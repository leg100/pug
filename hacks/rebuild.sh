#!/usr/bin/env bash

while true; do
    go build -o _build/pug && pkill -f '_build/pug'
    inotifywait -e attrib $(find . -name '*.go') || exit
done
