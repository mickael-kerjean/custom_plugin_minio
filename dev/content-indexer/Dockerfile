FROM python:3

WORKDIR /usr/src/app

ADD ./requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt

ADD ./app.py ./

CMD [ "python3", "./app.py" ]
