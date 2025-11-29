---
title: Microservice Decomposition - Bingo Go Architecture Evolution Guide
description: Learn when and how to decompose Bingo Go monolithic applications into microservices, including DDD principles, service discovery, distributed transactions, and API gateway patterns.
---

# Microservice Decomposition

When business complexity increases, you can decompose the monolithic application into multiple independent microservices. This document explains how to perform microservice decomposition.

## When to Decompose

Consider decomposing microservices in these situations:

1. **Team Growth**: Teams exceeding 10 people find monolithic application collaboration difficult
2. **Strong Module Independence**: Certain modules can be developed and deployed completely independently
3. **Performance Bottlenecks**: Certain modules need independent scaling
4. **Technology Heterogeneity**: Different modules need different technology stacks
5. **Different Release Frequency**: Some modules need frequent releases while others are relatively stable

## Decomposition Principles

### 1. Decompose by Business Domain (DDD)

```
bingo-user-service      # User service (users, authentication, permissions)
bingo-order-service     # Order service (orders, payment)
bingo-product-service   # Product service (products, inventory)
bingo-notification-service  # Notification service (email, SMS, push)
```

### 2. Decompose by Technical Responsibility

```
bingo-api-gateway       # API gateway
bingo-auth-service      # Authentication service
bingo-business-service  # Business service
bingo-data-service      # Data service
```

### 3. Single Responsibility Principle

Each service is responsible for one business domain:

✅ **Good Decomposition**:
- `user-service`: User management
- `auth-service`: Authentication/authorization
- `order-service`: Order management

❌ **Poor Decomposition**:
- `user-auth-order-service`: Mixed domains

## Decomposition Steps

### Step 1: Identify Boundaries

Analyze existing code and identify business boundaries:

```
internal/apiserver/
├── biz/
│   ├── user/       → user-service
│   ├── auth/       → auth-service
│   ├── order/      → order-service
│   └── product/    → product-service
```

### Step 2: Database Decomposition

Each service uses independent database:

```
bingo-user-service    → bingo_user_db
bingo-order-service   → bingo_order_db
bingo-product-service → bingo_product_db
```

**Note**: Cross-database joins need to be done through service calls.

### Step 3: Define Service Interfaces

Define service interfaces using gRPC:

```protobuf
// pkg/proto/user/user.proto
syntax = "proto3";

package user;

service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
}

message GetUserRequest {
  uint64 user_id = 1;
}

message GetUserResponse {
  uint64 id = 1;
  string username = 2;
  string email = 3;
}
```

### Step 4: Implement Services

Create independent service project:

```
bingo-user-service/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── biz/
│   ├── store/
│   └── grpc/          # gRPC service implementation
├── pkg/
│   └── proto/
└── go.mod
```

### Step 5: Inter-Service Calls

Call user service from other services:

```go
// order-service calls user-service
import userpb "github.com/bingo/user-service/pkg/proto/user"

type OrderBiz struct {
    userClient userpb.UserServiceClient
}

func (b *OrderBiz) CreateOrder(ctx context.Context, req *CreateOrderRequest) error {
    // Call user service to validate user
    userResp, err := b.userClient.GetUser(ctx, &userpb.GetUserRequest{
        UserId: req.UserID,
    })
    if err != nil {
        return err
    }

    // Create order logic
    // ...
}
```

## Service Discovery

### Using Consul

```go
// internal/pkg/discovery/consul.go
import "github.com/hashicorp/consul/api"

type ServiceDiscovery struct {
    client *api.Client
}

func (s *ServiceDiscovery) Register(name, addr string) error {
    registration := &api.AgentServiceRegistration{
        ID:      name + "-" + addr,
        Name:    name,
        Address: addr,
        Port:    8080,
        Check: &api.AgentServiceCheck{
            HTTP:     "http://" + addr + "/health",
            Interval: "10s",
        },
    }

    return s.client.Agent().ServiceRegister(registration)
}

func (s *ServiceDiscovery) Discover(name string) (string, error) {
    services, _, err := s.client.Health().Service(name, "", true, nil)
    if err != nil {
        return "", err
    }

    if len(services) == 0 {
        return "", errors.New("service not found")
    }

    // Simple load balancing
    service := services[rand.Intn(len(services))]
    addr := fmt.Sprintf("%s:%d", service.Service.Address, service.Service.Port)

    return addr, nil
}
```

## Distributed Transactions

### Saga Pattern

For cross-service transactions, use Saga pattern:

```go
// Order creation Saga
type CreateOrderSaga struct {
    orderService   *OrderService
    productService *ProductService
    paymentService *PaymentService
}

func (s *CreateOrderSaga) Execute(ctx context.Context, req *CreateOrderRequest) error {
    // Step 1: Create order
    orderID, err := s.orderService.Create(ctx, req)
    if err != nil {
        return err
    }

    // Step 2: Decrease stock
    if err := s.productService.DecreaseStock(ctx, req.ProductID, req.Quantity); err != nil {
        // Compensate: Cancel order
        s.orderService.Cancel(ctx, orderID)
        return err
    }

    // Step 3: Charge payment
    if err := s.paymentService.Charge(ctx, req.Amount); err != nil {
        // Compensate: Restore stock
        s.productService.IncreaseStock(ctx, req.ProductID, req.Quantity)
        // Compensate: Cancel order
        s.orderService.Cancel(ctx, orderID)
        return err
    }

    return nil
}
```

## Configuration Center

Use Consul as configuration center:

```go
import "github.com/hashicorp/consul/api"

type ConfigCenter struct {
    client *api.Client
}

func (c *ConfigCenter) Get(key string) (string, error) {
    pair, _, err := c.client.KV().Get(key, nil)
    if err != nil {
        return "", err
    }

    return string(pair.Value), nil
}

func (c *ConfigCenter) Watch(key string, callback func(string)) {
    // Watch configuration changes
    go func() {
        var index uint64
        for {
            pair, meta, err := c.client.KV().Get(key, &api.QueryOptions{
                WaitIndex: index,
            })
            if err != nil {
                continue
            }

            if meta.LastIndex > index {
                index = meta.LastIndex
                callback(string(pair.Value))
            }
        }
    }()
}
```

## Distributed Tracing

Integrate Jaeger for distributed request tracing:

```go
import (
    "github.com/opentracing/opentracing-go"
    "github.com/uber/jaeger-client-go"
)

func InitTracer(serviceName string) (opentracing.Tracer, io.Closer) {
    cfg := &config.Configuration{
        ServiceName: serviceName,
        Sampler: &config.SamplerConfig{
            Type:  jaeger.SamplerTypeConst,
            Param: 1,
        },
        Reporter: &config.ReporterConfig{
            LocalAgentHostPort: "127.0.0.1:6831",
        },
    }

    tracer, closer, _ := cfg.NewTracer()
    opentracing.SetGlobalTracer(tracer)

    return tracer, closer
}

// Use in HTTP middleware
func TracingMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        span := opentracing.GlobalTracer().StartSpan(c.Request.URL.Path)
        defer span.Finish()

        c.Set("tracing-span", span)
        c.Next()
    }
}
```

## API Gateway

Use Nginx or Kong as API gateway:

```nginx
# nginx.conf
upstream user-service {
    server user-service-1:8080;
    server user-service-2:8080;
}

upstream order-service {
    server order-service-1:8080;
    server order-service-2:8080;
}

server {
    listen 80;

    location /api/users {
        proxy_pass http://user-service;
    }

    location /api/orders {
        proxy_pass http://order-service;
    }
}
```

## Monitoring and Alerting

Each service exposes Prometheus metrics:

```go
import "github.com/prometheus/client_golang/prometheus"

var (
    requestCounter = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "api_requests_total",
            Help: "Total number of API requests",
        },
        []string{"service", "method", "status"},
    )
)

func init() {
    prometheus.MustRegister(requestCounter)
}

// Record in middleware
func MetricsMiddleware(serviceName string) gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()

        requestCounter.WithLabelValues(
            serviceName,
            c.Request.Method,
            strconv.Itoa(c.Writer.Status()),
        ).Inc()
    }
}
```

## Best Practices

1. **Clear service boundaries**: Avoid circular dependencies
2. **Independent databases**: Each service uses independent database
3. **Asynchronous communication**: Use message queues to decouple services
4. **Circuit breaker and degradation**: Use Hystrix etc. for circuit breakers
5. **Complete monitoring**: Each service needs monitoring and logging
6. **Version management**: Versioned APIs with backward compatibility

## Next Step

Congratulations! You've completed the Bingo documentation. Return to the [Introduction](../guide/what-is-bingo.md) to revisit any topics, or start building your project with confidence.
