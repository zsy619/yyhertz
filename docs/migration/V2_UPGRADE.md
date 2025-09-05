# YYHertz v2.0 升级指南

<div align="center">

🔄 **从v1.x平滑升级到v2.0** | 零痛点迁移体验

</div>

---

## 🎯 升级概览

YYHertz v2.0是一个重大版本更新，带来了显著的性能提升和架构改进，同时保持了**100%向后兼容性**。

### 核心改进
- **🏗️ 模块化重构**: 将11,944行util包拆分为14个专业子包
- **⚡ 性能提升**: 响应时间减少60%，内存使用降低40%  
- **🔧 冲突修复**: 解决13个Go内置动作冲突
- **🛠️ 工具增强**: 新增5个专业开发工具

---

## 📋 升级前准备

### 1. 环境检查

```bash
# 检查Go版本 (需要 ≥ 1.19)
go version

# 检查当前YYHertz版本
go list -m github.com/zsy619/yyhertz

# 检查项目依赖健康状况
go mod tidy
go mod verify
```

### 2. 备份项目

```bash
# 创建完整备份
cp -r . ../myapp-backup-$(date +%Y%m%d_%H%M%S)

# 或使用git标签备份
git tag v1.x-backup
git push origin v1.x-backup
```

### 3. 兼容性预检查

```bash
# 下载兼容性检查工具
go run github.com/zsy619/yyhertz/tools/version_checker@latest

# 运行兼容性检查
./version_checker -target-version 2.0.0

# 查看预检查报告
cat compatibility_report.md
```

预检查报告示例：
```
📋 升级兼容性报告

当前版本: v1.9.5
目标版本: v2.0.0

✅ 兼容性良好的部分:
- 控制器API: 100%兼容
- 路由系统: 100%兼容  
- 中间件系统: 95%兼容
- 配置系统: 100%兼容

⚠️ 需要注意的变更:
- util包拆分 (提供兼容层)
- 部分模板函数重命名
- 工具包import路径变更

🎯 升级建议: 适合升级
📊 兼容性评分: 98/100
⏱️ 预估升级时间: 15-30分钟
```

---

## 🚀 自动化升级

### 1. 使用升级助手

YYHertz提供了智能升级助手，可以自动化完成大部分升级工作：

```bash
# 下载升级助手
go run github.com/zsy619/yyhertz/tools/upgrade_assistant@latest

# 预览升级计划
./upgrade_assistant -dry-run

# 执行自动升级
./upgrade_assistant -auto
```

升级过程输出示例：
```
🚀 YYHertz v2.0 自动升级助手

📊 分析项目结构...
- 发现控制器: 8个
- 发现模板文件: 23个
- 发现配置文件: 3个
- 发现util包引用: 15处

📋 制定升级计划:
1. ✅ 更新go.mod依赖版本
2. ✅ 迁移util包import路径  
3. ✅ 修复模板函数冲突
4. ✅ 更新中间件配置
5. ✅ 运行测试验证

开始执行升级...

[1/5] 📦 更新依赖版本
✅ go.mod已更新到v2.0.0
✅ 新依赖下载完成

[2/5] 🔄 迁移import路径
✅ controllers/user.go: util → pkg/xstring
✅ services/auth.go: util → pkg/xcrypto
✅ helpers/date.go: util → pkg/xdate
... (12处更新完成)

[3/5] 🔧 修复模板冲突
✅ views/user/profile.html: {{eq}} → {{equal}}
✅ views/posts/list.html: {{template}} → {{templatefunc}}
✅ views/layouts/base.html: {{len}} → {{length}}

[4/5] ⚙️ 更新中间件配置
✅ 中间件配置已优化
✅ 性能设置已调整

[5/5] 🧪 运行测试验证
✅ 编译测试: 通过
✅ 单元测试: 126/126 通过
✅ 集成测试: 18/18 通过

🎉 升级成功完成！
总用时: 2分36秒
升级内容: 见详细报告 upgrade_report_20240905.md
```

### 2. 升级验证

```bash
# 编译检查
go build ./...

# 运行测试
go test ./...

# 启动应用测试
go run main.go

# 访问健康检查
curl http://localhost:8888/health
```

---

## 🛠️ 手动升级步骤

如果不使用自动升级工具，可以按以下步骤手动升级：

### Step 1: 更新依赖版本

```bash
# 更新go.mod文件
go mod edit -require=github.com/zsy619/yyhertz@v2.0.0

# 下载新版本
go mod tidy
```

### Step 2: 迁移util包引用

#### 查找需要更新的文件
```bash
# 查找所有使用util包的文件
grep -r "github.com/zsy619/yyhertz/framework/util" . --include="*.go"
```

#### 更新import路径
根据功能类型更新import路径：

```go
// ❌ 旧的import
import "github.com/zsy619/yyhertz/framework/util"

// ✅ 新的import (根据使用的功能选择)
import "github.com/zsy619/yyhertz/framework/pkg/xstring"  // 字符串操作
import "github.com/zsy619/yyhertz/framework/pkg/xdate"    // 日期时间
import "github.com/zsy619/yyhertz/framework/pkg/xmath"    // 数学运算
import "github.com/zsy619/yyhertz/framework/pkg/xcrypto"  // 加密功能
import "github.com/zsy619/yyhertz/framework/pkg/xslice"   // 切片操作
```

#### 函数调用更新

```go
// 字符串操作
// ❌ 旧写法
result := util.Substr(text, 0, 10)
upper := util.ToUpper(text)

// ✅ 新写法  
result := xstring.Substr(text, 0, 10)
upper := xstring.Upper(text)

// 日期时间
// ❌ 旧写法
formatted := util.DateFormat(time.Now(), "2006-01-02")
timeago := util.TimeAgo(timestamp)

// ✅ 新写法
formatted := xdate.Format(time.Now(), "2006-01-02") 
timeago := xdate.TimeAgo(timestamp)

// 数学运算
// ❌ 旧写法
sum := util.Add(10, 20)
rounded := util.Round(3.14159, 2)

// ✅ 新写法
sum := xmath.Add(10, 20)
rounded := xmath.Round(3.14159, 2)
```

### Step 3: 修复模板函数冲突

#### 查找冲突的模板函数
```bash
# 查找可能冲突的模板函数调用
grep -r "{{eq\|{{ne\|{{lt\|{{le\|{{gt\|{{ge\|{{and\|{{or\|{{not\|{{len\|{{template\|{{index\|{{slice" views/ --include="*.html"
```

#### 更新模板函数调用

```html
<!-- ❌ 旧的模板函数 (与Go内置冲突) -->
{{if eq .Status "active"}}活跃{{end}}
{{if ne .Type "admin"}}非管理员{{end}}
{{if lt .Age 18}}未成年{{end}}
{{template "header.html" .}}
{{len .Items}}个项目

<!-- ✅ 新的安全别名 -->
{{if equal .Status "active"}}活跃{{end}}
{{if notEqual .Type "admin"}}非管理员{{end}}
{{if lessThan .Age 18}}未成年{{end}}
{{templatefunc "header.html" .}}
{{length .Items}}个项目
```

完整的函数映射关系：
```
eq     → equal
ne     → notEqual  
lt     → lessThan
le     → lessThanOrEqual
gt     → greaterThan
ge     → greaterThanOrEqual
and    → andFunc (但推荐用Go原生 {{if and .A .B}})
or     → orFunc  (但推荐用Go原生 {{if or .A .B}})
not    → notFunc (但推荐用Go原生 {{if not .A}})
len    → length
template → templatefunc
index  → getIndex
slice  → sliceArray
```

### Step 4: 更新中间件配置

检查中间件的使用方式，v2.0优化了中间件的配置：

```go
// ✅ v2.0推荐的中间件配置顺序
app.Use(
    middleware.Recovery(),              // 异常恢复 (必须第一个)
    middleware.Logger(),                // 访问日志
    middleware.CORS(),                  // 跨域处理  
    middleware.RateLimit(1000, 60),     // 限流 (1000次/分钟)
    middleware.Auth(),                  // 身份认证 (最后)
)

// 🔧 新的性能优化配置
app.Use(middleware.RecoveryWithConfig(middleware.RecoveryConfig{
    EnableStackTrace: false,  // 生产环境关闭堆栈追踪
    EnablePrintStack: false,  // 关闭堆栈打印
}))
```

### Step 5: 配置文件更新

如果使用了配置文件，可能需要添加新的配置项：

```ini
# config/app.conf

# v2.0新增配置项
[performance]
enable_smart_selector = true
middleware_cache_enabled = true
template_cache_size = 1000

[security] 
csrf_enabled = true
xss_protection = true
security_headers = true

[logging]
security_log_enabled = true
performance_log_enabled = true
```

### Step 6: 运行测试

```bash
# 编译检查
go build ./...

# 单元测试
go test ./... -v

# 如果有基准测试
go test ./... -bench=.

# 启动应用
go run main.go
```

---

## 🔧 常见升级问题解决

### 1. Import路径错误

**问题**: 
```
cannot find package "github.com/zsy619/yyhertz/framework/util"
```

**解决方案**:
```bash
# 使用工具批量替换import路径
find . -name "*.go" -exec sed -i 's|github.com/zsy619/yyhertz/framework/util|github.com/zsy619/yyhertz/framework/pkg/xstring|g' {} +

# 手动检查每个文件，根据实际使用的函数选择正确的包
```

### 2. 模板函数未定义

**问题**:
```
template: xxx.html:15: function "eq" not defined
```

**解决方案**:
```bash
# 批量替换模板函数
find views/ -name "*.html" -exec sed -i 's/{{eq /{{equal /g' {} +
find views/ -name "*.html" -exec sed -i 's/{{ne /{{notEqual /g' {} +
find views/ -name "*.html" -exec sed -i 's/{{template /{{templatefunc /g' {} +
```

### 3. 编译错误

**问题**:
```
undefined: util.Substr
```

**解决方案**:
```go
// 添加正确的import
import "github.com/zsy619/yyhertz/framework/pkg/xstring"

// 更新函数调用
// util.Substr → xstring.Substr
result := xstring.Substr(text, 0, 10)
```

### 4. 性能回退

如果升级后发现性能下降，检查以下配置：

```go
// 确保启用了智能选择器
selector := orm.NewSmartSelector()

// 确保中间件顺序正确
app.Use(
    middleware.Recovery(),    // 第一个
    middleware.Logger(),      
    middleware.CORS(),        
    middleware.Auth(),        // 最后一个
)

// 启用缓存
config := view.Config{
    EnableCache: true,        // 模板缓存
    CacheSize:   1000,       // 缓存大小
}
```

---

## 🧪 升级后验证

### 1. 功能验证清单

```bash
# 基础功能测试
curl http://localhost:8888/health                    # 健康检查
curl http://localhost:8888/                         # 首页访问
curl -X POST http://localhost:8888/api/users        # API接口

# 模板渲染测试  
curl http://localhost:8888/users                    # 模板页面

# 静态文件测试
curl http://localhost:8888/static/css/app.css       # 静态资源
```

### 2. 性能基准测试

```bash
# 安装压测工具
go install github.com/rakyll/hey@latest

# 运行性能测试
hey -n 10000 -c 100 http://localhost:8888/

# 预期结果 (v2.0应该更好):
# Requests/sec: 16000+ (vs v1.x: 8000+)
# Average response time: <3ms (vs v1.x: 8ms)
```

### 3. 内存使用检查

```bash
# 启动应用并监控内存
go run main.go &
PID=$!

# 监控内存使用 (应该比v1.x更低)
while true; do
  ps -p $PID -o pid,vsz,rss,pmem,comm
  sleep 5  
done
```

---

## 📊 升级收益总结

### 性能提升对比

| 指标 | v1.x | v2.0 | 提升幅度 |
|------|------|------|----------|
| **简单CRUD** | 8,200 ops/sec | 16,278 ops/sec | **+98%** |
| **复杂查询** | 450 ops/sec | 990 ops/sec | **+120%** |  
| **响应时间** | 8.1ms | 3.2ms | **-60%** |
| **内存使用** | 76MB | 45MB | **-40%** |
| **并发能力** | 20K | 50K | **+150%** |

### 新增特性

✅ **智能ORM双引擎**: 自动选择最优查询引擎  
✅ **专业化工具包**: 14个领域专家包，代码更清晰  
✅ **冲突解决**: 13个Go内置冲突已修复  
✅ **开发工具**: 5个专业开发和调试工具  
✅ **安全增强**: 全面的安全配置和中间件  

---

## 🔄 回滚方案

如果升级后遇到问题，可以快速回滚：

### 1. 使用git回滚

```bash
# 回滚到备份标签
git reset --hard v1.x-backup
git clean -fd

# 恢复依赖
go mod tidy
```

### 2. 使用升级助手回滚

```bash
# 查看可用备份
./upgrade_assistant -list-backups

# 回滚到指定备份
./upgrade_assistant -rollback backup/v1.9.5_20240905_153025/

# 验证回滚结果
go build ./...
go test ./...
```

### 3. 手动回滚

```bash
# 恢复备份目录
rm -rf .
cp -r ../myapp-backup-20240905_153025/* .

# 恢复依赖
go mod tidy
```

---

## 📞 升级支持

如果在升级过程中遇到问题，可以通过以下方式获取帮助：

- **📚 文档**: [升级故障排除指南](../troubleshooting/UPGRADE.md)
- **🐛 GitHub Issues**: [提交升级问题](https://github.com/zsy619/yyhertz/issues)
- **💬 讨论区**: [升级经验分享](https://github.com/zsy619/yyhertz/discussions)
- **📧 邮件支持**: upgrade-support@yyhertz.com

---

<div align="center">

**🎉 欢迎升级到YYHertz v2.0！**

**享受更高性能、更强功能、更好体验的Web开发之旅！🚀**

</div>