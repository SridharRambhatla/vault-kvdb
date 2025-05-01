#!/bin/bash
set -e

trap 'killall go-kvdb' SIGINT

cd $(dirname $0)

killall go-kvdb || true
sleep 0.1

go install -v

# Default cache size is 1GB (1024*1024*1024 bytes)
CACHE_SIZE=${KVDB_CACHE_SIZE:-1073741824}

echo "Starting go-kvdb nodes with cache size: $CACHE_SIZE bytes..."

go-kvdb -db-location=database/sh-1.db -http-addr=127.0.0.1:8080 -config-file=sharding.toml -shard=sh-1 -cache-size=$CACHE_SIZE &
go-kvdb -db-location=database/sh-2.db -http-addr=127.0.0.1:8081 -config-file=sharding.toml -shard=sh-2 -cache-size=$CACHE_SIZE &
go-kvdb -db-location=database/sh-3.db -http-addr=127.0.0.1:8082 -config-file=sharding.toml -shard=sh-3 -cache-size=$CACHE_SIZE &

wait