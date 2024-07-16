terraform {
  backend "local" {}
}

variable "time_to_wait" {}

resource "time_sleep" "wait_n_seconds" {
  create_duration = var.time_to_wait
}
