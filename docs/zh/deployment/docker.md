# Docker 部署

本文介绍如何使用 Docker 部署 Bingo 项目。

## 本地开发环境

### 启动依赖服务

使用 Docker Compose 快速启动 MySQL 和 Redis:

```bash
docker-compose -f deployments/docker/docker-compose.yaml up -d mysql redis
```

### 启动所有服务

```bash
# 启动所有服务(包括应用服务)
docker-compose -f deployments/docker/docker-compose.yaml up -d

# 查看服务状态
docker-compose -f deployments/docker/docker-compose.yaml ps

# 查看日志
docker-compose -f deployments/docker/docker-compose.yaml logs -f bingo-apiserver

# 停止服务
docker-compose -f deployments/docker/docker-compose.yaml down
```

## 构建镜像

### 单服务构建

```bash
# 构建 API Server 镜像
make image

# 构建指定服务
docker build -t bingo-apiserver:latest \
  -f build/docker/Dockerfile.apiserver .
```

### Dockerfile 示例

```dockerfile
# build/docker/Dockerfile.apiserver
FROM golang:1.23.1-alpine AS builder

WORKDIR /build

# 复制依赖文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源码
COPY . .

# 编译
RUN CGO_ENABLED=0 GOOS=linux go build \
  -o bingo-apiserver \
  cmd/bingo-apiserver/main.go

# 运行镜像
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# 从构建镜像复制二进制文件
COPY --from=builder /build/bingo-apiserver .

# 复制配置文件
COPY configs/bingo-apiserver.example.yaml /etc/bingo/config.yaml

EXPOSE 8080 8081 8082

CMD ["./bingo-apiserver", "-c", "/etc/bingo/config.yaml"]
```

## 生产环境部署

### 1. 构建并推送镜像

```bash
# 构建镜像
make image

# 打标签
docker tag bingo-apiserver:latest registry.example.com/bingo-apiserver:v1.0.0

# 推送到镜像仓库
docker push registry.example.com/bingo-apiserver:v1.0.0
```

### 2. 准备配置文件

```yaml
# config/production.yaml
server:
  mode: release
  addr: 0.0.0.0:8080

mysql:
  host: mysql-prod.example.com:3306
  database: bingo_prod
  username: bingo
  password: ${MYSQL_PASSWORD}  # 使用环境变量

redis:
  host: redis-prod.example.com:6379
  password: ${REDIS_PASSWORD}

jwt:
  secretKey: ${JWT_SECRET}
```

### 3. 运行容器

```bash
# 拉取镜像
docker pull registry.example.com/bingo-apiserver:v1.0.0

# 运行容器
docker run -d \
  --name bingo-apiserver \
  --restart unless-stopped \
  -p 8080:8080 \
  -p 8081:8081 \
  -v /data/bingo/config.yaml:/etc/bingo/config.yaml \
  -v /data/bingo/logs:/app/storage/log \
  -e MYSQL_PASSWORD="your-mysql-password" \
  -e REDIS_PASSWORD="your-redis-password" \
  -e JWT_SECRET="your-jwt-secret" \
  registry.example.com/bingo-apiserver:v1.0.0
```

### 4. 使用 Docker Compose

```yaml
# docker-compose.prod.yaml
version: '3.8'

services:
  bingo-apiserver:
    image: registry.example.com/bingo-apiserver:v1.0.0
    ports:
      - "8080:8080"
      - "8081:8081"
    volumes:
      - ./config.yaml:/etc/bingo/config.yaml
      - ./logs:/app/storage/log
    environment:
      - MYSQL_PASSWORD=${MYSQL_PASSWORD}
      - REDIS_PASSWORD=${REDIS_PASSWORD}
      - JWT_SECRET=${JWT_SECRET}
    restart: unless-stopped
    depends_on:
      - mysql
      - redis

  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: ${MYSQL_ROOT_PASSWORD}
      MYSQL_DATABASE: bingo
    volumes:
      - mysql-data:/var/lib/mysql
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    command: redis-server --requirepass ${REDIS_PASSWORD}
    volumes:
      - redis-data:/data
    restart: unless-stopped

volumes:
  mysql-data:
  redis-data:
```

启动:

```bash
# 设置环境变量
export MYSQL_ROOT_PASSWORD="root-password"
export MYSQL_PASSWORD="bingo-password"
export REDIS_PASSWORD="redis-password"
export JWT_SECRET="jwt-secret"

# 启动服务
docker-compose -f docker-compose.prod.yaml up -d
```

## 多实例部署

使用 Nginx 负载均衡:

```nginx
# nginx.conf
upstream bingo_api {
    server api-1:8080;
    server api-2:8080;
    server api-3:8080;
}

server {
    listen 80;
    server_name api.example.com;

    location / {
        proxy_pass http://bingo_api;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

Docker Compose:

```yaml
services:
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
    depends_on:
      - api-1
      - api-2
      - api-3

  api-1:
    image: registry.example.com/bingo-apiserver:v1.0.0
    # ...

  api-2:
    image: registry.example.com/bingo-apiserver:v1.0.0
    # ...

  api-3:
    image: registry.example.com/bingo-apiserver:v1.0.0
    # ...
```

## 健康检查

添加健康检查:

```yaml
services:
  bingo-apiserver:
    image: registry.example.com/bingo-apiserver:v1.0.0
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
```

## 日志管理

### 输出到容器日志

```bash
# 查看日志
docker logs -f bingo-apiserver

# 限制日志大小
docker run -d \
  --log-driver json-file \
  --log-opt max-size=10m \
  --log-opt max-file=3 \
  bingo-apiserver
```

### 使用卷挂载

```yaml
services:
  bingo-apiserver:
    volumes:
      - ./logs:/app/storage/log
```

## 安全建议

1. **不要在镜像中硬编码密钥**,使用环境变量
2. **使用非 root 用户**运行容器
3. **定期更新基础镜像**
4. **限制容器资源**:

```yaml
services:
  bingo-apiserver:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G
        reservations:
          cpus: '1'
          memory: 1G
```

## 常见问题

### 容器无法连接数据库

检查网络配置:

```bash
# 查看容器网络
docker network ls
docker network inspect bridge

# 使用 Docker Compose 自动网络
# 服务间使用服务名连接
mysql:
  host: mysql:3306  # 使用服务名
```

### 性能优化

1. 使用多阶段构建减小镜像大小
2. 使用 Alpine 基础镜像
3. 合理设置资源限制

## 下一步

- [配置详解](./configuration.md) - 配置文件详细说明（待实现）
- [监控调试](./monitoring.md) - 生产环境监控（待实现）
- [常见问题](./troubleshooting.md) - 部署问题排查（待实现）
