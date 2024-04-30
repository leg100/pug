terraform {
  backend "local" {}
}

variable "foo" {
  default = "bar"
}

output "foo" {
  value = var.foo
}
