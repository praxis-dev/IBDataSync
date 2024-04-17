from fastapi import WebSocket, APIRouter
from starlette.websockets import  WebSocketDisconnect

router = APIRouter()

@router.websocket("/ws")
async def websocket_endpoint(websocket: WebSocket):
    await websocket.accept()
    try:
        while True:
            # Here we assume `app` is globally accessible or replace this with the correct way you manage state
            websocket.app.app_ib.observers.append(websocket)
            await websocket.receive_text()
    except WebSocketDisconnect:
        print("WebSocket disconnected.")
    except Exception as e:
        print(f"WebSocket error: {e}")
    finally:
        websocket.app.app_ib.observers.remove(websocket)
        await websocket.close()
