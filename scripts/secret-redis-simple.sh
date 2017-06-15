#!/bin/bash

kubectl create secret generic redis-api-simple \
  --from-literal=redis-addr=$CONFIGMAP_MASTER_REDIS_ADDR \
  --from-literal=redis-pass=$CONFIGMAP_MASTER_REDIS_PASS
