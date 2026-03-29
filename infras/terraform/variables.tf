variable "gcp_project_id" {
  description = "GCP project ID for Firestore"
  type        = string
}

variable "gcp_region" {
  description = "GCP region for Firestore database"
  type        = string
  default     = "us-central1"
}

variable "firestore_database_name" {
  description = "Firestore database name (default database uses '(default)')"
  type        = string
  default     = "(default)"
}

variable "create_firestore_database" {
  description = "Whether to create the Firestore database (set false if it already exists)"
  type        = bool
  default     = true
}

variable "service_account_name" {
  description = "Service account ID (short name) for Kotomi"
  type        = string
  default     = "kotomi-firestore"
}

variable "create_service_account_key" {
  description = "Whether to create a service account key (useful for GOOGLE_APPLICATION_CREDENTIALS)"
  type        = bool
  default     = false
}

variable "auth0_domain" {
  description = "Auth0 tenant domain (e.g., your-tenant.auth0.com)"
  type        = string
}

variable "auth0_mgmt_client_id" {
  description = "Auth0 Management API client ID"
  type        = string
  sensitive   = true
}

variable "auth0_mgmt_client_secret" {
  description = "Auth0 Management API client secret"
  type        = string
  sensitive   = true
}

variable "auth0_app_name" {
  description = "Auth0 application name"
  type        = string
  default     = "Kotomi Admin"
}

variable "auth0_callback_urls" {
  description = "Allowed callback URLs for Auth0"
  type        = list(string)
}

variable "auth0_logout_urls" {
  description = "Allowed logout URLs for Auth0"
  type        = list(string)
}

variable "auth0_web_origins" {
  description = "Allowed web origins for Auth0"
  type        = list(string)
}
