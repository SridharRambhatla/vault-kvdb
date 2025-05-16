#!/bin/bash

echo $RANDOM

for shard in localhost:8080 localhost:8081; do
    echo curl "http://$shard/set?key=$RANDOM&value=$RANDOM&bucketName=default"
done