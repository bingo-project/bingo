# 健康检查设计

## 概述

框架内置 K8s 风格健康检查端点，**运行在独立 HTTP 端口**，与业务 Server 解耦。

## 端点

```
GET /healthz   → Liveness（进程存活检查）
GET /readyz    → Readiness（服务就绪检查）
```

## 独立端口

健康检查运行在独立端口，与业务 Server（HTTP/gRPC/WebSocket）解耦：

```yaml
server:
  http:
    addr: ":8080"
  grpc:
    addr: ":9090"
  health:
    addr: ":8081"  # 健康检查独立端口
```

**设计理由**：
- 健康检查是基础设施级别，不应依赖业务 Server 是否启用
- 即使只有 gRPC Server，K8s 探针配置也统一指向 HTTP 端口
- 即使业务 Server panic/阻塞，健康检查端点仍然可用
- controller-runtime 也是这么做的（metrics server 独立端口）

## 固定行为

| 端点 | 行为 | 说明 |
|------|------|------|
| /healthz | 始终返回 200 | 能响应即存活 |
| /readyz | 就绪后 200，关闭时 503 | 框架内部状态控制 |

### 响应格式

```json
{"status": "ok"}
```

或关闭时：

```json
{"status": "shutting_down"}
```

## 生命周期中的行为

```
App 启动
    │
    ├── 健康检查端口立即启动
    │
    ├── /healthz 立即返回 200（进程存活）
    │
    ├── /readyz 返回 503（启动中）
    │
    │   ... Init/Register/Start 阶段 ...
    │
    ├── /readyz 返回 200（就绪，close Ready channel）
    │
    │   ... 运行中 ...
    │
    │   ctx 取消（用户通过 SetupSignalHandler 或手动）
    │
    └── /readyz 返回 503（从负载均衡摘除）
```

## Server 故障处理

如果业务 Server（HTTP/gRPC/WS）挂了：

| 情况 | 处理 |
|------|------|
| Runnable.Start() 返回 error | App.Run() 返回错误，进程退出 |
| Runnable goroutine panic | 进程崩溃 |

**结果**：进程退出 → 健康检查端点也没了 → K8s 检测到连接失败 → 重启 Pod

这符合设计决策 "Runnable 级让进程崩，Fail fast，K8s 会重启"。

## 为什么不提供自定义检查

### 常见错误做法

| 做法 | 问题 |
|------|------|
| Liveness 检查 DB | DB 短暂不可用 → 容器重启 → 雪崩 |
| Readiness 检查 DB | DB 短暂不可用 → 所有 Pod not ready → 服务完全不可用 |
| 检查外部 API | 同上，外部故障扩散到内部 |

### K8s 最佳实践

> "Having the liveness probe depend on external systems (like databases) is a bad practice."

- 连接池（gorm、go-redis）会自动重连
- 外部依赖故障应该由业务逻辑处理，而非健康检查
- 健康检查应该只反映"进程是否能工作"

### 框架立场

**不提供 AddHealthzCheck / AddReadyzCheck API。**

如果用户确实有特殊需求，可以自己注册 handler：

```go
app.HTTPEngine().GET("/my-check", myHandler)
```

## 决策记录

| 主题 | 决策 | 理由 |
|------|------|------|
| 端点暴露 | 独立 HTTP 端口 | 与业务 Server 解耦，始终可用 |
| 端点路径 | /healthz, /readyz | K8s 生态惯例 |
| 检查机制 | 不提供自定义，固定行为 | 自定义检查大多是错误做法 |
| Liveness | 始终 200 | 能响应即存活，不检查外部依赖 |
| Readiness | 框架状态控制 | 启动完成 200，关闭时 503 |
| Server 故障 | 进程退出 | Fail fast，K8s 重启 Pod |
