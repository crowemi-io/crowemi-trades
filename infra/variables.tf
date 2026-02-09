variable "gcp_project_id" {
  description = "GCP Project ID"
  type        = string
}
variable "gcp_region" {
  description = "GCP Region"
  type        = string
}
variable "gcp_organization_id" {
  description = "GCP Organization ID"
  type        = string
}
variable "gcp_billing_account_id" {
  description = "GCP Billing Account ID"
  type        = string
}
variable "env" {
  description = "Environment (e.g., dev/prod)"
  type        = string
}
