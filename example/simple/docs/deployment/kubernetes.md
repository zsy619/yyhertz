# ☸️ Kubernetes部署

将YYHertz应用部署到Kubernetes集群的完整指南。

## 基础部署

### Deployment配置

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: yyhertz-app
  labels:
    app: yyhertz
spec:
  replicas: 3
  selector:
    matchLabels:
      app: yyhertz
  template:
    metadata:
      labels:
        app: yyhertz
    spec:
      containers:
      - name: app
        image: yyhertz:latest
        ports:
        - containerPort: 8080
        env:
        - name: ENV
          value: "production"
        - name: DB_HOST
          value: "mysql-service"
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

### Service配置

```yaml
apiVersion: v1
kind: Service
metadata:
  name: yyhertz-service
spec:
  selector:
    app: yyhertz
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: LoadBalancer
```

## 配置管理

### ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: yyhertz-config
data:
  app.yaml: |
    server:
      host: "0.0.0.0"
      port: 8080
    database:
      host: "mysql-service"
      port: 3306
      database: "yyhertz"
```

### Secret

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: yyhertz-secret
type: Opaque
data:
  db-password: cGFzc3dvcmQ=  # base64编码
  jwt-secret: c2VjcmV0X2tleQ==
```

## Ingress配置

### NGINX Ingress

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: yyhertz-ingress
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
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

## 自动伸缩

### HPA配置

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: yyhertz-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: yyhertz-app
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

## 持久化存储

### PVC配置

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: yyhertz-storage
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
```

## Helm Chart

使用Helm简化部署和管理。

### Chart.yaml

```yaml
apiVersion: v2
name: yyhertz
description: YYHertz Web Application
version: 0.1.0
appVersion: "1.0"
```

### values.yaml

```yaml
replicaCount: 3

image:
  repository: yyhertz
  pullPolicy: IfNotPresent
  tag: "latest"

service:
  type: LoadBalancer
  port: 80

ingress:
  enabled: true
  className: "nginx"
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
  hosts:
    - host: api.yourdomain.com
      paths:
        - path: /
          pathType: ImplementationSpecific

resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 250m
    memory: 256Mi

autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70
```

### 部署命令

```bash
# 安装
helm install yyhertz ./yyhertz-chart

# 升级
helm upgrade yyhertz ./yyhertz-chart

# 卸载
helm uninstall yyhertz
```

在Kubernetes上运行YYHertz，实现高可用和自动伸缩！
