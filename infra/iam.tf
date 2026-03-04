resource "google_service_account" "this" {
  account_id   = "srv-${local.service}"
  display_name = "srv-${local.service}-${var.env}"
  description  = "A service account for ${local.service}"
}

# Permissions required for Firebase Extensions (e.g. firestore-bigquery-export) to deploy.
# The extension deploys Cloud Functions that need access to the GCF source bucket and Eventarc.

data "google_project" "project" {
  project_id = var.gcp_project_id
}

# Default Compute Engine SA must be able to read the Cloud Functions source bucket
# (gcf-sources-{PROJECT_NUMBER}-{REGION}) during function build.
resource "google_project_iam_member" "compute_sa_storage_object_viewer" {
  project = var.gcp_project_id
  role    = "roles/storage.objectViewer"
  member  = "serviceAccount:${data.google_project.project.number}-compute@developer.gserviceaccount.com"
}

# Eventarc Service Agent must have Eventarc Service Agent role so triggers can be created
# for the extension's Cloud Functions (e.g. Firestore document write trigger).
resource "google_project_iam_member" "eventarc_service_agent" {
  project = var.gcp_project_id
  role    = "roles/eventarc.serviceAgent"
  member  = "serviceAccount:service-${data.google_project.project.number}@gcp-sa-eventarc.iam.gserviceaccount.com"
}