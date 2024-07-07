#!/usr/bin/env bash

# in foreground, continously run app
while true; do
    _build/pug -w ./demo -d
done
