# Kotomi

Give your pages a voice

[![Go Version](https://img.shields.io/badge/Go-1.24-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Version](https://img.shields.io/badge/version-0.0.1-blue.svg)](https://github.com/saasuke-labs/kotomi/releases)

> ⚠️ **Early Development**: Kotomi is currently in early development (v0.0.1) and is not recommended for production use yet.

## Features

- 💬 **Comments System** - Enable discussions on your static pages
- 🔐 **Admin Panel** - Web-based dashboard with Auth0 authentication
- 🏢 **Multi-Site Management** - Manage multiple sites from a single instance
- 📄 **Page Tracking** - Organize comments by pages within sites
- 🛡️ **Comment Moderation** - Approve, reject, or delete comments with ease
- ⚡ **HTMX Interface** - Smooth, no-reload UI updates
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

### Admin Panel Routes

- `/admin` - Redirects to dashboard
- `/admin/dashboard` - Overview of sites and pending comments
- `/admin/sites` - List all sites
- `/admin/sites/{siteId}` - View site details and pages
- `/admin/sites/{siteId}/comments` - Moderate comments for a site
- `/login` - Auth0 login
- `/logout` - Logout and clear session

## API Documentation

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

**Endpoint:** `GET /api/site/{siteId}/page/{pageId}/comments`

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

**Endpoint:** `POST /api/site/{siteId}/page/{pageId}/comments`

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
- Real-time updates with HTMX
- User authentication via Auth0

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
- 🚧 CORS configuration
- 🚧 Rate limiting

### Future Versions

- **v0.2.0** - Reactions and voting system
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

## License

License to be determined.

## Links

- **Project Website:** Coming soon
- **Issue Tracker:** [GitHub Issues](https://github.com/saasuke-labs/kotomi/issues)
- **Discussions:** [GitHub Discussions](https://github.com/saasuke-labs/kotomi/discussions)

## Philosophy

Kotomi aims to bridge the gap between static sites and dynamic content. We believe that static sites shouldn't mean static experiences. By providing a lightweight, privacy-focused commenting system, we empower developers to add interactive features to their sites without compromising on performance or user privacy.
