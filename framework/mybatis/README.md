# MyBatis-Go 框架

一个受Java MyBatis启发的Golang ORM框架，完全兼容MyBatis XML映射文件，提供简洁高效的数据访问层解决方案。

## ✨ 核心特性

- 🎯 **XML映射器支持** - 完全兼容Java MyBatis mapper.xml文件格式
- 🔧 **动态SQL构建** - 支持`<if>`、`<where>`、`<foreach>`等标签
- 🚀 **简化版Session接口** - 基于Go语言惯用法设计的API
- 🔍 **DryRun调试模式** - 安全的SQL预览和调试功能  
- 📊 **结果映射** - 支持ResultMap复杂结果映射
- 🎣 **钩子系统** - 灵活的Before/After钩子机制
- 🔌 **插件扩展** - 可扩展的插件架构
- 📈 **性能监控** - 内置性能指标和监控

## 🚀 快速开始

### 安装

```go
import "github.com/zsy619/yyhertz/framework/mybatis"
```

### 基本使用

```go
package main

import (
    "context"
    "log"
    
    "gorm.io/gorm"
    "github.com/zsy619/yyhertz/framework/mybatis"
)

func main() {
    // 创建数据库连接（使用GORM）
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal(err)
    }

    // 创建简化版会话
    session := mybatis.NewSimpleSession(db)
    defer session.Close()

    // 基础SQL操作
    ctx := context.Background()
    
    // 查询单条记录
    user, err := session.SelectOne(ctx, "SELECT * FROM users WHERE id = ?", 1)
    if err != nil {
        log.Fatal(err)
    }
    
    // 查询列表
    users, err := session.SelectList(ctx, "SELECT * FROM users WHERE status = ?", "active")
    if err != nil {
        log.Fatal(err)
    }
    
    // 分页查询
    pageResult, err := session.SelectPage(ctx, "SELECT * FROM users", mybatis.PageRequest{
        Page: 1,
        Size: 10,
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // 插入记录
    affected, err := session.Insert(ctx, "INSERT INTO users (name, email) VALUES (?, ?)", "张三", "zhangsan@example.com")
}
```

### XML映射器使用

#### 1. 创建XML映射文件

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE mapper PUBLIC "-//mybatis.org//DTD Mapper 3.0//EN" "http://mybatis.org/dtd/mybatis-3-mapper.dtd">
<mapper namespace="UserMapper">
    
    <!-- 结果映射 -->
    <resultMap id="userResultMap" type="User">
        <id property="id" column="id" />
        <result property="name" column="name" />
        <result property="email" column="email" />
        <result property="createdAt" column="created_at" />
    </resultMap>

    <!-- 查询用户 -->
    <select id="selectById" parameterType="int" resultMap="userResultMap">
        SELECT id, name, email, created_at 
        FROM users 
        WHERE id = #{id}
    </select>

    <!-- 动态SQL查询 -->
    <select id="selectByCondition" parameterType="UserQuery" resultMap="userResultMap">
        SELECT id, name, email, created_at FROM users
        <where>
            <if test="name != null and name != ''">
                AND name LIKE CONCAT('%', #{name}, '%')
            </if>
            <if test="email != null and email != ''">
                AND email = #{email}
            </if>
            <if test="status != null">
                AND status = #{status}
            </if>
        </where>
        ORDER BY created_at DESC
    </select>

    <!-- 批量插入 -->
    <insert id="batchInsert" parameterType="list">
        INSERT INTO users (name, email) VALUES
        <foreach collection="list" item="user" separator=",">
            (#{user.name}, #{user.email})
        </foreach>
    </insert>

</mapper>
```

#### 2. 使用XML映射器

```go
// 创建XML映射器会话
session := mybatis.NewXMLMapper(db)

// 加载XML映射文件
err := session.LoadMapperXML("path/to/user_mapper.xml")
if err != nil {
    log.Fatal(err)
}

// 或者从字符串加载
xmlContent := `<?xml version="1.0" encoding="UTF-8"?>...`
err = session.LoadMapperFromString("UserMapper", xmlContent)

// 使用映射器方法
ctx := context.Background()

// 按ID查询
user, err := session.SelectOneByID(ctx, "UserMapper.selectById", 1)

// 动态SQL查询
query := UserQuery{
    Name:   "张",
    Status: "active",
}
users, err := session.SelectListByID(ctx, "UserMapper.selectByCondition", query)

// 分页查询
pageResult, err := session.SelectPageByID(ctx, "UserMapper.selectByCondition", query, mybatis.PageRequest{
    Page: 1,
    Size: 20,
})
```

## 📖 API参考

### SimpleSession接口

```go
type SimpleSession interface {
    // 基础查询方法
    SelectOne(ctx context.Context, sql string, args ...interface{}) (interface{}, error)
    SelectList(ctx context.Context, sql string, args ...interface{}) ([]interface{}, error)  
    SelectPage(ctx context.Context, sql string, page PageRequest, args ...interface{}) (*PageResult, error)
    
    // 数据操作方法
    Insert(ctx context.Context, sql string, args ...interface{}) (int64, error)
    Update(ctx context.Context, sql string, args ...interface{}) (int64, error)
    Delete(ctx context.Context, sql string, args ...interface{}) (int64, error)
    
    // 配置方法
    DryRun(enabled bool) SimpleSession  // 启用DryRun模式
    Debug(enabled bool) SimpleSession   // 启用Debug模式
    
    // 钩子方法
    AddBeforeHook(hook BeforeHook) SimpleSession
    AddAfterHook(hook AfterHook) SimpleSession
}
```

### XMLSession接口

```go
type XMLSession interface {
    SimpleSession
    
    // XML映射器加载
    LoadMapperXML(xmlPath string) error
    LoadMapperFromString(namespace string, xmlContent string) error
    LoadMappersFromDir(dir string) error
    
    // 映射器方法调用
    SelectOneByID(ctx context.Context, statementId string, parameter interface{}) (interface{}, error)
    SelectListByID(ctx context.Context, statementId string, parameter interface{}) ([]interface{}, error)
    SelectPageByID(ctx context.Context, statementId string, parameter interface{}, page PageRequest) (*PageResult, error)
    InsertByID(ctx context.Context, statementId string, parameter interface{}) (int64, error)
    UpdateByID(ctx context.Context, statementId string, parameter interface{}) (int64, error)
    DeleteByID(ctx context.Context, statementId string, parameter interface{}) (int64, error)
}
```

### 钩子系统

```go
// Before钩子：在SQL执行前调用
type BeforeHook func(ctx context.Context, sql string, args []interface{}) error

// After钩子：在SQL执行后调用  
type AfterHook func(ctx context.Context, result interface{}, duration time.Duration, err error)

// 使用示例
session := mybatis.NewSimpleSession(db)

// 添加SQL日志钩子
session.AddBeforeHook(func(ctx context.Context, sql string, args []interface{}) error {
    log.Printf("Executing SQL: %s with args: %v", sql, args)
    return nil
})

// 添加性能监控钩子
session.AddAfterHook(func(ctx context.Context, result interface{}, duration time.Duration, err error) {
    if duration > time.Second {
        log.Printf("Slow query detected: %v", duration)
    }
})
```

## 🔥 高级功能

### DryRun调试模式

```go
// 启用DryRun模式 - 只打印SQL，不实际执行
session := mybatis.NewSimpleSession(db).DryRun(true).Debug(true)

// 这将只打印SQL，不会实际插入数据
affected, err := session.Insert(ctx, "INSERT INTO users (name) VALUES (?)", "测试用户")
// 输出: [DryRun INSERT] SQL: INSERT INTO users (name) VALUES (?) Args: [测试用户]
```

### 动态SQL标签支持

框架支持以下MyBatis动态SQL标签：

- `<if test="condition">` - 条件判断
- `<where>` - WHERE子句，自动处理AND/OR
- `<set>` - SET子句，自动处理逗号  
- `<foreach>` - 循环遍历
- `<choose><when><otherwise>` - 选择结构
- `<trim>` - 字符串修剪
- `<bind>` - 变量绑定

### 结果映射

```xml
<resultMap id="userWithRoles" type="UserWithRoles">
    <id property="id" column="user_id" />
    <result property="name" column="user_name" />
    <collection property="roles" ofType="Role">
        <id property="id" column="role_id" />
        <result property="name" column="role_name" />
    </collection>
</resultMap>
```

## 🎯 最佳实践

### 1. 项目结构建议

```
project/
├── mapper/
│   ├── user_mapper.xml
│   ├── order_mapper.xml  
│   └── ...
├── model/
│   ├── user.go
│   ├── order.go
│   └── ...
├── service/
│   ├── user_service.go
│   └── ...
└── main.go
```

### 2. 错误处理

```go
user, err := session.SelectOne(ctx, sql, args...)
if err != nil {
    // 检查是否是"记录未找到"错误
    if errors.Is(err, gorm.ErrRecordNotFound) {
        // 处理记录不存在的情况
        return nil, fmt.Errorf("用户不存在")
    }
    // 其他数据库错误
    return nil, fmt.Errorf("查询用户失败: %w", err)
}

if user == nil {
    // SelectOne返回nil表示没有找到记录
    return nil, fmt.Errorf("用户不存在")
}
```

### 3. 性能优化

```go
// 使用分页查询避免大量数据
pageResult, err := session.SelectPage(ctx, sql, mybatis.PageRequest{
    Page: 1,
    Size: 100, // 建议单页不超过1000条
}, args...)

// 使用DryRun模式调试复杂SQL
session.DryRun(true).Debug(true).SelectList(ctx, complexSQL, args...)

// 添加性能监控钩子
session.AddAfterHook(func(ctx context.Context, result interface{}, duration time.Duration, err error) {
    if duration > 500*time.Millisecond {
        log.Printf("慢查询警告: 耗时 %v", duration)
    }
})
```

## 🔄 从Java MyBatis迁移

### 兼容性说明

✅ **完全兼容的特性：**
- mapper.xml文件格式
- 动态SQL标签（if、where、foreach等）
- ResultMap结果映射
- 参数占位符语法#{param}

⚠️ **需要适配的特性：**
- 接口映射器需要手动实现Go版本
- Java类型需要映射到Go类型
- 注解方式需要改为XML或代码方式

### 迁移步骤

1. **复制XML映射文件** - 可直接使用现有的mapper.xml文件
2. **定义Go结构体** - 对应Java的POJO类
3. **创建服务层** - 替代Java的Mapper接口
4. **调整类型映射** - Java类型对应到Go类型

```go
// Java: public interface UserMapper
// Go: 使用服务结构体
type UserService struct {
    session mybatis.XMLSession
}

func (s *UserService) SelectById(ctx context.Context, id int64) (*User, error) {
    result, err := s.session.SelectOneByID(ctx, "UserMapper.selectById", id)
    if err != nil {
        return nil, err
    }
    
    if result == nil {
        return nil, nil
    }
    
    // 类型转换
    user := mapToUser(result.(map[string]interface{}))
    return user, nil
}
```

## 📊 性能指标

- **查询性能**: 基于GORM，继承其优化特性
- **内存使用**: 轻量级设计，最小内存占用  
- **并发安全**: 全面的并发安全保护
- **连接池**: 支持数据库连接池管理

## 🤝 贡献指南

欢迎提交Issue和Pull Request！

## 📄 许可证

MIT License

---

> 💡 **提示**: 这是YYHertz框架的一部分，与其他框架模块（MVC、ORM等）无缝集成使用。