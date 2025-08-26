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

## 2508-0005 持续集成

目录 @framework/mybatis 是基于java mybatis框架的二次封装，在现有的功能上，请继续优化如下功能：
### 一、核心功能（基本使用）
1.  **SQL 与代码分离**
    **核心思想**：将 SQL 语句从 Java 代码中剥离出来，集中存放在 XML 配置文件或注解中。
    **好处**：SQL 变更无需重新编译 Java 代码，易于维护和优化，DBA 也可以直接参与 SQL 审核。
2.  **强大的参数映射 (Parameter Mapping)**
    能够轻松地将 Java 对象（如 POJO, Map, 基本类型）作为参数传递给 SQL 语句。
    支持在 SQL 中使用 `#{}`（预编译，防 SQL 注入）和 `${}`（字符串替换，慎用）来引用这些参数。
3.  **强大的结果集映射 (Result Mapping)**
    **自动映射**：如果数据库字段名（如 `user_name`）和 Java 对象属性名（如 `userName`）遵循一定的命名规则（如驼峰转换），MyBatis 可以自动完成映射，无需额外配置。
    **显式映射**：通过 `<resultMap>` 标签，可以自定义非常复杂的映射关系，解决以下问题：
       数据库字段名和 Java 属性名不一致。
       处理一对一、一对多等复杂的关联查询。
       将查询结果映射到复杂的对象树或集合中。
4.  **动态 SQL (Dynamic SQL)**
    **这是 MyBatis 的一个**标志性强大功能**。它允许在 XML 中编写条件判断、循环等动态生成的 SQL 语句。
    **常用标签**：
        `<if>`：条件判断。
        `<choose>`, `<when>`, `<otherwise>`：类似 Java 的 switch-case。
        `<trim>`, `<where>`, `<set>`：智能地处理 SQL 语句的前缀、后缀，避免语法错误（例如解决 `WHERE` 或 `SET` 后的冗余 `AND` 或逗号）。
        `<foreach>`：遍历集合，常用于 `IN` 条件或批量操作。
### 二、高级与扩展功能
1.  **缓存机制 (Caching)**
    **一级缓存**：默认开启，它是 **SqlSession 级别**的缓存。在同一个 SqlSession 中执行相同的查询，第二次会直接从缓存中取数据，而不再次访问数据库。
    **二级缓存**：需要手动配置开启，它是 **Mapper 级别**的缓存（跨 SqlSession）。多个 SqlSession 操作同一个 Mapper 的 SQL，数据会共享到二级缓存中。非常适合用于只读或读多写少的场景。
2.  **插件机制 (Plugin / Interceptor)**
    允许用户编写插件来拦截 MyBatis 核心组件的执行过程，例如拦截 Executor、ParameterHandler、ResultSetHandler、StatementHandler 的方法。
    **典型应用**：
      **分页**：编写分页插件（如 PageHelper）。
      **性能监控**：记录 SQL 执行时间。
      **自定义权限控制**：在 SQL 执行前自动加上某些条件。
      **通用审计字段（如 create_time, update_time）的自动填充**。
3.  **与 Spring 框架无缝集成**
    通过 `mybatis-spring` 集成包，可以非常方便地将 MyBatis 接入 Spring 和 Spring Boot 项目。
    主要好处：
      SqlSession 的生命周期交由 Spring 管理（通常是注入到 Mapper 接口中）。
      可以使用 `@Autowired` 直接注入 Mapper 接口，无需手动创建。
      支持与 Spring 的事务管理一起使用（`@Transactional`）。
4.  **注解支持**
    除了主流的 XML 配置方式，MyBatis 也提供了注解方式（如 `@Select`, `@Insert`, `@Update`, `@Delete`, `@Results` 等）来编写 SQL 和映射关系。
    **适用场景**：简单的、SQL 不复杂的场景。对于复杂的动态 SQL，使用 XML 方式可读性和维护性更高。
5.  **存储过程支持**
    支持调用数据库中的存储过程，并处理输入/输出参数。
提示：
- sqlx 库完美解决了结果集自动映射到结构体的问题
- sqlc: 它通过解析你编写的 SQL 文件，直接生成类型安全的 Go 代码和结构体（函数+结构体）。这相当于一个超级强大的代码生成器，编译时就能发现 SQL 错误，非常可靠。
- jet: 它是一个类型安全的 SQL 构建器，允许你用 Go 代码的方式编写 SQL，并在编译时检查类型安全。适合不喜欢拼接 SQL 字符串但又需要灵活性的开发者。


