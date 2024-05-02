#!/usr/bin/env bash

find ./demos/tasks -name .terraform -exec rm -rf {} \; > /dev/null 2>&1 || true
find ./demos/tasks -name terraform.tfstate -exec rm -rf {} \; > /dev/null 2>&1 || true
