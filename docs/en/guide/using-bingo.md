---
title: Using Bingo CLI - Bingo Go CLI Tool Guide
description: Master the Bingo CLI tool to create projects, generate CRUD code, database models, migrations, and more. Boost your Bingo Go microservices development efficiency.
---

# Using Bingo CLI

[Bingo CLI](https://github.com/bingo-project/bingoctl) is the official CLI tool for the Bingo framework, used for quickly creating projects and generating code, significantly improving development efficiency.

## Installation

```bash
go install github.com/bingo-project/bingoctl/cmd/bingo@latest
```

> To install an older version (v1.4.x with built-in templates), build from source: `git clone https://github.com/bingo-project/bingoctl && cd bingoctl && git checkout v1.4.7 && go build -o bingo ./cmd/bingoctl`
> See [CHANGELOG](https://github.com/bingo-project/bingoctl/blob/main/docs/en/CHANGELOG.md) for version history

Verify installation:

```bash
bingo version
```

## Shell Completion

Bingo supports command-line auto-completion for multiple shells.

### Zsh

```bash
# Temporary (current session)
source <(bingo completion zsh)

# Permanent
## Linux
bingo completion zsh > "${fpath[1]}/_bingo"

## macOS (Homebrew)
bingo completion zsh > $(brew --prefix)/share/zsh/site-functions/_bingo
```

> If completion doesn't work, ensure `.zshrc` has: `autoload -U compinit; compinit`

### Bash

```bash
# Temporary (current session)
source <(bingo completion bash)

# Permanent
## Linux
bingo completion bash > /etc/bash_completion.d/bingo

## macOS (Homebrew)
bingo completion bash > $(brew --prefix)/etc/bash_completion.d/bingo
```

> Requires the `bash-completion` package

### Fish

```bash
bingo completion fish > ~/.config/fish/completions/bingo.fish
```

### PowerShell

```powershell
bingo completion powershell > bingo.ps1
# Add the generated script to your PowerShell profile
```

## Core Features

### 1. Create Project

Create a complete Bingo project scaffold with one command:

```bash
bingo create github.com/myorg/myapp
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
bingo create github.com/myorg/myapp

# Customize module name (-m option)
bingo create myapp -m github.com/mycompany/myapp
```

**Template Version Control**

```bash
# Use recommended version (default)
bingo create myapp

# Use specific version
bingo create myapp -r v1.2.3

# Use branch (development version)
bingo create myapp -r main

# Force re-download template
bingo create myapp -r main --no-cache
```

**Service Selection**

Control which services are included in the project (apiserver, admserver, scheduler, bot, etc.):

```bash
# Only include apiserver (default)
bingo create myapp

# Create all available services
bingo create myapp --all
# or use shorthand
bingo create myapp -a

# Explicitly specify services (comma-separated)
bingo create myapp --services apiserver,admserver,scheduler

# Add service to default apiserver
bingo create myapp --add-service scheduler

# Exclude specific service
bingo create myapp --no-service bot

# Include only skeleton, no services
bingo create myapp --services none
```

**Git Initialization**

```bash
# Create project and initialize git repository (default)
bingo create myapp

# Create project without git initialization
bingo create myapp --init-git=false
```

**Build Options**

```bash
# Create project without building (default)
bingo create myapp

# Create project and run make build
bingo create myapp --build
```

**Cache Management and Mirror Configuration**

```bash
# Cache location: ~/.bingo/templates/

# For regions with GitHub access difficulties, configure mirror
export BINGO_TEMPLATE_MIRROR=https://ghproxy.com/
bingo create myapp

# Or temporarily set
BINGO_TEMPLATE_MIRROR=https://ghproxy.com/ bingo create myapp
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
bingo make model user

# Generate for specific service, auto-inferring path
bingo make model user --service admserver

# Generate complete CRUD (for specified service)
bingo make crud order --service admserver

# Explicitly specify directory (highest priority, overrides --service)
bingo make model user -d custom/path
```

**Path Inference Rules:**
1. Scan `cmd/` directory to identify existing services
2. If configured path contains service name, intelligently replace (e.g., `internal/apiserver/model` → `internal/admserver/model`)
3. Otherwise use default pattern: `internal/{service}/{suffix}`

#### Complete CRUD Code Generation

Generate all layers at once:

```bash
bingo make crud user
```

Auto-generates:
- `internal/pkg/model/user.go` - Data model
- `internal/apiserver/store/user.go` - Data access layer
- `internal/apiserver/biz/user/user.go` - Business logic layer
- `internal/apiserver/handler/http/user/user.go` - Handler layer
- `pkg/api/v1/user.go` - Request/response definitions

And auto-registers to:
- Store interface
- Biz interface
- Routes

#### Generate Individual Layers

```bash
# Generate Model
bingo make model article

# Generate Store layer
bingo make store article

# Generate Biz layer
bingo make biz article

# Generate Handler layer
bingo make handler article

# Generate Request validation
bingo make request article
```

### 3. Generate Code from Database

Auto-generate Model code from existing database tables:

```bash
# Generate single table
bingo gen -t users

# Generate multiple tables
bingo gen -t users,posts,comments
```

**Prerequisites**: Need to configure database connection in `.bingo.yaml`.

### 4. Database Migration

Generate database migration files:

```bash
bingo make migration create_users_table
```

Generated migration files are located in `internal/pkg/database/migration/`.

Optional: Generate migration from database table

```bash
bingo make migration create_posts_table -t posts
```

**Run Migrations**

```bash
bingo migrate <command> [options]

# Options
-v, --verbose   Show detailed compilation output
    --rebuild   Force recompile migration program
-f, --force     Force execution in production environment

# Subcommands
bingo migrate up          # Run all pending migrations
bingo migrate rollback    # Rollback the last batch of migrations
bingo migrate reset       # Rollback all migrations
bingo migrate refresh     # Rollback all and re-run migrations
bingo migrate fresh       # Drop all tables and re-run migrations
```

**Configure Migration Table Name** (optional, in `.bingo.yaml`):

```yaml
migrate:
  table: bingo_migration  # Default value
```

### 5. Generate Service Modules

Generate complete service modules:

```bash
# Generate API service with HTTP server
bingo make service api --http --with-store --with-handler

# Generate service with gRPC server
bingo make service rpc --grpc

# Generate service supporting both HTTP and gRPC
bingo make service gateway --http --grpc

# Generate pure worker service for business processing
bingo make service worker --no-biz
```

Service options:
- `--http`: Enable HTTP server
- `--grpc`: Enable gRPC server
- `--with-biz`: Generate business layer (default true)
- `--no-biz`: Don't generate business layer
- `--with-store`: Generate storage layer
- `--with-handler`: Generate handler layer
- `--with-middleware`: Generate middleware directory
- `--with-router`: Generate router directory

### 6. Other Generators

```bash
# Generate middleware
bingo make middleware auth

# Generate scheduled task
bingo make job cleanup

# Generate data seeder
bingo make seeder users

# Generate CLI command
bingo make cmd serve
```

### 7. Run Database Seeders

Run user-defined seeders to populate the database:

```bash
bingo db seed [options]

# Options
-v, --verbose      Show detailed compilation output
    --rebuild      Force recompile seeder program
    --seeder       Specify seeder class name to run

# Examples
bingo db seed                    # Run all seeders
bingo db seed --seeder=User      # Run only UserSeeder
bingo db seed -v                 # Show detailed output
```

## Configuration File

Copy example file to create `.bingo.yaml` in project root:

```bash
cp .bingo.example.yaml .bingo.yaml
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
  handler: internal/apiserver/handler/http
  middleware: internal/pkg/middleware
  job: internal/watcher/watcher
  migration: internal/pkg/database/migration
  seeder: internal/pkg/database/seeder

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
bingo create github.com/myorg/blog
cd blog

# 2. Configure database
vim .bingo.yaml  # Modify mysql configuration

# 3. Start dependency services
docker-compose -f deployments/docker/docker-compose.yaml up -d

# 4. Generate user module
bingo make crud user

# 5. Generate post module
bingo make crud post

# 6. Generate comment module
bingo make crud comment

# 7. Run service
make build
./_output/platforms/<os>/<arch>/blog-apiserver
```

### Example 2: Generate Code from Existing Database

If you have an existing database:

```bash
# 1. Configure database connection
vim .bingo.yaml

# 2. Generate Model from database
bingo gen -t users,posts,comments,tags

# 3. Generate complete CRUD code for each table
bingo make crud user
bingo make crud post
bingo make crud comment
bingo make crud tag
```

### Example 3: Generate Scheduled Tasks

```bash
# 1. Generate task code
bingo make job daily_report

# 2. Edit task logic
vim internal/watcher/watcher/daily_report.go

# 3. Register task in scheduler
```

### Example 4: Generate New Microservice

```bash
# Generate independent notification service
bingo make service notification \
  --http \
  --with-store \
  --with-handler \
  --with-router

# Generate pure gRPC service
bingo make service user-grpc --grpc
```

## Command Reference

### Global Options

```bash
-c, --config string   Configuration file path (default .bingo.yaml)
```

### create Command

```bash
bingo create <package-name>
```

### make Command

```bash
bingo make <type> <name> [options]

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
- `handler` - Handler layer
- `request` - Request validation
- `middleware` - Middleware
- `job` - Scheduled task
- `migration` - Database migration
- `seeder` - Data seeding
- `service` - Service module
- `cmd` - CLI command

### gen Command

```bash
bingo gen -t <table1,table2,...>
```

## Development Workflow

Recommended development workflow:

```
1. Create project
   ↓
   bingo create github.com/myorg/app

2. Configure database
   ↓
   Edit .bingo.yaml

3. Generate business modules
   ↓
   bingo make crud user
   bingo make crud post

4. Customize business logic
   ↓
   Edit Biz layer code

5. Add middleware/tasks
   ↓
   bingo make middleware auth
   bingo make job cleanup

6. Run and test
   ↓
   make build && ./_output/platforms/<os>/<arch>/app-apiserver
```

## Best Practices

### 1. Use CRUD Generator for Quick Start

For standard CRUD operations, use `make crud` directly:

```bash
bingo make crud product
```

### 2. Generate from Database to Reduce Manual Work

If you already have database design, use `gen` command:

```bash
bingo gen -t products,categories,orders
```

### 3. Test Immediately After Generation

Verify generated code works by running service:

```bash
bingo make crud user
make build
./_output/platforms/<os>/<arch>/app-apiserver
curl http://localhost:8080/v1/users
```

### 4. Customize Generated Code

Generated code is a starting point. Customize based on actual needs:
- Add business rules in Biz layer
- Add parameter validation in Handler layer
- Optimize queries in Store layer

### 5. Version Control Configuration Files

Commit `.bingo.yaml` to version control:

```bash
git add .bingo.yaml
git commit -m "feat: add bingo config"
```

## FAQ

### Q: Will generated code overwrite my changes?

A: bingo doesn't overwrite existing files by default. If a file exists, it prompts whether to overwrite.

### Q: How do I customize generation templates?

A: bingo uses built-in templates. To customize, fork the bingoctl repository and modify templates.

### Q: Can generated code be used in production directly?

A: Generated code follows Bingo best practices but still needs adjustment based on actual business:
- Add business validation
- Optimize query performance
- Add error handling
- Write unit tests

### Q: What's the difference between Bingo CLI and bingoctl in the project?

A: They are two different tools:
- **bingo (CLI tool)**: Independent project scaffold and code generation tool
- **cmd/bingoctl (project component)**: Built-in CLI tool in Bingo projects, can be extended with custom commands

## Next Step

- [Overall Architecture](../essentials/architecture.md) - Deep dive into microservices architecture design

## Reference Resources

- [Bingo CLI GitHub Repository](https://github.com/bingo-project/bingoctl)
- [Bingo CLI README](https://github.com/bingo-project/bingoctl/blob/main/README.md)
