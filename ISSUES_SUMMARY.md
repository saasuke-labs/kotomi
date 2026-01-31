# GitHub Issues Summary - Based on Status.md

This document contains a summary of all GitHub issues that should be created for Kotomi based on the Status.md analysis.

## Quick Reference

| # | Title | Priority | Status | Estimated Effort | Dependencies |
|---|-------|----------|--------|------------------|--------------|
| 1 | ✅ [RESOLVED] Implement CORS Configuration | Critical | ✅ Done | 2-4 hours | None |
| 2 | ✅ [RESOLVED] Implement Rate Limiting | Critical | ✅ Done | 4-8 hours | None |
| 3 | ⭐ [COMPLETED] Implement Reactions System | High | ✅ Done | 8-16 hours | After #1, #2 |
| 4 | 🔧 Implement API Versioning | Medium | Pending | 2-4 hours | Before #5 |
| 5 | 🎨 Create Frontend Widget / JavaScript Embed | High | Pending | 16-24 hours | After #1, #4 |
| 6 | 🔒 [COMPLETED] Conduct Security Audit | Critical | ✅ Done | 8-16 hours | After #1, #2 |
| 7 | 🤖 Implement Automatic/AI Moderation | Medium | Pending | 16-24 hours | Independent |
| 8 | 👤 Implement User Authentication for Comments | Medium | Pending | 24-40 hours | After #5 |
| 9 | 📧 Implement Email Notifications | Low | Pending | 12-16 hours | After #8 |
| 10 | 📊 Implement Analytics & Reporting | Low | Pending | 12-16 hours | After #3, #8 |
| 11 | 💾 Implement Export/Import Functionality | Low | Pending | 8-12 hours | Independent |
| 12 | 🔍 Improve Error Handling & Logging | Medium | Pending | 8-12 hours | Independent |

**Total Estimated Effort**: 118-178 hours (approximately 3-4 weeks of full-time development)
**Completed**: Issues #1, #2, #3, #6 (22-44 hours completed)
**Remaining**: 96-134 hours

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
**Status**: ✅ 1/4 Complete (25%)

3. ✅ Issue #3: Reactions System (8-16 hours) - **DONE**
4. Issue #4: API Versioning (2-4 hours)
5. Issue #5: Frontend Widget (16-24 hours)
6. Issue #12: Error Handling & Logging (8-12 hours)

**Deliverable**: Feature-complete v0.2.0 release

### Phase 3: Enhanced Features
**Goal**: Add advanced capabilities
**Timeline**: 2-3 weeks

8. Issue #7: AI Moderation (16-24 hours)
9. Issue #8: User Authentication (24-40 hours)

**Deliverable**: Enhanced user experience

### Phase 4: Nice-to-Have Features
**Goal**: Polish and additional features
**Timeline**: 1-2 weeks

10. Issue #9: Email Notifications (12-16 hours)
11. Issue #10: Analytics & Reporting (12-16 hours)
12. Issue #11: Export/Import (8-12 hours)

**Deliverable**: Full-featured production system

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

### Issue #4: 🔧 Implement API Versioning
**Priority**: Medium | **Effort**: Small (2-4 hours)

**Prevents**: Breaking changes affecting all clients simultaneously.

**Requirements**:
- Add `/api/v1/` prefix to all API routes
- Keep old routes for backward compatibility with deprecation headers
- Document versioning strategy

**Success Criteria**:
- All API routes use `/api/v1/` prefix
- Old routes redirect with deprecation warnings
- Tests updated to use versioned endpoints

**Files to Modify**: `cmd/main.go`, `tests/e2e/api_test.go`, `README.md`, `Status.md`

**Dependencies**: Should be done BEFORE #5 (Frontend Widget)

---

### Issue #5: 🎨 Create Frontend Widget / JavaScript Embed
**Priority**: High | **Effort**: Large (16-24 hours)

**Why It's Essential**: Site owners need an easy way to integrate Kotomi without custom development.

**Requirements**:
- JavaScript SDK with simple initialization
- UI components for comment list and submission form
- Responsive, accessible design
- Light/dark theme support
- Build system (webpack/rollup)

**Success Criteria**:
- Simple HTML embed snippet works
- Comments load and display correctly
- Comment submission works
- Cross-browser compatible
- Mobile responsive

**Files to Create**: `frontend/src/kotomi.js`, `frontend/src/ui.js`, `frontend/src/api.js`, `frontend/src/styles.css`

**Dependencies**: AFTER #1 (CORS), should use #4 (API Versioning) if available

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

### Issue #7: 🤖 Implement Automatic/AI Moderation
**Priority**: Medium | **Effort**: Large (16-24 hours)

**Reduces**: Manual moderation burden for site owners.

**Requirements**:
- OpenAI GPT integration for content analysis
- Flag spam, offensive, aggressive, off-topic content
- Confidence scoring (auto-reject, flag, auto-approve)
- Admin controls per site

**Success Criteria**:
- AI moderation runs on comment submission
- Confidence thresholds configurable
- Admin can view AI decisions and override
- Cost-effective (document API costs)

**Files to Create**: `pkg/moderation/ai.go`, `pkg/moderation/openai.go`

**Cost Note**: ~$0.75-$1.00 per 1000 comments analyzed

---

### Issue #8: 👤 Implement User Authentication for Comments
**Priority**: Medium | **Effort**: Large (24-40 hours)

**Enables**: Users to edit/delete their own comments and build profiles.

**Requirements**:
- Email/password authentication for commenters
- Social logins (Google, GitHub, Twitter)
- Optional anonymous posting per site
- Edit/delete own comments functionality
- User profiles

**Success Criteria**:
- Users can register and login
- Authenticated users can edit/delete their comments
- Anonymous posting still works if enabled
- Password reset flow works

**Files to Create**: `pkg/auth/public_auth.go`, `pkg/models/public_user.go`

**Dependencies**: AFTER #5 (Frontend Widget), requires email service

**Design Decisions Needed**:
- Separate auth system or share with admin Auth0?
- Social login providers?
- Edit time limit?

---

### Issue #9: 📧 Implement Email Notifications
**Priority**: Low | **Effort**: Medium (12-16 hours)

**Notifies**: Site owners and users about comments, replies, moderation.

**Requirements**:
- Support multiple email providers (SendGrid, AWS SES, Mailgun, SMTP)
- Notification types: new comment, reply, moderation update
- Email templates (HTML)
- Unsubscribe functionality
- Background job queue

**Success Criteria**:
- Emails sent successfully
- Templates render correctly
- Unsubscribe works
- Queue handles failures with retry

**Files to Create**: `pkg/notifications/email.go`, `pkg/notifications/queue.go`, `templates/email/`

**Dependencies**: After #8 (User Auth) for full functionality

**Cost Note**: SendGrid ~$14.95/month for 40K emails

---

### Issue #10: 📊 Implement Analytics & Reporting
**Priority**: Low | **Effort**: Medium (12-16 hours)

**Provides**: Engagement metrics and trends for site owners.

**Requirements**:
- Comment metrics (counts, approval rate, trends)
- User metrics (active users, top contributors)
- Reaction metrics (breakdown, most reacted)
- Moderation metrics (average time, spam detection)
- Admin UI with charts (Chart.js)

**Success Criteria**:
- Analytics dashboard shows key metrics
- Charts render correctly
- Date range filtering works
- Export to CSV works

**Files to Create**: `pkg/analytics/metrics.go`, `templates/admin/analytics.html`, `static/js/charts.js`

**Dependencies**: More valuable after #3 (Reactions) and #8 (User Auth)

---

### Issue #11: 💾 Implement Export/Import Functionality
**Priority**: Low | **Effort**: Small-Medium (8-12 hours)

**Enables**: Data portability and backup.

**Requirements**:
- Export comments/reactions in JSON or CSV
- Import from JSON/CSV with validation
- Handle duplicates (skip/update)
- Admin UI with file upload

**Success Criteria**:
- Export creates valid JSON/CSV files
- Import validates and loads data correctly
- Duplicate handling works
- Large files handled efficiently

**Files to Create**: `pkg/export/export.go`, `pkg/import/import.go`, templates

**Dependencies**: Can be implemented independently

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
