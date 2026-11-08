# Runtime identity for the Cloud Run service.
resource "google_service_account" "this" {
  account_id   = "srv-${local.service}"
  display_name = "srv-${local.service}"
  description  = "${local.service} service account."
}