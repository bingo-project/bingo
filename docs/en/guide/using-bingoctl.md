---
title: Using bingoctl - Bingo Go CLI Tool Guide
description: Master the bingoctl CLI tool to create projects, generate CRUD code, database models, migrations, and more. Boost your Bingo Go microservices development efficiency.
---

# Using bingoctl

[bingoctl](https://github.com/bingo-project/bingoctl) is the official CLI tool for the Bingo framework, used for quickly creating projects and generating code, significantly improving development efficiency.

## Installation

```bash
go install github.com/bingo-project/bingoctl@latest
```

Verify installation:

```bash
bingoctl version
```

## Core Features

### 1. Create Project

Create a complete Bingo project scaffold with one command:

```bash
bingoctl create github.com/myorg/myapp
```

Generated project includes:
- Complete directory structure
- Configuration file templates
- Docker Compose configuration
- Makefile
- Basic example code

#### Create Command Options

**Project Name and Module Name**

```bash
# Use default module name (same as project name)
bingoctl create github.com/myorg/myapp

# Customize module name (-m option)
bingoctl create myapp -m github.com/mycompany/myapp
```

**Template Version Control**

```bash
# Use recommended version (default)
bingoctl create myapp

# Use specific version
bingoctl create myapp -r v1.2.3

# Use branch (development version)
bingoctl create myapp -r main

# Force re-download template
bingoctl create myapp -r main --no-cache
```

**Service Selection**

Control which services are included in the project (apiserver, admserver, scheduler, bot, etc.):

```bash
# Only include apiserver (default)
bingoctl create myapp

# Create all available services
bingoctl create myapp --all
# or use shorthand
bingoctl create myapp -a

# Explicitly specify services (comma-separated)
bingoctl create myapp --services apiserver,admserver,scheduler

# Add service to default apiserver
bingoctl create myapp --add-service scheduler

# Exclude specific service
bingoctl create myapp --no-service bot

# Include only skeleton, no services
bingoctl create myapp --services none
```

**Git Initialization**

```bash
# Create project and initialize git repository (default)
bingoctl create myapp

# Create project without git initialization
bingoctl create myapp --init-git=false
```

**Cache Management and Mirror Configuration**

```bash
# Cache location: ~/.bingoctl/templates/

# For regions with GitHub access difficulties, configure mirror
export BINGOCTL_TEMPLATE_MIRROR=https://ghproxy.com/
bingoctl create myapp

# Or temporarily set
BINGOCTL_TEMPLATE_MIRROR=https://ghproxy.com/ bingoctl create myapp
```

### 2. Code Generation

Quickly generate code for each layer following Bingo's best practices.

#### Global Options

```bash
-d, --directory string   Specify directory for generated files
-p, --package string     Specify package name
-t, --table string       Read fields from database table
-s, --service string     Target service name, for auto-inferring path
```

#### Multi-Service Support

When project contains multiple services, use `--service` parameter to auto-infer generation path:

```bash
# Generate code for default service (usually apiserver)
bingoctl make model user

# Generate for specific service, auto-inferring path
bingoctl make model user --service admserver

# Generate complete CRUD (for specified service)
bingoctl make crud order --service admserver

# Explicitly specify directory (highest priority, overrides --service)
bingoctl make model user -d custom/path
```

**Path Inference Rules:**
1. Scan `cmd/` directory to identify existing services
2. If configured path contains service name, intelligently replace (e.g., `internal/apiserver/model` → `internal/admserver/model`)
3. Otherwise use default pattern: `internal/{service}/{suffix}`

#### Complete CRUD Code Generation

Generate all layers at once:

```bash
bingoctl make crud user
```

Auto-generates:
- `internal/pkg/model/user.go` - Data model
- `internal/apiserver/store/user.go` - Data access layer
- `internal/apiserver/biz/user/user.go` - Business logic layer
- `internal/apiserver/controller/v1/user/user.go` - Controller layer
- `pkg/api/v1/user.go` - Request/response definitions

And auto-registers to:
- Store interface
- Biz interface
- Routes

#### Generate Individual Layers

```bash
# Generate Model
bingoctl make model article

# Generate Store layer
bingoctl make store article

# Generate Biz layer
bingoctl make biz article

# Generate Controller layer
bingoctl make controller article

# Generate Request validation
bingoctl make request article
```

### 3. Generate Code from Database

Auto-generate Model code from existing database tables:

```bash
# Generate single table
bingoctl gen -t users

# Generate multiple tables
bingoctl gen -t users,posts,comments
```

**Prerequisites**: Need to configure database connection in `.bingoctl.yaml`.

### 4. Database Migration

Generate database migration files:

```bash
bingoctl make migration create_users_table
```

Generated migration files are located in `internal/apiserver/database/migration/`.

Optional: Generate migration from database table

```bash
bingoctl make migration create_posts_table -t posts
```

### 5. Generate Service Modules

Generate complete service modules:

```bash
# Generate API service with HTTP server
bingoctl make service api --http --with-store --with-controller

# Generate service with gRPC server
bingoctl make service rpc --grpc

# Generate service supporting both HTTP and gRPC
bingoctl make service gateway --http --grpc

# Generate pure worker service for business processing
bingoctl make service worker --no-biz
```

Service options:
- `--http`: Enable HTTP server
- `--grpc`: Enable gRPC server
- `--with-biz`: Generate business layer (default true)
- `--no-biz`: Don't generate business layer
- `--with-store`: Generate storage layer
- `--with-controller`: Generate controller layer
- `--with-middleware`: Generate middleware directory
- `--with-router`: Generate router directory

### 6. Other Generators

```bash
# Generate middleware
bingoctl make middleware auth

# Generate scheduled task
bingoctl make job cleanup

# Generate data seeder
bingoctl make seeder users

# Generate CLI command
bingoctl make cmd serve
```

## Configuration File

Copy example file to create `.bingoctl.yaml` in project root:

```bash
cp .bingoctl.example.yaml .bingoctl.yaml
```

Configuration file content:

```yaml
version: v1

# Project package name
rootPackage: github.com/myorg/myapp

# Directory configuration
directory:
  cmd: internal/bingoctl/cmd
  model: internal/pkg/model
  store: internal/apiserver/store
  request: pkg/api/v1
  biz: internal/apiserver/biz
  controller: internal/apiserver/controller/v1
  middleware: internal/pkg/middleware
  job: internal/watcher/watcher
  migration: internal/apiserver/database/migration
  seeder: internal/apiserver/database/seeder

# Registry configuration
registries:
  router: internal/apiserver/router/api.go
  store:
    filePath: internal/apiserver/store/store.go
    interface: "IStore"
  biz:
    filePath: internal/apiserver/biz/biz.go
    interface: "IBiz"

# Database configuration (for gen command)
mysql:
  host: 127.0.0.1:3306
  username: root
  password: your-password
  database: myapp
```

## Practical Examples

### Example 1: Create Blog System from Scratch

```bash
# 1. Create project
bingoctl create github.com/myorg/blog
cd blog

# 2. Configure database
vim .bingoctl.yaml  # Modify mysql configuration

# 3. Start dependency services
docker-compose -f deployments/docker/docker-compose.yaml up -d

# 4. Generate user module
bingoctl make crud user

# 5. Generate post module
bingoctl make crud post

# 6. Generate comment module
bingoctl make crud comment

# 7. Run service
make build
./blog-apiserver
```

### Example 2: Generate Code from Existing Database

If you have an existing database:

```bash
# 1. Configure database connection
vim .bingoctl.yaml

# 2. Generate Model from database
bingoctl gen -t users,posts,comments,tags

# 3. Generate complete CRUD code for each table
bingoctl make crud user
bingoctl make crud post
bingoctl make crud comment
bingoctl make crud tag
```

### Example 3: Generate Scheduled Tasks

```bash
# 1. Generate task code
bingoctl make job daily_report

# 2. Edit task logic
vim internal/watcher/watcher/daily_report.go

# 3. Register task in scheduler
```

### Example 4: Generate New Microservice

```bash
# Generate independent notification service
bingoctl make service notification \
  --http \
  --with-store \
  --with-controller \
  --with-router

# Generate pure gRPC service
bingoctl make service user-grpc --grpc
```

## Command Reference

### Global Options

```bash
-c, --config string   Configuration file path (default .bingoctl.yaml)
```

### create Command

```bash
bingoctl create <package-name>
```

### make Command

```bash
bingoctl make <type> <name> [options]

Options:
  -d, --directory string   Specify generation directory
  -p, --package string     Specify package name
  -t, --table string       Read fields from database table
```

Supported types:
- `crud` - Complete CRUD code
- `model` - Data model
- `store` - Storage layer
- `biz` - Business logic layer
- `controller` - Controller layer
- `request` - Request validation
- `middleware` - Middleware
- `job` - Scheduled task
- `migration` - Database migration
- `seeder` - Data seeding
- `service` - Service module
- `cmd` - CLI command

### gen Command

```bash
bingoctl gen -t <table1,table2,...>
```

## Development Workflow

Recommended development workflow:

```
1. Create project
   ↓
   bingoctl create github.com/myorg/app

2. Configure database
   ↓
   Edit .bingoctl.yaml

3. Generate business modules
   ↓
   bingoctl make crud user
   bingoctl make crud post

4. Customize business logic
   ↓
   Edit Biz layer code

5. Add middleware/tasks
   ↓
   bingoctl make middleware auth
   bingoctl make job cleanup

6. Run and test
   ↓
   make build && ./app-apiserver
```

## Best Practices

### 1. Use CRUD Generator for Quick Start

For standard CRUD operations, use `make crud` directly:

```bash
bingoctl make crud product
```

### 2. Generate from Database to Reduce Manual Work

If you already have database design, use `gen` command:

```bash
bingoctl gen -t products,categories,orders
```

### 3. Test Immediately After Generation

Verify generated code works by running service:

```bash
bingoctl make crud user
make build
./app-apiserver
curl http://localhost:8080/v1/users
```

### 4. Customize Generated Code

Generated code is a starting point. Customize based on actual needs:
- Add business rules in Biz layer
- Add parameter validation in Controller layer
- Optimize queries in Store layer

### 5. Version Control Configuration Files

Commit `.bingoctl.yaml` to version control:

```bash
git add .bingoctl.yaml
git commit -m "feat: add bingoctl config"
```

## FAQ

### Q: Will generated code overwrite my changes?

A: bingoctl doesn't overwrite existing files by default. If a file exists, it prompts whether to overwrite.

### Q: How do I customize generation templates?

A: bingoctl uses built-in templates. To customize, fork the bingoctl repository and modify templates.

### Q: Can generated code be used in production directly?

A: Generated code follows Bingo best practices but still needs adjustment based on actual business:
- Add business validation
- Optimize query performance
- Add error handling
- Write unit tests

### Q: What's the difference between bingoctl and bingoctl in the project?

A: They are two different tools:
- **bingoctl (CLI tool)**: Independent project scaffold and code generation tool
- **cmd/bingoctl (project component)**: Built-in CLI tool in Bingo projects for database migration etc.

## Next Step

- [Overall Architecture](../essentials/architecture.md) - Deep dive into microservices architecture design

## Reference Resources

- [bingoctl GitHub Repository](https://github.com/bingo-project/bingoctl)
- [bingoctl README](https://github.com/bingo-project/bingoctl/blob/main/README.md)
