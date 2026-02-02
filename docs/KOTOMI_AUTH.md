# Kotomi Built-in Authentication (Auth0-backed)

## Overview

Kotomi provides built-in authentication for sites that don't have their own authentication system. This authentication is powered by Auth0, ensuring enterprise-grade security without storing passwords locally in Kotomi.

## How It Works

When a site is configured with `auth_mode: "kotomi"`:

1. **User clicks login** → Redirected to Auth0 Universal Login
2. **User authenticates with Auth0** → Can use email/password, Google, GitHub, etc.
3. **Auth0 callback** → Kotomi receives Auth0 user info
4. **User profile created/updated** → Stored in Kotomi database (no passwords)
5. **JWT token issued** → Kotomi issues its own JWT token for API access
6. **Token used for API calls** → Comments and reactions APIs validate the JWT

## Key Benefits

✅ **No Password Storage**: Passwords are never stored in Kotomi  
✅ **Secure by Default**: Auth0 handles all authentication security  
✅ **Multiple Auth Methods**: Email/password, Google, GitHub, Twitter, etc.  
✅ **Easy Setup**: No Auth0 configuration needed by site owners  
✅ **Privacy-Focused**: Only essential user info cached locally  

## API Endpoints

### Get Auth Configuration

```bash
GET /api/v1/auth/config?siteId={siteId}
```

Returns the authentication configuration for a site. Helps clients know which auth flow to use.

**Response:**
```json
{
  "site_id": "my-site",
  "auth_mode": "kotomi",
  "auth0_domain": "kotomi.auth0.com",
  "auth0_client_id": "abc123"
}
```

### Login (Redirect to Auth0)

```bash
GET /api/v1/auth/login?siteId={siteId}&redirect_uri={redirectUri}
```

Redirects the user to Auth0 Universal Login for authentication.

**Parameters:**
- `siteId` (required): The site ID
- `redirect_uri` (optional): Where to redirect after successful auth

### Callback (OAuth Callback)

```bash
GET /api/v1/auth/callback?code={code}&state={state}
```

Handles the OAuth callback from Auth0. This endpoint:
1. Exchanges the authorization code for tokens
2. Fetches user info from Auth0
3. Creates or updates the user in Kotomi
4. Issues a Kotomi JWT token
5. Returns the token to the client

**Response:**
```json
{
  "user": {
    "id": "user-uuid",
    "site_id": "my-site",
    "email": "user@example.com",
    "auth0_sub": "auth0|12345",
    "name": "John Doe",
    "avatar_url": "https://...",
    "is_verified": true
  },
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "rbQV0vimG8o...",
  "expires_at": "2026-02-02T09:00:00Z"
}
```

### Logout

```bash
POST /api/v1/auth/logout
Authorization: Bearer {token}
```

Logs out the user by invalidating their session.

### Get Current User

```bash
GET /api/v1/auth/user
Authorization: Bearer {token}
```

Returns the currently authenticated user's profile.

## Database Schema

User profiles are cached in the `kotomi_auth_users` table:

```sql
CREATE TABLE kotomi_auth_users (
    id TEXT PRIMARY KEY,
    site_id TEXT NOT NULL,
    email TEXT NOT NULL,
    auth0_sub TEXT NOT NULL,        -- Auth0 subject identifier
    name TEXT,
    avatar_url TEXT,
    is_verified INTEGER DEFAULT 0,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    UNIQUE(site_id, auth0_sub)
);
```

**Note:** No `password_hash` field. Passwords are managed by Auth0.

## Client Integration

### JavaScript Example

```javascript
// 1. Get auth configuration
const response = await fetch('/api/v1/auth/config?siteId=my-site');
const config = await response.json();

if (config.auth_mode === 'kotomi') {
  // 2. Redirect to login
  window.location.href = `/api/v1/auth/login?siteId=my-site`;
  
  // 3. After callback, use the token
  const token = localStorage.getItem('kotomi_token');
  
  // 4. Make authenticated API calls
  fetch('/api/v1/site/my-site/page/my-page/comments', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      text: 'My comment'
    })
  });
}
```

### Embeddable Widget (Future)

```html
<div id="kotomi-comments"></div>
<script src="https://kotomi.example.com/widget.js"></script>
<script>
  Kotomi.init({
    siteId: 'my-site',
    pageId: 'blog-post-1',
    container: '#kotomi-comments'
  });
  // Widget handles Auth0 login automatically
</script>
```

## Admin Setup

### Configure a Site for Kotomi Auth

```bash
POST /api/v1/admin/sites/{siteId}/auth/config
Authorization: Bearer {admin-token}

{
  "auth_mode": "kotomi"
}
```

That's it! No JWT secrets, no key management. Kotomi handles everything through Auth0.

### Configure Auth0 (For Kotomi Operators)

Kotomi operators need to set up Auth0 once:

1. Create an Auth0 account at [auth0.com](https://auth0.com)
2. Create a "Regular Web Application"
3. Configure settings:
   - **Allowed Callback URLs**: `https://your-kotomi-instance.com/api/v1/auth/callback`
   - **Allowed Logout URLs**: `https://your-kotomi-instance.com/`
4. Set environment variables:
   ```bash
   export AUTH0_DOMAIN=your-tenant.auth0.com
   export AUTH0_CLIENT_ID=your_client_id
   export AUTH0_CLIENT_SECRET=your_client_secret
   export AUTH0_CALLBACK_URL=https://your-kotomi-instance.com/api/v1/auth/callback
   ```

## Security Considerations

### What's Stored in Kotomi

- ✅ User ID (UUID generated by Kotomi)
- ✅ Auth0 subject ID (`auth0_sub`)
- ✅ Email address
- ✅ Display name
- ✅ Avatar URL
- ✅ Email verification status

### What's NOT Stored

- ❌ Passwords
- ❌ Password hashes
- ❌ Auth0 access tokens (not persisted)
- ❌ Sensitive Auth0 data

### Token Security

- **JWT Tokens**: Signed with HMAC-SHA256
- **Expiration**: 1 hour (configurable)
- **Storage**: HTTP-only cookies (when used in browser)
- **Validation**: Every API request validates the token

### GDPR Compliance

- Users can request data export (includes cached profile)
- Users can request deletion (removes from `kotomi_auth_users`)
- Auth0 handles user consent and data management

## Comparison: Kotomi Auth vs External JWT

| Feature | Kotomi Auth | External JWT |
|---------|-------------|--------------|
| Password Storage | None (Auth0) | None (site handles) |
| Setup Complexity | Zero config for sites | Requires JWT configuration |
| User Management | Through Auth0 | Through site's system |
| Social Login | Included | Site must implement |
| Best For | Simple sites, blogs | Sites with existing auth |

## Migration Path

Sites can start with Kotomi auth and migrate to external JWT later:

1. Export user data from Kotomi
2. Import to site's auth system
3. Update auth_mode to "external"
4. Configure JWT validation
5. Users log in with site's system

## Troubleshooting

### "Authentication not configured"

Check that the site has an entry in `site_auth_configs` with `auth_mode = 'kotomi'`.

### "Failed to exchange code"

Verify Auth0 environment variables are set correctly:
- `AUTH0_DOMAIN`
- `AUTH0_CLIENT_ID`
- `AUTH0_CLIENT_SECRET`

### "User not found"

The user hasn't logged in yet. Direct them to `/api/v1/auth/login?siteId={siteId}`.

## Future Enhancements

- [ ] Magic link authentication (passwordless)
- [ ] 2FA/MFA support
- [ ] Refresh token rotation
- [ ] Session management UI
- [ ] User profile editing
- [ ] Account deletion flow
- [ ] Embeddable login widget

## Related Documentation

- [ADR 001: User Authentication](./adr/001-user-authentication-for-comments-and-reactions.md)
- [External JWT Authentication](./AUTHENTICATION_API.md)
- [Admin Panel Guide](./ADMIN_PANEL.md)
