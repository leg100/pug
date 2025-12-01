terraform {
  backend "local" {}

  required_providers {
    google = {
      version = "= 5.42.0"
    }
  }
}

provider "google" {
  region = "us-east-1"
}

resource "google_compute_instance" "default" {
  name         = "my-instance"
  machine_type = var.instance_type
  zone         = "us-central1-a"

  tags = ["foo", "bar"]

  boot_disk {
    initialize_params {
      image = "debian-cloud/debian-11"
      labels = {
        my_label = "value"
      }
    }
  }

  // Local SSD disk
  scratch_disk {
    interface = "NVME"
  }

  network_interface {
    network = "default"

    access_config {
      // Ephemeral public IP
    }
  }

  metadata = {
    foo = "bar"
  }

  metadata_startup_script = "echo hi > /test.txt"

  service_account {
    # Google recommends custom service accounts that have cloud-platform scope and permissions granted via IAM Roles.
    email  = google_service_account.default.email
    scopes = ["cloud-platform"]
  }
}

resource "google_service_account" "default" {
  account_id   = "my-custom-sa"
  display_name = "Custom SA for VM Instance"
}

variable "instance_type" {
  default = "n2-standard-2"
}
