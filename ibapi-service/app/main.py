from fastapi import FastAPI
from ibapi.client import EClient
from ibapi.wrapper import EWrapper
from ibapi.execution import ExecutionFilter
from ibapi.contract import Contract

from threading import Thread
import time

app = FastAPI()

class IBapi(EWrapper, EClient):
    def __init__(self):
        EClient.__init__(self, self)
        self.executions = []
        self.orders = []  


    def execDetails(self, reqId, contract, execution):
        print("Order Executed: ", execution.execId, execution.acctNumber, execution.side, execution.shares, execution.price)
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
        print("Order Submitted: ", orderId, contract.symbol, order.action, order.totalQuantity, order.lmtPrice)
        self.orders.append({
            "orderId": orderId,
            "contract": contract,
            "order": order,
            "orderState": orderState,
            "status": "",
            "filled": 0,
            "remaining": 0
        })

    def orderStatus(self, orderId, status, filled, remaining, avgFillPrice, permId, parentId, lastFillPrice, clientId, whyHeld, mktCapPrice):
        print("Order Status: ", orderId, status, filled)
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
    global app_ib
    return {"status": "IBapi is initialized" if 'app_ib' in globals() and app_ib.isConnected() else "IBapi not initialized"}

@app.on_event("startup")
async def startup_event():
    global app_ib
    app_ib = IBapi()
    app_ib.connect_and_start()

@app.get("/executions")
def get_executions():
    return {"executions": [str(e) for e in app_ib.executions]}

@app.get("/orders")
def get_orders():
    return {"orders": [serialize_order(order) for order in app_ib.orders]}

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
    }
