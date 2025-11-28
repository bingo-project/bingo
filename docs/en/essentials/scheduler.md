# Scheduler

Bingo Scheduler is a task scheduling service built on [Asynq](https://github.com/hibiken/asynq), supporting queue jobs, static periodic tasks, and dynamic periodic tasks.

## Core Features

Scheduler provides three types of task support:

### 1. Queue Jobs

One-time tasks for immediate or delayed execution, such as:
- Sending emails
- Push notifications
- Data processing
- Asynchronous operations

### 2. Static Periodic Tasks (Cron Jobs)

Periodic tasks defined in code, suitable for:
- Fixed system maintenance tasks
- Regular data statistics
- Log cleanup

### 3. Dynamic Periodic Tasks

Periodic tasks stored in database, supporting:
- Runtime task addition/modification
- No service restart required
- Configuration via admin dashboard

## Quick Start

### 1. Start Scheduler Service

```bash
# Use default configuration
./bingo-scheduler

# Specify configuration file
./bingo-scheduler -c /path/to/bingo-scheduler.yaml
```

### 2. Configuration

Create `bingo-scheduler.yaml`:

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
  queueDash: true  # Enable monitoring dashboard
```

### 3. Access Monitoring Dashboard

If `queueDash` is enabled, visit:

```
http://localhost:8080/queue
```

View task status, queue information, and execution statistics.

## Development Guide

### Adding Queue Jobs

Queue jobs are for one-time immediate or delayed execution.

#### Step 1: Define Task Type and Payload

Define in `internal/pkg/task/types.go`:

```go
package task

const (
    EmailVerificationCode = "email:verification"
    UserDataExport        = "user:export"  // New task type
)

type UserDataExportPayload struct {
    UserID   int64
    Format   string // csv, json, xlsx
    Email    string
}
```

#### Step 2: Implement Handler

Create handler file in `internal/scheduler/job/`, e.g., `user_export.go`:

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

    // Business logic
    // 1. Query user data
    // 2. Export to specified format
    // 3. Send email

    return nil
}
```

#### Step 3: Register Task

Register in `internal/scheduler/job/registry.go`:

```go
package job

import (
    "github.com/hibiken/asynq"
    "bingo/internal/pkg/task"
)

func Register(mux *asynq.ServeMux) {
    mux.HandleFunc(task.EmailVerificationCode, HandleEmailVerificationTask)
    mux.HandleFunc(task.UserDataExport, HandleUserDataExport)  // Add new
}
```

#### Step 4: Dispatch in Business Code

```go
import "bingo/internal/pkg/task"

// Execute immediately
task.T.Queue(ctx, task.UserDataExport, task.UserDataExportPayload{
    UserID: 123,
    Format: "csv",
    Email:  "user@example.com",
}).Dispatch()

// Delayed execution (after 10 minutes)
task.T.Queue(ctx, task.UserDataExport, payload).Dispatch(
    asynq.ProcessIn(10 * time.Minute),
)

// Set priority and retry
task.T.Queue(ctx, task.UserDataExport, payload).Dispatch(
    asynq.Queue("critical"),       // Use high-priority queue
    asynq.MaxRetry(3),              // Max 3 retries
    asynq.Timeout(5 * time.Minute), // Timeout duration
)
```

### Adding Static Periodic Tasks

Static periodic tasks are defined in code, suitable for fixed system tasks.

Register in `internal/scheduler/scheduler/registry.go`:

```go
package scheduler

import (
    "github.com/hibiken/asynq"
    "github.com/bingo-project/component-base/log"
    "bingo/internal/pkg/facade"
)

func RegisterPeriodicTasks() {
    // Daily cleanup at 2 AM
    t := asynq.NewTask("task:daily-cleanup", nil)
    _, err := facade.Scheduler.Register("0 2 * * *", t)
    if err != nil {
        log.Fatalw("Failed to register task", "err", err)
    }

    // Health check every 5 minutes
    healthCheck := asynq.NewTask("task:health-check", nil)
    facade.Scheduler.Register("*/5 * * * *", healthCheck)
}
```

**Common Cron Expressions:**

```
* * * * *        Every minute
0 * * * *        Every hour
0 2 * * *        Daily at 2 AM
0 9 * * 1        Every Monday at 9 AM
0 0 1 * *        First day of month at midnight
@hourly          Every hour (same as 0 * * * *)
@daily           Daily at midnight (same as 0 0 * * *)
@every 10s       Every 10 seconds
```

### Adding Dynamic Periodic Tasks

Dynamic tasks are stored in database and can be managed via admin dashboard or API.

#### Database Schema

Task configurations are stored in `sys_schedule` table:

```sql
CREATE TABLE `sys_schedule` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL COMMENT 'Task name',
  `job` varchar(255) NOT NULL COMMENT 'Task type (unique)',
  `spec` varchar(255) NOT NULL COMMENT 'Cron expression',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT 'Status: 1-enabled, 2-disabled',
  `description` varchar(1000) NOT NULL COMMENT 'Task description',
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_job` (`job`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

#### Adding Dynamic Tasks

Via admin dashboard or direct database insertion:

```sql
INSERT INTO `sys_schedule` (name, job, spec, status, description)
VALUES ('Daily Statistics Report', 'task:daily-stats', '0 1 * * *', 1, 'Generate daily data statistics report');
```

Dynamic tasks are automatically synced every 10 seconds without service restart.

## Monitoring and Operations

### View Task Status

Access monitoring dashboard to view:
- Pending task count
- Running tasks
- Completed task statistics
- Failed tasks and error messages

### View Logs

```bash
# Real-time logs
tail -f storage/log/scheduler.log

# Error logs
grep "ERROR" storage/log/scheduler.log

# Specific task logs
grep "task:daily-stats" storage/log/scheduler.log
```

### Common Issues

#### 1. Tasks Not Executing

**Check:**
- Redis connection status
- Timezone configuration (`server.timezone`)
- Cron expression validity
- Dynamic task status is 1 (enabled)

```bash
# Test Redis connection
redis-cli -h redis -p 6379 ping

# Check database task configuration
mysql -h mysql -u root -p bingo
SELECT * FROM sys_schedule WHERE status = 1;
```

#### 2. Task Execution Failed

Check error messages in logs:

```bash
grep "ERROR" storage/log/scheduler.log | tail -20
```

Common causes:
- Database connection failure
- Dependent services unavailable
- Task handler logic errors
- Task timeout

#### 3. Slow Task Execution

Optimization methods:
- Increase worker concurrency (modify in config)
- Optimize task handler logic
- Use task queue sharding

## Best Practices

### 1. Task Idempotency

Ensure tasks can be safely retried:

```go
func HandleTask(ctx context.Context, t *asynq.Task) error {
    taskID := t.ResultWriter().TaskID()

    // Check if task already executed
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
        log.Errorw("Task execution failed",
            "task_type", t.Type(),
            "payload", string(t.Payload()),
            "error", err)

        // Return error to trigger retry
        return fmt.Errorf("task failed: %w", err)
    }

    return nil
}
```

### 3. Monitor Task Execution Time

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

### 4. Proper Timeout and Retry Settings

```go
// Set when dispatching task
task.T.Queue(ctx, task.UserDataExport, payload).Dispatch(
    asynq.MaxRetry(3),              // Max 3 retries
    asynq.Timeout(30*time.Second),  // 30 second timeout
    asynq.Retention(24*time.Hour),  // Retain task result for 24 hours
)
```

## Integration with Other Services

### Send Email

```go
import "bingo/internal/pkg/facade"

func HandleDailyReport(ctx context.Context, t *asynq.Task) error {
    report := generateReport()

    err := facade.Mail.Send(
        "admin@example.com",
        "Daily Report",
        report,
    )

    return err
}
```

### Access Database

```go
import "bingo/internal/pkg/store"

func HandleDataSync(ctx context.Context, t *asynq.Task) error {
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

- [Asynq Official Documentation](https://github.com/hibiken/asynq) - Underlying task queue implementation
- [Cron Expression Generator](https://crontab.guru/) - Online cron expression tester

## Next Steps

- Learn how [Admin Server](/en/essentials/admserver) manages dynamic tasks
- Explore [Bot Service](/en/essentials/bot) for receiving scheduled notifications
