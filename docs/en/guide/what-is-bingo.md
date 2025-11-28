# What is Bingo

Bingo is a **production-grade Go backend scaffold** that provides a complete microservice architecture, core components, and best practices, helping teams quickly build scalable backend services.

## Project Positioning

An out-of-the-box Go backend scaffold based on microservice architecture design, allowing developers to focus solely on business development.

## Design Philosophy

- **Out-of-the-box**: Built-in complete tech stack and core components for quick project startup
- **Business-focused**: The scaffold handles technical details, developers focus on business logic
- **Flexible and Extensible**: Modular design, components can be freely combined or removed based on requirements
- **Production-ready**: Includes monitoring, logging, distributed tracing and other production-essential features
- **Best Practices**: Follows Go community best practices and design patterns

## Core Features

### Architecture Level
- **Microservice Architecture**: Multiple services deployed independently, supporting horizontal scaling
- **Layered Design**: Clear three-layer architecture (Controller → Biz → Store)
- **Dependency Injection**: Interface-based programming, easy to test and extend
- **Service Discovery**: Support gRPC inter-service communication

### Technology Components
- **Web Framework**: Gin - high-performance HTTP framework
- **ORM**: GORM - supporting multiple databases
- **Caching**: Redis integration - distributed caching
- **Task Queue**: Asynq - reliable asynchronous task handling
- **Permission Control**: Casbin - flexible RBAC permission engine
- **Configuration Management**: Viper - supporting multiple config formats
- **Logging**: Zap - structured high-performance logging
- **API Documentation**: Swagger - auto-generating API docs

### Engineering Capabilities
- **CLI Tool**: [bingoctl](https://github.com/bingo-project/bingoctl) - quickly create projects and generate code
- **Hot Reload**: Air support for development-time hot reload
- **Code Generation**: Auto-generating CRUD code and API documentation
- **Docker Support**: One-click containerization deployment
- **Monitoring Metrics**: Prometheus + pprof performance monitoring
- **Unit Testing**: Complete testing framework and examples

## Built-in Example Features

The scaffold includes some basic features as development references, these are **optional** and can be retained or removed based on actual needs:

- **User Authentication**: Examples of JWT, OAuth, Web3 and other authentication methods
- **Permission Management**: RBAC-based permission control examples
- **Application Management**: Multi-application and API Key management examples
- **Bot Service**: Discord/Telegram Bot integration examples
- **Scheduled Tasks**: Task scheduling based on Asynq examples

These built-in features are mainly for:
1. Demonstrating scaffold usage methods and best practices
2. Providing reusable code templates
3. Serving as a starting point for business development

> **Tip**: You can reference these examples to quickly develop your own business features, or directly delete unnecessary modules.

## Tech Stack

### Core Framework
- **Go**: 1.23.1+
- **Web Framework**: Gin v1.10.0
- **ORM**: GORM v1.25.10
- **Database**: MySQL 5.7+ / PostgreSQL (optional)
- **Cache**: Redis 6.0+
- **gRPC**: google.golang.org/grpc v1.64.0
- **Task Queue**: Asynq v0.24.1

### Tool Libraries
- **Logging**: Zap v1.27.0
- **Permission**: Casbin v2.89.0
- **JWT**: golang-jwt/jwt v4.5.0
- **Configuration**: Viper v1.18.2
- **CLI**: Cobra v1.8.0
- **Validation**: validator v10+
- **Utils**: Lancet v2.3.2

## System Requirements

- Go 1.23.1+
- MySQL 5.7+ or PostgreSQL
- Redis 6.0+
- Docker & Docker Compose (optional)

## Next Steps

- [Getting Started](./getting-started.md) - Launch a project in 10 minutes
- [Project Structure](./project-structure.md) - Understand the directory structure
- [Core Architecture](../essentials/architecture.md) - Deep dive into architectural design
