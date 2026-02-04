# Beta Iteration & Patch Release Process

This document outlines the process for releasing patches and iterating on Kotomi during the beta release period.

## Overview

During beta, we expect to release frequent patches to address bugs and incorporate feedback. This process ensures smooth, reliable releases while maintaining beta tester confidence.

## Release Cadence

### Patch Releases (0.1.0-beta.X)
**Purpose**: Bug fixes, security patches, documentation updates  
**Frequency**: As needed, typically weekly or when critical bugs are found  
**Version Format**: `0.1.0-beta.2`, `0.1.0-beta.3`, etc.

**When to Release**:
- Critical bug affecting beta testers
- Security vulnerability discovered
- Multiple high-priority bugs accumulated
- Documentation significantly improved
- Performance regression fixed

**When NOT to Release**:
- Single low-priority bug
- Feature request (wait for next minor version)
- Cosmetic/styling changes only
- Internal refactoring without user impact

### Minor Releases (0.X.0-beta.1)
**Purpose**: New features, breaking changes, major improvements  
**Frequency**: Monthly or when significant features complete  
**Version Format**: `0.2.0-beta.1`, `0.3.0-beta.1`, etc.

## Patch Release Process

### 1. Planning (1-2 hours)

#### Identify Issues to Address
- [ ] Review GitHub Issues with `priority: high` or `priority: critical` labels
- [ ] Check beta tester feedback from weekly check-ins
- [ ] Review monitoring alerts and error logs
- [ ] List all candidate fixes for this release

#### Prioritize Fixes
- **P0 - Critical**: Blocking bugs, data loss, security vulnerabilities
- **P1 - High**: Major features broken, widespread issues
- **P2 - Medium**: Important but has workaround
- **P3 - Low**: Minor issues, nice-to-haves

#### Create Release Issue
Create GitHub Issue: `[RELEASE] 0.1.0-beta.X`

```markdown
## Release: 0.1.0-beta.X

**Target Date**: [Date]
**Release Manager**: [Name]

### Included Fixes
- [ ] #123 - Fix JWT validation for ECDSA tokens (P1)
- [ ] #124 - Resolve comment deletion bug (P1)
- [ ] #125 - Update documentation for Cloud Run deployment (P2)

### Pre-Release Checklist
- [ ] All fixes merged to main
- [ ] Tests passing
- [ ] Documentation updated
- [ ] CHANGELOG updated
- [ ] Version bumped

### Release Checklist
- [ ] Tag created
- [ ] Docker image built and pushed
- [ ] Cloud Run deployed
- [ ] Smoke tests passed
- [ ] Beta testers notified

### Post-Release
- [ ] Monitor for errors (24 hours)
- [ ] Collect initial feedback
- [ ] Update release issue with results
```

### 2. Development (1-3 days)

#### Create Fix Branches
```bash
# For each fix
git checkout main
git pull origin main
git checkout -b fix/issue-123-jwt-validation

# Make changes
# ... edit files ...

# Test locally
go test ./...
go run cmd/main.go
# Run smoke tests
./scripts/smoke_test.sh http://localhost:8080

# Commit
git add .
git commit -m "Fix JWT validation for ECDSA tokens (#123)"

# Push and create PR
git push origin fix/issue-123-jwt-validation
```

#### Code Review
- [ ] Create pull request
- [ ] Link to original issue
- [ ] Describe fix and testing performed
- [ ] Request review from team member
- [ ] Address review comments
- [ ] Merge when approved

#### Testing
- [ ] All unit tests pass: `go test ./...`
- [ ] Integration tests pass (if applicable)
- [ ] E2E tests pass (if applicable)
- [ ] Manual testing of specific fixes
- [ ] Smoke tests pass locally

### 3. Pre-Release Preparation (1-2 hours)

#### Update VERSION File
```bash
# Example: Bump from 0.1.0-beta.1 to 0.1.0-beta.2
echo "0.1.0-beta.2" > VERSION
git add VERSION
```

#### Update CHANGELOG
Add new section at the top:

```markdown
## [0.1.0-beta.2] - 2026-02-10

### Fixed
- JWT validation for ECDSA tokens now works correctly (#123)
- Comment deletion no longer fails with foreign key error (#124)
- Admin panel loading spinner now displays correctly (#126)

### Changed
- Improved error messages for authentication failures
- Updated Cloud Run deployment documentation

### Security
- Fixed potential XSS vulnerability in comment rendering (#127)
```

#### Run Full Test Suite
```bash
# Unit tests
go test ./... -v -race -coverprofile=coverage.out

# Integration tests
go test ./pkg/comments/ -run Integration -v

# E2E tests (if available)
RUN_E2E_TESTS=true go test ./tests/e2e/... -v

# Smoke tests
go run cmd/main.go &
sleep 5
./scripts/smoke_test.sh http://localhost:8080
kill %1
```

#### Security Scan
```bash
# Run CodeQL or security scanner
# This should happen automatically in CI/CD
# But good to run manually for critical patches
```

#### Commit and Push
```bash
git add VERSION CHANGELOG.md
git commit -m "Bump version to 0.1.0-beta.2"
git push origin main
```

### 4. Release Execution (30 minutes)

#### Create Git Tag
```bash
# Tag the release
git tag -a v0.1.0-beta.2 -m "Release version 0.1.0-beta.2

Fixes:
- JWT validation for ECDSA tokens (#123)
- Comment deletion bug (#124)
- Admin panel loading spinner (#126)"

# Push tag
git push origin v0.1.0-beta.2
```

#### Verify CI/CD Pipeline
- [ ] GitHub Actions workflow triggered
- [ ] All tests passing in CI
- [ ] Docker image built successfully
- [ ] Image pushed to artifact registry
- [ ] Cloud Run deployment successful
- [ ] Health check passing

**Monitor at**: https://github.com/saasuke-labs/kotomi/actions

#### Manual Verification (If CI/CD Not Automated)
```bash
# Build Docker image
docker build -t kotomi:0.1.0-beta.2 .

# Tag for artifact registry
docker tag kotomi:0.1.0-beta.2 \
  us-central1-docker.pkg.dev/your-project/kotomi/kotomi:0.1.0-beta.2

# Push to registry
docker push us-central1-docker.pkg.dev/your-project/kotomi/kotomi:0.1.0-beta.2

# Deploy to Cloud Run
gcloud run deploy kotomi-prod \
  --image us-central1-docker.pkg.dev/your-project/kotomi/kotomi:0.1.0-beta.2 \
  --region us-central1 \
  --platform managed
```

#### Run Smoke Tests on Production
```bash
# Run against production deployment
./scripts/smoke_test.sh https://kotomi-prod-xyz.run.app

# Verify all tests pass
```

### 5. Post-Release Activities (1 day)

#### Create GitHub Release
1. Go to: https://github.com/saasuke-labs/kotomi/releases/new
2. Select tag: `v0.1.0-beta.2`
3. Release title: `0.1.0-beta.2 - Beta Patch Release`
4. Description: Copy from CHANGELOG
5. Mark as "pre-release" (it's beta)
6. Publish release

#### Notify Beta Testers
Send email to all active beta testers:

```
Subject: Kotomi 0.1.0-beta.2 Released

Hi [Beta Tester],

We've released Kotomi 0.1.0-beta.2 with several important fixes:

Highlights:
- Fixed JWT validation for ECDSA tokens (#123)
- Resolved comment deletion bug (#124)
- Improved admin panel UX (#126)

Full changelog: https://github.com/saasuke-labs/kotomi/blob/main/CHANGELOG.md

If you're using Docker:
  docker pull gcr.io/your-project/kotomi:0.1.0-beta.2
  docker-compose down && docker-compose up -d

If you're using Cloud Run:
  The deployment is automatic - already live!

Please let us know if you encounter any issues.

Thanks,
[Your Name]
Kotomi Team
```

#### Monitor Deployment
- [ ] Check error logs (first 1-2 hours)
- [ ] Monitor API response times
- [ ] Verify no spike in errors
- [ ] Check beta tester reports
- [ ] Review GitHub Issues for new bugs

#### Update Documentation
- [ ] Close release issue
- [ ] Update any affected documentation
- [ ] Mark fixed issues as closed
- [ ] Thank contributors

#### Team Debrief (within 48 hours)
- What went well?
- What could be improved?
- Any process changes needed?
- Lessons learned for next release

## Emergency Hotfix Process

### When to Use Emergency Hotfix
- **Production down**: Service completely unavailable
- **Data loss**: Active data corruption occurring
- **Security breach**: Actively exploited vulnerability
- **Critical bug**: Blocking all beta testers immediately

### Fast-Track Process (2-4 hours)

1. **Immediate Assessment** (15 minutes)
   - Confirm severity
   - Identify root cause
   - Estimate fix time
   - Decide: hotfix or rollback?

2. **Hotfix Development** (30-120 minutes)
   - Create hotfix branch: `hotfix/critical-issue-description`
   - Fix the issue (minimal changes only)
   - Test fix thoroughly
   - Skip normal review process (but have second pair of eyes)

3. **Emergency Release** (30 minutes)
   - Bump version: 0.1.0-beta.2 → 0.1.0-beta.3
   - Update CHANGELOG (brief entry)
   - Create tag immediately
   - Deploy to production
   - Run smoke tests

4. **Immediate Notification** (15 minutes)
   - Notify all beta testers via email
   - Post in communication channel
   - Update GitHub with incident report

5. **Post-Mortem** (within 24 hours)
   - Document what happened
   - Identify prevention measures
   - Update processes
   - Share learnings with team

### Rollback Procedure

If new release causes issues:

```bash
# 1. Identify previous working version
PREVIOUS_VERSION="0.1.0-beta.1"

# 2. Rollback Cloud Run
gcloud run deploy kotomi-prod \
  --image us-central1-docker.pkg.dev/your-project/kotomi/kotomi:$PREVIOUS_VERSION \
  --region us-central1 \
  --platform managed

# 3. Verify rollback
curl https://kotomi-prod-xyz.run.app/healthz

# 4. Notify beta testers
# Send email about rollback and investigation
```

## Version Numbering

### Beta Version Format
`MAJOR.MINOR.PATCH-beta.ITERATION`

Example: `0.1.0-beta.2`
- `0` - Major version (pre-1.0)
- `1` - Minor version (feature set)
- `0` - Patch version (always 0 for beta)
- `beta` - Pre-release identifier
- `2` - Beta iteration number

### When to Bump

**Beta Iteration** (0.1.0-beta.1 → 0.1.0-beta.2):
- Bug fixes
- Documentation updates
- Security patches
- No breaking changes

**Minor Version** (0.1.0-beta.X → 0.2.0-beta.1):
- New features added
- Breaking changes
- API changes
- Significant refactoring

**Major Version** (0.X.0 → 1.0.0):
- Production ready
- Stable API
- No longer beta

## Release Checklist Template

Use this for each release:

```markdown
## Release Checklist: 0.1.0-beta.X

### Pre-Release
- [ ] All planned fixes merged
- [ ] Tests passing (unit, integration, E2E)
- [ ] Documentation updated
- [ ] CHANGELOG updated
- [ ] VERSION bumped
- [ ] Security scan completed

### Release
- [ ] Git tag created
- [ ] CI/CD pipeline successful
- [ ] Docker image built and pushed
- [ ] Cloud Run deployment successful
- [ ] Production smoke tests passed

### Post-Release
- [ ] GitHub release created
- [ ] Beta testers notified
- [ ] Monitoring reviewed (first 2 hours)
- [ ] No critical errors detected
- [ ] Release issue closed
- [ ] Team debrief scheduled

### Rollback Plan
- [ ] Previous version ID recorded: ___________
- [ ] Rollback procedure ready
- [ ] Team alerted to rollback trigger
```

## Metrics to Track

### Release Health
- **Time to release**: Target < 1 week for patches
- **Deployment success rate**: Target 100%
- **Rollback rate**: Target < 5%
- **Post-release bugs**: Target < 2 per release

### Beta Impact
- **Beta tester satisfaction**: Survey after each release
- **Adoption rate**: % of testers on latest version
- **Feedback volume**: Issues/discussions per release
- **Active testers**: % still using Kotomi

## Tools & Resources

### Required Tools
- Git (version control)
- GitHub CLI or web interface (releases)
- gcloud CLI (Cloud Run deployment)
- Docker (image building)
- curl (smoke testing)

### Helpful Scripts
- `scripts/smoke_test.sh` - Basic smoke tests
- `scripts/smoke_test_authenticated.sh` - Full auth flow tests
- `scripts/release.sh` - Automated release script (TODO)

### Documentation
- [Release Process](RELEASE_PROCESS.md) - Detailed release guide
- [Beta Support Plan](BETA_SUPPORT_PLAN.md) - Support during releases
- [Deployment Monitoring](DEPLOYMENT_MONITORING.md) - Post-release monitoring
- [CHANGELOG](../CHANGELOG.md) - Release history

## FAQ

**Q: How often should we release patches?**  
A: As needed. For critical bugs, release immediately. Otherwise, batch fixes weekly.

**Q: Should every bug fix be a new release?**  
A: No. Batch non-critical fixes together. Release when you have 3-5 fixes or 1 high-priority fix.

**Q: What if a beta tester reports a bug after release?**  
A: Triage immediately. If critical, hotfix. Otherwise, add to next release.

**Q: Can we skip versions?**  
A: Yes, if you realize an issue before testers deploy. But be transparent about it.

**Q: Should we test on staging before production?**  
A: Ideally yes, but for small beta scale, production deployment with careful monitoring is acceptable.

**Q: How long to monitor after release?**  
A: Intensively for first 2-4 hours. Then daily for first week.

---

**Last Updated**: 2026-02-04  
**Version**: 1.0  
**Owner**: [Add release manager contact]
