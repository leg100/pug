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

module "child1" {
  source = "./child1"
}

module "child2" {
  source = "./child2"
}

output "pet1" {
  value = random_pet.pet.id
}

output "child_pet1" {
  value = module.child1.pet
}

output "child_pet2" {
  value = module.child2.pet
}

