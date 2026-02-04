# Release Process Documentation

This document describes the release process for Kotomi, including initial setup, version management, release checklists, and rollback procedures.

## Table of Contents

1. [Prerequisites & Initial Setup](#prerequisites--initial-setup)
2. [GCP Configuration](#gcp-configuration)
3. [GitHub Secrets Configuration](#github-secrets-configuration)
4. [Versioning Strategy](#versioning-strategy)
5. [Release Types](#release-types)
6. [Release Checklist](#release-checklist)
7. [Version Bumping](#version-bumping)
8. [Deployment Process](#deployment-process)
9. [Rollback Procedure](#rollback-procedure)
10. [Post-Release](#post-release)
11. [Troubleshooting](#troubleshooting)

## Prerequisites & Initial Setup

Before you can perform releases, you need to:

1. **GCP Account**: Access to a Google Cloud Platform project with billing enabled
2. **GitHub Access**: Admin access to the repository to configure secrets
3. **gcloud CLI**: Installed and authenticated on your local machine (for manual operations)
4. **Docker**: Installed locally (for testing builds)
5. **Go 1.25+**: For local development and testing

### Tools Installation

```bash
# Install gcloud CLI (macOS)
brew install --cask google-cloud-sdk

# Install gcloud CLI (Linux)
curl https://sdk.cloud.google.com | bash
exec -l $SHELL

# Authenticate with GCP
gcloud auth login

# Install Docker (macOS)
brew install docker

# Install Go (macOS)
brew install go@1.25
```

## GCP Configuration

The CI/CD pipeline deploys to Google Cloud Platform (GCP) using Artifact Registry and Cloud Run. Follow these steps to configure GCP for the first time.

### 1. Create or Select GCP Project

```bash
# Create a new project (or use existing one)
gcloud projects create YOUR-PROJECT-ID --name="Kotomi Production"

# Set as default project
gcloud config set project YOUR-PROJECT-ID

# Enable billing (required for Cloud Run and Artifact Registry)
# Visit: https://console.cloud.google.com/billing
```

### 2. Enable Required APIs

```bash
# Enable Artifact Registry API
gcloud services enable artifactregistry.googleapis.com

# Enable Cloud Run API
gcloud services enable run.googleapis.com

# Enable Cloud Build API (optional, for additional CI/CD features)
gcloud services enable cloudbuild.googleapis.com

# Verify enabled services
gcloud services list --enabled
```

### 3. Create Artifact Registry Repository

Artifact Registry stores Docker images for deployment.

```bash
# Set your region (choose closest to your users)
REGION="us-central1"  # or us-east1, europe-west1, asia-northeast1, etc.

# Create Docker repository
gcloud artifacts repositories create kotomi \
  --repository-format=docker \
  --location=$REGION \
  --description="Kotomi Docker images"

# Verify repository creation
gcloud artifacts repositories list --location=$REGION

# Configure Docker authentication
gcloud auth configure-docker ${REGION}-docker.pkg.dev
```

**Repository URL format**: `[REGION]-docker.pkg.dev/[PROJECT-ID]/kotomi/kotomi:[VERSION]`

Example: `us-central1-docker.pkg.dev/my-project/kotomi/kotomi:0.1.0-beta.1`

### 4. Create Service Account for CI/CD

Create a service account that GitHub Actions will use to deploy:

```bash
# Create service account
gcloud iam service-accounts create github-actions-deployer \
  --display-name="GitHub Actions Deployer" \
  --description="Service account for GitHub Actions CI/CD pipeline"

# Get service account email
SA_EMAIL=$(gcloud iam service-accounts list \
  --filter="displayName:GitHub Actions Deployer" \
  --format='value(email)')

echo "Service Account Email: $SA_EMAIL"
```

### 5. Grant Required Permissions

The service account needs permissions to push images and deploy to Cloud Run:

```bash
# Grant Artifact Registry Writer (to push Docker images)
gcloud projects add-iam-policy-binding YOUR-PROJECT-ID \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/artifactregistry.writer"

# Grant Cloud Run Admin (to deploy services)
gcloud projects add-iam-policy-binding YOUR-PROJECT-ID \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/run.admin"

# Grant Service Account User (required for Cloud Run deployment)
gcloud projects add-iam-policy-binding YOUR-PROJECT-ID \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/iam.serviceAccountUser"

# Grant Storage Admin (for accessing artifacts during deployment)
gcloud projects add-iam-policy-binding YOUR-PROJECT-ID \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/storage.admin"

# Verify permissions
gcloud projects get-iam-policy YOUR-PROJECT-ID \
  --flatten="bindings[].members" \
  --filter="bindings.members:serviceAccount:${SA_EMAIL}"
```

**Minimum Required Roles**:
- `roles/artifactregistry.writer` - Push Docker images
- `roles/run.admin` - Create and update Cloud Run services
- `roles/iam.serviceAccountUser` - Deploy as service account
- `roles/storage.admin` - Access build artifacts

### 6. Create Service Account Key

Generate a JSON key file for the service account (needed for GitHub Actions):

```bash
# Create and download key
gcloud iam service-accounts keys create ~/gcp-key.json \
  --iam-account=${SA_EMAIL}

# Display key content (you'll need this for GitHub Secrets)
cat ~/gcp-key.json

# IMPORTANT: Keep this key secure and delete the local copy after adding to GitHub
# Store the JSON content in GitHub Secrets (see next section)
```

⚠️ **Security Note**: The service account key provides full access to your GCP resources. Never commit it to version control, share it publicly, or leave it on your local machine. Delete it after adding to GitHub Secrets:

```bash
# After adding to GitHub Secrets, delete local copy
rm ~/gcp-key.json
```

### 7. Configure Cloud Run Service (Optional Pre-Setup)

You can create the Cloud Run service manually before the first deployment, or let the CI/CD pipeline create it automatically.

```bash
# Set region
REGION="us-central1"

# Create initial Cloud Run service (optional)
gcloud run services create kotomi-prod \
  --image=gcr.io/cloudrun/placeholder \
  --region=$REGION \
  --platform=managed \
  --allow-unauthenticated \
  --memory=512Mi \
  --cpu=1 \
  --min-instances=0 \
  --max-instances=10

# Get service URL
gcloud run services describe kotomi-prod \
  --region=$REGION \
  --format='value(status.url)'
```

### 8. Set Environment Variables (Cloud Run)

Configure required environment variables for the Cloud Run service:

```bash
# Set environment variables
gcloud run services update kotomi-prod \
  --region=$REGION \
  --set-env-vars="AUTH0_DOMAIN=YOUR_AUTH0_DOMAIN" \
  --set-env-vars="AUTH0_CLIENT_ID=YOUR_AUTH0_CLIENT_ID" \
  --set-env-vars="AUTH0_CLIENT_SECRET=YOUR_AUTH0_CLIENT_SECRET" \
  --set-env-vars="SESSION_SECRET=YOUR_SESSION_SECRET" \
  --set-env-vars="OPENAI_API_KEY=YOUR_OPENAI_KEY" \
  --set-env-vars="SMTP_HOST=smtp.gmail.com" \
  --set-env-vars="SMTP_PORT=587" \
  --set-env-vars="SMTP_USER=YOUR_EMAIL" \
  --set-env-vars="SMTP_PASSWORD=YOUR_APP_PASSWORD"
```

Or update via Cloud Console:
1. Go to https://console.cloud.google.com/run
2. Click on `kotomi-prod` service
3. Click "EDIT & DEPLOY NEW REVISION"
4. Scroll to "Container, Variables & Secrets"
5. Add environment variables

## GitHub Secrets Configuration

GitHub Actions workflow requires several secrets to deploy to GCP. These secrets must be configured in your GitHub repository settings.

### Accessing GitHub Secrets

1. Go to your repository: `https://github.com/saasuke-labs/kotomi`
2. Click **Settings** → **Secrets and variables** → **Actions**
3. Click **New repository secret** for each secret below

### Required Secrets

#### 1. `GCP_SA_KEY`

**Purpose**: Service account JSON key for authenticating with GCP

**Value**: The entire JSON content from the service account key file created in GCP Configuration step 6.

**How to get it**:
```bash
# Display the key content (if still available)
cat ~/gcp-key.json
```

**Format**: JSON object starting with `{` and ending with `}`

```json
{
  "type": "service_account",
  "project_id": "your-project-id",
  "private_key_id": "...",
  "private_key": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n",
  "client_email": "github-actions-deployer@your-project.iam.gserviceaccount.com",
  "client_id": "...",
  "auth_uri": "https://accounts.google.com/o/oauth2/auth",
  "token_uri": "https://oauth2.googleapis.com/token",
  "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
  "client_x509_cert_url": "..."
}
```

**Steps to add**:
1. In GitHub Secrets, click **New repository secret**
2. Name: `GCP_SA_KEY`
3. Value: Paste the entire JSON content
4. Click **Add secret**

#### 2. `GCP_PROJECT_ID`

**Purpose**: Your GCP project ID where Kotomi will be deployed

**Value**: Your GCP project ID (e.g., `kotomi-production-12345`)

**How to get it**:
```bash
# Get current project ID
gcloud config get-value project

# Or list all projects
gcloud projects list
```

**Steps to add**:
1. Click **New repository secret**
2. Name: `GCP_PROJECT_ID`
3. Value: Your project ID (e.g., `my-kotomi-project`)
4. Click **Add secret**

#### 3. `GCP_REGION`

**Purpose**: GCP region where Cloud Run service and Artifact Registry are located

**Value**: GCP region code (e.g., `us-central1`, `us-east1`, `europe-west1`)

**Common regions**:
- `us-central1` - Iowa, USA
- `us-east1` - South Carolina, USA
- `us-west1` - Oregon, USA
- `europe-west1` - Belgium, Europe
- `asia-northeast1` - Tokyo, Asia

**How to verify**:
```bash
# List available regions
gcloud run regions list

# Verify your Artifact Registry location
gcloud artifacts repositories list
```

**Steps to add**:
1. Click **New repository secret**
2. Name: `GCP_REGION`
3. Value: Your region (e.g., `us-central1`)
4. Click **Add secret**

### Verification

After adding all secrets, verify they're configured:

1. Go to **Settings** → **Secrets and variables** → **Actions**
2. You should see three repository secrets:
   - `GCP_SA_KEY`
   - `GCP_PROJECT_ID`
   - `GCP_REGION`

### Testing the Configuration

Test that secrets are correctly configured by triggering a deployment:

```bash
# Push to main branch (triggers deployment)
git checkout main
git pull origin main

# Make a small change to trigger workflow
echo "# Test deployment" >> README.md
git add README.md
git commit -m "Test: Verify CI/CD pipeline configuration"
git push origin main
```

Then monitor the GitHub Actions workflow:
1. Go to **Actions** tab in GitHub
2. Click on the latest "Build and Deploy Kotomi" workflow
3. Monitor each step for errors

If the deployment fails, check the [Troubleshooting](#troubleshooting) section below.

## Versioning Strategy

Kotomi follows [Semantic Versioning 2.0.0](https://semver.org/):

**Format**: `MAJOR.MINOR.PATCH[-PRERELEASE]`

### Version Components

- **MAJOR**: Breaking changes, incompatible API changes
- **MINOR**: New features, backwards-compatible
- **PATCH**: Bug fixes, backwards-compatible
- **PRERELEASE**: alpha, beta, rc (release candidate)

### Examples

- `0.0.1` - Initial development version
- `0.1.0-beta.1` - First beta release
- `0.1.0-beta.2` - Second beta release (bug fixes)
- `0.1.0` - First stable minor release
- `0.1.1` - Patch release (bug fixes)
- `0.2.0` - Second minor release (new features)
- `1.0.0` - First major release (production ready)

### Current Version

Current: `0.0.1`
Next: `0.1.0-beta.1` (Phase 1 ADR 004 completion)

## Release Types

### Alpha Release (Internal)

- **Purpose**: Internal testing
- **Audience**: Development team only
- **Stability**: Unstable, breaking changes expected
- **Versioning**: `0.x.0-alpha.N`

### Beta Release (External)

- **Purpose**: Real-world testing with limited users
- **Audience**: Selected beta testers
- **Stability**: Mostly stable, some bugs expected
- **Versioning**: `0.x.0-beta.N`
- **Support**: Limited, best-effort

### Release Candidate (Pre-Production)

- **Purpose**: Final validation before production
- **Audience**: Broader user base
- **Stability**: Stable, only critical bugs
- **Versioning**: `x.x.x-rc.N`
- **Support**: Full support

### Stable Release (Production)

- **Purpose**: Production use
- **Audience**: All users
- **Stability**: Stable, fully tested
- **Versioning**: `x.x.x`
- **Support**: Full support, SLA commitment

### Patch Release

- **Purpose**: Bug fixes, security patches
- **Audience**: All users on that major.minor
- **Stability**: Critical fixes only
- **Versioning**: `x.x.N` (increment patch)
- **Support**: Urgent support

## Release Checklist

Use this checklist for each release. Copy and customize for specific releases.

### Pre-Release Phase

#### Code Quality
- [ ] All tests passing (unit, integration, E2E)
- [ ] Code coverage maintained (>80%)
- [ ] No critical linting issues
- [ ] Security scan completed (gosec, dependency audit)
- [ ] Code review completed for all changes
- [ ] No known critical bugs

#### Documentation
- [ ] CHANGELOG.md updated
- [ ] API documentation updated (if API changes)
- [ ] README.md updated (if needed)
- [ ] Migration guide written (if breaking changes)
- [ ] Release notes drafted
- [ ] User-facing docs updated

#### Version Management
- [ ] VERSION file updated
- [ ] Git tag created
- [ ] Release branch created (for major/minor)
- [ ] Release notes finalized

#### Security
- [ ] Dependency vulnerabilities checked
- [ ] Security scan passed
- [ ] Secrets rotation completed (if needed)
- [ ] Security review completed (if significant changes)

### Build & Test Phase

#### Build
- [ ] Docker image built successfully
- [ ] Image tagged with version number
- [ ] Image pushed to artifact registry
- [ ] Build artifacts archived

#### Testing
- [ ] Smoke tests passed
- [ ] Integration tests passed
- [ ] E2E tests passed
- [ ] Performance tests passed (if applicable)
- [ ] Security tests passed

#### Staging Deployment
- [ ] Deployed to staging environment
- [ ] Staging smoke tests passed
- [ ] Manual testing completed
- [ ] Performance validated
- [ ] Security validated

### Release Phase

#### Deployment
- [ ] Deployment plan reviewed
- [ ] Rollback plan documented
- [ ] On-call team notified
- [ ] Deployment window scheduled
- [ ] Production deployment executed
- [ ] Health checks passing
- [ ] Logs reviewed (no errors)

#### Validation
- [ ] API endpoints responsive
- [ ] Admin panel accessible
- [ ] Authentication working
- [ ] Database connectivity confirmed
- [ ] Critical user paths tested
- [ ] No error spikes in logs

#### Communication
- [ ] Release notes published
- [ ] Users notified (if needed)
- [ ] Documentation published
- [ ] Status page updated
- [ ] Team notified

### Post-Release Phase

#### Monitoring
- [ ] Error rates monitored (first hour)
- [ ] Performance metrics reviewed
- [ ] User feedback collected
- [ ] Bug reports triaged
- [ ] Metrics dashboard reviewed

#### Follow-up
- [ ] Post-release retrospective scheduled
- [ ] Known issues documented
- [ ] Patch release planned (if needed)
- [ ] Next release planning started

## Version Bumping

### Update VERSION File

```bash
# Navigate to repository
cd /path/to/kotomi

# Update VERSION file
echo "0.1.0-beta.1" > VERSION

# Commit the change
git add VERSION
git commit -m "Bump version to 0.1.0-beta.1"
```

### Create Git Tag

```bash
# Create annotated tag
git tag -a v0.1.0-beta.1 -m "Release version 0.1.0-beta.1"

# Push tag to remote
git push origin v0.1.0-beta.1

# Verify tag
git tag -l "v0.1.0*"
```

### Update CHANGELOG

Add release section to `CHANGELOG.md`:

```markdown
## [0.1.0-beta.1] - 2026-02-04

### Added
- Beta tester guide
- Admin panel documentation
- Database backup procedures
- Automated deployment to Cloud Run

### Changed
- Updated golang.org/x/oauth2 to v0.27.0 (security fix)

### Fixed
- Security vulnerability in OAuth2 library

[0.1.0-beta.1]: https://github.com/saasuke-labs/kotomi/compare/v0.0.1...v0.1.0-beta.1
```

### Update README Badges (Optional)

```markdown
![Version](https://img.shields.io/badge/version-0.1.0--beta.1-blue)
![Status](https://img.shields.io/badge/status-beta-yellow)
```

## Deployment Process

### Automated Deployment (CI/CD)

The GitHub Actions workflow automatically:
1. Runs all tests
2. Builds Docker image
3. Pushes to artifact registry
4. Deploys to Cloud Run (when enabled)

### Manual Deployment Steps

If deploying manually:

#### 1. Build Docker Image

```bash
# Read version
VERSION=$(cat VERSION)

# Build image
docker build -t gcr.io/YOUR_PROJECT/kotomi:$VERSION \
  --build-arg VERSION=$VERSION \
  --build-arg COMMIT=$(git rev-parse --short HEAD) \
  --build-arg BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
  .

# Test image locally
docker run --rm -p 8080:8080 \
  -e AUTH0_DOMAIN=test.auth0.com \
  -e AUTH0_CLIENT_ID=test \
  -e AUTH0_CLIENT_SECRET=test \
  -e SESSION_SECRET=test-secret-min-32-chars-long \
  gcr.io/YOUR_PROJECT/kotomi:$VERSION
```

#### 2. Push to Registry

```bash
# Push image
docker push gcr.io/YOUR_PROJECT/kotomi:$VERSION

# Tag as latest (optional)
docker tag gcr.io/YOUR_PROJECT/kotomi:$VERSION gcr.io/YOUR_PROJECT/kotomi:latest
docker push gcr.io/YOUR_PROJECT/kotomi:latest
```

#### 3. Deploy to Cloud Run

```bash
# Deploy to Cloud Run
gcloud run deploy kotomi-prod \
  --image gcr.io/YOUR_PROJECT/kotomi:$VERSION \
  --region us-central1 \
  --platform managed \
  --allow-unauthenticated \
  --memory 512Mi \
  --cpu 1 \
  --min-instances 0 \
  --max-instances 10

# Get service URL
gcloud run services describe kotomi-prod \
  --region us-central1 \
  --format 'value(status.url)'
```

#### 4. Smoke Test

```bash
# Get service URL
SERVICE_URL=$(gcloud run services describe kotomi-prod --region us-central1 --format 'value(status.url)')

# Test health check
curl $SERVICE_URL/healthz

# Test API documentation
curl $SERVICE_URL/swagger/index.html

# Test admin panel (in browser)
open $SERVICE_URL/admin/dashboard
```

## Rollback Procedure

### When to Rollback

Rollback immediately if:
- Critical bugs affecting all users
- Security vulnerabilities discovered
- Data corruption issues
- Service completely down
- Error rate >5% within first hour

### Quick Rollback (Cloud Run)

Cloud Run makes rollback easy:

```bash
# List revisions
gcloud run revisions list \
  --service kotomi-prod \
  --region us-central1

# Rollback to previous revision
PREVIOUS_REVISION="kotomi-prod-00042-abc"  # Get from list above

gcloud run services update-traffic kotomi-prod \
  --to-revisions=$PREVIOUS_REVISION=100 \
  --region us-central1

# Verify rollback
gcloud run services describe kotomi-prod \
  --region us-central1
```

### Rollback with Database Restore

If database needs restoration:

```bash
# 1. Stop application (set to 0 instances temporarily)
gcloud run services update kotomi-prod \
  --min-instances=0 \
  --max-instances=0 \
  --region us-central1

# 2. Restore database (see DATABASE_BACKUP_RESTORE.md)

# 3. Deploy previous version
gcloud run deploy kotomi-prod \
  --image gcr.io/YOUR_PROJECT/kotomi:PREVIOUS_VERSION \
  --region us-central1

# 4. Resume traffic
gcloud run services update kotomi-prod \
  --min-instances=0 \
  --max-instances=10 \
  --region us-central1

# 5. Verify
curl $(gcloud run services describe kotomi-prod --region us-central1 --format 'value(status.url)')/healthz
```

### Rollback Communication

After rollback:
1. Update status page
2. Notify users via email/announcement
3. Document issue and resolution
4. Plan fix and re-deployment

## Post-Release

### Monitoring (First 24 Hours)

Monitor these metrics closely:

- **Error Rate**: Should be <1%
- **Response Time**: Should be <500ms p95
- **Availability**: Should be >99.5%
- **CPU/Memory**: Should be stable
- **Database**: No lock errors

### Gathering Feedback

Collect feedback from:
- Error logs
- User reports (GitHub issues)
- Beta tester feedback
- Monitoring dashboards
- Support channels

### Release Retrospective

Within 1 week, conduct retrospective:

**Topics to discuss**:
- What went well?
- What could be improved?
- Were there surprises?
- How can we prevent issues?
- Process improvements?

**Document**:
- Lessons learned
- Process changes
- Tool improvements needed

### Planning Next Release

Based on feedback:
1. Triage reported issues
2. Prioritize bug fixes
3. Plan next features
4. Schedule next release
5. Update roadmap

## Emergency Hotfix Process

For critical security or stability issues:

### Hotfix Workflow

1. **Create hotfix branch** from release tag
   ```bash
   git checkout -b hotfix/0.1.0-beta.2 v0.1.0-beta.1
   ```

2. **Make minimal fix** addressing only the critical issue

3. **Test thoroughly** on the hotfix branch

4. **Version bump** (increment patch or prerelease number)
   ```bash
   echo "0.1.0-beta.2" > VERSION
   ```

5. **Deploy to staging** and validate

6. **Deploy to production** following rollout procedure

7. **Tag and release**
   ```bash
   git tag -a v0.1.0-beta.2 -m "Hotfix: Fix critical issue"
   git push origin v0.1.0-beta.2
   ```

8. **Merge back** to main branch
   ```bash
   git checkout main
   git merge hotfix/0.1.0-beta.2
   git push origin main
   ```

9. **Communicate** to users about the fix

### Hotfix Approval

Hotfixes require:
- Minimal code changes
- Clear description of issue
- Test coverage for the fix
- Approval from team lead
- Rollback plan documented

## Automation Opportunities

Future improvements to release process:

- [ ] Automated version bumping
- [ ] Automated CHANGELOG generation
- [ ] Automated release notes
- [ ] Canary deployments
- [ ] Blue-green deployments
- [ ] Automated rollback on errors
- [ ] Release approval workflow
- [ ] Automated smoke tests post-deploy

## Troubleshooting

### Common Issues and Solutions

#### 1. Authentication Failed to Artifact Registry

**Error**: `denied: Permission "artifactregistry.repositories.uploadArtifacts" denied`

**Cause**: Service account doesn't have permission to push images

**Solution**:
```bash
# Grant Artifact Registry Writer role
gcloud projects add-iam-policy-binding YOUR-PROJECT-ID \
  --member="serviceAccount:SERVICE_ACCOUNT_EMAIL" \
  --role="roles/artifactregistry.writer"
```

#### 2. Cloud Run Deployment Permission Denied

**Error**: `ERROR: (gcloud.run.deploy) PERMISSION_DENIED: Permission 'run.services.create' denied`

**Cause**: Service account lacks Cloud Run permissions

**Solution**:
```bash
# Grant Cloud Run Admin role
gcloud projects add-iam-policy-binding YOUR-PROJECT-ID \
  --member="serviceAccount:SERVICE_ACCOUNT_EMAIL" \
  --role="roles/run.admin"

# Grant Service Account User role
gcloud projects add-iam-policy-binding YOUR-PROJECT-ID \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/iam.serviceAccountUser"
```

#### 3. Docker Login Failed

**Error**: `Error response from daemon: Get "https://[REGION]-docker.pkg.dev/v2/": denied: Permission "artifactregistry.repositories.get" denied`

**Cause**: Incorrect authentication or repository doesn't exist

**Solution**:
```bash
# Verify repository exists
gcloud artifacts repositories list --location=REGION

# Re-authenticate Docker
gcloud auth configure-docker REGION-docker.pkg.dev

# In GitHub Actions, ensure credentials_json secret is correct
```

#### 4. Invalid Service Account Key

**Error**: `Error: google-github-actions/auth failed with: retry function failed after 3 attempts`

**Cause**: GCP_SA_KEY secret is invalid or malformed

**Solution**:
1. Verify the JSON is valid (starts with `{`, ends with `}`)
2. Ensure no extra whitespace or line breaks
3. Regenerate the key if necessary:
```bash
gcloud iam service-accounts keys create new-key.json \
  --iam-account=SERVICE_ACCOUNT_EMAIL
```
4. Update GitHub Secret with new key content

#### 5. Artifact Registry Repository Not Found

**Error**: `ERROR: (gcloud.artifacts.docker.images) NOT_FOUND: Repository not found`

**Cause**: Artifact Registry repository doesn't exist in specified region

**Solution**:
```bash
# Create the repository
gcloud artifacts repositories create kotomi \
  --repository-format=docker \
  --location=REGION \
  --description="Kotomi Docker images"
```

#### 6. Wrong Region Configuration

**Error**: Images push to one region but Cloud Run looks in another

**Cause**: Mismatch between GCP_REGION secret and actual resource locations

**Solution**:
```bash
# Check where resources are located
gcloud artifacts repositories list
gcloud run services list

# Ensure GCP_REGION GitHub secret matches both locations
# Or migrate resources to same region
```

#### 7. Billing Not Enabled

**Error**: `Cloud Run API has not been used in project before or it is disabled`

**Cause**: Project doesn't have billing enabled

**Solution**:
1. Go to https://console.cloud.google.com/billing
2. Link a billing account to your project
3. Enable required APIs again

#### 8. Build Timeout in GitHub Actions

**Error**: Build step times out after 6 hours

**Cause**: Network issues or resource constraints

**Solution**:
```yaml
# Add timeout to workflow step
- name: Build and push Docker image
  timeout-minutes: 30
  run: |
    docker build ...
```

#### 9. Service Account Key Exposed

**Problem**: Accidentally committed service account key to git

**Solution**:
```bash
# 1. IMMEDIATELY delete the key in GCP
gcloud iam service-accounts keys list \
  --iam-account=SERVICE_ACCOUNT_EMAIL

gcloud iam service-accounts keys delete KEY_ID \
  --iam-account=SERVICE_ACCOUNT_EMAIL

# 2. Create new key
gcloud iam service-accounts keys create new-key.json \
  --iam-account=SERVICE_ACCOUNT_EMAIL

# 3. Update GitHub Secret

# 4. Remove from git history (contact GitHub support if public)
git filter-branch --force --index-filter \
  "git rm --cached --ignore-unmatch path/to/key.json" \
  --prune-empty --tag-name-filter cat -- --all

# 5. Force push (if private repo only)
git push origin --force --all
```

#### 10. Environment Variables Not Set

**Error**: Application fails to start with authentication errors

**Cause**: Cloud Run service missing required environment variables

**Solution**:
```bash
# Update environment variables
gcloud run services update kotomi-prod \
  --region=REGION \
  --update-env-vars="AUTH0_DOMAIN=YOUR_DOMAIN,AUTH0_CLIENT_ID=YOUR_ID"

# Or edit in Console:
# https://console.cloud.google.com/run → Select service → Edit → Variables
```

### Getting Help

If you encounter issues not covered here:

1. **Check GitHub Actions logs**: Detailed error messages in Actions tab
2. **Check Cloud Run logs**: 
   ```bash
   gcloud run services logs read kotomi-prod --region=REGION --limit=50
   ```
3. **Verify GCP resources**:
   ```bash
   gcloud artifacts repositories list
   gcloud run services list
   gcloud iam service-accounts list
   ```
4. **Test locally**:
   ```bash
   # Build and test Docker image locally
   docker build -t kotomi:test .
   docker run -p 8080:8080 kotomi:test
   ```
5. **Check GCP quotas**: https://console.cloud.google.com/iam-admin/quotas
6. **Review IAM permissions**: https://console.cloud.google.com/iam-admin/iam

### Useful Commands Reference

```bash
# Authenticate gcloud
gcloud auth login
gcloud config set project PROJECT_ID

# List all resources
gcloud artifacts repositories list
gcloud run services list
gcloud iam service-accounts list

# Check service account permissions
gcloud projects get-iam-policy PROJECT_ID \
  --flatten="bindings[].members" \
  --filter="bindings.members:serviceAccount:EMAIL"

# Describe Cloud Run service
gcloud run services describe kotomi-prod --region=REGION

# View Cloud Run logs
gcloud run services logs read kotomi-prod --region=REGION

# Test Docker image locally
docker build -t kotomi:test .
docker run --rm -p 8080:8080 kotomi:test

# Manual deploy (for testing)
gcloud run deploy kotomi-prod \
  --image=REGION-docker.pkg.dev/PROJECT/kotomi/kotomi:VERSION \
  --region=REGION
```

---

**Version**: 0.1.0-beta.1
**Last Updated**: February 4, 2026
