#!/usr/bin/env bash

# in foreground, continously run app
while true; do
    _build/pug -f tasks -w ./demos/workspaces -d
done
