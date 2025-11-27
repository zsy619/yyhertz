package view

// 重构后的render.go - 说明文档和引用
// 原本的render.go文件(2181行)已被重构拆分为以下功能明确的文件：
//
// 数据结构和接口：
// - render_data.go: 渲染数据结构(RenderData, MetaData, FlashData等)和CSRF接口
//
// 核心功能模块： 
// - render_core.go: 核心渲染功能(Render, RenderWithLayout等主要方法)
// - render_template.go: 模板管理(LoadTemplate, FindTemplateFile, GetTemplate等)
// - render_layout.go: 布局处理(布局嵌入、内容处理、区块处理等)
// - render_component.go: 组件渲染(组件加载、注册、渲染等)
//
// 扩展功能模块：
// - render_cache.go: 缓存管理(CacheKeyManager、缓存统计、清理优化等)
// - render_functions.go: 模板函数(formatDate、formatFileSize、safeHTML等工具函数)
// - render_preloader.go: 预加载机制(TemplatePreloader及相关功能)
// - render_stats.go: 统计分析(性能指标、健康检查、诊断信息等)
// - render_utils.go: 工具方法(HTML压缩、热重载、配置验证等)
//
// 重构优势：
// ✅ 单一职责：每个文件职责明确，便于维护
// ✅ 可读性提升：从2181行拆分为10个功能清晰的文件
// ✅ 模块化设计：功能模块化，便于测试和扩展
// ✅ 依赖清晰：避免循环依赖，依赖关系明确
// ✅ 向后兼容：保持所有公开API不变，现有代码无需修改
//
// 注意：TemplateEngine结构体定义在engine.go中，此文件不再重复定义