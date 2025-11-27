# YYHertz 社区贡献指南

<div align="center">

👥 **参与开源，共建生态** | 让YYHertz变得更好

</div>

---

## 🤝 欢迎贡献

YYHertz是一个开源项目，我们欢迎所有形式的贡献！无论你是初学者还是专家，都可以为项目做出有价值的贡献。

## 🎯 贡献方式

### 📝 文档改进
- 修复文档中的错误和拼写问题
- 补充缺失的文档内容
- 翻译文档到其他语言
- 改进代码示例和教程

### 🐛 Bug报告和修复
- 报告发现的bug
- 提供bug的复现步骤
- 提交bug修复的Pull Request
- 参与bug讨论和验证

### ✨ 新功能开发
- 提出新功能建议
- 实现新的功能特性
- 优化现有功能
- 开发实用工具和插件

### 🧪 测试贡献
- 编写和改进测试用例
- 进行性能测试
- 参与用户体验测试
- 验证新版本的稳定性

---

## 🚀 快速开始

### 1. 准备开发环境

```bash
# 1. Fork项目到你的GitHub账号

# 2. 克隆你的fork
git clone https://github.com/YOUR_USERNAME/yyhertz.git
cd yyhertz

# 3. 添加上游仓库
git remote add upstream https://github.com/zsy619/yyhertz.git

# 4. 创建开发分支
git checkout -b feature/my-awesome-feature

# 5. 安装依赖
go mod tidy
```

### 2. 开发流程

```bash
# 1. 保持代码最新
git fetch upstream
git rebase upstream/main

# 2. 进行开发
# ... 编写代码 ...

# 3. 运行测试
go test ./...

# 4. 提交代码
git add .
git commit -m "feat: add awesome new feature"

# 5. 推送到你的fork
git push origin feature/my-awesome-feature

# 6. 创建Pull Request
```

---

## 📋 贡献标准

### 代码质量要求
- ✅ 遵循Go代码规范
- ✅ 添加必要的注释
- ✅ 编写相应的测试
- ✅ 确保所有测试通过
- ✅ 保持代码覆盖率

### 提交信息规范
```bash
# 使用Conventional Commits格式
feat: 添加新功能
fix: 修复bug
docs: 文档更新
style: 代码格式调整
refactor: 代码重构
test: 测试相关
chore: 构建或工具更新

# 示例
feat: add user authentication middleware
fix: resolve template function conflicts
docs: update API documentation
```

### Pull Request要求
- 🎯 清晰的PR标题和描述
- 📝 详细说明变更内容
- 🧪 包含相关的测试
- 📚 更新相关文档
- ✅ 通过所有CI检查

---

## 🏆 贡献者权益

### 🎖️ 贡献者认可
- 在README中列出贡献者名单
- 获得贡献者专属徽章
- 项目重大决策的参与权
- 优先获得技术支持

### 🎁 特别奖励
- 优秀贡献者可获得定制纪念品
- 有机会参与项目路线图制定
- 技术大会演讲机会推荐
- 开源社区推荐信

---

## 📞 联系我们

- 💬 **GitHub Discussions**: [项目讨论区](https://github.com/zsy619/yyhertz/discussions)
- 🐛 **Issue追踪**: [提交问题](https://github.com/zsy619/yyhertz/issues)
- 📧 **邮件联系**: contributors@yyhertz.com
- 💬 **QQ群**: 123456789 (YYHertz贡献者群)

---

## 📜 行为准则

我们承诺为每个人提供友好、安全和欢迎的环境。请阅读并遵守我们的[行为准则](CODE_OF_CONDUCT.md)。

---

<div align="center">

**🌟 感谢每一位贡献者让YYHertz变得更好！**

**一起构建优秀的Go Web框架生态 🚀**

</div>