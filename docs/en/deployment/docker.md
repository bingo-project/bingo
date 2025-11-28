# Docker Deployment

This guide explains how to containerize and deploy Bingo applications using Docker.

## Prerequisites

- Docker (version 20.10+)
- Docker Compose (version 2.0+)
- Your Bingo project

## Building Docker Image

### 1. Create Dockerfile

Create `Dockerfile` in your project root:

```dockerfile
# Multi-stage build for minimal image size
FROM golang:1.23 AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -o myapp-apiserver ./cmd/myapp-apiserver/main.go

# Final stage - minimal runtime image
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/myapp-apiserver .

# Copy configuration (optional)
COPY configs/myapp-apiserver.yaml ./

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./myapp-apiserver", "-c", "myapp-apiserver.yaml"]
```

### 2. Create .dockerignore

```
node_modules
.git
.gitignore
.env
.env.local
.DS_Store
__pycache__
*.pyc
.pytest_cache
.vscode
.idea
*.log
_output
dist
build
.air.toml
```

### 3. Build the Image

```bash
# Build image with tag
docker build -t myapp:v1.0.0 .

# Or use latest tag
docker build -t myapp:latest .
```

## Using Docker Compose for Development

Create `docker-compose.yaml` in your project root:

```yaml
version: '3.8'

services:
  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: root
      MYSQL_DATABASE: myapp
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  apiserver:
    build:
      context: .
      dockerfile: Dockerfile
    environment:
      - DB_HOST=mysql
      - DB_PORT=3306
      - REDIS_HOST=redis
      - REDIS_PORT=6379
    ports:
      - "8080:8080"
    depends_on:
      mysql:
        condition: service_healthy
      redis:
        condition: service_healthy
    volumes:
      - ./configs:/root/configs
    command: ./myapp-apiserver -c myapp-apiserver.yaml

volumes:
  mysql_data:
  redis_data:
```

### Run with Docker Compose

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f apiserver

# Stop all services
docker-compose down

# Reset everything (remove volumes)
docker-compose down -v
```

## Docker Compose for Production

Create `docker-compose.prod.yaml` for production:

```yaml
version: '3.8'

services:
  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: ${MYSQL_ROOT_PASSWORD}
      MYSQL_DATABASE: ${MYSQL_DATABASE}
    volumes:
      - /data/mysql:/var/lib/mysql
    restart: always
    networks:
      - bingo-network

  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes --requirepass ${REDIS_PASSWORD}
    volumes:
      - /data/redis:/data
    restart: always
    networks:
      - bingo-network

  apiserver:
    image: myapp:${APP_VERSION}
    environment:
      - DB_HOST=mysql
      - DB_USERNAME=root
      - DB_PASSWORD=${MYSQL_ROOT_PASSWORD}
      - DB_PORT=3306
      - REDIS_HOST=redis
      - REDIS_PASSWORD=${REDIS_PASSWORD}
    ports:
      - "8080:8080"
    depends_on:
      - mysql
      - redis
    restart: always
    networks:
      - bingo-network
    deploy:
      replicas: 2

networks:
  bingo-network:
    driver: bridge

volumes:
  mysql_data:
  redis_data:
```

Run production compose:

```bash
# Create .env file
cat > .env << EOF
MYSQL_ROOT_PASSWORD=your-secure-password
MYSQL_DATABASE=myapp
REDIS_PASSWORD=your-redis-password
APP_VERSION=v1.0.0
EOF

# Start services
docker-compose -f docker-compose.prod.yaml up -d
```

## Container Registry

### Push to Docker Hub

```bash
# Login to Docker Hub
docker login

# Tag image with Docker Hub username
docker tag myapp:v1.0.0 myusername/myapp:v1.0.0

# Push image
docker push myusername/myapp:v1.0.0

# Pull image
docker pull myusername/myapp:v1.0.0
```

### Push to Private Registry

```bash
# Tag for private registry
docker tag myapp:v1.0.0 registry.example.com/myapp:v1.0.0

# Push to private registry
docker push registry.example.com/myapp:v1.0.0
```

## Environment Configuration

Use environment variables for configuration:

```bash
# Run container with environment variables
docker run -e DB_HOST=mysql \
           -e DB_PORT=3306 \
           -e REDIS_HOST=redis \
           -p 8080:8080 \
           myapp:latest
```

Or use .env file:

```bash
docker run --env-file .env -p 8080:8080 myapp:latest
```

## Volume Mounting

### Development

Mount source code for hot reload:

```bash
docker run -v $(pwd):/app \
           -e DB_HOST=mysql \
           -p 8080:8080 \
           myapp:latest
```

### Production

Mount configuration and data:

```bash
docker run -v /etc/myapp/config.yaml:/root/config.yaml \
           -v /data/logs:/root/logs \
           -p 8080:8080 \
           myapp:latest
```

## Health Checks

The Dockerfile includes a health check. Monitor container health:

```bash
# View container status
docker ps

# View health logs
docker inspect --format='{{json .State.Health}}' container_id
```

## Troubleshooting

### Container Exit Immediately

```bash
# View container logs
docker logs container_id

# Run with interactive terminal
docker run -it myapp:latest /bin/sh
```

### Database Connection Issues

```bash
# Check if MySQL is running
docker ps | grep mysql

# Test connection
docker exec mysql-container mysqladmin ping
```

### Port Already in Use

```bash
# Map to different port
docker run -p 8081:8080 myapp:latest

# Or kill process using port
lsof -i :8080
kill -9 <PID>
```

## Best Practices

1. **Use .dockerignore**: Exclude unnecessary files to reduce image size
2. **Multi-stage Build**: Use builder stage to reduce final image size
3. **Health Checks**: Always include health checks
4. **Non-root User**: Run container as non-root user for security
5. **Environment Variables**: Externalize configuration
6. **Volume Mounting**: Use volumes for persistent data
7. **Resource Limits**: Set CPU and memory limits
8. **Logging**: Redirect logs to stdout/stderr for Docker logs

## Next Steps

- Deploy to Kubernetes for production orchestration
- Set up CI/CD pipeline for automated builds
- Configure monitoring and logging
- Implement auto-scaling policies

## References

- [Docker Documentation](https://docs.docker.com/)
- [Docker Best Practices](https://docs.docker.com/develop/dev-best-practices/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
