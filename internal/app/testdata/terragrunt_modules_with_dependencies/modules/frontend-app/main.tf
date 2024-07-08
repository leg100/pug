terraform {
  backend "local" {}
}

data "terraform_remote_state" "vpc" {
  backend = "local"

  config = {
    path = "../vpc/terraform.tfstate"
  }
}

data "terraform_remote_state" "backend" {
  backend = "local"

  config = {
    path = "../backend-app/terraform.tfstate"
  }
}

output "name" {
  value = "frontend depends on ${data.terraform_remote_state.vpc.outputs.name}, ${data.terraform_remote_state.backend.outputs.name}"
}
