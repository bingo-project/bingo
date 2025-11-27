# 快速开始

本指南将帮助你在 10 分钟内启动 Bingo 项目并运行第一个 API。

## 1. 克隆项目

```bash
git clone <repository-url>
cd bingo
```

## 2. 配置环境

```bash
# 复制配置文件
cp configs/bingo-apiserver.example.yaml bingo-apiserver.yaml

# 根据实际环境修改配置
vim bingo-apiserver.yaml
```

主要配置项:
- 数据库连接(MySQL)
- Redis 连接
- JWT 密钥

## 3. 启动依赖服务

使用 Docker Compose 快速启动 MySQL 和 Redis:

```bash
docker-compose -f deployments/docker/docker-compose.yaml up -d mysql redis
```

## 4. 数据库迁移

```bash
# 编译项目(输出路径:./_output/platforms/<os>/<arch>/)
make build

# 复制配置文件,并修改数据库配置
cp configs/{app}-admserver.example.yaml {app}-admserver.yaml

# Build your app ctl
make build BINS="{app}ctl"

# 执行数据库迁移
./_output/platforms/{os}/{arch}/{app}ctl migrate up
```

> **说明**:`make build` 会将二进制文件输出到 `./_output/platforms/<os>/<arch>/` 目录(如 `./_output/platforms/darwin/arm64/`)

## 5. 启动服务

### 方式一:直接运行

```bash
make build
bingo-apiserver -c bingo-apiserver.yaml
```

### 方式二:开发模式(热重启)

```bash
cp .air.example.toml .air.toml
air
```

## 6. 验证服务

```bash
# 检查服务状态
curl http://localhost:8080/health

# 访问 Swagger 文档
open http://localhost:8080/swagger/index.html
```

## 常用命令

```bash
# 编译所有服务
make build

# 编译指定服务
make build BINS="bingo-apiserver"

# 运行测试
make test

# 代码检查
make lint

# 生成 Swagger 文档
make swagger

# 清理构建产物
make clean
```

## 下一步

- [开发第一个功能](./first-feature.md) - 通过实例学习 Bingo 开发
- [项目结构](./project-structure.md) - 理解项目目录组织
- [分层架构](../essentials/layered-design.md) - 理解三层架构设计

## 遇到问题?

查看 [常见问题](../deployment/troubleshooting.md) 或提交 Issue。
