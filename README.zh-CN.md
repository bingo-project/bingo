[English](README.md) | 中文

# Bingo - 生产级 Go 微服务脚手架

> 一个开箱即用的 Go/Golang 微服务脚手架框架，基于微服务架构设计，让开发者只需关注业务开发。

[![Go Version](https://img.shields.io/badge/Go-1.23%2B-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

## 文档

**官方文档**: [bingoctl.dev](https://bingoctl.dev)

- [快速开始](https://bingoctl.dev/guide/getting-started) - 10 分钟快速上手
- [什么是 Bingo](https://bingoctl.dev/guide/what-is-bingo) - 了解核心特性
- [整体架构](https://bingoctl.dev/essentials/architecture) - 微服务架构设计
- [使用 Bingo CLI](https://bingoctl.dev/guide/using-bingo) - CLI 工具指南
- [English Documentation](https://bingoctl.dev/en/) - English version

## 项目定位

**Bingo** 是一个**生产级的 Go/Golang 微服务脚手架框架**，提供完整的：
- 微服务架构设计（Controller → Biz → Store 三层架构）
- 核心组件预集成（Gin、GORM、Redis、Asynq、Casbin）
- 工程化能力（代码生成、热重启、Docker 支持）
- 生产级特性（日志、监控、链路追踪）
- 最佳实践和完整文档

**适用场景**: 中后台系统、微服务项目、RESTful API、gRPC 服务

**相关项目**: [Bingo CLI](https://github.com/bingo-project/bingoctl) - Bingo 项目脚手架工具

## 核心特性

### 架构层面
- **微服务架构**: 多服务独立部署，支持水平扩展
- **分层设计**: Controller → Biz → Store 清晰的三层架构
- **通用数据层**: 基于泛型的 Store[T] 设计，减少重复代码
- **依赖注入**: 基于接口编程，易于测试和扩展
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

```bash
# 安装 Bingo CLI
go install github.com/bingo-project/bingoctl/cmd/bingo@latest

# 创建新项目
bingo create github.com/myorg/myapp

# 进入项目目录
cd myapp

# 构建并运行
make build && ./_output/<os>/<arch>/myapp-apiserver
```

详细的安装和配置说明，请查看[快速开始](https://bingoctl.dev/guide/getting-started)。

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
- 代码规范检查 (golangci-lint)
- 单元测试
- 至少一位 Maintainer 的审查

## 许可证

本项目采用 [Apache License 2.0](LICENSE) 开源许可证。

## 联系方式

如有问题或建议，请:
- 提交 Issue
- 发送邮件到项目维护者

---

**开始使用 Bingo，专注于你的业务逻辑，让脚手架处理其他一切!**
