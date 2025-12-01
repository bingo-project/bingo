[English](README.md) | 中文

# Bingo - 生产级 Go 微服务脚手架

> 一个开箱即用的 Go/Golang 微服务脚手架框架，基于微服务架构设计，让开发者只需关注业务开发。

[![Go Version](https://img.shields.io/badge/Go-1.23%2B-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

## 🌐 完整文档

📚 **官方文档网站**: [bingoctl.dev](https://bingoctl.dev)

**快速导航**:
- 🚀 [快速开始](https://bingoctl.dev/guide/getting-started) - 10 分钟快速上手
- 📖 [什么是 Bingo](https://bingoctl.dev/guide/what-is-bingo) - 了解核心特性
- 🏗️ [整体架构](https://bingoctl.dev/essentials/architecture) - 微服务架构设计
- 🛠️ [使用 bingo CLI](https://bingoctl.dev/guide/using-bingo) - CLI 工具指南
- 🇬🇧 [English Documentation](https://bingoctl.dev/en/) - English version

## 🎯 项目定位

**Bingo** 是一个**生产级的 Go/Golang 微服务脚手架框架**，提供完整的：
- ✅ 微服务架构设计（Controller → Biz → Store 三层架构）
- ✅ 核心组件预集成（Gin、GORM、Redis、Asynq、Casbin）
- ✅ 工程化能力（代码生成、热重启、Docker 支持）
- ✅ 生产级特性（日志、监控、链路追踪）
- ✅ 最佳实践和完整文档

**适用场景**: 中后台系统、微服务项目、RESTful API、gRPC 服务

🔗 **相关项目**: [bingo CLI](https://github.com/bingo-project/bingoctl) - Bingo 项目脚手架工具

## 核心特性

### 架构层面
- **微服务架构**: 多服务独立部署,支持水平扩展
- **分层设计**: Controller → Biz → Store 清晰的三层架构
- **通用数据层**: 基于泛型的 Store[T] 设计,减少重复代码
- **依赖注入**: 基于接口编程,易于测试和扩展
- **服务发现**: 支持 gRPC 服务间通信

### 技术组件
- **Web 框架**: Gin - 高性能 HTTP 框架
- **ORM**: GORM - 支持多种数据库
- **缓存**: Redis - 分布式缓存
- **任务队列**: Asynq - 可靠的异步任务处理
- **权限控制**: Casbin - 灵活的 RBAC 权限引擎
- **日志系统**: Zap - 结构化高性能日志
- **API 文档**: Swagger - 自动生成 API 文档

### 工程能力
- **热重启**: Air 支持开发时热重启
- **代码生成**: 自动生成 CRUD 代码和 API 文档
- **Docker 支持**: 一键容器化部署
- **监控指标**: Prometheus + pprof 性能监控

## 技术栈

- **Go**: 1.23.1+
- **Web 框架**: Gin v1.10.0
- **ORM**: GORM v1.25.10
- **数据库**: MySQL 5.7+ / PostgreSQL
- **缓存**: Redis 6.0+
- **gRPC**: google.golang.org/grpc v1.64.0
- **任务队列**: Asynq v0.24.1

## 快速开始

### 方式一: 使用 bingo CLI 创建新项目 (推荐)

使用 [bingo CLI](https://github.com/bingo-project/bingoctl) 工具快速创建项目:

```bash
# 安装 bingo CLI
go install github.com/bingo-project/bingoctl/cmd/bingo@latest

# 创建新项目（只包含 apiserver）
bingo create github.com/myorg/myapp

# 或创建包含所有服务的项目
bingo create github.com/myorg/myapp --all

# 进入项目目录
cd myapp

# 启动依赖服务
docker-compose -f deployments/docker/docker-compose.yaml up -d

# 生成你的第一个模块 (如用户模块)
bingo make crud user

# 运行服务
make build
./myapp-apiserver
```

**创建项目的常用选项：**

```bash
# 创建并指定特定服务
bingo create myapp --services apiserver,admserver

# 添加额外的服务
bingo create myapp --add-service scheduler

# 排除某些服务
bingo create myapp --no-service bot

# 控制 git 初始化
bingo create myapp --init-git=false

# 使用特定的模板版本
bingo create myapp -r v1.2.3
```

详细的 bingo CLI 使用指南请查看 [使用 bingo CLI](https://bingoctl.dev/guide/using-bingo)。

### 方式二: 克隆 Bingo 仓库

如果你想基于 Bingo 源码进行开发:

#### 1. 克隆项目

```bash
git clone <repository-url>
cd bingo
```

#### 2. 配置环境

```bash
# 复制配置文件
cp configs/bingo-apiserver.example.yaml bingo-apiserver.yaml

# 根据实际环境修改配置
vim bingo-apiserver.yaml
```

#### 3. 启动依赖服务

```bash
# 使用 Docker Compose 启动 MySQL 和 Redis
docker-compose -f deployments/docker/docker-compose.yaml up -d mysql redis
```

#### 4. 数据库迁移

```bash
# 编译项目
make build

# 复制配置文件
cp configs/{app}ctl.example.yaml {app}ctl.yaml

# Build your app ctl
make build BINS="{app}ctl"

# 执行数据库迁移
./_output/platforms/{os}/{arch}/{app}ctl migrate up
```

#### 5. 启动服务

```bash
# 方式一:直接运行
make build
bingo-apiserver -c bingo-apiserver.yaml

# 方式二:开发模式(热重启)
cp .air.example.toml .air.toml
air
```

#### 6. 验证服务

```bash
# 检查服务状态
curl http://localhost:8080/health

# 访问 Swagger 文档
open http://localhost:8080/swagger/index.html
```

## 文档导航

### 📖 推荐学习路径

**初学者**：[什么是Bingo](https://bingoctl.dev/guide/what-is-bingo) → [快速开始](https://bingoctl.dev/guide/getting-started) → [项目结构](https://bingoctl.dev/guide/project-structure) → [开发第一个功能](https://bingoctl.dev/guide/first-feature)

**深入学习**：[整体架构](https://bingoctl.dev/essentials/architecture) → [分层架构详解](https://bingoctl.dev/essentials/layered-design) → [Store包设计](https://bingoctl.dev/essentials/store) → [开发规范](https://bingoctl.dev/development/standards)

**生产部署**：[Docker部署](https://bingoctl.dev/deployment/docker) → [微服务拆分](https://bingoctl.dev/advanced/microservices)

### 📚 新手入门

- [什么是 Bingo](https://bingoctl.dev/guide/what-is-bingo) - 了解 Bingo 的定位和特性
- [快速开始](https://bingoctl.dev/guide/getting-started) - 10 分钟快速启动项目
- [使用 bingo CLI](https://bingoctl.dev/guide/using-bingo) - CLI 工具完整指南
- [项目结构](https://bingoctl.dev/guide/project-structure) - 理解项目目录组织
- [开发第一个功能](https://bingoctl.dev/guide/first-feature) - 通过实例学习开发流程

### 🏗️ 核心概念

- [整体架构](https://bingoctl.dev/essentials/architecture) - 理解微服务架构设计
- [分层架构详解](https://bingoctl.dev/essentials/layered-design) - 掌握三层架构模式
- [Store 包设计](https://bingoctl.dev/essentials/store) - 数据访问层设计原理

### 💻 开发指南

- [开发规范](https://bingoctl.dev/development/standards) - 代码规范和最佳实践

### 🧩 组件参考

- [核心组件概览](https://bingoctl.dev/components/overview) - 了解所有可用组件

### 🚀 部署运维

- [Docker 部署](https://bingoctl.dev/deployment/docker) - 容器化部署指南

### 🔬 进阶主题

- [微服务拆分](https://bingoctl.dev/advanced/microservices) - 大型项目的微服务拆分

## 常用命令

```bash
# 开发相关
make build          # 编译所有服务
make run            # 运行服务(开发模式)
make test           # 运行单元测试
make cover          # 测试覆盖率报告

# 代码质量
make lint           # 代码检查
make format         # 代码格式化

# 代码生成
make swagger        # 生成 Swagger 文档
make protoc         # 编译 Protocol Buffers

# 部署相关
make image          # 构建 Docker 镜像

# 清理
make clean          # 清理构建产物
```

## 项目结构

```
bingo/
├── cmd/                    # 可执行程序入口
│   ├── bingo-apiserver/    # API 服务
│   ├── bingo-admserver/    # 管理服务
│   ├── bingo-scheduler/    # 调度服务
│   ├── bingo-bot/          # 机器人服务
│   └── bingoctl/           # CLI 工具
├── internal/               # 内部应用代码
│   ├── apiserver/          # API 服务实现
│   ├── admserver/          # 管理服务实现
│   └── pkg/                # 内部共享包
├── pkg/                    # 公共包
├── docs/                   # 项目文档
├── configs/                # 配置文件
├── deployments/            # 部署配置
└── scripts/                # 脚本工具
```

详细说明请查看 [项目结构文档](https://bingoctl.dev/guide/project-structure)。

## 贡献指南

欢迎提交 Issue 和 Pull Request!

### 开发流程

1. Fork 本仓库
2. 创建特性分支: `git checkout -b feature/amazing-feature`
3. 提交修改: `git commit -m 'feat: add amazing feature'`
4. 推送分支: `git push origin feature/amazing-feature`
5. 提交 Pull Request

### 代码审查

PR 需要通过:
- 代码规范检查(golangci-lint)
- 单元测试
- 至少一位 Maintainer 的审查

## 许可证

本项目采用 [Apache License 2.0](LICENSE) 开源许可证。

## 联系方式

如有问题或建议,请:
- 提交 Issue
- 发送邮件到项目维护者

---

**开始使用 Bingo,专注于你的业务逻辑,让脚手架处理其他一切!**
