# YYHertz 故障排除指南

<div align="center">

🔍 **常见问题解决方案** | 快速诊断和解决开发问题

</div>

---

## 🚨 常见问题分类

### 安装和环境问题
- [Go版本兼容性问题](#go版本兼容性问题)
- [依赖安装失败](#依赖安装失败)
- [编译错误](#编译错误)

### 运行时问题
- [应用启动失败](#应用启动失败)
- [数据库连接问题](#数据库连接问题)
- [模板渲染错误](#模板渲染错误)

### 性能问题
- [响应速度慢](#响应速度慢)
- [内存使用过高](#内存使用过高)
- [数据库查询慢](#数据库查询慢)

### 升级问题
- [v2.0升级问题](#v20升级问题)
- [包引用错误](#包引用错误)
- [模板函数冲突](#模板函数冲突)

---

## 🔧 问题诊断步骤

### 1. 检查环境信息
```bash
# 检查Go版本
go version

# 检查环境变量
echo $GOPATH
echo $GOROOT

# 检查模块状态
go mod tidy
go mod verify
```

### 2. 查看详细错误信息
```bash
# 启用详细日志
export YYH_DEBUG=true
export YYH_LOG_LEVEL=debug

# 运行应用查看详细错误
go run main.go
```

### 3. 检查依赖状态
```bash
# 查看依赖树
go mod graph

# 检查YYHertz版本
go list -m github.com/zsy619/yyhertz
```

---

## ❗ 具体问题解决

### Go版本兼容性问题

**问题症状：**
```
go: module requires Go 1.19 or higher
```

**解决方案：**
```bash
# 升级Go版本到1.19+
# macOS
brew install go

# Ubuntu
sudo apt update
sudo apt install golang-go

# 验证版本
go version
```

### 依赖安装失败

**问题症状：**
```
go: github.com/zsy619/yyhertz@v2.0.0: module not found
```

**解决方案：**
```bash
# 清理模块缓存
go clean -modcache

# 设置代理
export GOPROXY=https://goproxy.cn,direct

# 重新下载依赖
go mod download
```

### 应用启动失败

**问题症状：**
```
panic: failed to connect database
```

**解决方案：**
```bash
# 1. 检查数据库配置
cat config/app.conf

# 2. 测试数据库连接
mysql -h localhost -u root -p

# 3. 检查配置文件权限
ls -la config/

# 4. 使用默认配置测试
cp config/app.conf.example config/app.conf
```

### 模板渲染错误

**问题症状：**
```
template: views/user.html:15: function "eq" not defined
```

**解决方案：**
```bash
# 1. 使用冲突检测工具
go run tools/conflict_detector.go

# 2. 批量替换冲突函数
find views/ -name "*.html" -exec sed -i 's/{{eq /{{equal /g' {} +

# 3. 验证修复结果
go run tools/conflict_detector.go -verify
```

### v2.0升级问题

**问题症状：**
```
cannot find package "github.com/zsy619/yyhertz/framework/util"
```

**解决方案：**
```bash
# 1. 运行升级助手
go run tools/upgrade_assistant.go

# 2. 手动更新import路径
# 查找需要更新的文件
grep -r "framework/util" . --include="*.go"

# 3. 批量替换import路径
find . -name "*.go" -exec sed -i 's|framework/util|framework/pkg/xstring|g' {} +
```

---

## 🧪 调试技巧

### 1. 启用详细日志
```go
// 在main.go中添加
import "github.com/sirupsen/logrus"

func init() {
    logrus.SetLevel(logrus.DebugLevel)
    logrus.SetFormatter(&logrus.JSONFormatter{})
}
```

### 2. 使用pprof性能分析
```go
import _ "net/http/pprof"

go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()
```

### 3. 数据库查询调试
```go
// 启用SQL日志
db := orm.GetDB()
db.Logger = logger.New(
    log.New(os.Stdout, "\r\n", log.LstdFlags),
    logger.Config{
        SlowThreshold: time.Second,
        LogLevel:      logger.Info,
        Colorful:      true,
    },
)
```

---

## 📞 获取帮助

如果问题仍未解决：
- 📚 查看[完整文档](../README.md)
- 🐛 [提交Issue](https://github.com/zsy619/yyhertz/issues)
- 💬 [参与讨论](https://github.com/zsy619/yyhertz/discussions)
- 📧 发送邮件：support@yyhertz.com

---

<div align="center">

**🔍 遇到问题不要慌，按步骤排查很重要！**

</div>