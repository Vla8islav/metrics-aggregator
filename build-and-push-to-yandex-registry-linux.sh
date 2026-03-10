#!/bin/sh

yc container registry configure-docker

REGISTRY_ID=crpsssmuhjtqus0tfuhb

docker buildx build --platform linux/amd64 --load -t cr.yandex/$REGISTRY_ID/metrics-aggregator:latest --push .
