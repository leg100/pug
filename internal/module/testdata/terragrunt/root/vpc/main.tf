terraform {
  backend "local" {}
}

output "name" {
  value = "vpc"
}
