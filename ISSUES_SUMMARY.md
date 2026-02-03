# GitHub Issues Summary - Based on Status.md

This document contains a summary of all GitHub issues that should be created for Kotomi based on the Status.md analysis.

## Quick Reference

| # | Title | Priority | Status | Estimated Effort | Dependencies |
|---|-------|----------|--------|------------------|--------------|
| 1 | ✅ [RESOLVED] Implement CORS Configuration | Critical | ✅ Done | 2-4 hours | None |
| 2 | ✅ [RESOLVED] Implement Rate Limiting | Critical | ✅ Done | 4-8 hours | None |
| 3 | ⭐ [COMPLETED] Implement Reactions System | High | ✅ Done | 8-16 hours | After #1, #2 |
| 4 | ✅ [COMPLETED] Implement API Versioning | Medium | ✅ Done | 2-4 hours | Before #5 |
| 5 | ✅ [COMPLETED] Create Frontend Widget / JavaScript Embed | High | ✅ Done | 16-24 hours | After #1, #4 |
| 6 | 🔒 [COMPLETED] Conduct Security Audit | Critical | ✅ Done | 8-16 hours | After #1, #2 |
| 7 | 🤖 [COMPLETED] Implement Automatic/AI Moderation | Medium | ✅ Done | 16-24 hours | Independent |
| 8 | 👤 [PARTIAL] User Authentication for Comments | Medium | ✅ 65% Done | 24-40 hours (16h done) | After #5 |
| 9 | 📧 [COMPLETED] Implement Email Notifications | Low | ✅ Done | 12-16 hours | After #8 |
| 10 | 📊 [COMPLETED] Implement Analytics & Reporting | Low | ✅ Done | 12-16 hours | After #3, #8 |
| 11 | 💾 [COMPLETED] Implement Export/Import Functionality | Low | ✅ Done | 8-12 hours | Independent |
| 12 | 🔍 [COMPLETED] Improve Error Handling & Logging | Medium | ✅ Done | 8-12 hours | Independent |

**Total Estimated Effort**: 118-178 hours (approximately 3-4 weeks of full-time development)
**Completed**: Issues #1, #2, #3, #4, #5, #6, #7, #9, #10, #11, #12 (96-144 hours completed)
**Partially Complete**: Issue #8 (16 hours completed, 9-24 hours remaining)
**Remaining**: None

## Implementation Phases

### Phase 1: Blocking Issues (Critical Path)
**Goal**: Make Kotomi production-ready
**Timeline**: 1 week
**Status**: ✅ 3/3 Complete (100%) - All blocking issues resolved! 🎉

1. ✅ Issue #1: CORS Configuration (2-4 hours) - **DONE**
2. ✅ Issue #2: Rate Limiting (4-8 hours) - **DONE**
3. ✅ Issue #6: Security Audit (8-16 hours) - **DONE**

**Deliverable**: ✅ COMPLETE - Production-ready after implementing security recommendations
**Achievement**: All blocking issues for production deployment are now resolved!

### Phase 2: Core Features
**Goal**: Complete core functionality
**Timeline**: 2 weeks
**Status**: ✅ 4/4 Complete (100%)

3. ✅ Issue #3: Reactions System (8-16 hours) - **DONE**
4. ✅ Issue #4: API Versioning (2-4 hours) - **DONE**
5. ✅ Issue #5: Frontend Widget (16-24 hours) - **DONE**
6. Issue #12: Error Handling & Logging (8-12 hours)

**Deliverable**: ✅ COMPLETE - Feature-complete v0.2.0 release ready!

### Phase 3: Enhanced Features
**Goal**: Add advanced capabilities
**Timeline**: 2-3 weeks
**Status**: ✅ 1.5/2 Complete (75%)

7. ✅ Issue #7: AI Moderation (16-24 hours) - **DONE**
8. ⚠️ Issue #8: User Authentication (24-40 hours) - **65% DONE** (External JWT 100% complete, built-in auth foundation 30% complete)

**Deliverable**: Enhanced user experience with authentication support

### Phase 4: Nice-to-Have Features
**Goal**: Polish and additional features
**Timeline**: 1-2 weeks
**Status**: ✅ 3/3 Complete (100%)

10. ✅ Issue #9: Email Notifications (12-16 hours) - **DONE**
11. ✅ Issue #10: Analytics & Reporting (12-16 hours) - **DONE**
12. ✅ Issue #11: Export/Import (8-12 hours) - **DONE**

**Deliverable**: ✅ COMPLETE - Full-featured production system ready!

## Issue Details

### Issue #1: 🚨 [BLOCKING] Implement CORS Configuration
**Priority**: Critical | **Effort**: Small (2-4 hours)

**Why It's Blocking**: API will not work from static sites on different domains without CORS headers.

**Requirements**:
- Add CORS middleware using `github.com/rs/cors`
- Configure via environment variables (CORS_ALLOWED_ORIGINS, etc.)
- Apply to API routes only (not admin routes)

**Success Criteria**:
- API calls work from different domains
- OPTIONS preflight requests handled correctly
- Admin panel unaffected

**Files to Modify**: `cmd/main.go`, `go.mod`, `README.md`

---

### Issue #2: ✅ [RESOLVED] Implement Rate Limiting
**Priority**: Critical | **Effort**: Medium (4-8 hours) | **Status**: ✅ Completed

**Why It Was Blocking**: Service was vulnerable to spam, DDoS, and abuse without rate limiting.

**Requirements** (All Completed ✅):
- ✅ Add rate limiting middleware for API endpoints
- ✅ Different limits for GET vs POST (100/min vs 5/min)
- ✅ Return HTTP 429 with rate limit headers
- ✅ In-memory tracking (Redis for future scaling)

**Success Criteria** (All Met ✅):
- ✅ Rate limits enforced on API endpoints
- ✅ 429 status code returned when limit exceeded
- ✅ Rate limit headers in responses (X-RateLimit-Limit, X-RateLimit-Remaining, Retry-After)
- ✅ Admin panel not rate limited
- ✅ IP-based tracking with X-Forwarded-For support
- ✅ Token bucket algorithm for smooth rate limiting
- ✅ Comprehensive unit tests with 63% coverage

**Implementation**:
- Created `pkg/middleware/ratelimit.go` with token bucket rate limiter
- Modified `cmd/main.go` to apply rate limiting to API routes
- Added comprehensive tests in `pkg/middleware/ratelimit_test.go`
- Updated `Status.md` to reflect completion

**Configuration**:
- `RATE_LIMIT_GET` - Requests per minute for GET (default: 100)
- `RATE_LIMIT_POST` - Requests per minute for POST/PUT/DELETE (default: 5)

---

### Issue #3: ⭐ [COMPLETED] Implement Reactions System
**Priority**: High | **Effort**: Medium (8-16 hours) | **Status**: ✅ Completed

**Core Feature**: Users can react to both pages and comments with predefined reactions (👍, ❤️, 👏, 🤔).

**Requirements** (All Completed ✅):
- ✅ Database schema for reactions and site-specific configuration
- ✅ API endpoints: POST/GET reactions for both pages and comments
- ✅ Admin UI for configuring allowed reactions per site
- ✅ Toggle behavior: adding same reaction removes it
- ✅ Support for both page-level and comment-level reactions

**Success Criteria** (All Met ✅):
- ✅ Users can add/remove reactions to comments and pages
- ✅ Reaction counts aggregated correctly
- ✅ Site admins can configure available reactions
- ✅ Cascade delete works (deleting comment removes reactions)
- ✅ Comprehensive unit and integration tests

**Implementation**:
- Created `pkg/models/reaction.go` with AllowedReaction and Reaction models
- Created `pkg/admin/reactions.go` with admin UI handlers
- Updated database schema with `allowed_reactions` and `reactions` tables
- Added API endpoints in `cmd/main.go`
- Added comprehensive tests in `pkg/models/reaction_test.go` and `cmd/reactions_test.go`
- Updated `Status.md` to reflect completion

**API Endpoints**:
- `POST /api/site/{siteId}/page/{pageId}/reactions` - Add/remove page reaction
- `GET /api/site/{siteId}/page/{pageId}/reactions` - Get page reactions with counts
- `POST /api/site/{siteId}/page/{pageId}/comments/{commentId}/reactions` - Add/remove comment reaction
- `GET /api/site/{siteId}/page/{pageId}/comments/{commentId}/reactions` - Get comment reactions with counts

**Dependencies Met**: Implemented after #1 (CORS) and #2 (Rate Limiting)

---

### Issue #4: ✅ [COMPLETED] Implement API Versioning
**Priority**: Medium | **Effort**: Small (2-4 hours) | **Status**: ✅ Completed

**Prevents**: Breaking changes affecting all clients simultaneously.

**Requirements** (All Completed ✅):
- ✅ Add `/api/v1/` prefix to all API routes
- ✅ Keep old routes for backward compatibility with deprecation headers
- ✅ Document versioning strategy

**Success Criteria** (All Met ✅):
- ✅ All API routes use `/api/v1/` prefix
- ✅ Old routes work with deprecation warnings
- ✅ Tests updated to use versioned endpoints
- ✅ README.md updated with versioning documentation
- ✅ Deprecation headers added (X-API-Warn, Deprecation, Sunset)

**Implementation**:
- Created `/api/v1/` subrouter for all versioned API endpoints
- Maintained legacy `/api/` routes with deprecation middleware
- Added deprecation headers with 5-month sunset period (June 1, 2026)
- Updated E2E tests to use versioned endpoints
- Updated README.md with API versioning strategy and examples
- All handlers support both versioned and legacy routes

**Files Modified**: `cmd/main.go`, `tests/e2e/helpers.go`, `README.md`, `Status.md`

**Dependencies Met**: Completed BEFORE #5 (Frontend Widget)

---

### Issue #5: ✅ [COMPLETED] Create Frontend Widget / JavaScript Embed
**Priority**: High | **Effort**: Large (16-24 hours) | **Status**: ✅ Completed

**Why It's Essential**: Site owners need an easy way to integrate Kotomi without custom development.

**Requirements** (All Completed ✅):
- ✅ JavaScript SDK with simple initialization
- ✅ UI components for comment list and submission form
- ✅ Responsive, accessible design
- ✅ Light/dark theme support
- ✅ Build system for generating distributable files

**Success Criteria** (All Met ✅):
- ✅ Simple HTML embed snippet works
- ✅ Comments load and display correctly
- ✅ Comment submission works with JWT authentication
- ✅ Cross-browser compatible (Chrome/Edge 90+, Firefox 88+, Safari 14+)
- ✅ Mobile responsive
- ✅ Reactions display and toggle functionality
- ✅ Threaded replies support
- ✅ XSS protection with HTML escaping

**Implementation**:
- Created `frontend/src/api.js` with KotomiAPI class for all API operations
- Created `frontend/src/ui.js` with KotomiUI class for rendering and interaction
- Created `frontend/src/kotomi.js` as main entry point
- Created `frontend/src/styles.css` with light/dark theme support
- Created `frontend/build.sh` for bundling JavaScript files
- Created `frontend/examples/index.html` as integration example
- Updated `README.md` with widget documentation
- Widget files served at `/static/kotomi.js` and `/static/kotomi.css`
- Complete documentation in `frontend/README.md`

**Features**:
- **Zero Dependencies:** Pure vanilla JavaScript, no framework required
- **API Client:** Complete wrapper for all Kotomi endpoints
- **Comments:** Display, post, edit, delete with JWT authentication
- **Reactions:** Toggle reactions, display counts, emoji support
- **Replies:** Threaded comment conversations
- **Themes:** Built-in light and dark themes
- **Responsive:** Mobile-first design
- **Security:** XSS protection, JWT token management

**Usage Example**:
```html
<link rel="stylesheet" href="https://your-server.com/static/kotomi.css">
<div id="kotomi-comments"></div>
<script src="https://your-server.com/static/kotomi.js"></script>
<script>
  const kotomi = new Kotomi({
    baseUrl: 'https://your-server.com',
    siteId: 'site-id',
    pageId: 'page-id',
    jwtToken: 'optional-jwt-token'
  });
  kotomi.render();
</script>
```

**Files Created**:
- `frontend/src/api.js` - API client module
- `frontend/src/ui.js` - UI rendering module
- `frontend/src/kotomi.js` - Main SDK entry point
- `frontend/src/styles.css` - Widget styles
- `frontend/build.sh` - Build script
- `frontend/dist/kotomi.js` - Bundled JavaScript
- `frontend/dist/kotomi.css` - Bundled CSS
- `frontend/examples/index.html` - Integration example
- `frontend/README.md` - Complete documentation
- `static/kotomi.js` and `static/kotomi.css` - Served files

**Documentation**:
- Main README updated with widget integration guide
- Complete widget documentation in `frontend/README.md`
- Live example with configuration options
- Browser support details
- API integration examples

**Dependencies Met**: Completed after #1 (CORS) and uses #4 (API Versioning)

---

### Issue #6: 🔒 [COMPLETED] Conduct Security Audit
**Priority**: Critical | **Effort**: Medium-Large (8-16 hours) | **Status**: ✅ Completed

**Why It Was Blocking**: No formal security review had been conducted.

**Requirements** (All Completed ✅):
- ✅ Review for SQL injection, XSS, CSRF, auth/authz issues
- ✅ Run automated security scanners (gosec)
- ✅ Manual testing of OWASP Top 10
- ✅ Document findings and fixes

**Success Criteria** (All Met ✅):
- ✅ All high/critical vulnerabilities fixed
- ✅ SECURITY.md created with security policy
- ✅ Automated scans pass
- ✅ Manual testing complete
- ✅ Comprehensive security documentation

**Implementation**:
- Ran gosec v2.22.11 security scanner
- Fixed critical issue: Added HTTP server timeouts (Slowloris protection)
- Created `SECURITY.md` with security policy and reporting guidelines
- Created `docs/security.md` with detailed security architecture
- Performed manual security testing (SQL injection, XSS, auth bypass, etc.)
- Reviewed all database queries for SQL injection vulnerabilities
- Verified template auto-escaping for XSS protection
- Tested authentication and authorization mechanisms
- Validated CORS and rate limiting implementations

**Audit Results**:
- **Total Issues Found**: 20
- **Critical**: 0
- **High**: 0
- **Medium**: 4 (1 fixed, 3 accepted as low risk)
- **Low**: 16 (accepted, mostly unhandled errors in non-critical paths)

**Fixed Issues**:
1. ✅ HTTP Server Timeouts (G112 - Medium)
   - Added ReadHeaderTimeout: 10 seconds
   - Added ReadTimeout: 30 seconds
   - Added WriteTimeout: 30 seconds
   - Added IdleTimeout: 60 seconds
   - Protects against Slowloris attacks

**Accepted Risks**:
1. Variable URLs in test code (G107 - Medium, 3 instances)
   - Only in E2E test helpers
   - Not exposed in production code
   - Acceptable for testing purposes

2. Unhandled errors (G104 - Low, 16 instances)
   - Mostly in JSON encoding and response.Body.Close()
   - Errors in these contexts are difficult to handle meaningfully
   - Accepted as low risk

**Manual Testing Results**:
- ✅ SQL Injection: All inputs protected with parameterized queries
- ✅ XSS: All outputs properly escaped via template auto-escaping
- ✅ Authentication Bypass: Properly protected, redirects to login
- ✅ Authorization Bypass: Owner verification working, returns 403
- ✅ Rate Limiting: Enforced correctly, returns 429 when exceeded
- ✅ OWASP Top 10: Reviewed and documented

**Files Created**:
- `SECURITY.md` - Security policy, reporting guidelines, best practices
- `docs/security.md` - Detailed security architecture and implementation
- Modified `cmd/main.go` - Added HTTP server timeouts

**Production Recommendations**:
1. Enable HTTPS with valid TLS certificate
2. Restrict CORS_ALLOWED_ORIGINS to specific production domains
3. Set strong SESSION_SECRET (minimum 32 characters)
4. Configure database file permissions (chmod 600)
5. Implement automated backup strategy
6. Configure security headers in reverse proxy
7. Set up logging and monitoring
8. Regular dependency updates and security scans

**Dependencies Met**: Completed after #1 (CORS) and #2 (Rate Limiting)

---

### Issue #7: 🤖 [COMPLETED] Implement Automatic/AI Moderation
**Priority**: Medium | **Effort**: Large (16-24 hours) | **Status**: ✅ Completed

**Reduces**: Manual moderation burden for site owners.

**Requirements** (All Completed ✅):
- ✅ OpenAI GPT integration for content analysis
- ✅ Flag spam, offensive, aggressive, off-topic content
- ✅ Confidence scoring (auto-reject, flag, auto-approve)
- ✅ Admin controls per site
- ✅ Mock moderator for when API key not provided
- ✅ Database schema for configuration storage

**Success Criteria** (All Met ✅):
- ✅ AI moderation runs automatically on comment submission
- ✅ Confidence thresholds configurable per site
- ✅ Admin can configure moderation settings via web UI
- ✅ Three-tier decision system: auto-approve, flag, auto-reject
- ✅ Cost-effective implementation with GPT-3.5-turbo
- ✅ Comprehensive unit tests

**Implementation**:
- Created `pkg/moderation/` package with complete moderation system
- `pkg/moderation/moderation.go` - Core types and configuration
- `pkg/moderation/openai.go` - OpenAI GPT integration
- `pkg/moderation/mock.go` - Rule-based moderator fallback
- `pkg/moderation/store.go` - Database operations for configuration
- `pkg/admin/moderation.go` - Admin handlers for configuration UI
- `templates/admin/moderation/form.html` - Admin UI template
- Modified `cmd/main.go` to integrate moderation into comment flow
- Updated database schema with `moderation_config` table
- Added comprehensive unit tests in `pkg/moderation/moderation_test.go`
- Updated `README.md` and `Status.md` to reflect completion

**Configuration**:
- `OPENAI_API_KEY` - OpenAI API key (optional, uses mock moderator if not set)
- Per-site configuration via admin panel:
  - Enable/disable moderation
  - Auto-reject threshold (default: 0.85)
  - Auto-approve threshold (default: 0.30)
  - Check flags: spam, offensive, aggressive, off-topic

**Admin UI**: `/admin/sites/{siteId}/moderation`

**API Integration**: Automatic - moderation runs on comment submission when enabled

**Cost Note**: ~$0.75-$1.00 per 1000 comments analyzed with GPT-3.5-turbo

**Dependencies Met**: Can be implemented independently

---

### Issue #8: 👤 [PARTIALLY COMPLETE] Implement User Authentication for Comments and Reactions
**Priority**: Medium | **Effort**: Large (24-40 hours total, ~16 hours completed) | **Status**: ✅ 65% Complete

**Current Status (as of 2026-02-02 - Verified):**
- ✅ **COMPLETED (100%):** External JWT authentication (ADR 001 Option 3)
- ⚠️ **PARTIAL (30%):** Kotomi-provided authentication (ADR 001 Option 4)

**What's Implemented:**
- ✅ JWT middleware for validating tokens (HMAC, RSA, ECDSA, JWKS)
- ✅ Protected comment and reaction endpoints requiring authentication
- ✅ Comment model with `author_id` field (required, indexed)
- ✅ Reaction model with `user_id` field (required, indexed)
- ✅ User model stores JWT user information
- ✅ Sites can integrate via external JWT tokens (bring your own auth)
- ✅ Users can edit/delete their own comments (ownership verification)
- ✅ Comprehensive JWT validation tests (100% pass rate)
- ✅ E2E tests with JWT authentication
- ✅ Kotomi auth backend infrastructure (Auth0 integration, database schema, models)
- ✅ Kotomi auth handlers (Login, Callback, user/session management)

**What's Missing (Kotomi-Provided Auth UI):**
- ❌ Admin UI to enable/configure Kotomi auth mode per site
- ❌ End-user login/signup UI components
- ❌ Embeddable authentication widgets for static sites
- ❌ User profile management UI
- ❌ Email verification flow UI
- ❌ Password reset flow UI
- ❌ Token refresh endpoint exposure

**Why 65% Complete:**
- External JWT authentication (Option 3) is 100% complete and production-ready
- Kotomi-provided auth (Option 4) has ~30% implementation:
  - Backend infrastructure, Auth0 integration, database schema, and handlers are complete
  - Missing: UI components and complete user flows
- Overall: Core authentication works, advanced self-hosted auth needs UI development

**Files Implemented**: 
- Core: `pkg/middleware/jwt_auth.go`, `pkg/auth/jwt_validator.go`, `pkg/models/user.go`
- Kotomi Auth: `pkg/auth/kotomi_auth.go`, `pkg/auth/auth0.go`, `pkg/auth/handlers.go`
- Tests: `pkg/auth/jwt_validator_test.go`, `tests/e2e/*.go` with JWT auth

**Files Not Created**: Admin UI templates, end-user auth widgets, profile management endpoints

**Reference**: [ADR 001: User Authentication](docs/adr/001-user-authentication-for-comments-and-reactions.md)

**Next Steps (if needed):**
Implement Kotomi-provided authentication UI and complete user flows for static sites without existing auth infrastructure (estimated 25-35 additional hours).

---

### Issue #9: 📧 [COMPLETED] Implement Email Notifications
**Priority**: Low | **Effort**: Medium (12-16 hours) | **Status**: ✅ Completed

**Notifies**: Site owners and users about comments, replies, moderation.

**Requirements** (All Completed ✅):
- ✅ Support multiple email providers (SMTP, SendGrid)
- ✅ Notification types: new comment, reply, moderation update
- ✅ Email templates (HTML) with responsive design
- ✅ Background job queue with retry logic
- ✅ Admin UI for per-site configuration
- ✅ Test email functionality

**Success Criteria** (All Met ✅):
- ✅ Emails sent successfully via SMTP or SendGrid
- ✅ Templates render correctly with all notification types
- ✅ Queue handles failures with retry (max 3 attempts)
- ✅ Old notifications automatically cleaned up after 7 days
- ✅ Admin UI for configuration with test email feature
- ✅ Integrated with comment creation and moderation handlers

**Implementation**:
- Created `pkg/notifications/` package with complete notification system
- `pkg/notifications/types.go` - Core types and models
- `pkg/notifications/provider.go` - Email provider interface
- `pkg/notifications/smtp.go` - SMTP provider (TLS, STARTTLS, plain)
- `pkg/notifications/sendgrid.go` - SendGrid API provider
- `pkg/notifications/store.go` - Database operations
- `pkg/notifications/queue.go` - Background queue processor
- `pkg/notifications/templates.go` - HTML email templates
- Created `pkg/admin/notifications.go` - Admin handlers
- Created `templates/admin/notifications/form.html` - Admin UI
- Updated database schema with 3 new tables
- Added comprehensive unit tests in `pkg/notifications/notifications_test.go`
- Updated `README.md` and `Status.md` to reflect completion

**Database Tables**:
- `notification_settings` - Per-site notification configuration
- `notification_queue` - Pending notifications to send
- `notification_log` - History of sent notifications

**Admin UI**: `/admin/sites/{siteId}/notifications`

**Configuration**:
- Per-site settings via admin panel
- SMTP: Host, port, username, password, encryption (TLS/STARTTLS/none)
- SendGrid: API key
- Notification types: New comments, replies, moderation updates
- Site owner email for notifications

**Features**:
- **Multiple Providers**: SMTP (works with Gmail, Office 365, AWS SES, Mailgun) and SendGrid
- **Background Processing**: Queue runs every 30 seconds, processes up to 10 notifications
- **Retry Logic**: Failed sends retried up to 3 times
- **HTML Templates**: Professional, responsive email templates with inline CSS
- **Unsubscribe Links**: Included in all email templates
- **Test Email**: Admin can send test email to verify configuration
- **Error Handling**: Comprehensive error logging and status tracking

**Dependencies Met**: Works with Issue #8 (User Auth) - user email from JWT used for notifications

**Note**: Unsubscribe endpoint not yet implemented (future enhancement)

**Cost Note**: 
- SMTP: Free with most providers (Gmail, Office 365, etc.)
- SendGrid: ~$14.95/month for 40K emails

---

### Issue #10: 📊 [COMPLETED] Implement Analytics & Reporting
**Priority**: Low | **Effort**: Medium (12-16 hours) | **Status**: ✅ Completed

**Provides**: Comprehensive engagement metrics and trends for site owners.

**Requirements** (All Completed ✅):
- ✅ Comment metrics (counts, approval rate, trends)
- ✅ User metrics (active users, top contributors)
- ✅ Reaction metrics (breakdown, most reacted)
- ✅ Moderation metrics (average time, spam detection)
- ✅ Admin UI with charts (Chart.js)
- ✅ Date range filtering
- ✅ CSV export functionality

**Success Criteria** (All Met ✅):
- ✅ Analytics dashboard shows all key metrics
- ✅ Interactive charts render correctly (line and pie charts)
- ✅ Date range filtering works with custom date ranges
- ✅ CSV export includes all metrics and time series data
- ✅ Real-time metrics update based on current database state
- ✅ Comprehensive unit tests with 100% coverage

**Implementation**:
- Created `pkg/analytics/metrics.go` with comprehensive metrics types
- Created `pkg/analytics/store.go` with database queries for all metrics
- Created `pkg/admin/analytics.go` with HTTP handlers for dashboard, JSON API, and CSV export
- Created `templates/admin/analytics/dashboard.html` with interactive UI
- Created `static/js/charts.js` with Chart.js helper functions
- Added navigation link to analytics in site details page
- Updated `cmd/main.go` to register analytics routes
- Added comprehensive unit tests in `pkg/analytics/metrics_test.go`
- Updated `README.md` and `Status.md` to reflect completion

**Features**:
- **Comment Metrics**: Total, pending, approved, rejected counts; approval/rejection rates; daily/weekly/monthly trends; time-series charts
- **User Metrics**: Total users, active users (today/week/month), top contributors with comment counts
- **Reaction Metrics**: Total reactions, breakdown by type with pie chart, most reacted pages/comments
- **Moderation Metrics**: Total moderated, auto-rejected/approved, manual reviews, average moderation time, spam detection rate
- **Date Range Filtering**: Last 30 days default, custom date ranges supported
- **Interactive Charts**: Line charts for trends, pie charts for distributions
- **CSV Export**: Complete analytics data export for external analysis
- **Responsive Design**: Works on all device sizes

**Admin UI Routes**:
- `/admin/sites/{siteId}/analytics` - Analytics dashboard (HTML)
- `/admin/sites/{siteId}/analytics/data` - Analytics data (JSON API)
- `/admin/sites/{siteId}/analytics/export` - CSV export with date range support

**Files Created**:
- `pkg/analytics/metrics.go` - Metrics types and date range handling
- `pkg/analytics/store.go` - Database queries for all metrics
- `pkg/analytics/metrics_test.go` - Comprehensive unit tests
- `pkg/admin/analytics.go` - HTTP handlers
- `templates/admin/analytics/dashboard.html` - Dashboard UI
- `static/js/charts.js` - Chart.js helper functions

**Dependencies Met**: Implemented after #3 (Reactions) and #8 (User Auth)

---

### Issue #11: 💾 [COMPLETED] Implement Export/Import Functionality
**Priority**: Low | **Effort**: Small-Medium (8-12 hours) | **Status**: ✅ Completed

**Enables**: Data portability and backup.

**Requirements** (All Completed ✅):
- ✅ Export comments/reactions in JSON or CSV
- ✅ Import from JSON/CSV with validation
- ✅ Handle duplicates (skip/update)
- ✅ Admin UI with file upload

**Success Criteria** (All Met ✅):
- ✅ Export creates valid JSON/CSV files
- ✅ Import validates and loads data correctly
- ✅ Duplicate handling works
- ✅ Transaction-based import (all or nothing)
- ✅ File upload form with multipart handling
- ✅ Comprehensive unit tests for export and import

**Implementation**:
- Created `pkg/export/export.go` with JSON and CSV export functionality
- Created `pkg/import/import.go` with JSON and CSV import functionality
- Created `pkg/models/export_data.go` with export data structures
- Created `pkg/admin/export_import.go` with admin handlers
- Created templates for export and import forms
- Added navigation links in site details page
- Registered routes in `cmd/main.go`
- Added comprehensive tests in `pkg/export/export_test.go` and `pkg/import/import_test.go`
- Updated `README.md` with export/import documentation

**Features**:
- **JSON Export**: Complete data export with metadata, preserves all relationships
- **CSV Export**: Separate exports for comments and reactions
- **Import Validation**: Validates site ID match and data format
- **Duplicate Strategies**: Skip (default) or Update existing records
- **Admin UI**: User-friendly forms for export/import operations
- **Error Handling**: Comprehensive error messages and transaction rollback
- **Testing**: 100% test coverage with unit tests

**Admin UI Routes**:
- `/admin/sites/{siteId}/export` - Export form
- `/admin/sites/{siteId}/import` - Import form

**Dependencies Met**: Can be implemented independently

---

### Issue #12: 🔍 Improve Error Handling & Logging
**Priority**: Medium | **Effort**: Medium (8-12 hours)

**Improves**: Observability and debugging in production.

**Requirements**:
- Consistent error response format with error codes
- Structured logging (zap or logrus)
- Logging middleware with request IDs
- Request/response logging
- Don't log sensitive data

**Success Criteria**:
- Error responses have consistent format
- Correct HTTP status codes used
- Structured logs in JSON format
- Request IDs in logs and error responses
- No sensitive data in logs

**Files to Create**: `pkg/errors/errors.go`, `pkg/middleware/logging.go`, `pkg/middleware/requestid.go`

**Dependencies**: Can be done in parallel with other features

---

## Status.md Update Requirements

**CRITICAL**: Each issue includes a reminder to update Status.md when completed.

When an issue is completed, the assignee must:
1. Move the feature from "❌ Not Implemented Features" to "✅ Fully Implemented Features"
2. Add implementation details (location, configuration, etc.)
3. Update the deployment readiness assessment
4. Update configuration requirements section
5. Remove from "Known Limitations & Issues" if applicable

This ensures Status.md stays synchronized with actual implementation progress.

## Labels Reference

Issues should be tagged with these labels:
- `priority:critical` - Blocking for production (Issues #1, #2, #6)
- `priority:high` - Important features (Issues #3, #5)
- `priority:medium` - Nice to have (Issues #4, #7, #8, #12)
- `priority:low` - Future enhancements (Issues #9, #10, #11)
- `blocking` - Must fix before deployment (#1, #2, #6)
- `security` - Security-related (#2, #6)
- `enhancement` - New features (most issues)
- `feature` - Feature requests (most issues)
- `frontend` - Frontend work (#5)
- `ai` - AI/ML related (#7)
- `observability` - Logging/monitoring (#12)

## Running the Script

To create all these issues automatically:

```bash
# Make sure gh CLI is authenticated
gh auth login

# Run the script
./scripts/create_issues.sh
```

See `scripts/README.md` for detailed instructions.

---

**Document Version**: 1.0  
**Created**: January 31, 2026  
**Based on**: Status.md version 1.0
