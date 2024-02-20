#!/usr/bin/env bash

while true; do
    go build -o pug && pkill -f './pug'
    inotifywait -e attrib $(find . -name '*.go') || exit
done
