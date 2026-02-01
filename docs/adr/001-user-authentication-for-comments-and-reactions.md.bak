# ADR 001: User Authentication for Comments and Reactions

**Status:** Proposed  
**Date:** 2026-01-31  
**Authors:** Kotomi Development Team  
**Deciders:** Product Team, Engineering Team  

## Context and Problem Statement

Kotomi currently supports comments and reactions on static websites, but user identification is limited:

- **Comments**: Users provide an author name (string field), which can be arbitrary and doesn't establish true identity
- **Reactions**: User identification relies on IP addresses (`user_identifier` field), which is:
  - Unreliable (shared IPs, dynamic IPs, VPNs, proxies)
  - Privacy-concerning in some jurisdictions
  - Cannot distinguish between multiple users behind the same IP
  - Cannot link reactions to verified user identities

**Current Admin Authentication:**
- The admin panel uses Auth0 for authentication
- This provides secure authentication for site owners and moderators
- However, end-users (commenters and reactors) have no authentication

**The Challenge:**
Many Kotomi deployments are embedded into existing websites that already have their own authentication systems (Auth0, Firebase, custom OAuth, SAML, etc.). Sites need a way to "bring their own authentication" while Kotomi handles the comment and reaction functionality.

**Key Requirements:**
1. Support authenticated users for comments and reactions
2. Allow sites to integrate their existing authentication systems
3. Maintain backward compatibility with anonymous/guest commenting (if enabled)
4. Support user profile management (edit/delete own comments, view own reactions)
5. Enable moderation based on user reputation/history
6. Respect privacy regulations (GDPR, CCPA)
7. Minimize complexity for sites without authentication needs
8. Provide flexibility for various authentication providers

## Decision Drivers

* **Flexibility**: Sites must be able to use their existing authentication infrastructure
* **Security**: Authentication must be secure and tamper-proof
* **Privacy**: User data must be handled according to privacy regulations
* **User Experience**: Authentication should be seamless and not disruptive
* **Developer Experience**: Integration should be straightforward for site developers
* **Backward Compatibility**: Existing anonymous commenting must continue to work
* **Performance**: Authentication checks should not significantly impact API performance
* **Scalability**: Solution must scale to thousands of sites with different auth providers

## Considered Options

### Option 1: Centralized Authentication (Kotomi-Managed Auth)

Kotomi provides its own authentication system (similar to Disqus):

**Pros:**
- Single, consistent authentication experience across all sites
- Full control over security implementation
- Simplified user management
- Easier to implement user reputation/karma systems
- Users can have a single identity across multiple sites using Kotomi

**Cons:**
- Sites cannot use their existing authentication
- Users must create yet another account
- Privacy concerns (centralized user database)
- Doesn't meet the "bring your own key" requirement
- More maintenance burden for Kotomi team
- Potential regulatory compliance complexity

### Option 2: OAuth/OIDC Proxy (Kotomi as Auth Proxy)

Kotomi acts as an OAuth/OIDC relying party for multiple providers:

**Pros:**
- Supports major authentication providers (Google, GitHub, Auth0, etc.)
- Standards-based approach
- Users can authenticate with familiar services

**Cons:**
- Still requires configuration per provider in Kotomi
- Doesn't support custom/proprietary authentication systems
- Requires Kotomi to store OAuth client secrets
- Complex configuration for site owners
- Doesn't truly allow sites to "bring their own" authentication

### Option 3: JWT-Based Delegated Authentication (Recommended)

Sites handle authentication themselves and issue signed JWT tokens that Kotomi validates:

**Architecture:**
```
┌─────────────────┐         ┌──────────────────┐         ┌───────────────┐
│   Site's Web    │         │   Site's Auth    │         │    Kotomi     │
│   Application   │         │     System       │         │   Service     │
└────────┬────────┘         └────────┬─────────┘         └───────┬───────┘
         │                           │                           │
         │ 1. User logs in           │                           │
         │─────────────────────────>│                           │
         │                           │                           │
         │ 2. Auth successful        │                           │
         │   Returns JWT token       │                           │
         │<─────────────────────────│                           │
         │                           │                           │
         │ 3. Include JWT in API     │                           │
         │   requests to Kotomi      │                           │
         │───────────────────────────────────────────────────>│
         │                           │                           │
         │                           │   4. Validate JWT         │
         │                           │      using site's         │
         │                           │      public key/secret    │
         │                           │                           │
         │ 5. Process request with   │                           │
         │    authenticated user ID  │                           │
         │<───────────────────────────────────────────────────│
```

**Pros:**
- ✅ Sites maintain full control of authentication
- ✅ Works with any authentication system (OAuth, SAML, custom, etc.)
- ✅ Kotomi never handles passwords or sensitive auth data
- ✅ Standard JWT format (widely supported)
- ✅ Stateless validation (high performance)
- ✅ Can include custom claims (roles, permissions, display name)
- ✅ Sites can revoke access by rotating keys
- ✅ Supports both symmetric (HMAC) and asymmetric (RSA, ECDSA) signatures

**Cons:**
- Sites must generate JWT tokens (requires some technical knowledge)
- Requires key management by site owners
- Clock synchronization important for token expiration

### Option 4: Webhook-Based Validation

Kotomi calls back to the site's API to validate user tokens:

**Pros:**
- Sites have complete control over validation logic
- No key management needed in Kotomi
- Works with any authentication system

**Cons:**
- Performance overhead (HTTP request per validation)
- Reliability concerns (site's API must be available)
- Increased latency for API requests
- Privacy concerns (Kotomi sends tokens to site)
- More complex failure scenarios

## Decision Outcome

**Chosen option:** Option 3 - JWT-Based Delegated Authentication

This approach best satisfies the "bring your own key" requirement while maintaining security, performance, and flexibility.

### Implementation Details

#### 1. Site Configuration

Each site in Kotomi will have authentication configuration:

```go
type SiteAuthConfig struct {
    ID                    string    `json:"id"`
    SiteID                string    `json:"site_id"`
    AuthEnabled           bool      `json:"auth_enabled"`           // Enable/disable authentication
    AllowAnonymous        bool      `json:"allow_anonymous"`        // Allow unauthenticated comments/reactions
    RequireAuth           bool      `json:"require_auth"`           // Require authentication (cannot be anonymous)
    JWTValidationType     string    `json:"jwt_validation_type"`    // "hmac", "rsa", "ecdsa", "jwks"
    JWTSecret             string    `json:"jwt_secret,omitempty"`   // For HMAC (symmetric) - stored encrypted
    JWTPublicKey          string    `json:"jwt_public_key,omitempty"` // For RSA/ECDSA (asymmetric)
    JWKSEndpoint          string    `json:"jwks_endpoint,omitempty"`  // For JWKS (JSON Web Key Set) URL
    JWTIssuer             string    `json:"jwt_issuer"`             // Expected issuer claim
    JWTAudience           string    `json:"jwt_audience"`           // Expected audience claim
    TokenExpirationBuffer int       `json:"token_expiration_buffer"` // Grace period in seconds
    CreatedAt             time.Time `json:"created_at"`
    UpdatedAt             time.Time `json:"updated_at"`
}
```

#### 2. JWT Token Format

Sites will issue JWT tokens with the following claims:

**Required Claims:**
```json
{
  "iss": "https://example.com",           // Issuer (site's domain)
  "sub": "user-uuid-12345",               // Subject (unique user ID)
  "aud": "kotomi",                        // Audience (Kotomi service)
  "exp": 1738368515,                      // Expiration timestamp
  "iat": 1738364915,                      // Issued at timestamp
  "kotomi_user": {
    "id": "user-uuid-12345",              // User ID (consistent identifier)
    "name": "Jane Doe",                   // Display name for comments
    "email": "jane@example.com"           // Email (optional, for moderation)
  }
}
```

**Optional Claims:**
```json
{
  "kotomi_user": {
    "avatar_url": "https://example.com/avatars/jane.jpg",
    "profile_url": "https://example.com/users/jane",
    "roles": ["member", "premium"],
    "verified": true,
    "created_at": "2024-01-15T10:00:00Z"
  }
}
```

#### 3. API Changes

**a) Request Headers:**

Clients will send JWT tokens via the `Authorization` header:

```
Authorization: Bearer <JWT_TOKEN>
```

**b) User Model:**

New user model for authenticated users:

```go
type User struct {
    ID          string    `json:"id"`           // Unique user ID from JWT sub claim
    SiteID      string    `json:"site_id"`      // Site this user belongs to
    Name        string    `json:"name"`         // Display name
    Email       string    `json:"email,omitempty"` // Email (optional)
    AvatarURL   string    `json:"avatar_url,omitempty"` // Avatar URL
    ProfileURL  string    `json:"profile_url,omitempty"` // Profile page URL
    IsVerified  bool      `json:"is_verified"`  // Verified user status
    Roles       []string  `json:"roles,omitempty"` // User roles
    FirstSeen   time.Time `json:"first_seen"`   // First time seen in Kotomi
    LastSeen    time.Time `json:"last_seen"`    // Last activity
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

**c) Comment Model Updates:**

```go
type Comment struct {
    ID          string    `json:"id"`
    SiteID      string    `json:"site_id"`
    PageID      string    `json:"page_id"`
    Author      string    `json:"author"`         // Display name (kept for backward compatibility)
    AuthorID    string    `json:"author_id,omitempty"` // NEW: User ID if authenticated
    Email       string    `json:"email,omitempty"`     // NEW: User email (for moderation)
    Text        string    `json:"text"`
    ParentID    string    `json:"parent_id,omitempty"`
    Status      string    `json:"status"`
    IsAnonymous bool      `json:"is_anonymous"`   // NEW: True if posted anonymously
    ModeratedBy string    `json:"moderated_by,omitempty"`
    ModeratedAt time.Time `json:"moderated_at,omitempty"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

**d) Reaction Model Updates:**

```go
type Reaction struct {
    ID                string    `json:"id"`
    PageID            string    `json:"page_id,omitempty"`
    CommentID         string    `json:"comment_id,omitempty"`
    AllowedReactionID string    `json:"allowed_reaction_id"`
    UserIdentifier    string    `json:"user_identifier"` // CHANGED: Now stores user_id instead of IP
    UserID            string    `json:"user_id,omitempty"` // NEW: Explicit user ID field
    IsAnonymous       bool      `json:"is_anonymous"`     // NEW: True if reacted anonymously
    CreatedAt         time.Time `json:"created_at"`
}
```

#### 4. Authentication Middleware

New middleware to validate JWT tokens:

```go
func JWTAuthMiddleware(siteAuthConfig *SiteAuthConfig) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. Extract token from Authorization header
        authHeader := c.GetHeader("Authorization")
        
        // 2. If no token and anonymous allowed, continue
        if authHeader == "" && siteAuthConfig.AllowAnonymous {
            c.Next()
            return
        }
        
        // 3. If no token and auth required, reject
        if authHeader == "" && siteAuthConfig.RequireAuth {
            c.JSON(401, gin.H{"error": "Authentication required"})
            c.Abort()
            return
        }
        
        // 4. Parse and validate JWT
        token, err := validateJWT(authHeader, siteAuthConfig)
        if err != nil {
            c.JSON(401, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }
        
        // 5. Extract user info and store in context
        user := extractUserFromToken(token)
        c.Set("authenticated_user", user)
        c.Set("is_authenticated", true)
        
        c.Next()
    }
}
```

#### 5. API Endpoints

**New Endpoints:**

- `POST /api/v1/admin/sites/{siteId}/auth/config` - Configure authentication for a site
- `GET /api/v1/admin/sites/{siteId}/auth/config` - Get authentication configuration
- `PUT /api/v1/admin/sites/{siteId}/auth/config` - Update authentication configuration
- `DELETE /api/v1/admin/sites/{siteId}/auth/config` - Disable authentication
- `GET /api/v1/admin/sites/{siteId}/users` - List users who have commented/reacted
- `GET /api/v1/admin/sites/{siteId}/users/{userId}` - Get user details
- `DELETE /api/v1/admin/sites/{siteId}/users/{userId}` - Ban/remove a user

**Modified Endpoints:**

Existing comment and reaction endpoints will:
1. Accept optional JWT token in Authorization header
2. Extract user info if token is valid
3. Fall back to anonymous mode if no token and allowed
4. Store user_id instead of IP for authenticated users

#### 6. Database Schema Changes

**New Table: site_auth_configs**

```sql
CREATE TABLE site_auth_configs (
    id TEXT PRIMARY KEY,
    site_id TEXT NOT NULL UNIQUE,
    auth_enabled BOOLEAN DEFAULT FALSE,
    allow_anonymous BOOLEAN DEFAULT TRUE,
    require_auth BOOLEAN DEFAULT FALSE,
    jwt_validation_type TEXT NOT NULL, -- 'hmac', 'rsa', 'ecdsa', 'jwks'
    jwt_secret TEXT,                   -- Encrypted at rest
    jwt_public_key TEXT,
    jwks_endpoint TEXT,
    jwt_issuer TEXT NOT NULL,
    jwt_audience TEXT NOT NULL,
    token_expiration_buffer INTEGER DEFAULT 60,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (site_id) REFERENCES sites(id) ON DELETE CASCADE
);
```

**New Table: users**

```sql
CREATE TABLE users (
    id TEXT PRIMARY KEY,               -- user_id from JWT
    site_id TEXT NOT NULL,
    name TEXT NOT NULL,
    email TEXT,
    avatar_url TEXT,
    profile_url TEXT,
    is_verified BOOLEAN DEFAULT FALSE,
    roles TEXT,                        -- JSON array of roles
    first_seen DATETIME NOT NULL,
    last_seen DATETIME NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (site_id) REFERENCES sites(id) ON DELETE CASCADE,
    UNIQUE(site_id, id)
);

CREATE INDEX idx_users_site_id ON users(site_id);
CREATE INDEX idx_users_email ON users(site_id, email);
```

**Modified Table: comments**

```sql
-- Add new columns (backward compatible):
ALTER TABLE comments ADD COLUMN author_id TEXT;
ALTER TABLE comments ADD COLUMN email TEXT;
ALTER TABLE comments ADD COLUMN is_anonymous BOOLEAN DEFAULT FALSE;

-- Add index for author_id
CREATE INDEX idx_comments_author_id ON comments(site_id, author_id);

-- Add foreign key (optional, depends on implementation)
-- FOREIGN KEY (site_id, author_id) REFERENCES users(site_id, id) ON DELETE SET NULL
```

**Modified Table: reactions**

```sql
-- Add new columns (backward compatible):
ALTER TABLE reactions ADD COLUMN user_id TEXT;
ALTER TABLE reactions ADD COLUMN is_anonymous BOOLEAN DEFAULT FALSE;

-- Add index for user_id
CREATE INDEX idx_reactions_user_id ON reactions(user_id);

-- Migration note: user_identifier will continue to exist for backward compatibility
-- For authenticated users: user_identifier = user_id
-- For anonymous users: user_identifier = IP address (if allowed)
```

#### 7. Security Considerations

**a) Key Storage:**
- HMAC secrets will be encrypted at rest using application-level encryption
- Public keys for RSA/ECDSA will be stored in plain text (they're public)
- Admin panel will provide secure key generation tools

**b) Token Validation:**
- Verify signature using configured validation method
- Check expiration (`exp` claim) with configurable buffer
- Verify issuer (`iss` claim) matches site configuration
- Verify audience (`aud` claim) is "kotomi" or site-specific
- Check not-before (`nbf` claim) if present
- Implement token caching with short TTL to reduce validation overhead

**c) Rate Limiting:**
- Apply stricter rate limits to anonymous users
- More generous limits for authenticated users
- Track rate limits per user_id for authenticated users
- Track rate limits per IP for anonymous users

**d) User Privacy:**
- Email addresses are optional and used only for moderation
- User IDs are site-specific (no cross-site tracking)
- Sites control what user data is included in JWT
- GDPR compliance: users can request deletion via site owner

**e) Token Expiration:**
- Sites should issue short-lived tokens (5-60 minutes recommended)
- Tokens should be refreshed by the site's auth system
- Kotomi will reject expired tokens (with configurable grace period)

#### 8. Migration Path

**Phase 1: Backward Compatibility (Required)**
- All existing API endpoints continue to work without authentication
- IP-based user identification continues for anonymous users
- Comments with just "author" string continue to work

**Phase 2: Opt-in Authentication (Per Site)**
- Sites can enable authentication via admin panel
- Sites configure JWT validation method
- Sites can choose to allow/disallow anonymous comments

**Phase 3: Gradual Migration**
- Existing comments remain unchanged (no author_id)
- New comments from authenticated users include author_id
- Admin panel shows mixed authenticated/anonymous comments

**Phase 4: Enhanced Features (Authenticated Users Only)**
- Edit own comments (requires author_id)
- Delete own comments (requires author_id)
- View comment history (requires author_id)
- User reputation/karma (requires author_id)
- Block/mute users (requires author_id)

#### 9. Developer Experience

**Site Integration Steps:**

**Step 1: Enable Authentication in Kotomi Admin Panel**
```
1. Log in to Kotomi admin panel
2. Navigate to Site Settings → Authentication
3. Enable "User Authentication"
4. Choose validation method: HMAC or RSA
5. For HMAC: Generate a shared secret (or provide your own)
6. For RSA: Upload your public key
7. Set issuer (your domain) and audience (kotomi)
8. Choose whether to allow anonymous comments
9. Save configuration
```

**Step 2: Generate JWT Tokens in Your Application**

Example using Node.js:
```javascript
const jwt = require('jsonwebtoken');

function generateKotomiToken(user) {
  const payload = {
    iss: 'https://example.com',
    sub: user.id,
    aud: 'kotomi',
    exp: Math.floor(Date.now() / 1000) + (60 * 60), // 1 hour
    iat: Math.floor(Date.now() / 1000),
    kotomi_user: {
      id: user.id,
      name: user.display_name,
      email: user.email,
      avatar_url: user.avatar_url,
      verified: user.is_verified
    }
  };
  
  return jwt.sign(payload, process.env.KOTOMI_JWT_SECRET);
}
```

Example using Python:
```python
import jwt
import time

def generate_kotomi_token(user):
    payload = {
        'iss': 'https://example.com',
        'sub': user.id,
        'aud': 'kotomi',
        'exp': int(time.time()) + 3600,  # 1 hour
        'iat': int(time.time()),
        'kotomi_user': {
            'id': user.id,
            'name': user.display_name,
            'email': user.email,
            'avatar_url': user.avatar_url,
            'verified': user.is_verified
        }
    }
    
    return jwt.encode(payload, os.getenv('KOTOMI_JWT_SECRET'), algorithm='HS256')
```

**Step 3: Include Token in API Requests**

JavaScript example:
```javascript
// After user logs in, get the token
const token = generateKotomiToken(currentUser);

// Include in API requests to Kotomi
fetch('https://kotomi.example.com/api/v1/site/{siteId}/page/{pageId}/comments', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`
  },
  body: JSON.stringify({
    text: 'This is my comment'
  })
});
```

#### 10. Configuration Examples

**Example 1: Blog with Auth0**
```yaml
auth_enabled: true
allow_anonymous: false  # Require authentication
require_auth: true
jwt_validation_type: "jwks"
jwks_endpoint: "https://example.auth0.com/.well-known/jwks.json"
jwt_issuer: "https://example.auth0.com/"
jwt_audience: "kotomi"
```

**Example 2: Open Source Project (Allow Both)**
```yaml
auth_enabled: true
allow_anonymous: true   # Allow both authenticated and anonymous
require_auth: false
jwt_validation_type: "rsa"
jwt_public_key: "-----BEGIN PUBLIC KEY-----\n..."
jwt_issuer: "https://github.com"
jwt_audience: "kotomi"
```

**Example 3: Private Community (Symmetric Key)**
```yaml
auth_enabled: true
allow_anonymous: false
require_auth: true
jwt_validation_type: "hmac"
jwt_secret: "your-secret-key-min-32-chars"  # Stored encrypted
jwt_issuer: "https://community.example.com"
jwt_audience: "kotomi"
```

**Example 4: Public Site (No Auth)**
```yaml
auth_enabled: false
allow_anonymous: true
# No JWT configuration needed
```

## Consequences

### Positive

1. **Flexibility**: Sites can use any authentication system
2. **Security**: Kotomi doesn't handle sensitive authentication data
3. **Privacy**: No centralized user database (site-specific users)
4. **Performance**: Stateless JWT validation is fast
5. **Scalability**: No additional infrastructure for authentication
6. **Developer-Friendly**: Standard JWT format, well-documented
7. **User Experience**: Seamless authentication (no extra login)
8. **Features**: Enables user-specific features (edit, delete, history)

### Negative

1. **Complexity**: Sites must implement JWT token generation
2. **Technical Barrier**: Requires understanding of JWT and key management
3. **Key Management**: Sites must secure their JWT secrets/keys
4. **Clock Sync**: Token expiration requires synchronized clocks
5. **Migration**: Existing anonymous comments cannot be attributed to users retroactively

### Neutral

1. **Backward Compatibility**: Anonymous commenting continues to work
2. **Gradual Adoption**: Sites can adopt authentication at their own pace
3. **Optional Feature**: Sites without authentication needs can ignore it

## Compliance and Privacy

### GDPR Compliance

1. **Right to Access**: Sites can query user data via API
2. **Right to Deletion**: Sites can delete user data via API
3. **Right to Portability**: User data available via JSON API
4. **Data Minimization**: Only store necessary user data
5. **Purpose Limitation**: User data used only for comments/reactions
6. **Storage Limitation**: User data stored only while active

### Data Protection

1. **Encryption**: JWT secrets encrypted at rest
2. **Access Control**: Only site admins can configure authentication
3. **Audit Logging**: Authentication configuration changes logged
4. **Rate Limiting**: Prevent abuse of authentication endpoints

## Alternatives for Low-Technical Sites

For sites that cannot or don't want to implement JWT:

1. **Option A: Anonymous Only**
   - Disable authentication
   - Use IP-based rate limiting
   - Rely on moderation

2. **Option B: Third-Party Auth Widget**
   - Future: Kotomi provides a JavaScript widget
   - Widget handles OAuth with popular providers
   - Widget generates JWT tokens automatically
   - Site owners just configure providers in admin panel

3. **Option C: Simple Email Verification**
   - Future: Kotomi provides optional email verification
   - User enters email to comment
   - Kotomi sends verification link
   - User clicks link to verify comment
   - Not true authentication, but better than anonymous

## References

- [JWT Best Practices (RFC 8725)](https://datatracker.ietf.org/doc/html/rfc8725)
- [JSON Web Token (RFC 7519)](https://datatracker.ietf.org/doc/html/rfc7519)
- [JWKS (JSON Web Key Set)](https://datatracker.ietf.org/doc/html/rfc7517)
- [OAuth 2.0 (RFC 6749)](https://datatracker.ietf.org/doc/html/rfc6749)
- [GDPR Guidelines](https://gdpr.eu/)

## Future Considerations

1. **Webhook Validation**: Add webhook-based validation as an option
2. **Social Login Widget**: Provide optional widget for sites without auth
3. **User Reputation**: Build reputation systems on authenticated users
4. **User Profiles**: Allow users to have profiles across sites (opt-in)
5. **SSO Integration**: Provide pre-built integrations with popular auth providers
6. **API Keys**: Alternative authentication method for automated systems
7. **Two-Factor Authentication**: Support 2FA in JWT claims
8. **Session Management**: Track active sessions and allow revocation

## Implementation Roadmap

### Phase 1: Foundation (v0.4.0)
- [ ] Database schema updates
- [ ] JWT validation library integration
- [ ] Authentication configuration model
- [ ] Admin panel UI for auth configuration

### Phase 2: Core Features (v0.4.1)
- [ ] JWT middleware implementation
- [ ] User model and storage
- [ ] Updated comment/reaction APIs
- [ ] API documentation updates

### Phase 3: Developer Experience (v0.4.2)
- [ ] JWT token examples in documentation
- [ ] Client library updates (if any)
- [ ] Integration guides for popular platforms
- [ ] Testing tools for token validation

### Phase 4: Advanced Features (v0.5.0)
- [ ] User management endpoints
- [ ] Edit/delete own comments
- [ ] User activity history
- [ ] Reputation system foundation

### Phase 5: Enhancements (Future)
- [ ] JavaScript auth widget
- [ ] Pre-built OAuth integrations
- [ ] Email verification option
- [ ] User profile system

## Questions and Answers

**Q: Can sites rotate JWT keys?**  
A: Yes, sites can update their keys in the admin panel at any time. Existing tokens will fail validation after key rotation.

**Q: What happens to existing anonymous comments?**  
A: They remain anonymous. There's no way to retroactively attribute them to users.

**Q: Can a user edit comments across devices?**  
A: Yes, as long as they authenticate with the same user ID on all devices.

**Q: How long should JWT tokens be valid?**  
A: We recommend 5-60 minutes. Sites should refresh tokens as needed.

**Q: Can sites use multiple JWT validation methods?**  
A: Not initially. Each site configures one validation method. Future versions may support fallback methods.

**Q: What if the site's auth system is compromised?**  
A: Sites should immediately rotate their JWT keys in Kotomi admin panel to invalidate all tokens.

**Q: Does this work with mobile apps?**  
A: Yes, mobile apps can generate JWT tokens the same way web apps do.

**Q: Can anonymous and authenticated users co-exist on the same site?**  
A: Yes, if the site enables "allow_anonymous" in the configuration.

## Conclusion

JWT-based delegated authentication provides the best balance of flexibility, security, and developer experience for Kotomi. It allows sites to "bring their own authentication" while maintaining a simple, standards-based integration. The implementation will be gradual, ensuring backward compatibility and allowing sites to adopt authentication at their own pace.
