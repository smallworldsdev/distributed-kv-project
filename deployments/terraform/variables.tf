# variables.tf: Configuration Variables
variable "project_id" {
  description = "The GCP Project ID"
  type        = string
}

variable "region" {
  description = "Region for the GKE cluster"
  type        = string
  default     = "us-central1"
}
