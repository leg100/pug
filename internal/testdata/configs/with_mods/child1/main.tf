resource "random_integer" "suffix" {
  min = "1000"
  max = "9999"
  keepers = {
    now = timestamp()
  }
}

resource "random_pet" "pet" {
  keepers = {
    now = timestamp()
  }
}

output "pet" {
  value = random_pet.pet.id
}

