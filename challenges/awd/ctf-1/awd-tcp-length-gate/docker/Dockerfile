FROM python:3.12-alpine

WORKDIR /app
COPY service.py /app/service.py

ENV PORT=8080
EXPOSE 8080

CMD ["python", "/app/service.py"]
