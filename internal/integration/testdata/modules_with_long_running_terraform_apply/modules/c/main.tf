terraform {
  backend "local" {}
}

resource "time_sleep" "wait" {
  create_duration = "5s"
}

output "waited" {
  value = time_sleep.wait.create_duration
}
