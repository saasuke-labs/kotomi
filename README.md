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
- **In-memory storage** - Lightning-fast data access
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
docker run -p 8080:8080 kotomi
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

The comments API endpoints are currently under development and will be available in the next release:

- `GET /api/site/{siteId}/page/{pageId}/comments` - Get comments for a page (Coming soon)
- `POST /api/site/{siteId}/page/{pageId}/comments` - Post a comment (Coming soon)

## Configuration

Kotomi can be configured using environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |

Example:
```bash
PORT=3000 go run cmd/main.go
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
- ✅ In-memory comment storage
- 🚧 REST API for comments
- 🚧 CORS configuration
- 🚧 Basic moderation features

### Future Versions

- **v0.2.0** - Reactions and voting system
- **v0.3.0** - Persistent storage options
- **v0.4.0** - Authentication and user management
- **v0.5.0** - Advanced moderation tools
- **v1.0.0** - Production-ready release

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
