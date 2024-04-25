#!/usr/bin/env bash

find ./demos/state -name .terraform -exec rm -rf {} \; > /dev/null 2>&1 || true
find ./demos/state -name terraform.tfstate* -exec rm -rf {} \; > /dev/null 2>&1 || true
find ./demos/state -name .pug -exec rm -rf {} \; > /dev/null 2>&1 || true
