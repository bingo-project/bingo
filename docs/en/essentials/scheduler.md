# Scheduler

Bingo Scheduler is a distributed task scheduling service based on [Asynq](https://github.com/hibiken/asynq), supporting scheduled tasks, periodic tasks, and dynamic task management.

## Core Features

- **Scheduled Task Execution** - Support for cron expression-based task scheduling
- **Dynamic Task Management** - Add, modify, and delete tasks at runtime
- **Distributed Architecture** - Redis-based distributed task queue
- **High Availability** - Multi-instance deployment with automatic failover
- **Task Monitoring** - Built-in web monitoring dashboard (optional)

## Architecture

The Scheduler service consists of two core components:

### 1. Static Task Scheduler

Registers and schedules predefined periodic tasks.

```go
// Initialize scheduler
facade.Scheduler = asynq.NewScheduler(opt, &asynq.SchedulerOpts{
    Location: location,
})
```

### 2. Dynamic Task Manager (PeriodicTaskManager)

Dynamically loads task configurations from the database, allowing task updates without service restart.

```go
// Dynamic task manager
facade.TaskManager, err = asynq.NewPeriodicTaskManager(
    asynq.PeriodicTaskManagerOpts{
        RedisConnOpt:               opt,
        PeriodicTaskConfigProvider: syscfg.NewSchedule(store.S),
        SyncInterval:               time.Second * 10,  // Sync every 10 seconds
    })
```

## Quick Start

### 1. Configuration

Create `bingo-scheduler.yaml` configuration file:

```yaml
# Scheduler Server
server:
  name: bingo-scheduler
  mode: release
  addr: :8080
  timezone: Asia/Shanghai  # Timezone setting
  key: your-secret-key

# Redis Configuration (Task queue storage)
redis:
  host: redis:6379
  password: ""
  database: 1

# MySQL Configuration (Dynamic task configuration storage)
mysql:
  host: mysql:3306
  username: root
  password: root
  database: bingo
  maxIdleConnections: 100
  maxOpenConnections: 100
  maxConnectionLifeTime: 10s
  logLevel: 4

# Logging Configuration
log:
  level: info
  days: 7
  format: console
  console: true
  maxSize: 100
  compress: true
  path: storage/log/scheduler.log

# Feature Flags
feature:
  queueDash: true  # Enable queue monitoring dashboard
```

### 2. Start Service

```bash
# Use default configuration file
./bingo-scheduler

# Specify configuration file
./bingo-scheduler -c /path/to/bingo-scheduler.yaml
```

### 3. Access Monitoring Dashboard

If `queueDash` is enabled, access the monitoring dashboard at:

```
http://localhost:8080/queue
```

## Task Types

### Static Tasks (Code-Defined)

Register periodic tasks in code:

```go
import (
    "github.com/hibiken/asynq"
    "bingo/internal/pkg/facade"
)

// Register a task that runs daily at 2 AM
facade.Scheduler.Register(
    "0 2 * * *",  // Cron expression
    asynq.NewTask("task:daily-report", nil),
)
```

**Common Cron Expressions:**

```
# Every minute
* * * * *

# Every hour
0 * * * *

# Daily at 2 AM
0 2 * * *

# Every Monday at 9 AM
0 9 * * 1

# First day of every month at midnight
0 0 1 * *
```

### Dynamic Tasks (Database-Configured)

Dynamic tasks are stored in the `sys_crontab` table and can be managed through the admin dashboard:

```sql
CREATE TABLE `sys_crontab` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(100) NOT NULL COMMENT 'Task name',
  `type` varchar(50) NOT NULL COMMENT 'Task type',
  `spec` varchar(100) NOT NULL COMMENT 'Cron expression',
  `payload` text COMMENT 'Task parameters (JSON)',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT 'Status: 1-enabled, 0-disabled',
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

**Example: Add Dynamic Task**

```sql
INSERT INTO `sys_crontab`
(`name`, `type`, `spec`, `payload`, `status`)
VALUES
('Daily Statistics', 'task:daily-stats', '0 1 * * *', '{"date":"today"}', 1);
```

Dynamic tasks are automatically synced to the scheduler every 10 seconds without requiring a service restart.

## Task Handlers

### 1. Define Task Handler

```go
package tasks

import (
    "context"
    "encoding/json"

    "github.com/hibiken/asynq"
    "github.com/bingo-project/component-base/log"
)

// DailyReportPayload task parameters
type DailyReportPayload struct {
    Date string `json:"date"`
}

// HandleDailyReport processes daily report task
func HandleDailyReport(ctx context.Context, t *asynq.Task) error {
    var p DailyReportPayload
    if err := json.Unmarshal(t.Payload(), &p); err != nil {
        return err
    }

    log.Infof("Generating daily report: %s", p.Date)

    // Business logic
    // ...

    return nil
}
```

### 2. Register Handler

```go
import (
    "github.com/hibiken/asynq"
    "bingo/internal/scheduler/tasks"
)

func registerHandlers(srv *asynq.Server) {
    mux := asynq.NewServeMux()

    // Register task handlers
    mux.HandleFunc("task:daily-report", tasks.HandleDailyReport)
    mux.HandleFunc("task:daily-stats", tasks.HandleDailyStats)

    // Start worker
    if err := srv.Run(mux); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}
```

## Advanced Usage

### Task Priority

```go
// High priority task
task := asynq.NewTask("task:important", payload, asynq.MaxRetry(3))
facade.Scheduler.Register("*/5 * * * *", task)
```

### Task Retry

```go
// Set maximum retry count
task := asynq.NewTask(
    "task:with-retry",
    payload,
    asynq.MaxRetry(5),              // Maximum 5 retries
    asynq.Timeout(30*time.Second),  // Timeout duration
)
```

### Task Timeout

```go
func HandleTask(ctx context.Context, t *asynq.Task) error {
    // Check if context is cancelled or timed out
    select {
    case <-ctx.Done():
        return ctx.Err()  // Task was cancelled or timed out
    default:
        // Execute task
    }

    return nil
}
```

## Monitoring and Operations

### View Task Status

Access the monitoring dashboard to view:
- Pending tasks count
- Running tasks
- Completed tasks
- Failed tasks and retry counts

### View Logs

```bash
# View real-time logs
tail -f storage/log/scheduler.log

# View error logs
grep "ERROR" storage/log/scheduler.log
```

### Common Issues

#### 1. Tasks Not Executing

**Check:**
- Redis connection status
- Timezone configuration
- Cron expression validity
- Task status (enabled/disabled)

```bash
# Test Redis connection
redis-cli -h redis -p 6379 ping
```

#### 2. Dynamic Tasks Not Working

**Causes:**
- Database connection failure
- `sys_crontab` table doesn't exist
- Sync interval not reached (default 10 seconds)

**Solutions:**
```bash
# Check database connection
mysql -h mysql -u root -p bingo

# Manually trigger sync (restart service)
pkill -USR1 bingo-scheduler
```

#### 3. Slow Task Execution

**Optimization Tips:**
- Increase worker concurrency
- Optimize task handler logic
- Use task queue sharding

```go
srv := asynq.NewServer(
    opt,
    asynq.Config{
        Concurrency: 20,  // Increase concurrency
    },
)
```

## Best Practices

### 1. Task Idempotency

Ensure tasks can be safely retried:

```go
func HandleTask(ctx context.Context, t *asynq.Task) error {
    // Use unique identifier to check if task was already executed
    taskID := t.ResultWriter().TaskID()

    if exists := checkTaskExecuted(taskID); exists {
        return nil  // Already executed, skip
    }

    // Execute task
    if err := doWork(); err != nil {
        return err
    }

    // Mark as executed
    markTaskExecuted(taskID)
    return nil
}
```

### 2. Error Handling

```go
func HandleTask(ctx context.Context, t *asynq.Task) error {
    if err := doWork(); err != nil {
        // Log detailed error information
        log.Errorf("Task execution failed: %v, payload: %s", err, t.Payload())

        // Return error to trigger retry
        return fmt.Errorf("task failed: %w", err)
    }

    return nil
}
```

### 3. Task Monitoring

```go
// Track task execution time
func HandleTask(ctx context.Context, t *asynq.Task) error {
    start := time.Now()
    defer func() {
        log.Infof("Task execution time: %v", time.Since(start))
    }()

    // Execute task
    return doWork()
}
```

### 4. Proper Timezone Configuration

```yaml
# Set timezone in configuration file
server:
  timezone: Asia/Shanghai  # Use local timezone
```

```go
// Load timezone in code
location, err := time.LoadLocation(facade.Config.Server.Timezone)
if err != nil {
    log.Fatalf("Failed to load timezone: %v", err)
}
```

## Integration with Other Services

### Send Email Notifications

```go
import "bingo/internal/pkg/mail"

func HandleDailyReport(ctx context.Context, t *asynq.Task) error {
    // Generate report
    report := generateReport()

    // Send email
    err := mail.Send(mail.Message{
        To:      []string{"admin@example.com"},
        Subject: "Daily Report",
        Body:    report,
    })

    return err
}
```

### Call API Server

```go
import "bingo/internal/apiserver/biz"

func HandleDataSync(ctx context.Context, t *asynq.Task) error {
    // Access data through Store
    users, err := store.S.Users().List(ctx)
    if err != nil {
        return err
    }

    // Process data
    for _, user := range users {
        // ...
    }

    return nil
}
```

## Related Resources

- [Asynq Official Documentation](https://github.com/hibiken/asynq)
- [Cron Expression Generator](https://crontab.guru/)
- [Task Queue Best Practices](/components/overview#asynchronous-tasks)

## Next Steps

- Learn how [Admin Server](/essentials/admserver) manages dynamic tasks
- Explore [Task Queue Component](/components/overview#asynchronous-tasks) usage
