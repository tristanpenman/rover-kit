#!/usr/bin/env python3

import aiohttp
import asyncio
import json
import os
import time

import RPi.GPIO as GPIO

from aiohttp import web
from adafruit_motorkit import MotorKit

# Used to find web resources
script_dir = os.path.dirname(__file__)

# Set pin-numbering mode (Broadcom SOC channel number)
GPIO.setmode(GPIO.BCM)

# Sensor 1
GPIO_TRIGGER = 18
GPIO_ECHO = 24

# Set GPIO direction (IN / OUT)
GPIO.setup(GPIO_TRIGGER, GPIO.OUT)
GPIO.setup(GPIO_ECHO, GPIO.IN)

# Threshold for starting motors when analog input is used
threshold = 0.5

kit = MotorKit()

websockets = []


def stop():
    kit.motor1.throttle = 0
    kit.motor2.throttle = 0
    kit.motor3.throttle = 0
    kit.motor4.throttle = 0


def forwards():
    kit.motor1.throttle = -1
    kit.motor2.throttle = 1
    kit.motor3.throttle = -1
    kit.motor4.throttle = 1


def backwards():
    kit.motor1.throttle = 1
    kit.motor2.throttle = -1
    kit.motor3.throttle = 1
    kit.motor4.throttle = -1


def spin_cw():
    kit.motor1.throttle = 1
    kit.motor2.throttle = 1
    kit.motor3.throttle = 1
    kit.motor4.throttle = 1


def spin_ccw():
    kit.motor1.throttle = -1
    kit.motor2.throttle = -1
    kit.motor3.throttle = -1
    kit.motor4.throttle = -1


def throttle(value):
    if abs(value) > threshold:
        kit.motor1.throttle = value
        kit.motor2.throttle = -value
        kit.motor3.throttle = value
        kit.motor4.throttle = -value
        return True
    else:
        stop()
        return False


async def distance():
    global websockets

    while True:
        # set Trigger to HIGH
        GPIO.output(GPIO_TRIGGER, True)

        # set Trigger after 0.01ms to LOW
        time.sleep(0.00001)
        GPIO.output(GPIO_TRIGGER, False)

        StartTime = time.time()
        StopTime = time.time()

        # save start time, which is last instant at which ECHO is low
        while GPIO.input(GPIO_ECHO) == 0:
            StartTime = time.time()

        # save time of arrival, which is the last instant at which ECHO is high
        while GPIO.input(GPIO_ECHO) == 1:
            StopTime = time.time()

        # time difference between start and arrival
        TimeElapsed = StopTime - StartTime
        # multiply with the sonic speed (34300 cm/s)
        # and divide by 2, because there and back
        distance = (TimeElapsed * 34300) / 2

        # Publish distance measurement to all connected clients
        msg = json.dumps({'type': 'distance', 'value': distance})
        for ws in websockets:
            await ws.send_str(msg)

        await asyncio.sleep(1)


async def index_handler(_):
    return web.FileResponse('web/index.html')


async def websocket_handler(request):
    global websockets

    ws = web.WebSocketResponse()
    websockets.append(ws)

    await ws.prepare(request)

    async for msg in ws:
        if msg.type == aiohttp.WSMsgType.TEXT:
            print('ws message: ' + msg.data)
            try:
                payload = json.loads(msg.data)
                if payload['type'] == 'forwards':
                    forwards()
                    await ws.send_str(json.dumps({'type': 'forwards'}))
                elif payload['type'] == 'backwards':
                    backwards()
                    await ws.send_str(json.dumps({'type': 'backwards'}))
                elif payload['type'] == 'spin_cw':
                    spin_cw()
                    await ws.send_str(json.dumps({'type': 'spin_cw'}))
                elif payload['type'] == 'spin_ccw':
                    spin_ccw()
                    await ws.send_str(json.dumps({'type': 'spin_ccw'}))
                elif payload['type'] == 'stop':
                    stop()
                    await ws.send_str(json.dumps({'type': 'stop'}))
                elif payload['type'] == 'throttle':
                    active = throttle(payload['value'])
                    await ws.send_str(json.dumps({'type': 'throttle', 'active': active}))
                else:
                    await ws.send_str('invalid payload type: ' + payload['type'])

            except json.JSONDecodeError as err:
                await ws.send_str('bad request: ' + err.msg)

        elif msg.type == aiohttp.WSMsgType.ERROR:
            print('ws connection closed with exception %s' % ws.exception())

    return ws


def create_runner():
    app = web.Application()
    app.add_routes([
        web.get('/', index_handler),
        web.get('/ws', websocket_handler)
    ])
    app.router.add_static('/', path=os.path.join(script_dir, 'web'), name='web')
    return web.AppRunner(app)


async def start_server(host="0.0.0.0", port=8000):
    runner = create_runner()
    await runner.setup()
    site = web.TCPSite(runner, host, port)
    await site.start()


if __name__ == "__main__":
    stop()
    loop = asyncio.get_event_loop()
    asyncio.ensure_future(distance())
    loop.run_until_complete(start_server())
    loop.run_forever()
