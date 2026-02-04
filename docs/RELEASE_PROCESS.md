# Release Process Documentation

This document describes the release process for Kotomi, including version management, release checklists, and rollback procedures.

## Table of Contents

1. [Versioning Strategy](#versioning-strategy)
2. [Release Types](#release-types)
3. [Release Checklist](#release-checklist)
4. [Version Bumping](#version-bumping)
5. [Deployment Process](#deployment-process)
6. [Rollback Procedure](#rollback-procedure)
7. [Post-Release](#post-release)

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

---

**Version**: 0.1.0-beta.1
**Last Updated**: February 4, 2026
