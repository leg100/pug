terraform {
  backend "s3" {
    bucket = "pug-terragrunt-example-terraform-state-dev-us-east-1"
    key    = "terraform.tfstate"
    region = "us-east-1"
  }
}
