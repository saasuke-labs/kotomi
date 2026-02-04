# Kotomi Beta Tester Guide

Welcome to the Kotomi beta program! This guide will help you deploy and configure Kotomi for your website.

## Table of Contents

1. [System Requirements](#system-requirements)
2. [Deployment Options](#deployment-options)
3. [Configuration Reference](#configuration-reference)
4. [Quick Start Tutorial](#quick-start-tutorial)
5. [Authentication Setup](#authentication-setup)
6. [Frontend Integration](#frontend-integration)
7. [Troubleshooting](#troubleshooting)
8. [Known Limitations](#known-limitations)

## System Requirements

### Minimum Requirements

- **Operating System**: Linux, macOS, or Windows
- **Memory**: 512 MB RAM minimum, 1 GB recommended
- **Storage**: 100 MB for application + database storage
- **Network**: HTTPS recommended for production

### Software Dependencies

- **Docker** (recommended): Docker 20.10+ and Docker Compose 2.0+
- **Or Go**: Go 1.25+ (for building from source)

## Deployment Options

Kotomi can be deployed in several ways. Choose the option that best fits your infrastructure.

### Option 1: Docker (Recommended)

**Pros**: Easy setup, isolated environment, production-ready
**Cons**: Requires Docker installed

#### Using Docker Compose

1. Create a `docker-compose.yml` file:

```yaml
version: '3.8'

services:
  kotomi:
    image: gcr.io/your-project/kotomi:0.1.0-beta.1
    ports:
      - "8080:8080"
    environment:
      # Auth0 Configuration
      AUTH0_DOMAIN: your-domain.auth0.com
      AUTH0_CLIENT_ID: your-client-id
      AUTH0_CLIENT_SECRET: your-client-secret
      AUTH0_CALLBACK_URL: http://localhost:8080/callback
      
      # Session Security
      SESSION_SECRET: your-strong-random-secret-at-least-32-chars
      
      # CORS Configuration
      CORS_ALLOWED_ORIGINS: http://localhost:3000,https://yourdomain.com
      CORS_ALLOWED_METHODS: GET,POST,PUT,DELETE,OPTIONS
      CORS_ALLOWED_HEADERS: Content-Type,Authorization
      
      # Rate Limiting
      RATE_LIMIT_GET: 100
      RATE_LIMIT_POST: 5
      
      # Database
      DB_PATH: /app/data/kotomi.db
      
      # Server
      PORT: 8080
    volumes:
      - kotomi-data:/app/data
    restart: unless-stopped

volumes:
  kotomi-data:
```

2. Start the service:

```bash
docker-compose up -d
```

3. Check the logs:

```bash
docker-compose logs -f kotomi
```

4. Access the admin panel at `http://localhost:8080/admin/dashboard`

#### Using Docker CLI

```bash
# Create a volume for data persistence
docker volume create kotomi-data

# Run the container
docker run -d \
  --name kotomi \
  -p 8080:8080 \
  -v kotomi-data:/app/data \
  -e AUTH0_DOMAIN=your-domain.auth0.com \
  -e AUTH0_CLIENT_ID=your-client-id \
  -e AUTH0_CLIENT_SECRET=your-client-secret \
  -e AUTH0_CALLBACK_URL=http://localhost:8080/callback \
  -e SESSION_SECRET=your-strong-random-secret \
  -e DB_PATH=/app/data/kotomi.db \
  gcr.io/your-project/kotomi:0.1.0-beta.1

# View logs
docker logs -f kotomi

# Stop the container
docker stop kotomi

# Start the container
docker start kotomi
```

### Option 2: Google Cloud Run

**Pros**: Fully managed, auto-scaling, HTTPS included
**Cons**: Requires Google Cloud account

#### Prerequisites

1. Google Cloud account with billing enabled
2. `gcloud` CLI installed and authenticated
3. Auth0 account configured

#### Deployment Steps

1. Set up environment:

```bash
# Set your project ID
gcloud config set project YOUR_PROJECT_ID

# Enable required APIs
gcloud services enable run.googleapis.com
gcloud services enable cloudbuild.googleapis.com
```

2. Deploy to Cloud Run:

```bash
gcloud run deploy kotomi \
  --image gcr.io/your-project/kotomi:0.1.0-beta.1 \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated \
  --set-env-vars "AUTH0_DOMAIN=your-domain.auth0.com" \
  --set-env-vars "AUTH0_CLIENT_ID=your-client-id" \
  --set-secrets "AUTH0_CLIENT_SECRET=auth0-secret:latest" \
  --set-secrets "SESSION_SECRET=session-secret:latest" \
  --set-env-vars "CORS_ALLOWED_ORIGINS=https://yourdomain.com" \
  --set-env-vars "DB_PATH=/app/data/kotomi.db" \
  --memory 512Mi \
  --cpu 1 \
  --min-instances 0 \
  --max-instances 10
```

3. Configure secrets in Google Secret Manager:

```bash
# Create secrets
echo -n "your-auth0-client-secret" | gcloud secrets create auth0-secret --data-file=-
echo -n "your-session-secret-min-32-chars" | gcloud secrets create session-secret --data-file=-

# Grant access to Cloud Run
gcloud secrets add-iam-policy-binding auth0-secret \
  --member="serviceAccount:YOUR_PROJECT_NUMBER-compute@developer.gserviceaccount.com" \
  --role="roles/secretmanager.secretAccessor"

gcloud secrets add-iam-policy-binding session-secret \
  --member="serviceAccount:YOUR_PROJECT_NUMBER-compute@developer.gserviceaccount.com" \
  --role="roles/secretmanager.secretAccessor"
```

4. Get your service URL:

```bash
gcloud run services describe kotomi --region us-central1 --format 'value(status.url)'
```

**Note**: For production use, configure a custom domain and update your Auth0 callback URL.

### Option 3: Binary Deployment

**Pros**: No Docker required, direct control
**Cons**: Manual dependency management

#### Building from Source

1. Clone the repository:

```bash
git clone https://github.com/saasuke-labs/kotomi.git
cd kotomi
```

2. Build the binary:

```bash
go build -o kotomi ./cmd/main.go
```

3. Create environment file `.env`:

```bash
# Auth0 Configuration
AUTH0_DOMAIN=your-domain.auth0.com
AUTH0_CLIENT_ID=your-client-id
AUTH0_CLIENT_SECRET=your-client-secret
AUTH0_CALLBACK_URL=http://localhost:8080/callback

# Session Security
SESSION_SECRET=your-strong-random-secret-min-32-chars

# CORS Configuration
CORS_ALLOWED_ORIGINS=http://localhost:3000
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=Content-Type,Authorization

# Database
DB_PATH=./kotomi.db

# Server
PORT=8080
```

4. Run the application:

```bash
# Load environment and run
export $(cat .env | xargs)
./kotomi
```

## Configuration Reference

### Required Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `AUTH0_DOMAIN` | Your Auth0 domain | `your-domain.auth0.com` |
| `AUTH0_CLIENT_ID` | Auth0 application client ID | `abc123xyz...` |
| `AUTH0_CLIENT_SECRET` | Auth0 application client secret | `secret-key-here` |
| `AUTH0_CALLBACK_URL` | OAuth callback URL | `http://localhost:8080/callback` |
| `SESSION_SECRET` | Secret for session encryption (min 32 chars) | `strong-random-secret-32-chars-min` |

### Optional Environment Variables

| Variable | Description | Default | Production Value |
|----------|-------------|---------|------------------|
| `PORT` | HTTP server port | `8080` | `8080` |
| `DB_PATH` | SQLite database file path | `./kotomi.db` | `/app/data/kotomi.db` |
| `CORS_ALLOWED_ORIGINS` | Allowed origins for CORS | `*` | `https://yourdomain.com` |
| `CORS_ALLOWED_METHODS` | Allowed HTTP methods | `GET,POST,PUT,DELETE,OPTIONS` | Same |
| `CORS_ALLOWED_HEADERS` | Allowed request headers | `Content-Type,Authorization` | Same |
| `CORS_ALLOW_CREDENTIALS` | Allow credentials in CORS | `false` | `false` |
| `RATE_LIMIT_GET` | GET requests per minute per IP | `100` | `100` (adjust as needed) |
| `RATE_LIMIT_POST` | POST/PUT/DELETE per minute per IP | `5` | `5` (adjust as needed) |

### Security Best Practices

1. **Never commit secrets to version control**
2. **Use strong random values** for `SESSION_SECRET` (minimum 32 characters)
3. **Restrict CORS origins** in production (don't use `*`)
4. **Use HTTPS** in production (configure via reverse proxy)
5. **Set appropriate rate limits** based on expected traffic
6. **Regularly backup** your database file

## Quick Start Tutorial

This tutorial walks you through setting up Kotomi from scratch to posting your first comment.

### Step 1: Set Up Auth0 (5 minutes)

1. Create a free Auth0 account at [auth0.com](https://auth0.com)

2. Create a new application:
   - Click "Applications" → "Create Application"
   - Name: "Kotomi Beta"
   - Type: "Regular Web Application"
   - Click "Create"

3. Configure application settings:
   - **Allowed Callback URLs**: `http://localhost:8080/callback`
   - **Allowed Logout URLs**: `http://localhost:8080`
   - **Allowed Web Origins**: `http://localhost:8080`
   - Click "Save Changes"

4. Note your credentials:
   - **Domain**: Found at top of settings page
   - **Client ID**: Found in "Basic Information"
   - **Client Secret**: Found in "Basic Information" (click "Show")

### Step 2: Deploy Kotomi (5 minutes)

Using Docker Compose (easiest):

1. Create `docker-compose.yml` (see Docker section above)

2. Update environment variables with your Auth0 credentials

3. Start the service:

```bash
docker-compose up -d
```

4. Verify it's running:

```bash
# Check logs
docker-compose logs kotomi

# Should see: "Server starting on port 8080"
```

5. Test the health check:

```bash
curl http://localhost:8080/healthz
# Should return: {"status":"healthy","database":"connected"}
```

### Step 3: Access Admin Panel (2 minutes)

1. Open your browser to `http://localhost:8080/admin/dashboard`

2. Click "Login" - you'll be redirected to Auth0

3. Sign up or log in with your Auth0 account

4. You'll be redirected back to the admin dashboard

### Step 4: Create Your First Site (3 minutes)

1. In the admin dashboard, click "Sites" in the navigation

2. Click "Add New Site"

3. Fill in the form:
   - **Name**: Your website name (e.g., "My Blog")
   - **Domain**: Your website domain (e.g., "myblog.com")
   - **Description**: Optional description

4. Click "Create Site"

5. Note the **Site ID** - you'll need this for API calls

### Step 5: Add a Page (2 minutes)

1. Click on your newly created site

2. Click "Add Page"

3. Fill in the form:
   - **URL**: Page URL (e.g., "/blog/my-first-post")
   - **Title**: Page title (e.g., "My First Post")

4. Click "Create Page"

5. Note the **Page ID** - you'll need this for comments

### Step 6: Post Your First Comment (5 minutes)

#### Option A: Using the API directly

```bash
# Set your IDs
SITE_ID="your-site-id"
PAGE_ID="your-page-id"

# Post a comment
curl -X POST http://localhost:8080/api/comments \
  -H "Content-Type: application/json" \
  -d '{
    "siteId": "'$SITE_ID'",
    "pageId": "'$PAGE_ID'",
    "author": "Test User",
    "email": "test@example.com",
    "text": "This is my first comment!",
    "parentId": null
  }'
```

#### Option B: Using the Swagger UI

1. Navigate to `http://localhost:8080/swagger/index.html`

2. Find the `POST /api/comments` endpoint

3. Click "Try it out"

4. Fill in the request body with your site ID and page ID

5. Click "Execute"

### Step 7: View Comments in Admin Panel (1 minute)

1. Go back to the admin dashboard

2. Click "Comments" in the navigation

3. You should see your test comment

4. Try moderating it (approve/reject)

### Step 8: Integrate with Your Website

See the [Frontend Integration](#frontend-integration) section below.

## Authentication Setup

### Auth0 Configuration

Kotomi uses Auth0 for admin authentication. Here's how to set it up properly:

#### Creating an Auth0 Application

1. **Sign up** for Auth0 at [auth0.com](https://auth0.com) (free tier available)

2. **Create a new application**:
   - Go to Applications → Create Application
   - Choose "Regular Web Application"
   - Name it (e.g., "Kotomi Production")

3. **Configure the application**:
   ```
   Allowed Callback URLs:
     https://yourdomain.com/callback
     http://localhost:8080/callback (for testing)
   
   Allowed Logout URLs:
     https://yourdomain.com
     http://localhost:8080 (for testing)
   
   Allowed Web Origins:
     https://yourdomain.com
     http://localhost:8080 (for testing)
   ```

4. **Save your credentials**:
   - Domain (e.g., `dev-abc123.us.auth0.com`)
   - Client ID
   - Client Secret (keep this secure!)

#### JWT for Public API

For the public comments API, you can implement JWT token generation in your backend:

**Important**: The public API does NOT require authentication currently. JWT support is planned for future versions to allow users to authenticate with your existing auth system.

Current approach:
- Public API accepts any author name (rate-limited)
- Admin moderates comments through the admin panel
- Future: JWT will allow verified user identity

## Frontend Integration

### Static Site Integration (HTML + JavaScript)

Add this to your HTML page where you want comments to appear:

```html
<!DOCTYPE html>
<html>
<head>
    <title>My Blog Post</title>
</head>
<body>
    <article>
        <h1>My Blog Post Title</h1>
        <p>Your blog content here...</p>
    </article>

    <!-- Comments Section -->
    <div id="kotomi-comments">
        <h2>Comments</h2>
        <div id="comments-list"></div>
        
        <!-- Comment Form -->
        <form id="comment-form">
            <input type="text" id="author-name" placeholder="Your name" required>
            <input type="email" id="author-email" placeholder="Your email (optional)">
            <textarea id="comment-text" placeholder="Write your comment..." required></textarea>
            <button type="submit">Post Comment</button>
        </form>
    </div>

    <script>
        // Configuration
        const KOTOMI_API = 'http://localhost:8080'; // Change to your Kotomi URL
        const SITE_ID = 'your-site-id';
        const PAGE_ID = 'your-page-id';

        // Load comments
        async function loadComments() {
            try {
                const response = await fetch(
                    `${KOTOMI_API}/api/sites/${SITE_ID}/pages/${PAGE_ID}/comments`
                );
                const comments = await response.json();
                displayComments(comments);
            } catch (error) {
                console.error('Error loading comments:', error);
            }
        }

        // Display comments
        function displayComments(comments) {
            const commentsList = document.getElementById('comments-list');
            commentsList.innerHTML = comments.map(comment => `
                <div class="comment">
                    <div class="comment-author">${escapeHtml(comment.author)}</div>
                    <div class="comment-date">${new Date(comment.created_at).toLocaleString()}</div>
                    <div class="comment-text">${escapeHtml(comment.text)}</div>
                </div>
            `).join('');
        }

        // Post comment
        document.getElementById('comment-form').addEventListener('submit', async (e) => {
            e.preventDefault();
            
            const author = document.getElementById('author-name').value;
            const email = document.getElementById('author-email').value;
            const text = document.getElementById('comment-text').value;

            try {
                const response = await fetch(`${KOTOMI_API}/api/comments`, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({
                        siteId: SITE_ID,
                        pageId: PAGE_ID,
                        author: author,
                        email: email,
                        text: text,
                        parentId: null
                    })
                });

                if (response.ok) {
                    alert('Comment submitted! It will appear after moderation.');
                    document.getElementById('comment-form').reset();
                    loadComments(); // Reload comments
                } else {
                    alert('Error submitting comment. Please try again.');
                }
            } catch (error) {
                console.error('Error posting comment:', error);
                alert('Error submitting comment. Please try again.');
            }
        });

        // Helper function to escape HTML
        function escapeHtml(text) {
            const div = document.createElement('div');
            div.textContent = text;
            return div.innerHTML;
        }

        // Load comments on page load
        loadComments();
    </script>

    <style>
        #kotomi-comments {
            max-width: 800px;
            margin: 2rem auto;
            padding: 1rem;
        }
        .comment {
            border-left: 3px solid #007bff;
            padding: 1rem;
            margin-bottom: 1rem;
            background: #f8f9fa;
        }
        .comment-author {
            font-weight: bold;
            margin-bottom: 0.25rem;
        }
        .comment-date {
            font-size: 0.875rem;
            color: #6c757d;
            margin-bottom: 0.5rem;
        }
        #comment-form {
            display: flex;
            flex-direction: column;
            gap: 0.5rem;
            margin-top: 1rem;
        }
        #comment-form input,
        #comment-form textarea {
            padding: 0.5rem;
            border: 1px solid #ddd;
            border-radius: 4px;
        }
        #comment-form textarea {
            min-height: 100px;
        }
        #comment-form button {
            padding: 0.5rem 1rem;
            background: #007bff;
            color: white;
            border: none;
            border-radius: 4px;
            cursor: pointer;
        }
        #comment-form button:hover {
            background: #0056b3;
        }
    </style>
</body>
</html>
```

### React Integration Example

```javascript
import React, { useState, useEffect } from 'react';

const KotomiComments = ({ siteId, pageId }) => {
  const KOTOMI_API = process.env.REACT_APP_KOTOMI_API || 'http://localhost:8080';
  const [comments, setComments] = useState([]);
  const [author, setAuthor] = useState('');
  const [email, setEmail] = useState('');
  const [text, setText] = useState('');

  useEffect(() => {
    loadComments();
  }, [siteId, pageId]);

  const loadComments = async () => {
    try {
      const response = await fetch(
        `${KOTOMI_API}/api/sites/${siteId}/pages/${pageId}/comments`
      );
      const data = await response.json();
      setComments(data);
    } catch (error) {
      console.error('Error loading comments:', error);
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    try {
      const response = await fetch(`${KOTOMI_API}/api/comments`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          siteId,
          pageId,
          author,
          email,
          text,
          parentId: null
        })
      });

      if (response.ok) {
        alert('Comment submitted! It will appear after moderation.');
        setAuthor('');
        setEmail('');
        setText('');
        loadComments();
      }
    } catch (error) {
      console.error('Error posting comment:', error);
    }
  };

  return (
    <div className="kotomi-comments">
      <h2>Comments</h2>
      
      <div className="comments-list">
        {comments.map(comment => (
          <div key={comment.id} className="comment">
            <div className="comment-author">{comment.author}</div>
            <div className="comment-date">
              {new Date(comment.created_at).toLocaleString()}
            </div>
            <div className="comment-text">{comment.text}</div>
          </div>
        ))}
      </div>

      <form onSubmit={handleSubmit} className="comment-form">
        <input
          type="text"
          placeholder="Your name"
          value={author}
          onChange={(e) => setAuthor(e.target.value)}
          required
        />
        <input
          type="email"
          placeholder="Your email (optional)"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
        />
        <textarea
          placeholder="Write your comment..."
          value={text}
          onChange={(e) => setText(e.target.value)}
          required
        />
        <button type="submit">Post Comment</button>
      </form>
    </div>
  );
};

export default KotomiComments;
```

## Troubleshooting

### Common Issues

#### 1. "Server failed to start" - Port Already in Use

**Symptom**: Error message saying port 8080 is already in use.

**Solution**:
```bash
# Check what's using port 8080
lsof -i :8080  # macOS/Linux
netstat -ano | findstr :8080  # Windows

# Either stop that process or change Kotomi's port
docker-compose up -d -e PORT=8081
```

#### 2. "Auth0 login fails" - Invalid Callback URL

**Symptom**: Auth0 redirects to an error page.

**Solution**:
1. Check your Auth0 application settings
2. Ensure callback URL matches exactly: `http://your-domain:8080/callback`
3. Update environment variable `AUTH0_CALLBACK_URL`
4. Restart Kotomi

#### 3. "Comments not appearing" - CORS Error

**Symptom**: Browser console shows CORS error.

**Solution**:
```bash
# Update CORS_ALLOWED_ORIGINS to include your website
CORS_ALLOWED_ORIGINS=https://yourdomain.com,http://localhost:3000
```

#### 4. "Database locked" - SQLite Concurrent Access

**Symptom**: Error messages about database being locked.

**Solution**:
- This is expected under high load with SQLite
- For beta, this is acceptable
- For production scale, consider PostgreSQL

#### 5. "Rate limit exceeded" - Too Many Requests

**Symptom**: 429 Too Many Requests response.

**Solution**:
```bash
# Increase rate limits
RATE_LIMIT_GET=200
RATE_LIMIT_POST=10
```

### Getting Help

1. **Check the logs**:
   ```bash
   # Docker
   docker-compose logs -f kotomi
   
   # Cloud Run
   gcloud run services logs read kotomi --region us-central1
   ```

2. **Verify configuration**:
   - Check all required environment variables are set
   - Verify Auth0 credentials are correct
   - Test health endpoint: `curl http://localhost:8080/healthz`

3. **Test the API**:
   - Use Swagger UI: `http://localhost:8080/swagger/index.html`
   - Test with curl (see examples in this guide)

4. **Report issues**:
   - GitHub Issues: https://github.com/saasuke-labs/kotomi/issues
   - Include: logs, configuration (redact secrets), steps to reproduce

## Known Limitations

These limitations are acceptable for beta but will be addressed in future releases:

### Functional Limitations
- No email verification for comment authors
- No user profile management UI
- No real-time updates (must refresh)
- No bulk moderation operations
- No webhook support

### Scalability Limitations
- SQLite may not scale to millions of comments
- No horizontal scaling (single instance)
- No caching layer
- No CDN integration

### Deployment Limitations
- Cloud Run only documented (Kubernetes/ECS guides coming)
- No automated backup solution
- No disaster recovery plan
- No multi-region deployment

### Documentation Limitations
- No video tutorials yet
- Limited troubleshooting scenarios
- Few integration examples
- No architecture diagrams

## Beta Feedback

We want to hear from you! Please provide feedback on:

1. **Deployment Experience**
   - Was setup straightforward?
   - Any blockers or confusion?
   - Documentation gaps?

2. **Features**
   - What works well?
   - What's missing?
   - What's confusing?

3. **Performance**
   - Response times acceptable?
   - Any errors or crashes?
   - Database performance?

4. **Use Cases**
   - How are you using Kotomi?
   - What's your traffic volume?
   - Any specific needs?

**How to Provide Feedback**:
- GitHub Issues: https://github.com/saasuke-labs/kotomi/issues
- GitHub Discussions: https://github.com/saasuke-labs/kotomi/discussions
- Email: beta@saasuke-labs.com

## Next Steps

1. **Complete your setup** following this guide
2. **Integrate with your website** using the examples
3. **Test thoroughly** with real usage
4. **Provide feedback** to help us improve
5. **Join the community** in GitHub Discussions

Thank you for being an early adopter! Your feedback will shape the future of Kotomi.

---

**Version**: 0.1.0-beta.1
**Last Updated**: February 4, 2026
**Support**: beta@saasuke-labs.com
