#!/bin/sh

./redisapi \
--read-redis-addr=$REDIS_ADDR_READ \
--write-redis-pass=$REDIS_PASS_WRITE \
--write-redis-pass=$REDIS_PASS_WRITE
