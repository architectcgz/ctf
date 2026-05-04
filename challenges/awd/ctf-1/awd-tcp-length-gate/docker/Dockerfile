FROM python:3.12-alpine

WORKDIR /app
COPY app.py challenge_app.py ctf_runtime.py /app/

ENV PORT=8080
EXPOSE 8080

CMD ["python", "/app/app.py"]
