#!/usr/bin/env bash

# in foreground, continously run app
while true; do
    _build/pug -w ./internal/state/testdata -d
done
