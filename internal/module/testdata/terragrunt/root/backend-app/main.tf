terraform {
  backend "local" {}
}

data "terraform_remote_state" "vpc" {
  backend = "local"

  config = {
    path = "../vpc"
  }
}

data "terraform_remote_state" "mysql" {
  backend = "local"

  config = {
    path = "../mysql"
  }
}

data "terraform_remote_state" "redis" {
  backend = "local"

  config = {
    path = "../redis"
  }
}

output "name" {
  value = "backend depends on ${terraform_remote_state.vpc.outputs.name}, ${terraform_remote_state.mysql.outputs.name}, ${terraform_remote_state.redis.outputs.name}"
}
