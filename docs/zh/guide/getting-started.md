---
title: 快速开始 - 10分钟上手 Bingo Go 微服务框架
description: 使用 bingo CLI 快速创建 Bingo Go 微服务项目，10分钟内启动并运行第一个 API。本指南提供完整的安装、配置和运行步骤，帮助你快速开始 Golang 后端开发。
---

# 快速开始

本指南将帮助你在 10 分钟内启动 Bingo 项目并运行第一个 API。

## 创建项目

### 方式一: 使用 bingo CLI（推荐）

使用 [bingo CLI](https://github.com/bingo-project/bingoctl) 工具是创建 Bingo 项目最快的方式。

```bash
# 安装 bingo CLI
go install github.com/bingo-project/bingoctl/cmd/bingo@latest

# 创建新项目
bingo create github.com/myorg/myapp

# 进入项目目录
cd myapp
```

bingo 会自动生成完整的项目结构,包括:
- 基础配置文件
- Docker Compose 配置
- Makefile
- 示例代码

### 方式二: 克隆 Bingo 仓库

如果你想基于 Bingo 源码进行开发:

```bash
git clone https://github.com/bingo-project/bingo.git
cd bingo
```

---

## 配置服务

复制示例配置文件到项目根目录:

```bash
cp configs/*.example.yaml .
# 重命名配置文件（去掉 .example 后缀）
for f in *.example.yaml; do mv "$f" "${f%.example.yaml}.yaml"; done
```

编辑配置文件，修改 MySQL 和 Redis 连接信息:

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

## 构建和运行

### 1. 构建项目

```bash
make build
```

> **说明**: `make build` 会将二进制文件输出到 `./_output/platforms/<os>/<arch>/` 目录（如 `./_output/platforms/darwin/arm64/`）

### 2. 数据库迁移

```bash
# 执行数据库迁移
bingo migrate up
```

### 3. 运行服务

```bash
./_output/platforms/<os>/<arch>/<app>-apiserver

# 例如 macOS ARM64 + bingo 项目:
./_output/platforms/darwin/arm64/bingo-apiserver
```

### 4. 验证服务

```bash
# 检查服务状态
curl http://localhost:8080/health

# 访问 Swagger 文档
open http://localhost:8080/swagger/index.html
```

## 常用命令

```bash
# 编译所有服务
make build

# 编译指定服务
make build BINS="bingo-apiserver"

# 运行测试
make test

# 代码检查
make lint

# 生成 Swagger 文档
make swagger

# 清理构建产物
make clean
```

## 下一步

- [项目结构](./project-structure.md) - 理解项目目录组织

## 遇到问题?

如果遇到任何问题，请在 GitHub 上提交 Issue。
