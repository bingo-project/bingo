# 微服务拆分

当业务复杂度增加时,可以将单体应用拆分为多个独立的微服务。本文介绍如何进行微服务拆分。

## 何时拆分

在以下情况考虑拆分微服务:

1. **团队规模增长**: 超过 10 人的团队,单体应用协作困难
2. **模块独立性强**: 某些模块可以完全独立开发和部署
3. **性能瓶颈**: 某些模块需要独立扩展
4. **技术异构**: 不同模块需要使用不同技术栈
5. **发布频率不同**: 某些模块需要频繁发布,其他模块相对稳定

## 拆分原则

### 1. 按业务领域拆分(DDD)

```
bingo-user-service      # 用户服务(用户、认证、权限)
bingo-order-service     # 订单服务(订单、支付)
bingo-product-service   # 商品服务(商品、库存)
bingo-notification-service  # 通知服务(邮件、短信、推送)
```

### 2. 按技术职责拆分

```
bingo-api-gateway       # API 网关
bingo-auth-service      # 认证服务
bingo-business-service  # 业务服务
bingo-data-service      # 数据服务
```

### 3. 单一职责原则

每个服务只负责一个业务领域:

✅ **好的拆分**:
- `user-service`: 用户管理
- `auth-service`: 认证授权
- `order-service`: 订单管理

❌ **不好的拆分**:
- `user-auth-order-service`: 混合多个领域

## 拆分步骤

### 步骤 1: 识别边界

分析现有代码,识别业务边界:

```
internal/apiserver/
├── biz/
│   ├── user/       → user-service
│   ├── auth/       → auth-service
│   ├── order/      → order-service
│   └── product/    → product-service
```

### 步骤 2: 数据库拆分

每个服务使用独立数据库:

```
bingo-user-service    → bingo_user_db
bingo-order-service   → bingo_order_db
bingo-product-service → bingo_product_db
```

**注意**: 跨库关联查询需要通过服务调用实现。

### 步骤 3: 定义服务接口

使用 gRPC 定义服务接口:

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

### 步骤 4: 实现服务

创建独立的服务项目:

```
bingo-user-service/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── biz/
│   ├── store/
│   └── grpc/          # gRPC 服务实现
├── pkg/
│   └── proto/
└── go.mod
```

### 步骤 5: 服务间调用

在其他服务中调用用户服务:

```go
// order-service 调用 user-service
import userpb "github.com/bingo/user-service/pkg/proto/user"

type OrderBiz struct {
    userClient userpb.UserServiceClient
}

func (b *OrderBiz) CreateOrder(ctx context.Context, req *CreateOrderRequest) error {
    // 调用用户服务验证用户
    userResp, err := b.userClient.GetUser(ctx, &userpb.GetUserRequest{
        UserId: req.UserID,
    })
    if err != nil {
        return err
    }

    // 创建订单逻辑
    // ...
}
```

## 服务发现

### 使用 Consul

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

    // 简单负载均衡
    service := services[rand.Intn(len(services))]
    addr := fmt.Sprintf("%s:%d", service.Service.Address, service.Service.Port)

    return addr, nil
}
```

## 分布式事务

### Saga 模式

对于跨服务事务,使用 Saga 模式:

```go
// 订单创建 Saga
type CreateOrderSaga struct {
    orderService   *OrderService
    productService *ProductService
    paymentService *PaymentService
}

func (s *CreateOrderSaga) Execute(ctx context.Context, req *CreateOrderRequest) error {
    // 步骤 1: 创建订单
    orderID, err := s.orderService.Create(ctx, req)
    if err != nil {
        return err
    }

    // 步骤 2: 减库存
    if err := s.productService.DecreaseStock(ctx, req.ProductID, req.Quantity); err != nil {
        // 补偿: 取消订单
        s.orderService.Cancel(ctx, orderID)
        return err
    }

    // 步骤 3: 扣款
    if err := s.paymentService.Charge(ctx, req.Amount); err != nil {
        // 补偿: 恢复库存
        s.productService.IncreaseStock(ctx, req.ProductID, req.Quantity)
        // 补偿: 取消订单
        s.orderService.Cancel(ctx, orderID)
        return err
    }

    return nil
}
```

## 配置中心

使用 Consul 作为配置中心:

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
    // 监听配置变化
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

## 链路追踪

集成 Jaeger 实现分布式链路追踪:

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

// 在 HTTP 中间件中使用
func TracingMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        span := opentracing.GlobalTracer().StartSpan(c.Request.URL.Path)
        defer span.Finish()

        c.Set("tracing-span", span)
        c.Next()
    }
}
```

## API 网关

使用 Nginx 或 Kong 作为 API 网关:

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

## 监控告警

每个服务暴露 Prometheus 指标:

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

// 在中间件中记录
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

## 最佳实践

1. **服务边界清晰**: 避免循环依赖
2. **独立数据库**: 每个服务使用独立数据库
3. **异步通信**: 使用消息队列解耦服务
4. **熔断降级**: 使用 Hystrix 等熔断器
5. **监控完善**: 每个服务都要有监控和日志
6. **版本管理**: API 版本化,向后兼容

## 下一步

有关服务发现、链路追踪和性能优化的详细文档正在筹备中，敬请期待！
