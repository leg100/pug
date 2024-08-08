terraform {
  backend "local" {}
}

data "terraform_remote_state" "vpc" {
  backend = "local"

  config = {
    path = "../vpc/terraform.tfstate"
  }
}

data "terraform_remote_state" "mysql" {
  backend = "local"

  config = {
    path = "../mysql/terraform.tfstate"
  }
}

data "terraform_remote_state" "redis" {
  backend = "local"

  config = {
    path = "../redis/terraform.tfstate"
  }
}

output "name" {
  value = "backend depends on ${data.terraform_remote_state.vpc.outputs.name}, ${data.terraform_remote_state.mysql.outputs.name}, ${data.terraform_remote_state.redis.outputs.name}"
}
