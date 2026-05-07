FROM python:3.12-alpine

WORKDIR /app
COPY runtime/app.py runtime/ctf_runtime.py /app/
COPY workspace /workspace
COPY workspace/src/challenge_app.py /app/challenge_app.py

ENV PYTHONPATH=/workspace/src

ENV PORT=8080
EXPOSE 8080

CMD ["python", "/app/app.py"]
