#!/usr/bin/env bash

find ./demos -name .terraform -exec rm -rf {} \; > /dev/null 2>&1 || true
find ./demos -name terraform.tfstate* -exec rm -rf {} \; > /dev/null 2>&1 || true
find ./demos -name .pug -exec rm -rf {} \; > /dev/null 2>&1 || true
