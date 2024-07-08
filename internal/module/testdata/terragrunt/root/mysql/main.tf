terraform {
  backend "local" {}
}

data "terraform_remote_state" "vpc" {
  backend = "local"

  config = {
    path = "../vpc"
  }
}

output "name" {
  value = "mysql depends on ${terraform_remote_state.vpc.outputs.name}"
}
