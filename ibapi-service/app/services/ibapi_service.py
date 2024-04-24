# services/ibapi_service.py
from ibapi.client import EClient
from ibapi.wrapper import EWrapper
from datetime import datetime
from zoneinfo import ZoneInfo
from threading import Thread
import asyncio
from ..core.config import settings
from .notification_service import NotificationService

def serialize_contract(contract):
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
        self.orders = {}
        self.observers = []
        self.notifier = NotificationService(self.observers)

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
        self.notifier.schedule_notification(execution_data)

    def connect_and_start(self):
        self.connect(settings.ib_host, settings.ib_port, clientId=settings.ib_client_id)
        thread = Thread(target=self.run_thread, daemon=True)
        thread.start()

    def run_thread(self):
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
        self.run()

    def openOrder(self, orderId, contract, order, orderState):
        current_time = datetime.now(ZoneInfo(settings.timezone)).strftime('%Y-%m-%d %H:%M:%S')
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
        self.orders[orderId] = order_info
        self.notifier.schedule_notification({"type": "order", "data": order_info})

    def orderStatus(self, orderId, status, filled, remaining, avgFillPrice, permId, parentId, lastFillPrice, clientId, whyHeld, mktCapPrice):
        print("Order Status:", orderId, status, filled)
        if orderId in self.orders:
            order = self.orders[orderId]
            order.update({
                "status": status,
                "filled": filled,
                "remaining": remaining,
                "avgFillPrice": avgFillPrice,
                "lastFillPrice": lastFillPrice
            })
            self.notifier.schedule_notification({"type": "order_update", "data": order})