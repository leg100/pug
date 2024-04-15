#!/usr/bin/env bash

find ./demo -name .terraform -exec rm -rf {} \; > /dev/null 2>&1 || true
find ./demo -name .terraform.lock.hcl -exec rm {} \; > /dev/null 2>&1 || true
find ./demo -name terraform.tfstate -exec rm {} \; > /dev/null 2>&1 || true
find ./demo -name terraform.tfstate.backup -exec rm {} \; > /dev/null 2>&1 || true
