# core/config/config.py

from pydantic import BaseSettings

class Settings(BaseSettings):
    ib_host: str = "127.0.0.1"
    ib_port: int = 7496
    ib_client_id: int = 0
    timezone: str = "America/New_York"
    
    class Config:
        env_file = ".env"

settings = Settings()
