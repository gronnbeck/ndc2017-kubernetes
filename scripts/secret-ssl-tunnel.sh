#!/bin/bash

kubectl create secret generic compose-ssl-tunnel \
  --from-literal=host=$COMPOSE_REDIS_ADDR \
  --from-literal=port=$COMPOSE_REDIS_PORT \
  --from-literal=dst_ip=$COMPOSE_REDIS_PROXY_IP  \
  --from-literal=src_port=$COMPOSE_REDIS_PROXY_PORT
