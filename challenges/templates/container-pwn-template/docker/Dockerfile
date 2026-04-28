FROM debian:bookworm

RUN apt-get update \
    && apt-get install -y --no-install-recommends gcc libc6-dev socat python3 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /opt/chal
COPY src/challenge.c /opt/chal/challenge.c
RUN gcc -O0 -fno-stack-protector -no-pie -z execstack -s -o /opt/chal/challenge /opt/chal/challenge.c

COPY site/ /srv/site/
RUN mkdir -p /srv/site/downloads \
    && cp /opt/chal/challenge /srv/site/downloads/challenge \
    && chmod 755 /srv/site/downloads/challenge

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

EXPOSE 80 9999

CMD ["/entrypoint.sh"]
