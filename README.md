# pastebin

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat-square&logo=go)](https://go.dev/)
[![Docker Image](https://img.shields.io/badge/Docker-ghcr.io%2Fskidoodle%2Fpastebin-blue?style=flat-square&logo=docker)](https://github.com/skidoodle/pastebin/pkgs/container/pastebin)

A minimalist, high-performance paste service written in Go.

## Deployment

### Docker Compose

```yaml
services:
  pastebin:
    image: ghcr.io/skidoodle/pastebin:latest
    container_name: pastebin
    restart: unless-stopped
    ports:
      - "3000:3000"
    volumes:
      - ./data:/app/database
```

### Manual Installation

Requires Go 1.25 or higher.

```bash
go build -o pastebin .
./pastebin -addr :3000 -db-path pastebin.db
```

## Configuration

| Flag | Description | Default |
| :--- | :--- | :--- |
| `-addr` | Socket address to bind to. | `:3000` |
| `-db-path` | Path to the BoltDB database file. | `pastebin.db` |
| `-max-size` | Maximum allowed paste size in bytes. | `10485760` |
| `-ttl` | Time to live for pastes before expiration. | `168h0m0s` |

## License

This project is licensed under the [Zero-Clause BSD](LICENSE).
