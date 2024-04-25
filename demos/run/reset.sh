#!/usr/bin/env bash

find ./demos/run -name .terraform -exec rm -rf {} \; > /dev/null 2>&1 || true
find ./demos/run -name terraform.tfstate* -exec rm -rf {} \; > /dev/null 2>&1 || true
find ./demos/run -name .pug -exec rm -rf {} \; > /dev/null 2>&1 || true
