resource "google_service_account" "this" {
  account_id   = "srv-${local.service}"
  display_name = "srv-${local.service}-${var.env}"
  description  = "A service account for ${local.service}"
}

resource "google_project_iam_member" "this" {
  project = var.gcp_project_id
  role    = "roles/datastore.user"
  member  = "serviceAccount:${google_service_account.this.email}"

  # Restrict to the Firestore database only (not project-wide)
  condition {
    title       = "firestore_database_only"
    description = "Grant read/write only to the crowemi-trades Firestore database"
    expression  = "resource.name == \"projects/${var.gcp_project_id}/databases/${local.service}\""
  }
}
