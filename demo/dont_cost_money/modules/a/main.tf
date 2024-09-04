terraform {
  backend "local" {}
  required_providers {
    random = {
      version = "= 3.6.0"
    }
  }
}

resource "time_sleep" "wait_three_seconds" {
  create_duration = "3s"
}

resource "random_pet" "pet" {
  count = 10

  keepers = {
    now = timestamp()
  }
}

resource "random_password" "password" {
  length           = 16
  special          = true
  override_special = "!#$%&*()-_=+[]{}<>:?"
}

output "waited" {
  value = time_sleep.wait_three_seconds.create_duration
}

output "pets" {
  value = random_pet.pet[*].id
}

output "secret" {
  sensitive = true
  value     = "topsecret"
}
