# Go Reverse Proxy

A lightweight, easy-to-use reverse proxy server with a beautiful web-based management panel.

![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)
![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker)
![License](https://img.shields.io/badge/License-MIT-green?style=flat)

**[中文文档](README_CN.md)** | **Official Website: [www.nsmao.com](https://www.nsmao.com)**

## Features

- **Web Management Panel** - Beautiful Zen-iOS Hybrid design with glassmorphism effects
- **Dynamic Configuration** - Add/edit/delete proxy rules without restart
- **Authentication** - Secure admin panel with login system
- **JSON Storage** - No database required, configuration stored in JSON files
- **Docker Ready** - Easy deployment with Docker and docker-compose
- **Multi-Platform** - Supports linux/amd64 and linux/arm64

## Quick Start

### Using Docker (Recommended)

```bash
# Pull from GitHub Container Registry
docker pull ghcr.io/nsmao-com/go_proxy_every:latest

# Run container
docker run -d \
  --name go-proxy-every \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  ghcr.io/nsmao-com/go_proxy_every:latest
```

### Using Docker Compose

```bash
# Clone the repository
git clone https://github.com/nsmao-com/go_proxy_every.git
cd go_proxy_every

# Start with docker-compose
docker-compose up -d
```

### Build from Source

```bash
# Clone the repository
git clone https://github.com/nsmao-com/go_proxy_every.git
cd go_proxy_every

# Build
go build -o proxy .

# Run
./proxy
```

## Usage

### Access the Service

- **Homepage**: http://localhost:8080
- **Admin Panel**: http://localhost:8080/admin/

### Default Credentials

| Field | Value |
|-------|-------|
| Username | `admin` |
| Password | `admin123` |

> **Important**: Please change the default password after first login!

### Adding a Proxy Rule

1. Login to the admin panel
2. Click "Add Rule" button
3. Fill in the form:
   - **Name**: A friendly name for the rule
   - **Path**: The local path prefix (e.g., `nsmao`)
   - **Target**: The target URL to proxy (e.g., `https://www.nsmao.com`)
4. Enable the rule and save

### Example

If you add a rule:
- Path: `nsmao`
- Target: `https://www.nsmao.com`

Then accessing `http://localhost:8080/nsmao` will proxy to `https://www.nsmao.com`

## Configuration

Configuration is stored in `data/rules.json`:

```json
{
  "rules": [
    {
      "id": "uuid",
      "name": "Example Site",
      "path": "example",
      "target": "https://example.com",
      "enabled": true,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

Authentication config is stored in `data/auth.json`:

```json
{
  "username": "admin",
  "password": "admin123"
}
```

## API Reference

| Endpoint | Method | Description | Auth |
|----------|--------|-------------|------|
| `/api/login` | POST | Login | No |
| `/api/logout` | POST | Logout | No |
| `/api/check-auth` | GET | Check authentication | No |
| `/api/rules` | GET | List all rules | Yes |
| `/api/rules` | POST | Create a rule | Yes |
| `/api/rules` | PUT | Update a rule | Yes |
| `/api/rules` | DELETE | Delete a rule | Yes |
| `/api/rules/toggle` | POST | Toggle rule status | Yes |
| `/api/change-password` | POST | Change password | Yes |

## Project Structure

```
go-reverse-proxy/
├── main.go              # Entry point
├── go.mod               # Go module
├── Dockerfile           # Docker build file
├── docker-compose.yml   # Docker compose config
├── auth/
│   └── auth.go          # Authentication logic
├── config/
│   └── config.go        # Configuration management
├── handlers/
│   └── api.go           # API handlers
├── proxy/
│   └── reverse_proxy.go # Reverse proxy core
├── static/
│   └── index.html       # Web UI
├── data/
│   ├── rules.json       # Proxy rules
│   └── auth.json        # Auth config
└── .github/
    └── workflows/
        └── docker.yml   # GitHub Actions
```

## Docker Hub / GHCR

After pushing to GitHub, the Docker image will be automatically built and pushed to GitHub Container Registry.

```bash
# Pull the image
docker pull ghcr.io/nsmao-com/go_proxy_every:latest

# Or with specific version
docker pull ghcr.io/nsmao-com/go_proxy_every:v1.0.0
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `TZ` | `Asia/Shanghai` | Timezone |

## Security Notes

1. Change the default password immediately after deployment
2. Use HTTPS in production (deploy behind nginx/caddy with SSL)
3. Consider using firewall rules to restrict access to the admin panel

## License

MIT License

## Contributing

Pull requests are welcome. For major changes, please open an issue first.
