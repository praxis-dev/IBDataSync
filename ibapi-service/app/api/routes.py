# app/api/routes.py
from fastapi import APIRouter

router = APIRouter()

@router.get("/")
def read_root():
    return {"Hello": "World"}

@router.get("/check-ibapi")
def check_ibapi(app):
    return {"status": "IBapi is initialized" if hasattr(app, 'app_ib') and app.app_ib.isConnected() else "IBapi not initialized"}
