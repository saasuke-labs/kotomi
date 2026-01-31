#!/bin/bash

# Script to create GitHub issues from Status.md
# This script creates issues for all unimplemented features mentioned in Status.md
# Each issue includes a reminder to update Status.md when completed

set -e

REPO="saasuke-labs/kotomi"

echo "Creating GitHub issues for Kotomi from Status.md..."
echo ""

# Check if gh is authenticated
if ! gh auth status > /dev/null 2>&1; then
    echo "❌ Error: gh CLI is not authenticated."
    echo "Please run: gh auth login"
    exit 1
fi

echo "✅ GitHub CLI is authenticated"
echo ""

# Issue 1: CORS Configuration (BLOCKING - Critical Priority)
echo "Creating Issue 1: CORS Configuration..."
gh issue create \
    --repo "$REPO" \
    --title "🚨 [BLOCKING] Implement CORS Configuration" \
    --label "priority:critical,blocking,enhancement" \
    --body "## Description
Cross-Origin Resource Sharing (CORS) headers are needed for the API to work with static sites hosted on different domains. This is a **BLOCKING** issue for production deployment.

## Current State
- No CORS middleware in the request pipeline
- Cannot configure allowed origins, methods, headers
- API requests from static sites on different domains will be blocked

## Requirements
1. Add CORS middleware using \`github.com/rs/cors\` or similar package
2. Configure via environment variables:
   - \`CORS_ALLOWED_ORIGINS\` (comma-separated list or \`*\` for development)
   - \`CORS_ALLOWED_METHODS\` (default: GET, POST, PUT, DELETE, OPTIONS)
   - \`CORS_ALLOWED_HEADERS\` (default: Content-Type, Authorization)
   - \`CORS_ALLOW_CREDENTIALS\` (default: true)
3. Apply middleware to all API routes (\`/api/*\`)
4. Do NOT apply CORS to admin routes (\`/admin/*\`)

## Implementation Details
- Add CORS package to \`go.mod\`
- Create CORS middleware configuration in \`cmd/main.go\` or \`pkg/middleware/cors.go\`
- Parse environment variables with sensible defaults
- Apply middleware before route handlers
- Test with different origins to ensure it works

## Testing
- [ ] Test API calls from localhost with different port
- [ ] Test API calls from a different domain
- [ ] Verify OPTIONS preflight requests work correctly
- [ ] Ensure admin panel is NOT affected by CORS

## Files to Modify
- \`cmd/main.go\` - Add CORS middleware to router
- \`go.mod\` / \`go.sum\` - Add CORS package dependency
- \`README.md\` - Document new environment variables

## Priority
**CRITICAL** - This is a blocking issue for production deployment.

## Estimated Effort
Small (2-4 hours)

## Dependencies
None - This can be implemented immediately.

## ⚠️ Important
**After completing this issue, please update Status.md:**
- Move CORS Configuration from \"❌ Not Implemented Features\" to \"✅ Fully Implemented Features\"
- Update the deployment readiness assessment
- Update the configuration requirements section"

echo "✅ Issue 1 created"
echo ""

# Issue 2: Rate Limiting (BLOCKING - Critical Priority)
echo "Creating Issue 2: Rate Limiting..."
gh issue create \
    --repo "$REPO" \
    --title "🚨 [BLOCKING] Implement Rate Limiting" \
    --label "priority:critical,blocking,security,enhancement" \
    --body "## Description
Rate limiting is essential to prevent spam and abuse. Without rate limiting, the service is vulnerable to spam attacks, DDoS, and API abuse. This is a **BLOCKING** issue for production deployment.

## Current State
- No rate limiting on any API endpoints
- Service is vulnerable to spam attacks and abuse
- No IP-based or user-based throttling
- Anyone can flood the API with requests

## Requirements
1. Add rate limiting middleware for all API endpoints
2. Implement different limits for different endpoint types:
   - **POST /api/site/{siteId}/page/{pageId}/comments**: 5 requests per minute per IP
   - **GET /api/site/{siteId}/page/{pageId}/comments**: 100 requests per minute per IP
   - **POST /api/site/{siteId}/page/{pageId}/comments/{commentId}/reactions**: 10 requests per minute per IP (when reactions are implemented)
3. Use in-memory storage for rate limit tracking (consider Redis for distributed deployments in future)
4. Return proper HTTP 429 (Too Many Requests) status code when limit exceeded
5. Include rate limit headers in responses:
   - \`X-RateLimit-Limit\`
   - \`X-RateLimit-Remaining\`
   - \`X-RateLimit-Reset\`

## Suggested Implementation
- Use \`golang.org/x/time/rate\` package or \`github.com/didip/tollbooth\`
- Create rate limiting middleware in \`pkg/middleware/ratelimit.go\`
- Track limits by IP address
- Configure via environment variables:
  - \`RATE_LIMIT_ENABLED\` (default: true)
  - \`RATE_LIMIT_COMMENTS_POST\` (default: 5 per minute)
  - \`RATE_LIMIT_COMMENTS_GET\` (default: 100 per minute)

## Implementation Details
1. Add rate limiting package to \`go.mod\`
2. Create middleware with configurable limits
3. Apply to API routes (not admin routes)
4. Add tests to verify rate limiting works
5. Document in README.md

## Testing
- [ ] Test that rate limits are enforced on POST requests
- [ ] Test that rate limits are enforced on GET requests
- [ ] Verify 429 status code is returned when limit exceeded
- [ ] Verify rate limit headers are present in responses
- [ ] Test that different IPs have independent limits
- [ ] Ensure admin panel is NOT rate limited

## Files to Modify
- \`pkg/middleware/ratelimit.go\` (new file)
- \`cmd/main.go\` - Apply rate limiting middleware
- \`go.mod\` / \`go.sum\` - Add rate limiting package
- \`README.md\` - Document rate limiting configuration

## Priority
**CRITICAL** - This is a blocking security issue for production deployment.

## Estimated Effort
Medium (4-8 hours)

## Dependencies
None - This can be implemented immediately.

## ⚠️ Important
**After completing this issue, please update Status.md:**
- Move Rate Limiting from \"❌ Not Implemented Features\" to \"✅ Fully Implemented Features\"
- Update the deployment readiness assessment
- Remove from \"Known Limitations & Issues\"
- Update the configuration requirements section"

echo "✅ Issue 2 created"
echo ""

# Issue 3: Reactions System (High Priority)
echo "Creating Issue 3: Reactions System..."
gh issue create \
    --repo "$REPO" \
    --title "⭐ Implement Reactions System" \
    --label "priority:high,enhancement,feature" \
    --body "## Description
Users should be able to react to comments with predefined reactions (e.g., like, love, clap). This is a core feature mentioned in the PRD and v0.2.0 roadmap.

## Current State
- No database schema for reactions
- No API endpoints for reactions
- No admin panel UI for configuring reactions
- Comments can only be replied to, not reacted to

## Requirements
1. **Database Schema**: Create reactions table with:
   - \`id\` (TEXT PRIMARY KEY)
   - \`comment_id\` (TEXT NOT NULL, FOREIGN KEY to comments)
   - \`reaction_type\` (TEXT NOT NULL) - e.g., 'like', 'love', 'clap', 'thinking'
   - \`author\` (TEXT NOT NULL) - User identifier (IP or authenticated user)
   - \`created_at\` (TIMESTAMP)
   - Unique constraint on (comment_id, reaction_type, author) to prevent duplicate reactions

2. **Site Configuration**: Create site_reactions table for site-specific reaction configuration:
   - \`id\` (TEXT PRIMARY KEY)
   - \`site_id\` (TEXT NOT NULL, FOREIGN KEY to sites)
   - \`reaction_type\` (TEXT NOT NULL) - e.g., 'like', 'love', 'clap'
   - \`display_name\` (TEXT NOT NULL) - Human-readable name
   - \`emoji\` (TEXT) - Optional emoji representation
   - \`enabled\` (BOOLEAN DEFAULT true)
   - \`sort_order\` (INTEGER) - Display order
   - Unique constraint on (site_id, reaction_type)

3. **API Endpoints**:
   - \`POST /api/site/{siteId}/page/{pageId}/comments/{commentId}/reactions\` - Add/toggle reaction
     - Request body: \`{\"reaction_type\": \"like\", \"author\": \"user@example.com\"}\`
     - Response: \`{\"success\": true, \"total\": 5}\`
   - \`GET /api/site/{siteId}/page/{pageId}/comments/{commentId}/reactions\` - Get reaction counts
     - Response: \`{\"like\": 5, \"love\": 2, \"clap\": 3}\`
   - \`DELETE /api/site/{siteId}/page/{pageId}/comments/{commentId}/reactions/{reactionType}\` - Remove reaction
     - Response: \`{\"success\": true, \"total\": 4}\`

4. **Admin Panel UI**:
   - Add reactions configuration section to site management page
   - Allow enabling/disabling specific reaction types per site
   - Allow customizing reaction display names and emojis
   - Default reactions: like (👍), love (❤️), clap (👏), thinking (🤔)

5. **Business Logic**:
   - One reaction type per author per comment (can change reaction)
   - Toggling: Adding same reaction removes it
   - Aggregate counts for display
   - Cascade delete: deleting comment removes all its reactions

## Implementation Steps
1. Create database migrations for reactions tables
2. Create models in \`pkg/models/reaction.go\`
3. Implement reaction storage in \`pkg/comments/sqlite.go\`
4. Create API handlers in \`pkg/reactions/handlers.go\` or extend \`cmd/main.go\`
5. Add admin UI for reaction configuration in \`pkg/admin/reactions.go\` and templates
6. Write comprehensive tests (unit, integration, E2E)
7. Update documentation

## Testing
- [ ] Unit tests for reaction storage/retrieval
- [ ] Test unique constraint (one reaction per user per comment)
- [ ] Test toggle behavior (add/remove same reaction)
- [ ] Test cascade delete (comment deletion removes reactions)
- [ ] E2E tests for all API endpoints
- [ ] Test site-specific reaction configuration
- [ ] Test admin UI for managing reactions

## Files to Create/Modify
- \`pkg/models/reaction.go\` (new)
- \`pkg/reactions/handlers.go\` (new) or extend \`cmd/main.go\`
- \`pkg/comments/sqlite.go\` - Add reaction methods
- \`pkg/admin/reactions.go\` (new)
- \`templates/admin/site_reactions.html\` (new)
- \`cmd/main.go\` - Add reaction routes
- \`README.md\` - Document reactions feature

## Priority
**HIGH** - This is a core feature mentioned in the PRD and affects v0.2.0 release.

## Estimated Effort
Medium (8-16 hours)

## Dependencies
- This issue should be completed AFTER #1 (CORS) and #2 (Rate Limiting) are done
- Consider implementing rate limiting for reaction endpoints as part of #2

## ⚠️ Important
**After completing this issue, please update Status.md:**
- Move Reactions System from \"❌ Not Implemented Features\" to \"✅ Fully Implemented Features\"
- Add detailed documentation of the reactions feature
- Update the deployment readiness assessment
- Update the \"Must-Have Before Deployment\" checklist"

echo "✅ Issue 3 created"
echo ""

# Issue 4: API Versioning (Medium Priority)
echo "Creating Issue 4: API Versioning..."
gh issue create \
    --repo "$REPO" \
    --title "🔧 Implement API Versioning" \
    --label "priority:medium,enhancement,breaking-change" \
    --body "## Description
API endpoints currently lack versioning. Breaking changes would affect all clients. Implementing API versioning is important for API stability and future evolution.

## Current State
- Endpoints like \`/api/site/{siteId}/...\` have no version prefix
- No versioning strategy in place
- Breaking changes would impact all API consumers simultaneously

## Requirements
1. Add version prefix to all API routes: \`/api/v1/site/{siteId}/...\`
2. Keep existing \`/api/...\` routes for backward compatibility (redirect or proxy to v1)
3. Document versioning strategy in README.md
4. Add deprecation headers for old endpoints
5. Plan for future versions (v2, v3, etc.)

## Implementation Details
1. Update all API routes in \`cmd/main.go\`:
   - \`/api/site/{siteId}/...\` → \`/api/v1/site/{siteId}/...\`
   - \`/api/healthz\` → \`/api/v1/healthz\`
2. Add route aliases for backward compatibility:
   - \`/api/...\` redirects to \`/api/v1/...\` with deprecation warning
3. Add response headers to old routes:
   - \`X-API-Deprecation-Warning: This endpoint is deprecated. Use /api/v1/... instead\`
   - \`X-API-Sunset-Date: 2026-12-31\` (example)
4. Update all documentation and examples
5. Update test files to use versioned endpoints

## Testing
- [ ] All existing tests pass with v1 endpoints
- [ ] Old endpoints still work (backward compatibility)
- [ ] Deprecation headers are present on old endpoints
- [ ] Update E2E tests to use versioned endpoints

## Files to Modify
- \`cmd/main.go\` - Update all API route definitions
- \`tests/e2e/api_test.go\` - Update test endpoints
- \`README.md\` - Document API versioning
- \`Status.md\` - Document versioning strategy
- Any example code or documentation

## Priority
**MEDIUM** - Important for API stability but not blocking deployment.

## Estimated Effort
Small (2-4 hours)

## Dependencies
- Should be implemented before Issue #5 (Frontend Widget) so the widget uses versioned endpoints from the start

## ⚠️ Important
**After completing this issue, please update Status.md:**
- Move API Versioning from \"❌ Not Implemented Features\" to \"✅ Fully Implemented Features\"
- Document the versioning strategy
- Update all endpoint examples to use /api/v1/..."

echo "✅ Issue 4 created"
echo ""

# Issue 5: Frontend Widget/JavaScript Embed (High Priority)
echo "Creating Issue 5: Frontend Widget..."
gh issue create \
    --repo "$REPO" \
    --title "🎨 Create Frontend Widget / JavaScript Embed" \
    --label "priority:high,enhancement,feature,frontend" \
    --body "## Description
JavaScript widget for easy integration of Kotomi comments and reactions into static sites. Currently, site owners must build their own frontend integration, which is a significant barrier to adoption.

## Current State
- No JavaScript SDK/library for embedding comments
- No HTML snippet for easy integration
- No UI components for displaying comments on static sites
- Site owners must manually call API endpoints and render UI

## Requirements

### Core Features
1. **JavaScript SDK** (\`kotomi.js\`):
   - Simple initialization: \`Kotomi.init({siteId: 'xxx', pageId: 'yyy'})\`
   - Load comments from API
   - Display comments in a customizable UI
   - Handle comment submission
   - Display and handle reactions (when implemented)
   - Responsive design (mobile-friendly)
   - Accessible (ARIA labels, keyboard navigation)

2. **Simple Embed**:
   \`\`\`html
   <div id=\"kotomi-comments\"></div>
   <script src=\"https://kotomi.example.com/widget/kotomi.js\"></script>
   <script>
     Kotomi.init({
       siteId: 'your-site-id',
       pageId: 'blog/my-post',
       element: '#kotomi-comments'
     });
   </script>
   \`\`\`

3. **Configuration Options**:
   - \`apiUrl\` - Kotomi API base URL
   - \`siteId\` - Site identifier
   - \`pageId\` - Page identifier (URL path)
   - \`element\` - DOM selector for mount point
   - \`theme\` - 'light' or 'dark'
   - \`maxComments\` - Pagination limit
   - \`sortBy\` - 'newest' or 'oldest'
   - \`enableReactions\` - true/false (when reactions implemented)

### UI Components
1. **Comment List**:
   - Display approved comments
   - Nested/threaded replies
   - Author name and timestamp
   - Reaction counts (when implemented)
   - Reply button
   - Pagination/load more

2. **Comment Form**:
   - Author name input
   - Comment text textarea
   - Submit button
   - Loading state
   - Success/error messages

3. **Styling**:
   - Clean, minimal default theme
   - CSS variables for customization
   - Responsive design
   - Dark mode support

### Technical Details
1. Build JavaScript SDK with vanilla JS (no framework dependencies)
2. Bundle with webpack/rollup for single-file distribution
3. Minified and unminified versions
4. Source maps for debugging
5. TypeScript types included
6. CDN-ready (can be served from static directory)

## Implementation Steps
1. Create \`frontend/\` directory for widget code
2. Set up build system (webpack/rollup)
3. Implement core JavaScript SDK
4. Create UI components (HTML/CSS)
5. Add API integration layer
6. Create example HTML pages
7. Write documentation
8. Build and bundle for distribution
9. Serve from \`/widget/kotomi.js\` endpoint

## Testing
- [ ] Test widget initialization
- [ ] Test comment loading and display
- [ ] Test comment submission
- [ ] Test nested replies
- [ ] Test error handling
- [ ] Cross-browser testing (Chrome, Firefox, Safari, Edge)
- [ ] Mobile responsiveness
- [ ] Accessibility testing
- [ ] Test with different themes

## Files to Create
- \`frontend/src/kotomi.js\` - Main SDK
- \`frontend/src/ui.js\` - UI components
- \`frontend/src/api.js\` - API client
- \`frontend/src/styles.css\` - Default styles
- \`frontend/webpack.config.js\` - Build configuration
- \`frontend/examples/simple.html\` - Example usage
- \`static/widget/kotomi.js\` - Built widget (gitignored, built during deployment)
- \`static/widget/kotomi.min.js\` - Minified version
- \`docs/widget.md\` - Widget documentation

## Files to Modify
- \`cmd/main.go\` - Add route to serve widget files
- \`README.md\` - Add widget usage documentation
- \`.gitignore\` - Ignore built widget files

## Priority
**HIGH** - Essential for end-users to integrate Kotomi without custom development.

## Estimated Effort
Large (16-24 hours)

## Dependencies
- Should be implemented AFTER Issue #1 (CORS) is complete
- Should use Issue #4 (API Versioning) endpoints if available
- Will need updates when Issue #3 (Reactions) is implemented

## ⚠️ Important
**After completing this issue, please update Status.md:**
- Move Frontend Widget from \"❌ Not Implemented Features\" to \"✅ Fully Implemented Features\"
- Add comprehensive documentation of widget features and usage
- Update deployment requirements
- Update README.md with widget examples"

echo "✅ Issue 5 created"
echo ""

# Issue 6: Security Audit (Critical Priority)
echo "Creating Issue 6: Security Audit..."
gh issue create \
    --repo "$REPO" \
    --title "🔒 Conduct Security Audit" \
    --label "priority:critical,blocking,security" \
    --body "## Description
No formal security review has been conducted. Before production deployment, a comprehensive security audit is required to identify and fix vulnerabilities.

## Current State
- Code has not been formally reviewed for security vulnerabilities
- Some security measures are in place (Auth0, session encryption) but need validation
- Security testing has not been conducted

## Requirements

### Code Review Areas
1. **SQL Injection**:
   - [ ] Verify all database queries use prepared statements
   - [ ] Check \`pkg/comments/sqlite.go\` for query construction
   - [ ] Review all raw SQL queries

2. **XSS (Cross-Site Scripting)**:
   - [ ] Verify HTML templates properly escape user input
   - [ ] Check all template files in \`templates/\`
   - [ ] Review comment display in admin panel
   - [ ] Test with malicious payloads

3. **CSRF (Cross-Site Request Forgery)**:
   - [ ] Verify CSRF protection on admin forms
   - [ ] Check if gorilla/csrf or similar is used
   - [ ] Test admin panel forms

4. **Authentication & Authorization**:
   - [ ] Review Auth0 integration for best practices
   - [ ] Verify session security (encryption, httpOnly, secure flags)
   - [ ] Check that admin routes require authentication
   - [ ] Test authorization bypass attempts

5. **Input Validation**:
   - [ ] Review all API input validation
   - [ ] Check for buffer overflow vulnerabilities
   - [ ] Test with oversized inputs
   - [ ] Validate all user-supplied data

6. **Session Management**:
   - [ ] Review session configuration
   - [ ] Check session timeout settings
   - [ ] Verify secure cookie flags (httpOnly, secure, sameSite)
   - [ ] Test session fixation attacks

7. **Rate Limiting** (after Issue #2):
   - [ ] Verify rate limiting is properly implemented
   - [ ] Test bypass attempts

8. **CORS Configuration** (after Issue #1):
   - [ ] Review CORS settings
   - [ ] Ensure admin routes are not exposed to CORS
   - [ ] Test cross-origin attacks

9. **Error Handling**:
   - [ ] Verify errors don't leak sensitive information
   - [ ] Check stack traces in production
   - [ ] Review logging for sensitive data

10. **Dependencies**:
    - [ ] Run \`go list -m all | go list -m -json all | nancy sleuth\` or similar
    - [ ] Check for known vulnerabilities in dependencies
    - [ ] Update vulnerable packages

### Security Testing
1. **Automated Scanning**:
   - Run \`gosec\` for static analysis
   - Run \`go-critic\` for code quality
   - Use \`nancy\` or \`snyk\` for dependency vulnerabilities

2. **Manual Testing**:
   - Test common OWASP Top 10 vulnerabilities
   - Attempt SQL injection on all inputs
   - Test XSS with various payloads
   - Test authentication bypass
   - Test authorization bypass

3. **Penetration Testing**:
   - Consider hiring external security consultant
   - Or conduct internal penetration testing
   - Document findings

## Implementation Steps
1. Install security scanning tools
2. Run automated scans
3. Review and fix findings
4. Conduct manual testing
5. Document all findings and fixes
6. Create security.md with security best practices
7. Add security section to README.md

## Tools to Use
- \`gosec\` - Go security scanner
- \`go-critic\` - Go linter with security checks
- \`nancy\` or \`snyk\` - Dependency vulnerability scanner
- OWASP ZAP - Web application security testing

## Files to Create/Modify
- \`SECURITY.md\` (new) - Security policy and vulnerability reporting
- \`docs/security.md\` (new) - Security documentation
- \`README.md\` - Add security section
- Various code files - Fix identified vulnerabilities

## Priority
**CRITICAL** - This is a blocking issue for production deployment.

## Estimated Effort
Medium to Large (8-16 hours depending on findings)

## Dependencies
- Should be done AFTER Issue #1 (CORS) and Issue #2 (Rate Limiting)
- Can be done in parallel with other features

## ⚠️ Important
**After completing this issue, please update Status.md:**
- Update \"Testing Status\" section to include security tests
- Document security measures in place
- Update deployment readiness assessment
- Add security audit results summary"

echo "✅ Issue 6 created"
echo ""

# Issue 7: Automatic/AI Moderation (Medium Priority)
echo "Creating Issue 7: Automatic/AI Moderation..."
gh issue create \
    --repo "$REPO" \
    --title "🤖 Implement Automatic/AI Moderation" \
    --label "priority:medium,enhancement,feature,ai" \
    --body "## Description
Automatic moderation using AI to flag spam, offensive language, or off-topic comments. This would reduce the manual moderation burden for site owners.

## Current State
- Only manual moderation is available
- All comments must be reviewed manually
- No automatic spam detection
- No automatic content filtering

## Requirements

### Core Features
1. **AI Integration**:
   - Integrate with OpenAI GPT API for content analysis
   - Configurable via environment variables
   - Optional feature (can be disabled)

2. **Moderation Categories**:
   - **Spam**: Commercial spam, repetitive content
   - **Offensive**: Profanity, hate speech, harassment
   - **Aggressive**: Personal attacks, threatening language
   - **Off-topic**: Content unrelated to discussion
   
3. **Confidence Scoring**:
   - High confidence (>90%): Auto-reject
   - Medium confidence (50-90%): Flag for review with yellow indicator
   - Low confidence (<50%): Auto-approve

4. **Admin Controls**:
   - Enable/disable AI moderation per site
   - Configure confidence thresholds
   - Configure auto-reject vs. flag-for-review
   - View AI moderation decisions and rationale

### Implementation Details
1. **Environment Variables**:
   - \`OPENAI_API_KEY\` - OpenAI API key
   - \`AI_MODERATION_ENABLED\` - Enable/disable globally (default: false)
   - \`AI_MODERATION_AUTO_REJECT_THRESHOLD\` - Auto-reject threshold (default: 0.9)
   - \`AI_MODERATION_FLAG_THRESHOLD\` - Flag threshold (default: 0.5)

2. **Database Schema Updates**:
   Add to comments table:
   - \`ai_moderation_score\` (REAL) - Confidence score 0-1
   - \`ai_moderation_category\` (TEXT) - 'spam', 'offensive', 'aggressive', 'off-topic', 'clean'
   - \`ai_moderation_reason\` (TEXT) - AI's explanation

   Add site_settings table:
   - \`site_id\` (TEXT PRIMARY KEY)
   - \`ai_moderation_enabled\` (BOOLEAN)
   - \`auto_reject_threshold\` (REAL)
   - \`flag_threshold\` (REAL)

3. **API Integration**:
   - Create \`pkg/moderation/ai.go\` with OpenAI integration
   - Send comment text to OpenAI for analysis
   - Parse response and extract category and confidence
   - Store results in database

4. **Workflow**:
   - When comment is submitted, run AI moderation if enabled
   - Apply automatic actions based on thresholds
   - Store AI decision for admin review
   - Allow admin to override AI decisions

### Testing
- [ ] Unit tests for AI moderation logic
- [ ] Mock OpenAI API for testing
- [ ] Test different comment types (spam, offensive, clean)
- [ ] Test threshold configurations
- [ ] Test admin override functionality
- [ ] Integration tests with database

## Files to Create/Modify
- \`pkg/moderation/ai.go\` (new) - AI moderation logic
- \`pkg/moderation/openai.go\` (new) - OpenAI API client
- \`pkg/comments/sqlite.go\` - Add AI moderation fields
- \`pkg/admin/comments.go\` - Display AI moderation info
- \`templates/admin/comments.html\` - Show AI flags
- \`cmd/main.go\` - Integrate AI moderation in comment creation
- \`README.md\` - Document AI moderation feature
- \`go.mod\` - Add OpenAI SDK dependency

## Priority
**MEDIUM** - Nice to have, but manual moderation works for now.

## Estimated Effort
Large (16-24 hours)

## Dependencies
- Can be implemented independently
- Consider implementing after core features are stable

## Cost Considerations
- OpenAI API calls have costs ($0.0015-0.002 per 1K tokens)
- Estimate ~500 tokens per comment analysis
- Cost per 1000 comments: ~$0.75-$1.00
- Document costs in README

## ⚠️ Important
**After completing this issue, please update Status.md:**
- Move Automatic/AI Moderation from \"❌ Not Implemented Features\" to \"✅ Fully Implemented Features\"
- Document AI moderation feature and configuration
- Add cost considerations to documentation
- Update configuration requirements"

echo "✅ Issue 7 created"
echo ""

# Issue 8: User Authentication for Comments (Medium Priority)
echo "Creating Issue 8: User Authentication for Comments..."
gh issue create \
    --repo "$REPO" \
    --title "👤 Implement User Authentication for Comments" \
    --label "priority:medium,enhancement,feature" \
    --body "## Description
End-users (commenters) currently cannot authenticate. All comments are anonymous. Implementing user authentication would allow users to edit/delete their own comments and build user profiles.

## Current State
- Only admin panel has authentication
- Public API accepts any author name (no verification)
- No way to track authenticated users vs. guests
- Users cannot edit or delete their own comments
- No user profiles

## Requirements

### Core Features
1. **Authentication Options**:
   - Email/password authentication
   - Social logins (Google, GitHub, Twitter)
   - Guest/anonymous posting (optional per site)
   - Remember me functionality

2. **User Management**:
   - User registration and login
   - Password reset flow
   - Email verification
   - User profiles (display name, avatar, bio)
   - User preferences

3. **Comment Ownership**:
   - Link comments to authenticated users
   - Show verified badge for authenticated users
   - Allow users to edit their own comments (within time limit)
   - Allow users to delete their own comments

4. **Admin Controls**:
   - Configure whether to require authentication per site
   - Allow/disallow anonymous comments
   - Manage user accounts
   - Ban/suspend users

### Database Schema Updates
1. **Public Users Table** (separate from admin users):
   - \`id\` (TEXT PRIMARY KEY)
   - \`email\` (TEXT UNIQUE)
   - \`password_hash\` (TEXT)
   - \`display_name\` (TEXT)
   - \`avatar_url\` (TEXT)
   - \`bio\` (TEXT)
   - \`email_verified\` (BOOLEAN)
   - \`created_at\` (TIMESTAMP)
   - \`last_login\` (TIMESTAMP)

2. **Comments Table Updates**:
   - \`user_id\` (TEXT, FOREIGN KEY to public_users, nullable)
   - Keep \`author\` field for anonymous comments
   - \`is_authenticated\` (BOOLEAN)

3. **Site Settings Updates**:
   - \`require_authentication\` (BOOLEAN)
   - \`allow_anonymous\` (BOOLEAN)
   - \`allow_guest_edit\` (BOOLEAN)

### API Endpoints
- \`POST /api/auth/register\` - Register new user
- \`POST /api/auth/login\` - Login
- \`POST /api/auth/logout\` - Logout
- \`POST /api/auth/forgot-password\` - Request password reset
- \`POST /api/auth/reset-password\` - Reset password
- \`GET /api/auth/verify-email\` - Verify email
- \`GET /api/users/me\` - Get current user profile
- \`PUT /api/users/me\` - Update user profile
- \`PUT /api/site/{siteId}/page/{pageId}/comments/{commentId}\` - Edit own comment
- \`DELETE /api/site/{siteId}/page/{pageId}/comments/{commentId}\` - Delete own comment

### Security Considerations
- Hash passwords with bcrypt
- Implement JWT or session-based auth
- Rate limit login attempts
- Email verification required
- Password strength requirements
- CSRF protection

## Implementation Steps
1. Design authentication architecture
2. Create public users database schema
3. Implement authentication endpoints
4. Create password hashing and verification
5. Implement JWT/session management
6. Add email verification system
7. Update comment creation to support authenticated users
8. Add edit/delete comment functionality
9. Update admin panel to show user info
10. Update frontend widget to support authentication

## Testing
- [ ] Test user registration
- [ ] Test login/logout
- [ ] Test password reset flow
- [ ] Test email verification
- [ ] Test authenticated comment creation
- [ ] Test edit own comment
- [ ] Test delete own comment
- [ ] Test authorization (cannot edit others' comments)
- [ ] Security testing (password strength, SQL injection, etc.)

## Files to Create/Modify
- \`pkg/auth/public_auth.go\` (new) - Public user authentication
- \`pkg/models/public_user.go\` (new) - Public user model
- \`pkg/comments/sqlite.go\` - Update comment methods
- \`cmd/main.go\` - Add authentication routes
- \`README.md\` - Document authentication
- Frontend widget - Add authentication UI

## Priority
**MEDIUM** - Depends on use case. Many comment systems allow anonymous posting.

## Estimated Effort
Large (24-40 hours)

## Dependencies
- Should be implemented AFTER Issue #5 (Frontend Widget)
- Requires email service for verification (consider SendGrid, AWS SES)

## Design Decisions
- Should public users share Auth0 with admins, or separate system?
- Should we support social logins?
- Time limit for editing comments (e.g., 5 minutes)?
- Soft delete vs. hard delete for user comments?

## ⚠️ Important
**After completing this issue, please update Status.md:**
- Move User Authentication for Comments from \"❌ Not Implemented Features\" to \"✅ Fully Implemented Features\"
- Document authentication system
- Update API documentation
- Add security considerations"

echo "✅ Issue 8 created"
echo ""

# Issue 9: Email Notifications (Low Priority)
echo "Creating Issue 9: Email Notifications..."
gh issue create \
    --repo "$REPO" \
    --title "📧 Implement Email Notifications" \
    --label "priority:low,enhancement,feature" \
    --body "## Description
Email notifications for site owners and users when comments are posted, moderated, or replied to.

## Current State
- No email notifications
- Site owners must manually check for new comments
- Users don't know when someone replies to their comments

## Requirements

### Notification Types
1. **For Site Owners**:
   - New comment pending moderation
   - Comment approved/rejected (if user provided email)
   - Spam detected (if AI moderation enabled)
   
2. **For Users** (when auth is implemented):
   - Someone replied to your comment
   - Your comment was moderated (approved/rejected)
   - Digest of activity on pages you commented on

### Configuration
- Enable/disable notifications per site
- Configure notification frequency (immediate, daily digest, weekly digest)
- Configure which events trigger notifications
- Unsubscribe links in all emails

### Email Service Integration
- Support multiple email providers:
  - SendGrid (recommended)
  - AWS SES
  - Mailgun
  - SMTP

### Environment Variables
- \`EMAIL_PROVIDER\` - 'sendgrid', 'ses', 'mailgun', 'smtp'
- \`EMAIL_API_KEY\` - API key for provider
- \`EMAIL_FROM\` - From email address
- \`EMAIL_FROM_NAME\` - From name
- For SMTP:
  - \`SMTP_HOST\`
  - \`SMTP_PORT\`
  - \`SMTP_USERNAME\`
  - \`SMTP_PASSWORD\`

## Implementation Details

### Database Updates
1. Add to users table:
   - \`notification_email\` (TEXT)
   - \`notification_preferences\` (JSON)
   
2. Create notifications table:
   - \`id\` (TEXT PRIMARY KEY)
   - \`user_id\` (TEXT, FOREIGN KEY)
   - \`type\` (TEXT) - 'new_comment', 'reply', 'moderation'
   - \`subject\` (TEXT)
   - \`body\` (TEXT)
   - \`sent\` (BOOLEAN)
   - \`sent_at\` (TIMESTAMP)
   - \`created_at\` (TIMESTAMP)

### Email Templates
Create HTML email templates:
- \`templates/email/new_comment.html\`
- \`templates/email/reply_notification.html\`
- \`templates/email/moderation_update.html\`
- \`templates/email/digest.html\`

### Implementation
- Create \`pkg/notifications/email.go\`
- Implement email service abstraction
- Create template rendering system
- Add background job for sending emails (queue system)
- Add retry logic for failed sends

## Testing
- [ ] Test email sending with each provider
- [ ] Test email templates render correctly
- [ ] Test unsubscribe links work
- [ ] Test notification preferences
- [ ] Test email queue and retry logic

## Files to Create/Modify
- \`pkg/notifications/email.go\` (new)
- \`pkg/notifications/queue.go\` (new)
- \`templates/email/\` (new directory)
- \`pkg/admin/notifications.go\` (new) - Admin UI for notification settings
- \`cmd/main.go\` - Initialize notification system
- \`README.md\` - Document notification configuration

## Priority
**LOW** - Nice to have, but not essential for initial launch.

## Estimated Effort
Medium (12-16 hours)

## Dependencies
- Should be implemented after core features are stable
- Consider implementing after Issue #8 (User Authentication)

## Cost Considerations
- Email services have costs (SendGrid: $14.95/month for 40K emails)
- Document costs in README

## ⚠️ Important
**After completing this issue, please update Status.md:**
- Move Email Notifications from \"❌ Not Implemented Features\" to \"✅ Fully Implemented Features\"
- Document notification system
- Add cost considerations
- Update configuration requirements"

echo "✅ Issue 9 created"
echo ""

# Issue 10: Analytics & Reporting (Low Priority)
echo "Creating Issue 10: Analytics & Reporting..."
gh issue create \
    --repo "$REPO" \
    --title "📊 Implement Analytics & Reporting" \
    --label "priority:low,enhancement,feature" \
    --body "## Description
Analytics dashboard for site owners to track engagement, comment activity, and user interactions.

## Current State
- No analytics or reporting features
- Site owners cannot see trends or metrics
- No visibility into engagement patterns

## Requirements from PRD
- Comment counts
- Active users
- Reaction breakdowns (when reactions implemented)
- Engagement trends over time

### Analytics to Track
1. **Comment Metrics**:
   - Total comments per site/page
   - Comments over time (daily, weekly, monthly)
   - Average comments per page
   - Comment approval rate
   - Comment rejection rate
   - Pending moderation count

2. **User Metrics** (when auth implemented):
   - Unique commenters
   - Active users (commented in last 30 days)
   - Top contributors
   - User growth over time

3. **Reaction Metrics** (when reactions implemented):
   - Total reactions per site/page
   - Reaction type breakdown
   - Most reacted comments

4. **Moderation Metrics**:
   - Average moderation time
   - Moderation actions per day
   - Spam detection rate (if AI enabled)

5. **Engagement Metrics**:
   - Pages with most comments
   - Pages with most reactions
   - Average time to first comment
   - Reply rate (% of comments that get replies)

### API Endpoints
- \`GET /api/site/{siteId}/analytics\` - Overall site analytics
- \`GET /api/site/{siteId}/analytics/comments\` - Comment trends
- \`GET /api/site/{siteId}/analytics/users\` - User metrics
- \`GET /api/site/{siteId}/analytics/reactions\` - Reaction metrics
- \`GET /api/site/{siteId}/analytics/moderation\` - Moderation metrics

### Admin Panel UI
- Dashboard with key metrics and charts
- Date range selector
- Export reports as CSV
- Visualizations:
  - Line charts for trends over time
  - Pie charts for breakdowns
  - Tables for top pages/users

## Implementation Details

### Database
- Create analytics views/materialized views for performance
- Pre-calculate common metrics
- Add indexes for analytics queries

### Technology
- Backend: Go with analytics calculations
- Frontend: Chart.js or similar for visualizations
- Consider time-series database for high-volume sites

### Caching
- Cache analytics results (e.g., 1 hour TTL)
- Recalculate when new data arrives
- Use Redis for caching if available

## Implementation Steps
1. Design analytics database schema
2. Create analytics calculation functions
3. Implement API endpoints
4. Create admin panel UI with charts
5. Add export functionality
6. Implement caching
7. Add tests

## Testing
- [ ] Test analytics calculations
- [ ] Test date range filtering
- [ ] Test chart rendering
- [ ] Test export functionality
- [ ] Performance test with large datasets
- [ ] Test caching

## Files to Create/Modify
- \`pkg/analytics/metrics.go\` (new)
- \`pkg/analytics/calculations.go\` (new)
- \`pkg/admin/analytics.go\` (new)
- \`templates/admin/analytics.html\` (new)
- \`static/js/charts.js\` (new)
- \`cmd/main.go\` - Add analytics routes
- \`README.md\` - Document analytics feature

## Priority
**LOW** - Nice to have, can be added after launch.

## Estimated Effort
Medium (12-16 hours)

## Dependencies
- Should be implemented after core features are stable
- Will be more valuable after Issue #3 (Reactions) and Issue #8 (User Auth)

## ⚠️ Important
**After completing this issue, please update Status.md:**
- Move Analytics & Reporting from \"❌ Not Implemented Features\" to \"✅ Fully Implemented Features\"
- Document analytics features
- Add example screenshots to documentation"

echo "✅ Issue 10 created"
echo ""

# Issue 11: Export/Import Functionality (Low Priority)
echo "Creating Issue 11: Export/Import Functionality..."
gh issue create \
    --repo "$REPO" \
    --title "💾 Implement Export/Import Functionality" \
    --label "priority:low,enhancement,feature" \
    --body "## Description
Ability to export and import comments, reactions, and site data in JSON or CSV format for data portability and backup purposes.

## Current State
- No export functionality
- No import functionality
- No easy way to backup or migrate data
- Data portability is limited

## Requirements

### Export Features
1. **Export Comments**:
   - Export all comments for a site or specific page
   - Format: JSON or CSV
   - Include metadata (author, timestamp, status, moderation info)
   - Include nested structure for threaded comments

2. **Export Reactions** (when implemented):
   - Export all reactions for a site or page
   - Include reaction counts and types

3. **Export Full Site Data**:
   - All pages, comments, and reactions for a site
   - Complete backup that can be restored

### Import Features
1. **Import Comments**:
   - Import from JSON or CSV
   - Map fields from external sources
   - Handle duplicates (skip or update)
   - Validate data before import

2. **Import from Other Platforms**:
   - Disqus export format
   - WordPress comments export
   - Generic JSON/CSV format

3. **Conflict Resolution**:
   - Skip duplicates
   - Update existing
   - Create new with different ID

### API Endpoints
- \`GET /api/site/{siteId}/export/comments?format=json|csv\` - Export comments
- \`GET /api/site/{siteId}/export/full\` - Export full site data
- \`POST /api/site/{siteId}/import/comments\` - Import comments
- \`POST /api/site/{siteId}/import/validate\` - Validate import file before importing

### Admin Panel UI
- Export button on site management page
- Import wizard with file upload
- Preview import data before committing
- Show import progress and results
- Error handling for invalid data

### File Formats

**JSON Export Example**:
\`\`\`json
{
  \"site_id\": \"my-blog\",
  \"exported_at\": \"2026-01-31T12:00:00Z\",
  \"version\": \"1.0\",
  \"comments\": [
    {
      \"id\": \"comment-1\",
      \"page_id\": \"blog/post-1\",
      \"author\": \"John Doe\",
      \"text\": \"Great post!\",
      \"parent_id\": null,
      \"status\": \"approved\",
      \"created_at\": \"2026-01-30T10:00:00Z\"
    }
  ]
}
\`\`\`

**CSV Export Example**:
\`\`\`csv
id,page_id,author,text,parent_id,status,created_at
comment-1,blog/post-1,John Doe,Great post!,,approved,2026-01-30T10:00:00Z
\`\`\`

## Implementation Details

### Security
- Require authentication for export/import
- Validate uploaded files (file size, format)
- Sanitize imported data
- Rate limit export/import operations

### Performance
- Stream large exports (don't load all in memory)
- Show progress for large imports
- Handle timeouts for large operations
- Consider background jobs for very large exports/imports

## Implementation Steps
1. Create export service in \`pkg/export/\`
2. Create import service in \`pkg/import/\`
3. Implement JSON export/import
4. Implement CSV export/import
5. Add API endpoints
6. Create admin UI
7. Add validation and error handling
8. Write tests

## Testing
- [ ] Test JSON export
- [ ] Test CSV export
- [ ] Test import with valid data
- [ ] Test import with invalid data
- [ ] Test duplicate handling
- [ ] Test large file handling
- [ ] Test error cases

## Files to Create/Modify
- \`pkg/export/export.go\` (new)
- \`pkg/export/json.go\` (new)
- \`pkg/export/csv.go\` (new)
- \`pkg/import/import.go\` (new)
- \`pkg/import/validate.go\` (new)
- \`pkg/admin/export.go\` (new)
- \`templates/admin/export.html\` (new)
- \`templates/admin/import.html\` (new)
- \`cmd/main.go\` - Add export/import routes
- \`README.md\` - Document export/import features

## Priority
**LOW** - Nice to have for data portability, but not essential for launch.

## Estimated Effort
Small to Medium (8-12 hours)

## Dependencies
- Can be implemented independently
- Consider implementing after core features are stable

## ⚠️ Important
**After completing this issue, please update Status.md:**
- Move Export/Import Functionality from \"❌ Not Implemented Features\" to \"✅ Fully Implemented Features\"
- Document export/import formats and usage
- Add examples to documentation"

echo "✅ Issue 11 created"
echo ""

# Issue 12: Error Handling & Logging Improvements (Medium Priority)
echo "Creating Issue 12: Error Handling & Logging..."
gh issue create \
    --repo "$REPO" \
    --title "🔍 Improve Error Handling & Logging" \
    --label "priority:medium,enhancement,observability" \
    --body "## Description
Improve error handling, logging, and observability for production debugging. Current implementation has limited error messages and no structured logging.

## Current State (from Status.md)
- Some endpoints may not have comprehensive error handling
- Limited error messages - some errors return generic HTTP 500 without details
- No structured logging for production debugging
- HTTP requests are not logged
- No request tracing

## Requirements

### Error Handling
1. **Consistent Error Responses**:
   - Standardized error response format
   - Include error code, message, and optional details
   - Don't expose internal errors to clients
   - Log detailed errors server-side

   Example error response:
   \`\`\`json
   {
     \"error\": {
       \"code\": \"INVALID_INPUT\",
       \"message\": \"Comment text cannot be empty\",
       \"details\": {
         \"field\": \"text\"
       }
     }
   }
   \`\`\`

2. **Error Codes**:
   - Define error code constants
   - Use specific HTTP status codes correctly:
     - 400 - Bad Request (invalid input)
     - 401 - Unauthorized (not authenticated)
     - 403 - Forbidden (not authorized)
     - 404 - Not Found
     - 409 - Conflict (duplicate, constraint violation)
     - 429 - Too Many Requests
     - 500 - Internal Server Error
     - 503 - Service Unavailable

3. **Input Validation**:
   - Validate all API inputs
   - Return detailed validation errors
   - Check required fields
   - Validate field formats and lengths

### Logging
1. **Structured Logging**:
   - Use structured logging library (e.g., \`go.uber.org/zap\` or \`logrus\`)
   - Include context in logs (request ID, user ID, etc.)
   - Log levels: DEBUG, INFO, WARN, ERROR
   - JSON format for production

2. **What to Log**:
   - HTTP requests (method, path, status, duration)
   - Database queries (for debugging)
   - Authentication events
   - Moderation actions
   - Errors with stack traces
   - Performance metrics

3. **What NOT to Log**:
   - Passwords or secrets
   - Session tokens
   - Personal information (unless necessary)
   - Credit card data

4. **Request Tracing**:
   - Add request ID to all logs
   - Include request ID in error responses
   - Trace requests across services

### Configuration
- \`LOG_LEVEL\` - Log level (debug, info, warn, error)
- \`LOG_FORMAT\` - Log format (json, text)
- \`LOG_OUTPUT\` - Log output (stdout, file path)

## Implementation Details

### Error Handling Package
Create \`pkg/errors/errors.go\`:
\`\`\`go
type AppError struct {
    Code    string
    Message string
    Details map[string]interface{}
    Err     error
}

func NewBadRequestError(message string, details map[string]interface{}) *AppError
func NewNotFoundError(message string) *AppError
func NewInternalError(err error) *AppError
\`\`\`

### Logging Middleware
Create \`pkg/middleware/logging.go\`:
- Log all HTTP requests
- Include request/response details
- Calculate request duration
- Add request ID to context

### Implementation Steps
1. Choose logging library (recommend zap)
2. Create error handling package
3. Create logging middleware
4. Update all handlers to use new error handling
5. Add input validation to all endpoints
6. Update error responses to use consistent format
7. Add tests

## Testing
- [ ] Test error responses have consistent format
- [ ] Test correct HTTP status codes are returned
- [ ] Test input validation
- [ ] Test logging middleware
- [ ] Test structured logs are properly formatted
- [ ] Verify sensitive data is not logged

## Files to Create/Modify
- \`pkg/errors/errors.go\` (new) - Error types
- \`pkg/middleware/logging.go\` (new) - Logging middleware
- \`pkg/middleware/requestid.go\` (new) - Request ID middleware
- \`cmd/main.go\` - Initialize logging, add middleware
- \`pkg/admin/*.go\` - Update error handling
- \`go.mod\` - Add logging library
- All API handlers - Update error handling
- \`README.md\` - Document error codes and logging

## Priority
**MEDIUM** - Important for production operations but not blocking.

## Estimated Effort
Medium (8-12 hours)

## Dependencies
- Should be implemented before production deployment
- Can be done in parallel with other features

## ⚠️ Important
**After completing this issue, please update Status.md:**
- Update \"Important Issues\" section to mark Error Handling as complete
- Update \"Known Limitations & Issues\" to remove error handling items
- Document error codes and logging configuration
- Add observability section to documentation"

echo "✅ Issue 12 created"
echo ""

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "✅ All GitHub issues created successfully!"
echo ""
echo "Summary:"
echo "  1. 🚨 [BLOCKING] Implement CORS Configuration (Critical)"
echo "  2. 🚨 [BLOCKING] Implement Rate Limiting (Critical, Security)"
echo "  3. ⭐ Implement Reactions System (High Priority)"
echo "  4. 🔧 Implement API Versioning (Medium Priority)"
echo "  5. 🎨 Create Frontend Widget / JavaScript Embed (High Priority)"
echo "  6. 🔒 Conduct Security Audit (Critical, Blocking)"
echo "  7. 🤖 Implement Automatic/AI Moderation (Medium Priority)"
echo "  8. 👤 Implement User Authentication for Comments (Medium Priority)"
echo "  9. 📧 Implement Email Notifications (Low Priority)"
echo " 10. 📊 Implement Analytics & Reporting (Low Priority)"
echo " 11. 💾 Implement Export/Import Functionality (Low Priority)"
echo " 12. 🔍 Improve Error Handling & Logging (Medium Priority)"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "⚠️  IMPORTANT NOTES:"
echo "  • Issues #1 (CORS) and #2 (Rate Limiting) are BLOCKING for production"
echo "  • Issue #6 (Security Audit) should be done after #1 and #2"
echo "  • Issue #5 (Frontend Widget) depends on #1 (CORS)"
echo "  • Issue #3 (Reactions) should be done after #1 and #2"
echo "  • All issues include reminders to update Status.md when completed"
echo ""
echo "Next steps:"
echo "  1. Review the created issues on GitHub"
echo "  2. Prioritize and assign issues to team/agents"
echo "  3. Start with blocking issues (#1, #2, #6)"
echo "  4. Update Status.md as issues are completed"
