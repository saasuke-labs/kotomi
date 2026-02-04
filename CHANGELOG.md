# Changelog

All notable changes to Kotomi will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
