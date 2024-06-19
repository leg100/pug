#!/usr/bin/env bash

find ./demos/modules -name .terraform -exec rm -rf {} \; > /dev/null 2>&1 || true
find ./demos/modules -name terraform.tfstate* -exec rm -rf {} \; > /dev/null 2>&1 || true
