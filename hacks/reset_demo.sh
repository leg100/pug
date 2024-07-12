find demo/modules -name .terraform -exec rm -rf {} \; > /dev/null 2>&1 || true
find demo/modules -name terraform.tfstate -exec rm {} \; > /dev/null 2>&1 || true
find demo/modules -name terraform.tfstate.* -exec rm {} \; > /dev/null 2>&1 || true
find demo/modules -name environment -exec rm {} \; > /dev/null 2>&1 || true
