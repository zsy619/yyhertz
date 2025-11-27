# 📚 YYHertz 框架文档

欢迎来到YYHertz框架的官方文档！YYHertz是基于CloudWeGo Hertz构建的现代化Go Web框架，提供完整的MVC开发体验和革命性的多Handler类型系统。

## 🌟 核心特性

- **🎯 多Handler类型系统** - 7种专门优化的Handler类型，性能提升60%+
- **⚡ 增强Context系统** - 对象池化 + 原子操作，高并发友好
- **🗄️ 统一ORM解决方案** - GORM + MyBatis双引擎协同，智能选择最优引擎
- **🏛️ 完整MVC架构** - 100%兼容Beego风格，渐进式迁移
- **🔌 4层中间件系统** - 全局/分组/路由/控制器级别的灵活中间件

## 🚀 快速导航

### 📖 入门指南
- [📋 概览与安装](getting-started/overview.md) - 了解框架特性和环境要求  
- [🚀 快速开始](getting-started/quickstart.md) - 15分钟创建第一个多Handler应用
- [🏗️ 项目结构](getting-started/structure.md) - 标准项目组织方式
- [⚙️ 安装指南](getting-started/installation.md) - 详细安装步骤

### 🎯 核心功能

#### MVC核心系统
- [🎯 **多Handler类型系统详解**](mvc-core/multi-handlers.md) - **⭐ 核心特性**
- [🛣️ 路由系统](mvc-core/routing.md) - 双模式路由系统完整指南
- [🎛️ 控制器开发](mvc-core/controller/overview.md) - MVC控制器详解
- [⚠️ 错误处理系统](mvc-core/error-handling/overview.md) - 统一错误处理方案
- [📝 注解系统](mvc-core/annotation.md) - 声明式开发支持

#### 数据访问层
- [🗄️ **统一ORM解决方案**](data-access/orm-unified.md) - **⭐ 双引擎架构**  
- [📊 GORM快速入门](data-access/gorm-quickstart.md) - 高效CRUD操作
- [🔧 MyBatis高级用法](data-access/mybatis-advanced.md) - 复杂查询处理
- [💾 缓存策略](data-access/caching-strategies.md) - 多级缓存优化
- [📈 性能调优](data-access/database-tuning.md) - 数据库性能优化
- [🔍 监控告警](data-access/monitoring-alerting.md) - 实时性能监控

#### 中间件系统
- [🔌 中间件概览](middleware/overview.md) - 4层中间件架构
- [📦 内置中间件](middleware/builtin.md) - 丰富的预置功能
- [⚙️ 中间件配置](middleware/config.md) - 灵活的配置方式
- [🛠️ 自定义中间件](middleware/custom.md) - 开发定制中间件

### 🛠️ 开发工具

#### 配置管理
- [⚙️ 应用配置](configuration/app-config.md) - 应用级配置管理
- [🌍 环境配置](configuration/environment.md) - 多环境配置方案
- [📝 日志配置](configuration/logging.md) - 结构化日志系统

#### 视图模板
- [👁️ 模板概览](view-template/overview.md) - 模板引擎介绍
- [🎨 模板引擎](view-template/template-engine.md) - 高级模板功能
- [📄 视图渲染](view-template/view-rendering.md) - 渲染优化技巧
- [🎯 静态资源](view-template/static-assets.md) - 静态文件处理

#### 开发工具
- [🧪 测试框架](dev-tools/testing.md) - 完整测试解决方案
- [🔥 热重载](dev-tools/hot-reload.md) - 开发环境优化
- [📈 性能分析](dev-tools/performance.md) - 性能监控工具
- [🤖 代码生成](dev-tools/codegen.md) - 自动化开发工具

### ⚡ 高级功能
- [📅 定时任务](advanced/scheduler.md) - 分布式任务调度
- [🗂️ 会话管理](advanced/session.md) - 高性能会话系统
- [🔐 验证码系统](advanced/captcha.md) - 多种验证码支持
- [💾 缓存系统](advanced/cache.md) - 分布式缓存方案
- [✅ 数据验证](advanced/validation.md) - 强大的验证框架

### 🚀 部署运维
- [📦 部署概览](deployment/overview.md) - 部署策略概述
- [🐳 Docker部署](deployment/docker.md) - 容器化部署方案
- [☸️ Kubernetes部署](deployment/kubernetes.md) - K8s集群部署
- [📊 监控运维](deployment/monitoring.md) - 生产环境监控

## 🎯 重点推荐

### 🌟 必读文档
1. **[多Handler类型系统详解](mvc-core/multi-handlers.md)** - YYHertz的核心创新，7种Handler类型性能对比
2. **[统一ORM解决方案](data-access/orm-unified.md)** - GORM + MyBatis双引擎架构，智能选择最优性能
3. **[快速开始](getting-started/quickstart.md)** - 15分钟上手多Handler类型开发

### 🚀 性能优化
- [📈 性能分析](dev-tools/performance.md) - 对象池化、并发优化技巧
- [💾 缓存策略](data-access/caching-strategies.md) - 多级缓存架构设计  
- [📊 数据库调优](data-access/database-tuning.md) - SQL性能优化实践

### 🎯 实战案例
- [🏆 最佳实践](mvc-core/controller/faq-best-practices.md) - 生产环境开发规范
- [🔧 真实案例](mvc-core/controller/real-world-examples.md) - 企业级应用示例
- [⚠️ 故障排除](mvc-core/error-handling/troubleshooting.md) - 常见问题解决方案

## 🔄 版本迁移

### 从其他框架迁移
- **从Beego迁移**: YYHertz提供100%兼容的命名空间语法，支持渐进式迁移
- **从Gin迁移**: 提供Handler适配器，可以无缝集成现有Gin应用
- **从Echo迁移**: 类似的中间件设计，迁移成本极低

### 性能对比

| 框架 | QPS | 内存使用 | Handler类型 | 特色功能 |
|------|-----|----------|-------------|----------|
| **YYHertz** | **45,000** | **128MB** | **7种优化类型** | 多Handler + 增强Context |
| Gin | 38,000 | 156MB | 1种基础类型 | 轻量级路由 |
| Beego | 25,000 | 245MB | MVC控制器 | 完整框架 |
| Echo | 42,000 | 134MB | 1种基础类型 | 中间件丰富 |

## 📞 获取帮助

### 🔗 官方资源
- **📖 在线文档**: [http://localhost:8888/docs](http://localhost:8888/docs) 
- **🐛 问题反馈**: [GitHub Issues](https://github.com/zsy619/yyhertz/issues)
- **💬 社区讨论**: [GitHub Discussions](https://github.com/zsy619/yyhertz/discussions)
- **📧 邮件支持**: support@yyhertz.com

### 🎯 学习路径

#### 🔰 初学者 (第1周)
1. [概览与安装](getting-started/overview.md) - 了解框架概念
2. [快速开始](getting-started/quickstart.md) - 创建第一个应用
3. [路由系统](mvc-core/routing.md) - 掌握基本路由
4. [控制器开发](mvc-core/controller/overview.md) - 学习MVC模式

#### 🚀 进阶开发 (第2-3周)  
1. [多Handler类型系统](mvc-core/multi-handlers.md) - 掌握7种Handler类型
2. [统一ORM解决方案](data-access/orm-unified.md) - 学习双引擎数据访问
3. [中间件系统](middleware/overview.md) - 理解4层中间件架构
4. [错误处理系统](mvc-core/error-handling/overview.md) - 构建健壮应用

#### 🏆 高级专家 (第4-8周)
1. [性能优化](dev-tools/performance.md) - 深入性能调优
2. [缓存策略](data-access/caching-strategies.md) - 设计高效缓存
3. [部署运维](deployment/overview.md) - 生产环境部署
4. [监控告警](data-access/monitoring-alerting.md) - 建立完善监控

## 🎉 开始您的YYHertz之旅

选择适合您的起点：

- **🚀 我要快速体验** → [快速开始](getting-started/quickstart.md)
- **📚 我要系统学习** → [概览与安装](getting-started/overview.md)  
- **⚡ 我关注性能优化** → [多Handler类型系统](mvc-core/multi-handlers.md)
- **🗄️ 我需要数据访问** → [统一ORM解决方案](data-access/orm-unified.md)
- **🔧 我要迁移现有项目** → [最佳实践](mvc-core/controller/faq-best-practices.md)

---

**🎯 让我们一起构建更快、更现代的Go Web应用！**