#!/bin/bash
set -e

trap 'killall go-kvdb' SIGINT

cd $(dirname $0)

killall go-kvdb || true
sleep 0.1

go install -v

go-kvdb -db-location=database/sh-1.db -http-addr=127.0.0.1:8080 -config-file=sharding.toml -shard=sh-1
go-kvdb -db-location=database/sh-2.db -http-addr=127.0.0.1:8081 -config-file=sharding.toml -shard=sh-2 &
go-kvdb -db-location=database/sh-3.db -http-addr=127.0.0.1:8082 -config-file=sharding.toml -shard=sh-3 &

wait