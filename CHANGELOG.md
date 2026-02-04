# Changelog

All notable changes to Kotomi will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Smoke test scripts for deployment validation
  - Basic smoke test (`scripts/smoke_test.sh`) - validates health, APIs, admin panel, docs
  - Authenticated smoke test (`scripts/smoke_test_authenticated.sh`) - tests full JWT flow
- Comprehensive beta program documentation
  - Beta Tester Onboarding Checklist with day-by-day process
  - Beta Feedback Collection Guide with templates and processes
  - Beta Support Plan with SLAs and response times
  - Deployment Monitoring Guide with metrics and procedures
  - Beta Iteration & Patch Release Process documentation
- GitHub issue templates for structured feedback
  - Bug report template with impact assessment
  - Feature request template with priority levels
- Phase 2 implementation summary document

### Changed
- Updated documentation to use correct health endpoint (`/healthz` instead of `/health`)

## [0.1.0-beta.1] - 2026-02-04

### Added
- Comprehensive beta tester guide with deployment options (Docker, Cloud Run, binary)
- Admin panel guide with feature documentation and workflows
- Database backup and restore procedures
- Release process documentation with checklists and rollback procedures
- Production-ready Dockerfile with multi-stage build, health checks, and security hardening
- Automated Cloud Run deployment in CI/CD pipeline

### Changed
- Updated golang.org/x/oauth2 from v0.15.0 to v0.27.0 (security patch)
- Enhanced Dockerfile for production use (non-root user, health checks, optimized builds)
- Enabled Cloud Run deployment in GitHub Actions workflow

### Security
- Fixed vulnerability CVE in golang.org/x/oauth2 (Improper Validation of Syntactic Correctness)
- Added non-root user to Docker container
- Added health check to Docker container
- Comprehensive security documentation and audit results

## [0.0.1] - 2026-01-31

### Added
- JWT-based authentication system for public comments API
- Auth0 integration for admin panel authentication
- Comments API with full CRUD operations
- Reactions system (like/dislike for pages)
- Admin panel with HTMX for interactive UI
- AI-powered moderation with OpenAI integration
- Email notifications for new comments
- Rate limiting for API endpoints
- CORS middleware for cross-origin requests
- SQLite database with WAL mode optimization
- Comprehensive test coverage (>80%)
- Docker containerization
- CI/CD pipeline with GitHub Actions
- Swagger/OpenAPI API documentation
- Health check endpoint

### Security
- SQL injection prevention (parameterized queries)
- XSS protection (automatic HTML escaping)
- Slowloris attack protection (HTTP timeouts)
- Session encryption
- Owner-based access control

[Unreleased]: https://github.com/saasuke-labs/kotomi/compare/v0.0.1...HEAD
[0.0.1]: https://github.com/saasuke-labs/kotomi/releases/tag/v0.0.1
