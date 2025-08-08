# 📦 安装部署

YYHertz框架提供了多种安装和部署方式，适合不同的开发环境和部署需求。本文档将指导您完成框架的安装、配置和部署。

## 🛠️ 环境要求

### 基础要求
- **Go版本**: 1.19+ (推荐1.21+)
- **操作系统**: Linux, macOS, Windows
- **内存**: 最低512MB RAM (推荐2GB+)
- **磁盘空间**: 最低100MB可用空间
- **网络**: 需要访问Go模块代理 (如Go Proxy)

### 开发环境推荐
- **IDE**: VS Code + Go插件 / GoLand / Vim + vim-go
- **工具**: Git, Make, Docker (可选)
- **数据库**: MySQL 8.0+ / PostgreSQL 13+ / SQLite 3.36+
- **缓存**: Redis 6.0+ (可选)

### 验证环境
```bash
# 检查Go版本
go version
# 输出: go version go1.21.0 linux/amd64

# 检查Go环境变量
go env GOPATH GOROOT GOPROXY

# 验证网络连接
go list -m golang.org/x/text
```

## 📥 安装方式

### 方式一：使用go get安装 (推荐)

这是最简单和推荐的安装方式：

```bash
# 1. 创建新项目
mkdir my-yyhertz-app
cd my-yyhertz-app

# 2. 初始化Go模块
go mod init my-yyhertz-app

# 3. 安装YYHertz框架
go get -u github.com/zsy619/yyhertz

# 4. 验证安装
go list -m github.com/zsy619/yyhertz
# 输出: github.com/zsy619/yyhertz v1.4.0

# 5. 创建主文件
cat > main.go << 'EOF'
package main

import (
    "github.com/zsy619/yyhertz/framework/mvc"
)

func main() {
    app := mvc.HertzApp
    app.GET("/", func(c *mvc.Context) {
        c.JSON(map[string]string{"message": "Hello YYHertz!"})
    })
    app.Run(":8080")
}
EOF

# 6. 运行测试
go run main.go
```

### 方式二：从源码编译

适用于需要自定义框架或参与开发的场景：

```bash
# 1. 克隆源码
git clone https://github.com/zsy619/yyhertz.git
cd yyhertz

# 2. 查看版本和分支
git tag -l | tail -5
git checkout v1.4.0  # 使用稳定版本

# 3. 编译框架
go mod tidy
go build ./...

# 4. 运行测试
go test ./...

# 5. 安装到GOPATH (可选)
go install ./...

# 6. 在项目中使用本地版本
cd ../my-project
go mod init my-project
go mod edit -replace github.com/zsy619/yyhertz=../yyhertz
go mod tidy
```

### 方式三：使用项目模板

快速创建标准项目结构：

```bash
# 1. 克隆官方模板
git clone https://github.com/zsy619/yyhertz-template.git my-app
cd my-app

# 2. 清理git历史
rm -rf .git
git init

# 3. 重新初始化模块
rm go.mod go.sum
go mod init my-app

# 4. 安装依赖
go mod tidy

# 5. 配置项目信息
# 编辑 config/app.yaml 中的应用信息
sed -i 's/yyhertz-template/my-app/g' config/app.yaml

# 6. 启动开发服务器
make dev
```

### 方式四：使用Docker

适用于容器化部署：

```bash
# 1. 创建Dockerfile
cat > Dockerfile << 'EOF'
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/views ./views
COPY --from=builder /app/static ./static
COPY --from=builder /app/config ./config

EXPOSE 8080
CMD ["./main"]
EOF

# 2. 构建镜像
docker build -t my-yyhertz-app .

# 3. 运行容器
docker run -p 8080:8080 my-yyhertz-app
```

## ⚙️ 配置管理

### 基础配置结构

YYHertz使用YAML格式的配置文件：

```
config/
├── app.yaml          # 应用基础配置
├── database.yaml     # 数据库配置
├── middleware.yaml   # 中间件配置
├── template.yaml     # 模板引擎配置
├── log.yaml         # 日志配置
└── redis.yaml       # 缓存配置
```

### 应用配置 (config/app.yaml)

```yaml
app:
  name: "my-yyhertz-app"
  version: "1.0.0"
  description: "YYHertz应用"
  debug: true
  
server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: "30s"
  write_timeout: "30s"
  idle_timeout: "60s"
  max_header_bytes: 1048576

security:
  secret_key: "your-secret-key-here"
  jwt_secret: "your-jwt-secret-here"
  csrf_enabled: true
  cors_enabled: true

performance:
  max_multipart_memory: 33554432  # 32MB
  enable_gzip: true
  gzip_level: 6
  
development:
  hot_reload: true
  auto_build: false
  profiler_enabled: true
```

### 数据库配置 (config/database.yaml)

```yaml
database:
  driver: "mysql"
  dsn: "root:password@tcp(localhost:3306)/yyhertz?charset=utf8mb4&parseTime=True&loc=Local"
  max_open_conns: 100
  max_idle_conns: 10
  conn_max_lifetime: "1h"
  conn_max_idle_time: "30m"
  
  # 开发环境配置
  debug_sql: true
  log_level: "info"
  slow_threshold: "200ms"
  
  # 生产环境优化
  prepare_stmt: true
  disable_automatic_ping: false
```

### 环境变量支持

YYHertz支持通过环境变量覆盖配置：

```bash
# 应用配置
export YYHERTZ_APP_DEBUG=false
export YYHERTZ_SERVER_PORT=9000
export YYHERTZ_SERVER_HOST="127.0.0.1"

# 数据库配置
export YYHERTZ_DB_DSN="user:pass@tcp(db:3306)/app"
export YYHERTZ_DB_MAX_OPEN_CONNS=200

# 安全配置
export YYHERTZ_SECRET_KEY="production-secret-key"
export YYHERTZ_JWT_SECRET="production-jwt-secret"

# 启动应用
go run main.go
```

## 🚀 部署方案

### 单机部署

适用于小型应用和开发环境：

```bash
# 1. 编译生产版本
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

# 2. 创建部署目录结构
mkdir -p /opt/my-app/{bin,config,logs,static,views}

# 3. 复制文件
cp app /opt/my-app/bin/
cp -r config/* /opt/my-app/config/
cp -r static/* /opt/my-app/static/
cp -r views/* /opt/my-app/views/

# 4. 创建systemd服务
cat > /etc/systemd/system/my-app.service << 'EOF'
[Unit]
Description=My YYHertz App
After=network.target

[Service]
Type=simple
User=app
WorkingDirectory=/opt/my-app
ExecStart=/opt/my-app/bin/app
Restart=on-failure
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

# 5. 启动服务
systemctl daemon-reload
systemctl enable my-app
systemctl start my-app
```

### Docker部署

使用Docker Compose进行完整部署：

```yaml
# docker-compose.yml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - YYHERTZ_DB_DSN=root:password@tcp(mysql:3306)/app
      - YYHERTZ_REDIS_ADDR=redis:6379
    depends_on:
      - mysql
      - redis
    restart: unless-stopped
    
  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: password
      MYSQL_DATABASE: app
    volumes:
      - mysql_data:/var/lib/mysql
    restart: unless-stopped
    
  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data
    restart: unless-stopped
    
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/nginx/ssl
    depends_on:
      - app
    restart: unless-stopped

volumes:
  mysql_data:
  redis_data:
```

### Kubernetes部署

使用K8s进行大规模部署：

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: yyhertz-app
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
        image: my-yyhertz-app:latest
        ports:
        - containerPort: 8080
        env:
        - name: YYHERTZ_DB_DSN
          valueFrom:
            secretKeyRef:
              name: app-secrets
              key: database-dsn
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
  name: yyhertz-service
spec:
  selector:
    app: yyhertz-app
  ports:
  - port: 80
    targetPort: 8080
  type: LoadBalancer
```

## 🔧 故障排查

### 常见问题

**Q: 启动时提示"端口被占用"**

```bash
# 检查端口占用
lsof -i :8080
netstat -tulpn | grep :8080

# 杀死占用进程
kill -9 <PID>

# 或修改配置端口
export YYHERTZ_SERVER_PORT=8081
```

**Q: 数据库连接失败**

```bash
# 检查数据库连接
mysql -h localhost -u root -p

# 测试网络连通性
telnet localhost 3306
ping mysql-server

# 检查配置文件
cat config/database.yaml
```

**Q: 静态文件404错误**

```bash
# 检查文件路径
ls -la static/
ls -la views/

# 检查配置
grep -i "static\|view" config/app.yaml

# 确认工作目录
pwd
echo $PWD
```

### 性能调优

```bash
# 启用性能分析
go tool pprof http://localhost:8080/debug/pprof/profile

# 内存分析
go tool pprof http://localhost:8080/debug/pprof/heap

# 查看goroutine
go tool pprof http://localhost:8080/debug/pprof/goroutine

# 系统监控
htop
iostat -x 1
netstat -i
```

## 📈 生产优化

### 编译优化

```bash
# 优化编译参数
go build -ldflags="-s -w" -o app

# 压缩可执行文件
upx --best app

# 交叉编译
GOOS=linux GOARCH=amd64 go build -o app-linux
```

### 运行时优化

```bash
# 设置Go运行时参数
export GOMAXPROCS=4
export GOGC=100
export GODEBUG=gctrace=1

# 系统优化
ulimit -n 65536
sysctl net.core.somaxconn=65535
```

---

**🎉 恭喜！您已成功完成YYHertz框架的安装和部署配置！**

现在您可以继续阅读其他文档，深入了解框架的各种功能和最佳实践。