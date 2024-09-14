# remove all asdf installs of terraform
asdf uninstall terraform

find demo/ -name .terraform -exec rm -rf {} \; > /dev/null 2>&1 || true
find demo/ -name terraform.tfstate -exec rm {} \; > /dev/null 2>&1 || true
find demo/ -name terraform.tfstate.* -exec rm {} \; > /dev/null 2>&1 || true
find demo/ -name .terraform.lock.hcl -exec rm {} \; > /dev/null 2>&1 || true
find demo/ -name environment -exec rm {} \; > /dev/null 2>&1 || true
