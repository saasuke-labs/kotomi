# Terraform: GCP Firestore + Auth0

This module provisions:

- Firestore API enablement + (optional) Firestore database
- Service account with Firestore access
- Auth0 Regular Web Application for Kotomi admin login

## Prereqs

- Terraform >= 1.5
- GCP project + billing enabled
- Auth0 Management API application with `create:clients` permissions

## Usage

1. Create a `terraform.tfvars` with your values.

Required variables:

- `gcp_project_id`
- `auth0_domain`
- `auth0_mgmt_client_id`
- `auth0_mgmt_client_secret`
- `auth0_callback_urls`
- `auth0_logout_urls`
- `auth0_web_origins`

Optional:

- `gcp_region` (default: us-central1)
- `create_firestore_database` (default: true)
- `create_service_account_key` (default: false)

2. Apply:

- `terraform init`
- `terraform plan`
- `terraform apply`

## App configuration (Kotomi)

Set these environment variables for production:

- `DB_PROVIDER=firestore`
- `FIRESTORE_PROJECT_ID=<gcp_project_id>`
- `GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json` (if you created a key)
- `AUTH0_DOMAIN=<tenant>`
- `AUTH0_CLIENT_ID=<auth0_client_id>`
- `AUTH0_CLIENT_SECRET=<auth0_client_secret>`
- `AUTH0_CALLBACK_URL=https://<your-domain>/callback`
- `SESSION_SECRET=<random-32+>`

## Firestore indexes

Firestore composite indexes are defined in `firestore.indexes.json` at the repo root. Apply them via gcloud or the console.
