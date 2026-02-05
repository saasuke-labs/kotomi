# Quick Fix for Deployment Permission Error

If you're seeing this error in your GitHub Actions deployment:

```
ERROR: Permission 'iam.serviceaccounts.actAs' denied on service account 
[PROJECT_NUMBER]-compute@developer.gserviceaccount.com
```

## What's Wrong?

Your GitHub Actions service account doesn't have permission to act as the Compute Engine default service account when deploying to Cloud Run.

## Quick Fix (Recommended)

Run these commands in your terminal:

```bash
# Set your project ID
export PROJECT_ID="your-project-id"  # Replace with your actual project ID

# Get project number
export PROJECT_NUMBER=$(gcloud projects describe $PROJECT_ID --format='value(projectNumber)')

# Set service account email
export SA_EMAIL="github-actions@${PROJECT_ID}.iam.gserviceaccount.com"

# Grant the permission
gcloud iam service-accounts add-iam-policy-binding \
  ${PROJECT_NUMBER}-compute@developer.gserviceaccount.com \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/iam.serviceAccountUser" \
  --project=$PROJECT_ID
```

## What This Does

This command grants your GitHub Actions service account (`github-actions@[PROJECT].iam.gserviceaccount.com`) permission to act as the Compute Engine default service account (`[PROJECT_NUMBER]-compute@developer.gserviceaccount.com`), which is required for Cloud Run deployments.

## Verify the Fix

After running the command:

1. Re-run your failed GitHub Actions workflow
2. The deployment should now succeed

## Alternative Solution: Use Custom Service Account

If you prefer not to use the default Compute Engine service account, you can create a custom runtime service account for Cloud Run:

### Step 1: Create Runtime Service Account

```bash
gcloud iam service-accounts create kotomi-runtime \
  --display-name="Kotomi Runtime Service Account" \
  --project=$PROJECT_ID

export RUNTIME_SA_EMAIL="kotomi-runtime@${PROJECT_ID}.iam.gserviceaccount.com"
```

### Step 2: Grant Permissions (if needed)

Only grant permissions if your Cloud Run service needs to access other GCP services:

```bash
# Example: Grant access to Cloud Storage
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:${RUNTIME_SA_EMAIL}" \
  --role="roles/storage.objectViewer"
```

### Step 3: Allow GitHub Actions SA to Act as Runtime SA

```bash
gcloud iam service-accounts add-iam-policy-binding \
  $RUNTIME_SA_EMAIL \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/iam.serviceAccountUser" \
  --project=$PROJECT_ID
```

### Step 4: Update Workflow File

Edit `.github/workflows/deploy_kotomi.yaml` and add `--service-account` flag to the deployment command:

```yaml
- name: Deploy to Cloud Run
  run: |
    IMAGE="${{ secrets.GCP_REGION }}-docker.pkg.dev/${{ secrets.GCP_PROJECT_ID }}/kotomi/kotomi:${{ steps.vars.outputs.version }}"
    echo "Deploying $IMAGE to Cloud Run"
    gcloud run deploy kotomi-prod \
      --image "$IMAGE" \
      --project ${{ secrets.GCP_PROJECT_ID }} \
      --region ${{ secrets.GCP_REGION }} \
      --service-account "kotomi-runtime@${{ secrets.GCP_PROJECT_ID }}.iam.gserviceaccount.com" \
      --platform managed \
      --allow-unauthenticated \
      --port 8080 \
      --memory 512Mi \
      --cpu 1 \
      --max-instances 10 \
      --timeout 300
```

## Understanding the Permissions

### Required IAM Roles for GitHub Actions Service Account

The GitHub Actions service account needs these project-level roles:

1. **`roles/run.admin`** - To deploy and manage Cloud Run services
2. **`roles/artifactregistry.writer`** - To push Docker images
3. **`roles/iam.serviceAccountUser`** - To act as other service accounts

### Additional Permission Required

Additionally, it needs **service account-level** permission:
- Permission to act as the Cloud Run runtime service account (either default compute SA or custom SA)

This is granted with:
```bash
gcloud iam service-accounts add-iam-policy-binding \
  [RUNTIME_SERVICE_ACCOUNT] \
  --member="serviceAccount:[GITHUB_ACTIONS_SA]" \
  --role="roles/iam.serviceAccountUser"
```

## Why Is This Needed?

When Cloud Run deploys a service:
1. The deployment is done by the GitHub Actions service account
2. The running service executes under a runtime service account
3. The deploying SA must have permission to "act as" the runtime SA

This is a security feature to ensure that deployment accounts can't grant themselves arbitrary permissions by deploying services with highly privileged service accounts.

## Need Help?

See the full [GCP Deployment Setup Guide](./GCP_DEPLOYMENT_SETUP.md) for complete instructions.

## Security Best Practices

1. **Use custom runtime service accounts** for production deployments
2. **Grant minimal permissions** to both GitHub Actions and runtime service accounts
3. **Rotate service account keys** regularly (at least every 90 days)
4. **Audit IAM bindings** periodically to remove unnecessary permissions
5. **Enable Cloud Audit Logs** to track service account usage
