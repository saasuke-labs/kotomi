output "firestore_service_account_email" {
  value       = google_service_account.kotomi.email
  description = "Service account email for Kotomi"
}

output "firestore_service_account_key_json" {
  value       = var.create_service_account_key ? google_service_account_key.kotomi[0].private_key : null
  description = "Base64-encoded service account key JSON (decode before writing to file)"
  sensitive   = true
}

output "auth0_client_id" {
  value       = auth0_client.kotomi_admin.client_id
  description = "Auth0 application client ID"
}

output "auth0_client_name" {
  value       = auth0_client.kotomi_admin.name
  description = "Auth0 application name"
}
