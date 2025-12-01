English | [中文](README.zh-CN.md)

# Bingo - Production-ready Go Microservice Scaffold

> A production-ready Go/Golang microservice scaffold framework for rapid development, letting developers focus on business logic.

[![Go Version](https://img.shields.io/badge/Go-1.23%2B-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

## Documentation

**Official Documentation**: [bingoctl.dev](https://bingoctl.dev/en/)

**Quick Links**:
- [Getting Started](https://bingoctl.dev/en/guide/getting-started) - Get up and running in 10 minutes
- [What is Bingo](https://bingoctl.dev/en/guide/what-is-bingo) - Learn about core features
- [Architecture](https://bingoctl.dev/en/essentials/architecture) - Microservice architecture design
- [Using bingo CLI](https://bingoctl.dev/en/guide/using-bingo) - CLI tool guide
- [中文文档](https://bingoctl.dev/) - Chinese documentation

## Overview

**Bingo** is a **production-ready Go/Golang microservice scaffold framework** that provides:
- Clean architecture design (Controller → Biz → Store three-layer architecture)
- Pre-integrated core components (Gin, GORM, Redis, Asynq, Casbin)
- Engineering capabilities (code generation, hot reload, Docker support)
- Production-grade features (logging, monitoring, tracing)
- Best practices and comprehensive documentation

**Use Cases**: Backend systems, microservice projects, RESTful APIs, gRPC services

**Related Project**: [bingo CLI](https://github.com/bingo-project/bingoctl) - Bingo project scaffold tool

## Core Features

### Architecture
- **Microservice Architecture**: Independent service deployment with horizontal scaling
- **Layered Design**: Clean Controller → Biz → Store three-layer architecture
- **Generic Data Layer**: Generic-based Store[T] design to reduce boilerplate
- **Dependency Injection**: Interface-based programming for easy testing and extension
- **Service Discovery**: Support for gRPC inter-service communication

### Tech Components
- **Web Framework**: Gin - High-performance HTTP framework
- **ORM**: GORM - Multi-database support
- **Cache**: Redis - Distributed caching
- **Task Queue**: Asynq - Reliable async task processing
- **Access Control**: Casbin - Flexible RBAC permission engine
- **Logging**: Zap - Structured high-performance logging
- **API Docs**: Swagger - Auto-generated API documentation

### Engineering
- **Hot Reload**: Air support for development hot reload
- **Code Generation**: Auto-generate CRUD code and API docs
- **Docker Support**: One-click containerized deployment
- **Monitoring**: Prometheus + pprof performance monitoring

## Tech Stack

- **Go**: 1.23.1+
- **Web Framework**: Gin v1.10.0
- **ORM**: GORM v1.25.10
- **Database**: MySQL 5.7+ / PostgreSQL
- **Cache**: Redis 6.0+
- **gRPC**: google.golang.org/grpc v1.64.0
- **Task Queue**: Asynq v0.24.1

## Quick Start

### Option 1: Create New Project with bingo CLI (Recommended)

Use the [bingo CLI](https://github.com/bingo-project/bingoctl) tool to quickly create a project:

```bash
# Install bingo CLI
go install github.com/bingo-project/bingoctl/cmd/bingo@latest

# Create new project (apiserver only)
bingo create github.com/myorg/myapp

# Or create project with all services
bingo create github.com/myorg/myapp --all

# Enter project directory
cd myapp

# Start dependency services
docker-compose -f deployments/docker/docker-compose.yaml up -d

# Generate your first module (e.g., user module)
bingo make crud user

# Run service
make build
./myapp-apiserver
```

**Common options for project creation:**

```bash
# Create with specific services
bingo create myapp --services apiserver,admserver

# Add additional services
bingo create myapp --add-service scheduler

# Exclude certain services
bingo create myapp --no-service bot

# Control git initialization
bingo create myapp --init-git=false

# Use specific template version
bingo create myapp -r v1.2.3
```

See [Using bingo CLI](https://bingoctl.dev/en/guide/using-bingo) for detailed guide.

### Option 2: Clone Bingo Repository

If you want to develop based on Bingo source code:

#### 1. Clone Project

```bash
git clone <repository-url>
cd bingo
```

#### 2. Configure Environment

```bash
# Copy config file
cp configs/bingo-apiserver.example.yaml bingo-apiserver.yaml

# Edit config for your environment
vim bingo-apiserver.yaml
```

#### 3. Start Dependencies

```bash
# Use Docker Compose to start MySQL and Redis
docker-compose -f deployments/docker/docker-compose.yaml up -d mysql redis
```

#### 4. Database Migration

```bash
# Build project
make build

# Copy config file
cp configs/{app}ctl.example.yaml {app}ctl.yaml

# Build your app ctl
make build BINS="{app}ctl"

# Run database migration
./_output/platforms/{os}/{arch}/{app}ctl migrate up
```

#### 5. Start Service

```bash
# Option 1: Run directly
make build
bingo-apiserver -c bingo-apiserver.yaml

# Option 2: Development mode (hot reload)
cp .air.example.toml .air.toml
air
```

#### 6. Verify Service

```bash
# Check service status
curl http://localhost:8080/health

# Access Swagger docs
open http://localhost:8080/swagger/index.html
```

## Documentation Guide

### Recommended Learning Path

**Beginners**: [What is Bingo](https://bingoctl.dev/en/guide/what-is-bingo) → [Getting Started](https://bingoctl.dev/en/guide/getting-started) → [Project Structure](https://bingoctl.dev/en/guide/project-structure) → [First Feature](https://bingoctl.dev/en/guide/first-feature)

**Deep Dive**: [Architecture](https://bingoctl.dev/en/essentials/architecture) → [Layered Design](https://bingoctl.dev/en/essentials/layered-design) → [Store Package](https://bingoctl.dev/en/essentials/store) → [Development Standards](https://bingoctl.dev/en/development/standards)

**Production**: [Docker Deployment](https://bingoctl.dev/en/deployment/docker) → [Microservice Decomposition](https://bingoctl.dev/en/advanced/microservices)

## Common Commands

```bash
# Development
make build          # Build all services
make run            # Run service (dev mode)
make test           # Run unit tests
make cover          # Test coverage report

# Code Quality
make lint           # Code linting
make format         # Code formatting

# Code Generation
make swagger        # Generate Swagger docs
make protoc         # Compile Protocol Buffers

# Deployment
make image          # Build Docker image

# Cleanup
make clean          # Clean build artifacts
```

## Project Structure

```
bingo/
├── cmd/                    # Executable entry points
│   ├── bingo-apiserver/    # API service
│   ├── bingo-admserver/    # Admin service
│   ├── bingo-scheduler/    # Scheduler service
│   ├── bingo-bot/          # Bot service
│   └── bingoctl/           # CLI tool
├── internal/               # Internal application code
│   ├── apiserver/          # API service implementation
│   ├── admserver/          # Admin service implementation
│   └── pkg/                # Internal shared packages
├── pkg/                    # Public packages
├── docs/                   # Documentation
├── configs/                # Configuration files
├── deployments/            # Deployment configs
└── scripts/                # Scripts
```

See [Project Structure](https://bingoctl.dev/en/guide/project-structure) for details.

## Contributing

Issues and Pull Requests are welcome!

### Development Workflow

1. Fork this repository
2. Create feature branch: `git checkout -b feature/amazing-feature`
3. Commit changes: `git commit -m 'feat: add amazing feature'`
4. Push branch: `git push origin feature/amazing-feature`
5. Submit Pull Request

### Code Review

PRs must pass:
- Code linting (golangci-lint)
- Unit tests
- Review by at least one Maintainer

## License

This project is licensed under the [Apache License 2.0](LICENSE).

## Contact

For questions or suggestions:
- Submit an Issue
- Email project maintainers

---

**Start using Bingo - Focus on your business logic, let the scaffold handle the rest!**
