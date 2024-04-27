terraform {
  backend "local" {}
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
