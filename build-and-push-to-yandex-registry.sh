#!/bin/sh

yc container registry configure-docker

REGISTRY_ID=crpsssmuhjtqus0tfuhb

docker buildx build --load -t cr.yandex/$REGISTRY_ID/metrics-aggregator:latest .
docker push cr.yandex/$REGISTRY_ID/metrics-aggregator:latest
