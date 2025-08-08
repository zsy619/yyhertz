# 📊 监控告警

YYHertz应用的全面监控和告警体系搭建。

## Prometheus监控

### 指标收集

```go
package middleware

import (
    "time"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    httpRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "yyhertz_http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )

    httpRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "yyhertz_http_request_duration_seconds",
            Help: "Duration of HTTP requests",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "endpoint"},
    )

    dbConnectionsActive = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "yyhertz_db_connections_active",
        Help: "Number of active database connections",
    })
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

### Prometheus配置

```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'yyhertz'
    static_configs:
      - targets: ['yyhertz:8080']
    metrics_path: '/metrics'
    scrape_interval: 10s
```

## Grafana仪表板

### 关键指标仪表板

1. **应用性能指标**
   - HTTP请求率 (RPS)
   - 响应时间分布
   - 错误率统计
   - 并发连接数

2. **系统资源指标**
   - CPU使用率
   - 内存使用率
   - 磁盘I/O
   - 网络流量

3. **业务指标**
   - 用户注册数
   - 活跃用户数
   - 订单处理量
   - 收入统计

### 仪表板JSON配置

```json
{
  "dashboard": {
    "title": "YYHertz Application Metrics",
    "panels": [
      {
        "title": "HTTP Request Rate",
        "type": "stat",
        "targets": [
          {
            "expr": "rate(yyhertz_http_requests_total[5m])",
            "legendFormat": "{{method}} {{endpoint}}"
          }
        ]
      },
      {
        "title": "Response Time",
        "type": "histogram",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(yyhertz_http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "95th percentile"
          }
        ]
      }
    ]
  }
}
```

## 日志监控

### ELK Stack集成

#### Logstash配置

```ruby
input {
  beats {
    port => 5044
  }
}

filter {
  if [fields][app] == "yyhertz" {
    json {
      source => "message"
    }
    
    date {
      match => [ "timestamp", "ISO8601" ]
    }
    
    mutate {
      add_field => { "application" => "yyhertz" }
    }
  }
}

output {
  elasticsearch {
    hosts => ["elasticsearch:9200"]
    index => "yyhertz-%{+YYYY.MM.dd}"
  }
}
```

#### Filebeat配置

```yaml
filebeat.inputs:
- type: log
  enabled: true
  paths:
    - /var/log/yyhertz/*.log
  fields:
    app: yyhertz
  fields_under_root: true

output.logstash:
  hosts: ["logstash:5044"]
```

## 告警规则

### Prometheus告警

```yaml
# alerts.yml
groups:
- name: yyhertz.rules
  rules:
  - alert: HighErrorRate
    expr: rate(yyhertz_http_requests_total{status=~"5.."}[5m]) > 0.1
    for: 2m
    labels:
      severity: critical
    annotations:
      summary: "High error rate detected"
      description: "Error rate is {{ $value }} requests per second"

  - alert: HighResponseTime
    expr: histogram_quantile(0.95, rate(yyhertz_http_request_duration_seconds_bucket[5m])) > 1
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "High response time detected"
      description: "95th percentile response time is {{ $value }}s"

  - alert: ApplicationDown
    expr: up{job="yyhertz"} == 0
    for: 1m
    labels:
      severity: critical
    annotations:
      summary: "YYHertz application is down"
      description: "YYHertz application has been down for more than 1 minute"
```

### Alertmanager配置

```yaml
# alertmanager.yml
global:
  smtp_smarthost: 'smtp.gmail.com:587'
  smtp_from: 'alerts@yourdomain.com'

route:
  group_by: ['alertname']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'web.hook'

receivers:
- name: 'web.hook'
  email_configs:
  - to: 'admin@yourdomain.com'
    subject: 'YYHertz Alert: {{ .GroupLabels.alertname }}'
    body: |
      {{ range .Alerts }}
      Alert: {{ .Annotations.summary }}
      Description: {{ .Annotations.description }}
      {{ end }}
  
  slack_configs:
  - api_url: 'https://hooks.slack.com/services/...'
    channel: '#alerts'
    title: 'YYHertz Alert'
    text: '{{ range .Alerts }}{{ .Annotations.summary }}{{ end }}'
```

## 健康检查

### 应用健康检查

```go
func HealthCheckHandler(c *gin.Context) {
    health := map[string]interface{}{
        "status": "ok",
        "timestamp": time.Now(),
        "version": version.Version,
    }

    // 数据库健康检查
    if err := db.Ping(); err != nil {
        health["database"] = "unhealthy"
        health["status"] = "degraded"
    } else {
        health["database"] = "healthy"
    }

    // Redis健康检查
    if err := redisClient.Ping().Err(); err != nil {
        health["redis"] = "unhealthy"
        health["status"] = "degraded"
    } else {
        health["redis"] = "healthy"
    }

    status := 200
    if health["status"] == "degraded" {
        status = 503
    }

    c.JSON(status, health)
}
```

### Kubernetes探针

```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10
  failureThreshold: 3

readinessProbe:
  httpGet:
    path: /ready
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
  failureThreshold: 3
```

## 性能监控

### APM集成

```go
// Jaeger分布式追踪
import "github.com/opentracing/opentracing-go"

func TracingMiddleware() gin.HandlerFunc {
    return gin.HandlerFunc(func(c *gin.Context) {
        span := opentracing.StartSpan(c.Request.URL.Path)
        defer span.Finish()
        
        span.SetTag("http.method", c.Request.Method)
        span.SetTag("http.url", c.Request.URL.String())
        
        c.Set("tracing-span", span)
        c.Next()
        
        span.SetTag("http.status_code", c.Writer.Status())
    })
}
```

### 自定义指标

```go
// 业务指标收集
var (
    userRegistrations = promauto.NewCounter(prometheus.CounterOpts{
        Name: "yyhertz_user_registrations_total",
        Help: "Total number of user registrations",
    })
    
    ordersProcessed = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "yyhertz_orders_processed_total",
            Help: "Total number of orders processed",
        },
        []string{"status"},
    )
)

func (c *UserController) Register() {
    // 业务逻辑
    
    // 指标记录
    userRegistrations.Inc()
}
```

完整的监控告警体系，确保应用稳定运行！