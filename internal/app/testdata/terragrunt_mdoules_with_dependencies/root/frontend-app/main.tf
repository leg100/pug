terraform {
  backend "local" {}
}

data "terraform_remote_state" "vpc" {
  backend = "local"

  config = {
    path = "../vpc"
  }
}

data "terraform_remote_state" "backend" {
  backend = "local"

  config = {
    path = "../backend"
  }
}

output "name" {
  value = "frontend depends on ${terraform_remote_state.vpc.outputs.name}, ${terraform_remote_state.backend.outputs.name}"
}
