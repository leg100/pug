terraform {
  backend "s3" {
    bucket = "pug-terragrunt-example-terraform-state-dev-us-east-1"
    key    = "terraform.tfstate"
    region = "us-east-1"
  }
}

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

resource "aws_s3_bucket" "pug" {
  bucket = "pug-${random_pet.pet.id}-${random_integer.suffix.id}"
}
