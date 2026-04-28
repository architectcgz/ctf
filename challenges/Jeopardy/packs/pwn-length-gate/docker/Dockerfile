FROM python:3.12-alpine

WORKDIR /app
COPY server.py /app/server.py

ENV FLAG=flag{pwn_length_gate_passed}
EXPOSE 8080

CMD ["python", "/app/server.py"]
