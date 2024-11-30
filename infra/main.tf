locals {
  region       = var.google_region
  service      = var.service_name
  project      = var.google_project_id
  display_name = "srv_${replace(local.service, "-", "_")}"
  name         = local.service
}

resource "google_service_account" "this" {
  account_id   = "srv-${local.service}"
  display_name = local.display_name
  description  = "A service account for ${local.service}"
}


resource "google_cloud_run_v2_service" "this" {
  name     = local.name
  project  = local.project
  location = local.region
  ingress  = "INGRESS_TRAFFIC_ALL"
  template {
    containers {
      image = "us-west1-docker.pkg.dev/${local.project}/crowemi-io/${local.service}:${var.docker_image_tag}"
      ports {
        container_port = 3000
      }
      env {
        name = "URL"
        value_source {
          secret_key_ref {
            secret  = data.google_secret_manager_secret.this.secret_id
            version = "latest"
          }
        }
      }
    }
    scaling {
      max_instance_count = 1
    }
    vpc_access {
      network_interfaces {
        network    = "crowemi-io-network"
        subnetwork = "crowemi-io-subnet-01"
      }
      egress = "ALL_TRAFFIC"
    }
    service_account = google_service_account.this.email
  }
}

data "google_iam_policy" "noauth" {
    binding {
        role = "roles/run.invoker"
        members = [
            "allUsers",
        ]
    }
}


resource "google_cloud_run_service_iam_policy" "noauth" {
  location = google_cloud_run_v2_service.this.location
  project  = google_cloud_run_v2_service.this.project
  service  = google_cloud_run_v2_service.this.name

  policy_data = data.google_iam_policy.noauth.policy_data
}

