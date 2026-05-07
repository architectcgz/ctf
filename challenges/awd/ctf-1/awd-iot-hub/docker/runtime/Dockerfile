FROM python:3.12-slim

WORKDIR /app
COPY runtime/requirements.txt /tmp/requirements.txt
RUN pip install --no-cache-dir -r /tmp/requirements.txt
COPY runtime/app.py runtime/ctf_runtime.py /app/
COPY workspace /workspace
COPY workspace/src/challenge_app.py /app/challenge_app.py

ENV PYTHONPATH=/workspace/src

EXPOSE 8080
CMD ["python", "app.py"]
