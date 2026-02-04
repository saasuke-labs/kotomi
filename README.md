# Kotomi

Give your pages a voice

[![Go Version](https://img.shields.io/badge/Go-1.24-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Version](https://img.shields.io/badge/version-0.0.1-blue.svg)](https://github.com/saasuke-labs/kotomi/releases)

> ⚠️ **Early Development**: Kotomi is currently in early development (v0.0.1) and is not recommended for production use yet.

## Features

- 💬 **Comments System** - Enable discussions on your static pages
- 👍 **Reactions System** - Allow users to react to pages and comments with configurable emoji
- 🔐 **JWT Authentication** - Secure user authentication with external JWT support (Phase 1)
- 🤖 **AI Moderation** - Automatic content moderation using OpenAI GPT (optional)
- 📧 **Email Notifications** - Notify site owners and users about comments and moderation events
- 📊 **Analytics & Reporting** - Track engagement metrics, user activity, and trends
- 🔐 **Admin Panel** - Web-based dashboard with Auth0 authentication
- 🏢 **Multi-Site Management** - Manage multiple sites from a single instance
- 📄 **Page Tracking** - Organize comments by pages within sites
- 🛡️ **Comment Moderation** - Approve, reject, or delete comments with ease
- ⚡ **HTMX Interface** - Smooth, no-reload UI updates
- 📖 **OpenAPI/Swagger** - Interactive API documentation (development mode)
- 🪶 **Lightweight** - Built with Go for minimal resource usage
- 🔌 **Easy Integration** - Simple REST API for seamless integration
- 🔒 **Privacy-Focused** - Designed with user privacy in mind

## Architecture

Kotomi is built with simplicity and performance in mind:

- **Go 1.24** - Modern, fast, and efficient
- **SQLite Storage** - Persistent, reliable database with zero configuration
- **REST API** - Standard HTTP endpoints for easy integration
- **HTMX** - Server-side rendering with smooth interactivity
- **Auth0** - Secure authentication for admin panel
- **Docker** - Containerized for easy deployment

## Quick Start

### Prerequisites

- Go 1.24+ (for local development)
- Docker (optional, for containerized deployment)

### Local Development

1. Clone the repository:
```bash
git clone https://github.com/saasuke-labs/kotomi.git
cd kotomi
```

2. Install dependencies:
```bash
go mod download
```

3. Run the server:
```bash
go run cmd/main.go
```

The server will start on `http://localhost:8080` by default.

### Docker Deployment

1. Build the Docker image:
```bash
docker build -t kotomi .
```

2. Run the container:
```bash
docker run -p 8080:8080 -v kotomi-data:/app/data kotomi
```

**Note:** The `-v kotomi-data:/app/data` flag creates a Docker volume to persist your comment database across container restarts.

### GCP Cloud Run Deployment

For production deployment to Google Cloud Platform, see the **[GCP Deployment Setup Guide](docs/GCP_DEPLOYMENT_SETUP.md)**.

The GitHub Actions workflow automatically:
- Runs tests on every push
- Builds and pushes Docker images to Artifact Registry
- Deploys to Cloud Run on main branch

## Documentation & Guides

### 🚀 Quick Starts

- **[5-Minute Gengo Quick Start](docs/GENGO_QUICK_START.md)** - Get Kotomi running on your Gengo blog in 5 minutes
- **[Complete Gengo Integration Guide](docs/GENGO_INTEGRATION_GUIDE.md)** - Comprehensive step-by-step guide for Gengo users

### Integration Guides

- [Frontend Widget Documentation](frontend/README.md) - Widget API reference and configuration
- [Widget Examples](frontend/examples/) - Working examples and templates
- [Authentication API Guide](docs/AUTHENTICATION_API.md) - JWT authentication setup and examples

### Release & Deployment

- **[Release Setup Quick Start](docs/RELEASE_SETUP_QUICKSTART.md)** - 30-minute setup guide for CI/CD pipeline
- [Release Process Documentation](docs/RELEASE_PROCESS.md) - Complete release process with GCP configuration
- [Deployment Monitoring Guide](docs/DEPLOYMENT_MONITORING.md) - Monitor production deployments

### Additional Documentation

- [Security Guide](docs/security.md) - Security best practices and vulnerability reporting
- [OpenAPI/Swagger Documentation](http://localhost:8080/swagger/index.html) - Interactive API documentation (development mode only)

## Running Tests

### Unit Tests

Run all unit tests:
```bash
go test ./pkg/... ./cmd/... -v
```

Run tests with coverage:
```bash
go test ./pkg/... ./cmd/... -cover
```

Run tests for a specific package:
```bash
go test ./pkg/comments/... -v
```

### Integration Tests

Run integration tests:
```bash
go test ./pkg/comments/integration_test.go -v
```

### E2E Tests

E2E tests validate the API endpoints by starting a real server and making HTTP requests. To run E2E tests:

```bash
# Set environment variable to enable E2E tests
export RUN_E2E_TESTS=true
export TEST_MODE=true
export DB_PATH=./kotomi_test.db

# Run E2E tests
go test ./tests/e2e/... -v -timeout=10m
```

**Note:** E2E tests will start a test server automatically on port 8888.

### All Tests

Run all tests (unit, integration, and E2E):
```bash
RUN_E2E_TESTS=true go test ./... -v -cover
```

### With Coverage Report

Generate a detailed coverage report:
```bash
go test ./pkg/... ./cmd/... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

Then open `coverage.html` in your browser.

### Health Check

Verify the server is running:

```bash
curl http://localhost:8080/healthz
```

Response:
```json
{"message":"OK"}
```

## API Documentation

### Interactive API Documentation (Swagger UI)

Kotomi includes an interactive API documentation interface powered by OpenAPI/Swagger, available in **development mode only**.

**Access Swagger UI:**
```bash
# Start the server in development mode (default)
go run cmd/main.go

# Visit the Swagger UI in your browser
# http://localhost:8080/swagger/index.html
```

**Features:**
- 📖 Complete API endpoint documentation
- 🧪 Interactive "Try it out" functionality
- 🔐 Built-in authorization support for protected endpoints
- 📋 Request/response schema definitions
- 🏷️ Organized by endpoint categories (health, comments, reactions)

**Note:** Swagger UI is automatically disabled in production mode (when `ENV=production`). This ensures documentation is only available during development and testing.

**Regenerating Documentation:**
```bash
# Install swag if not already installed
go install github.com/swaggo/swag/cmd/swag@latest

# Generate updated documentation
swag init -g cmd/main.go -o docs
```

### Health Check

**Endpoint:** `GET /healthz`

Returns the health status of the service.

**Response:**
```json
{
  "message": "OK"
}
```

## Admin Panel

Kotomi includes a web-based admin panel for managing sites, pages, and moderating comments. The admin panel uses Auth0 for authentication and provides a smooth HTMX-based interface.

### Accessing the Admin Panel

1. Configure Auth0 (see Configuration section above)
2. Start Kotomi with Auth0 environment variables set
3. Visit `http://localhost:8080/admin`
4. Click "Login with Auth0"
5. Authenticate using your Auth0 credentials

### Features

**Site Management:**
- Create and manage multiple sites
- Track site metadata (name, domain, description)
- View all pages associated with each site
- Delete sites (cascade deletes all associated pages and comments)

**Page Management:**
- Add pages to your sites
- Track page paths and titles
- Edit or remove pages

**Comment Moderation:**
- View all comments across your sites
- Filter by status (pending, approved, rejected)
- Approve or reject comments with one click
- Delete spam or inappropriate comments
- Real-time updates without page refreshes

**Reaction Management:**
- Configure allowed reactions per site
- Set reactions for pages, comments, or both
- Add custom emoji reactions (👍, ❤️, 🎉, etc.)
- View reaction statistics and usage
- Delete reaction types (cascade deletes user reactions)

**Export/Import:**
- Export site data to JSON or CSV formats
- Import previously exported data
- Backup and restore comments and reactions
- Duplicate handling strategies (skip or update)
- Data portability between Kotomi instances

### Admin Panel Routes

- `/admin` - Redirects to dashboard
- `/admin/dashboard` - Overview of sites and pending comments
- `/admin/sites` - List all sites
- `/admin/sites/{siteId}` - View site details and pages
- `/admin/sites/{siteId}/analytics` - View analytics and engagement metrics
- `/admin/sites/{siteId}/reactions` - Manage allowed reactions for a site
- `/admin/sites/{siteId}/comments` - Moderate comments for a site
- `/admin/sites/{siteId}/export` - Export site data
- `/admin/sites/{siteId}/import` - Import site data
- `/login` - Auth0 login
- `/logout` - Logout and clear session

### Analytics & Reporting

Kotomi provides comprehensive analytics for tracking engagement metrics, user activity, and trends. Access the analytics dashboard through the admin panel at `/admin/sites/{siteId}/analytics`.

**Metrics Available:**

- **Comment Metrics**
  - Total comments, pending, approved, rejected counts
  - Approval and rejection rates
  - Daily, weekly, and monthly trends
  - Time-series charts showing comment activity over time

- **User Metrics**
  - Total registered users
  - Active users (today, this week, this month)
  - Top contributors with comment counts

- **Reaction Metrics**
  - Total reactions with daily/weekly/monthly breakdowns
  - Reactions by type with distribution charts
  - Most reacted pages and comments

- **Moderation Metrics**
  - Total moderated comments
  - Auto-rejected vs auto-approved vs manual reviews
  - Average moderation time
  - Spam detection rate

**Features:**

- **Date Range Filtering** - View metrics for custom date ranges
- **Interactive Charts** - Powered by Chart.js for visual insights
- **CSV Export** - Download complete analytics data for external analysis
- **Real-time Updates** - Metrics update based on current database state

**API Endpoints:**

- `GET /admin/sites/{siteId}/analytics` - View analytics dashboard (HTML)
- `GET /admin/sites/{siteId}/analytics/data` - Get analytics data (JSON)
- `GET /admin/sites/{siteId}/analytics/export` - Export analytics to CSV

## API Documentation

### Authentication

**All write operations** (POST/DELETE/PUT) require JWT-based authentication. Read operations (GET) remain unauthenticated.

**Current Status:** External JWT authentication is fully implemented (ADR 001 Option 3). Sites with existing authentication systems can integrate by providing JWT tokens.

📚 **See [Authentication API Documentation](docs/AUTHENTICATION_API.md) for complete details:**
- JWT token format and requirements
- Configuration examples (HMAC, RSA, ECDSA)
- Code examples in Node.js, Python, and Go
- Integration guides and troubleshooting

**Quick Example:**
```bash
# Generate a JWT token (see scripts/generate_jwt.js or scripts/generate_jwt.py)
curl -X POST https://kotomi.example.com/api/v1/site/{siteId}/page/{pageId}/comments \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"text": "This is my comment"}'
```

### API Versioning

Kotomi uses versioned API endpoints to ensure stability and backward compatibility.

**Current Version:** `v1`

**Versioned Endpoints:** All API endpoints should use the `/api/v1/` prefix for the current version.

**Legacy Support:** For backward compatibility, the unversioned `/api/` endpoints are still supported but deprecated. These endpoints return a deprecation warning in the response headers:
- `X-API-Warn: Deprecated API endpoint. Please use /api/v1/ prefix instead.`
- `Deprecation: true`
- `Sunset: Sun, 01 Jun 2026 00:00:00 GMT` (5 months deprecation period)

**Recommendation:** Use versioned endpoints in all new integrations to future-proof your application.

### Health Check

**Endpoint:** `GET /healthz`

Returns the health status of the service.

**Response:**
```json
{
  "message": "OK"
}
```

### Comments API

**Get Comments**

**Endpoint:** `GET /api/v1/site/{siteId}/page/{pageId}/comments`

Retrieve all comments for a specific page.

**Parameters:**
- `siteId` - Unique identifier for your site
- `pageId` - Unique identifier for the page

**Response:**
```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "author": "John Doe",
    "text": "Great article!",
    "parent_id": "",
    "created_at": "2024-01-01T12:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  }
]
```

**Post Comment**

**Endpoint:** `POST /api/v1/site/{siteId}/page/{pageId}/comments`

Create a new comment on a page.

**Parameters:**
- `siteId` - Unique identifier for your site
- `pageId` - Unique identifier for the page

**Request Body:**
```json
{
  "author": "John Doe",
  "text": "This is my comment",
  "parent_id": ""
}
```

**Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "author": "John Doe",
  "text": "This is my comment",
  "parent_id": "",
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:00:00Z"
}
```

### Reactions API

Reactions can be applied to both pages and comments. Site admins can configure which reactions are available for pages vs comments vs both.

**Get Allowed Reactions**

**Endpoint:** `GET /api/v1/site/{siteId}/allowed-reactions[?type=page|comment]`

Retrieve all allowed reaction types for a site. Optionally filter by type.

**Parameters:**
- `siteId` - Unique identifier for your site
- `type` (optional) - Filter by reaction type: `page`, `comment`, or omit for all

**Response:**
```json
[
  {
    "id": "reaction-uuid",
    "site_id": "site-uuid",
    "name": "thumbs_up",
    "emoji": "👍",
    "reaction_type": "page",
    "created_at": "2024-01-01T12:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  },
  {
    "id": "reaction-uuid-2",
    "site_id": "site-uuid",
    "name": "heart",
    "emoji": "❤️",
    "reaction_type": "both",
    "created_at": "2024-01-01T12:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  }
]
```

**Add Comment Reaction (Toggle)**

**Endpoint:** `POST /api/v1/comments/{commentId}/reactions`

Add a reaction to a comment. If the user has already reacted with this type, it will be removed (toggle behavior).

**Parameters:**
- `commentId` - Unique identifier for the comment

**Request Body:**
```json
{
  "allowed_reaction_id": "reaction-uuid"
}
```

**Response (Added):**
```json
{
  "id": "user-reaction-uuid",
  "comment_id": "comment-uuid",
  "allowed_reaction_id": "reaction-uuid",
  "user_identifier": "192.168.1.1",
  "created_at": "2024-01-01T12:00:00Z"
}
```

**Response (Removed):** HTTP 204 No Content

**Get Reaction Counts**

**Endpoint:** `GET /api/v1/comments/{commentId}/reactions/counts`

Get aggregated reaction counts for a comment.

**Parameters:**
- `commentId` - Unique identifier for the comment

**Response:**
```json
[
  {
    "name": "thumbs_up",
    "emoji": "👍",
    "count": 5
  },
  {
    "name": "heart",
    "emoji": "❤️",
    "count": 3
  }
]
```

**Get All Reactions**

**Endpoint:** `GET /api/v1/comments/{commentId}/reactions`

Get all individual reactions for a comment (includes user identifiers).

**Parameters:**
- `commentId` - Unique identifier for the comment

**Response:**
```json
[
  {
    "id": "reaction-instance-uuid",
    "comment_id": "comment-uuid",
    "name": "thumbs_up",
    "emoji": "👍",
    "user_identifier": "192.168.1.1",
    "created_at": "2024-01-01T12:00:00Z"
  }
]
```

**Remove Reaction**

**Endpoint:** `DELETE /api/v1/reactions/{reactionId}`

Remove a specific reaction instance.

**Parameters:**
- `reactionId` - Unique identifier for the reaction instance

**Response:** HTTP 204 No Content

**Add Page Reaction (Toggle)**

**Endpoint:** `POST /api/v1/pages/{pageId}/reactions`

Add a reaction to a page. If the user has already reacted with this type, it will be removed (toggle behavior).

**Parameters:**
- `pageId` - Unique identifier for the page

**Request Body:**
```json
{
  "allowed_reaction_id": "reaction-uuid"
}
```

**Response (Added):**
```json
{
  "id": "user-reaction-uuid",
  "page_id": "page-uuid",
  "allowed_reaction_id": "reaction-uuid",
  "user_identifier": "192.168.1.1",
  "created_at": "2024-01-01T12:00:00Z"
}
```

**Response (Removed):** HTTP 204 No Content

**Get Page Reaction Counts**

**Endpoint:** `GET /api/v1/pages/{pageId}/reactions/counts`

Get aggregated reaction counts for a page.

**Parameters:**
- `pageId` - Unique identifier for the page

**Response:**
```json
[
  {
    "name": "thumbs_up",
    "emoji": "👍",
    "count": 5
  },
  {
    "name": "heart",
    "emoji": "❤️",
    "count": 3
  }
]
```

**Get All Page Reactions**

**Endpoint:** `GET /api/v1/pages/{pageId}/reactions`

Get all individual reactions for a page (includes user identifiers).

**Parameters:**
- `pageId` - Unique identifier for the page

**Response:**
```json
[
  {
    "id": "reaction-instance-uuid",
    "page_id": "page-uuid",
    "name": "thumbs_up",
    "emoji": "👍",
    "user_identifier": "192.168.1.1",
    "created_at": "2024-01-01T12:00:00Z"
  }
]
```

## Frontend Widget

Kotomi includes a JavaScript widget that makes it easy to embed comments and reactions into your static websites.

### Quick Start

Add the widget to your HTML page:

```html
<!-- Include Kotomi CSS -->
<link rel="stylesheet" href="https://your-kotomi-server.com/static/kotomi.css">

<!-- Comment widget container -->
<div id="kotomi-comments"></div>

<!-- Include Kotomi JavaScript -->
<script src="https://your-kotomi-server.com/static/kotomi.js"></script>

<!-- Initialize the widget -->
<script>
  const kotomi = new Kotomi({
    baseUrl: 'https://your-kotomi-server.com',
    siteId: 'your-site-id',
    pageId: 'page-slug',
    container: '#kotomi-comments',
    theme: 'light', // or 'dark'
    jwtToken: null // Optional: Set if you have authentication
  });
  
  kotomi.render();
</script>
```

### Features

- 💬 Display and post comments
- 👍 React to comments with emoji
- 💬 Threaded replies
- 🎨 Light and dark themes
- 📱 Responsive design
- 🔒 JWT authentication support
- 🚀 Zero dependencies

### Configuration Options

| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| `baseUrl` | string | Yes | - | URL of your Kotomi server |
| `siteId` | string | Yes | - | Your site identifier |
| `pageId` | string | Yes | - | Unique identifier for the page |
| `container` | string | No | `#kotomi-comments` | CSS selector for the container |
| `theme` | string | No | `light` | Theme: `light` or `dark` |
| `enableReactions` | boolean | No | `true` | Enable/disable reactions |
| `enableReplies` | boolean | No | `true` | Enable/disable threaded replies |
| `jwtToken` | string | No | `null` | JWT token for authenticated requests |

### Full Documentation

For complete widget documentation, examples, and integration guides, see:
- [Frontend Widget README](frontend/README.md)
- [Widget Examples](frontend/examples/)
- **[Gengo Integration Guide](docs/GENGO_INTEGRATION_GUIDE.md)** - Complete guide for using Kotomi with Gengo static site generator

### Building the Widget

If you need to customize or rebuild the widget:

```bash
cd frontend
./build.sh
```

This will generate distributable files in `frontend/dist/`:
- `kotomi.js` and `kotomi.min.js` - JavaScript bundle
- `kotomi.css` and `kotomi.min.css` - Styles

## Configuration

Kotomi can be configured using environment variables:

### Basic Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `DB_PATH` | Path to SQLite database file | `./kotomi.db` |

### CORS Configuration (Optional)

Configure Cross-Origin Resource Sharing (CORS) for API endpoints:

| Variable | Description | Default |
|----------|-------------|---------|
| `CORS_ALLOWED_ORIGINS` | Comma-separated list of allowed origins (e.g., `https://example.com,https://blog.example.com`) or `*` for all origins | `*` |
| `CORS_ALLOWED_METHODS` | Comma-separated list of allowed HTTP methods | `GET,POST,PUT,DELETE,OPTIONS` |
| `CORS_ALLOWED_HEADERS` | Comma-separated list of allowed headers | `Content-Type,Authorization` |
| `CORS_ALLOW_CREDENTIALS` | Allow credentials in CORS requests | `false` |

**Note:** CORS is only applied to `/api/*` routes. Admin panel routes (`/admin/*`) are not affected by CORS configuration.

**Production Example:**
```bash
export CORS_ALLOWED_ORIGINS=https://example.com,https://blog.example.com
export CORS_ALLOW_CREDENTIALS=true
```

**Development Example:**
```bash
export CORS_ALLOWED_ORIGINS=*  # Allow all origins (default)
```

### Rate Limiting Configuration (Optional)

Rate limiting is enabled by default on all API endpoints to prevent spam and abuse:

| Variable | Description | Default |
|----------|-------------|---------|
| `RATE_LIMIT_GET` | Maximum GET requests per minute per IP address | `100` |
| `RATE_LIMIT_POST` | Maximum POST/PUT/DELETE requests per minute per IP address | `5` |

**Features:**
- IP-based rate limiting (supports X-Forwarded-For and X-Real-IP headers)
- Token bucket algorithm for smooth rate limiting
- Returns HTTP 429 (Too Many Requests) when limit exceeded
- Rate limit headers: `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `Retry-After`
- Automatic cleanup of old visitor data

**Note:** Rate limiting is only applied to `/api/*` routes. Admin panel routes (`/admin/*`) are not rate limited.

**Production Example:**
```bash
export RATE_LIMIT_GET=200      # Allow 200 GET requests per minute
export RATE_LIMIT_POST=10      # Allow 10 POST/PUT/DELETE requests per minute
```

**Development Example:**
```bash
# Use defaults (100 GET/min, 5 POST/min)
# Or customize as needed
export RATE_LIMIT_GET=1000     # Higher limits for development
export RATE_LIMIT_POST=50
```

### Logging & Error Handling

Kotomi uses structured JSON logging for all HTTP requests and responses:

**Features:**
- **Structured JSON logs**: All logs are output as JSON for easy parsing and analysis
- **Request tracking**: Every request receives a unique `X-Request-ID` header for tracing
- **Automatic logging**: All HTTP requests and responses are logged with:
  - Timestamp (UTC)
  - Log level (INFO, WARN, ERROR)
  - Request ID
  - HTTP method and path
  - Status code
  - Duration
  - Remote IP address
  - User agent
- **Error responses**: API errors return consistent JSON responses with error codes and request IDs
- **Privacy-focused**: Query parameters are stripped from logs to prevent logging sensitive data

**Log Format Example:**
```json
{
  "timestamp": "2026-02-03T12:00:00Z",
  "level": "INFO",
  "message": "HTTP Request",
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "method": "GET",
  "path": "/api/v1/site/my-site/page/home/comments",
  "status_code": 200,
  "duration": "15.234ms",
  "remote_addr": "192.168.1.1",
  "user_agent": "Mozilla/5.0..."
}
```

**Error Response Format:**
```json
{
  "code": "BAD_REQUEST",
  "message": "Invalid JSON format",
  "details": "unexpected end of JSON input",
  "request_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Request Tracking:**
- Each request automatically receives a unique request ID
- Request IDs are returned in the `X-Request-ID` response header
- Request IDs are included in all log entries and error responses
- External request IDs (e.g., from load balancers) are preserved if provided

### Admin Panel Configuration (Optional)

To enable the admin panel with Auth0 authentication, set these environment variables:

| Variable | Description | Required |
|----------|-------------|----------|
| `AUTH0_DOMAIN` | Your Auth0 tenant domain (e.g., `your-tenant.auth0.com`) | Yes |
| `AUTH0_CLIENT_ID` | Auth0 application client ID | Yes |
| `AUTH0_CLIENT_SECRET` | Auth0 application client secret | Yes |
| `AUTH0_CALLBACK_URL` | Callback URL for Auth0 | No (default: `http://localhost:8080/callback`) |
| `SESSION_SECRET` | Secret key for encrypting session cookies | No (auto-generated in dev) |

**Setting up Auth0:**

1. Create a free account at [auth0.com](https://auth0.com)
2. Create a new "Regular Web Application"
3. Configure the following settings in your Auth0 application:
   - **Allowed Callback URLs**: `http://localhost:8080/callback` (or your custom URL)
   - **Allowed Logout URLs**: `http://localhost:8080/`
   - **Allowed Web Origins**: `http://localhost:8080`
4. Copy your Domain, Client ID, and Client Secret from the application settings
5. Set the environment variables before starting Kotomi

Example with Auth0:
```bash
export AUTH0_DOMAIN=your-tenant.auth0.com
export AUTH0_CLIENT_ID=your_client_id
export AUTH0_CLIENT_SECRET=your_client_secret
export AUTH0_CALLBACK_URL=http://localhost:8080/callback
export SESSION_SECRET=your-random-secret-key-minimum-32-chars
PORT=3000 DB_PATH=/data/comments.db go run cmd/main.go
```

**Admin Panel Features:**
- Manage multiple sites
- Track pages within each site
- Moderate comments (approve, reject, delete)
- Configure AI moderation per site
- Real-time updates with HTMX
- User authentication via Auth0

### AI Moderation Configuration (Optional)

Kotomi supports automatic content moderation using OpenAI GPT models or a built-in rule-based moderator.

| Variable | Description | Default |
|----------|-------------|---------|
| `OPENAI_API_KEY` | OpenAI API key for AI-powered moderation | None (uses mock moderator if not set) |

**Features:**
- Automatically analyze comments for spam, offensive language, aggressive tone, and off-topic content
- Configurable confidence thresholds per site
- Three-tier decision system:
  - **Auto-Approve**: Comments with low confidence scores (< 0.30 by default)
  - **Manual Review**: Comments with medium confidence scores (0.30 - 0.85)
  - **Auto-Reject**: Comments with high confidence scores (> 0.85 by default)
- Admin UI for configuration at `/admin/sites/{siteId}/moderation`

**Setting up OpenAI:**

1. Create an account at [platform.openai.com](https://platform.openai.com)
2. Generate an API key from your account settings
3. Set the environment variable:
   ```bash
   export OPENAI_API_KEY=sk-your-api-key-here
   ```

**Cost Estimate:** ~$0.75-$1.00 per 1000 comments analyzed with GPT-3.5-turbo

**Without OpenAI:** If no API key is provided, Kotomi uses a rule-based mock moderator that checks for common spam patterns, offensive words, and aggressive language.

Example with AI moderation:
```bash
export OPENAI_API_KEY=sk-your-api-key-here
export AUTH0_DOMAIN=your-tenant.auth0.com
export AUTH0_CLIENT_ID=your_client_id
export AUTH0_CLIENT_SECRET=your_client_secret
go run cmd/main.go
```

### Export/Import Configuration

Kotomi includes built-in export and import functionality for data portability and backup.

**Features:**
- Export site data to JSON or CSV formats
- Import previously exported data
- Duplicate handling (skip or update existing records)
- Data validation during import
- Transaction-based import (all or nothing)

**Export Formats:**

1. **JSON (Complete Export)**:
   - Exports all comments, reactions, pages, and metadata
   - Preserves relationships between data
   - Can be re-imported to the same or different Kotomi instance
   - Recommended for backups and data portability

2. **CSV (Partial Export)**:
   - Comments CSV: For analysis in spreadsheet applications
   - Reactions CSV: Separate file for reaction data
   - Suitable for reporting and data analysis
   - Cannot be fully re-imported (metadata lost)

**Using Export/Import:**

1. **Export via Admin Panel**:
   - Navigate to your site in the admin panel
   - Click "Export Data"
   - Choose your format (JSON, CSV Comments, or CSV Reactions)
   - Download the file

2. **Import via Admin Panel**:
   - Navigate to your site in the admin panel
   - Click "Import Data"
   - Upload your JSON or CSV file
   - Choose duplicate handling strategy:
     - **Skip**: Skip existing records (recommended)
     - **Update**: Update existing records with new data
   - Review import results

**Example Export Filename**: `kotomi_export_my_site_20260203_120000.json`

**Important Notes:**
- Always export before importing to prevent data loss
- Import files must match the target site ID
- Large imports may take a few seconds
- Import is transactional - either all data imports or none

### Email Notifications Configuration

Kotomi can send email notifications to site owners and users for comment events.

**Features:**
- Notify site owners when new comments are posted
- Notify users when someone replies to their comment
- Notify users when their comment is approved or rejected
- Support for multiple email providers (SMTP, SendGrid)
- Background queue processing with retry logic
- HTML email templates with unsubscribe links
- Per-site notification configuration via admin panel

**Configuration:**

Email notifications are configured per-site through the admin panel at `/admin/sites/{siteId}/notifications`.

**Supported Email Providers:**

1. **SMTP (Generic):**
   - Works with any SMTP-compatible email service
   - Supports TLS, STARTTLS, and plain connections
   - Common providers: Gmail, Office 365, AWS SES, Mailgun

2. **SendGrid:**
   - API-based email delivery
   - No SMTP configuration needed
   - Requires SendGrid API key

**SMTP Configuration Examples:**

**Gmail (with App Password):**
```
Provider: SMTP
Host: smtp.gmail.com
Port: 587
Encryption: STARTTLS
Username: your-email@gmail.com
Password: your-app-password
```

**Office 365:**
```
Provider: SMTP
Host: smtp.office365.com
Port: 587
Encryption: STARTTLS
Username: your-email@office365.com
Password: your-password
```

**AWS SES:**
```
Provider: SMTP
Host: email-smtp.{region}.amazonaws.com
Port: 587
Encryption: STARTTLS
Username: your-smtp-username
Password: your-smtp-password
```

**SendGrid:**
```
Provider: SendGrid
API Key: your-sendgrid-api-key
From Email: noreply@yourdomain.com
From Name: Your Site Name
```

**Notification Types:**

- **New Comments**: Sent to site owner when a comment is posted
- **Comment Replies**: Sent to the original commenter when someone replies (requires user email)
- **Moderation Updates**: Sent to commenter when their comment is approved or rejected

**Important Notes:**
- Users must have email addresses in their JWT tokens to receive notifications
- Notification emails are queued and sent in the background
- Failed sends are retried up to 3 times
- Old processed notifications are automatically cleaned up after 7 days
- Test email functionality available in admin panel

**Gmail Users:**
If using Gmail, you must create an App Password:
1. Enable 2-factor authentication on your Google account
2. Go to Google Account → Security → 2-Step Verification → App passwords
3. Generate a new app password for "Mail"
4. Use the generated password in Kotomi (not your regular Gmail password)

### Docker Configuration

When running with Docker, use environment variables and volumes:

```bash
docker run -p 8080:8080 \
  -e PORT=8080 \
  -e DB_PATH=/app/data/kotomi.db \
  -e AUTH0_DOMAIN=your-tenant.auth0.com \
  -e AUTH0_CLIENT_ID=your_client_id \
  -e AUTH0_CLIENT_SECRET=your_client_secret \
  -e AUTH0_CALLBACK_URL=http://localhost:8080/callback \
  -e SESSION_SECRET=your-random-secret \
  -v kotomi-data:/app/data \
  kotomi
```

## Project Structure

```
kotomi/
├── cmd/                # Application entry point
│   └── main.go
├── pkg/                # Public packages
│   ├── admin/          # Admin handlers
│   ├── auth/           # Auth0 authentication
│   ├── comments/       # Comments storage
│   └── models/         # Data models (users, sites, pages)
├── templates/          # HTML templates
│   ├── admin/          # Admin panel templates
│   └── base.html
├── static/             # Static assets (CSS)
├── docs/               # Documentation
│   └── adr/            # Architecture Decision Records
├── internal_docs/      # Internal documentation
├── .github/            # GitHub workflows and configurations
├── Dockerfile          # Docker configuration
├── go.mod              # Go module definition
├── go.sum              # Go module checksums
├── VERSION             # Current version
└── README.md           # This file
```

## Roadmap

### v0.1.0 - Current Focus

- ✅ Basic server setup with health check
- ✅ SQLite persistent storage
- ✅ REST API for comments
- ✅ Comprehensive test coverage (>90%)
- ✅ Admin panel with Auth0 authentication
- ✅ Multi-site and page management
- ✅ Comment moderation (approve, reject, delete)
- ✅ HTMX-based UI
- ✅ CORS configuration
- ✅ Rate limiting
- ✅ Reactions system (emoji reactions to comments)
- ✅ AI moderation (OpenAI GPT integration)

### Future Versions

- **v0.2.0** - Frontend widget and JavaScript SDK
- **v0.3.0** - Additional storage backends
- **v0.4.0** - Authentication and user management
- **v0.5.0** - Advanced moderation tools
- **v1.0.0** - Production-ready release

## Development

### Testing

Kotomi has comprehensive test coverage (>90%). To run tests:

```bash
# Run all tests
go test ./...

# Run with coverage
go test ./... -cover

# Run with verbose output
go test ./... -v

# Run specific package tests
go test ./pkg/comments/...
```

### Code Style

- Follow standard Go formatting (`gofmt`)
- Write meaningful test names: `TestFunctionName_Scenario_ExpectedResult`
- Include error handling for all operations
- Use prepared statements for database queries

## Contributing

We welcome contributions! Here's how you can help:

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Commit** your changes (`git commit -m 'Add some amazing feature'`)
4. **Push** to the branch (`git push origin feature/amazing-feature`)
5. **Open** a Pull Request

Please follow Go best practices and include tests for new features.

## Architecture

For information about architectural decisions, see the [Architecture Decision Records (ADR)](docs/adr/) directory. Key decisions include:

- [ADR 001: User Authentication for Comments and Reactions](docs/adr/001-user-authentication-for-comments-and-reactions.md) - External JWT-based authentication is implemented; built-in Kotomi auth pending (Status: Partially Implemented)

## License

License to be determined.

## Links

- **Project Website:** Coming soon
- **Issue Tracker:** [GitHub Issues](https://github.com/saasuke-labs/kotomi/issues)
- **Discussions:** [GitHub Discussions](https://github.com/saasuke-labs/kotomi/discussions)

## Philosophy

Kotomi aims to bridge the gap between static sites and dynamic content. We believe that static sites shouldn't mean static experiences. By providing a lightweight, privacy-focused commenting system, we empower developers to add interactive features to their sites without compromising on performance or user privacy.
