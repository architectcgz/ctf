FROM python:3.12-alpine

WORKDIR /app
COPY runtime/app.py runtime/ctf_runtime.py /app/

ENV PYTHONPATH=/workspace/src

ENV PORT=8080
EXPOSE 8080

CMD ["python", "/app/app.py"]
