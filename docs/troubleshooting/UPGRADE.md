# YYHertz v2.0 升级故障排除指南

<div align="center">

🔧 **升级过程中的问题解决方案** | 从v1.x到v2.0的完整故障排除

</div>

---

## 📋 目录

- [升级前准备](#升级前准备)
- [常见升级错误](#常见升级错误)
- [依赖冲突解决](#依赖冲突解决)
- [API破坏性变更](#api破坏性变更)
- [配置迁移问题](#配置迁移问题)
- [性能回归分析](#性能回归分析)
- [回滚操作](#回滚操作)

---

## 🎯 升级前准备

### 1. 环境检查清单

```bash
# 检查当前YYHertz版本
go list -m github.com/zsy619/yyhertz

# 检查Go版本兼容性
go version  # 要求 >= 1.19

# 检查项目依赖状态
go mod verify
go mod tidy

# 备份当前代码
git tag backup-before-upgrade-$(date +%Y%m%d)
git push origin backup-before-upgrade-$(date +%Y%m%d)
```

### 2. 升级前测试

```bash
# 运行完整测试套件
go test -v ./...

# 运行集成测试
make test-integration

# 检查API兼容性
make test-api-compatibility
```

---

## 🚨 常见升级错误

### 错误1: 模块导入失败

**错误信息:**
```
build github.com/zsy619/yyhertz/framework/mvc: cannot find module providing package
```

**解决方案:**
```bash
# 清理模块缓存
go clean -modcache

# 重新下载依赖
go mod download

# 如果问题持续，强制更新
go get -u github.com/zsy619/yyhertz@latest
go mod tidy
```

### 错误2: 编译失败 - 未知类型

**错误信息:**
```
undefined: mvc.BaseController
undefined: middleware.Recovery
```

**解决方案:**
```go
// 旧的导入方式 (v1.x)
import (
    "github.com/zsy619/yyhertz/mvc"
    "github.com/zsy619/yyhertz/middleware"
)

// 新的导入方式 (v2.0)
import (
    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/framework/mvc/middleware"
)
```

### 错误3: ORM引擎初始化失败

**错误信息:**
```
panic: ORM engine not initialized properly
```

**解决方案:**
```go
// v2.0 新的初始化方式
func initDatabase() {
    // 使用智能选择器
    selector, err := orm.NewSmartSelector(&orm.Config{
        Driver: "mysql",
        DSN:    "user:password@tcp(localhost:3306)/dbname",
    })
    if err != nil {
        panic(err)
    }
    
    // 注册到全局实例
    orm.SetDefaultSelector(selector)
}
```

### 错误4: 路由注册冲突

**错误信息:**
```
panic: route already exists: GET /user/index
```

**解决方案:**
```go
// 检查重复路由注册
app := mvc.HertzApp

// 避免重复注册同一个控制器
// app.AutoRouters(&UserController{}) // 错误：如果已经注册过
if !app.IsControllerRegistered(&UserController{}) {
    app.AutoRouters(&UserController{})
}

// 或使用条件注册
app.AutoRoutersIf(!production, &DebugController{})
```

---

## 🔄 依赖冲突解决

### 1. Go模块冲突

**问题:** 不同版本的依赖包冲突

**解决方案:**
```bash
# 查看依赖关系图
go mod graph | grep yyhertz

# 查看具体冲突
go mod why -m github.com/zsy619/yyhertz

# 强制使用特定版本
go mod edit -replace github.com/old/package=github.com/new/package@v2.0.0
go mod tidy
```

### 2. 间接依赖问题

**go.mod 修复示例:**
```go
module your-app

go 1.21

require (
    github.com/zsy619/yyhertz v2.0.0
    // 其他直接依赖
)

// 解决版本冲突
replace (
    github.com/old/conflicting-package => github.com/compatible/package v1.2.3
)

// 排除有问题的版本
exclude github.com/problematic/package v1.0.0
```

### 3. 第三方包兼容性

**常见冲突包及解决方案:**
```bash
# GORM版本冲突
go get gorm.io/gorm@latest
go get gorm.io/driver/mysql@latest

# Hertz版本冲突
go get github.com/cloudwego/hertz@latest

# 日志库冲突
go get github.com/sirupsen/logrus@latest
```

---

## 💥 API破坏性变更

### 1. 控制器方法签名变更

**v1.x 旧版本:**
```go
type UserController struct {
    beego.Controller
}

func (c *UserController) Get() {
    c.Data["json"] = map[string]string{"message": "hello"}
    c.ServeJSON()
}
```

**v2.0 新版本:**
```go
type UserController struct {
    mvc.BaseController
}

func (c *UserController) GetIndex() { // 注意方法名变更
    c.JSON(map[string]string{"message": "hello"})
}
```

### 2. 中间件接口变更

**v1.x 旧版本:**
```go
func AuthMiddleware(ctx *context.Context) {
    // 旧的中间件接口
    token := ctx.Input.Header("Authorization")
    if token == "" {
        ctx.Abort(401, "Unauthorized")
        return
    }
}
```

**v2.0 新版本:**
```go
func AuthMiddleware() mvc.HandlerFunc {
    return func(c *mvc.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.JSON(401, map[string]any{"error": "Unauthorized"})
            c.Abort()
            return
        }
        c.Next()
    }
}
```

### 3. 配置系统变更

**v1.x 旧版本:**
```go
// 使用 beego.AppConfig
port := beego.AppConfig.String("httpport")
dbHost := beego.AppConfig.String("db.host")
```

**v2.0 新版本:**
```go
import "github.com/zsy619/yyhertz/framework/config"

// 使用新的配置系统
port := config.GetInt("app.port")
dbHost := config.GetString("db.host")
```

---

## ⚙️ 配置迁移问题

### 1. 配置文件格式变更

**app.conf 迁移:**
```ini
# v1.x 格式
httpport = 8888
runmode = prod
db::host = localhost
db::port = 3306

# v2.0 格式
app.port = 8888
app.mode = production
db.host = localhost
db.port = 3306
```

### 2. 环境变量映射

**旧环境变量映射:**
```bash
export BEEGO_RUNMODE=prod
export BEEGO_HTTPPORT=8888
```

**新环境变量映射:**
```bash
export YYHERTZ_APP_MODE=production
export YYHERTZ_APP_PORT=8888
```

### 3. 数据库配置迁移

**迁移脚本:**
```bash
#!/bin/bash
# migrate-config.sh

# 备份旧配置
cp conf/app.conf conf/app.conf.backup

# 转换配置格式
sed -i 's/httpport/app.port/g' conf/app.conf
sed -i 's/runmode/app.mode/g' conf/app.conf
sed -i 's/db::/db./g' conf/app.conf

echo "配置文件迁移完成"
```

---

## 📉 性能回归分析

### 1. 性能基准测试

**升级前后对比:**
```bash
# 升级前性能基准
go test -bench=. -benchmem > benchmark-before.txt

# 升级后性能测试
go test -bench=. -benchmem > benchmark-after.txt

# 性能对比
benchcmp benchmark-before.txt benchmark-after.txt
```

### 2. 内存使用分析

```go
// 内存使用监控
func monitorMemoryUsage() {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    log.Printf("Memory Usage:")
    log.Printf("Alloc = %d KB", bToKb(m.Alloc))
    log.Printf("TotalAlloc = %d KB", bToKb(m.TotalAlloc))
    log.Printf("Sys = %d KB", bToKb(m.Sys))
    log.Printf("NumGC = %v\n", m.NumGC)
}

func bToKb(b uint64) uint64 {
    return b / 1024
}
```

### 3. HTTP性能测试

```bash
# 使用wrk进行压力测试
wrk -t12 -c400 -d30s http://localhost:8888/api/users

# 对比升级前后的结果
echo "升级前:" > performance-report.txt
wrk -t12 -c400 -d30s http://localhost:8888/api/users >> performance-report.txt

echo -e "\n升级后:" >> performance-report.txt  
wrk -t12 -c400 -d30s http://localhost:8888/api/users >> performance-report.txt
```

---

## 🔙 回滚操作

### 1. 快速回滚步骤

```bash
# 1. 停止应用服务
systemctl stop yyhertz-app

# 2. 切换到备份分支
git checkout backup-before-upgrade-$(date +%Y%m%d)

# 3. 恢复依赖版本
go mod edit -dropreplace github.com/zsy619/yyhertz
go get github.com/zsy619/yyhertz@v1.9.0  # 假设这是稳定版本
go mod tidy

# 4. 重新编译
go build -o app main.go

# 5. 恢复配置文件
cp conf/app.conf.backup conf/app.conf

# 6. 启动服务
systemctl start yyhertz-app
```

### 2. 数据库回滚

```sql
-- 如果有数据库迁移，需要回滚
-- 查看迁移历史
SELECT * FROM schema_migrations ORDER BY version DESC LIMIT 10;

-- 回滚到升级前的版本
-- 具体操作取决于你的迁移工具
migrate down 5  # 回滚5个版本
```

### 3. 生产环境回滚策略

**蓝绿部署回滚:**
```bash
# 切换负载均衡到绿色环境（旧版本）
kubectl patch service yyhertz-service -p '{"spec":{"selector":{"version":"v1"}}}'

# 验证服务正常
curl -f http://your-app.com/health

# 删除蓝色环境（新版本）
kubectl delete deployment yyhertz-app-v2
```

---

## 🔍 调试和诊断

### 1. 日志分析

```bash
# 查看升级相关日志
grep -i "upgrade\|migration\|version" /var/log/yyhertz/app.log

# 查看错误日志
grep -i "error\|panic\|fatal" /var/log/yyhertz/app.log | tail -50

# 实时监控日志
tail -f /var/log/yyhertz/app.log | grep -i "error"
```

### 2. 健康检查脚本

```bash
#!/bin/bash
# health-check.sh

APP_URL="http://localhost:8888"
HEALTH_ENDPOINT="/health"

echo "正在检查应用健康状态..."

# 检查HTTP响应
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" $APP_URL$HEALTH_ENDPOINT)

if [ $HTTP_CODE -eq 200 ]; then
    echo "✅ 应用健康检查通过"
    
    # 检查关键API
    API_CODE=$(curl -s -o /dev/null -w "%{http_code}" $APP_URL/api/users)
    if [ $API_CODE -eq 200 ]; then
        echo "✅ API接口正常"
    else
        echo "❌ API接口异常 (HTTP $API_CODE)"
        exit 1
    fi
else
    echo "❌ 应用健康检查失败 (HTTP $HTTP_CODE)"
    exit 1
fi

echo "🎉 所有检查通过，升级成功！"
```

### 3. 性能监控脚本

```go
package main

import (
    "fmt"
    "net/http"
    "time"
    "log"
)

func main() {
    endpoints := []string{
        "http://localhost:8888/health",
        "http://localhost:8888/api/users",
        "http://localhost:8888/api/products",
    }
    
    for _, endpoint := range endpoints {
        start := time.Now()
        resp, err := http.Get(endpoint)
        duration := time.Since(start)
        
        if err != nil {
            log.Printf("❌ %s: 错误 - %v", endpoint, err)
            continue
        }
        
        if resp.StatusCode == 200 {
            log.Printf("✅ %s: %v (响应时间: %v)", endpoint, resp.StatusCode, duration)
        } else {
            log.Printf("⚠️  %s: %v (响应时间: %v)", endpoint, resp.StatusCode, duration)
        }
        
        resp.Body.Close()
    }
}
```

---

## 📞 获取帮助

### 1. 官方支持渠道

- 📧 **技术支持**: support@yyhertz.com
- 💬 **技术交流群**: [QQ群链接]
- 📝 **GitHub Issues**: https://github.com/zsy619/yyhertz/issues
- 📖 **官方文档**: https://docs.yyhertz.com

### 2. 社区资源

- 🌟 **知识库**: https://wiki.yyhertz.com
- 🎥 **视频教程**: https://video.yyhertz.com
- 📚 **最佳实践**: https://practices.yyhertz.com

### 3. 紧急联系方式

```
如果遇到生产环境紧急问题：

1. 立即回滚到稳定版本
2. 发送邮件到 emergency@yyhertz.com
3. 在GitHub创建Priority高的Issue
4. 联系技术支持热线：400-XXX-XXXX
```

---

<div align="center">

**🔧 升级过程中遇到问题不要慌，按照指南逐步排查解决！**

**记住：备份是最好的保险，测试是最大的保障！🛡️**

</div>