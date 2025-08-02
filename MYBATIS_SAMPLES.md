# MyBatis-Go 示例和测试用例

MyBatis-Go框架的完整示例和测试用例已迁移到 `sample/mybat` 目录。

## 📁 目录说明

```
sample/mybat/                          # MyBatis-Go完整示例
├── mappers/                           # XML映射文件
│   └── UserMapper.xml                # 用户映射器XML配置
├── models.go                         # 数据模型定义
├── user_mapper.go                    # 用户映射器接口和实现
├── sql_mappings.go                   # SQL映射常量定义
├── main_test.go                      # 主要测试运行器
├── integration_test.go               # 集成测试
├── database_setup.go                 # 数据库设置工具
├── xml_mapper_loader.go              # XML映射器加载器
├── xml_based_test.go                 # 基于XML的测试用例
├── mybatis-config.xml                # MyBatis主配置文件
├── database.properties               # 数据库属性配置
└── README.md                         # 详细文档
```

## 🚀 快速开始

### 1. 进入示例目录

```bash
cd sample/mybat
```

### 2. 配置数据库

修改 `database.properties` 或 `database_setup.go` 中的数据库连接信息：

```properties
database.url=jdbc:mysql://localhost:3306/mybatis_test
database.username=root
database.password=your_password
```

### 3. 运行测试

```bash
# 运行所有测试
go test -v ./

# 运行传统Go代码测试
go test -v ./ -run "^((?!XML).)*$"

# 运行基于XML的测试
go test -v ./ -run ".*XML.*"

# 运行性能基准测试
go test -v ./ -bench=.
```

## 🎯 主要特性

### 双重配置方式
- **Go代码配置**: 传统的Go代码配置方式
- **XML文件配置**: 完整的MyBatis XML配置支持

### 核心功能
- ✅ SQL映射和动态SQL
- ✅ 结果集映射
- ✅ 多级缓存机制
- ✅ 事务管理
- ✅ 批量操作
- ✅ 复杂查询和关联查询
- ✅ 存储过程支持

### 测试覆盖
- **80+ 个测试用例**
- **完整的功能测试覆盖**
- **性能基准测试**
- **集成测试**
- **错误处理测试**

## 📖 文档

详细的使用文档请参考：`sample/mybat/README.md`

## 🔄 版本历史

### v1.1.0 (2024-08-02)
- ✅ 修订所有编译错误
- ✅ 将测试用例迁移到sample/mybat目录
- ✅ 新增基于XML文件的完整测试用例
- ✅ 重新组织目录结构
- ✅ 完善文档和使用说明

### v1.0.0 (之前版本)
- ✅ 完成基础MyBatis-Go框架
- ✅ 实现基础CRUD操作
- ✅ 支持动态SQL
- ✅ 缓存和事务功能

## 🤝 贡献

欢迎提交问题和改进建议：

1. 克隆项目
2. 进入 `sample/mybat` 目录
3. 运行测试确保功能正常
4. 提交改进代码

---

**MyBatis-Go Framework** - 企业级Golang ORM解决方案！