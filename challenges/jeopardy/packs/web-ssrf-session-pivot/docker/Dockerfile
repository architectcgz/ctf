FROM python:3.12-alpine

WORKDIR /app
COPY app.py /app/app.py

ENV FLAG=flag{local_web_ssrf_session_pivot}
ENV SESSION_SECRET=relay-signing-secret

EXPOSE 8080

CMD ["python", "/app/app.py"]
