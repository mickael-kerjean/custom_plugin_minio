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

minio_client = Minio(
  'localhost:9000',
  'minio',
  'minio123',
  secure=False
)

request_headers = {'Content-Type':'application/json'}

# Configure Ingest Attachment Pipeline
attachment_config = '{"description" : "Extract attachment information","processors" : [{"attachment" : {"field" : "data"}},{"remove": {"field": "data"}}]}'
r = requests.put('http://localhost:9200/_ingest/pipeline/attachment', headers = request_headers, data = attachment_config)
logging.info('Configure Ingest Attachment Pipeline: %s', r.status_code)

if r.status_code != 200:
  raise SystemExit('Unable to configure ingest attachment pipeline')

bucket = 'test'
prefix = 'filestash'
objects = minio_client.list_objects(bucket, prefix=prefix, recursive=True)
for obj in objects:
    if obj.is_dir:
      continue
    path = '/' + bucket + '/' + obj.object_name
    unique_id = hashlib.sha1(path.encode('utf-8')).hexdigest()
    logging.info('Processing new put file event %s: %s', unique_id, path)

    size = obj.size

    base64_encoded_data = ""
    # Get data of an object.
    obj_data = minio_client.get_object(bucket, obj.object_name).read()
    logging.info('Retrieved %s object from MinIO: %s', unique_id, path)

    # Read data from response.
    base64_encoded_data = base64.b64encode(obj_data).decode("utf-8")

    logging.info('Encoding %s object into base64 string', unique_id)

    request_json = {'data' : base64_encoded_data, 'path' : path, 'size' : size}
    try:
      response = requests.put('http://localhost:9200/minio_file/_doc/' + unique_id + '?pipeline=attachment', data = json.dumps(request_json), headers = request_headers)
      logging.info('Trigger ingest attachment pipeline %s return status_code: %s', unique_id, response.status_code)
    except requests.exceptions.HTTPError as err:
      print(err.response.text)