# go-lite-auth

[![Go CI](https://github.com/joshleeeeee/go-lite-auth/actions/workflows/go.yml/badge.svg)](https://github.com/joshleeeeee/go-lite-auth/actions/workflows/go.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/joshleeeeee/go-lite-auth)](https://golang.org/)

A lightweight Single Sign-On (SSO) system built with Go.

[中文文档 (Chinese)](README_zh.md) | [Contributing](CONTRIBUTING.md) | [Changelog](CHANGELOG.md)

## Tech Stack

- **Web Framework**: [Gin](https://github.com/gin-gonic/gin)
- **Database**: [SQLite](https://sqlite.org/) (Default) / [MySQL 8](https://www.mysql.com/) / [PostgreSQL](https://www.postgresql.org/) + [GORM](https://gorm.io/)
- **Cache**: [Redis](https://redis.io/)
- **Authentication**: JWT (Access Token + Refresh Token)
- **Configuration**: [Viper](https://github.com/spf13/viper)

## Project Structure

```text
lite-auth/
├── cmd/
│   └── server/
│       └── main.go           # Entry point
├── config/
│   └── config.yaml           # Configuration file
├── internal/
│   ├── config/               # Configuration loading
│   ├── database/             # Database initialization (SQLite, MySQL, Postgres) & Redis
│   ├── handler/              # HTTP handlers (Controllers)
│   ├── middleware/           # Middleware (JWT, CORS, etc.)
│   ├── model/                # Data models
│   ├── repository/           # Data access layer (DAO)
│   ├── router/               # Route definitions
│   └── service/              # Business logic layer
├── pkg/
│   └── jwt/                  # JWT utilities
├── test/
│   └── api/
│       └── auth.http         # API test scripts
├── go.mod
└── README.md
```

## Quick Start

### 1. Prerequisites

Ensure you have the following installed:
- Go 1.21+
- MySQL 8.0+
- Redis 6.0+

### 2. Create Database

```sql
CREATE DATABASE lite_auth CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### 3. Configuration

Edit `config/config.yaml` with your MySQL and Redis connection details.

### 4. Install Dependencies

```bash
go mod tidy
```

### 5. Run the Server

```bash
go run cmd/server/main.go
```

The server will start at http://localhost:8080.

## API Endpoints

### Authentication

| Method | Path | Description | Auth Required |
|--------|------|-------------|---------------|
| POST | `/api/auth/register` | User Registration | ❌ |
| POST | `/api/auth/login` | User Login | ❌ |
| POST | `/api/auth/logout` | User Logout | ✅ |
| POST | `/api/auth/refresh` | Refresh Token | ❌ |
| GET | `/api/auth/validate` | Validate Token | ❌ |

### User Profile

| Method | Path | Description | Auth Required |
|--------|------|-------------|---------------|
| GET | `/api/user/info` | Get current user info | ✅ |

## Sample Requests

### Register
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","email":"test@example.com","password":"123456"}'
```

### Login
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"123456"}'
```

### Get User Info
```bash
curl http://localhost:8080/api/user/info \
  -H "Authorization: Bearer <access_token>"
```

## Redis Key Design

| Prefix | Purpose | TTL |
|------|------|----------|
| `session:` | User session data | 24 hours |
| `blacklist:` | Revoked JWT tokens | Remaining JWT TTL |
| `ticket:` | SSO Tickets | 60 seconds |
| `login_fail:` | Login failure counter | 5 minutes |

## Roadmap

- [ ] SSO Ticket mechanism (CAS-style)
- [ ] OAuth 2.0 Authorization Code Flow
- [ ] Frontend login page
- [ ] Client application management
- [ ] Admin dashboard

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT
