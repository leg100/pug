terraform {
  backend "local" {}
}

// Requires that an http server be setup locally on $url - if the server
// doesn't respond to the request then this should hang indefinitely.
data "http" "forever" {
  url = var.url
}

variable "url" {}
