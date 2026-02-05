# GCP Deployment Setup Guide

This guide covers how to set up Kotomi for automated deployment to Google Cloud Platform (GCP) via GitHub Actions.

## Overview

Kotomi uses GitHub Actions to automatically build, test, and deploy to Google Cloud Run whenever changes are pushed to the `main` branch. The workflow:

1. Runs unit tests, integration tests, and E2E tests
2. Builds a Docker image with version metadata
3. Pushes the image to Google Artifact Registry
4. Deploys to Cloud Run

## Prerequisites

- A GCP project with billing enabled
- Google Cloud CLI installed locally (for initial setup)
- Admin access to your GitHub repository

> **⚠️ Important Note**: When setting up Cloud Run deployments via GitHub Actions, the service account used by GitHub Actions must have permission to act as the Cloud Run runtime service account. This guide includes all necessary IAM bindings in Step 4. If you encounter "Permission 'iam.serviceaccounts.actAs' denied" errors, see the Troubleshooting section.

## GCP Setup

### 1. Create a GCP Project

If you don't have a GCP project yet:

```bash
gcloud projects create kotomi-production --name="Kotomi Production"
gcloud config set project kotomi-production
```

### 2. Enable Required APIs

```bash
# Enable required GCP APIs
gcloud services enable run.googleapis.com
gcloud services enable artifactregistry.googleapis.com
gcloud services enable cloudbuild.googleapis.com
```

### 3. Create Artifact Registry Repository

Create a Docker repository to store your images:

```bash
# Replace REGION with your preferred region (e.g., us-central1, europe-west1)
export REGION="us-central1"
export PROJECT_ID=$(gcloud config get-value project)

gcloud artifacts repositories create kotomi \
  --repository-format=docker \
  --location=$REGION \
  --description="Kotomi container images"
```

### 4. Create Service Account

Create a service account for GitHub Actions:

```bash
# Create service account
gcloud iam service-accounts create github-actions \
  --display-name="GitHub Actions Deployment"

# Get the service account email
export SA_EMAIL="github-actions@${PROJECT_ID}.iam.gserviceaccount.com"

# Grant necessary permissions
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/run.admin"

gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/artifactregistry.writer"

gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/iam.serviceAccountUser"

# Allow GitHub Actions SA to act as the Compute Engine default service account
# This is required for Cloud Run to deploy services
export PROJECT_NUMBER=$(gcloud projects describe $PROJECT_ID --format='value(projectNumber)')
gcloud iam service-accounts add-iam-policy-binding \
  ${PROJECT_NUMBER}-compute@developer.gserviceaccount.com \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/iam.serviceAccountUser"
```

### 5. Create Service Account Key

Generate a JSON key for the service account:

```bash
gcloud iam service-accounts keys create github-actions-key.json \
  --iam-account=$SA_EMAIL
```

**Important**: This file contains sensitive credentials. Never commit it to version control!

## GitHub Secrets Configuration

Add the following secrets to your GitHub repository:

1. Go to your repository on GitHub
2. Navigate to Settings > Secrets and variables > Actions
3. Click "New repository secret" for each of the following:

### Required Secrets

| Secret Name | Description | Example Value |
|-------------|-------------|---------------|
| `GCP_PROJECT_ID` | Your GCP project ID | `kotomi-production` |
| `GCP_REGION` | GCP region for deployment | `us-central1` |
| `GCP_SA_KEY` | Service account JSON key | Contents of `github-actions-key.json` |

### Setting the Secrets

```bash
# For GCP_SA_KEY, copy the entire contents of the JSON file
cat github-actions-key.json
# Copy the output and paste it as the secret value

# Alternatively, use GitHub CLI
gh secret set GCP_PROJECT_ID --body "kotomi-production"
gh secret set GCP_REGION --body "us-central1"
gh secret set GCP_SA_KEY < github-actions-key.json
```

## Workflow Configuration

The deployment workflow is defined in `.github/workflows/deploy_kotomi.yaml` and includes:

### Workflow Triggers

- **Push to main**: Runs all tests and deploys on success
- **Pull requests**: Runs all tests only (no deployment)

### Build and Deploy Job

The `build-and-deploy` job:
1. Authenticates with GCP using the service account
2. Sets up the gcloud CLI
3. Configures Docker authentication for Artifact Registry
4. Builds Docker image with version metadata from `VERSION` file
5. Pushes image to Artifact Registry
6. Deploys to Cloud Run with:
   - Port: 8080
   - Memory: 512Mi
   - CPU: 1
   - Max instances: 10
   - Timeout: 300 seconds
   - Public access (unauthenticated)

## Cloud Run Configuration

The deployment creates/updates a Cloud Run service named `kotomi-prod` with the following configuration:

- **Port**: 8080 (matches the Docker container's exposed port)
- **Memory**: 512Mi (sufficient for SQLite-based application)
- **CPU**: 1 (adequate for moderate traffic)
- **Max Instances**: 10 (auto-scales based on load)
- **Timeout**: 300 seconds (5 minutes)
- **Authentication**: Allows unauthenticated requests (public access)

### Customizing the Deployment

To modify these settings, edit the `Deploy to Cloud Run` step in `.github/workflows/deploy_kotomi.yaml`:

```yaml
- name: Deploy to Cloud Run
  run: |
    IMAGE="${{ secrets.GCP_REGION }}-docker.pkg.dev/${{ secrets.GCP_PROJECT_ID }}/kotomi/kotomi:${{ steps.vars.outputs.version }}"
    echo "Deploying $IMAGE to Cloud Run"
    gcloud run deploy kotomi-prod \
      --image "$IMAGE" \
      --project ${{ secrets.GCP_PROJECT_ID }} \
      --region ${{ secrets.GCP_REGION }} \
      --platform managed \
      --allow-unauthenticated \
      --port 8080 \
      --memory 512Mi \
      --cpu 1 \
      --max-instances 10 \
      --timeout 300
```

### Environment Variables

To add environment variables to your Cloud Run deployment:

```yaml
--set-env-vars "KEY1=value1,KEY2=value2"
```

For secrets, use Secret Manager:

```yaml
--update-secrets "SECRET_NAME=projects/PROJECT_ID/secrets/SECRET_NAME/versions/latest"
```

## Verifying Deployment

### Check Workflow Status

1. Go to your repository on GitHub
2. Click the "Actions" tab
3. View the latest workflow run

### Check Cloud Run Service

```bash
# List services
gcloud run services list --platform managed --region $REGION

# Get service URL
gcloud run services describe kotomi-prod \
  --platform managed \
  --region $REGION \
  --format 'value(status.url)'
```

### Test the Deployment

```bash
# Get the service URL
SERVICE_URL=$(gcloud run services describe kotomi-prod \
  --platform managed \
  --region $REGION \
  --format 'value(status.url)')

# Test the health endpoint
curl $SERVICE_URL/healthz
```

Expected response:
```json
{
  "message": "OK"
}
```

## Troubleshooting

### Deployment Fails: Authentication Error

**Symptom**: `ERROR: (gcloud.auth.configure-docker) Docker configuration failed`

**Solution**: Verify that:
- The `GCP_SA_KEY` secret contains a valid JSON service account key
- The service account has the necessary permissions
- The service account key is not expired

### Deployment Fails: Permission Denied

**Symptom**: `ERROR: Permission denied` when pushing to Artifact Registry or deploying to Cloud Run

**Solution**: Ensure the service account has the following roles:
- `roles/run.admin`
- `roles/artifactregistry.writer`
- `roles/iam.serviceAccountUser`

### Deployment Fails: Service Account actAs Permission Error

**Symptom**: 
```
ERROR: Permission 'iam.serviceaccounts.actAs' denied on service account 
[PROJECT_NUMBER]-compute@developer.gserviceaccount.com
```

**Root Cause**: The GitHub Actions service account doesn't have permission to act as the Compute Engine default service account, which is required for Cloud Run deployments.

**Solution**: Grant the GitHub Actions service account permission to act as the compute service account:

```bash
# Get your project ID and number
export PROJECT_ID=$(gcloud config get-value project)
export PROJECT_NUMBER=$(gcloud projects describe $PROJECT_ID --format='value(projectNumber)')
export SA_EMAIL="github-actions@${PROJECT_ID}.iam.gserviceaccount.com"

# Grant actAs permission on the compute service account
gcloud iam service-accounts add-iam-policy-binding \
  ${PROJECT_NUMBER}-compute@developer.gserviceaccount.com \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/iam.serviceAccountUser"
```

**Alternative Solution**: Use a custom service account for Cloud Run instead of the default compute service account:

1. Create a custom service account for Cloud Run:
```bash
gcloud iam service-accounts create kotomi-runtime \
  --display-name="Kotomi Runtime Service Account"
```

2. Grant it necessary permissions (if accessing other GCP services).

3. Update the deployment command in `.github/workflows/deploy_kotomi.yaml` to specify the service account:
```yaml
gcloud run deploy kotomi-prod \
  --image "$IMAGE" \
  --service-account "kotomi-runtime@${PROJECT_ID}.iam.gserviceaccount.com" \
  ...other flags...
```

### Docker Build Fails

**Symptom**: Build fails with network or dependency errors

**Solution**:
- Check the Dockerfile for any issues
- Ensure all dependencies in `go.mod` are accessible
- Verify the Go version matches the Dockerfile (1.25)

### Cloud Run Deployment Timeout

**Symptom**: Service deployment times out or health checks fail

**Solution**:
- Verify the container starts correctly locally: `docker run -p 8080:8080 kotomi`
- Check Cloud Run logs: `gcloud run services logs read kotomi-prod --region $REGION`
- Ensure the application listens on the correct port (8080)

## Monitoring and Logs

### View Logs

```bash
# Stream logs
gcloud run services logs tail kotomi-prod --region $REGION

# View logs in Cloud Console
https://console.cloud.google.com/run/detail/$REGION/kotomi-prod/logs
```

### Monitor Metrics

```bash
# View service details
gcloud run services describe kotomi-prod --region $REGION
```

For detailed monitoring, see [Deployment Monitoring Guide](./DEPLOYMENT_MONITORING.md).

## Database Persistence

**Important**: Cloud Run is stateless. The SQLite database stored in the container will be reset on each deployment.

For production use, consider:
1. Using Cloud SQL for PostgreSQL
2. Mounting a persistent volume (Cloud Run with mounts)
3. Using Cloud Storage for SQLite file backup/restore

See [Database Backup & Restore Guide](./DATABASE_BACKUP_RESTORE.md) for more details.

## Security Best Practices

1. **Rotate Service Account Keys**: Regularly rotate the GitHub Actions service account key
2. **Limit Permissions**: Follow the principle of least privilege for service accounts
3. **Enable VPC**: Consider using VPC Service Controls for enhanced security
4. **Use Secret Manager**: Store sensitive configuration in GCP Secret Manager
5. **Enable Cloud Armor**: Add DDoS protection and WAF rules
6. **Review IAM**: Regularly audit IAM permissions and access logs

## Cost Optimization

Cloud Run pricing is based on:
- Request count
- CPU/memory allocation
- Network egress

To optimize costs:
1. Set appropriate `--min-instances` (default is 0 for cost savings)
2. Tune `--memory` and `--cpu` based on actual usage
3. Use `--max-instances` to prevent runaway scaling
4. Monitor usage in Cloud Console billing section

## Next Steps

- Set up Cloud SQL for production database
- Configure custom domain with Cloud Load Balancer
- Set up Cloud CDN for static assets
- Implement Cloud Monitoring alerts
- Configure Cloud Armor for security
- Set up automated backups

## References

- [Cloud Run Documentation](https://cloud.google.com/run/docs)
- [Artifact Registry Documentation](https://cloud.google.com/artifact-registry/docs)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [gcloud CLI Reference](https://cloud.google.com/sdk/gcloud/reference)
