# Bingo

一个开箱即用的 Go 语言中后台脚手架,基于微服务架构设计,让开发者只需关注业务开发。

## 项目定位

Bingo 是一个**生产级的 Go 中后台脚手架**,提供了完整的微服务架构、核心组件和最佳实践,帮助团队快速搭建可扩展的后端服务。

## 核心特性

### 架构层面
- **微服务架构**: 多服务独立部署,支持水平扩展
- **分层设计**: Controller → Biz → Store 清晰的三层架构
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

### 方式一: 使用 bingoctl 创建新项目 (推荐)

使用 [bingoctl](https://github.com/bingo-project/bingoctl) CLI 工具快速创建项目:

```bash
# 安装 bingoctl
go install github.com/bingo-project/bingoctl@latest

# 创建新项目
bingoctl create github.com/myorg/myapp

# 进入项目目录
cd myapp

# 启动依赖服务
docker-compose -f deployments/docker/docker-compose.yaml up -d

# 生成你的第一个模块 (如用户模块)
bingoctl make crud user

# 运行服务
make build
./myapp-apiserver
```

详细的 bingoctl 使用指南请查看 [使用 bingoctl](./docs/guide/using-bingoctl.md)。

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
cp configs/{app}-admserver.example.yaml {app}-admserver.yaml

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

### 📚 新手入门

- [什么是 Bingo](./docs/guide/what-is-bingo.md) - 了解 Bingo 的定位和特性
- [快速开始](./docs/guide/getting-started.md) - 10 分钟快速启动项目
- [使用 bingoctl](./docs/guide/using-bingoctl.md) - CLI 工具完整指南
- [项目结构](./docs/guide/project-structure.md) - 理解项目目录组织
- [开发第一个功能](./docs/guide/first-feature.md) - 通过实例学习开发流程

### 🏗️ 核心概念

- [整体架构](./docs/essentials/architecture.md) - 理解微服务架构设计
- [分层架构详解](./docs/essentials/layered-design.md) - 掌握三层架构模式
- [Store 包设计](./docs/essentials/store.md) - 数据访问层设计原理

### 💻 开发指南

- [开发规范](./docs/development/standards.md) - 代码规范和最佳实践

### 🧩 组件参考

- [核心组件概览](./docs/components/overview.md) - 了解所有可用组件

### 🚀 部署运维

- [Docker 部署](./docs/deployment/docker.md) - 容器化部署指南

### 🔬 进阶主题

- [微服务拆分](./docs/advanced/microservices.md) - 大型项目的微服务拆分

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

详细说明请查看 [项目结构文档](./docs/guide/project-structure.md)。

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

本项目采用 [MIT License](LICENSE) 开源许可证。

## 联系方式

如有问题或建议,请:
- 提交 Issue
- 发送邮件到项目维护者

---

**开始使用 Bingo,专注于你的业务逻辑,让脚手架处理其他一切!**
