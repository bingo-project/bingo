# Scheduler 调度器

Bingo Scheduler 是一个基于 [Asynq](https://github.com/hibiken/asynq) 的分布式任务调度服务，支持定时任务、周期任务和动态任务管理。

## 核心特性

- **定时任务调度** - 支持 Cron 表达式配置定时任务
- **动态任务管理** - 支持在运行时动态添加、修改、删除任务
- **分布式架构** - 基于 Redis 实现分布式任务队列
- **高可用性** - 支持多实例部署，自动故障转移
- **任务监控** - 内置 Web 监控面板（可选）

## 架构设计

Scheduler 服务由两个核心组件组成：

### 1. 静态任务调度器（Scheduler）

用于注册和调度预定义的周期性任务。

```go
// 初始化调度器
facade.Scheduler = asynq.NewScheduler(opt, &asynq.SchedulerOpts{
    Location: location,
})
```

### 2. 动态任务管理器（PeriodicTaskManager）

支持从数据库动态加载任务配置，无需重启服务即可更新任务。

```go
// 动态任务管理器
facade.TaskManager, err = asynq.NewPeriodicTaskManager(
    asynq.PeriodicTaskManagerOpts{
        RedisConnOpt:               opt,
        PeriodicTaskConfigProvider: syscfg.NewSchedule(store.S),
        SyncInterval:               time.Second * 10,  // 每 10 秒同步一次
    })
```

## 快速开始

### 1. 配置文件

创建 `bingo-scheduler.yaml` 配置文件：

```yaml
# Scheduler Server
server:
  name: bingo-scheduler
  mode: release
  addr: :8080
  timezone: Asia/Shanghai  # 时区设置
  key: your-secret-key

# Redis 配置（任务队列存储）
redis:
  host: redis:6379
  password: ""
  database: 1

# MySQL 配置（动态任务配置存储）
mysql:
  host: mysql:3306
  username: root
  password: root
  database: bingo
  maxIdleConnections: 100
  maxOpenConnections: 100
  maxConnectionLifeTime: 10s
  logLevel: 4

# 日志配置
log:
  level: info
  days: 7
  format: console
  console: true
  maxSize: 100
  compress: true
  path: storage/log/scheduler.log

# 功能开关
feature:
  queueDash: true  # 开启队列监控面板
```

### 2. 启动服务

```bash
# 使用默认配置文件
./bingo-scheduler

# 指定配置文件
./bingo-scheduler -c /path/to/bingo-scheduler.yaml
```

### 3. 访问监控面板

如果启用了 `queueDash`，可以通过以下地址访问监控面板：

```
http://localhost:8080/queue
```

## 任务类型

### 静态任务（代码中定义）

在代码中注册周期性任务：

```go
import (
    "github.com/hibiken/asynq"
    "bingo/internal/pkg/facade"
)

// 注册每天凌晨 2 点执行的任务
facade.Scheduler.Register(
    "0 2 * * *",  // Cron 表达式
    asynq.NewTask("task:daily-report", nil),
)
```

**常用 Cron 表达式：**

```
# 每分钟执行
* * * * *

# 每小时执行
0 * * * *

# 每天凌晨 2 点执行
0 2 * * *

# 每周一早上 9 点执行
0 9 * * 1

# 每月 1 号凌晨执行
0 0 1 * *
```

### 动态任务（数据库配置）

动态任务存储在 `sys_crontab` 表中，支持通过管理后台进行 CRUD 操作：

```sql
CREATE TABLE `sys_crontab` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(100) NOT NULL COMMENT '任务名称',
  `type` varchar(50) NOT NULL COMMENT '任务类型',
  `spec` varchar(100) NOT NULL COMMENT 'Cron 表达式',
  `payload` text COMMENT '任务参数（JSON）',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '状态：1-启用，0-禁用',
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

**示例：添加动态任务**

```sql
INSERT INTO `sys_crontab`
(`name`, `type`, `spec`, `payload`, `status`)
VALUES
('每日数据统计', 'task:daily-stats', '0 1 * * *', '{"date":"today"}', 1);
```

动态任务会每 10 秒自动同步到调度器，无需重启服务。

## 任务处理器

### 1. 定义任务处理器

```go
package tasks

import (
    "context"
    "encoding/json"

    "github.com/hibiken/asynq"
    "github.com/bingo-project/component-base/log"
)

// DailyReportPayload 任务参数
type DailyReportPayload struct {
    Date string `json:"date"`
}

// HandleDailyReport 处理每日报告任务
func HandleDailyReport(ctx context.Context, t *asynq.Task) error {
    var p DailyReportPayload
    if err := json.Unmarshal(t.Payload(), &p); err != nil {
        return err
    }

    log.Infof("生成每日报告: %s", p.Date)

    // 业务逻辑
    // ...

    return nil
}
```

### 2. 注册处理器

```go
import (
    "github.com/hibiken/asynq"
    "bingo/internal/scheduler/tasks"
)

func registerHandlers(srv *asynq.Server) {
    mux := asynq.NewServeMux()

    // 注册任务处理器
    mux.HandleFunc("task:daily-report", tasks.HandleDailyReport)
    mux.HandleFunc("task:daily-stats", tasks.HandleDailyStats)

    // 启动 Worker
    if err := srv.Run(mux); err != nil {
        log.Fatalf("服务器启动失败: %v", err)
    }
}
```

## 高级用法

### 任务优先级

```go
// 高优先级任务
task := asynq.NewTask("task:important", payload, asynq.MaxRetry(3))
facade.Scheduler.Register("*/5 * * * *", task)
```

### 任务重试

```go
// 设置最大重试次数
task := asynq.NewTask(
    "task:with-retry",
    payload,
    asynq.MaxRetry(5),      // 最多重试 5 次
    asynq.Timeout(30*time.Second),  // 超时时间
)
```

### 任务超时

```go
func HandleTask(ctx context.Context, t *asynq.Task) error {
    // 检查上下文是否超时
    select {
    case <-ctx.Done():
        return ctx.Err()  // 任务被取消或超时
    default:
        // 执行任务
    }

    return nil
}
```

## 监控和运维

### 查看任务状态

访问监控面板查看：
- 待执行任务数量
- 正在执行任务
- 已完成任务
- 失败任务及重试次数

### 日志查看

```bash
# 查看实时日志
tail -f storage/log/scheduler.log

# 查看错误日志
grep "ERROR" storage/log/scheduler.log
```

### 常见问题

#### 1. 任务未执行

**检查项：**
- Redis 连接是否正常
- 时区配置是否正确
- Cron 表达式是否有效
- 任务状态是否为启用

```bash
# 测试 Redis 连接
redis-cli -h redis -p 6379 ping
```

#### 2. 动态任务不生效

**原因：**
- 数据库连接失败
- `sys_crontab` 表不存在
- 同步间隔未到（默认 10 秒）

**解决方案：**
```bash
# 检查数据库连接
mysql -h mysql -u root -p bingo

# 手动触发同步（重启服务）
pkill -USR1 bingo-scheduler
```

#### 3. 任务执行缓慢

**优化建议：**
- 增加 Worker 并发数
- 优化任务处理逻辑
- 使用任务队列分片

```go
srv := asynq.NewServer(
    opt,
    asynq.Config{
        Concurrency: 20,  // 增加并发数
    },
)
```

## 最佳实践

### 1. 任务幂等性

确保任务可以安全重试：

```go
func HandleTask(ctx context.Context, t *asynq.Task) error {
    // 使用唯一标识符检查任务是否已执行
    taskID := t.ResultWriter().TaskID()

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
        log.Errorf("任务执行失败: %v, payload: %s", err, t.Payload())

        // 返回错误以触发重试
        return fmt.Errorf("任务失败: %w", err)
    }

    return nil
}
```

### 3. 任务监控

```go
// 记录任务执行时间
func HandleTask(ctx context.Context, t *asynq.Task) error {
    start := time.Now()
    defer func() {
        log.Infof("任务执行耗时: %v", time.Since(start))
    }()

    // 执行任务
    return doWork()
}
```

### 4. 合理设置时区

```yaml
# 配置文件中设置时区
server:
  timezone: Asia/Shanghai  # 使用本地时区
```

```go
// 代码中加载时区
location, err := time.LoadLocation(facade.Config.Server.Timezone)
if err != nil {
    log.Fatalf("时区加载失败: %v", err)
}
```

## 与其他服务集成

### 发送邮件通知

```go
import "bingo/internal/pkg/mail"

func HandleDailyReport(ctx context.Context, t *asynq.Task) error {
    // 生成报告
    report := generateReport()

    // 发送邮件
    err := mail.Send(mail.Message{
        To:      []string{"admin@example.com"},
        Subject: "每日报告",
        Body:    report,
    })

    return err
}
```

### 调用 API Server

```go
import "bingo/internal/apiserver/biz"

func HandleDataSync(ctx context.Context, t *asynq.Task) error {
    // 通过 Store 访问数据
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

- [Asynq 官方文档](https://github.com/hibiken/asynq)
- [Cron 表达式生成器](https://crontab.guru/)
- [任务队列最佳实践](/components/overview#任务队列)

## 下一步

- 了解 [Admin Server](/essentials/admserver) 如何管理动态任务
- 学习 [任务队列组件](/components/overview#异步任务) 的使用
