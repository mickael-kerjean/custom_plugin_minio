from kafka import KafkaConsumer
from minio import Minio
import logging
import requests
import base64
import json
import hashlib

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
    handlers=[
        logging.FileHandler("debug.log"),
        logging.StreamHandler()
    ]
)

kafka_consumer = KafkaConsumer(
  'put.file',
  bootstrap_servers=['kafka:9092'],
  auto_offset_reset='latest',
  enable_auto_commit=False,
  group_id='indexer1',
  value_deserializer=lambda x: json.loads(x.decode('utf-8'))
)
kafka_consumer.subscribe('put.file')

minio_client = Minio(
  'minio:9000',
  'Q3AM3UQ867SPQQA43P2F', 
  'zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG',
  secure=False
)

request_headers = {'Content-Type':'application/json'}

# Configure Ingest Attachment Pipeline
attachment_config = '{"description" : "Extract attachment information","processors" : [{"attachment" : {"field" : "data"}},{"remove": {"field": "data"}}]}'
r = requests.put('http://elasticsearch:9200/_ingest/pipeline/attachment', headers = request_headers, data = attachment_config)
logging.info('Configure Ingest Attachment Pipeline: %s', r.status_code)

if r.status_code != 200:
  raise SystemExit('Unable to configure ingest attachment pipeline')

while True:
  for msg in kafka_consumer:
    path = msg.value['Key']
    unique_id = hashlib.sha1(path.encode('utf-8')).hexdigest()
    logging.info('Processing new put file event %s: %s', unique_id, path)

    bucket, obj_name = path.split('/', 1)
    obj = minio_client.get_object(bucket, obj_name).read()
    logging.info('Retrieved %s object from MinIO: %s', unique_id, path)

    base64_encoded_data = base64.b64encode(obj).decode("utf-8") 
    logging.info('Encoding %s object into base64 string', unique_id)
    
    request_json = {'data' : base64_encoded_data, 'path' : path}
    try:
      response = requests.put('http://elasticsearch:9200/minio_file/_doc/' + unique_id + '?pipeline=attachment', data = json.dumps(request_json), headers = request_headers)
      logging.info('Trigger ingest attachment pipeline %s return status_code: %s', unique_id, response.status_code)
    except requests.exceptions.HTTPError as err:
      print(err.response.text)
