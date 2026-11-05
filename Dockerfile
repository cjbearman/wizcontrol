FROM golang:1.26 AS builder
WORKDIR /src
COPY . /src/
RUN go build -o cmd/wizcontrol/wizcontrol ./cmd/wizcontrol/wizcontrol.go
RUN go build -o cmd/wizfind/wizfind ./cmd/wizfind/wizfind.go


FROM ubuntu:latest
RUN apt-get update && \
	apt-get upgrade && \
	apt-get install tzdata && \
	mkdir -p /app /config

VOLUME /config
WORKDIR /config


COPY --from=builder /src/cmd/wizcontrol/wizcontrol /app
COPY --from=builder /src/cmd/wizfind/wizfind /app
COPY config.yaml /config/config.yaml

ENV PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/app

CMD [ "wizcontrol", "-d" ]

