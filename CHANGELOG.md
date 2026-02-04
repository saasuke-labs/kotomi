# Changelog

All notable changes to Kotomi will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial beta release preparation
- Comprehensive security documentation
- Beta tester guide and deployment documentation
- Release process and checklist

### Changed
- Updated golang.org/x/oauth2 from v0.15.0 to v0.27.0 (security patch)

### Security
- Fixed vulnerability in golang.org/x/oauth2 (Improper Validation of Syntactic Correctness)

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
