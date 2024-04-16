from fastapi import FastAPI
from ibapi.client import EClient
from ibapi.wrapper import EWrapper

app = FastAPI()

class IBapi(EWrapper, EClient):
    def __init__(self):
        EClient.__init__(self, self)

@app.get("/")
def read_root():
    return {"Hello": "Worldddd12"}

@app.get("/check-ibapi")
def check_ibapi():
    app = IBapi()
    return {"status": "IBapi initialized successfully"}
