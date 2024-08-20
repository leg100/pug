terraform {
  backend "local" {}
}

# Configure the AWS Provider
provider "aws" {
  region = "us-east-1"
}

resource "aws_instance" "web" {
  ami           = data.aws_ami.ubuntu.id
  instance_type = var.instance_type

  tags = {
    Name = "HelloWorld"
  }
}

variable "instance_type" {
  default = "t3.micro"
}
