#!/usr/bin/env bash

find ./demos/workspaces -name .terraform -exec rm -rf {} \; > /dev/null 2>&1 || true
find ./demos/workspaces -name terraform.tfstate -exec rm -rf {} \; > /dev/null 2>&1 || true
