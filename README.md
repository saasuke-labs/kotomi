# Kotomi

Give your pages a voice

[![Go Version](https://img.shields.io/badge/Go-1.24-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Version](https://img.shields.io/badge/version-0.0.1-blue.svg)](https://github.com/saasuke-labs/kotomi/releases)

> ⚠️ **Early Development**: Kotomi is currently in early development (v0.0.1) and is not recommended for production use yet.

## Features

- 💬 **Comments System** - Enable discussions on your static pages
- 👍 **Reactions** - Let users express themselves with reactions
- 🛡️ **Moderation** - Tools to manage and moderate content
- 🪶 **Lightweight** - Built with Go for minimal resource usage
- 🔌 **Easy Integration** - Simple REST API for seamless integration
- 🔒 **Privacy-Focused** - Designed with user privacy in mind

## Architecture

Kotomi is built with simplicity and performance in mind:

- **Go 1.24** - Modern, fast, and efficient
- **SQLite Storage** - Persistent, reliable database with zero configuration
- **REST API** - Standard HTTP endpoints for easy integration
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

### Running Tests

Run all tests:
```bash
go test ./...
```

Run tests with coverage:
```bash
go test ./... -cover
```

Run tests for a specific package:
```bash
go test ./pkg/comments/...
```

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

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `DB_PATH` | Path to SQLite database file | `./kotomi.db` |

Example:
```bash
PORT=3000 DB_PATH=/data/comments.db go run cmd/main.go
```

### Docker Configuration

When running with Docker, use environment variables and volumes:

```bash
docker run -p 8080:8080 \
  -e PORT=8080 \
  -e DB_PATH=/app/data/kotomi.db \
  -v kotomi-data:/app/data \
  kotomi
```

## Project Structure

```
kotomi/
├── cmd/                # Application entry point
│   └── main.go
├── pkg/                # Public packages
│   └── comments/
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
- 🚧 CORS configuration
- 🚧 Basic moderation features

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
