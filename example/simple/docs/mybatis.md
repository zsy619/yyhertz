# MyBatis集成

YYHertz框架内置了MyBatis集成支持，提供简单易用的SQL映射和数据库操作功能。

## 快速开始

### 1. 配置数据库连接

在 `conf/database.yaml` 中配置数据库连接：

```yaml
database:
  driver: "mysql"
  host: "localhost"
  port: 3306
  username: "root"
  password: "password"
  dbname: "myapp"
  charset: "utf8mb4"
  parseTime: true
  loc: "Local"
```

### 2. 初始化MyBatis

```go
import (
    "github.com/zsy619/yyhertz/framework/mybatis"
    "github.com/zsy619/yyhertz/framework/config"
)

func main() {
    // 加载数据库配置
    dbConfig := config.DatabaseConfig{}
    config.LoadConfig("database", &dbConfig)
    
    // 初始化MyBatis
    mybatis.Initialize(&dbConfig.Database)
    
    // 启动应用
    app := mvc.HertzApp
    app.Run()
}
```

## 实体映射

### 定义实体

```go
type User struct {
    ID       int64     `db:"id" json:"id"`
    Username string    `db:"username" json:"username"`
    Email    string    `db:"email" json:"email"`
    Status   int       `db:"status" json:"status"`
    CreateAt time.Time `db:"create_at" json:"create_at"`
    UpdateAt time.Time `db:"update_at" json:"update_at"`
}
```

### 创建Mapper接口

```go
type UserMapper interface {
    // 根据ID查询用户
    GetUserById(id int64) (*User, error)
    
    // 根据用户名查询用户
    GetUserByUsername(username string) (*User, error)
    
    // 获取用户列表
    GetUserList(offset, limit int) ([]*User, error)
    
    // 创建用户
    CreateUser(user *User) error
    
    // 更新用户
    UpdateUser(user *User) error
    
    // 删除用户
    DeleteUser(id int64) error
}
```

## SQL映射

### XML映射文件

创建 `mappers/user_mapper.xml`：

```xml
<?xml version="1.0" encoding="UTF-8"?>
<mapper namespace="UserMapper">
    
    <!-- 根据ID查询用户 -->
    <select id="GetUserById" parameterType="int64" resultType="User">
        SELECT id, username, email, status, create_at, update_at
        FROM users 
        WHERE id = #{id} AND status = 1
    </select>
    
    <!-- 根据用户名查询用户 -->
    <select id="GetUserByUsername" parameterType="string" resultType="User">
        SELECT id, username, email, status, create_at, update_at
        FROM users 
        WHERE username = #{username} AND status = 1
    </select>
    
    <!-- 获取用户列表 -->
    <select id="GetUserList" resultType="User">
        SELECT id, username, email, status, create_at, update_at
        FROM users 
        WHERE status = 1
        ORDER BY id DESC
        LIMIT #{offset}, #{limit}
    </select>
    
    <!-- 创建用户 -->
    <insert id="CreateUser" parameterType="User">
        INSERT INTO users (username, email, status, create_at, update_at)
        VALUES (#{username}, #{email}, #{status}, NOW(), NOW())
    </insert>
    
    <!-- 更新用户 -->
    <update id="UpdateUser" parameterType="User">
        UPDATE users SET 
            username = #{username},
            email = #{email},
            status = #{status},
            update_at = NOW()
        WHERE id = #{id}
    </update>
    
    <!-- 删除用户 -->
    <delete id="DeleteUser" parameterType="int64">
        DELETE FROM users WHERE id = #{id}
    </delete>
    
</mapper>
```

## 在控制器中使用

```go
type UserController struct {
    mvc.BaseController
    userMapper UserMapper
}

func (c *UserController) GetIndex() {
    // 获取用户列表
    users, err := c.userMapper.GetUserList(0, 10)
    if err != nil {
        c.Error(500, "获取用户列表失败")
        return
    }
    
    c.JSON(map[string]interface{}{
        "success": true,
        "data":    users,
    })
}

func (c *UserController) GetShow() {
    id := c.GetParamInt64("id")
    
    user, err := c.userMapper.GetUserById(id)
    if err != nil {
        c.Error(500, "获取用户信息失败")
        return
    }
    
    if user == nil {
        c.Error(404, "用户不存在")
        return
    }
    
    c.JSON(map[string]interface{}{
        "success": true,
        "data":    user,
    })
}

func (c *UserController) PostCreate() {
    user := &User{
        Username: c.GetForm("username"),
        Email:    c.GetForm("email"),
        Status:   1,
    }
    
    err := c.userMapper.CreateUser(user)
    if err != nil {
        c.Error(500, "创建用户失败")
        return
    }
    
    c.JSON(map[string]interface{}{
        "success": true,
        "message": "用户创建成功",
    })
}
```

## 高级功能

### 动态SQL

```xml
<select id="SearchUsers" resultType="User">
    SELECT id, username, email, status, create_at, update_at
    FROM users 
    <where>
        <if test="username != null and username != ''">
            AND username LIKE CONCAT('%', #{username}, '%')
        </if>
        <if test="email != null and email != ''">
            AND email LIKE CONCAT('%', #{email}, '%')
        </if>
        <if test="status != null">
            AND status = #{status}
        </if>
    </where>
    ORDER BY id DESC
</select>
```

### 批量操作

```xml
<insert id="BatchCreateUsers" parameterType="list">
    INSERT INTO users (username, email, status, create_at, update_at)
    VALUES 
    <foreach collection="users" item="user" separator=",">
        (#{user.username}, #{user.email}, #{user.status}, NOW(), NOW())
    </foreach>
</insert>
```

### 关联查询

```xml
<select id="GetUserWithProfile" resultMap="UserWithProfileMap">
    SELECT 
        u.id, u.username, u.email, u.status,
        p.real_name, p.phone, p.avatar
    FROM users u
    LEFT JOIN user_profiles p ON u.id = p.user_id
    WHERE u.id = #{id}
</select>

<resultMap id="UserWithProfileMap" type="UserWithProfile">
    <id column="id" property="id"/>
    <result column="username" property="username"/>
    <result column="email" property="email"/>
    <result column="status" property="status"/>
    <association property="profile" javaType="UserProfile">
        <result column="real_name" property="realName"/>
        <result column="phone" property="phone"/>
        <result column="avatar" property="avatar"/>
    </association>
</resultMap>
```

## 事务管理

```go
import "github.com/zsy619/yyhertz/framework/mybatis/transaction"

func (c *UserController) PostTransferMoney() {
    fromUserId := c.GetFormInt64("from_user_id")
    toUserId := c.GetFormInt64("to_user_id")
    amount := c.GetFormFloat64("amount")
    
    // 开启事务
    tx, err := transaction.Begin()
    if err != nil {
        c.Error(500, "开启事务失败")
        return
    }
    
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
            c.Error(500, "操作失败，已回滚")
        }
    }()
    
    // 扣除转出用户余额
    err = c.userMapper.WithTx(tx).DeductBalance(fromUserId, amount)
    if err != nil {
        tx.Rollback()
        c.Error(500, "扣除余额失败")
        return
    }
    
    // 增加转入用户余额
    err = c.userMapper.WithTx(tx).AddBalance(toUserId, amount)
    if err != nil {
        tx.Rollback()
        c.Error(500, "增加余额失败")
        return
    }
    
    // 提交事务
    err = tx.Commit()
    if err != nil {
        c.Error(500, "提交事务失败")
        return
    }
    
    c.JSON(map[string]interface{}{
        "success": true,
        "message": "转账成功",
    })
}
```

## 配置选项

### MyBatis配置

```yaml
mybatis:
  # XML映射文件路径
  mapper_locations: "mappers/*.xml"
  
  # 类型别名包路径
  type_aliases_package: "models"
  
  # 配置选项
  configuration:
    # 是否开启驼峰命名转换
    map_underscore_to_camel_case: true
    
    # 缓存配置
    cache_enabled: true
    
    # 延迟加载
    lazy_loading_enabled: true
    
    # SQL执行超时时间（秒）
    default_statement_timeout: 30
```

## 性能优化

### 1. 连接池配置

```yaml
database:
  # 最大连接数
  max_open_conns: 100
  
  # 最大空闲连接数
  max_idle_conns: 10
  
  # 连接最大生存时间
  conn_max_lifetime: 3600
```

### 2. 一级缓存

MyBatis默认开启一级缓存（Session级别），同一个Session内相同查询会被缓存。

### 3. 二级缓存

```xml
<mapper namespace="UserMapper">
    <!-- 开启二级缓存 -->
    <cache eviction="LRU" flushInterval="60000" size="512" readOnly="true"/>
    
    <select id="GetUserById" useCache="true">
        SELECT * FROM users WHERE id = #{id}
    </select>
</mapper>
```

## 最佳实践

1. **合理使用索引**：确保查询条件字段有适当的数据库索引
2. **避免N+1问题**：使用关联查询代替循环查询
3. **分页查询**：大数据量查询使用LIMIT分页
4. **参数验证**：在Mapper方法中验证输入参数
5. **异常处理**：妥善处理数据库异常和业务异常
6. **连接管理**：及时关闭数据库连接，避免连接泄露

## 示例项目

查看完整的MyBatis集成示例：[example/mybatis](../mybat/)