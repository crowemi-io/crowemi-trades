resource "google_firestore_database" "crowmei-trades" {
  project     = var.gcp_project_id
  name        = local.service
  location_id = var.gcp_region
  type        = "FIRESTORE_NATIVE"
}
