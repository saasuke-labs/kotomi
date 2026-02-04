# Release Setup Quick Start Guide

This is a quick reference for setting up the release pipeline. For detailed explanations, see [RELEASE_PROCESS.md](RELEASE_PROCESS.md).

## Overview

The release pipeline automatically builds and deploys Kotomi to Google Cloud Run when changes are pushed to the `main` branch.

**Prerequisites**: GCP account with billing, GitHub admin access, gcloud CLI installed

## Step 1: GCP Setup (15-20 minutes)

### 1.1 Create/Select Project
```bash
gcloud projects create YOUR-PROJECT-ID --name="Kotomi Production"
gcloud config set project YOUR-PROJECT-ID
```
Then enable billing at: https://console.cloud.google.com/billing

### 1.2 Enable APIs
```bash
gcloud services enable artifactregistry.googleapis.com
gcloud services enable run.googleapis.com
```

### 1.3 Create Artifact Registry
```bash
REGION="us-central1"  # Change to your preferred region
gcloud artifacts repositories create kotomi \
  --repository-format=docker \
  --location=$REGION \
  --description="Kotomi Docker images"
```

### 1.4 Create Service Account
```bash
# Create service account
gcloud iam service-accounts create github-actions-deployer \
  --display-name="GitHub Actions Deployer"

# Get service account email
SA_EMAIL=$(gcloud iam service-accounts list \
  --filter="displayName:GitHub Actions Deployer" \
  --format='value(email)')
```

### 1.5 Grant Permissions
```bash
PROJECT_ID="YOUR-PROJECT-ID"  # Replace with your project ID

gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/artifactregistry.writer"

gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/run.admin"

gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/iam.serviceAccountUser"

gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/storage.admin"
```

### 1.6 Create Service Account Key
```bash
gcloud iam service-accounts keys create ~/gcp-key.json \
  --iam-account=${SA_EMAIL}

# Display key (copy this for GitHub Secrets)
cat ~/gcp-key.json
```

⚠️ **IMPORTANT**: Keep this key secure! Delete local copy after adding to GitHub.

## Step 2: GitHub Secrets Setup (5 minutes)

Go to: `https://github.com/saasuke-labs/kotomi/settings/secrets/actions`

### Add Three Secrets:

#### 2.1 GCP_SA_KEY
- Click **New repository secret**
- Name: `GCP_SA_KEY`
- Value: Paste entire JSON from `~/gcp-key.json`
- Click **Add secret**

#### 2.2 GCP_PROJECT_ID
- Click **New repository secret**
- Name: `GCP_PROJECT_ID`
- Value: Your GCP project ID (e.g., `kotomi-production-12345`)
- Click **Add secret**

#### 2.3 GCP_REGION
- Click **New repository secret**
- Name: `GCP_REGION`
- Value: Your region (e.g., `us-central1`)
- Click **Add secret**

### Delete Local Key File
```bash
rm ~/gcp-key.json
```

## Step 3: Configure Cloud Run Environment (5 minutes)

Set environment variables for your deployed application:

```bash
REGION="us-central1"  # Must match GCP_REGION secret

gcloud run services update kotomi-prod \
  --region=$REGION \
  --set-env-vars="AUTH0_DOMAIN=YOUR_AUTH0_DOMAIN" \
  --set-env-vars="AUTH0_CLIENT_ID=YOUR_AUTH0_CLIENT_ID" \
  --set-env-vars="AUTH0_CLIENT_SECRET=YOUR_AUTH0_CLIENT_SECRET" \
  --set-env-vars="SESSION_SECRET=YOUR_SESSION_SECRET_MIN_32_CHARS" \
  --set-env-vars="OPENAI_API_KEY=YOUR_OPENAI_KEY" \
  --set-env-vars="SMTP_HOST=smtp.gmail.com" \
  --set-env-vars="SMTP_PORT=587" \
  --set-env-vars="SMTP_USER=YOUR_EMAIL" \
  --set-env-vars="SMTP_PASSWORD=YOUR_APP_PASSWORD"
```

**Note**: If the service doesn't exist yet, the first deployment will create it automatically.

## Step 4: Test Deployment (5 minutes)

Push a change to main branch to trigger deployment:

```bash
git checkout main
git pull origin main

# Make a test change
echo "# Test" >> README.md
git add README.md
git commit -m "Test: Verify CI/CD pipeline"
git push origin main
```

Monitor the deployment:
1. Go to https://github.com/saasuke-labs/kotomi/actions
2. Click on latest "Build and Deploy Kotomi" workflow
3. Watch each step complete successfully

## Step 5: Verify Deployment (2 minutes)

```bash
# Get service URL
SERVICE_URL=$(gcloud run services describe kotomi-prod \
  --region=$REGION \
  --format='value(status.url)')

echo "Service URL: $SERVICE_URL"

# Test health endpoint
curl $SERVICE_URL/healthz

# Open in browser
open $SERVICE_URL/admin/dashboard
```

## Checklist

- [ ] GCP project created with billing enabled
- [ ] Artifact Registry repository created
- [ ] Service account created with all required permissions
- [ ] Service account key generated
- [ ] Three GitHub secrets added: GCP_SA_KEY, GCP_PROJECT_ID, GCP_REGION
- [ ] Local service account key file deleted
- [ ] Cloud Run environment variables configured
- [ ] Test deployment successful
- [ ] Service accessible and health check passing

## Common Issues

### Authentication Failed
```bash
# Verify service account has correct roles
gcloud projects get-iam-policy PROJECT_ID \
  --flatten="bindings[].members" \
  --filter="bindings.members:serviceAccount:${SA_EMAIL}"
```

### Repository Not Found
```bash
# Verify repository exists
gcloud artifacts repositories list --location=$REGION
```

### Deployment Fails
```bash
# Check Cloud Run logs
gcloud run services logs read kotomi-prod --region=$REGION --limit=50
```

## Next Steps

Once setup is complete:

1. **Update VERSION file** to prepare for release
2. **Follow release checklist** in RELEASE_PROCESS.md
3. **Monitor deployment** using Cloud Run console
4. **Configure alerts** in Cloud Console

## Additional Resources

- **Full Documentation**: [RELEASE_PROCESS.md](RELEASE_PROCESS.md)
- **Deployment Monitoring**: [DEPLOYMENT_MONITORING.md](DEPLOYMENT_MONITORING.md)
- **GCP Console**: https://console.cloud.google.com
- **GitHub Actions**: https://github.com/saasuke-labs/kotomi/actions
- **Cloud Run Console**: https://console.cloud.google.com/run

## Time Estimate

- **Total setup time**: ~30-35 minutes (first time)
- **Subsequent deployments**: Automatic (5-10 minutes per deployment)

---

**Last Updated**: February 4, 2026
