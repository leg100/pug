terraform {
  backend "local" {}
}

resource "random_integer" "seconds" {
  min = "1"
  max = "2"
  keepers = {
    now = timestamp()
  }
}

resource "time_sleep" "wait_random_seconds" {
  create_duration = "${random_integer.seconds.result}s"

  triggers = {
    rerun = random_integer.seconds.result
  }
}
