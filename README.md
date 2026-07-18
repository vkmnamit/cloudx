# CloudX

CloudX is a self-hosted Platform as a Service (PaaS) designed to simplify application deployment and infrastructure management. It provides a unified platform for deploying, managing, and monitoring containerized applications while maintaining full control over the underlying infrastructure.

The project is being built as a production-grade platform inspired by modern cloud deployment systems such as Railway, Render, and Fly.io.

## Features

### Current

- User Authentication
- JWT-based Authorization
- PostgreSQL Integration
- Dockerized Backend
- REST API
- Layered Architecture (Handler → Service → Repository)
- Environment-based Configuration

### Planned

- GitHub OAuth
- Project Management
- Application Deployments
- Docker Engine Integration
- Reverse Proxy
- SSL Certificate Management
- Build Pipeline
- Deployment Logs
- Metrics & Monitoring
- Resource Management
- Multi-Node Support

---

## Technology Stack

### Backend

- Go
- Gin
- PostgreSQL
- pgx
- Docker
- Docker Compose

### Architecture

- Clean Architecture
- Repository Pattern
- Service Layer
- RESTful APIs

---

## Project Structure

```
cloudx/

├── backend/
│   ├── cmd/
│   │   └── server/
│   ├── configs/
│   ├── internal/
│   │   ├── auth/
│   │   ├── middleware/
│   │   └── router/
│   ├── migrations/
│   ├── pkg/
│   ├── Dockerfile
│   ├── go.mod
│   └── go.sum
│
├── docker-compose.yml
└── README.md
```

---

## Getting Started

### Prerequisites

- Go 1.26+
- Docker
- Docker Compose

### Clone the Repository

```bash
git clone https://github.com/vkmnamit/cloudx.git

cd cloudx
```

### Run with Docker

```bash
docker compose up --build
```

The backend will be available at:

```
http://localhost:8080
```

---

## Local Development

```bash
cd backend

go mod tidy

go run ./cmd/server
```

---

## Configuration

Create a `.env` file inside the `backend` directory.

Example:

```env
APP_NAME=CloudX
PORT=8080
ENV=development

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=cloudx
DB_SSLMODE=disable
```

When running with Docker:

```env
DB_HOST=postgres
```

---

## API

| Method | Endpoint | Description |
|----------|----------|-------------|
| POST | `/api/signup` | Register a new user |
| POST | `/api/login` | Authenticate a user |

---

## Roadmap

- Authentication
- Authorization
- Organizations
- Projects
- Deployments
- GitHub Integration
- Docker Runtime
- Reverse Proxy
- SSL Automation
- Monitoring
- CI/CD
- Multi-node Infrastructure

---

## License

This project is licensed under the MIT License.
