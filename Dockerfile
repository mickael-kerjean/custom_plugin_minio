FROM ubuntu:21.04
MAINTAINER mickael@kerjean.me

ENV TZ Europe/Berlin
ENV DEBIAN_FRONTEND noninteractive

COPY filestash/dist/ /app/
RUN apt-get -y update && \
    apt install -y ca-certificates libvips-dev && \
    useradd filestash && \
    chown -R filestash:filestash /app/ && \
    sed -i 's|"admin":.*||' /app/data/state/config/config.json && \
    sed -i 's|"secret_key":.*||' /app/data/state/config/config.json && \
    rm -rf /var/lib/apt/lists/* && \
    rm -rf /tmp/*

USER filestash
RUN timeout 1 /app/filestash | grep -q start

EXPOSE 8334
VOLUME ["/app/data/state/"]
CMD ["/app/filestash"]