English | [中文](README.zh-CN.md)

# Bingo - Production-ready Go Microservice Scaffold

> A production-ready Go/Golang microservice scaffold framework for rapid development, letting developers focus on business logic.

[![Go Version](https://img.shields.io/badge/Go-1.23%2B-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

## Documentation

**Official Documentation**: [bingoctl.dev](https://bingoctl.dev/en/)

- [Getting Started](https://bingoctl.dev/en/guide/getting-started) - Get up and running in 10 minutes
- [What is Bingo](https://bingoctl.dev/en/guide/what-is-bingo) - Learn about core features
- [Architecture](https://bingoctl.dev/en/essentials/architecture) - Microservice architecture design
- [Using Bingo CLI](https://bingoctl.dev/en/guide/using-bingo) - CLI tool guide
- [中文文档](https://bingoctl.dev/) - Chinese documentation

## Overview

**Bingo** is a **production-ready Go/Golang microservice scaffold framework** that provides:
- Clean architecture design (Controller → Biz → Store three-layer architecture)
- Pre-integrated core components (Gin, GORM, Redis, Asynq, Casbin)
- Engineering capabilities (code generation, hot reload, Docker support)
- Production-grade features (logging, monitoring, tracing)
- Best practices and comprehensive documentation

**Use Cases**: Backend systems, microservice projects, RESTful APIs, gRPC services

**Related Project**: [Bingo CLI](https://github.com/bingo-project/bingoctl) - Bingo project scaffold tool

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

```bash
# Install Bingo CLI
go install github.com/bingo-project/bingoctl/cmd/bingo@latest

# Create new project
bingo create github.com/myorg/myapp

# Enter project directory
cd myapp

# Build and run
make build && ./_output/<os>/<arch>/myapp-apiserver
```

For detailed setup instructions, see [Getting Started](https://bingoctl.dev/en/guide/getting-started).

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
