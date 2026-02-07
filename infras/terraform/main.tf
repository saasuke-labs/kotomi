resource "google_project_service" "firestore" {
  project = var.gcp_project_id
  service = "firestore.googleapis.com"
}

resource "google_firestore_database" "default" {
  count       = var.create_firestore_database ? 1 : 0
  project     = var.gcp_project_id
  name        = var.firestore_database_name
  location_id = var.gcp_region
  type        = "NATIVE"

  depends_on = [google_project_service.firestore]
}

resource "google_service_account" "kotomi" {
  account_id   = var.service_account_name
  display_name = "Kotomi Firestore Access"
}

resource "google_project_iam_member" "kotomi_firestore_user" {
  project = var.gcp_project_id
  role    = "roles/datastore.user"
  member  = "serviceAccount:${google_service_account.kotomi.email}"
}

resource "google_service_account_key" "kotomi" {
  count              = var.create_service_account_key ? 1 : 0
  service_account_id = google_service_account.kotomi.name
}

resource "auth0_client" "kotomi_admin" {
  name                                = var.auth0_app_name
  app_type                            = "regular_web"
  callbacks                           = var.auth0_callback_urls
  allowed_logout_urls                 = var.auth0_logout_urls
  web_origins                         = var.auth0_web_origins
  token_endpoint_auth_method          = "client_secret_post"
  oidc_conformant                     = true
  grant_types                         = ["authorization_code", "refresh_token"]
  jwt_configuration {
    alg = "RS256"
  }
}
