FROM python:3.12-alpine

WORKDIR /app
COPY app.py /app/app.py

ENV FLAG=flag{web_header_door_opened}
EXPOSE 8080

CMD ["python", "/app/app.py"]
