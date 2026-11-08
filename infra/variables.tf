variable "gcp_project_id" {
  description = "GCP Project ID hosting the site (e.g. crowemi-dev)"
  type        = string
}

variable "gcp_region" {
  description = "GCP region (e.g. us-west1)"
  type        = string
}

# Supplied by the global `gcp` Terraform Cloud variable set. Unused here, but
# must be declared or Terraform rejects the values the variable set provides
# to every workspace.
variable "gcp_organization_id" {
  description = "GCP organization ID (unused; required by global variable set)"
  type        = string
  default     = ""
}

variable "gcp_billing_account_id" {
  description = "GCP billing account ID (unused; required by global variable set)"
  type        = string
  default     = ""
}

variable "env" {
  description = "Environment (dev or prod)"
  type        = string
}