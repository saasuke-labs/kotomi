# Security Policy

## Reporting a Vulnerability

We take the security of Kotomi seriously. If you believe you have found a security vulnerability, please report it to us as described below.

### How to Report a Security Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, please send an email to security@saasuke-labs.com (or the appropriate security contact for your organization).

You should receive a response within 48 hours. If for some reason you do not, please follow up via email to ensure we received your original message.

Please include the following information in your report:

* Type of vulnerability
* Full paths of source file(s) related to the vulnerability
* The location of the affected source code (tag/branch/commit or direct URL)
* Any special configuration required to reproduce the issue
* Step-by-step instructions to reproduce the issue
* Proof-of-concept or exploit code (if possible)
* Impact of the issue, including how an attacker might exploit it

This information will help us triage your report more quickly.

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.0.x   | :white_check_mark: |

## Security Measures

### Authentication & Authorization

* **Auth0 Integration**: All admin panel access is secured through Auth0 OAuth 2.0
* **Session Management**: Encrypted session cookies with configurable secret keys
* **Role-Based Access Control**: Middleware enforces authentication for all admin routes
* **Owner Verification**: Users can only access and modify their own sites, pages, and content

### Data Protection

* **SQL Injection Prevention**: All database queries use parameterized prepared statements
* **Input Validation**: User inputs are validated and sanitized
* **XSS Protection**: HTML templates use Go's built-in automatic escaping
* **Database Encryption**: Session secrets are encrypted; consider enabling SQLite encryption for production

### Network Security

* **CORS Configuration**: Configurable CORS middleware protects API endpoints
  - Default allows all origins in development (`*`)
  - Should be restricted to specific domains in production
* **Rate Limiting**: IP-based rate limiting protects against abuse
  - GET requests: 100 requests/minute per IP (configurable)
  - POST/PUT/DELETE requests: 5 requests/minute per IP (configurable)
* **Slowloris Protection**: HTTP server configured with appropriate timeouts
  - ReadHeaderTimeout: 10 seconds
  - ReadTimeout: 30 seconds
  - WriteTimeout: 30 seconds
  - IdleTimeout: 60 seconds

### Application Security

* **Foreign Key Constraints**: Database enforces referential integrity
* **Cascade Deletes**: Properly configured to maintain data consistency
* **Error Handling**: Error messages don't expose sensitive information
* **Dependency Management**: Regular updates and vulnerability scanning

## Security Best Practices

### For Deployment

1. **Environment Variables**: Never commit secrets to version control
   ```bash
   # Required secrets
   AUTH0_DOMAIN=your-domain.auth0.com
   AUTH0_CLIENT_ID=your-client-id
   AUTH0_CLIENT_SECRET=your-client-secret
   SESSION_SECRET=your-random-secret-key
   ```

2. **CORS Configuration**: Restrict allowed origins in production
   ```bash
   CORS_ALLOWED_ORIGINS=https://yourdomain.com,https://www.yourdomain.com
   CORS_ALLOW_CREDENTIALS=false
   ```

3. **Database Security**:
   - Use SQLite encryption for production databases
   - Regular backups with secure storage
   - Restrict file system permissions on database file
   ```bash
   chmod 600 kotomi.db
   ```

4. **HTTPS Only**: Always use HTTPS in production
   - Configure reverse proxy (nginx, Caddy) with TLS
   - Set secure cookie flags
   - Use HSTS headers

5. **Rate Limiting**: Configure appropriate limits based on expected traffic
   ```bash
   RATE_LIMIT_GET=100    # Requests per minute
   RATE_LIMIT_POST=5     # Requests per minute
   ```

### For Development

1. **Auto-Generated Secrets**: Development mode auto-generates session secrets
   - These are not suitable for production
   - Always set `SESSION_SECRET` in production

2. **Test Credentials**: Never use production credentials in tests
3. **Code Review**: All changes should be reviewed for security implications
4. **Dependency Audits**: Run `go mod tidy` and check for vulnerabilities regularly

## Known Security Considerations

### Current Limitations

1. **Anonymous Comments**: Public API accepts any author name without verification
   - Users cannot edit/delete their own comments (no user auth for commenters)
   - Consider implementing user authentication for comment authors in future versions

2. **SQLite Limitations**: 
   - Not recommended for high-concurrency production deployments
   - Consider PostgreSQL for production at scale
   - Single file database - ensure proper backup strategy

3. **Session Storage**:
   - Cookie-based sessions stored in memory
   - Consider Redis for distributed deployments
   - Sessions lost on server restart

4. **No CSRF Protection**: 
   - Admin panel uses HTMX but lacks CSRF tokens
   - Auth0 provides some protection
   - Consider adding explicit CSRF middleware

5. **IP-Based Rate Limiting**:
   - Can be bypassed with VPN/proxy rotation
   - Consider adding additional rate limiting strategies
   - Supports X-Forwarded-For and X-Real-IP headers

## Security Audit Results

**Last Audit**: January 31, 2026  
**Tool**: gosec v2.22.11  
**Status**: ✅ PASSED with minor findings

### Automated Scan Results

Total Issues: 20
- **Critical**: 0
- **High**: 0
- **Medium**: 4 (2 resolved, 2 accepted)
- **Low**: 16 (mostly unhandled errors in non-critical paths)

### Resolved Issues

1. ✅ **G112 - Slowloris Attack Protection**: Added HTTP server timeouts
   - ReadHeaderTimeout: 10 seconds
   - ReadTimeout: 30 seconds
   - WriteTimeout: 30 seconds
   - IdleTimeout: 60 seconds

### Accepted Risks

1. **G107 - Variable URLs in Tests**: Low risk, only in test code
   - Tests use dynamic URLs for E2E testing
   - Not exposed in production code
   - Acceptable for testing purposes

2. **G104 - Unhandled Errors**: Low risk, mostly in response encoding
   - Errors in `json.NewEncoder().Encode()` are difficult to handle meaningfully
   - Errors in `resp.Body.Close()` are logged elsewhere
   - Database close errors in defer/error paths are already handled

### Manual Security Review

✅ **SQL Injection**: All queries use parameterized statements  
✅ **XSS Protection**: Templates use automatic HTML escaping  
✅ **Authentication**: Auth0 OAuth 2.0 properly implemented  
✅ **Authorization**: Owner-based access control enforced  
✅ **Session Security**: Encrypted cookies with secure configuration  
✅ **Password Storage**: Delegated to Auth0 (no passwords stored locally)  
✅ **File Upload**: Not implemented (no file upload vulnerabilities)  
✅ **Directory Traversal**: Not applicable (no file serving by path)

### Recommendations for Production

1. **Enable HTTPS**: Required for production deployment
2. **Restrict CORS**: Set specific allowed origins
3. **Monitor Rate Limits**: Track and adjust based on traffic patterns
4. **Regular Updates**: Keep dependencies updated
5. **Log Monitoring**: Implement centralized logging and alerting
6. **Database Backups**: Automated backup strategy
7. **Secret Rotation**: Regular rotation of session secrets
8. **Security Headers**: Add security headers via reverse proxy:
   - Content-Security-Policy
   - X-Frame-Options
   - X-Content-Type-Options
   - Strict-Transport-Security (HSTS)

## Secure Configuration Example

```bash
# Production Environment Variables

# Auth0 Configuration
AUTH0_DOMAIN=production.auth0.com
AUTH0_CLIENT_ID=prod_client_id
AUTH0_CLIENT_SECRET=prod_secret_key
AUTH0_CALLBACK_URL=https://yourdomain.com/callback

# Session Security
SESSION_SECRET=your-strong-random-secret-min-32-chars

# CORS Configuration
CORS_ALLOWED_ORIGINS=https://yourdomain.com
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=Content-Type,Authorization
CORS_ALLOW_CREDENTIALS=false

# Rate Limiting
RATE_LIMIT_GET=100
RATE_LIMIT_POST=5

# Database
DB_PATH=/var/lib/kotomi/kotomi.db

# Server
PORT=8080
```

## Security Checklist for Production

- [ ] HTTPS enabled with valid TLS certificate
- [ ] All environment variables configured with production values
- [ ] Session secret set to strong random value (min 32 characters)
- [ ] CORS origins restricted to actual domain(s)
- [ ] Rate limits configured appropriately
- [ ] Database file permissions restricted (chmod 600)
- [ ] Database backups automated
- [ ] Logging configured and monitored
- [ ] Security headers configured in reverse proxy
- [ ] Dependency vulnerabilities checked (go list -m all | nancy)
- [ ] Regular security updates scheduled

## Contact

For security concerns, please contact: security@saasuke-labs.com

---

**Last Updated**: January 31, 2026  
**Document Version**: 1.0
