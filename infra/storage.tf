# General-purpose app bucket, named after the project. It is in the project
# default region (var.gcp_region), no lifecycle rule.
# No IAM yet: grant the runtime SA object access (for example
# roles/storage.objectUser on this bucket) when the app starts using it.
resource "google_storage_bucket" "app" {
  project                     = var.gcp_project_id
  name                        = var.gcp_project_id
  location                    = var.gcp_region
  uniform_bucket_level_access = true
  public_access_prevention    = "enforced"
}