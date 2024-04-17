from fastapi import FastAPI
from ibapi.client import EClient
from ibapi.wrapper import EWrapper
from datetime import datetime
from zoneinfo import ZoneInfo
from threading import Thread
import time

app = FastAPI()

class IBapi(EWrapper, EClient):
    def __init__(self):
        EClient.__init__(self, self)
        self.executions = []
        self.orders = []

    def execDetails(self, reqId, contract, execution):
        print("Order Executed:", execution.execId, execution.acctNumber, execution.side, execution.shares, execution.price)
        self.executions.append(execution)

    def connect_and_start(self):
        self.connect("127.0.0.1", 7496, clientId=0)
        thread = Thread(target=self.run)
        thread.start()
        setattr(self, "_thread", thread)
        while not self.isConnected():
            time.sleep(1)
        self.reqAllOpenOrders()
        self.reqAutoOpenOrders(True)

    def openOrder(self, orderId, contract, order, orderState):
        current_time = datetime.now(ZoneInfo("America/New_York")).strftime('%Y-%m-%d %H:%M:%S')
        print(f"Order Submitted at {current_time}:", orderId, contract.symbol, order.action, order.totalQuantity, order.lmtPrice)

        if not any(o["orderId"] == orderId for o in self.orders):
            self.orders.append({
                "orderId": orderId,
                "contract": contract,
                "order": order,
                "orderState": orderState,
                "status": orderState.status,
                "filled": 0,
                "remaining": order.totalQuantity,
                "submissionTime": current_time
            })

    def orderStatus(self, orderId, status, filled, remaining, avgFillPrice, permId, parentId, lastFillPrice, clientId, whyHeld, mktCapPrice):
        print("Order Status:", orderId, status, filled)
        for order in self.orders:
            if order["orderId"] == orderId:
                order.update({
                    "status": status,
                    "filled": filled,
                    "remaining": remaining
                })

@app.get("/")
def read_root():
    return {"Hello": "World"}

@app.get("/check-ibapi")
def check_ibapi():
    return {"status": "IBapi is initialized" if hasattr(app, 'app_ib') and app.app_ib.isConnected() else "IBapi not initialized"}

@app.on_event("startup")
async def startup_event():
    app.app_ib = IBapi()
    app.app_ib.connect_and_start()

@app.get("/executions")
def get_executions():
    return {"executions": [str(e) for e in app.app_ib.executions]}

@app.get("/orders")
def get_orders():
    return {"orders": [serialize_order(order) for order in app.app_ib.orders]}

def serialize_order(order):
    return {
        "orderId": order["orderId"],
        "symbol": order["contract"].symbol,
        "action": order["order"].action,
        "totalQuantity": order["order"].totalQuantity,
        "limitPrice": order["order"].lmtPrice,
        "orderType": order["order"].orderType,
        "status": order["status"],
        "filled": order["filled"],
        "remaining": order["remaining"],
        "submissionTime": order["submissionTime"]
    }
