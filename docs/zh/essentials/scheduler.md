# Scheduler 调度器

Bingo Scheduler 是基于 [Asynq](https://github.com/hibiken/asynq) 封装的任务调度服务，支持队列任务、静态周期任务和动态周期任务。

## 核心功能

Scheduler 提供三种类型的任务支持：

### 1. 队列任务（Queue Jobs）

用于即时或延迟执行的一次性任务，例如：
- 发送邮件
- 推送通知
- 数据处理
- 异步操作

### 2. 静态周期任务（Cron Jobs）

在代码中定义的周期性任务，适合：
- 固定的系统维护任务
- 定期数据统计
- 日志清理

### 3. 动态周期任务（Dynamic Cron Jobs）

存储在数据库的周期性任务，支持：
- 运行时动态添加/修改任务
- 无需重启服务
- 通过管理后台配置

## 快速开始

### 1. 启动 Scheduler 服务

```bash
# 使用默认配置
./bingo-scheduler

# 指定配置文件
./bingo-scheduler -c /path/to/bingo-scheduler.yaml
```

### 2. 配置文件

创建 `bingo-scheduler.yaml`：

```yaml
server:
  name: bingo-scheduler
  mode: release
  addr: :8080
  timezone: Asia/Shanghai

redis:
  host: redis:6379
  password: ""
  database: 1

mysql:
  host: mysql:3306
  username: root
  password: root
  database: bingo

log:
  level: info
  path: storage/log/scheduler.log

feature:
  queueDash: true  # 开启监控面板
```

### 3. 访问监控面板

如果启用了 `queueDash`，访问：

```
http://localhost:8080/queue
```

可以查看任务状态、队列情况和执行统计。

## 开发指南

### 添加队列任务

队列任务用于即时或延迟执行的一次性操作。

#### 第一步：定义任务类型和 Payload

在 `internal/pkg/task/types.go` 中定义：

```go
package task

const (
    EmailVerificationCode = "email:verification"
    UserDataExport        = "user:export"  // 新增任务类型
)

type UserDataExportPayload struct {
    UserID   int64
    Format   string // csv, json, xlsx
    Email    string
}
```

#### 第二步：实现处理函数

在 `internal/scheduler/job/` 目录创建处理文件，例如 `user_export.go`：

```go
package job

import (
    "context"
    "encoding/json"

    "github.com/hibiken/asynq"
    "github.com/bingo-project/component-base/log"

    "bingo/internal/pkg/task"
)

func HandleUserDataExport(ctx context.Context, t *asynq.Task) error {
    var payload task.UserDataExportPayload
    if err := json.Unmarshal(t.Payload(), &payload); err != nil {
        return err
    }

    log.Infow("Processing user data export",
        "user_id", payload.UserID,
        "format", payload.Format)

    // 业务逻辑
    // 1. 查询用户数据
    // 2. 导出为指定格式
    // 3. 发送邮件

    return nil
}
```

#### 第三步：注册任务

在 `internal/scheduler/job/registry.go` 中注册：

```go
package job

import (
    "github.com/hibiken/asynq"
    "bingo/internal/pkg/task"
)

func Register(mux *asynq.ServeMux) {
    mux.HandleFunc(task.EmailVerificationCode, HandleEmailVerificationTask)
    mux.HandleFunc(task.UserDataExport, HandleUserDataExport)  // 新增
}
```

#### 第四步：在业务代码中分发任务

```go
import "bingo/internal/pkg/task"

// 立即执行
task.T.Queue(ctx, task.UserDataExport, task.UserDataExportPayload{
    UserID: 123,
    Format: "csv",
    Email:  "user@example.com",
}).Dispatch()

// 延迟执行（10 分钟后）
task.T.Queue(ctx, task.UserDataExport, payload).Dispatch(
    asynq.ProcessIn(10 * time.Minute),
)

// 设置优先级和重试
task.T.Queue(ctx, task.UserDataExport, payload).Dispatch(
    asynq.Queue("critical"),       // 使用高优先级队列
    asynq.MaxRetry(3),              // 最多重试 3 次
    asynq.Timeout(5 * time.Minute), // 超时时间
)
```

### 添加静态周期任务

静态周期任务在代码中定义，适合固定的系统任务。

在 `internal/scheduler/scheduler/registry.go` 中注册：

```go
package scheduler

import (
    "github.com/hibiken/asynq"
    "github.com/bingo-project/component-base/log"
    "bingo/internal/pkg/facade"
)

func RegisterPeriodicTasks() {
    // 每天凌晨 2 点执行数据清理
    t := asynq.NewTask("task:daily-cleanup", nil)
    _, err := facade.Scheduler.Register("0 2 * * *", t)
    if err != nil {
        log.Fatalw("Failed to register task", "err", err)
    }

    // 每 5 分钟执行健康检查
    healthCheck := asynq.NewTask("task:health-check", nil)
    facade.Scheduler.Register("*/5 * * * *", healthCheck)
}
```

**常用 Cron 表达式：**

```
* * * * *        每分钟
0 * * * *        每小时
0 2 * * *        每天凌晨 2 点
0 9 * * 1        每周一早上 9 点
0 0 1 * *        每月 1 号凌晨
@hourly          每小时（等同于 0 * * * *）
@daily           每天凌晨（等同于 0 0 * * *）
@every 10s       每 10 秒
```

### 添加动态周期任务

动态任务存储在数据库，可以通过管理后台或 API 动态管理。

#### 数据库表结构

任务配置存储在 `sys_schedule` 表：

```sql
CREATE TABLE `sys_schedule` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL COMMENT '任务名称',
  `job` varchar(255) NOT NULL COMMENT '任务类型（唯一）',
  `spec` varchar(255) NOT NULL COMMENT 'Cron 表达式',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '状态：1-启用，2-禁用',
  `description` varchar(1000) NOT NULL COMMENT '任务描述',
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_job` (`job`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

#### 添加动态任务

通过管理后台或直接插入数据库：

```sql
INSERT INTO `sys_schedule` (name, job, spec, status, description)
VALUES ('每日统计报告', 'task:daily-stats', '0 1 * * *', 1, '生成每日数据统计报告');
```

动态任务会每 10 秒自动同步，无需重启服务。

## 监控和运维

### 查看任务状态

访问监控面板查看：
- 待执行任务数量
- 正在执行的任务
- 已完成任务统计
- 失败任务和错误信息

### 日志查看

```bash
# 实时日志
tail -f storage/log/scheduler.log

# 错误日志
grep "ERROR" storage/log/scheduler.log

# 特定任务日志
grep "task:daily-stats" storage/log/scheduler.log
```

### 常见问题

#### 1. 任务未执行

**检查项：**
- Redis 连接是否正常
- 时区配置是否正确（`server.timezone`）
- Cron 表达式是否有效
- 动态任务的 status 是否为 1（启用）

```bash
# 测试 Redis 连接
redis-cli -h redis -p 6379 ping

# 查看数据库任务配置
mysql -h mysql -u root -p bingo
SELECT * FROM sys_schedule WHERE status = 1;
```

#### 2. 任务执行失败

查看日志中的错误信息：

```bash
grep "ERROR" storage/log/scheduler.log | tail -20
```

常见原因：
- 数据库连接失败
- 依赖服务不可用
- 任务处理逻辑错误
- 任务超时

#### 3. 任务执行缓慢

优化方法：
- 增加 Worker 并发数（配置文件中修改）
- 优化任务处理逻辑
- 使用任务队列分片

## 最佳实践

### 1. 任务幂等性

确保任务可以安全重试：

```go
func HandleTask(ctx context.Context, t *asynq.Task) error {
    taskID := t.ResultWriter().TaskID()

    // 检查任务是否已执行
    if exists := checkTaskExecuted(taskID); exists {
        return nil  // 已执行，跳过
    }

    // 执行任务
    if err := doWork(); err != nil {
        return err
    }

    // 标记为已执行
    markTaskExecuted(taskID)
    return nil
}
```

### 2. 错误处理

```go
func HandleTask(ctx context.Context, t *asynq.Task) error {
    if err := doWork(); err != nil {
        // 记录详细错误日志
        log.Errorw("Task execution failed",
            "task_type", t.Type(),
            "payload", string(t.Payload()),
            "error", err)

        // 返回错误以触发重试
        return fmt.Errorf("task failed: %w", err)
    }

    return nil
}
```

### 3. 监控任务执行时间

```go
func HandleTask(ctx context.Context, t *asynq.Task) error {
    start := time.Now()
    defer func() {
        log.Infow("Task execution completed",
            "task_type", t.Type(),
            "duration", time.Since(start))
    }()

    return doWork()
}
```

### 4. 合理设置超时和重试

```go
// 分发任务时设置
task.T.Queue(ctx, task.UserDataExport, payload).Dispatch(
    asynq.MaxRetry(3),              // 最多重试 3 次
    asynq.Timeout(30*time.Second),  // 30 秒超时
    asynq.Retention(24*time.Hour),  // 保留任务结果 24 小时
)
```

## 与其他服务集成

### 发送邮件

```go
import "bingo/internal/pkg/facade"

func HandleDailyReport(ctx context.Context, t *asynq.Task) error {
    report := generateReport()

    err := facade.Mail.Send(
        "admin@example.com",
        "每日报告",
        report,
    )

    return err
}
```

### 访问数据库

```go
import "bingo/internal/pkg/store"

func HandleDataSync(ctx context.Context, t *asynq.Task) error {
    users, err := store.S.Users().List(ctx)
    if err != nil {
        return err
    }

    // 处理数据
    for _, user := range users {
        // ...
    }

    return nil
}
```

## 相关资源

- [Asynq 官方文档](https://github.com/hibiken/asynq) - 底层任务队列实现
- [Cron 表达式生成器](https://crontab.guru/) - 在线测试 Cron 表达式

## 下一步

- 学习 [Bot 服务](/essentials/bot) 如何接收定时通知
