terraform {
  backend "local" {}
}

// Requires that an http server be setup locally on port 12345 - if the server
// doesn't respond to the request then this should hang indefinitely.
data "http" "forever" {
  url = "http://localhost:12345"
}
