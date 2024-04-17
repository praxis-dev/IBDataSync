from fastapi import FastAPI

from .api.routes import router as api_router
from .api.websocket import router as ws_router
from .services.ibapi_service import IBapi

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


@app.on_event("startup")
async def startup_event():
    app.app_ib = IBapi()
    app.app_ib.connect_and_start()
