FROM python:3.9-slim

WORKDIR /app

COPY . /app

RUN pip install --trusted-host pypi.python.org -r requirements.txt

EXPOSE 5000

ENV UVICORN_CMD="uvicorn app.main:app --host 0.0.0.0 --port 5000"

CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "5000", "--reload"]
