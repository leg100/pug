terraform {
  backend "local" {}
  required_providers {
    random = {
      version = "= 3.6.0"
    }
  }
}

resource "random_pet" "pet" {
  count = 10

  keepers = {
    now = timestamp()
  }
}

output "pets" {
  value = random_pet.pet[*].id
}
