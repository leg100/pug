find demo/configs -name .terraform -exec rm -rf {} \; > /dev/null 2>&1 || true
find demo/configs -name terraform.tfstate* -exec rm -rf {} \; > /dev/null 2>&1 || true
