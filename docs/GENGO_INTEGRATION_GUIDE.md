# Using Kotomi with Gengo - Complete Integration Guide

> **Add dynamic comments and reactions to your Gengo static site**

**🚀 Want to get started quickly?** Check out the [5-Minute Quick Start Guide](GENGO_QUICK_START.md) first!

This guide will walk you through integrating Kotomi (a dynamic commenting and reaction system) with your Gengo static site generator. By the end, you'll have a fully functional commenting system on your static blog.

## Table of Contents

1. [What is Kotomi?](#what-is-kotomi)
2. [Prerequisites](#prerequisites)
3. [Quick Start](#quick-start)
4. [Step-by-Step Setup](#step-by-step-setup)
5. [Gengo Integration](#gengo-integration)
6. [Configuration Options](#configuration-options)
7. [Authentication Setup](#authentication-setup)
8. [Customization](#customization)
9. [Production Deployment](#production-deployment)
10. [Troubleshooting](#troubleshooting)

## What is Kotomi?

Kotomi is a lightweight, privacy-focused commenting and reaction system designed specifically for static websites. It provides:

- 💬 **Comments System** - Enable discussions on your blog posts
- 👍 **Reactions** - Allow readers to react with emoji (👍, ❤️, 🎉, etc.)
- 🔐 **Secure Authentication** - JWT-based authentication for verified users
- 🤖 **AI Moderation** - Optional automatic content moderation
- 🛡️ **Admin Dashboard** - Manage comments and configure settings
- 🪶 **Lightweight** - Minimal JavaScript footprint (~50KB)

### How It Works

```
┌─────────────────────────────────────────────────────────────────────┐
│                      Kotomi Architecture                             │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌─────────────┐         ┌──────────────┐        ┌──────────────┐  │
│  │   Gengo     │ builds  │ Static Site  │ embeds │   Kotomi     │  │
│  │   (SSG)     │────────→│   (HTML)     │───────→│   Widget     │  │
│  └─────────────┘         └──────────────┘        └──────┬───────┘  │
│                                                          │           │
│                                                          │ API calls │
│                                                          ↓           │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                    Kotomi Server                             │   │
│  ├─────────────────────────────────────────────────────────────┤   │
│  │  • REST API (comments, reactions)                           │   │
│  │  • SQLite Database (persistent storage)                     │   │
│  │  • Admin Panel (Auth0)                                      │   │
│  │  • AI Moderation (optional)                                 │   │
│  │  • JWT Authentication                                       │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

## Prerequisites

Before you begin, ensure you have:

- ✅ A **Gengo site** set up and running
- ✅ **Node.js** or **Go** installed (for JWT token generation)
- ✅ **Docker** (recommended) or **Go 1.24+** (for running Kotomi server)
- ✅ Basic understanding of HTML and JavaScript

## Quick Start

For those who want to get started immediately:

### 1. Start Kotomi Server

Using Docker (recommended):

```bash
docker run -p 8080:8080 -v kotomi-data:/app/data ghcr.io/saasuke-labs/kotomi:latest
```

Or using Go:

```bash
git clone https://github.com/saasuke-labs/kotomi.git
cd kotomi
go run cmd/main.go
```

### 2. Add to Your Gengo Template

Add this snippet to your Gengo blog post template (typically in `themes/your-theme/layouts/post.html` or similar):

```html
<!-- Add this in the <head> section -->
<link rel="stylesheet" href="http://localhost:8080/static/kotomi.css">

<!-- Add this where you want comments to appear (usually after blog post content) -->
<div id="kotomi-comments"></div>

<!-- Add this before closing </body> tag -->
<script src="http://localhost:8080/static/kotomi.js"></script>
<script>
  const kotomi = new Kotomi({
    baseUrl: 'http://localhost:8080',
    siteId: 'my-gengo-blog',
    pageId: '{{ .Slug }}',  // Use your Gengo template variable for unique page identifier
    theme: 'light'
  });
  kotomi.render();
</script>
```

### 3. Build and Preview

Build your Gengo site and preview locally:

```bash
gengo build
gengo serve
```

Visit your blog post and you should see the Kotomi comment widget!

## Step-by-Step Setup

Let's walk through a complete setup, including creating your first site and configuring everything properly.

### Step 1: Install and Run Kotomi Server

#### Option A: Using Docker (Recommended)

Docker provides the easiest way to run Kotomi with persistence:

```bash
# Pull the latest image
docker pull ghcr.io/saasuke-labs/kotomi:latest

# Run with persistent storage
docker run -d \
  --name kotomi \
  -p 8080:8080 \
  -v kotomi-data:/app/data \
  -e ENV=production \
  ghcr.io/saasuke-labs/kotomi:latest
```

**Important**: The `-v kotomi-data:/app/data` flag ensures your comments persist across container restarts.

#### Option B: Building from Source

If you prefer to build from source:

```bash
# Clone the repository
git clone https://github.com/saasuke-labs/kotomi.git
cd kotomi

# Install dependencies
go mod download

# Run the server
go run cmd/main.go
```

The server will start on `http://localhost:8080`.

#### Verify the Server is Running

Test that Kotomi is running correctly:

```bash
curl http://localhost:8080/healthz
```

You should see: `{"message":"OK"}`

### Step 2: Create Your Site in Kotomi

Before integrating with Gengo, you need to create a site in Kotomi's admin panel:

1. **Access the Admin Panel**  
   Visit `http://localhost:8080/admin`

2. **Configure Auth0 (for admin access)**
   
   For now, you can skip Auth0 setup for development. In production, you'll want to configure it. See [Authentication Setup](#authentication-setup) below.

3. **Create a Site**
   
   - Navigate to "Sites" in the admin panel
   - Click "Create New Site"
   - Fill in the details:
     - **Site ID**: `my-gengo-blog` (use lowercase, no spaces)
     - **Name**: "My Gengo Blog"
     - **Domain**: `yourdomain.com` (or `localhost` for development)
     - **Description**: Brief description of your blog

4. **Note Your Site ID**
   
   Save the site ID - you'll use it in your Gengo templates.

### Step 3: Integrate with Gengo

Now let's integrate Kotomi into your Gengo blog.

#### Understanding Gengo's Structure

Gengo typically has this structure:

```
your-gengo-site/
├── content/
│   └── posts/
│       ├── first-post.md
│       └── second-post.md
├── themes/
│   └── your-theme/
│       ├── layouts/
│       │   ├── post.html
│       │   └── index.html
│       └── static/
└── gengo.yaml
```

#### Modify Your Post Layout

Edit your Gengo theme's post layout file (e.g., `themes/your-theme/layouts/post.html`):

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>{{ .Title }} - {{ .SiteName }}</title>
  
  <!-- Your existing CSS -->
  <link rel="stylesheet" href="/css/style.css">
  
  <!-- Add Kotomi CSS -->
  <link rel="stylesheet" href="http://localhost:8080/static/kotomi.css">
</head>
<body>
  <article>
    <h1>{{ .Title }}</h1>
    <time>{{ .Date }}</time>
    
    <div class="content">
      {{ .Content }}
    </div>
  </article>

  <!-- Kotomi Comment Section -->
  <section class="comments-section">
    <h2>Comments</h2>
    <div id="kotomi-comments"></div>
  </section>

  <!-- Add Kotomi JavaScript -->
  <script src="http://localhost:8080/static/kotomi.js"></script>
  <script>
    const kotomi = new Kotomi({
      baseUrl: 'http://localhost:8080',
      siteId: 'my-gengo-blog',
      pageId: '{{ .Slug }}',  // Gengo provides .Slug for unique page identifier
      theme: 'light',
      enableReactions: true,
      enableReplies: true
    });
    
    kotomi.render();
  </script>
</body>
</html>
```

**Key Points:**

- Replace `'my-gengo-blog'` with your actual site ID from Step 2
- `{{ .Slug }}` is the Gengo template variable for the post's slug (unique identifier)
- Adjust the template variables based on your Gengo version and configuration

### Step 4: Build Your Site

Build your Gengo site:

```bash
gengo build
```

This generates static HTML files in your output directory (typically `public/` or `dist/`).

### Step 5: Test Locally

Serve your Gengo site locally:

```bash
gengo serve
```

Visit a blog post in your browser. You should see:

1. Your blog post content
2. A "Comments" section below
3. The Kotomi comment widget (showing "No comments yet" if empty)

## Gengo Integration

### Template Variables

Gengo provides several template variables you can use with Kotomi:

| Gengo Variable | Description | Example Use |
|----------------|-------------|-------------|
| `{{ .Slug }}` | Unique post identifier | Use as `pageId` |
| `{{ .Title }}` | Post title | Display above comments |
| `{{ .Date }}` | Post publication date | Show post metadata |
| `{{ .URL }}` | Post URL | Use for sharing |
| `{{ .SiteName }}` | Site name | Branding |

### Example: Using Post Title in Comments Section

```html
<section class="comments-section">
  <h2>Comments on "{{ .Title }}"</h2>
  <p>Share your thoughts about this post!</p>
  <div id="kotomi-comments"></div>
</section>
```

### Adding Comments to Multiple Pages

If you want comments on multiple page types (posts, pages, etc.), you can create a reusable partial in Gengo:

**Create `themes/your-theme/partials/comments.html`:**

```html
<section class="comments-section">
  <h2>Comments</h2>
  <div id="kotomi-comments"></div>
</section>

<script src="http://localhost:8080/static/kotomi.js"></script>
<script>
  const kotomi = new Kotomi({
    baseUrl: 'http://localhost:8080',
    siteId: 'my-gengo-blog',
    pageId: '{{ .Slug }}',
    theme: 'light'
  });
  kotomi.render();
</script>
```

**Then include it in your layouts:**

```html
<!-- In your post layout -->
<article>
  {{ .Content }}
</article>

{{ partial "comments.html" . }}
```

### Conditional Comments

To enable comments only on specific posts, you can use Gengo's front matter:

**In your post (e.g., `content/posts/my-post.md`):**

```yaml
---
title: "My Post"
slug: "my-post"
date: 2024-01-15
enable_comments: true
---

Post content here...
```

**In your layout:**

```html
{{ if .Params.enable_comments }}
  {{ partial "comments.html" . }}
{{ end }}
```

## Configuration Options

### Widget Configuration

The Kotomi widget accepts these configuration options:

```javascript
const kotomi = new Kotomi({
  // Required
  baseUrl: 'http://localhost:8080',      // Kotomi server URL
  siteId: 'my-gengo-blog',                // Your site ID
  pageId: 'unique-page-identifier',       // Unique ID for this page
  
  // Optional
  container: '#kotomi-comments',          // CSS selector for container
  theme: 'light',                         // 'light' or 'dark'
  enableReactions: true,                  // Enable emoji reactions
  enableReplies: true,                    // Enable threaded replies
  jwtToken: null,                         // JWT token (for authenticated users)
  placeholder: 'Write a comment...',      // Comment input placeholder
  submitText: 'Post Comment'              // Submit button text
});
```

### Server Configuration

Configure Kotomi server using environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `DB_PATH` | SQLite database path | `./kotomi.db` |
| `CORS_ORIGINS` | Allowed CORS origins (comma-separated) | `*` |
| `ENV` | Environment (`development` or `production`) | `development` |
| `JWT_SECRET` | Secret for JWT validation (HMAC) | - |
| `JWT_PUBLIC_KEY` | Public key for JWT validation (RSA/ECDSA) | - |

**Example with environment variables:**

```bash
docker run -d \
  --name kotomi \
  -p 8080:8080 \
  -v kotomi-data:/app/data \
  -e PORT=8080 \
  -e DB_PATH=/app/data/kotomi.db \
  -e CORS_ORIGINS=https://yourdomain.com,https://www.yourdomain.com \
  -e ENV=production \
  -e JWT_SECRET=your-secret-key-here \
  ghcr.io/saasuke-labs/kotomi:latest
```

### CORS Configuration

When deploying to production, configure CORS to allow requests from your domain:

```bash
export CORS_ORIGINS=https://yourdomain.com,https://www.yourdomain.com
```

Or in Docker:

```bash
-e CORS_ORIGINS=https://yourdomain.com,https://www.yourdomain.com
```

## Authentication Setup

Kotomi requires JWT authentication for write operations (posting comments, adding reactions). Here's how to set it up:

### Understanding JWT Authentication

- **Read operations** (viewing comments) are public - no authentication required
- **Write operations** (posting, editing, deleting comments, reactions) require JWT tokens
- You generate JWT tokens in your authentication system and pass them to Kotomi

### Step 1: Choose a JWT Method

Kotomi supports three JWT signing methods:

1. **HMAC (Shared Secret)** - Simple, good for single-application setups
2. **RSA (Public Key)** - Better security, use for distributed systems
3. **ECDSA (Public Key)** - Modern alternative to RSA

For most Gengo blogs, **HMAC is the simplest option**.

### Step 2: Generate a Secret Key

Generate a secure random secret:

```bash
# Using OpenSSL
openssl rand -base64 32
```

Save this secret securely - you'll use it in both your authentication system and Kotomi.

### Step 3: Configure Kotomi with Your Secret

```bash
# For Docker
docker run -d \
  --name kotomi \
  -p 8080:8080 \
  -v kotomi-data:/app/data \
  -e JWT_SECRET=your-generated-secret-here \
  ghcr.io/saasuke-labs/kotomi:latest
```

### Step 4: Generate JWT Tokens

When users log into your site, generate a JWT token with this payload:

**Required JWT Claims:**

```json
{
  "iss": "yourdomain.com",
  "aud": "kotomi",
  "sub": "user-unique-id",
  "exp": 1234567890,
  "kotomi_user": {
    "id": "user-unique-id",
    "name": "John Doe",
    "email": "john@example.com"
  }
}
```

**Example: Generate JWT in Node.js**

```javascript
const jwt = require('jsonwebtoken');

function generateKotomiToken(user) {
  const payload = {
    iss: 'yourdomain.com',
    aud: 'kotomi',
    sub: user.id,
    exp: Math.floor(Date.now() / 1000) + (60 * 60 * 24), // 24 hours
    kotomi_user: {
      id: user.id,
      name: user.name,
      email: user.email
    }
  };
  
  return jwt.sign(payload, process.env.JWT_SECRET, { algorithm: 'HS256' });
}

// Usage
const token = generateKotomiToken({
  id: 'user123',
  name: 'John Doe',
  email: 'john@example.com'
});
```

**Example: Generate JWT in Python**

```python
import jwt
import time

def generate_kotomi_token(user, secret_key):
    payload = {
        'iss': 'yourdomain.com',
        'aud': 'kotomi',
        'sub': user['id'],
        'exp': int(time.time()) + (60 * 60 * 24),  # 24 hours
        'kotomi_user': {
            'id': user['id'],
            'name': user['name'],
            'email': user['email']
        }
    }
    
    return jwt.encode(payload, secret_key, algorithm='HS256')

# Usage
token = generate_kotomi_token({
    'id': 'user123',
    'name': 'John Doe',
    'email': 'john@example.com'
}, 'your-secret-key')
```

**Example: Generate JWT in Go**

```go
package main

import (
    "github.com/golang-jwt/jwt/v5"
    "time"
)

func generateKotomiToken(userID, userName, userEmail, secret string) (string, error) {
    claims := jwt.MapClaims{
        "iss": "yourdomain.com",
        "aud": "kotomi",
        "sub": userID,
        "exp": time.Now().Add(24 * time.Hour).Unix(),
        "kotomi_user": map[string]string{
            "id":    userID,
            "name":  userName,
            "email": userEmail,
        },
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(secret))
}
```

### Step 5: Pass Token to Kotomi Widget

In your Gengo template, pass the JWT token to the widget:

```html
<script>
  // Get the JWT token from your authentication system
  // This could be from a cookie, localStorage, or server-rendered variable
  const userToken = getUserToken(); // Your function to get the token
  
  const kotomi = new Kotomi({
    baseUrl: 'http://localhost:8080',
    siteId: 'my-gengo-blog',
    pageId: '{{ .Slug }}',
    jwtToken: userToken  // Pass the token here
  });
  
  kotomi.render();
</script>
```

**Example with Cookie:**

```html
<script>
  // Read JWT from cookie
  function getCookie(name) {
    const value = `; ${document.cookie}`;
    const parts = value.split(`; ${name}=`);
    if (parts.length === 2) return parts.pop().split(';').shift();
  }
  
  const kotomi = new Kotomi({
    baseUrl: 'http://localhost:8080',
    siteId: 'my-gengo-blog',
    pageId: '{{ .Slug }}',
    jwtToken: getCookie('kotomi_token')
  });
  
  kotomi.render();
</script>
```

### Authentication Flow Summary

1. User logs into your site → Your auth system generates JWT token
2. Token is stored (cookie, localStorage, etc.)
3. Gengo page loads → Widget reads token
4. User posts comment → Widget sends token with request
5. Kotomi validates token → Comment is saved

For complete authentication details, see [docs/AUTHENTICATION_API.md](AUTHENTICATION_API.md).

## Customization

### Styling the Widget

Kotomi provides CSS classes you can override to match your Gengo theme:

```css
/* Override Kotomi styles in your theme CSS */

/* Comment container */
.kotomi-comments {
  font-family: your-font-family;
}

/* Comment item */
.kotomi-comment {
  border-left: 3px solid your-accent-color;
  padding: 1rem;
  margin-bottom: 1rem;
}

/* Comment author */
.kotomi-comment-author {
  font-weight: bold;
  color: your-primary-color;
}

/* Comment text */
.kotomi-comment-text {
  color: your-text-color;
  line-height: 1.6;
}

/* Submit button */
.kotomi-submit-btn {
  background-color: your-primary-color;
  color: white;
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 4px;
}

/* Reactions */
.kotomi-reaction {
  cursor: pointer;
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
}

.kotomi-reaction:hover {
  background-color: your-hover-color;
}
```

### Dark Mode Support

Kotomi supports both light and dark themes. You can switch dynamically:

```html
<script>
  // Detect user's theme preference
  const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
  
  const kotomi = new Kotomi({
    baseUrl: 'http://localhost:8080',
    siteId: 'my-gengo-blog',
    pageId: '{{ .Slug }}',
    theme: prefersDark ? 'dark' : 'light'
  });
  
  kotomi.render();
</script>
```

### Custom Reactions

Configure custom emoji reactions in the admin panel:

1. Go to `http://localhost:8080/admin/sites/your-site-id/reactions`
2. Add reactions:
   - 👍 Thumbs Up
   - ❤️ Heart
   - 😂 Laugh
   - 🎉 Celebrate
   - 🤔 Thinking
   - 🙏 Thank you

3. Choose whether reactions apply to pages, comments, or both

## Production Deployment

### Deploying Kotomi Server

#### Option 1: Docker on a VPS

Deploy Kotomi on any VPS (DigitalOcean, AWS EC2, etc.):

```bash
# SSH into your server
ssh user@your-server.com

# Pull and run Kotomi
docker run -d \
  --name kotomi \
  -p 8080:8080 \
  --restart unless-stopped \
  -v kotomi-data:/app/data \
  -e ENV=production \
  -e CORS_ORIGINS=https://yourdomain.com \
  -e JWT_SECRET=your-production-secret \
  ghcr.io/saasuke-labs/kotomi:latest

# Set up nginx reverse proxy (recommended)
```

#### Option 2: Docker Compose

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  kotomi:
    image: ghcr.io/saasuke-labs/kotomi:latest
    container_name: kotomi
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - kotomi-data:/app/data
    environment:
      - ENV=production
      - PORT=8080
      - DB_PATH=/app/data/kotomi.db
      - CORS_ORIGINS=https://yourdomain.com,https://www.yourdomain.com
      - JWT_SECRET=${JWT_SECRET}

volumes:
  kotomi-data:
```

Run with:

```bash
docker-compose up -d
```

#### Option 3: Build from Source

```bash
# Clone and build
git clone https://github.com/saasuke-labs/kotomi.git
cd kotomi
go build -o kotomi cmd/main.go

# Run with systemd (example)
sudo cp kotomi /usr/local/bin/
sudo systemctl enable kotomi
sudo systemctl start kotomi
```

### Reverse Proxy Setup (Nginx)

Create an nginx configuration for SSL and domain mapping:

```nginx
server {
    listen 80;
    server_name comments.yourdomain.com;
    
    # Redirect to HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name comments.yourdomain.com;
    
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Update Your Gengo Site for Production

Update your template to use the production URL:

```html
<script>
  const kotomi = new Kotomi({
    baseUrl: 'https://comments.yourdomain.com',  // Production URL
    siteId: 'my-gengo-blog',
    pageId: '{{ .Slug }}'
  });
  kotomi.render();
</script>
```

### Build and Deploy Your Gengo Site

```bash
# Build your Gengo site
gengo build

# Deploy to your hosting (example with rsync)
rsync -avz --delete public/ user@yourserver:/var/www/yourdomain.com/

# Or use your hosting provider's deployment method
# (Netlify, Vercel, GitHub Pages, etc.)
```

## Troubleshooting

### Comments Not Showing

**Problem**: The widget appears but says "No comments yet" even though comments exist.

**Solutions**:

1. **Check Site ID and Page ID**
   
   Verify you're using the correct IDs:
   ```javascript
   console.log('Site ID:', 'my-gengo-blog');
   console.log('Page ID:', '{{ .Slug }}');
   ```

2. **Check Browser Console**
   
   Open browser DevTools (F12) and look for errors in the Console tab.

3. **Verify API Endpoints**
   
   Test the API directly:
   ```bash
   curl http://localhost:8080/api/v1/site/my-gengo-blog/page/test-post/comments
   ```

4. **Check CORS Settings**
   
   If running on different domains, ensure CORS is configured:
   ```bash
   -e CORS_ORIGINS=https://yourdomain.com
   ```

### Cannot Post Comments

**Problem**: Users see comments but cannot post new ones.

**Solutions**:

1. **Check JWT Token**
   
   Ensure a valid JWT token is provided:
   ```javascript
   console.log('JWT Token:', kotomi.config.jwtToken);
   ```

2. **Verify JWT Secret**
   
   Ensure Kotomi is configured with the correct JWT secret:
   ```bash
   docker exec kotomi env | grep JWT_SECRET
   ```

3. **Check Token Expiration**
   
   JWT tokens expire. Verify the `exp` claim is in the future:
   ```javascript
   const payload = JSON.parse(atob(token.split('.')[1]));
   console.log('Expires:', new Date(payload.exp * 1000));
   ```

### Widget Not Loading

**Problem**: The widget doesn't appear at all.

**Solutions**:

1. **Check JavaScript Loading**
   
   Verify the script loads without errors:
   ```html
   <!-- Check Network tab in DevTools -->
   <script src="http://localhost:8080/static/kotomi.js"></script>
   ```

2. **Check Container Element**
   
   Ensure the container exists:
   ```javascript
   console.log(document.querySelector('#kotomi-comments'));
   ```

3. **Check for JavaScript Errors**
   
   Look for errors in the browser console.

### CORS Errors

**Problem**: Browser shows CORS-related errors.

**Solution**:

Configure CORS origins properly:

```bash
# Allow multiple origins
docker run -d \
  -e CORS_ORIGINS=https://yourdomain.com,https://www.yourdomain.com,http://localhost:3000 \
  ghcr.io/saasuke-labs/kotomi:latest
```

### Database Issues

**Problem**: Comments disappear after container restart.

**Solution**:

Ensure you're using a persistent volume:

```bash
# Correct - with volume
docker run -v kotomi-data:/app/data ghcr.io/saasuke-labs/kotomi:latest

# Wrong - no persistence
docker run ghcr.io/saasuke-labs/kotomi:latest
```

### Styling Issues

**Problem**: Widget doesn't match your site's theme.

**Solution**:

1. Add custom CSS overrides (see [Customization](#customization))
2. Use the correct theme option (`light` or `dark`)
3. Check for CSS conflicts with your theme

## Next Steps

Now that you have Kotomi integrated with your Gengo blog, consider:

1. **Setting Up Moderation** - Configure AI moderation or manual approval in the admin panel
2. **Enabling Notifications** - Set up email notifications for new comments
3. **Adding Analytics** - Track engagement through the analytics dashboard
4. **Customizing Reactions** - Add custom emoji reactions relevant to your blog
5. **Implementing Authentication** - Set up user authentication for verified comments

## Additional Resources

- [Kotomi Main README](../README.md) - Complete feature overview
- [Authentication API Documentation](AUTHENTICATION_API.md) - Detailed JWT setup
- [Frontend Widget Documentation](../frontend/README.md) - Widget API reference
- [Security Guide](security.md) - Security best practices
- [Gengo Documentation](https://github.com/saasuke-labs/gengo) - Static site generator docs

## Support

- **Issues**: Report bugs at [GitHub Issues](https://github.com/saasuke-labs/kotomi/issues)
- **Discussions**: Ask questions in [GitHub Discussions](https://github.com/saasuke-labs/kotomi/discussions)
- **Documentation**: Full docs at [saasuke-labs.com/kotomi](https://saasuke-labs.com/kotomi)

---

**Happy blogging with Kotomi! 🎉**
