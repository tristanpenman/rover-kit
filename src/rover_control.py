#!/usr/bin/env python3

import aiohttp
import asyncio
import http.server
import socketserver
import time
import websockets

from aiohttp import web

from adafruit_motorkit import MotorKit


kit = MotorKit()


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


async def index_handler(request):
    return web.FileResponse('web/index.html')


async def websocket_handler(request):
    ws = web.WebSocketResponse()
    await ws.prepare(request)
    async for msg in ws:
        if msg.type == aiohttp.WSMsgType.TEXT:
            if msg.data == 'forwards':
                await ws.send_str('forwards')
                forwards()
                await ws.send_str('ok')
            elif msg.data == 'backwards':
                await ws.send_str('backwards')
                backwards()
                await ws.send_str('ok')
            elif msg.data == 'spin_cw':
                await ws.send_str('spin_cw')
                spin_cw()
                await ws.send_str('ok')
            elif msg.data == 'spin_ccw':
                await ws.send_str('spin_ccw')
                spin_ccw()
                await ws.send_str('ok')
            elif msg.data == 'stop':
                await ws.send_str('stop')
                stop()
                await ws.send_str('ok')
            else:
                await ws.send_str('some websocket message payload: ' + msg.data)
        elif msg.type == aiohttp.WSMsgType.ERROR:
            print('ws connection closed with exception %s' % ws.exception())

    return ws


def create_runner():
    app = web.Application()
    app.add_routes([
        web.get('/', index_handler),
        web.get('/ws', websocket_handler)
    ])
    app.router.add_static('/', path='web', name='web')
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
