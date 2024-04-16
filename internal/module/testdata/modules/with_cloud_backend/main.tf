terraform {
  cloud {
    organization = "my-org"
    hostname     = "app.terraform.io" # Optional; defaults to app.terraform.io

    workspaces {
      project = "networking-development"
      tags    = ["networking", "source:cli"]
    }
  }
}
