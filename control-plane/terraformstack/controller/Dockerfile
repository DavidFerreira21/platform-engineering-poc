FROM python:3.13-slim
WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY controller.py leadership.py real.py .
USER 10001:10001
ENTRYPOINT ["python", "-u", "/app/controller.py"]
