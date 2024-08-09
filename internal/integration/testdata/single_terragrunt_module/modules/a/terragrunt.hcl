terraform {
  source = "."
}

remote_state {
  backend = "local"
  config = {}
}
