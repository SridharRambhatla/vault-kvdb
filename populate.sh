#!/bin/bash

echo $RANDOM

for shard in localhost:8080 localhost:8081; do
    for i in {1..1000}; do
        echo curl "http://$shard/set?key=$RANDOM&value=$RANDOM&bucketName=default"
    done
done