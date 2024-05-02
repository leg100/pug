#!/usr/bin/env bash

find ./demos/getting_started -name .terraform -exec rm -rf {} \; > /dev/null 2>&1 || true
find ./demos/getting_started -name terraform.tfstate* -exec rm -rf {} \; > /dev/null 2>&1 || true
