#!/usr/bin/env bash

while true; do
    go build -o _build/pug && pkill -f '_build/pug'

    # OSX doesn't have inotifywait, so we use fswatch instead
    if [[ "$(uname)" == "Darwin" ]]; then
      fswatch -e .git -o $(find . -name '*.go') || exit
    elif [[ "$(uname)" == "Linux" ]]; then
      inotifywait -e attrib $(find . -name '*.go') || exit
    else
      echo "Unknown OS: $(uname)"
    fi
done
