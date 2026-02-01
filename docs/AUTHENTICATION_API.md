# Kotomi Authentication API Documentation

## Phase 1: External JWT Authentication

This document describes the authentication requirements for Kotomi's comment and reaction APIs as implemented in Phase 1.

## Overview

As of Phase 1, **all write operations** (creating comments, adding reactions, etc.) require JWT-based authentication. Read operations remain unauthenticated for backwards compatibility.

### Authentication Modes

- **External JWT**: Sites provide their own authentication and issue JWT tokens (Phase 1 - implemented)
- **Kotomi-provided Auth**: Built-in authentication service (Future phases)

## JWT Token Format

All JWT tokens must follow this structure:

### Required Claims

```json
{
  "iss": "https://example.com",           // Issuer (your site's domain)
  "sub": "user-uuid-12345",               // Subject (unique user ID)
  "aud": "kotomi",                        // Audience (must be "kotomi")
  "exp": 1738368515,                      // Expiration timestamp (Unix)
  "iat": 1738364915,                      // Issued at timestamp (Unix)
  "kotomi_user": {
    "id": "user-uuid-12345",              // User ID (must match 'sub')
    "name": "Jane Doe"                    // Display name (required)
  }
}
```

### Optional Claims

```json
{
  "kotomi_user": {
    "email": "jane@example.com",          // Email (for moderation)
    "avatar_url": "https://example.com/avatars/jane.jpg",
    "profile_url": "https://example.com/users/jane",
    "verified": true,                      // Email verified status
    "roles": ["member", "premium"]         // User roles
  }
}
```

## Configuration

### Admin API Endpoints

Configure authentication for your site using the admin API:

#### Get Auth Configuration
```http
GET /admin/sites/{siteId}/auth/config
Authorization: Bearer {admin_token}
```

#### Create Auth Configuration
```http
POST /admin/sites/{siteId}/auth/config
Authorization: Bearer {admin_token}
Content-Type: application/json

{
  "auth_mode": "external",
  "jwt_validation_type": "hmac",
  "jwt_secret": "your-secret-key-min-32-characters",
  "jwt_issuer": "https://example.com",
  "jwt_audience": "kotomi",
  "token_expiration_buffer": 60
}
```

#### Update Auth Configuration
```http
PUT /admin/sites/{siteId}/auth/config
Authorization: Bearer {admin_token}
Content-Type: application/json

{
  "jwt_validation_type": "rsa",
  "jwt_public_key": "-----BEGIN PUBLIC KEY-----\n...",
  "jwt_issuer": "https://example.com",
  "jwt_audience": "kotomi"
}
```

#### Delete Auth Configuration
```http
DELETE /admin/sites/{siteId}/auth/config
Authorization: Bearer {admin_token}
```

### Validation Types

#### HMAC (Symmetric Key)
```json
{
  "jwt_validation_type": "hmac",
  "jwt_secret": "your-secret-key-min-32-characters"
}
```

#### RSA (Asymmetric Key)
```json
{
  "jwt_validation_type": "rsa",
  "jwt_public_key": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBg...\n-----END PUBLIC KEY-----"
}
```

#### ECDSA (Asymmetric Key)
```json
{
  "jwt_validation_type": "ecdsa",
  "jwt_public_key": "-----BEGIN PUBLIC KEY-----\n...\n-----END PUBLIC KEY-----"
}
```

#### JWKS (JSON Web Key Set)
```json
{
  "jwt_validation_type": "jwks",
  "jwks_endpoint": "https://example.com/.well-known/jwks.json"
}
```

## API Usage

### Comments

#### Create Comment (Authenticated)
```http
POST /api/v1/site/{siteId}/page/{pageId}/comments
Authorization: Bearer {jwt_token}
Content-Type: application/json

{
  "text": "This is my comment",
  "parent_id": "optional-parent-comment-id"
}
```

**Response:**
```json
{
  "id": "comment-uuid",
  "author": "Jane Doe",
  "author_id": "user-uuid-12345",
  "author_email": "jane@example.com",
  "text": "This is my comment",
  "status": "pending",
  "created_at": "2026-02-01T12:34:56Z",
  "updated_at": "2026-02-01T12:34:56Z"
}
```

#### Get Comments (Unauthenticated)
```http
GET /api/v1/site/{siteId}/page/{pageId}/comments
```

### Reactions

#### Add Reaction (Authenticated)
```http
POST /api/v1/comments/{commentId}/reactions
Authorization: Bearer {jwt_token}
Content-Type: application/json

{
  "allowed_reaction_id": "reaction-uuid"
}
```

**Note:** If the user already has this reaction, it will be toggled off (removed).

#### Add Page Reaction (Authenticated)
```http
POST /api/v1/pages/{pageId}/reactions
Authorization: Bearer {jwt_token}
Content-Type: application/json

{
  "allowed_reaction_id": "reaction-uuid"
}
```

#### Get Reactions (Unauthenticated)
```http
GET /api/v1/comments/{commentId}/reactions
GET /api/v1/comments/{commentId}/reactions/counts
GET /api/v1/pages/{pageId}/reactions
GET /api/v1/pages/{pageId}/reactions/counts
```

#### Remove Reaction (Authenticated)
```http
DELETE /api/v1/reactions/{reactionId}
Authorization: Bearer {jwt_token}
```

## Error Responses

### 401 Unauthorized
```json
{
  "error": "Authentication required"
}
```

Returned when:
- No Authorization header is provided
- JWT token is missing
- JWT token is invalid
- JWT token has expired
- Site auth configuration not found

### 400 Bad Request
```json
{
  "error": "Invalid request body"
}
```

### 500 Internal Server Error
```json
{
  "error": "Failed to add comment"
}
```

## Security Best Practices

### Token Generation

1. **Use strong secrets**: HMAC secrets should be at least 32 characters
2. **Short expiration**: Tokens should expire within 5-60 minutes
3. **Secure storage**: Never expose JWT secrets in client-side code
4. **HTTPS only**: Always transmit tokens over HTTPS

### Token Validation

1. **Validate all claims**: Check iss, aud, exp, and sub
2. **Grace period**: Configure `token_expiration_buffer` for clock skew (default: 60s)
3. **Key rotation**: Regularly rotate HMAC secrets or RSA keys
4. **Revocation**: Rotate keys immediately if compromised

## Examples

### Node.js

See `scripts/generate_jwt.js` for a complete example.

```javascript
const jwt = require('jsonwebtoken');

const token = jwt.sign({
  iss: 'https://example.com',
  sub: 'user-123',
  aud: 'kotomi',
  exp: Math.floor(Date.now() / 1000) + 3600,
  iat: Math.floor(Date.now() / 1000),
  kotomi_user: {
    id: 'user-123',
    name: 'John Doe',
    email: 'john@example.com'
  }
}, 'your-secret-key', { algorithm: 'HS256' });
```

### Python

See `scripts/generate_jwt.py` for a complete example.

```python
import jwt
import time

token = jwt.encode({
    'iss': 'https://example.com',
    'sub': 'user-123',
    'aud': 'kotomi',
    'exp': int(time.time()) + 3600,
    'iat': int(time.time()),
    'kotomi_user': {
        'id': 'user-123',
        'name': 'John Doe',
        'email': 'john@example.com'
    }
}, 'your-secret-key', algorithm='HS256')
```

### Go

```go
import (
    "time"
    "github.com/golang-jwt/jwt/v5"
)

token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
    "iss": "https://example.com",
    "sub": "user-123",
    "aud": "kotomi",
    "exp": time.Now().Add(time.Hour).Unix(),
    "iat": time.Now().Unix(),
    "kotomi_user": map[string]interface{}{
        "id":    "user-123",
        "name":  "John Doe",
        "email": "john@example.com",
    },
})

tokenString, err := token.SignedString([]byte("your-secret-key"))
```

## Migration from Pre-Auth Version

If you have an existing Kotomi deployment without authentication:

1. **Configure authentication** in the admin panel
2. **Update your client code** to:
   - Generate JWT tokens after user login
   - Include `Authorization: Bearer {token}` header in API requests
   - Handle 401 errors (redirect to login)
3. **Test thoroughly** before deploying to production

### Breaking Changes

- Comments no longer accept `author` field in request body (derived from JWT)
- Reactions use `user_id` instead of IP-based `user_identifier`
- All POST/DELETE endpoints require authentication

## Troubleshooting

### "Authentication required" error
- Ensure Authorization header is present
- Verify token format: `Bearer {token}`
- Check token hasn't expired

### "Invalid token" error
- Verify JWT secret/public key matches site configuration
- Check issuer and audience claims
- Ensure kotomi_user claim is present

### "Site auth config not found" error
- Configure authentication in admin panel
- Verify site ID in URL is correct

## Future Enhancements (Not in Phase 1)

- Kotomi-provided authentication service
- Social login (Google, GitHub, Twitter)
- Magic link authentication
- User profile management
- Session management
- JWKS auto-refresh
