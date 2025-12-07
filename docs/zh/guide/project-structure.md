---
title: 项目结构 - Bingo Go 微服务项目目录组织
description: 了解 Bingo Go 微服务框架的项目目录结构，包括 cmd、internal、pkg 等目录的组织方式和各文件的职责说明。
---

# 项目结构

本文介绍 Bingo 项目的目录组织和文件职责。

## 目录概览

```
bingo/
├── cmd/                            # 可执行程序入口
├── internal/                       # 内部应用代码(不可被外部导入)
├── pkg/                            # 可被外部导入的公共包
├── api/                            # API 文档
├── configs/                        # 配置文件
├── deployments/                    # 部署配置
├── build/                          # 构建脚本
├── scripts/                        # 开发脚本
├── storage/                        # 运行时数据
├── Makefile                        # 构建配置
├── go.mod                          # Go 模块定义
└── README.md                       # 项目文档
```

## cmd/ - 可执行程序入口

```
cmd/
├── bingo-apiserver/            # API 服务入口
│   └── main.go                 # 主函数
├── bingo-admserver/            # 管理服务入口
├── bingo-scheduler/            # 调度服务入口
├── bingo-bot/                  # 机器人服务入口
└── bingoctl/                   # CLI 工具入口
```

每个服务一个目录,包含独立的 `main.go`。遵循 Go 标准项目布局。

## internal/ - 内部应用代码

内部代码,不会被外部项目导入。这是 Go 的包可见性特性,确保内部实现不被外部依赖。

### 服务实现标准结构

以 `apiserver` 为例:

```
internal/apiserver/
├── app.go                      # 应用初始化
├── run.go                      # 服务启动逻辑
├── biz/                        # 业务逻辑层
│   ├── auth/                   # 认证业务
│   ├── user/                   # 用户业务
│   └── ...                     # 其他业务模块
├── handler/                    # HTTP Handler 层
│   └── http/                   # HTTP 处理器
│       ├── auth/               # 认证相关接口
│       ├── user/               # 用户相关接口
│       └── ...
├── store/                      # 数据访问层
│   ├── user.go                 # 用户数据访问
│   └── ...
├── router/                     # 路由定义
│   └── router.go
├── middleware/                 # 中间件
│   ├── authn.go                # 认证中间件
│   ├── authz.go                # 授权中间件
│   └── ...
└── grpc/                       # gRPC 服务实现
```

**每个服务都遵循相同的结构**:
- `app.go` / `run.go`: 应用初始化和启动逻辑
- `biz/`: 业务逻辑层,处理业务规则
- `handler/`: HTTP 处理器,负责请求响应
- `store/`: 数据访问层,封装数据库操作
- `router/`: 路由配置
- `middleware/`: 中间件
- `grpc/`: gRPC 服务实现

### internal/pkg/ - 内部共享包

```
internal/pkg/
├── bootstrap/              # 应用启动引导
├── config/                 # 配置定义
├── model/                  # 数据模型
├── logger/                 # 日志组件
├── db/                     # 数据库组件
├── auth/                   # 认证组件
├── util/                   # 工具函数
└── ...
```

被多个内部服务使用,但不对外暴露。

## pkg/ - 公共包

```
pkg/
├── api/                        # API 定义
├── proto/                      # Protocol Buffer 定义
└── ...
```

可以被外部项目导入的公共包。如果你的项目需要提供 SDK,可以放在这里。

## api/ - API 文档

```
api/
├── swagger/                    # Swagger 文档
└── openapi/                    # OpenAPI 规范
```

## configs/ - 配置文件

```
configs/
├── bingo-apiserver.example.yaml
├── bingo-admserver.example.yaml
└── ...
```

各服务的配置文件模板。

## deployments/ - 部署配置

```
deployments/
└── docker/
    └── docker-compose.yaml
```

Docker、Kubernetes 等部署配置。

## build/ - 构建脚本

```
build/
├── docker/                     # Dockerfile
└── scripts/                    # 构建脚本
```

## scripts/ - 开发脚本

```
scripts/
└── make-rules/                 # Makefile 规则
```

Makefile 辅助脚本和规则。

## storage/ - 运行时数据

```
storage/
├── log/                        # 日志文件
└── public/                     # 静态资源
```

运行时生成的数据,不纳入版本控制。

## 为什么这样组织?

1. **符合 Go 标准**: 遵循 [Go 项目布局标准](https://github.com/golang-standards/project-layout)
2. **职责清晰**: 每个目录有明确的职责边界
3. **可维护性**: 新成员快速理解项目结构
4. **可扩展性**: 添加新服务只需复制标准结构
5. **包可见性**: `internal/` 确保内部实现不被外部依赖

## 下一步

- [开发第一个功能](./first-feature.md) - 通过实例理解各层职责
