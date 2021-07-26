import newrelic.agent
import time
import redis
import os
import json
import requests
import time
import random


newrelic.agent.initialize('/usr/src/app/newrelic.ini')

@newrelic.agent.background_task(name="testing")
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

        log_message(message)
