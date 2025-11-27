# YYHertz 生产部署指南

<div align="center">

🚀 **生产环境部署最佳实践** | Docker + Kubernetes + 云原生

</div>

---

## 📋 目录

- [部署准备](#部署准备)
- [Docker部署](#docker部署)
- [Kubernetes部署](#kubernetes部署)
- [云平台部署](#云平台部署)
- [负载均衡配置](#负载均衡配置)
- [监控与日志](#监控与日志)
- [安全配置](#安全配置)
- [性能优化](#性能优化)

---

## 🎯 部署准备

### 1. 环境检查清单

```bash
# 检查Go版本 (推荐1.21+)
go version

# 检查Docker版本
docker --version

# 检查Kubernetes集群
kubectl version --client

# 检查必要的端口
netstat -tlnp | grep :8888
```

### 2. 项目构建优化

```dockerfile
# 多阶段构建的Dockerfile示例
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制依赖文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用程序
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# 最终运行镜像
FROM alpine:latest

# 安装ca-certificates for SSL
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/main .
COPY --from=builder /app/views ./views
COPY --from=builder /app/static ./static
COPY --from=builder /app/config ./config

# 暴露端口
EXPOSE 8888

# 运行应用
CMD ["./main"]
```

### 3. 配置文件管理

```bash
# 生产环境配置文件 config/prod.conf
cat > config/prod.conf << 'EOF'
# 应用配置
app.name = yyhertz-prod
app.mode = production
app.host = 0.0.0.0
app.port = 8888

# 数据库配置
db.driver = mysql
db.host = db-cluster.example.com
db.port = 3306
db.name = yyhertz_prod
db.user = ${DB_USER}
db.password = ${DB_PASSWORD}
db.max_idle_conns = 10
db.max_open_conns = 100
db.conn_max_lifetime = 300s

# Redis配置
redis.host = redis-cluster.example.com
redis.port = 6379
redis.password = ${REDIS_PASSWORD}
redis.db = 0
redis.pool_size = 10

# 日志配置
log.level = warn
log.file = /var/log/yyhertz/app.log
log.max_size = 100
log.max_backups = 5
log.max_age = 30

# 性能配置
performance.read_timeout = 30s
performance.write_timeout = 30s
performance.max_header_bytes = 1048576
performance.enable_gzip = true

# 安全配置
security.enable_csrf = true
security.csrf_token_length = 32
security.session_timeout = 3600
security.max_login_attempts = 5
EOF
```

---

## 🐳 Docker部署

### 1. 基础Docker部署

```bash
# 构建Docker镜像
docker build -t yyhertz-app:latest .

# 运行容器
docker run -d \
  --name yyhertz-app \
  -p 8888:8888 \
  -v $(pwd)/logs:/var/log/yyhertz \
  -e DB_USER=root \
  -e DB_PASSWORD=password \
  -e REDIS_PASSWORD=redispass \
  --restart unless-stopped \
  yyhertz-app:latest
```

### 2. Docker Compose部署

```yaml
# docker-compose.yml
version: '3.8'

services:
  # YYHertz应用服务
  app:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8888:8888"
    environment:
      - DB_USER=${DB_USER}
      - DB_PASSWORD=${DB_PASSWORD}
      - REDIS_PASSWORD=${REDIS_PASSWORD}
    depends_on:
      - mysql
      - redis
    volumes:
      - ./logs:/var/log/yyhertz
      - ./config:/app/config
    restart: unless-stopped
    networks:
      - yyhertz-network

  # MySQL数据库
  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: ${DB_PASSWORD}
      MYSQL_DATABASE: yyhertz_prod
    ports:
      - "3306:3306"
    volumes:
      - mysql-data:/var/lib/mysql
      - ./init.sql:/docker-entrypoint-initdb.d/init.sql
    restart: unless-stopped
    networks:
      - yyhertz-network

  # Redis缓存
  redis:
    image: redis:7-alpine
    command: redis-server --requirepass ${REDIS_PASSWORD}
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data
    restart: unless-stopped
    networks:
      - yyhertz-network

  # Nginx负载均衡
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
    networks:
      - yyhertz-network

volumes:
  mysql-data:
  redis-data:

networks:
  yyhertz-network:
    driver: bridge
```

### 3. 环境变量配置

```bash
# .env文件
DB_USER=yyhertz_user
DB_PASSWORD=strong_password_here
REDIS_PASSWORD=redis_password_here

# 启动服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f app
```

---

## ☸️ Kubernetes部署

### 1. Kubernetes配置文件

```yaml
# k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: yyhertz-prod

---
# k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: yyhertz-config
  namespace: yyhertz-prod
data:
  app.conf: |
    app.name = yyhertz-k8s
    app.mode = production
    app.host = 0.0.0.0
    app.port = 8888
    
    db.driver = mysql
    db.host = mysql-service
    db.port = 3306
    db.name = yyhertz_prod
    
    redis.host = redis-service
    redis.port = 6379

---
# k8s/secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: yyhertz-secrets
  namespace: yyhertz-prod
type: Opaque
data:
  db-user: eXloZXJ0el91c2Vy  # base64 encoded
  db-password: c3Ryb25nX3Bhc3N3b3JkX2hlcmU=  # base64 encoded
  redis-password: cmVkaXNfcGFzc3dvcmRfaGVyZQ==  # base64 encoded

---
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: yyhertz-app
  namespace: yyhertz-prod
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
      - name: yyhertz-app
        image: yyhertz-app:latest
        ports:
        - containerPort: 8888
        env:
        - name: DB_USER
          valueFrom:
            secretKeyRef:
              name: yyhertz-secrets
              key: db-user
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: yyhertz-secrets
              key: db-password
        - name: REDIS_PASSWORD
          valueFrom:
            secretKeyRef:
              name: yyhertz-secrets
              key: redis-password
        volumeMounts:
        - name: config
          mountPath: /app/config
        - name: logs
          mountPath: /var/log/yyhertz
        livenessProbe:
          httpGet:
            path: /health
            port: 8888
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8888
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
      volumes:
      - name: config
        configMap:
          name: yyhertz-config
      - name: logs
        emptyDir: {}

---
# k8s/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: yyhertz-service
  namespace: yyhertz-prod
spec:
  selector:
    app: yyhertz-app
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8888
  type: ClusterIP

---
# k8s/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: yyhertz-ingress
  namespace: yyhertz-prod
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/rate-limit: "100"
    nginx.ingress.kubernetes.io/rate-limit-window: "1m"
spec:
  tls:
  - hosts:
    - api.yourdomain.com
    secretName: yyhertz-tls
  rules:
  - host: api.yourdomain.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: yyhertz-service
            port:
              number: 80
```

### 2. 部署脚本

```bash
#!/bin/bash
# deploy.sh

# 创建命名空间
kubectl apply -f k8s/namespace.yaml

# 部署配置和密钥
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/secret.yaml

# 部署应用
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
kubectl apply -f k8s/ingress.yaml

# 等待部署完成
kubectl rollout status deployment/yyhertz-app -n yyhertz-prod

# 检查Pod状态
kubectl get pods -n yyhertz-prod

# 查看服务状态
kubectl get svc -n yyhertz-prod

# 查看Ingress状态
kubectl get ingress -n yyhertz-prod
```

### 3. 数据库迁移Job

```yaml
# k8s/migration-job.yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: yyhertz-migration
  namespace: yyhertz-prod
spec:
  template:
    spec:
      containers:
      - name: migration
        image: yyhertz-app:latest
        command: ["./main", "migrate"]
        env:
        - name: DB_USER
          valueFrom:
            secretKeyRef:
              name: yyhertz-secrets
              key: db-user
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: yyhertz-secrets
              key: db-password
        volumeMounts:
        - name: config
          mountPath: /app/config
      volumes:
      - name: config
        configMap:
          name: yyhertz-config
      restartPolicy: Never
  backoffLimit: 3
```

---

## ☁️ 云平台部署

### 1. AWS ECS部署

```json
{
  "family": "yyhertz-app",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "512",
  "memory": "1024",
  "executionRoleArn": "arn:aws:iam::123456789012:role/ecsTaskExecutionRole",
  "taskRoleArn": "arn:aws:iam::123456789012:role/ecsTaskRole",
  "containerDefinitions": [
    {
      "name": "yyhertz-app",
      "image": "123456789012.dkr.ecr.us-west-2.amazonaws.com/yyhertz-app:latest",
      "portMappings": [
        {
          "containerPort": 8888,
          "protocol": "tcp"
        }
      ],
      "environment": [
        {
          "name": "APP_ENV",
          "value": "production"
        }
      ],
      "secrets": [
        {
          "name": "DB_PASSWORD",
          "valueFrom": "arn:aws:secretsmanager:us-west-2:123456789012:secret:prod/yyhertz/db-password"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/yyhertz-app",
          "awslogs-region": "us-west-2",
          "awslogs-stream-prefix": "ecs"
        }
      },
      "healthCheck": {
        "command": [
          "CMD-SHELL",
          "curl -f http://localhost:8888/health || exit 1"
        ],
        "interval": 30,
        "timeout": 5,
        "retries": 3,
        "startPeriod": 60
      }
    }
  ]
}
```

### 2. 阿里云ACK部署

```bash
# 使用阿里云容器服务
# 1. 构建镜像并推送到ACR
docker build -t registry.cn-hangzhou.aliyuncs.com/your-namespace/yyhertz-app:latest .
docker push registry.cn-hangzhou.aliyuncs.com/your-namespace/yyhertz-app:latest

# 2. 使用kubectl部署到ACK集群
kubectl apply -f k8s/
```

### 3. 腾讯云TKE部署

```yaml
# tke-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: yyhertz-app
  namespace: default
spec:
  replicas: 2
  selector:
    matchLabels:
      app: yyhertz-app
  template:
    metadata:
      labels:
        app: yyhertz-app
    spec:
      containers:
      - name: yyhertz-app
        image: ccr.ccs.tencentyun.com/your-namespace/yyhertz-app:latest
        ports:
        - containerPort: 8888
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
```

---

## ⚖️ 负载均衡配置

### 1. Nginx配置

```nginx
# nginx.conf
upstream yyhertz_backend {
    least_conn;
    server app1:8888 weight=3 max_fails=3 fail_timeout=30s;
    server app2:8888 weight=3 max_fails=3 fail_timeout=30s;
    server app3:8888 weight=2 max_fails=3 fail_timeout=30s;
}

server {
    listen 80;
    server_name api.yourdomain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name api.yourdomain.com;

    # SSL配置
    ssl_certificate /etc/nginx/ssl/cert.pem;
    ssl_certificate_key /etc/nginx/ssl/key.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512;
    ssl_prefer_server_ciphers off;

    # 安全头
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header X-XSS-Protection "1; mode=block";
    add_header Strict-Transport-Security "max-age=31536000; includeSubdomains; preload";

    # 限流配置
    limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
    limit_req zone=api burst=20 nodelay;

    # 反向代理
    location / {
        proxy_pass http://yyhertz_backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
        
        # 超时配置
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
        
        # 健康检查
        proxy_next_upstream error timeout invalid_header http_500 http_502 http_503 http_504;
    }

    # 静态文件缓存
    location ~* \.(css|js|png|jpg|jpeg|gif|ico|svg)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
        try_files $uri @backend;
    }

    location @backend {
        proxy_pass http://yyhertz_backend;
    }

    # 健康检查端点
    location /nginx-health {
        access_log off;
        return 200 "healthy\n";
        add_header Content-Type text/plain;
    }
}
```

### 2. HAProxy配置

```cfg
# haproxy.cfg
global
    daemon
    maxconn 4096
    log stdout local0

defaults
    mode http
    timeout connect 5000ms
    timeout client 50000ms
    timeout server 50000ms
    option httplog
    option dontlognull
    retries 3
    option redispatch

frontend yyhertz_frontend
    bind *:80
    redirect scheme https if !{ ssl_fc }
    bind *:443 ssl crt /etc/ssl/certs/yourdomain.com.pem
    default_backend yyhertz_backend

backend yyhertz_backend
    balance roundrobin
    option httpchk GET /health
    http-check expect status 200
    
    server app1 app1:8888 check inter 2000 rise 2 fall 3
    server app2 app2:8888 check inter 2000 rise 2 fall 3
    server app3 app3:8888 check inter 2000 rise 2 fall 3

listen stats
    bind *:8080
    stats enable
    stats uri /stats
    stats refresh 30s
    stats admin if TRUE
```

---

## 📊 监控与日志

### 1. Prometheus监控配置

```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'yyhertz-app'
    static_configs:
      - targets: ['app1:8888', 'app2:8888', 'app3:8888']
    metrics_path: /metrics
    scrape_interval: 5s

rule_files:
  - "yyhertz_rules.yml"

alerting:
  alertmanagers:
    - static_configs:
        - targets: ['alertmanager:9093']
```

### 2. Grafana仪表板配置

```json
{
  "dashboard": {
    "title": "YYHertz Application Metrics",
    "panels": [
      {
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total[5m])",
            "legendFormat": "{{instance}}"
          }
        ]
      },
      {
        "title": "Response Time",
        "type": "graph", 
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "95th percentile"
          }
        ]
      }
    ]
  }
}
```

### 3. 日志收集配置

```yaml
# filebeat.yml
filebeat.inputs:
- type: log
  enabled: true
  paths:
    - /var/log/yyhertz/*.log
  fields:
    service: yyhertz-app
    environment: production

output.elasticsearch:
  hosts: ["elasticsearch:9200"]
  index: "yyhertz-logs-%{+yyyy.MM.dd}"

setup.kibana:
  host: "kibana:5601"

logging.level: info
```

---

## 🔐 安全配置

### 1. HTTPS/TLS配置

```bash
# 使用Let's Encrypt获取SSL证书
certbot certonly --standalone -d api.yourdomain.com

# 自动续期
echo "0 12 * * * /usr/bin/certbot renew --quiet" | crontab -
```

### 2. 防火墙配置

```bash
# iptables规则
iptables -A INPUT -p tcp --dport 80 -j ACCEPT
iptables -A INPUT -p tcp --dport 443 -j ACCEPT
iptables -A INPUT -p tcp --dport 22 -s YOUR_ADMIN_IP -j ACCEPT
iptables -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT
iptables -P INPUT DROP

# 保存规则
iptables-save > /etc/iptables/rules.v4
```

### 3. 应用安全配置

```go
// 安全中间件配置
func setupSecurity(app *mvc.Application) {
    // HTTPS重定向
    app.Use(middleware.SecureHeaders())
    
    // CSRF保护
    app.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
        TokenLength:  32,
        CookieName:   "_csrf",
        CookieSecure: true,
        CookieHTTPOnly: true,
    }))
    
    // 速率限制
    app.Use(middleware.RateLimitWithConfig(middleware.RateLimitConfig{
        Rate:   100,
        Window: time.Minute,
        KeyFunc: func(c *mvc.Context) string {
            return c.ClientIP()
        },
    }))
    
    // 身份验证
    app.Use(middleware.JWTAuth(middleware.JWTConfig{
        SigningKey: []byte(os.Getenv("JWT_SECRET")),
        TokenLookup: "header:Authorization",
    }))
}
```

---

## ⚡ 性能优化

### 1. 应用层优化

```go
// 性能优化配置
func optimizePerformance(app *mvc.Application) {
    // 启用压缩
    app.Use(middleware.GzipWithConfig(middleware.GzipConfig{
        Level: middleware.BestCompression,
    }))
    
    // 缓存静态资源
    app.Use(middleware.CacheWithConfig(middleware.CacheConfig{
        TTL: 24 * time.Hour,
        KeyFunc: func(c *mvc.Context) string {
            return c.Request.URL.Path
        },
    }))
    
    // 连接池优化
    db := orm.GetDB()
    sqlDB, _ := db.DB()
    sqlDB.SetMaxIdleConns(10)
    sqlDB.SetMaxOpenConns(100)
    sqlDB.SetConnMaxLifetime(time.Hour)
}
```

### 2. 数据库优化

```sql
-- 数据库索引优化
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_status_created ON users(status, created_at);
CREATE INDEX idx_orders_user_id_status ON orders(user_id, status);

-- 查询优化
EXPLAIN SELECT * FROM users WHERE status = 1 ORDER BY created_at DESC LIMIT 10;
```

### 3. Redis缓存优化

```go
// Redis缓存配置
func setupRedis() *redis.Client {
    return redis.NewClient(&redis.Options{
        Addr:         os.Getenv("REDIS_ADDR"),
        Password:     os.Getenv("REDIS_PASSWORD"),
        DB:           0,
        PoolSize:     10,
        MinIdleConns: 2,
        MaxRetries:   3,
        DialTimeout:  5 * time.Second,
        ReadTimeout:  3 * time.Second,
        WriteTimeout: 3 * time.Second,
        IdleTimeout:  5 * time.Minute,
    })
}
```

---

## 🔄 自动化部署

### 1. CI/CD Pipeline

```yaml
# .github/workflows/deploy.yml
name: Deploy to Production

on:
  push:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v2
    - name: Setup Go
      uses: actions/setup-go@v2
      with:
        go-version: 1.21
    - name: Run tests
      run: go test ./...

  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v2
    - name: Build Docker image
      run: |
        docker build -t yyhertz-app:${{ github.sha }} .
        docker tag yyhertz-app:${{ github.sha }} yyhertz-app:latest
    - name: Push to registry
      run: |
        echo ${{ secrets.DOCKER_PASSWORD }} | docker login -u ${{ secrets.DOCKER_USERNAME }} --password-stdin
        docker push yyhertz-app:${{ github.sha }}
        docker push yyhertz-app:latest

  deploy:
    needs: build
    runs-on: ubuntu-latest
    steps:
    - name: Deploy to Kubernetes
      run: |
        echo ${{ secrets.KUBE_CONFIG }} | base64 -d > kubeconfig
        export KUBECONFIG=kubeconfig
        kubectl set image deployment/yyhertz-app yyhertz-app=yyhertz-app:${{ github.sha }} -n yyhertz-prod
        kubectl rollout status deployment/yyhertz-app -n yyhertz-prod
```

### 2. 滚动更新脚本

```bash
#!/bin/bash
# rolling-update.sh

IMAGE_TAG=${1:-latest}
NAMESPACE=${2:-yyhertz-prod}

echo "Starting rolling update to $IMAGE_TAG"

# 更新镜像
kubectl set image deployment/yyhertz-app yyhertz-app=yyhertz-app:$IMAGE_TAG -n $NAMESPACE

# 等待更新完成
kubectl rollout status deployment/yyhertz-app -n $NAMESPACE

# 检查健康状态
kubectl get pods -n $NAMESPACE -l app=yyhertz-app

echo "Rolling update completed successfully"
```

---

## 📋 部署检查清单

### 上线前检查

- [ ] 代码已通过所有测试
- [ ] 安全扫描通过
- [ ] 性能测试达标
- [ ] 数据库迁移脚本就绪
- [ ] 配置文件已更新
- [ ] SSL证书有效
- [ ] 监控告警已配置
- [ ] 备份策略已实施

### 上线后验证

- [ ] 服务正常启动
- [ ] 健康检查通过
- [ ] 日志输出正常
- [ ] 监控指标正常
- [ ] 数据库连接正常
- [ ] 缓存服务可用
- [ ] API接口响应正常
- [ ] 前端页面加载正常

---

<div align="center">

**🚀 遵循这些最佳实践，让你的YYHertz应用稳定运行在生产环境中！**

**生产部署不仅是技术问题，更是系统工程 🏗️**

</div>