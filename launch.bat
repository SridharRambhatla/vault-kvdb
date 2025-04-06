@echo off
setlocal enabledelayedexpansion

echo Killing existing go-kvdb processes...
taskkill /F /IM go-kvdb.exe >nul 2>&1

echo Building go-kvdb...
go install -v
if errorlevel 1 (
    echo Build failed.
    exit /b 1
)

echo Starting go-kvdb nodes...

start "" go-kvdb.exe -db-location=database/sh-1.db -http-addr=127.0.0.1:8080 -config-file=sharding.toml -shard=sh-1
start "" go-kvdb.exe -db-location=database/sh-2.db -http-addr=127.0.0.1:8081 -config-file=sharding.toml -shard=sh-2
start "" go-kvdb.exe -db-location=database/sh-3.db -http-addr=127.0.0.1:8082 -config-file=sharding.toml -shard=sh-3
echo All nodes started.
pause
