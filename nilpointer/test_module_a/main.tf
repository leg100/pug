terraform {
  backend "local" {}
}

locals {
  time_to_wait = "1s"
}

resource "time_sleep" "wait_n_seconds" {
  create_duration = local.time_to_wait
}

output "time_waited" {
  value = time_sleep.wait_n_seconds.create_duration
}
