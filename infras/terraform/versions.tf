terraform {
  required_version = ">= 1.5.0"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = ">= 5.20.0"
    }
    auth0 = {
      source  = "auth0/auth0"
      version = ">= 1.5.0"
    }
  }
}

provider "google" {
  project = var.gcp_project_id
  region  = var.gcp_region
}

provider "auth0" {
  domain        = var.auth0_domain
  client_id     = var.auth0_mgmt_client_id
  client_secret = var.auth0_mgmt_client_secret
}
