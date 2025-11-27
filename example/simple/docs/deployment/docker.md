# 🐳 Docker部署

YYHertz应用的Docker化部署方案，包括单容器部署和Docker Compose多容器编排。

## 基础Docker部署

### Dockerfile编写

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/views ./views
COPY --from=builder /app/static ./static

EXPOSE 8080
CMD ["./main"]
```

### 构建和运行

```bash
# 构建镜像
docker build -t yyhertz-app .

# 运行容器
docker run -p 8080:8080 yyhertz-app
```

## Docker Compose部署

完整的多服务编排方案，包含应用、数据库、缓存等服务。

### docker-compose.yml

```yaml
version: '3.8'
services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - ENV=production
      - DB_HOST=mysql
      - REDIS_HOST=redis
    depends_on:
      - mysql
      - redis
    restart: unless-stopped

  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: rootpassword
      MYSQL_DATABASE: yyhertz
      MYSQL_USER: yyhertz
      MYSQL_PASSWORD: password
    volumes:
      - mysql_data:/var/lib/mysql
    restart: unless-stopped

  redis:
    image: redis:alpine
    command: redis-server --requirepass password
    volumes:
      - redis_data:/data
    restart: unless-stopped

volumes:
  mysql_data:
  redis_data:
```

### 部署命令

```bash
# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f app

# 停止服务
docker-compose down
```

## 生产环境优化

### 多阶段构建优化

```dockerfile
FROM golang:1.21-alpine AS builder
RUN apk add --no-cache git ca-certificates tzdata
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o app main.go

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /app/app /app
COPY --from=builder /app/views /views
COPY --from=builder /app/static /static
EXPOSE 8080
CMD ["/app"]
```

### 健康检查

```dockerfile
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1
```

## 容器编排最佳实践

### 环境变量管理

```bash
# .env文件
DB_PASSWORD=secure_password
REDIS_PASSWORD=secure_redis_password
JWT_SECRET=your_jwt_secret
```

### 数据持久化

```yaml
volumes:
  - ./data/mysql:/var/lib/mysql
  - ./data/redis:/data
  - ./logs:/var/log/app
```

### 网络配置

```yaml
networks:
  app-network:
    driver: bridge
```

完整的Docker化部署解决方案！