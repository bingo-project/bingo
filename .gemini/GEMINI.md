# Bingo 项目指南

## 必读文档

制定实现计划、开发或 Review 代码前必须阅读 [docs/guides/CONVENTIONS.md](../docs/guides/CONVENTIONS.md)，了解：

- 三层架构原则（Handler → Biz → Store）
- 文件规范（ABOUTME 注释、目录结构）
- 命名规范（包名、文件名、接口名）
- 错误处理（统一错误码、core.Response）
- 日志规范（结构化日志）
- 测试规范（分层测试策略）
- 生成代码检查清单
- 构建规范

## 技术栈

- Go 1.24+
- Gin（HTTP 框架）
- GORM（ORM）
- Redis（缓存/队列）
- Asynq（任务调度）

## 构建规范

```bash
# 代码修改后重新构建（不要用 go build ./...）
make build

# 修改 API 参数定义后（pkg/api/ 下的结构体）
make swag   # 先更新 Swagger 文档
make build  # 再构建

# commit 前必须执行
make lint
```

## 数据库命令

> **注意**：`bingo` 是全局安装的 CLI 工具，不是项目内的命令。

```bash
bingo migrate up        # 执行迁移
bingo migrate rollback  # 回滚上一次迁移
bingo migrate refresh   # 重置并重新迁移（开发环境）
bingo db seed           # 执行所有 seeder
```

## 常用命令

```bash
make build                    # 编译所有服务
make build BINS="svc1 svc2"   # 编译指定服务
make test                     # 测试
make lint                     # 代码检查
make swag                     # 生成 Swagger 文档
```

## 开发偏好

- **执行计划时**：使用 Subagent-Driven 方式执行计划, 按阶段（而非每个任务）进行代码审查，提高效率
- **Worktree 目录**：使用 `.worktrees/` 存放隔离开发环境
