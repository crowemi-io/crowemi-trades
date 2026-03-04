resource "google_service_account" "this" {
  account_id   = "srv-${local.service}"
  display_name = "srv-${local.service}-${var.env}"
  description  = "A service account for ${local.service}"
}
