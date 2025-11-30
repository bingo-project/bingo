---
title: Getting Started - Quick Start with Bingo Go Microservices Framework
description: Quickly create a Bingo Go microservices project using bingoctl in 10 minutes. Complete guide with installation, configuration and running steps to start Golang backend development.
---

# Getting Started

This guide will help you launch a Bingo project and run your first API within 10 minutes.

## Create Project

### Option 1: Using bingoctl (Recommended)

Using the [bingoctl](https://github.com/bingo-project/bingoctl) CLI tool is the fastest way to create a Bingo project.

```bash
# Install bingoctl
go install github.com/bingo-project/bingoctl@latest

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

### Option 2: Clone Bingo Repository

If you want to develop based on Bingo source code:

```bash
git clone https://github.com/bingo-project/bingo.git
cd bingo
```

---

## Configure Services

Copy example configuration files to project root:

```bash
cp configs/*.example.yaml .
# Rename config files (remove .example suffix)
for f in *.example.yaml; do mv "$f" "${f%.example.yaml}.yaml"; done
```

Edit configuration files to set MySQL and Redis connection:

```yaml
# <app>-apiserver.yaml
mysql:
  host: 127.0.0.1:3306
  username: root
  password: your-password
  database: your-database

redis:
  host: 127.0.0.1:6379
  password: ""
```

---

## Build and Run

### 1. Build Project

```bash
make build
```

> **Note**: `make build` outputs binary files to `./_output/platforms/<os>/<arch>/` directory (e.g., `./_output/platforms/darwin/arm64/`)

### 2. Database Migration

```bash
# Execute database migration (choose path based on your OS and architecture)
./_output/platforms/<os>/<arch>/<app>ctl migrate up

# Example for macOS ARM64 + bingo project:
./_output/platforms/darwin/arm64/bingoctl migrate up
```

### 3. Run Service

```bash
./_output/platforms/<os>/<arch>/<app>-apiserver

# Example for macOS ARM64 + bingo project:
./_output/platforms/darwin/arm64/bingo-apiserver
```

### 4. Verify Service

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

## Next Step

- [Project Structure](./project-structure.md) - Understand project directory organization

## Troubleshooting

If you encounter any issues, please submit an Issue on GitHub.
