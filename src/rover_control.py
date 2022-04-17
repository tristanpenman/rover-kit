#!/usr/bin/env python3

import aiohttp
import asyncio
import http.server
import json
import os
import socketserver
import time
import websockets

from aiohttp import web

from adafruit_motorkit import MotorKit

threshold = 0.5

kit = MotorKit()

script_dir = os.path.dirname(__file__)


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
    else:
        stop()


async def index_handler(request):
    return web.FileResponse('web/index.html')


async def websocket_handler(request):
    ws = web.WebSocketResponse()
    await ws.prepare(request)
    async for msg in ws:
        if msg.type == aiohttp.WSMsgType.TEXT:
            print(msg.data)
            try:
                payload = json.loads(msg.data)
                print(payload)
                if payload['type'] == 'forwards':
                    await ws.send_str('forwards')
                    forwards()
                elif payload['type'] == 'backwards':
                    await ws.send_str('backwards')
                    backwards()
                elif payload['type'] == 'spin_cw':
                    await ws.send_str('spin_cw')
                    spin_cw()
                elif payload['type'] == 'spin_ccw':
                    await ws.send_str('spin_ccw')
                    spin_ccw()
                elif payload['type'] == 'stop':
                    await ws.send_str('stop')
                    stop()
                elif payload['type'] == 'throttle':
                    await ws.send_str('throttle');
                    throttle(payload['value']);
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
    loop.run_until_complete(start_server())
    loop.run_forever()
