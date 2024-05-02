#!/usr/bin/env bash

find ./demos/runs -name .terraform -exec rm -rf {} \; > /dev/null 2>&1 || true
find ./demos/runs -name terraform.tfstate* -exec rm -rf {} \; > /dev/null 2>&1 || true
