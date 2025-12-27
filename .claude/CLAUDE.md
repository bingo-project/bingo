# Bingo 项目指南

## 必读文档

开发前必须阅读 [docs/guides/CONVENTIONS.md](../docs/guides/CONVENTIONS.md)，了解：

- 三层架构原则（Handler → Biz → Store）
- 文件规范（ABOUTME 注释、目录结构）
- 命名规范（包名、文件名、接口名）
- 错误处理（统一错误码、core.Response）
- 日志规范（结构化日志）
- 测试规范（分层测试策略）
- 生成代码检查清单

## 技术栈

- Go 1.24+
- Gin（HTTP 框架）
- GORM（ORM）
- Redis（缓存/队列）
- Asynq（任务调度）

## 开发偏好

- **执行计划时**：使用 Subagent-Driven 方式执行计划, 按阶段（而非每个任务）进行代码审查，提高效率
- **Worktree 目录**：使用 `.worktrees/` 存放隔离开发环境
