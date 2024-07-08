terraform {
  backend "local" {}
}

data "terraform_remote_state" "vpc" {
  backend = "local"

  config = {
    path = "../vpc/terraform.tfstate"
  }
}

output "name" {
  value = "redis depends on ${data.terraform_remote_state.vpc.outputs.name}"
}
