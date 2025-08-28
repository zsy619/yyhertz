# ParamBinder 配置示例说明

这个目录包含了 ParamBinder 的各种配置示例，适用于不同的环境和使用场景。

## 配置文件说明

### 1. basic_config.yaml
**适用场景**: 基础项目、学习和测试
**特点**:
- 简单的配置选项
- 适中的性能设置
- 基础的错误处理
- 适合小型项目

### 2. development_config.yaml
**适用场景**: 开发环境
**特点**:
- 详细的调试信息
- 完整的错误堆栈
- 性能分析工具
- 热重载支持
- 测试数据生成

### 3. production_config.yaml
**适用场景**: 生产环境
**特点**:
- 高性能优化设置
- 完善的监控指标
- 错误采样和限流
- 内存和并发优化
- 安全性考虑

## 配置项说明

### 缓存设置 (cache_*)
```yaml
cache_enabled: true           # 是否启用缓存
cache_max_size: 1000         # 最大缓存条目数
cache_ttl: 3600              # 缓存生存时间（秒）
```

### 对象池设置 (object_pool_*)
```yaml
object_pool_enabled: true     # 启用对象池
object_pool_initial_size: 8   # 初始池大小
object_pool_max_size: 1000    # 最大池大小
```

### 转换器设置 (converters)
```yaml
converters:
  string:
    trim_spaces: true         # 自动去除字符串空格
    max_length: 1000          # 字符串最大长度
  numeric:
    strict_mode: false        # 数值转换严格模式
    overflow_check: true      # 检查数值溢出
  slice:
    separator: ","            # 切片分隔符
    max_elements: 100         # 最大元素数量
```

### 错误处理 (error_handling)
```yaml
error_handling:
  detailed_errors: true       # 详细错误信息
  log_conversion_errors: true # 记录转换错误
  stack_trace: false          # 包含堆栈跟踪
```

### 监控设置 (monitoring)
```yaml
monitoring:
  metrics_enabled: true       # 启用指标收集
  performance_tracking: true  # 性能跟踪
  cache_hit_ratio_alert: 0.8  # 缓存命中率告警
```

## 使用方法

### 1. Go 代码中使用配置

```go
import (
    "gopkg.in/yaml.v2"
    "github.com/zsy619/yyhertz/framework/mvc/routing"
)

type Config struct {
    ParamBinder struct {
        CacheEnabled bool `yaml:"cache_enabled"`
        CacheMaxSize int  `yaml:"cache_max_size"`
        // ... 其他配置
    } `yaml:"param_binder"`
}

func LoadConfig(configPath string) (*Config, error) {
    data, err := ioutil.ReadFile(configPath)
    if err != nil {
        return nil, err
    }
    
    var config Config
    err = yaml.Unmarshal(data, &config)
    return &config, err
}

func main() {
    config, _ := LoadConfig("config_examples/production_config.yaml")
    
    // 根据配置创建 ParamBinder
    pb := routing.NewParamBinder()
    // 应用配置设置...
}
```

### 2. 环境变量支持

可以结合环境变量使用：

```yaml
param_binder:
  cache_enabled: ${CACHE_ENABLED:true}
  cache_max_size: ${CACHE_MAX_SIZE:1000}
```

### 3. 配置验证

建议在应用启动时验证配置：

```go
func ValidateConfig(config *Config) error {
    if config.ParamBinder.CacheMaxSize <= 0 {
        return errors.New("cache_max_size must be positive")
    }
    // ... 其他验证
    return nil
}
```

## 配置最佳实践

### 1. 环境分离
- 开发环境：使用 development_config.yaml
- 测试环境：使用 basic_config.yaml  
- 生产环境：使用 production_config.yaml

### 2. 性能调优
- 根据实际负载调整缓存大小
- 监控缓存命中率，低于80%考虑增加缓存
- 根据内存使用情况调整对象池大小

### 3. 安全考虑
- 生产环境关闭详细错误信息
- 启用参数验证和类型检查
- 设置合适的限流参数

### 4. 监控配置
- 启用性能指标收集
- 设置适当的告警阈值
- 定期分析性能数据

## 扩展配置

可以根据具体需求添加自定义配置项：

```yaml
param_binder:
  # 自定义转换器配置
  custom_converters:
    uuid_converter:
      enabled: true
      validation: strict
    
    time_converter:
      enabled: true
      formats: 
        - "2006-01-02"
        - "2006-01-02T15:04:05Z"
  
  # 业务相关配置
  business_rules:
    max_page_size: 100
    default_timeout: 30
```

通过合理配置，ParamBinder 可以在各种环境下提供最优的性能和稳定性。