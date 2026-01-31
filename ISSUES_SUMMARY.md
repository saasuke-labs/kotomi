# GitHub Issues Summary - Based on Status.md

This document contains a summary of all GitHub issues that should be created for Kotomi based on the Status.md analysis.

## Quick Reference

| # | Title | Priority | Status | Estimated Effort | Dependencies |
|---|-------|----------|--------|------------------|--------------|
| 1 | ✅ [RESOLVED] Implement CORS Configuration | Critical | ✅ Done | 2-4 hours | None |
| 2 | ✅ [RESOLVED] Implement Rate Limiting | Critical | ✅ Done | 4-8 hours | None |
| 3 | ⭐ Implement Reactions System | High | Pending | 8-16 hours | After #1, #2 |
| 4 | 🔧 Implement API Versioning | Medium | Pending | 2-4 hours | Before #5 |
| 5 | 🎨 Create Frontend Widget / JavaScript Embed | High | Pending | 16-24 hours | After #1, #4 |
| 6 | 🔒 Conduct Security Audit | Critical | Pending | 8-16 hours | After #1, #2 |
| 7 | 🤖 Implement Automatic/AI Moderation | Medium | Pending | 16-24 hours | Independent |
| 8 | 👤 Implement User Authentication for Comments | Medium | Pending | 24-40 hours | After #5 |
| 9 | 📧 Implement Email Notifications | Low | Pending | 12-16 hours | After #8 |
| 10 | 📊 Implement Analytics & Reporting | Low | Pending | 12-16 hours | After #3, #8 |
| 11 | 💾 Implement Export/Import Functionality | Low | Pending | 8-12 hours | Independent |
| 12 | 🔍 Improve Error Handling & Logging | Medium | Pending | 8-12 hours | Independent |

**Total Estimated Effort**: 118-178 hours (approximately 3-4 weeks of full-time development)
**Completed**: Issues #1, #2 (6-12 hours completed)
**Remaining**: 112-166 hours

## Implementation Phases

### Phase 1: Blocking Issues (Critical Path)
**Goal**: Make Kotomi production-ready
**Timeline**: 1 week
**Status**: ✅ 2/3 Complete (67%)

1. ✅ Issue #1: CORS Configuration (2-4 hours) - **DONE**
2. ✅ Issue #2: Rate Limiting (4-8 hours) - **DONE**
3. ⏳ Issue #6: Security Audit (8-16 hours) - **REMAINING**

**Deliverable**: Minimal viable version ready for production deployment
**Next Step**: Security audit is now the only remaining blocker for production

### Phase 2: Core Features
**Goal**: Complete core functionality
**Timeline**: 2 weeks

4. Issue #4: API Versioning (2-4 hours)
5. Issue #5: Frontend Widget (16-24 hours)
6. Issue #3: Reactions System (8-16 hours)
7. Issue #12: Error Handling & Logging (8-12 hours)

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

### Issue #3: ⭐ Implement Reactions System
**Priority**: High | **Effort**: Medium (8-16 hours)

**Core Feature**: Users can react to comments with predefined reactions (👍, ❤️, 👏, 🤔).

**Requirements**:
- Database schema for reactions and site-specific configuration
- API endpoints: POST/GET/DELETE reactions
- Admin UI for configuring allowed reactions per site
- Toggle behavior: adding same reaction removes it

**Success Criteria**:
- Users can add/remove reactions to comments
- Reaction counts aggregated correctly
- Site admins can configure available reactions
- Cascade delete works (deleting comment removes reactions)

**Files to Create/Modify**: `pkg/models/reaction.go`, `pkg/reactions/handlers.go`, `pkg/comments/sqlite.go`, templates

**Dependencies**: Should be done AFTER #1 (CORS) and #2 (Rate Limiting)

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

### Issue #6: 🔒 Conduct Security Audit
**Priority**: Critical | **Effort**: Medium-Large (8-16 hours)

**Why It's Blocking**: No formal security review has been conducted.

**Requirements**:
- Review for SQL injection, XSS, CSRF, auth/authz issues
- Run automated security scanners (gosec, nancy/snyk)
- Manual testing of OWASP Top 10
- Document findings and fixes

**Success Criteria**:
- All high/critical vulnerabilities fixed
- Security.md created with security policy
- Automated scans pass
- Manual testing complete

**Files to Create/Modify**: `SECURITY.md` (new), `docs/security.md` (new), various code files

**Dependencies**: Should be done AFTER #1 (CORS) and #2 (Rate Limiting)

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
