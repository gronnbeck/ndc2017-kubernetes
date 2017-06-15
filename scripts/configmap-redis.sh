#!/bin/bash

_port=$CONFIGMAP_REDIS_PORT
_masterHost=$CONFIGMAP_REDIS_ADDR
_masterPort=$COMPOSE_REDIS_PROXY_PORT
_masterPass=$CONFIGMAP_MASTER_REDIS_PASS

cat ./manifest/redis.conf.tmpl | \
sed -e s/'$REDIS_MASTER_HOST'/${_masterHost:-127.0.0.1}/g | \
sed -e s/'$REDIS_MASTER_PORT'/${_masterPort:-6379}/g | \
sed -e s/'$REDIS_MASTER_PASS'/$_masterPass/g | \
sed -e s/'$REDIS_PORT'/${_port:-6379}/g \
> redis.conf

kubectl create configmap redis-sidekick --from-file=redis.conf

kubectl create configmap redis-write \
  --from-literal=host=$CONFIGMAP_MASTER_REDIS_ADDR \
  --from-literal=pass=$CONFIGMAP_MASTER_REDIS_PASS
