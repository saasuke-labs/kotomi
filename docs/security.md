# Security Architecture & Implementation

## Overview

This document provides detailed technical information about Kotomi's security implementation, architecture decisions, and security considerations for developers and security auditors.

## Authentication Architecture

### Auth0 Integration

Kotomi uses Auth0 for authentication, providing enterprise-grade security without maintaining password databases.

**Flow:**
1. User accesses protected route (e.g., `/admin/dashboard`)
2. Middleware checks for valid session
3. If no session, redirects to `/login`
4. Auth0 handles authentication (OAuth 2.0)
5. User redirected to `/callback` with authorization code
6. Server exchanges code for tokens with Auth0
7. Session created with encrypted cookie
8. User redirected to original destination

**Implementation:** `pkg/auth/auth0.go`

```go
// Session cookie configuration
store := sessions.NewCookieStore([]byte(sessionSecret))
store.Options = &sessions.Options{
    Path:     "/",
    MaxAge:   3600 * 24, // 24 hours
    HttpOnly: true,      // Prevents XSS attacks
    Secure:   false,     // Set to true in production with HTTPS
    SameSite: http.SameSiteLaxMode,
}
```

**Security Features:**
- Encrypted session cookies
- HTTP-only cookies (XSS protection)
- SameSite cookies (CSRF protection)
- Configurable expiration
- Secure flag for HTTPS (production)

### Authorization Middleware

**Implementation:** `pkg/auth/middleware.go`

```go
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        session, _ := store.Get(r, "auth-session")
        
        // Check if user is authenticated
        if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
            http.Redirect(w, r, "/login", http.StatusSeeOther)
            return
        }
        
        // Add user info to context
        ctx := context.WithValue(r.Context(), "userID", session.Values["userID"])
        next.ServeHTTP(w, r.WithContext(ctx))
    }
}
```

**Protection Scope:**
- All `/admin/*` routes require authentication
- Public API routes (`/api/*`) do not require authentication (by design)
- Health check (`/healthz`) is public

### Owner-Based Access Control

Every resource (site, page, comment) is owned by a user. The application enforces ownership:

**Example (Site Access):**
```go
func (h *SitesHandler) GetSiteByID(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value("userID").(string)
    siteID := mux.Vars(r)["siteId"]
    
    site, err := h.siteStore.GetByID(siteID)
    if err != nil {
        http.Error(w, "Site not found", http.StatusNotFound)
        return
    }
    
    // Verify ownership
    if site.OwnerID != userID {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }
    
    // ... return site data
}
```

## Database Security

### SQL Injection Prevention

All database queries use parameterized prepared statements via Go's `database/sql` package.

**Safe Example:**
```go
query := `
    INSERT INTO comments (id, site_id, page_id, author, text, created_at)
    VALUES (?, ?, ?, ?, ?, ?)
`
_, err := db.Exec(query, id, siteID, pageID, author, text, time.Now())
```

**What's Protected:**
- User inputs are never concatenated into SQL strings
- Database driver handles proper escaping
- No dynamic SQL construction

**Verified Files:**
- ✅ `pkg/comments/sqlite.go` - All queries parameterized
- ✅ `pkg/models/site.go` - All queries parameterized
- ✅ `pkg/models/page.go` - All queries parameterized
- ✅ `pkg/models/user.go` - All queries parameterized
- ✅ `pkg/models/reaction.go` - All queries parameterized

### Foreign Key Constraints

The database schema enforces referential integrity:

```sql
CREATE TABLE IF NOT EXISTS comments (
    id TEXT PRIMARY KEY,
    site_id TEXT NOT NULL,
    page_id TEXT NOT NULL,
    FOREIGN KEY (site_id) REFERENCES sites(id) ON DELETE CASCADE,
    FOREIGN KEY (page_id) REFERENCES pages(id) ON DELETE CASCADE
);
```

**Benefits:**
- Prevents orphaned records
- Maintains data consistency
- Cascade deletes ensure cleanup

### Database File Security

**Recommendations:**
1. Set restrictive file permissions:
   ```bash
   chmod 600 kotomi.db
   chown kotomi-user:kotomi-group kotomi.db
   ```

2. Store database outside web root
3. Enable SQLite encryption for sensitive data:
   ```bash
   # Using SQLCipher
   PRAGMA key = 'encryption-key';
   ```

## Web Security

### XSS (Cross-Site Scripting) Protection

Go's `html/template` package provides automatic context-aware escaping.

**Template Rendering:**
```go
tmpl := template.Must(template.ParseFiles("templates/admin/dashboard.html"))
tmpl.Execute(w, data) // Automatic HTML escaping
```

**Protected Contexts:**
- HTML content: `<div>{{.UserInput}}</div>`
- HTML attributes: `<div title="{{.Title}}">`
- JavaScript strings: `<script>var x = "{{.Data}}";</script>`
- CSS: `<style>.class { color: {{.Color}}; }</style>`

**Manual Review Confirmed:**
- ✅ All templates use automatic escaping
- ✅ No use of `template.HTML` or unescaped content
- ✅ JSON responses properly encoded

### CORS (Cross-Origin Resource Sharing)

**Implementation:** `pkg/middleware/cors.go`

```go
corsMiddleware := cors.New(cors.Options{
    AllowedOrigins:   parseOrigins(allowedOrigins),
    AllowedMethods:   parseMethods(allowedMethods),
    AllowedHeaders:   parseHeaders(allowedHeaders),
    AllowCredentials: allowCredentials,
})
```

**Configuration:**
- Default: Allows all origins (`*`) for development
- Production: Should restrict to specific domains
- OPTIONS preflight requests handled automatically
- Applied only to `/api/*` routes (not admin panel)

**Environment Variables:**
```bash
CORS_ALLOWED_ORIGINS=https://example.com,https://www.example.com
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=Content-Type,Authorization
CORS_ALLOW_CREDENTIALS=false
```

### Rate Limiting

**Implementation:** `pkg/middleware/ratelimit.go`

**Token Bucket Algorithm:**
```go
type Visitor struct {
    Limiter  *rate.Limiter
    LastSeen time.Time
}

func (v *Visitor) Allow() bool {
    return v.Limiter.Allow()
}
```

**Features:**
- IP-based rate limiting
- Separate limits for GET vs POST/PUT/DELETE
- Supports `X-Forwarded-For` and `X-Real-IP` headers
- Automatic cleanup of old visitors
- Rate limit headers in responses:
  - `X-RateLimit-Limit`
  - `X-RateLimit-Remaining`
  - `Retry-After` (when limit exceeded)

**Default Limits:**
- GET: 100 requests/minute per IP
- POST/PUT/DELETE: 5 requests/minute per IP

**Limitations:**
- In-memory storage (lost on restart)
- Can be bypassed with proxy rotation
- Consider Redis for distributed deployments

### Slowloris Attack Protection

**Implementation:** HTTP server timeouts

```go
server := &http.Server{
    Addr:              ":8080",
    Handler:           router,
    ReadHeaderTimeout: 10 * time.Second, // Slowloris protection
    ReadTimeout:       30 * time.Second, // Full read timeout
    WriteTimeout:      30 * time.Second, // Response timeout
    IdleTimeout:       60 * time.Second, // Keep-alive timeout
}
```

**Protection Against:**
- Slowloris attacks (slow header sending)
- Slow read attacks
- Connection exhaustion
- Keep-alive abuse

## Security Limitations & Mitigations

### 1. Anonymous Comments

**Current State:**
- Public API accepts any author name
- No verification of commenter identity
- Users cannot edit/delete their own comments

**Risks:**
- Impersonation
- Spam
- Abuse

**Mitigations:**
- Rate limiting prevents automated spam
- Manual moderation via admin panel
- IP tracking for abuse prevention

**Future Enhancement:**
- Implement user authentication for commenters
- Optional anonymous posting per site
- User profiles and comment ownership

### 2. No CSRF Protection on Admin Panel

**Current State:**
- Admin panel uses HTMX but lacks explicit CSRF tokens
- Auth0 session provides some protection

**Risks:**
- Cross-site request forgery attacks

**Mitigations:**
- SameSite cookies (Lax mode)
- Auth0 session validation
- HTTPS only in production

**Future Enhancement:**
- Add CSRF token middleware
- Validate tokens on state-changing operations

### 3. SQLite Concurrency

**Current State:**
- SQLite supports limited concurrent writes
- Not recommended for high-traffic production

**Risks:**
- Database lock errors under high load
- Performance degradation

**Mitigations:**
- Appropriate for small to medium sites
- Write operations are quick (no long transactions)
- WAL mode for better concurrency

**Future Enhancement:**
- PostgreSQL support for production scale
- Connection pooling
- Read replicas

### 4. Session Storage

**Current State:**
- Cookie-based sessions
- In-memory session store
- Sessions lost on restart

**Risks:**
- Users logged out on deployment
- No session persistence

**Mitigations:**
- 24-hour session expiration (users re-login occasionally)
- Only affects admin users (not end-users)

**Future Enhancement:**
- Redis session store
- Persistent session storage
- Session migration on deployment

## Security Testing

### Automated Scanning

**Tool:** gosec v2.22.11

**Results:** 20 findings (0 critical, 0 high, 4 medium, 16 low)

**Critical Findings:** None

**Medium Severity:**
1. ✅ **Fixed** - G112: Missing HTTP server timeouts (Slowloris protection)
2. ✅ **Accepted** - G107: Variable URLs in test code (acceptable in tests, 3 instances)

**Low Severity:**
- G104: Unhandled errors (16 instances, mostly in response encoding)
- Risk accepted: Errors in JSON encoding are difficult to handle meaningfully

### Manual Testing

**Tested Scenarios:**

✅ **SQL Injection:**
- Attempted injection in all input fields
- Special characters: `'; DROP TABLE comments; --`
- Result: Properly escaped, no injection possible

✅ **XSS Attacks:**
- Injected scripts in comment text: `<script>alert('XSS')</script>`
- Injected HTML: `<img src=x onerror=alert('XSS')>`
- Result: Properly escaped, rendered as text

✅ **Authentication Bypass:**
- Attempted direct access to admin routes without session
- Attempted session manipulation
- Result: Properly redirected to login

✅ **Authorization Bypass:**
- Attempted access to other users' sites/pages/comments
- Result: 403 Forbidden responses

✅ **Rate Limiting:**
- Sent rapid requests beyond limits
- Result: 429 Too Many Requests after threshold

### OWASP Top 10 Coverage

| Vulnerability | Status | Protection |
|---------------|--------|------------|
| A01:2021 Broken Access Control | ✅ Protected | Authentication + owner verification |
| A02:2021 Cryptographic Failures | ✅ Protected | Encrypted sessions, HTTPS in production |
| A03:2021 Injection | ✅ Protected | Parameterized queries, input validation |
| A04:2021 Insecure Design | ✅ Protected | Security-first architecture |
| A05:2021 Security Misconfiguration | ⚠️ Caution | Proper production config required |
| A06:2021 Vulnerable Components | ✅ Protected | Regular dependency updates |
| A07:2021 Authentication Failures | ✅ Protected | Auth0 OAuth 2.0 |
| A08:2021 Software/Data Integrity | ✅ Protected | Database constraints, no file uploads |
| A09:2021 Logging Failures | ⚠️ Limited | Basic logging, needs enhancement |
| A10:2021 Server-Side Request Forgery | ✅ Protected | No outbound requests from user input |

## Incident Response

### If a Vulnerability is Found

1. **Report:** Email security@saasuke-labs.com
2. **Triage:** Security team reviews within 48 hours
3. **Fix:** Develop and test patch
4. **Notify:** Inform affected users if necessary
5. **Release:** Deploy security update
6. **Disclose:** Public disclosure after fix is available

### Security Update Process

1. Create security advisory
2. Develop fix in private branch
3. Test thoroughly
4. Release patch version
5. Update SECURITY.md
6. Notify users via GitHub Security Advisories

## Compliance Considerations

### GDPR Compliance

**User Data Stored:**
- Auth0 user ID (sub)
- Email address
- Name
- IP addresses (for rate limiting, temporary)
- Comment data (author names, text)

**User Rights:**
- Right to access: Users can view their data via admin panel
- Right to deletion: Users can delete their sites/comments
- Right to rectification: Users can update their content

**Recommendations:**
- Implement data export functionality
- Add privacy policy
- Add cookie consent banner
- Implement data retention policies

### Data Retention

**Current Policy:** Indefinite retention

**Recommendations:**
1. Delete inactive sites after X months of inactivity
2. Anonymize old comments after X years
3. Purge rate limiting data after 24 hours (already implemented)
4. Regular database cleanup jobs

## Security Maintenance

### Regular Tasks

**Weekly:**
- [ ] Review application logs for anomalies
- [ ] Check rate limit effectiveness

**Monthly:**
- [ ] Update dependencies: `go get -u ./...`
- [ ] Run security scan: `gosec ./...`
- [ ] Review new Auth0 security advisories

**Quarterly:**
- [ ] Full security audit
- [ ] Penetration testing
- [ ] Review and rotate secrets
- [ ] Update security documentation

**Annually:**
- [ ] Third-party security audit
- [ ] Compliance review
- [ ] Disaster recovery drill

### Dependency Management

**Check for vulnerabilities:**
```bash
# Update dependencies
go get -u ./...
go mod tidy

# Audit dependencies (requires nancy)
go list -m all | nancy sleuth
```

**Current Dependencies:** (See `go.mod`)
- gorilla/mux - HTTP router
- gorilla/sessions - Session management
- rs/cors - CORS middleware
- mattn/go-sqlite3 - SQLite driver
- google/uuid - UUID generation
- coreos/go-oidc/v3 - OpenID Connect

## Production Security Checklist

### Pre-Deployment

- [ ] All environment variables configured
- [ ] SESSION_SECRET set to strong random value (min 32 chars)
- [ ] CORS_ALLOWED_ORIGINS restricted to production domains
- [ ] Rate limits configured appropriately
- [ ] HTTPS certificate obtained and configured
- [ ] Database file permissions set (chmod 600)
- [ ] Automated backups configured
- [ ] Monitoring and logging configured
- [ ] Security headers configured in reverse proxy
- [ ] Firewall rules configured

### Post-Deployment

- [ ] Verify HTTPS is working
- [ ] Test authentication flow
- [ ] Verify CORS restrictions
- [ ] Test rate limiting
- [ ] Monitor logs for errors
- [ ] Perform smoke tests
- [ ] Document incident response contacts

### Ongoing

- [ ] Monitor logs regularly
- [ ] Review security advisories
- [ ] Update dependencies monthly
- [ ] Backup verification
- [ ] Performance monitoring
- [ ] Security scan automation

## Recommended Security Headers

Configure these in your reverse proxy (nginx, Caddy, etc.):

```nginx
# Nginx example
add_header X-Frame-Options "SAMEORIGIN" always;
add_header X-Content-Type-Options "nosniff" always;
add_header X-XSS-Protection "1; mode=block" always;
add_header Referrer-Policy "strict-origin-when-cross-origin" always;
add_header Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline' https://unpkg.com; style-src 'self' 'unsafe-inline';" always;
add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
```

## Contact & Resources

**Security Contact:** security@saasuke-labs.com

**Resources:**
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Auth0 Security](https://auth0.com/security)
- [Go Security](https://go.dev/security/)
- [gosec](https://github.com/securego/gosec)

---

**Last Updated:** January 31, 2026  
**Version:** 1.0  
**Next Review:** April 30, 2026
