# 🚀 部署上线

YYHertz应用可以部署到多种环境，本文档介绍常见的部署方案和最佳实践。

## 部署准备

### 环境要求

- **Go版本**: 1.19+
- **操作系统**: Linux/Windows/macOS
- **内存**: 建议2GB+
- **CPU**: 建议2核+
- **磁盘**: 建议10GB+

### 构建配置

#### 生产环境配置

```go
// config/production.go
package config

type ProductionConfig struct {
    Server   ServerConfig   `yaml:"server"`
    Database DatabaseConfig `yaml:"database"`
    Redis    RedisConfig    `yaml:"redis"`
    Logger   LoggerConfig   `yaml:"logger"`
}

type ServerConfig struct {
    Host         string `yaml:"host"`
    Port         int    `yaml:"port"`
    ReadTimeout  int    `yaml:"read_timeout"`
    WriteTimeout int    `yaml:"write_timeout"`
    TLS          struct {
        Cert string `yaml:"cert"`
        Key  string `yaml:"key"`
    } `yaml:"tls"`
}

type DatabaseConfig struct {
    Host     string `yaml:"host"`
    Port     int    `yaml:"port"`
    Username string `yaml:"username"`
    Password string `yaml:"password"`
    Database string `yaml:"database"`
    MaxIdle  int    `yaml:"max_idle"`
    MaxOpen  int    `yaml:"max_open"`
}
```

#### 配置文件示例

```yaml
# config/production.yaml
server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: 30
  write_timeout: 30
  tls:
    cert: "/etc/ssl/certs/app.crt"
    key: "/etc/ssl/private/app.key"

database:
  host: "localhost"
  port: 3306
  username: "app_user"
  password: "${DB_PASSWORD}"
  database: "production_db"
  max_idle: 10
  max_open: 100

redis:
  host: "localhost"
  port: 6379
  password: "${REDIS_PASSWORD}"
  database: 0

logger:
  level: "info"
  format: "json"
  output: "/var/log/app/app.log"
```

## 编译构建

### 标准编译

```bash
# 编译Linux版本
GOOS=linux GOARCH=amd64 go build -o app-linux main.go

# 编译Windows版本
GOOS=windows GOARCH=amd64 go build -o app-windows.exe main.go

# 编译macOS版本
GOOS=darwin GOARCH=amd64 go build -o app-darwin main.go
```

### 优化编译

```bash
# 减小二进制文件大小
go build -ldflags="-s -w" -o app main.go

# 添加版本信息
VERSION=$(git describe --tags --always)
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
go build -ldflags="-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}" -o app main.go
```

### 交叉编译脚本

```bash
#!/bin/bash
# build.sh

APP_NAME="yyhertz-app"
VERSION=$(git describe --tags --always)
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')

LDFLAGS="-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -s -w"

# 创建构建目录
mkdir -p dist

# Linux amd64
echo "Building for Linux amd64..."
GOOS=linux GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o "dist/${APP_NAME}-linux-amd64" main.go

# Linux arm64
echo "Building for Linux arm64..."
GOOS=linux GOARCH=arm64 go build -ldflags="${LDFLAGS}" -o "dist/${APP_NAME}-linux-arm64" main.go

# Windows amd64
echo "Building for Windows amd64..."
GOOS=windows GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o "dist/${APP_NAME}-windows-amd64.exe" main.go

# macOS amd64
echo "Building for macOS amd64..."
GOOS=darwin GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o "dist/${APP_NAME}-darwin-amd64" main.go

# macOS arm64 (M1)
echo "Building for macOS arm64..."
GOOS=darwin GOARCH=arm64 go build -ldflags="${LDFLAGS}" -o "dist/${APP_NAME}-darwin-arm64" main.go

echo "Build completed!"
```

## Docker部署

### Dockerfile

```dockerfile
# 多阶段构建
FROM golang:1.21-alpine AS builder

# 安装必要工具
RUN apk add --no-cache git ca-certificates tzdata

# 设置工作目录
WORKDIR /app

# 复制依赖文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 编译应用
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o app main.go

# 生产镜像
FROM alpine:latest

# 安装ca证书和时区数据
RUN apk --no-cache add ca-certificates tzdata

# 创建非root用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/app .
COPY --from=builder /app/views ./views
COPY --from=builder /app/static ./static
COPY --from=builder /app/config ./config

# 创建日志目录
RUN mkdir -p /var/log/app && \
    chown -R appuser:appgroup /app /var/log/app

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 启动应用
CMD ["./app"]
```

### Docker Compose

```yaml
# docker-compose.yml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - ENV=production
      - DB_HOST=mysql
      - DB_PASSWORD=${DB_PASSWORD}
      - REDIS_HOST=redis
      - REDIS_PASSWORD=${REDIS_PASSWORD}
    depends_on:
      - mysql
      - redis
    volumes:
      - ./logs:/var/log/app
    restart: unless-stopped
    networks:
      - app-network

  mysql:
    image: mysql:8.0
    environment:
      - MYSQL_ROOT_PASSWORD=${MYSQL_ROOT_PASSWORD}
      - MYSQL_DATABASE=production_db
      - MYSQL_USER=app_user
      - MYSQL_PASSWORD=${DB_PASSWORD}
    volumes:
      - mysql_data:/var/lib/mysql
      - ./init.sql:/docker-entrypoint-initdb.d/init.sql
    ports:
      - "3306:3306"
    restart: unless-stopped
    networks:
      - app-network

  redis:
    image: redis:7-alpine
    command: redis-server --requirepass ${REDIS_PASSWORD}
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    restart: unless-stopped
    networks:
      - app-network

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/ssl/certs
    depends_on:
      - app
    restart: unless-stopped
    networks:
      - app-network

volumes:
  mysql_data:
  redis_data:

networks:
  app-network:
    driver: bridge
```

### 环境变量文件

```bash
# .env
DB_PASSWORD=your_secure_db_password
MYSQL_ROOT_PASSWORD=your_secure_root_password
REDIS_PASSWORD=your_secure_redis_password
```

## Kubernetes部署

### Deployment配置

```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: yyhertz-app
  labels:
    app: yyhertz-app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: yyhertz-app
  template:
    metadata:
      labels:
        app: yyhertz-app
    spec:
      containers:
      - name: app
        image: yyhertz-app:latest
        ports:
        - containerPort: 8080
        env:
        - name: ENV
          value: "production"
        - name: DB_HOST
          value: "mysql-service"
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: db-secret
              key: password
        - name: REDIS_HOST
          value: "redis-service"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ping
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
---
apiVersion: v1
kind: Service
metadata:
  name: yyhertz-app-service
spec:
  selector:
    app: yyhertz-app
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: LoadBalancer
```

### ConfigMap和Secret

```yaml
# k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  server.yaml: |
    server:
      host: "0.0.0.0"
      port: 8080
      read_timeout: 30
      write_timeout: 30

---
apiVersion: v1
kind: Secret
metadata:
  name: db-secret
type: Opaque
data:
  password: eW91cl9zZWN1cmVfcGFzc3dvcmQ=  # base64编码
```

### Ingress配置

```yaml
# k8s/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: yyhertz-ingress
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
spec:
  tls:
  - hosts:
    - your-domain.com
    secretName: yyhertz-tls
  rules:
  - host: your-domain.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: yyhertz-app-service
            port:
              number: 80
```

## 传统服务器部署

### Systemd服务

```ini
# /etc/systemd/system/yyhertz.service
[Unit]
Description=YYHertz Web Application
After=network.target

[Service]
Type=simple
User=app
Group=app
WorkingDirectory=/opt/yyhertz
ExecStart=/opt/yyhertz/app
ExecReload=/bin/kill -HUP $MAINPID
Restart=always
RestartSec=5
Environment=ENV=production
Environment=CONFIG_PATH=/opt/yyhertz/config

# 安全设置
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/yyhertz/logs

[Install]
WantedBy=multi-user.target
```

### 部署脚本

```bash
#!/bin/bash
# deploy.sh

APP_NAME="yyhertz"
APP_DIR="/opt/${APP_NAME}"
SERVICE_NAME="${APP_NAME}.service"
BACKUP_DIR="/opt/backup"

echo "Starting deployment..."

# 创建备份
if [ -f "${APP_DIR}/app" ]; then
    echo "Creating backup..."
    mkdir -p ${BACKUP_DIR}
    cp ${APP_DIR}/app ${BACKUP_DIR}/app.$(date +%Y%m%d_%H%M%S)
fi

# 停止服务
echo "Stopping service..."
sudo systemctl stop ${SERVICE_NAME}

# 更新应用
echo "Updating application..."
cp app ${APP_DIR}/
chmod +x ${APP_DIR}/app

# 更新静态文件
cp -r views ${APP_DIR}/
cp -r static ${APP_DIR}/
cp -r config ${APP_DIR}/

# 设置权限
chown -R app:app ${APP_DIR}

# 启动服务
echo "Starting service..."
sudo systemctl start ${SERVICE_NAME}
sudo systemctl enable ${SERVICE_NAME}

# 检查状态
sleep 5
if systemctl is-active --quiet ${SERVICE_NAME}; then
    echo "Deployment successful!"
    echo "Service status:"
    systemctl status ${SERVICE_NAME}
else
    echo "Deployment failed!"
    exit 1
fi
```

## 负载均衡

### Nginx配置

```nginx
# nginx.conf
upstream yyhertz_backend {
    least_conn;
    server 127.0.0.1:8080 weight=1 max_fails=3 fail_timeout=30s;
    server 127.0.0.1:8081 weight=1 max_fails=3 fail_timeout=30s;
    server 127.0.0.1:8082 weight=1 max_fails=3 fail_timeout=30s;
}

server {
    listen 80;
    server_name your-domain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com;

    # SSL配置
    ssl_certificate /etc/ssl/certs/your-domain.crt;
    ssl_certificate_key /etc/ssl/private/your-domain.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA384:ECDHE-RSA-AES128-GCM-SHA256;
    ssl_prefer_server_ciphers off;

    # 安全头
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header X-XSS-Protection "1; mode=block";
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;

    # 静态文件
    location /static/ {
        alias /opt/yyhertz/static/;
        expires 1y;
        add_header Cache-Control "public, immutable";
    }

    # API代理
    location / {
        proxy_pass http://yyhertz_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # 超时设置
        proxy_connect_timeout 30s;
        proxy_send_timeout 30s;
        proxy_read_timeout 30s;
        
        # 健康检查
        proxy_set_header Connection "";
        proxy_http_version 1.1;
    }

    # 健康检查
    location /health {
        access_log off;
        proxy_pass http://yyhertz_backend;
    }
}
```

## 监控告警

### Prometheus监控

```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'yyhertz'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
    scrape_interval: 5s
```

### 应用监控代码

```go
package middleware

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    httpRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )

    httpRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "Duration of HTTP requests",
        },
        []string{"method", "endpoint"},
    )
)

func PrometheusMiddleware() gin.HandlerFunc {
    return gin.HandlerFunc(func(c *gin.Context) {
        start := time.Now()
        
        c.Next()
        
        duration := time.Since(start).Seconds()
        status := strconv.Itoa(c.Writer.Status())
        
        httpRequestsTotal.WithLabelValues(c.Request.Method, c.FullPath(), status).Inc()
        httpRequestDuration.WithLabelValues(c.Request.Method, c.FullPath()).Observe(duration)
    })
}
```

## 日志管理

### 结构化日志

```go
package logger

import (
    "os"
    "github.com/sirupsen/logrus"
)

type Logger struct {
    *logrus.Logger
}

func NewLogger() *Logger {
    log := logrus.New()
    
    // 设置日志格式
    log.SetFormatter(&logrus.JSONFormatter{
        TimestampFormat: "2006-01-02 15:04:05",
    })
    
    // 设置日志输出
    if os.Getenv("ENV") == "production" {
        file, err := os.OpenFile("/var/log/app/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
        if err != nil {
            log.Fatalln("Failed to open log file:", err)
        }
        log.SetOutput(file)
    } else {
        log.SetOutput(os.Stdout)
    }
    
    // 设置日志级别
    level, err := logrus.ParseLevel(os.Getenv("LOG_LEVEL"))
    if err != nil {
        level = logrus.InfoLevel
    }
    log.SetLevel(level)
    
    return &Logger{log}
}
```

### 日志轮转配置

```json
// /etc/logrotate.d/yyhertz
/var/log/app/*.log {
    daily
    rotate 30
    compress
    delaycompress
    missingok
    notifempty
    create 644 app app
    postrotate
        systemctl reload yyhertz
    endscript
}
```

## 性能优化

### 应用优化

```go
// 连接池配置
func optimizeDatabase(db *gorm.DB) {
    sqlDB, _ := db.DB()
    sqlDB.SetMaxIdleConns(10)
    sqlDB.SetMaxOpenConns(100)
    sqlDB.SetConnMaxLifetime(time.Hour)
}

// Redis连接池
func optimizeRedis() *redis.Client {
    return redis.NewClient(&redis.Options{
        Addr:         "localhost:6379",
        PoolSize:     100,
        MinIdleConns: 10,
        MaxRetries:   3,
    })
}

// 缓存中间件
func CacheMiddleware(duration time.Duration) gin.HandlerFunc {
    return gin.HandlerFunc(func(c *gin.Context) {
        // 缓存逻辑
    })
}
```

## 安全配置

### HTTPS和安全头

```go
func SecurityMiddleware() gin.HandlerFunc {
    return gin.HandlerFunc(func(c *gin.Context) {
        c.Header("X-Frame-Options", "DENY")
        c.Header("X-Content-Type-Options", "nosniff")
        c.Header("X-XSS-Protection", "1; mode=block")
        c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        c.Next()
    })
}
```

### 防火墙配置

```bash
# UFW防火墙配置
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow ssh
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
```

## 备份策略

### 数据库备份

```bash
#!/bin/bash
# backup.sh

DB_HOST="localhost"
DB_USER="backup_user"
DB_PASS="backup_password"
DB_NAME="production_db"
BACKUP_DIR="/opt/backup/mysql"
DATE=$(date +%Y%m%d_%H%M%S)

# 创建备份目录
mkdir -p ${BACKUP_DIR}

# 执行备份
mysqldump -h${DB_HOST} -u${DB_USER} -p${DB_PASS} ${DB_NAME} | gzip > ${BACKUP_DIR}/backup_${DATE}.sql.gz

# 保留最近30天的备份
find ${BACKUP_DIR} -name "backup_*.sql.gz" -mtime +30 -delete

echo "Backup completed: backup_${DATE}.sql.gz"
```

### 文件备份

```bash
#!/bin/bash
# 应用文件备份
tar -czf /opt/backup/app_$(date +%Y%m%d).tar.gz /opt/yyhertz --exclude=/opt/yyhertz/logs
```

## 故障排查

### 常见问题

1. **应用启动失败**
   - 检查配置文件
   - 查看系统日志: `journalctl -u yyhertz.service`
   - 检查端口占用: `netstat -tlnp | grep :8080`

2. **数据库连接失败**
   - 检查数据库服务状态
   - 验证连接字符串
   - 检查防火墙设置

3. **内存泄漏**
   - 使用pprof分析: `go tool pprof http://localhost:8080/debug/pprof/heap`
   - 监控内存使用: `top -p $(pgrep app)`

### 健康检查

```go
func HealthCheck() gin.HandlerFunc {
    return gin.HandlerFunc(func(c *gin.Context) {
        // 数据库健康检查
        if err := db.Ping(); err != nil {
            c.JSON(503, gin.H{"status": "unhealthy", "database": "down"})
            return
        }
        
        // Redis健康检查
        if err := redisClient.Ping().Err(); err != nil {
            c.JSON(503, gin.H{"status": "unhealthy", "redis": "down"})
            return
        }
        
        c.JSON(200, gin.H{"status": "healthy"})
    })
}
```

---

合理的部署策略和运维实践是保证应用稳定运行的关键！根据实际需求选择合适的部署方案。