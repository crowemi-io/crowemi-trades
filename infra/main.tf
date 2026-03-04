locals {
    service = "crowemi-trades"
}

resource "google_cloud_run_v2_service" "this" {
  provider             = google-beta
  name                 = local.service
  location             = var.gcp_region
  ingress              = "INGRESS_TRAFFIC_ALL"
  launch_stage         = "BETA"
  default_uri_disabled = true
  template {
    containers {
      image = "us-west1-docker.pkg.dev/${var.gcp_project_id}/${var.gcp_project_id}/${local.service}:${var.docker_image_tag}"
      env {
        name = "CONFIG"
        value_source {
          secret_key_ref {
            secret  = data.google_secret_manager_secret.this.secret_id
            version = "latest"
          }
        }
      }
    }
    service_account = google_service_account.this.email
    vpc_access {
      network_interfaces {
        network    = "projects/crowemi-io-${var.env}/global/networks/crowemi-io-${var.env}-network"                         # TODO: ref data
        subnetwork = "projects/crowemi-io-${var.env}/regions/${var.gcp_region}/subnetworks/crowemi-io-${var.env}-subnet-01" # TODO: ref data
      }
      egress = "ALL_TRAFFIC"
    }
  }
}