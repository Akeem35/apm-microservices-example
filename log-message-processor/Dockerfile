FROM python:3.7-alpine

WORKDIR /usr/src/app
RUN apk add --no-cache build-base
COPY requirements.txt .
RUN pip3 install -r requirements.txt

COPY . .

ENV NEW_RELIC_CONFIG_FILE=/usr/src/app/newrelic.ini
CMD ["python3","-u","main.py"]

