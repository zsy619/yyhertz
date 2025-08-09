# MyBatis高级特性

YYHertz框架的MyBatis集成提供了丰富的高级特性，包括XML映射器、动态SQL、钩子系统、调试模式等Go语言化的增强功能。

## 🎯 XMLSession - 动态SQL之王

### XML映射器配置

在 `conf/mybatis.yaml` 中启用XML映射器：

```yaml
# 基础配置
basic:
  enable: true                                    # 是否启用MyBatis
  config_file: "./config/mybatis-config.xml"     # MyBatis配置文件路径
  mapper_locations: "./mappers/*.xml"            # Mapper文件位置
  type_aliases_package: ""                       # 类型别名包

# 缓存配置
cache:
  enable: false                                 # 是否启用缓存
  type: "memory"                                # 缓存类型: memory, redis
  ttl: 3600                                     # 缓存生存时间(秒)
  max_size: 1000                                # 最大缓存条目数
  redis_addr: "localhost:6379"                  # Redis地址
  redis_db: 0                                   # Redis数据库
```

### XML映射文件示例

创建 `mappers/UserMapper.xml`：

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE mapper PUBLIC "-//mybatis.org//DTD Mapper 3.0//EN" 
    "http://mybatis.org/dtd/mybatis-3-mapper.dtd">
<mapper namespace="UserMapper">

    <!-- 动态条件查询 -->
    <select id="selectByCondition" parameterType="UserQuery" resultType="map">
        SELECT id, name, email, status, age, created_at 
        FROM users
        <where>
            <if test="name != null and name != ''">
                AND name LIKE CONCAT('%', #{name}, '%')
            </if>
            <if test="email != null and email != ''">
                AND email = #{email}
            </if>
            <if test="status != null and status != ''">
                AND status = #{status}
            </if>
            <if test="ageMin > 0">
                AND age >= #{ageMin}
            </if>
            <if test="ageMax > 0">
                AND age <= #{ageMax}
            </if>
            <if test="keyword != null and keyword != ''">
                AND (name LIKE CONCAT('%', #{keyword}, '%') OR email LIKE CONCAT('%', #{keyword}, '%'))
            </if>
        </where>
        <if test="orderBy != null and orderBy != ''">
            ORDER BY #{orderBy}
            <if test="orderDesc">DESC</if>
        </if>
    </select>

    <!-- 批量插入 -->
    <insert id="batchInsert" parameterType="list">
        INSERT INTO users (name, email, age, status, created_at, updated_at)
        VALUES 
        <foreach collection="list" item="user" separator=",">
            (#{user.name}, #{user.email}, #{user.age}, #{user.status}, NOW(), NOW())
        </foreach>
    </insert>

    <!-- 复杂关联查询 -->
    <select id="selectWithProfile" parameterType="int" resultMap="UserWithProfileMap">
        SELECT 
            u.id, u.name, u.email, u.status, u.created_at,
            p.bio, p.phone, p.avatar, p.location
        FROM users u
        LEFT JOIN user_profiles p ON u.id = p.user_id
        WHERE u.id = #{id} AND u.deleted_at IS NULL
    </select>

    <!-- 结果映射 -->
    <resultMap id="UserWithProfileMap" type="map">
        <id column="id" property="id"/>
        <result column="name" property="name"/>
        <result column="email" property="email"/>
        <result column="status" property="status"/>
        <result column="created_at" property="createdAt"/>
        <association property="profile" javaType="map">
            <result column="bio" property="bio"/>
            <result column="phone" property="phone"/>
            <result column="avatar" property="avatar"/>
            <result column="location" property="location"/>
        </association>
    </resultMap>

</mapper>
```

### 在控制器中使用XMLSession

```go
package controllers

import (
    "context"
    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/framework/mybatis"
)

type UserController struct {
    mvc.BaseController
    xmlSession mybatis.XMLSession  // 注入XMLSession
}

func NewUserController(xmlSession mybatis.XMLSession) *UserController {
    return &UserController{xmlSession: xmlSession}
}

// 动态条件查询
func (c *UserController) GetSearch() {
    ctx := context.Background()
    
    // 构建查询条件
    query := UserQuery{
        Name:      c.GetQuery("name", ""),
        Email:     c.GetQuery("email", ""),
        Status:    c.GetQuery("status", ""),
        Keyword:   c.GetQuery("keyword", ""),
        AgeMin:    c.GetQueryInt("age_min", 0),
        AgeMax:    c.GetQueryInt("age_max", 0),
        OrderBy:   c.GetQuery("order_by", "id"),
        OrderDesc: c.GetQueryBool("desc", false),
    }
    
    // 使用XML映射器查询
    users, err := c.xmlSession.SelectListByID(ctx, 
        "UserMapper.selectByCondition", query)
    if err != nil {
        c.Error(500, "查询失败: "+err.Error())
        return
    }
    
    c.JSON(mvc.Result{Success: true, Data: users})
}

// 分页动态查询
func (c *UserController) GetPageSearch() {
    ctx := context.Background()
    
    query := UserQuery{
        Name:    c.GetQuery("name", ""),
        Status:  c.GetQuery("status", ""),
        Keyword: c.GetQuery("keyword", ""),
    }
    
    pageReq := mybatis.PageRequest{
        Page: c.GetQueryInt("page", 1),
        Size: c.GetQueryInt("size", 10),
    }
    
    // XML分页查询
    pageResult, err := c.xmlSession.SelectPageByID(ctx,
        "UserMapper.selectByCondition", query, pageReq)
    if err != nil {
        c.Error(500, "查询失败")
        return
    }
    
    c.JSON(mvc.Result{Success: true, Data: pageResult})
}
```

### 查询参数结构体

```go
type UserQuery struct {
    Name      string `json:"name" form:"name"`
    Email     string `json:"email" form:"email"`
    Status    string `json:"status" form:"status"`
    Keyword   string `json:"keyword" form:"keyword"`
    AgeMin    int    `json:"age_min" form:"age_min"`
    AgeMax    int    `json:"age_max" form:"age_max"`
    OrderBy   string `json:"order_by" form:"order_by"`
    OrderDesc bool   `json:"order_desc" form:"order_desc"`
}
```

## 🎣 钩子系统 - 强大的扩展机制

YYHertz的MyBatis提供了Go函数式的钩子系统，支持链式配置：

### 钩子类型

```go
// 执行前钩子
type BeforeHook func(ctx context.Context, sql string, args []interface{}) error

// 执行后钩子  
type AfterHook func(ctx context.Context, result interface{}, duration time.Duration, err error)
```

### 内置钩子函数

#### 1. 性能监控钩子

```go
// 创建带性能监控的会话
session := mybatis.NewSimpleSession(db).
    AddAfterHook(mybatis.PerformanceHook(100 * time.Millisecond))

// 使用后自动监控慢查询
users, err := session.SelectList(ctx, "SELECT * FROM users")
// 如果查询超过100ms会自动记录警告日志
```

#### 2. 审计日志钩子

```go
// 审计钩子记录所有SQL操作
auditHook := func(ctx context.Context, sql string, args []interface{}) error {
    // 获取用户信息（从context或其他方式）
    userID := getUserIDFromContext(ctx)
    
    // 记录审计日志
    logrus.WithFields(logrus.Fields{
        "user_id": userID,
        "sql":     sql,
        "args":    args,
        "time":    time.Now(),
    }).Info("SQL执行审计")
    
    return nil
}

session := mybatis.NewSimpleSession(db).AddBeforeHook(auditHook)
```

#### 3. 安全检查钩子

```go
// SQL注入防护钩子
securityHook := func(ctx context.Context, sql string, args []interface{}) error {
    sql = strings.ToLower(sql)
    
    // 检查危险操作
    if strings.Contains(sql, "drop table") || 
       strings.Contains(sql, "delete from") && !strings.Contains(sql, "where") {
        return fmt.Errorf("危险SQL操作被阻止: %s", sql)
    }
    
    return nil
}

session := mybatis.NewSimpleSession(db).AddBeforeHook(securityHook)
```

### 复合钩子配置

```go
// 创建具有多个钩子的会话
session := mybatis.NewSimpleSession(db).
    AddBeforeHook(auditHook).                                    // 审计日志
    AddBeforeHook(securityHook).                                 // 安全检查
    AddAfterHook(mybatis.PerformanceHook(100*time.Millisecond)). // 性能监控
    AddAfterHook(metricsHook).                                   // 指标收集
    Debug(true)                                                  // 开启调试
```

## 🔍 DryRun调试模式

YYHertz独有的DryRun模式，让SQL调试变得简单：

### 基础DryRun

```go
// 创建DryRun会话
debugSession := mybatis.NewSimpleSession(db).DryRun(true).Debug(true)

// 执行查询（只打印SQL，不实际执行）
user, err := debugSession.SelectOne(ctx, 
    "SELECT * FROM users WHERE id = ? AND status = ?", 1, "active")

// 输出:
// [DryRun SELECT] SQL: SELECT * FROM users WHERE id = ? AND status = ?
// Args: [1 active]
```

### 复杂SQL调试

```go
// 调试分页查询
pageResult, err := debugSession.SelectPage(ctx,
    "SELECT * FROM users WHERE status = ? ORDER BY created_at DESC",
    mybatis.PageRequest{Page: 1, Size: 10},
    "active")

// 输出:
// [DryRun COUNT] SQL: SELECT COUNT(*) FROM (SELECT * FROM users WHERE status = ? ORDER BY created_at DESC) AS count_query
// Args: [active]
// [DryRun SELECT] SQL: SELECT * FROM users WHERE status = ? ORDER BY created_at DESC LIMIT ? OFFSET ?
// Args: [active 10 0]
```

### XML映射器DryRun

```go
xmlSession := mybatis.NewXMLMapper(db).DryRun(true).Debug(true)

query := UserQuery{
    Name:   "张",
    Status: "active",
    AgeMin: 25,
}

users, err := xmlSession.SelectListByID(ctx, 
    "UserMapper.selectByCondition", query)

// 输出动态生成的SQL:
// [DryRun XML] Namespace: UserMapper, Statement: selectByCondition
// [DryRun SELECT] SQL: SELECT id, name, email, status, age, created_at FROM users WHERE name LIKE CONCAT('%', ?, '%') AND status = ? AND age >= ?
// Args: [张 active 25]
```

## ⚡ 批量操作优化

### XML批量插入

```xml
<!-- 批量插入用户 -->
<insert id="batchInsertUsers" parameterType="list">
    INSERT INTO users (name, email, age, status, created_at, updated_at)
    VALUES 
    <foreach collection="list" item="user" separator=",">
        (#{user.name}, #{user.email}, #{user.age}, 'active', NOW(), NOW())
    </foreach>
</insert>
```

```go
func (c *UserController) PostBatchCreate() {
    ctx := context.Background()
    
    // 接收用户数组
    var users []User
    if err := c.ShouldBindJSON(&users); err != nil {
        c.Error(400, "参数错误: "+err.Error())
        return
    }
    
    // 批量插入
    affected, err := c.xmlSession.InsertByID(ctx, 
        "UserMapper.batchInsertUsers", users)
    if err != nil {
        c.Error(500, "批量创建失败")
        return
    }
    
    c.JSON(mvc.Result{
        Success: true,
        Data:    map[string]interface{}{"affected": affected},
        Message: fmt.Sprintf("成功创建%d个用户", affected),
    })
}
```

### 批量更新

```xml
<!-- 批量更新状态 -->
<update id="batchUpdateStatus" parameterType="map">
    UPDATE users SET 
        status = #{status},
        updated_at = NOW()
    WHERE id IN 
    <foreach collection="ids" item="id" open="(" separator="," close=")">
        #{id}
    </foreach>
</update>
```

```go
func (c *UserController) PutBatchUpdateStatus() {
    ctx := context.Background()
    
    var req struct {
        IDs    []int64 `json:"ids" binding:"required"`
        Status string  `json:"status" binding:"required"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.Error(400, "参数错误")
        return
    }
    
    params := map[string]interface{}{
        "ids":    req.IDs,
        "status": req.Status,
    }
    
    affected, err := c.xmlSession.UpdateByID(ctx,
        "UserMapper.batchUpdateStatus", params)
    if err != nil {
        c.Error(500, "批量更新失败")
        return
    }
    
    c.JSON(mvc.Result{Success: true, Data: map[string]interface{}{"affected": affected}})
}
```

## 🔄 动态SQL标签

### 条件判断 - if

```xml
<select id="selectUsers" parameterType="UserQuery" resultType="map">
    SELECT * FROM users
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
</select>
```

### 选择结构 - choose/when/otherwise

```xml
<select id="selectByPriority" parameterType="map" resultType="map">
    SELECT * FROM users
    WHERE 1=1
    <choose>
        <when test="priority == 'vip'">
            AND status = 'vip'
        </when>
        <when test="priority == 'premium'">
            AND status IN ('premium', 'vip')
        </when>
        <otherwise>
            AND status = 'active'
        </otherwise>
    </choose>
    ORDER BY created_at DESC
</select>
```

### 集合遍历 - foreach

```xml
<!-- IN查询 -->
<select id="selectByIds" parameterType="list" resultType="map">
    SELECT * FROM users 
    WHERE id IN
    <foreach collection="list" item="id" open="(" separator="," close=")">
        #{id}
    </foreach>
</select>

<!-- 批量条件 -->
<select id="selectMultipleConditions" parameterType="list" resultType="map">
    SELECT * FROM users WHERE
    <foreach collection="list" item="condition" separator=" OR ">
        (name = #{condition.name} AND status = #{condition.status})
    </foreach>
</select>
```

### 动态SET - set

```xml
<update id="updateSelective" parameterType="User">
    UPDATE users
    <set>
        <if test="name != null">name = #{name},</if>
        <if test="email != null">email = #{email},</if>
        <if test="status != null">status = #{status},</if>
        updated_at = NOW()
    </set>
    WHERE id = #{id}
</update>
```

## 🎯 结果映射和类型处理

### 复杂结果映射

```xml
<!-- 用户及其文章列表 -->
<select id="selectUserWithArticles" parameterType="int" resultMap="UserWithArticlesMap">
    SELECT 
        u.id as user_id, u.name as user_name, u.email,
        a.id as article_id, a.title, a.content, a.created_at as article_created_at
    FROM users u
    LEFT JOIN articles a ON u.id = a.author_id
    WHERE u.id = #{id}
</select>

<resultMap id="UserWithArticlesMap" type="map">
    <id column="user_id" property="id"/>
    <result column="user_name" property="name"/>
    <result column="email" property="email"/>
    <collection property="articles" ofType="map">
        <id column="article_id" property="id"/>
        <result column="title" property="title"/>
        <result column="content" property="content"/>
        <result column="article_created_at" property="createdAt"/>
    </collection>
</resultMap>
```

### Go结构体映射

```go
// 定义嵌套结构体
type UserWithArticles struct {
    ID       int64     `json:"id"`
    Name     string    `json:"name"`
    Email    string    `json:"email"`
    Articles []Article `json:"articles"`
}

type Article struct {
    ID        int64     `json:"id"`
    Title     string    `json:"title"`
    Content   string    `json:"content"`
    CreatedAt time.Time `json:"created_at"`
}

// 使用
func (c *UserController) GetUserWithArticles() {
    id := c.GetParamInt64("id")
    ctx := context.Background()
    
    result, err := c.xmlSession.SelectOneByID(ctx,
        "UserMapper.selectUserWithArticles", id)
    if err != nil {
        c.Error(500, "查询失败")
        return
    }
    
    // 可以直接使用map结果，或转换为结构体
    c.JSON(mvc.Result{Success: true, Data: result})
}
```

## 🔗 事务中的高级操作

### 事务内使用XML映射器

```go
func (c *UserController) PostTransferWithProfile() {
    ctx := context.Background()
    
    // 开启事务
    err := c.xmlSession.ExecuteInTransaction(ctx, func(txCtx context.Context, txSession mybatis.XMLSession) error {
        // 1. 更新用户余额
        _, err := txSession.UpdateByID(txCtx, "UserMapper.updateBalance", map[string]interface{}{
            "userId": fromUserID,
            "amount": -amount,
        })
        if err != nil {
            return err
        }
        
        // 2. 创建转账记录
        _, err = txSession.InsertByID(txCtx, "TransferMapper.insertRecord", transferRecord)
        if err != nil {
            return err
        }
        
        // 3. 批量更新相关数据
        _, err = txSession.UpdateByID(txCtx, "UserMapper.batchUpdateStatus", params)
        return err
    })
    
    if err != nil {
        c.Error(500, "转账失败: "+err.Error())
        return
    }
    
    c.JSON(mvc.Result{Success: true, Message: "转账成功"})
}
```

## 📚 下一步学习

掌握了MyBatis高级特性后，建议继续学习：

- **[MyBatis性能优化](./mybatis-performance)** - 性能测试、监控指标、生产环境调优
- **[数据库配置详解](./database-config)** - database.yaml的完整配置参数
- **[事务管理](./transaction)** - 复杂事务场景的处理方案

## 🔗 参考资源

- [GoBatis完整示例](../../gobatis/) - 包含所有高级特性的完整代码示例
- [MyBatis官方文档](https://mybatis.org/mybatis-3/) - XML映射器的标准参考
- [Go语言最佳实践](https://golang.org/doc/effective_go.html) - Go代码风格指南