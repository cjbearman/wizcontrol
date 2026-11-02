FROM ubuntu:latest

RUN apt-get update && \
	apt-get upgrade && \
	apt-get install tzdata && \
	mkdir -p /app /config

VOLUME /config
WORKDIR /app


COPY cmd/wizcontrol/wizcontrol /app
COPY config.yaml /config/config.yaml

ENTRYPOINT ["/app/wizcontrol", "-d", "-config", "/config/config.yaml"]

