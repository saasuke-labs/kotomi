# Kotomi – Product Requirements Document (PRD)

## 1. Introduction
Kotomi is a dynamic content service for static websites. It provides features such as comments, reactions, and moderation, while ensuring ease of integration with existing static site generators and workflows. Kotomi aims to deliver a lightweight, privacy-respecting, and developer-friendly solution.

## 2. Goals
- Provide a simple way for static sites to support interactive content (comments, reactions, etc.).
- Support moderation workflows (manual, automatic, AI-assisted).
- Minimize the impact on site performance and end-user experience.
- Ensure developer flexibility (works with multiple SSGs and deployment styles).
- Offer a scalable model for both small personal sites and larger communities.

## 3. Functional Requirements

### 3.1 Core Features
1. **Comments**
   - Users can post comments under articles or posts.
   - Users can edit and delete their own comments (if authenticated).
   - Comments are displayed in chronological and/or threaded order.
   - Support for replies and nested discussions.
   - Optional anonymous posting (if enabled by site owner).

2. **Reactions**
   - Users can react with predefined options (e.g., like, love, clap).
   - Multiple reaction types supported per site.
   - Aggregate reaction counts displayed per post or comment.

3. **Moderation**
   - Admin dashboard for site owners.
   - Manual approval/rejection of comments.
   - AI-powered moderation option (flagging spam, offensive language, etc.).
   - Configurable moderation workflows (pre-moderation, post-moderation, hybrid).
   - Moderation roles: owner, moderator, admin.

4. **Authentication**
   - Support for guest/anonymous users (optional).
   - Support for authenticated users (via external identity providers like Auth0).
   - Role-based permissions (admin, moderator, user).

5. **Integration**
   - Provide a JavaScript snippet or HTML embed to display comments/reactions.
   - Provide an API for direct integration with static site generators (e.g., during build step).
   - Support server-side rendering of comments (static HTML embedding).
   - Support API endpoints for dynamic interaction (adding/deleting comments, reactions).

6. **Admin/Developer Tools**
   - Admin UI to manage comments, reactions, and moderation settings.
   - Export/import of comments (JSON/CSV).
   - API for analytics (comment counts, active users, reaction breakdowns).
   - Configurable branding (CSS classes, styling hooks).

## 4. Non-Functional Requirements

### 4.1 Performance
- Page load impact must be minimal (<= 50ms added to TTFB for static rendering).
- API endpoints must respond within 200ms for 95% of requests.
- Support at least 1000 concurrent requests per site without degradation.

### 4.2 Scalability
- Handle sites ranging from small personal blogs (<100 monthly comments) to larger communities (>100k monthly comments).
- Support horizontal scaling without downtime.

### 4.3 Reliability
- 99.9% uptime target.
- Automatic recovery in case of failures.
- Comments must never be lost (guaranteed persistence).

### 4.4 Security
- All communication must be encrypted (TLS).
- User data stored securely (hashed passwords if any, encrypted tokens).
- Role-based access control for admins/moderators.
- Protection against spam (rate limiting, captcha options).

### 4.5 Privacy & Compliance
- GDPR compliance (right to delete data, data export).
- No unnecessary tracking of users.
- Clear consent for cookies/local storage usage.

### 4.6 Maintainability
- APIs must follow consistent naming and versioning conventions.
- Deprecation strategy for old endpoints.
- Logging and monitoring for debugging and analytics.

### 4.7 Usability
- Admin interface must be intuitive and usable on desktop and mobile.
- User-facing UI must be simple, minimal, and accessible (WCAG AA compliance).
- Support localization (multi-language UI for users and admins).

## 5. Quality Attributes
- **Extensibility:** Ability to add new interaction types (polls, Q&A) without major redesign.
- **Portability:** Should run on multiple hosting environments (cloud providers, containers, serverless).
- **Interoperability:** Works with multiple SSGs (Hugo, Jekyll, Next.js static export, Gengo).
- **Configurability:** Site owners can choose features (comments only, reactions only, or both).
- **Minimalism:** Default footprint on frontend should be small (<50KB JS bundle).
- **Testability:** APIs and moderation logic must be testable with automated tools.

## 6. Out of Scope
- Full-fledged forum or social network features.
- Complex WYSIWYG editor for comments (basic markdown only).
- Native mobile apps (web-first approach only).

