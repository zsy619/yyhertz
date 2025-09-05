# YYHertz 开发工具使用指南

<div align="center">

🛠️ **分析和调试工具使用指南** | 提升开发效率的利器

</div>

---

## 🔍 分析工具概览

YYHertz v2.0提供了完整的开发工具链，帮助开发者进行代码分析、性能监控、冲突检测等工作。

```
tools/
├── template_function_analyzer.go    # 模板函数分析器
├── conflict_detector.go             # 冲突检测工具  
├── performance_profiler.go          # 性能分析工具
├── version_checker.go               # 版本兼容性检查器
└── upgrade_assistant.go             # 升级助手工具
```

---

## 📊 模板函数分析器

### 1. 工具概述

模板函数分析器用于分析YYHertz项目中的模板函数使用情况，检测命名规范，发现潜在问题。

### 2. 使用方法

```bash
# 运行模板函数分析
go run tools/template_function_analyzer.go

# 指定分析路径
go run tools/template_function_analyzer.go -path ./views

# 输出详细报告
go run tools/template_function_analyzer.go -verbose

# 生成JSON报告
go run tools/template_function_analyzer.go -format json > analysis_report.json
```

### 3. 分析报告示例

```
=== YYHertz 模板函数分析报告 ===

📊 统计信息:
- 总函数数量: 152
- 字符串函数: 28
- 日期函数: 18  
- 数学函数: 15
- 逻辑函数: 12
- 其他函数: 79

🏷️ 命名规范分析:
✅ 符合规范: 144 (94.7%)
❌ 不符合规范: 8 (5.3%)

⚠️ 发现的问题:
1. 函数 'eq' 与Go内置冲突 - 建议使用 'equal'
2. 函数 'ne' 与Go内置冲突 - 建议使用 'notEqual'  
3. 函数 'lt' 与Go内置冲突 - 建议使用 'lessThan'

🔧 建议修复:
- 重命名冲突函数: 3个
- 添加缺失别名: 5个
- 统一命名风格: 2个

📈 覆盖率分析:
- 字符串处理: 完整 ✅
- 日期时间: 完整 ✅  
- 数学运算: 需补充 ⚠️
- 类型转换: 基本完整 ✅
```

### 4. 工具配置

```go
// tools/config/analyzer.json
{
    "template_paths": [
        "./views",
        "./templates", 
        "./example/views"
    ],
    "function_categories": {
        "string": ["substr", "upper", "lower", "trim"],
        "date": ["dateformat", "timeago", "now"],
        "math": ["add", "sub", "mul", "div"],
        "logic": ["and", "or", "not", "equal"]
    },
    "naming_rules": {
        "camelCase": true,
        "no_underscores": true,
        "max_length": 20
    },
    "go_builtins": ["and", "or", "not", "eq", "ne", "lt", "le", "gt", "ge"]
}
```

---

## ⚠️ 冲突检测工具

### 1. 工具功能

冲突检测工具专门用于检测YYHertz模板函数与Go内置动作的命名冲突，并提供修复建议。

### 2. 使用方法

```bash
# 运行冲突检测
go run tools/conflict_detector.go

# 检测特定文件
go run tools/conflict_detector.go -file beego_functions.go

# 自动修复模式（谨慎使用）
go run tools/conflict_detector.go -auto-fix

# 生成修复脚本
go run tools/conflict_detector.go -generate-script > fix_conflicts.sh
```

### 3. 检测结果示例

```
🔍 Go内置动作冲突检测报告

发现 13 个严重冲突:

❌ 关键冲突 (需立即修复):
1. 'and' 函数与Go内置 {{and}} 冲突
   - 位置: framework/mvc/view/beego_functions.go:45
   - 建议: 重命名为 'andFunc' 或使用别名
   - 影响: 模板渲染可能失败

2. 'eq' 函数与Go内置 {{eq}} 冲突  
   - 位置: framework/mvc/view/beego_functions.go:78
   - 建议: 使用安全别名 'equal'
   - 影响: 条件判断异常

⚠️ 潜在冲突 (建议修复):
3. 'len' 函数与Go内置 {{len}} 冲突
   - 建议: 使用别名 'length'
   - 影响: 可能导致混淆

🔧 推荐修复方案:
1. 立即重命名严重冲突的函数
2. 为常用函数添加安全别名  
3. 更新模板文件中的函数调用
4. 运行测试确保兼容性

📊 冲突统计:
- 严重冲突: 8个
- 潜在冲突: 5个
- 安全函数: 139个
- 修复进度: 0% (0/13)
```

### 4. 自动修复功能

```bash
# 生成修复补丁
go run tools/conflict_detector.go -generate-patch

# 应用修复补丁
go run tools/conflict_detector.go -apply-patch conflict_fixes.patch

# 验证修复结果
go run tools/conflict_detector.go -verify
```

---

## ⚡ 性能分析工具

### 1. 工具功能

性能分析工具提供全面的应用性能监控和分析，包括ORM性能、中间件耗时、内存使用等。

### 2. 使用方法

```bash
# 启动性能监控
go run tools/performance_profiler.go

# 分析特定组件
go run tools/performance_profiler.go -component orm
go run tools/performance_profiler.go -component middleware
go run tools/performance_profiler.go -component template

# 生成性能报告
go run tools/performance_profiler.go -report -duration 5m

# 实时监控模式
go run tools/performance_profiler.go -monitor -interval 10s
```

### 3. 性能报告示例

```
🚀 YYHertz 性能分析报告

📊 ORM性能测试 (持续5分钟):
┌─────────────────┬─────────────┬─────────────┬─────────────┐
│ 操作类型        │ 平均耗时    │ QPS         │ 成功率      │
├─────────────────┼─────────────┼─────────────┼─────────────┤
│ 简单查询(GORM)  │ 2.3ms       │ 16,278      │ 99.97%      │
│ 复杂查询(MyBatis)│ 15.8ms     │ 990         │ 99.85%      │  
│ 插入操作        │ 3.1ms       │ 12,450      │ 99.99%      │
│ 更新操作        │ 2.8ms       │ 13,890      │ 99.98%      │
│ 删除操作        │ 1.9ms       │ 18,200      │ 100.00%     │
└─────────────────┴─────────────┴─────────────┴─────────────┘

🔧 中间件性能分析:
┌─────────────────┬─────────────┬─────────────┬─────────────┐
│ 中间件名称      │ 平均耗时    │ 内存分配    │ 调用次数    │
├─────────────────┼─────────────┼─────────────┼─────────────┤
│ Recovery        │ 0.05ms      │ 128B        │ 25,680      │
│ Logger          │ 0.12ms      │ 256B        │ 25,680      │
│ CORS            │ 0.03ms      │ 64B         │ 25,680      │
│ Auth            │ 1.2ms       │ 1.2KB       │ 20,450      │
│ RateLimit       │ 0.08ms      │ 192B        │ 25,680      │
└─────────────────┴─────────────┴─────────────┴─────────────┘

💾 内存使用情况:
- 堆内存分配: 45.2 MB
- 栈内存使用: 8.7 MB  
- GC触发次数: 23次
- 平均GC耗时: 2.1ms

🎯 性能建议:
1. ✅ ORM性能良好，智能选择器工作正常
2. ⚠️ Auth中间件耗时较长，考虑缓存优化
3. ✅ 内存使用合理，GC压力较小
4. 🔧 建议启用连接池预热提升首次请求性能
```

### 4. 实时监控界面

```bash
# 启动Web监控界面
go run tools/performance_profiler.go -web -port 6060

# 访问 http://localhost:6060/debug/performance
```

监控界面显示：
- 实时QPS图表
- 内存使用趋势  
- 响应时间分布
- 错误率统计
- 热点接口排行

---

## ✅ 版本兼容性检查器

### 1. 工具功能

版本检查器用于检测当前代码与YYHertz不同版本的兼容性，帮助平滑升级。

### 2. 使用方法

```bash
# 检查当前版本兼容性
go run tools/version_checker.go

# 检查升级到指定版本的兼容性
go run tools/version_checker.go -target-version 2.0.0

# 生成兼容性报告
go run tools/version_checker.go -report compatibility_report.md

# 检查API变更
go run tools/version_checker.go -check-api-changes
```

### 3. 检查报告示例

```
📋 版本兼容性检查报告

当前版本: v1.9.5
目标版本: v2.0.0

🔄 API变更检测:
✅ 向后兼容的变更:
- 新增 orm.SmartSelector 智能选择器
- 新增 middleware.RateLimit 限流中间件
- 新增 pkg/xstring、pkg/xdate 等工具包

⚠️ 需要注意的变更:
- util 包已拆分为多个子包，但提供兼容层
- 部分模板函数重命名以避免冲突
- 中间件配置方式有轻微调整

❌ 破坏性变更:
- 无破坏性变更

🛠️ 升级步骤:
1. 更新依赖版本
2. 运行自动升级脚本
3. 更新import路径（可选）
4. 测试应用功能
5. 更新文档

📊 兼容性评分: 95/100 (优秀)
🎯 升级难度: 简单
⏱️ 预估升级时间: 30分钟
```

---

## 🚀 升级助手工具

### 1. 工具功能

升级助手提供自动化的版本升级支持，包括代码迁移、配置更新、测试验证等。

### 2. 使用方法

```bash
# 启动升级助手
go run tools/upgrade_assistant.go

# 预览升级内容
go run tools/upgrade_assistant.go -dry-run

# 自动升级
go run tools/upgrade_assistant.go -auto

# 回滚升级
go run tools/upgrade_assistant.go -rollback
```

### 3. 升级过程示例

```
🚀 YYHertz 升级助手 v2.0

检测到当前版本: v1.9.5
目标版本: v2.0.0

📋 升级计划:
1. ✅ 备份当前代码
2. ⏳ 更新依赖版本
3. ⏳ 迁移util包引用
4. ⏳ 修复模板函数冲突
5. ⏳ 更新中间件配置
6. ⏳ 运行兼容性测试
7. ⏳ 更新配置文件

是否继续? [y/N]: y

🔧 执行升级...

步骤 1/7: 备份当前代码
✅ 代码已备份至 backup/v1.9.5_20240905_153025/

步骤 2/7: 更新依赖版本
✅ go.mod 已更新
✅ 依赖下载完成

步骤 3/7: 迁移util包引用  
📝 发现 15 个文件需要更新import路径
✅ controllers/user.go - 已更新
✅ services/auth.go - 已更新
... (13个文件)
✅ 所有import路径已更新

步骤 4/7: 修复模板函数冲突
📝 发现 3 个模板文件使用冲突函数
✅ views/user/profile.html - {{eq}} → {{equal}}
✅ views/post/list.html - {{template}} → {{templatefunc}}
✅ views/common/header.html - {{len}} → {{length}}

步骤 5/7: 更新中间件配置
✅ 中间件注册方式已更新
✅ 配置参数已迁移

步骤 6/7: 运行兼容性测试
✅ 编译测试通过
✅ 单元测试通过 (125/125)
✅ 集成测试通过 (23/23)

步骤 7/7: 更新配置文件
✅ 配置文件已更新到v2.0格式

🎉 升级成功完成!

📊 升级总结:
- 用时: 3分25秒
- 更新文件: 18个
- 通过测试: 148个
- 发现问题: 0个

🔧 后续建议:
1. 运行应用并检查功能
2. 更新部署脚本
3. 培训团队新特性
4. 更新项目文档
```

### 4. 升级回滚

如果升级过程中出现问题，可以使用回滚功能：

```bash
# 查看可回滚的版本
go run tools/upgrade_assistant.go -list-backups

# 回滚到指定备份
go run tools/upgrade_assistant.go -rollback backup/v1.9.5_20240905_153025/

# 验证回滚结果
go run tools/upgrade_assistant.go -verify
```

---

## 🧪 工具集成测试

### 1. 工具测试套件

```bash
# 运行所有工具的测试
go test ./tools/...

# 测试特定工具
go test ./tools/template_function_analyzer/
go test ./tools/conflict_detector/
go test ./tools/performance_profiler/

# 生成测试覆盖率报告
go test -coverprofile=coverage.out ./tools/...
go tool cover -html=coverage.out -o coverage.html
```

### 2. CI/CD集成

```yaml
# .github/workflows/tools-validation.yml
name: Tools Validation

on:
  push:
    paths: ['tools/**']
  pull_request:
    paths: ['tools/**']

jobs:
  validate-tools:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v2
    
    - name: Set up Go
      uses: actions/setup-go@v2
      with:
        go-version: 1.21
    
    - name: Run template function analyzer
      run: go run tools/template_function_analyzer.go
      
    - name: Run conflict detector
      run: go run tools/conflict_detector.go
      
    - name: Run performance profiler test
      run: go run tools/performance_profiler.go -test-mode
      
    - name: Run version checker
      run: go run tools/version_checker.go
      
    - name: Test upgrade assistant
      run: go run tools/upgrade_assistant.go -dry-run
```

---

## 📝 工具开发指南

### 1. 创建新工具

```go
// tools/my_tool.go
package main

import (
    "flag"
    "fmt"
    "log"
    "os"
    
    "github.com/zsy619/yyhertz/tools/common"
)

var (
    verbose = flag.Bool("verbose", false, "Enable verbose output")
    output  = flag.String("output", "", "Output file path")
)

func main() {
    flag.Parse()
    
    fmt.Println("🛠️ YYHertz Custom Tool v1.0")
    
    if *verbose {
        log.SetFlags(log.LstdFlags | log.Lshortfile)
    }
    
    // 工具逻辑
    result, err := runAnalysis()
    if err != nil {
        log.Fatalf("Analysis failed: %v", err)
    }
    
    // 输出结果
    if *output != "" {
        err = common.WriteReport(*output, result)
        if err != nil {
            log.Fatalf("Failed to write report: %v", err)
        }
        fmt.Printf("✅ Report saved to %s\n", *output)
    } else {
        common.PrintReport(result)
    }
}

func runAnalysis() (*common.AnalysisResult, error) {
    // 实现分析逻辑
    return &common.AnalysisResult{
        Summary: "Analysis completed successfully",
        Issues:  []common.Issue{},
        Metrics: map[string]interface{}{},
    }, nil
}
```

### 2. 通用工具库

```go
// tools/common/report.go
package common

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
)

type AnalysisResult struct {
    Summary string                 `json:"summary"`
    Issues  []Issue               `json:"issues"`
    Metrics map[string]interface{} `json:"metrics"`
}

type Issue struct {
    Type        string `json:"type"`
    Severity    string `json:"severity"`
    Message     string `json:"message"`
    File        string `json:"file,omitempty"`
    Line        int    `json:"line,omitempty"`
    Suggestion  string `json:"suggestion,omitempty"`
}

func WriteReport(filename string, result *AnalysisResult) error {
    data, err := json.MarshalIndent(result, "", "  ")
    if err != nil {
        return err
    }
    return ioutil.WriteFile(filename, data, 0644)
}

func PrintReport(result *AnalysisResult) {
    fmt.Println("📊", result.Summary)
    
    if len(result.Issues) > 0 {
        fmt.Println("\n🔍 发现的问题:")
        for i, issue := range result.Issues {
            fmt.Printf("%d. [%s] %s\n", i+1, issue.Severity, issue.Message)
            if issue.File != "" {
                fmt.Printf("   位置: %s:%d\n", issue.File, issue.Line)
            }
            if issue.Suggestion != "" {
                fmt.Printf("   建议: %s\n", issue.Suggestion)
            }
        }
    }
    
    if len(result.Metrics) > 0 {
        fmt.Println("\n📈 统计数据:")
        for key, value := range result.Metrics {
            fmt.Printf("- %s: %v\n", key, value)
        }
    }
}
```

---

## ⚙️ 工具配置管理

### 1. 全局配置文件

```json
{
  "tools": {
    "template_analyzer": {
      "enabled": true,
      "scan_paths": ["./views", "./templates"],
      "ignore_patterns": ["*.backup", "*.tmp"],
      "rules": {
        "naming_convention": "camelCase",
        "max_function_length": 20
      }
    },
    "conflict_detector": {
      "enabled": true,
      "auto_fix": false,
      "backup_before_fix": true,
      "go_builtins": ["and", "or", "not", "eq", "ne", "lt", "le", "gt", "ge"]
    },
    "performance_profiler": {
      "enabled": true,
      "monitor_interval": "10s",
      "report_format": "table",
      "thresholds": {
        "response_time": "100ms",
        "memory_usage": "512MB",
        "error_rate": "1%"
      }
    }
  },
  "output": {
    "format": "console",
    "verbose": false,
    "log_file": "logs/tools.log"
  }
}
```

---

<div align="center">

**🛠️ YYHertz开发工具链让开发更高效，问题发现更及时！**

**善用这些工具，提升项目质量和开发体验 🚀**

</div>