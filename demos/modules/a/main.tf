terraform {
  backend "local" {}
  required_providers {
    random = {
      version = "= 3.6.0"
    }
  }
}

resource "random_integer" "zero_to_nine" {
  min = "0"
  max = "9"
  keepers = {
    now = timestamp()
  }
}

resource "time_sleep" "wait_random_seconds" {
  create_duration = "${random_integer.zero_to_nine.result}s"

  triggers = {
    rerun = random_integer.zero_to_nine.result
  }
}

resource "random_pet" "pet" {
  count = 10

  keepers = {
    now = timestamp()
  }
}

output "waited" {
  value = time_sleep.wait_random_seconds.create_duration
}

output "pets" {
  value = random_pet.pet[*].id
}
