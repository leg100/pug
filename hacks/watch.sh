#!/usr/bin/env bash

# in foreground, continously run app
while true; do
    _build/pug -w ./demo/dont_cost_money -d -l debug
done
