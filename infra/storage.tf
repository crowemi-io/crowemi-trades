# General-purpose app bucket. 
resource "google_storage_bucket" "app" {
  project                     = var.gcp_project_id
  name                        = var.gcp_project_id
  location                    = var.gcp_region
  uniform_bucket_level_access = true
  public_access_prevention    = "enforced"
}