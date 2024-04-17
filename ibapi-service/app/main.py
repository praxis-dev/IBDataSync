from fastapi import FastAPI, WebSocket
from .api.routes import router as api_router
from .api.websocket import router as ws_router
from ibapi.client import EClient
from ibapi.wrapper import EWrapper
from datetime import datetime
from zoneinfo import ZoneInfo
from threading import Thread
import asyncio
from starlette.websockets import WebSocketState


app = FastAPI()

app.include_router(api_router)
app.include_router(ws_router)

def serialize_contract(contract):
    """ Convert a Contract object into a JSON-serializable dictionary. """
    return {
        "symbol": contract.symbol,
        "secType": contract.secType,
        "currency": contract.currency,
        "exchange": contract.exchange
    }

class IBapi(EWrapper, EClient):
    def __init__(self):
        EClient.__init__(self, self)
        self.executions = []
        self.orders = []
        self.observers = []

    async def notify_observers(self, message):
        for observer in self.observers:
            if observer.client_state == WebSocketState.CONNECTED:
                try:
                    await observer.send_json(message)
                except RuntimeError as e:
                    print(f"Error sending message: {e}")

    def schedule_notification(self, message):
        loop = asyncio.get_event_loop()
        if loop.is_running():
            loop.create_task(self.notify_observers(message))
        else:
            loop.run_until_complete(self.notify_observers(message))



    def execDetails(self, reqId, contract, execution):
        print("Order Executed:", execution.execId, execution.acctNumber, execution.side, execution.shares, execution.price)
        self.executions.append(execution)
        execution_data = {"type": "execution", "data": {
            "execId": execution.execId,
            "acctNumber": execution.acctNumber,
            "side": execution.side,
            "shares": execution.shares,
            "price": execution.price
        }}
        self.schedule_notification(execution_data)

    def connect_and_start(self):
        self.connect("127.0.0.1", 7496, clientId=0)
        thread = Thread(target=self.run_thread, daemon=True)
        thread.start()

    def run_thread(self):
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
        self.run()

    def openOrder(self, orderId, contract, order, orderState):
            current_time = datetime.now(ZoneInfo("America/New_York")).strftime('%Y-%m-%d %H:%M:%S')
            print(f"Order Submitted at {current_time}:", orderId, contract.symbol, order.action, order.totalQuantity, order.lmtPrice)
            order_info = {
                "orderId": orderId,
                "contract": serialize_contract(contract),  
                "order": {
                    "action": order.action,
                    "totalQuantity": order.totalQuantity,
                    "lmtPrice": order.lmtPrice
                },
                "orderState": {
                    "status": orderState.status
                },
                "submissionTime": current_time
            }
            if not any(o["orderId"] == orderId for o in self.orders):
                self.orders.append(order_info)
            self.schedule_notification({"type": "order", "data": order_info})


    def orderStatus(self, orderId, status, filled, remaining, avgFillPrice, permId, parentId, lastFillPrice, clientId, whyHeld, mktCapPrice):
        print("Order Status:", orderId, status, filled)
        for order in self.orders:
            if order["orderId"] == orderId:
                order.update({
                    "status": status,
                    "filled": filled,
                    "remaining": remaining,
                    "avgFillPrice": avgFillPrice,
                    "lastFillPrice": lastFillPrice
                })
                self.schedule_notification({"type": "order_update", "data": order})

@app.on_event("startup")
async def startup_event():
    app.app_ib = IBapi()
    app.app_ib.connect_and_start()
