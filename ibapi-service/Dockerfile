FROM python:3.8-slim

WORKDIR /app

COPY requirements.txt .

RUN pip install --trusted-host pypi.python.org -r requirements.txt

EXPOSE 5000

ENV UVICORN_CMD="uvicorn app:app --host 0.0.0.0 --port 5000"

CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "5000"]
