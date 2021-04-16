import time
import redis
import os
import json
import requests
import time
import random
import elasticapm

from elasticapm.utils.disttracing import TraceParent

from elasticapm import Client

client = Client({'SERVICE_NAME': 'python'})


@elasticapm.capture_span()
def log_message(message):
    time_delay = random.randrange(0, 2000)
    time.sleep(time_delay / 1000)
    print('message received after waiting for {}ms: {}'.format(time_delay, message))

if __name__ == '__main__':
    redis_host = os.environ['REDIS_HOST']
    redis_port = int(os.environ['REDIS_PORT'])
    redis_channel = os.environ['REDIS_CHANNEL']

    pubsub = redis.Redis(host=redis_host, port=redis_port, db=0).pubsub()
    pubsub.subscribe([redis_channel])
    for item in pubsub.listen():
        try:
            if isinstance(item['data'], int):
                #do nothing
                print('data is an integer')
                continue
            else:
                message = json.loads(str(item['data'].decode("utf-8")))
        except Exception as e:
            log_message(e)
            continue

        spanTransaction = message['spanTransaction']
        trace_parent1 = spanTransaction['context']['request']['headers']['traceparent']
        print('trace_parent_log: {}'.format(trace_parent1))
        trace_parent = TraceParent.from_string(trace_parent1)
        client.begin_transaction("logger-transaction", trace_parent=trace_parent)

        log_message(message)

        client.end_transaction('logger-transaction')
