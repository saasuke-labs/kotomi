# ADR 001: User Authentication for Comments and Reactions

**Status:** Proposed  
**Date:** 2026-01-31  
**Updated:** 2026-02-01 (Added Kotomi-provided authentication option for sites without auth)  
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
Many Kotomi deployments are embedded into existing websites that already have their own authentication systems (Auth0, Firebase, custom OAuth, SAML, etc.). These sites need a way to "bring their own authentication" while Kotomi handles the comment and reaction functionality.

However, **most static sites don't have authentication systems** in place. Setting up Auth0, Firebase, or similar services can be complex and time-consuming for simple blogs and documentation sites. These sites need an easy-to-setup authentication solution that works out of the box.

**Key Requirements:**
1. Require authenticated users for all comments and reactions
2. **Provide built-in authentication for sites without their own auth system** (easy setup)
3. Allow sites with existing auth to integrate via JWT tokens (bring your own auth)
4. Support user profile management (edit/delete own comments, view own reactions)
5. Enable moderation based on user reputation/history
6. Respect privacy regulations (GDPR, CCPA)
7. Provide flexibility for various authentication providers

**Important Note:** Since Kotomi has not been released yet, we do not need to maintain backward compatibility with anonymous commenting. All users must authenticate to comment or react.

## Decision Drivers

* **Ease of Setup**: Most static sites don't have authentication - Kotomi should provide an easy option
* **Flexibility**: Sites with existing auth must be able to integrate seamlessly
* **Security**: Authentication must be secure and tamper-proof
* **Privacy**: User data must be handled according to privacy regulations
* **User Experience**: Authentication should be seamless and not disruptive
* **Developer Experience**: Integration should be straightforward for site developers
* **Authentication Required**: All commenting and reactions require authenticated users
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
- Not suitable for simple static sites without existing auth infrastructure

### Option 4: Kotomi-Provided Authentication (For Sites Without Auth)

Kotomi provides a built-in authentication service for sites that don't have their own auth system:

**Architecture:**
```
┌─────────────────┐         ┌──────────────────┐         ┌───────────────┐
│   Site's Web    │         │    Kotomi Auth   │         │    Kotomi     │
│   Application   │         │     Service      │         │   Comments    │
└────────┬────────┘         └────────┬─────────┘         └───────┬───────┘
         │                           │                           │
         │ 1. User clicks login      │                           │
         │─────────────────────────>│                           │
         │                           │                           │
         │ 2. Email/password or      │                           │
         │    social login           │                           │
         │<─────────────────────────│                           │
         │                           │                           │
         │ 3. Kotomi returns JWT     │                           │
         │    (stored in cookie)     │                           │
         │<─────────────────────────│                           │
         │                           │                           │
         │ 4. Comment API requests   │                           │
         │    include JWT cookie     │                           │
         │───────────────────────────────────────────────────>│
         │                           │                           │
         │                           │   5. Validate JWT         │
         │                           │      internally           │
         │                           │                           │
         │ 6. Process request        │                           │
         │<───────────────────────────────────────────────────│
```

**Features:**
- Email/password authentication with password reset
- Social login integration (Google, GitHub, Twitter/X)
- Magic link authentication (passwordless)
- User profile management
- Email verification
- Session management
- Built-in UI widgets for login/signup

**Pros:**
- ✅ Zero configuration for simple sites (works out of the box)
- ✅ No need for sites to implement authentication
- ✅ Built-in UI components (login forms, signup, profile)
- ✅ Handles all auth complexity (password hashing, email verification, etc.)
- ✅ Ideal for static sites, blogs, documentation sites
- ✅ Social login support without site owner needing OAuth setup
- ✅ Professional authentication experience
- ✅ Users can have identity across multiple Kotomi-powered sites

**Cons:**
- Kotomi becomes an identity provider (more operational responsibility)
- Sites don't control the authentication experience
- Centralized user database (privacy considerations)
- Users must create Kotomi account even if they have site account
- More complex than pure JWT delegation

### Option 5: Webhook-Based Validation

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

**Chosen option:** Hybrid Approach - **Option 3 (JWT Delegation) + Option 4 (Kotomi-Provided Auth)**

This hybrid approach provides the best of both worlds:

1. **For sites with existing authentication** → Use Option 3 (JWT-Based Delegated Authentication)
   - Sites maintain full control
   - Works with any auth system
   - No dependency on Kotomi's auth service

2. **For sites without authentication** → Use Option 4 (Kotomi-Provided Authentication)
   - Zero configuration setup
   - Professional authentication experience
   - Built-in UI components
   - Social login support

**Implementation Strategy:**
- Kotomi-provided auth is treated as "just another authentication provider"
- Both modes use the same JWT token format internally
- Sites choose their auth mode during initial setup
- Can migrate between modes if needed (with user data export/import)

This approach satisfies the requirement that most static sites don't have authentication while still supporting advanced sites with their own auth infrastructure.

### Implementation Details

#### 1. Site Configuration

Each site in Kotomi will have authentication configuration:

```go
type SiteAuthConfig struct {
    ID                    string    `json:"id"`
    SiteID                string    `json:"site_id"`
    AuthMode              string    `json:"auth_mode"`          // "kotomi" or "external"
    // Fields for external JWT validation (auth_mode = "external")
    JWTValidationType     string    `json:"jwt_validation_type,omitempty"` // "hmac", "rsa", "ecdsa", "jwks"
    JWTSecret             string    `json:"jwt_secret,omitempty"`          // For HMAC (symmetric) - stored encrypted
    JWTPublicKey          string    `json:"jwt_public_key,omitempty"`      // For RSA/ECDSA (asymmetric)
    JWKSEndpoint          string    `json:"jwks_endpoint,omitempty"`       // For JWKS (JSON Web Key Set) URL
    JWTIssuer             string    `json:"jwt_issuer,omitempty"`          // Expected issuer claim
    JWTAudience           string    `json:"jwt_audience,omitempty"`        // Expected audience claim
    TokenExpirationBuffer int       `json:"token_expiration_buffer"`       // Grace period in seconds
    // Fields for Kotomi-provided auth (auth_mode = "kotomi")
    EnableEmailAuth       bool      `json:"enable_email_auth"`      // Enable email/password login
    EnableSocialAuth      bool      `json:"enable_social_auth"`     // Enable social login
    EnableMagicLink       bool      `json:"enable_magic_link"`      // Enable passwordless magic link
    SocialProviders       []string  `json:"social_providers,omitempty"` // ["google", "github", "twitter"]
    RequireEmailVerify    bool      `json:"require_email_verify"`   // Require email verification
    CreatedAt             time.Time `json:"created_at"`
    UpdatedAt             time.Time `json:"updated_at"`
}
```

**Auth Modes:**
- `"kotomi"`: Use Kotomi-provided authentication (default, easy setup)
- `"external"`: Use external JWT-based authentication (bring your own auth)

**Note:** Authentication is always required. Sites choose between Kotomi-provided auth or external JWT validation.

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

#### 3. Kotomi-Provided Authentication

For sites using `auth_mode = "kotomi"`, Kotomi provides a complete authentication service:

**Authentication Methods:**

1. **Email/Password Authentication**
   - Standard email and password registration
   - Secure password hashing (bcrypt)
   - Password reset via email
   - Email verification required (optional)

2. **Social Login**
   - Google OAuth
   - GitHub OAuth
   - Twitter/X OAuth
   - Site admin configures which providers to enable
   - No OAuth setup required by site owner

3. **Magic Link (Passwordless)**
   - User enters email
   - Receives login link via email
   - Click link to authenticate
   - No password required

**User Experience:**

Kotomi provides embeddable UI components:
```html
<!-- Login widget -->
<div id="kotomi-auth"></div>
<script src="https://kotomi.example.com/auth-widget.js"></script>
<script>
  KotomiAuth.init({
    siteId: 'your-site-id',
    container: '#kotomi-auth',
    onLogin: (user) => {
      console.log('User logged in:', user);
    }
  });
</script>
```

**API Endpoints:**

- `POST /api/v1/auth/signup` - Create new account
- `POST /api/v1/auth/login` - Login with email/password
- `POST /api/v1/auth/logout` - Logout
- `POST /api/v1/auth/password-reset` - Request password reset
- `POST /api/v1/auth/verify-email` - Verify email address
- `POST /api/v1/auth/magic-link` - Request magic link
- `GET /api/v1/auth/social/{provider}` - Initiate social login
- `GET /api/v1/auth/social/{provider}/callback` - Social login callback
- `GET /api/v1/auth/user` - Get current user info
- `PUT /api/v1/auth/user` - Update user profile

**Token Management:**

Kotomi auth issues JWT tokens with the same format as external auth:
- Stored in HTTP-only cookies (secure by default)
- Refresh tokens for long-lived sessions
- Token rotation on refresh
- Automatic token validation in comment APIs

#### 4. API Changes (Unified for Both Auth Modes)

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
    Author      string    `json:"author"`         // Display name from JWT
    AuthorID    string    `json:"author_id"`      // User ID from JWT (required)
    Email       string    `json:"email,omitempty"` // User email (for moderation)
    Text        string    `json:"text"`
    ParentID    string    `json:"parent_id,omitempty"`
    Status      string    `json:"status"`
    ModeratedBy string    `json:"moderated_by,omitempty"`
    ModeratedAt time.Time `json:"moderated_at,omitempty"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

**Note:** `author_id` is always required since all comments must be from authenticated users.

**d) Reaction Model Updates:**

```go
type Reaction struct {
    ID                string    `json:"id"`
    PageID            string    `json:"page_id,omitempty"`
    CommentID         string    `json:"comment_id,omitempty"`
    AllowedReactionID string    `json:"allowed_reaction_id"`
    UserID            string    `json:"user_id"`           // User ID from JWT (required)
    CreatedAt         time.Time `json:"created_at"`
}
```

**Note:** `user_id` is always required since all reactions must be from authenticated users. The legacy `user_identifier` field for IP addresses is removed.

#### 5. Authentication Middleware

Unified middleware handles both auth modes:

```go
func JWTAuthMiddleware(siteAuthConfig *SiteAuthConfig) gin.HandlerFunc {
    return func(c *gin.Context) {
        var token string
        
        // 1. Extract token from Authorization header or cookie
        authHeader := c.GetHeader("Authorization")
        if authHeader != "" {
            token = strings.TrimPrefix(authHeader, "Bearer ")
        } else if siteAuthConfig.AuthMode == "kotomi" {
            // For Kotomi auth, check cookie
            cookie, err := c.Cookie("kotomi_auth_token")
            if err == nil {
                token = cookie
            }
        }
        
        // 2. If no token, reject with 401
        if token == "" {
            c.JSON(401, gin.H{"error": "Authentication required"})
            c.Abort()
            return
        }
        
        // 3. Validate JWT based on auth mode
        var user *User
        var err error
        
        if siteAuthConfig.AuthMode == "kotomi" {
            // Validate using Kotomi's internal key
            user, err = validateKotomiToken(token, siteAuthConfig.SiteID)
        } else {
            // Validate using external configuration
            user, err = validateExternalJWT(token, siteAuthConfig)
        }
        
        if err != nil {
            c.JSON(401, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }
        
        // 4. Store user info in context
        c.Set("authenticated_user", user)
        
        c.Next()
    }
}
```

**Note:** Both auth modes use the same JWT token format internally. The middleware adapts based on the site's auth mode configuration.

#### 6. API Endpoints

**New Endpoints (Admin - Site Configuration):**

- `POST /api/v1/admin/sites/{siteId}/auth/config` - Configure authentication for a site
- `GET /api/v1/admin/sites/{siteId}/auth/config` - Get authentication configuration
- `PUT /api/v1/admin/sites/{siteId}/auth/config` - Update authentication configuration
- `POST /api/v1/admin/sites/{siteId}/auth/mode` - Switch between kotomi/external auth mode
- `GET /api/v1/admin/sites/{siteId}/users` - List users who have commented/reacted
- `GET /api/v1/admin/sites/{siteId}/users/{userId}` - Get user details
- `DELETE /api/v1/admin/sites/{siteId}/users/{userId}` - Ban/remove a user

**New Endpoints (Kotomi Authentication - Public):**

- `POST /api/v1/auth/signup` - Register new user (Kotomi auth mode)
- `POST /api/v1/auth/login` - Login with email/password
- `POST /api/v1/auth/logout` - Logout and invalidate token
- `POST /api/v1/auth/password-reset` - Request password reset email
- `POST /api/v1/auth/magic-link` - Request magic link login
- `POST /api/v1/auth/verify-email` - Verify email with token
- `GET /api/v1/auth/social/{provider}` - Start social login flow
- `GET /api/v1/auth/social/{provider}/callback` - Handle social login callback
- `GET /api/v1/auth/user` - Get current user profile
- `PUT /api/v1/auth/user` - Update user profile

**Modified Endpoints:**

Existing comment and reaction endpoints work with both auth modes:
1. Accept JWT token in Authorization header OR cookie (for Kotomi auth)
2. Validate token based on site's auth mode
3. Extract and store user_id for all comments and reactions
4. Return 401 Unauthorized if token is missing or invalid

#### 7. Database Schema Changes

**New Table: site_auth_configs**

```sql
CREATE TABLE site_auth_configs (
    id TEXT PRIMARY KEY,
    site_id TEXT NOT NULL UNIQUE,
    auth_mode TEXT NOT NULL DEFAULT 'kotomi', -- 'kotomi' or 'external'
    -- External JWT validation fields (for auth_mode = 'external')
    jwt_validation_type TEXT,           -- 'hmac', 'rsa', 'ecdsa', 'jwks'
    jwt_secret TEXT,                     -- Encrypted at rest
    jwt_public_key TEXT,
    jwks_endpoint TEXT,
    jwt_issuer TEXT,
    jwt_audience TEXT,
    token_expiration_buffer INTEGER DEFAULT 60,
    -- Kotomi auth fields (for auth_mode = 'kotomi')
    enable_email_auth BOOLEAN DEFAULT TRUE,
    enable_social_auth BOOLEAN DEFAULT TRUE,
    enable_magic_link BOOLEAN DEFAULT TRUE,
    social_providers TEXT,               -- JSON array: ["google", "github", "twitter"]
    require_email_verify BOOLEAN DEFAULT TRUE,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (site_id) REFERENCES sites(id) ON DELETE CASCADE
);
```

**New Table: kotomi_users (for Kotomi-provided auth)**

```sql
CREATE TABLE kotomi_users (
    id TEXT PRIMARY KEY,                 -- User UUID
    email TEXT NOT NULL UNIQUE,
    email_verified BOOLEAN DEFAULT FALSE,
    password_hash TEXT,                  -- NULL for social-only or magic link users
    name TEXT NOT NULL,
    avatar_url TEXT,
    provider TEXT,                       -- 'email', 'google', 'github', 'twitter'
    provider_id TEXT,                    -- ID from OAuth provider
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    last_login DATETIME
);

CREATE INDEX idx_kotomi_users_email ON kotomi_users(email);
CREATE INDEX idx_kotomi_users_provider ON kotomi_users(provider, provider_id);
```

**New Table: kotomi_sessions**

```sql
CREATE TABLE kotomi_sessions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    token_hash TEXT NOT NULL,            -- Hash of JWT token
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES kotomi_users(id) ON DELETE CASCADE
);

CREATE INDEX idx_kotomi_sessions_user ON kotomi_sessions(user_id);
CREATE INDEX idx_kotomi_sessions_token ON kotomi_sessions(token_hash);
CREATE INDEX idx_kotomi_sessions_expires ON kotomi_sessions(expires_at);
```

**New Table: kotomi_email_verifications**

```sql
CREATE TABLE kotomi_email_verifications (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    token TEXT NOT NULL UNIQUE,
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES kotomi_users(id) ON DELETE CASCADE
);

CREATE INDEX idx_kotomi_email_verifications_token ON kotomi_email_verifications(token);
```

**New Table: kotomi_password_resets**

```sql
CREATE TABLE kotomi_password_resets (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    token TEXT NOT NULL UNIQUE,
    expires_at DATETIME NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    created_at DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES kotomi_users(id) ON DELETE CASCADE
);

CREATE INDEX idx_kotomi_password_resets_token ON kotomi_password_resets(token);
```

**Existing Table: users (unified for both auth modes)**

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
-- Add new columns:
ALTER TABLE comments ADD COLUMN author_id TEXT NOT NULL;
ALTER TABLE comments ADD COLUMN email TEXT;

-- Add index for author_id
CREATE INDEX idx_comments_author_id ON comments(site_id, author_id);

-- Add foreign key
ALTER TABLE comments ADD FOREIGN KEY (site_id, author_id) REFERENCES users(site_id, id) ON DELETE CASCADE;
```

**Modified Table: reactions**

```sql
-- Replace user_identifier with user_id:
ALTER TABLE reactions DROP COLUMN user_identifier;
ALTER TABLE reactions ADD COLUMN user_id TEXT NOT NULL;

-- Add index for user_id
CREATE INDEX idx_reactions_user_id ON reactions(user_id);

-- Add foreign key
ALTER TABLE reactions ADD FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Ensure unique constraint: one reaction per user per comment/page
CREATE UNIQUE INDEX idx_reactions_unique_comment 
    ON reactions(comment_id, allowed_reaction_id, user_id) 
    WHERE comment_id IS NOT NULL;

CREATE UNIQUE INDEX idx_reactions_unique_page 
    ON reactions(page_id, allowed_reaction_id, user_id) 
    WHERE page_id IS NOT NULL;
```

**Note:** Since Kotomi hasn't been released, we don't need migration scripts. The schema is designed from scratch to require authentication.

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
- Apply rate limits per authenticated user_id
- Track user activity for abuse detection
- Implement stricter limits for new or unverified users

**d) User Privacy:**
- Email addresses are optional and used only for moderation
- User IDs are site-specific (no cross-site tracking)
- Sites control what user data is included in JWT
- GDPR compliance: users can request deletion via site owner

**e) Token Expiration:**
- Sites should issue short-lived tokens (5-60 minutes recommended)
- Tokens should be refreshed by the site's auth system
- Kotomi will reject expired tokens (with configurable grace period)

#### 8. Implementation Strategy

**Since Kotomi has not been released yet, we implement authentication from the start:**

**Phase 1: Core Implementation**
- Create authentication configuration per site
- Implement JWT validation for all supported methods (HMAC, RSA, ECDSA, JWKS)
- Enforce authentication on all comment and reaction endpoints
- Store user information from JWT claims

**Phase 2: User Management**
- Track user first seen / last seen timestamps
- Store user profiles with name, email, avatar, etc.
- Enable user-based moderation features
- Implement user activity history

**Phase 3: Enhanced Features**
- Allow users to edit their own comments
- Allow users to delete their own comments
- Show user badges/verification status
- Implement user reputation system foundations

#### 9. Developer Experience

**Option A: Use Kotomi-Provided Authentication (Easiest)**

**Step 1: Enable Kotomi Auth in Admin Panel**
```
1. Log in to Kotomi admin panel
2. Navigate to Site Settings → Authentication
3. Select "Use Kotomi Authentication" (default)
4. Choose authentication methods:
   ☑ Email/Password
   ☑ Social Login (Google, GitHub, Twitter)
   ☑ Magic Link
5. Configure options:
   ☑ Require email verification
6. Save configuration
```

**Step 2: Add Kotomi Auth Widget to Your Site**
```html
<!DOCTYPE html>
<html>
<head>
  <title>My Blog</title>
</head>
<body>
  <!-- Your site content -->
  
  <!-- Kotomi Auth Widget -->
  <div id="kotomi-auth"></div>
  
  <!-- Kotomi Comments Widget -->
  <div id="kotomi-comments"></div>
  
  <!-- Load Kotomi SDK -->
  <script src="https://kotomi.example.com/sdk.js"></script>
  <script>
    // Initialize auth
    KotomiAuth.init({
      siteId: 'your-site-id',
      container: '#kotomi-auth',
      onLogin: (user) => {
        console.log('User logged in:', user);
        // Initialize comments after login
        initComments();
      }
    });
    
    // Initialize comments
    function initComments() {
      Kotomi.Comments.init({
        siteId: 'your-site-id',
        pageId: window.location.pathname,
        container: '#kotomi-comments'
      });
    }
  </script>
</body>
</html>
```

**That's it!** No backend code required. Kotomi handles all authentication.

---

**Option B: Use External Authentication (Bring Your Own Auth)**

**Step 1: Configure External JWT in Admin Panel**
```
1. Log in to Kotomi admin panel
2. Navigate to Site Settings → Authentication
3. Choose validation method: HMAC or RSA
4. For HMAC: Generate a shared secret (or provide your own)
5. For RSA: Upload your public key
6. Set issuer (your domain) and audience (kotomi)
7. Save configuration
```

**Note:** Authentication is always required. All users must provide valid JWT tokens to comment or react.

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

**Example 1: Static Blog (Kotomi Auth - Easy Setup)**
```yaml
auth_mode: "kotomi"
enable_email_auth: true
enable_social_auth: true
social_providers: ["google", "github"]
enable_magic_link: true
require_email_verify: true
```

**Use Case:** Simple blog or documentation site without existing auth. Zero backend code required.

---

**Example 2: Enterprise Site (External Auth with Auth0)**
```yaml
auth_mode: "external"
jwt_validation_type: "jwks"
jwks_endpoint: "https://example.auth0.com/.well-known/jwks.json"
jwt_issuer: "https://example.auth0.com/"
jwt_audience: "kotomi"
```

**Use Case:** Corporate website with existing Auth0 integration. Bring your own auth.

---

**Example 3: Open Source Project (External Auth with GitHub)**
```yaml
auth_mode: "external"
jwt_validation_type: "rsa"
jwt_public_key: "-----BEGIN PUBLIC KEY-----\n..."
jwt_issuer: "https://github.com"
jwt_audience: "kotomi"
```

**Use Case:** OSS project where users already have GitHub accounts. Use GitHub OAuth + JWT.

---

**Example 4: Custom Platform (External Auth with HMAC)**
```yaml
auth_mode: "external"
jwt_validation_type: "hmac"
jwt_secret: "your-secret-key-min-32-chars"  # Stored encrypted
jwt_issuer: "https://community.example.com"
jwt_audience: "kotomi"
```

**Use Case:** Custom platform with proprietary auth system. Simple symmetric key validation.

---

**Comparison:**

| Auth Mode | Setup Time | Backend Required | Best For |
|-----------|------------|------------------|----------|
| Kotomi | 5 minutes | No | Static sites, blogs, simple docs |
| External (JWT) | 30-60 minutes | Yes | Sites with existing auth, enterprises |

**Note:** All sites require authentication. Choose between easy Kotomi auth or flexible external JWT auth.

## Consequences

### Positive

1. **Dual Approach**: Serves both simple static sites and complex platforms
2. **Easy Setup**: Static sites can use Kotomi auth with zero backend code
3. **Flexibility**: Advanced sites can bring their own authentication
4. **Security**: Kotomi auth provides professional-grade security out of the box
5. **Privacy**: User data handled according to regulations
6. **Performance**: Stateless JWT validation is fast for both modes
7. **Scalability**: No additional infrastructure bottlenecks
8. **Developer-Friendly**: Both modes well-documented with examples
9. **User Experience**: Seamless authentication (integrated or delegated)
10. **User Features**: Enables edit/delete own comments, user profiles, reputation
11. **Quality Control**: All interactions from authenticated, trackable users
12. **Unified Architecture**: Both modes use same JWT format internally

### Negative

1. **Dual Complexity**: Must maintain two authentication modes
2. **Operational Burden**: Kotomi becomes identity provider (for Kotomi auth mode)
3. **User Accounts**: Users need Kotomi account for Kotomi-auth sites
4. **Storage Overhead**: Must store passwords, sessions for Kotomi auth
5. **Email Infrastructure**: Need reliable email service for verifications/resets
6. **OAuth Management**: Kotomi must maintain OAuth integrations for social login
7. **Security Responsibility**: Password breaches, account security for Kotomi auth
8. **External JWT Setup**: Still requires technical knowledge for external mode

### Neutral

1. **Pre-Release Decision**: Made before v1.0 release, no migration needed
2. **Mode Selection**: Sites choose at setup, can migrate later
3. **Default Mode**: Kotomi auth is default (easier for most users)

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

## Solution for Sites Without Authentication

**Kotomi-provided authentication solves this problem!**

Sites without authentication can use Kotomi's built-in auth service:

✅ **Zero Backend Code**: Just add the Kotomi auth widget to your HTML  
✅ **5-Minute Setup**: Configure in admin panel, copy/paste widget code  
✅ **Professional Experience**: Email/password, social login, magic links  
✅ **Secure by Default**: Password hashing, session management, email verification  
✅ **No OAuth Setup**: Social login works without site owner needing OAuth credentials  

**Example for Static Blog:**
```html
<!-- Add to your HTML template -->
<div id="kotomi-auth"></div>
<script src="https://kotomi.example.com/sdk.js"></script>
<script>
  KotomiAuth.init({
    siteId: 'your-site-id',
    container: '#kotomi-auth'
  });
</script>
```

This makes Kotomi accessible to **all static sites**, not just those with sophisticated authentication infrastructure.

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

**Q: Which auth mode should I use?**  
A: Use Kotomi auth if you don't have existing authentication. Use external JWT if you already have an auth system.

**Q: Can I switch auth modes later?**  
A: Yes, with user data export/import. Contact support for migration assistance.

**Q: Does Kotomi auth work with static site generators?**  
A: Yes! It's designed specifically for static sites (Hugo, Jekyll, Next.js static, etc.). Just add the widget script.

**Q: Can sites rotate JWT keys (external mode)?**  
A: Yes, sites can update their keys in the admin panel at any time. Existing tokens will fail validation after key rotation.

**Q: Can a user edit comments across devices?**  
A: Yes, as long as they authenticate with the same user ID on all devices.

**Q: How long should JWT tokens be valid?**  
A: We recommend 5-60 minutes for external JWT. Kotomi auth manages this automatically.

**Q: What if the site's auth system is compromised (external mode)?**  
A: Sites should immediately rotate their JWT keys in Kotomi admin panel to invalidate all tokens.

**Q: Does this work with mobile apps?**  
A: Yes. Mobile apps can use Kotomi auth SDK or generate JWT tokens for external mode.

**Q: Is anonymous commenting supported?**  
A: No. Since Kotomi has not been released yet, we made the decision to require authentication for all interactions. This simplifies the implementation and ensures better quality control.

**Q: What if users don't want to create a Kotomi account?**  
A: For Kotomi auth mode, offer social login (Google, GitHub) for one-click access. For external mode, users use your site's existing auth.

**Q: Can I use Kotomi auth for multiple sites?**  
A: Yes. Users can have a single Kotomi account across multiple Kotomi-powered sites (optional cross-site identity).

**Q: What about GDPR and user data?**  
A: Both modes are GDPR-compliant. Users can export data and request deletion via the admin panel or API.

## Conclusion

The hybrid authentication approach provides the best balance of ease-of-use and flexibility for Kotomi:

**For Most Users (Static Sites):**
- Use **Kotomi-provided authentication**
- Zero backend code required
- Professional auth experience out of the box
- 5-minute setup

**For Advanced Users (Existing Auth):**
- Use **External JWT authentication**
- Full control over user experience
- Bring your own auth system
- Standards-based integration

**Key Decision:** Since Kotomi has not been released yet, we require authentication for all comments and reactions. However, we recognize that most static sites don't have authentication infrastructure, so we provide Kotomi auth as the default, easy option alongside the flexible external JWT option.

This dual approach makes Kotomi accessible to everyone—from simple blogs to enterprise platforms—while maintaining security and enabling rich user features from day one.
