---
title: Project Structure - Bingo Go Microservices Directory Organization
description: Understand the Bingo Go microservices framework project directory structure, including cmd, internal, pkg directories and the responsibilities of each file.
---

# Project Structure

This document describes the directory structure of a Bingo project and the responsibility of each module.

## Directory Structure

```
bingo/
├── cmd/                    # Executable program entry points
│   ├── bingo-apiserver/    # API server
│   ├── bingo-admserver/    # Admin server
│   ├── bingo-scheduler/    # Scheduler service
│   ├── bingo-bot/          # Bot service
│   └── bingoctl/           # CLI tool
├── internal/               # Internal application code
│   ├── apiserver/          # API server implementation
│   │   ├── controller/     # HTTP handlers
│   │   ├── biz/            # Business logic layer
│   │   ├── store/          # Data access layer
│   │   ├── middleware/     # Middleware
│   │   └── router/         # Route definitions
│   ├── admserver/          # Admin server implementation
│   ├── scheduler/          # Scheduler implementation
│   ├── bot/                # Bot service implementation
│   └── pkg/                # Internal shared packages
│       ├── model/          # Data models
│       ├── middleware/     # Shared middleware
│       └── ...
├── pkg/                    # Public packages (can be imported by external projects)
│   ├── api/                # API definitions
│   └── ...
├── docs/                   # Project documentation
│   ├── en/                 # English documentation
│   ├── zh/                 # Chinese documentation
│   └── .vitepress/         # VitePress configuration
├── configs/                # Configuration files
│   ├── bingo-apiserver.example.yaml
│   ├── bingo-admserver.example.yaml
│   └── ...
├── deployments/            # Deployment configurations
│   ├── docker/             # Docker configurations
│   ├── k8s/                # Kubernetes configurations
│   └── ...
├── scripts/                # Utility scripts
├── Makefile                # Build configuration
├── go.mod                  # Go module definition
└── go.sum                  # Go module checksums
```

## Core Directories

### cmd/ - Entry Points

Each service has an entry point in `cmd/`:

- **bingo-apiserver**: Main API service, serves HTTP requests
- **bingo-admserver**: Admin backend service, handles admin operations
- **bingo-scheduler**: Task scheduler service, handles scheduled tasks
- **bingo-bot**: Bot service, handles bot integrations
- **bingoctl**: CLI tool for code generation and management

### internal/ - Application Code

Private application code that follows the layered architecture:

#### Three-Layer Architecture

Each service implements the three-layer model:

```
Request
   ↓
Controller (HTTP Handler)
   ↓
Biz (Business Logic)
   ↓
Store (Data Access)
   ↓
Database/External Services
```

- **controller/**: HTTP request handlers, parameter validation, response formatting
- **biz/**: Business logic implementation, contains core application logic
- **store/**: Data access layer, database operations and queries
- **middleware/**: Request/response processing middleware
- **router/**: Route definitions and registration

#### Package Organization

- **model/**: Data structures and database models
- **middleware/**: Shared middleware components
- **util/**: Utility functions
- **constant/**: Constants and enums

### pkg/ - Public Packages

Public code that can be imported by other projects:

- **api/**: API request/response definitions
- **errors/**: Custom error types
- **middleware/**: Reusable middleware

### configs/ - Configuration

Configuration files for each service:

```yaml
# Example: bingo-apiserver.yaml
server:
  port: 8080

database:
  driver: mysql
  dsn: root:password@tcp(localhost:3306)/bingo

redis:
  addr: localhost:6379

log:
  level: info
  format: json
```

### deployments/ - Deployment

- **docker/**: Docker Compose files for local development
- **k8s/**: Kubernetes manifests for production deployment

## Naming Conventions

### File and Directory Names

- Use lowercase with hyphens for directories and files
- Example: `user_service.go`, `user-controller/`

### Package Names

- Package names should be lowercase, single word when possible
- Example: `controller`, `biz`, `store`

### Variable and Function Names

- Use camelCase for variables and functions
- Example: `getUserByID()`, `userData`

## Best Practices

1. **Separation of Concerns**: Each layer has clear responsibilities
2. **Interface-based**: Use interfaces for dependency injection
3. **Testing**: Place test files alongside code with `_test.go` suffix
4. **Constants**: Keep magic numbers and strings in dedicated constant files
5. **Error Handling**: Use custom error types for better error handling

## Next Step

- [Develop Your First Feature](./first-feature.md) - Learn layer responsibilities through hands-on practice
