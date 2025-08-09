## 2508-0001 数据库-mybatis

目录 @framework/mybatis 是基于java mybatis框架的二次封装，可以简单、方便、快速的完成基本的增删改查功能。
请参考 https://github.com/gobatis/gobatis 项目进行对比与优化。

```kimi
你是一名资深 Java 架构师。下面有两份输入：

1. 我的项目：目录 @framework/mybatis  
   作用：基于原生 MyBatis 的二次封装，目标是用最简单的方式完成日常 CRUD。

2. 参考项目：Gobatis（Golang 版 MyBatis 风格框架）  
   仓库：https://github.com/gobatis/gobatis  
   亮点：  
   • 工程化 ORM、API 直观  
   • 支持 MyBatis 动态 SQL 语法  
   • 提供 Before/After 钩子、事务追踪、DryRun 等高级特性  
   • 日志、单测、Debug 模式等开发者友好能力

任务步骤（请严格按顺序输出）：

Step-1  差距对比  
- 用表格列出「Gobatis 已支持但 @framework/mybatis 缺失」的 8～10 个关键特性；  
- 标注每一项对「简单、方便、快速」目标的影响级别：High / Medium / Low。

Step-2  优化设计  
- 针对 High & Medium 级差距，给出在 Java 侧的落地方案（不改动 MyBatis 内核，只动封装层）；  
- 每个方案包含：设计思路、新增或调整的模块/类、对现有代码的最小侵入点。

Step-3  示例代码  
- 用 Java 17 + MyBatis 3.5.x 语法，为「新增钩子和分页查询」各给一段 20 行以内的精简示例；  
- 示例需可直接拷进 @framework/mybatis 目录运行，保留包名、注解风格与现有代码一致。

Step-4  迁移计划  
- 以 checklist 形式列出：  
  – 需新建的文件  
  – 需改动的旧文件（精确到类名）  
  – 单测用例增补建议  
  – 回归验证点（3 条以内）

Step-5  风险提示  
- 指出 2 个最可能的兼容性或性能陷阱，并给出规避手段。

输出格式：  
用二级标题对应 Step-1 ~ Step-5，标题下直接放内容，不要多余解释。
```

## 2508-0002 命名

1.分析 @framework/mybatis 这个文件夹命名是否合适
2.如不合适给出合理的命名，如果多个，给出列表供选择
3.根据功能生成 @framework/mybatis/README.md 文档

## 2508-0004 完善示例

1.@sample/mybatis 更改为 @sample/gobatis 
2.分析 @framework/mybatis 里面 md文档及功能，完善 @sample/gobatis 示例，给出必要的性能测试、压力测试
3.生成 @sample/gobatis/README.md 文档

```claude

  关于命名分析的总结：

  1. 当前命名评估：mybatis 文件夹名称虽然技术化，但考虑到：
    - 功能确实基于MyBatis设计理念
    - 用户已熟悉这个名称
    - 保持现有命名可避免大量重构工作
  2. 如需重命名的最佳选择：sqlmapper 是最佳候选，但你选择保持现状是明智的。

  生成的README.md特色：

  ✅ 完整的功能介绍 - 涵盖所有核心特性和高级功能
  ✅ 丰富的代码示例 - 从基础使用到高级特性的完整示例
  ✅ 详细的API文档 - SimpleSession和XMLSession接口说明
  ✅ 最佳实践指南 - 项目结构、错误处理、性能优化建议
  ✅ MyBatis迁移指南 - 帮助Java开发者快速上手
  ✅ 专业的文档格式 - 使用emoji图标和清晰的层级结构

```

## 2508-0003 帮助文档

依据 @framework/mybatis/ 下md文档 及 本功能特点 及 @sample/gobatis 下的示例与md文档，还有 @conf/database.yaml 配置文件，做如下优化：
1.修订 @README.md 文档
2.分析 @example/simple/views/home/docs.html 文档中【MyBatis集成】栏目是否拆分成多个
2.1 如拆分则给出合理的命名，并完成相关栏目与内容
2.2.修订 @example/simple/docs/data-access/mybatis.md 文档

```claude



```

