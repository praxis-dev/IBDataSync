#  services/notification_service.py

import asyncio
from starlette.websockets import WebSocketState

class NotificationService:
    def __init__(self, observers):
        self.observers = observers

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
