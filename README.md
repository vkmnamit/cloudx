

#  CloudX

> **CloudX** is a self-hosted Platform as a Service (PaaS) that automatically builds, deploys, and manages applications directly from GitHub repositories using Docker.

CloudX is inspired by platforms like **Railway**, **Render**, **Fly.io**, and **Heroku**, while giving developers complete ownership of their infrastructure.

Instead of manually provisioning servers, writing deployment scripts, or configuring reverse proxies, CloudX automates the deployment pipeline from repository to running application.

---

# Features

## Authentication & Security

* JWT Authentication
* Refresh Tokens
* Password Hashing (bcrypt)
* Protected APIs
* Role-Based Authorization (RBAC)
* Authentication Middleware

---

## Organizations

* Create Organizations
* Invite Members
* Role Management
* Multi-Organization Support

---

## Projects

* Create Projects
* Connect GitHub Repository
* Branch Selection
* Project Configuration

---

## Services

Each project can contain multiple deployable services.

Supports:

* Environment Variables
* Build Context
* Custom Build Commands
* Start Commands
* Automatic Port Detection
* Language Detection

---

## GitHub Integration

* GitHub OAuth
* Repository Selection
* Branch Selection
* Commit Tracking
* Clone Repository

---

## Intelligent Project Analyzer

CloudX automatically detects:

* Programming Language
* Framework
* Project Root
* Build Context
* Dockerfile
* Start Command
* Exposed Port

Supports:

* React
* Vite
* Next.js
* Go
* Node.js
* Python

Automatically detects nested repositories like:

```
repo/
    frontend/
```

without any manual configuration.

---

## Automatic Dockerfile Generation

CloudX generates optimized Dockerfiles automatically.

Supports:

* Node
* Go
* Python

If a Dockerfile already exists, CloudX uses it instead.

---

## Docker Runtime

Automatically:

* Build Docker Images
* Create Containers
* Create Networks
* Inject Environment Variables
* Publish Ports
* Start Containers
* Stop Containers
* Remove Containers

---

## Deployments

One-click deployment pipeline:

```
GitHub Repository
        ↓
Clone Repository
        ↓
Analyze Project
        ↓
Generate Dockerfile
        ↓
Build Docker Image
        ↓
Create Container
        ↓
Health Check
        ↓
Running
```

Each deployment stores:

* Commit SHA
* Commit Message
* Image Tag
* Build Status
* Logs
* Start Time
* Finish Time

---

## Health Checks

CloudX continuously validates deployments before marking them as running.

Health checks verify:

* Container Status
* HTTP Response
* Application Availability

---

## Environment Variables

Supports runtime environment variables.

Example:

```
DATABASE_URL
JWT_SECRET
API_KEY
REDIS_URL
```

---

## REST API

Production-ready REST APIs built using Gin.

---

## PostgreSQL

Stores:

* Users
* Organizations
* Projects
* Services
* Deployments
* Environment Variables

---

## Dockerized Infrastructure

Everything runs inside Docker.

Services:

* Backend
* PostgreSQL

---

# 🛠 Technology Stack

## Backend

* Go
* Gin
* PostgreSQL
* pgx
* JWT
* bcrypt

---

## Infrastructure

* Docker
* Docker Compose
* Docker Engine API

---

## Deployment

* GitHub
* Docker Build
* Docker Runtime

---

#  Architecture

```
                     Client
                        │
                        ▼
               Gin HTTP Server
                        │
                        ▼
                 Authentication
                        │
                        ▼
                  API Handlers
                        │
                        ▼
               Application Services
                        │
                        ▼
                 Repository Layer
                        │
                        ▼
                 PostgreSQL Database
```

Deployment Pipeline

```
Deploy Request
      │
      ▼
GitHub Clone
      │
      ▼
Analyzer
      │
      ▼
Dockerfile Generator
      │
      ▼
Docker Build
      │
      ▼
Runtime Manager
      │
      ▼
Health Check
      │
      ▼
Deployment Running
```

---

# 📂 Project Structure

```
cloudx/

backend/

cmd/
server/

configs/

internal/

auth/
deployment/
docker/
github/
middleware/
organization/
project/
repository/
runtime/
service/
router/
analyzer/
database/

migrations/

pkg/

Dockerfile

docker-compose.yml

README.md
```

---

# ⚙ Requirements

* Go 1.26+
* Docker 29+
* Docker Compose
* PostgreSQL 17

---

#  Installation

Clone

```bash
git clone https://github.com/vkmnamit/cloudx.git

cd cloudx
```

---

Run

```bash
docker compose up --build
```

Backend

```
http://localhost:8080
```

---

# API Testing

## Register

```http
POST /api/signup
```

Body

```json
{
  "name": "Namit",
  "email": "namit@example.com",
  "password": "password123"
}
```

Response

```json
{
  "message": "User created successfully"
}
```

---

## Login

```http
POST /api/login
```

Body

```json
{
  "email": "namit@example.com",
  "password": "password123"
}
```

Response

```json
{
  "access_token": "...",
  "refresh_token": "..."
}
```

---

## Create Organization

```
POST /api/organizations
Authorization: Bearer TOKEN
```

---

## Create Project

```
POST /api/projects
Authorization: Bearer TOKEN
```

---

## Connect GitHub Repository

```
POST /api/github/connect
```

---

## Create Service

```
POST /api/services
```

---

## Deploy Application

```
POST /api/deployments
```

Example

```json
{
    "service_id":"xxxxxxxx"
}
```

---

## Deployment Status

```
GET /api/deployments/:id
```

---

## Deployment Logs

```
GET /api/deployments/:id/logs
```

---

#  Database

Current tables

```
users

organizations

organization_members

projects

services

deployments

deployment_logs

environment_variables

refresh_tokens
```

---

#  Roadmap

* Custom Domains
* Automatic SSL (Let's Encrypt)
* Reverse Proxy (Traefik/Caddy)
* Webhooks
* Auto Deploy on Push
* Live Deployment Logs
* Metrics Dashboard
* CPU & Memory Monitoring
* Horizontal Scaling
* Multi-Node Clusters
* Rollbacks
* Secret Management
* CLI
* Kubernetes Runtime
* AI-assisted Deployment Insights

---

# Contributing

```bash
git clone

git checkout -b feature/my-feature

git commit

git push

open pull request
```

---

# 📄 License

This project is licensed under the MIT License.

---

## One suggestion

Your README has outgrown a simple document. At CloudX's current stage, it would be worth adding:


