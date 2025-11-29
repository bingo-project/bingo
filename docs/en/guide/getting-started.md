---
title: Getting Started - Quick Start with Bingo Go Microservices Framework
description: Quickly create a Bingo Go microservices project using bingoctl in 10 minutes. Complete guide with installation, configuration and running steps to start Golang backend development.
---

# Getting Started

This guide will help you launch a Bingo project and run your first API within 10 minutes.

## Recommended: Create Project with bingoctl

Using the [bingoctl](https://github.com/bingo-project/bingoctl) CLI tool is the fastest way to create a Bingo project.

### 1. Install bingoctl

```bash
go install github.com/bingo-project/bingoctl@latest
```

### 2. Create Project

```bash
# Create a new project
bingoctl create github.com/myorg/myapp

# Enter project directory
cd myapp
```

bingoctl will automatically generate a complete project structure, including:
- Basic configuration files
- Docker Compose configuration
- Makefile
- Example code

### 3. Configure Database Connection

Copy the example configuration file:

```bash
cp .bingoctl.example.yaml .bingoctl.yaml
```

Edit `.bingoctl.yaml` to configure database connection:

```yaml
mysql:
  host: 127.0.0.1:3306
  username: root
  password: your-password
  database: myapp
```

### 4. Start Dependency Services

```bash
docker-compose -f deployments/docker/docker-compose.yaml up -d
```

### 5. Generate Your First Module

```bash
# Generate complete CRUD code for user module
bingoctl make crud user
```

This will automatically generate:
- Model (data model)
- Store (data access layer)
- Biz (business logic layer)
- Controller (HTTP handler layer)
- Request (request validation)

### 6. Run Service

```bash
make build
./myapp-apiserver
```

### 7. Verify Service

```bash
# Check service status
curl http://localhost:8080/health

# Access Swagger documentation
open http://localhost:8080/swagger/index.html
```

---

## Alternative: Clone Bingo Repository

If you want to develop based on Bingo source code:

### 1. Clone Project

```bash
git clone <repository-url>
cd bingo
```

### 2. Configure Environment

```bash
# Copy configuration file
cp configs/bingo-apiserver.example.yaml bingo-apiserver.yaml

# Modify configuration based on your environment
vim bingo-apiserver.yaml
```

Main configuration items:
- Database connection (MySQL)
- Redis connection
- JWT secret

### 3. Start Dependency Services

Use Docker Compose to quickly start MySQL and Redis:

```bash
docker-compose -f deployments/docker/docker-compose.yaml up -d mysql redis
```

### 4. Database Migration

```bash
# Build project (output path: ./_output/platforms/<os>/<arch>/)
make build

# Copy and modify configuration file
cp configs/{app}-admserver.example.yaml {app}-admserver.yaml

# Build your app ctl
make build BINS="{app}ctl"

# Execute database migration
./_output/platforms/{os}/{arch}/{app}ctl migrate up
```

> **Note**: `make build` outputs binary files to `./_output/platforms/<os>/<arch>/` directory (e.g., `./_output/platforms/darwin/arm64/`)

### 5. Start Service

**Method 1: Direct Run**

```bash
make build
bingo-apiserver -c bingo-apiserver.yaml
```

**Method 2: Development Mode (Hot Reload)**

```bash
cp .air.example.toml .air.toml
air
```

### 6. Verify Service

```bash
# Check service status
curl http://localhost:8080/health

# Access Swagger documentation
open http://localhost:8080/swagger/index.html
```

## Common Commands

```bash
# Build all services
make build

# Build specific service
make build BINS="bingo-apiserver"

# Run tests
make test

# Code lint check
make lint

# Generate Swagger documentation
make swagger

# Clean build artifacts
make clean
```

## Next Steps

- [Using bingoctl](./using-bingoctl.md) - Deep dive into bingoctl's powerful features
- [Develop Your First Feature](./first-feature.md) - Learn Bingo development through examples
- [Project Structure](./project-structure.md) - Understand project directory organization
- [Layered Architecture](../essentials/layered-design.md) - Understand three-layer architecture design

## Troubleshooting

If you encounter any issues, please submit an Issue on GitHub.
