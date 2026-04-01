# locals {
#   firestore_export_collections = toset([
#     "accounts",
#     "activities",
#     "corporate_actions",
#     "orders",
#     "portfolios",
#     "positions",
#     "watchlists",
#     "withholdings",
#   ])
# }

# resource "google_firebase_extensions_instance" "firestore_bigquery_export" {
#   for_each    = local.firestore_export_collections
#   provider    = google-beta
#   project     = var.gcp_project_id
#   instance_id = "firestore-bigquery-export-${each.key}"

#   config {
#     extension_ref     = "firebase/firestore-bigquery-export"
#     extension_version = "0.2.9"
#     params = {
#       COLLECTION_PATH     = each.key
#       DATASET_ID          = google_bigquery_dataset.firestore_export.dataset_id
#       TABLE_ID            = each.key
#       BIGQUERY_PROJECT_ID = var.gcp_project_id
#       DATABASE            = local.service
#       DATABASE_REGION     = var.gcp_region
#       DATASET_LOCATION    = var.gcp_region
#     }
#     system_params = {
#       "firebaseextensions.v1beta.function/location" = var.gcp_region
#     }
#   }
# }

# resource "google_bigquery_dataset" "firestore_export" {
#   dataset_id = "firestore_export"
#   project    = var.gcp_project_id
#   location   = var.gcp_region
# }