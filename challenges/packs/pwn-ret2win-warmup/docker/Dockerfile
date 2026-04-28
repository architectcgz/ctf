FROM debian:bookworm-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends gcc libc6-dev socat \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /opt/chal

COPY src/challenge.c /opt/chal/challenge.c
RUN gcc -O0 -fno-stack-protector -no-pie -o /opt/chal/challenge /opt/chal/challenge.c \
    && chmod 755 /opt/chal/challenge

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

EXPOSE 8080

CMD ["/entrypoint.sh"]
