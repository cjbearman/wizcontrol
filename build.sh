#!/bin/bash

(cd cmd/wizcontrol && go build ./...) && \
docker build -t wizcontrol:latest .

