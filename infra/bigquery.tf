resource "google_bigquery_dataset" "firestore_export" {
  dataset_id = "firestore_export"
  project    = var.gcp_project_id
  location   = var.gcp_region
}