# YYHertz MyBatis 使用示例

<div align="center">

🔧 **MyBatis引擎使用指南** | 复杂查询的最佳选择

</div>

---

## 📋 目录

- [MyBatis基础概念](#mybatis基础概念)
- [XML配置和映射](#xml配置和映射)
- [复杂查询示例](#复杂查询示例)
- [动态SQL使用](#动态sql使用)
- [存储过程调用](#存储过程调用)
- [批量操作优化](#批量操作优化)
- [智能引擎选择](#智能引擎选择)

---

## 🎯 MyBatis基础概念

### 什么时候使用MyBatis引擎？
- 复杂的多表关联查询
- 需要优化的SQL语句
- 存储过程调用
- 复杂的统计分析查询
- 需要精确控制SQL的场景

### YYHertz中的MyBatis集成

```go
package main

import (
    "github.com/zsy619/yyhertz/framework/orm"
    "github.com/zsy619/yyhertz/framework/mvc"
)

func main() {
    app := mvc.HertzApp
    
    // 初始化MyBatis引擎
    mybatisEngine, err := orm.NewMyBatisEngine(&orm.MyBatisConfig{
        Driver:   "mysql",
        DataSource: "user:password@tcp(localhost:3306)/dbname?charset=utf8mb4",
        MapperLocations: []string{
            "mappers/*.xml",
            "mappers/user/*.xml",
        },
        ConfigPath: "mybatis-config.xml",
    })
    if err != nil {
        panic(err)
    }
    
    // 注册到智能选择器
    orm.RegisterEngine("mybatis", mybatisEngine)
    
    app.Run(":8888")
}
```

---

## 📁 XML配置和映射

### 1. MyBatis配置文件

```xml
<!-- mybatis-config.xml -->
<?xml version="1.0" encoding="UTF-8" ?>
<!DOCTYPE configuration
  PUBLIC "-//mybatis.org//DTD Config 3.0//EN"
  "http://mybatis.org/dtd/mybatis-3-config.dtd">
<configuration>
  <!-- 设置 -->
  <settings>
    <!-- 开启驼峰命名转换 -->
    <setting name="mapUnderscoreToCamelCase" value="true"/>
    <!-- 开启延迟加载 -->
    <setting name="lazyLoadingEnabled" value="true"/>
    <!-- 设置超时时间 -->
    <setting name="defaultStatementTimeout" value="30"/>
    <!-- 开启二级缓存 -->
    <setting name="cacheEnabled" value="true"/>
  </settings>
  
  <!-- 类型别名 -->
  <typeAliases>
    <typeAlias type="models.User" alias="User"/>
    <typeAlias type="models.Order" alias="Order"/>
    <typeAlias type="models.Product" alias="Product"/>
  </typeAliases>
  
  <!-- 插件 -->
  <plugins>
    <!-- 分页插件 -->
    <plugin interceptor="com.github.pagehelper.PageInterceptor">
      <property name="helperDialect" value="mysql"/>
      <property name="reasonable" value="true"/>
      <property name="supportMethodsArguments" value="true"/>
      <property name="params" value="count=countSql"/>
    </plugin>
  </plugins>
</configuration>
```

### 2. 基础Mapper映射文件

```xml
<!-- mappers/UserMapper.xml -->
<?xml version="1.0" encoding="UTF-8" ?>
<!DOCTYPE mapper
  PUBLIC "-//mybatis.org//DTD Mapper 3.0//EN"
  "http://mybatis.org/dtd/mybatis-3-mapper.dtd">
<mapper namespace="user">

  <!-- 结果映射 -->
  <resultMap id="UserResultMap" type="User">
    <id property="id" column="id"/>
    <result property="username" column="username"/>
    <result property="email" column="email"/>
    <result property="firstName" column="first_name"/>
    <result property="lastName" column="last_name"/>
    <result property="status" column="status"/>
    <result property="createdAt" column="created_at"/>
    <result property="updatedAt" column="updated_at"/>
  </resultMap>

  <!-- 用户详情结果映射(包含关联数据) -->
  <resultMap id="UserDetailResultMap" type="User">
    <id property="id" column="user_id"/>
    <result property="username" column="username"/>
    <result property="email" column="email"/>
    <result property="firstName" column="first_name"/>
    <result property="lastName" column="last_name"/>
    <result property="status" column="status"/>
    <result property="createdAt" column="created_at"/>
    
    <!-- 一对一关联 - 用户资料 -->
    <association property="profile" javaType="UserProfile">
      <id property="id" column="profile_id"/>
      <result property="avatar" column="avatar"/>
      <result property="bio" column="bio"/>
      <result property="website" column="website"/>
    </association>
    
    <!-- 一对多关联 - 用户角色 -->
    <collection property="roles" ofType="Role">
      <id property="id" column="role_id"/>
      <result property="name" column="role_name"/>
      <result property="displayName" column="role_display_name"/>
    </collection>
  </resultMap>

  <!-- 基本查询 -->
  <select id="selectById" parameterType="int" resultMap="UserResultMap">
    SELECT * FROM users WHERE id = #{id} AND deleted_at IS NULL
  </select>

  <select id="selectByUsername" parameterType="string" resultMap="UserResultMap">
    SELECT * FROM users 
    WHERE username = #{username} 
    AND deleted_at IS NULL
  </select>

  <!-- 分页查询 -->
  <select id="selectPageList" parameterType="map" resultMap="UserResultMap">
    SELECT * FROM users 
    WHERE deleted_at IS NULL
    <if test="status != null">
      AND status = #{status}
    </if>
    <if test="keyword != null and keyword != ''">
      AND (username LIKE CONCAT('%', #{keyword}, '%') 
           OR email LIKE CONCAT('%', #{keyword}, '%'))
    </if>
    ORDER BY created_at DESC
  </select>

  <!-- 复杂关联查询 -->
  <select id="selectUserDetail" parameterType="int" resultMap="UserDetailResultMap">
    SELECT 
      u.id as user_id,
      u.username,
      u.email,
      u.first_name,
      u.last_name,
      u.status,
      u.created_at,
      p.id as profile_id,
      p.avatar,
      p.bio,
      p.website,
      r.id as role_id,
      r.name as role_name,
      r.display_name as role_display_name
    FROM users u
    LEFT JOIN user_profiles p ON u.id = p.user_id
    LEFT JOIN user_roles ur ON u.id = ur.user_id
    LEFT JOIN roles r ON ur.role_id = r.id
    WHERE u.id = #{userId} AND u.deleted_at IS NULL
  </select>

  <!-- 插入用户 -->
  <insert id="insert" parameterType="User" useGeneratedKeys="true" keyProperty="id">
    INSERT INTO users (username, email, password, first_name, last_name, status, created_at, updated_at)
    VALUES (#{username}, #{email}, #{password}, #{firstName}, #{lastName}, #{status}, NOW(), NOW())
  </insert>

  <!-- 批量插入 -->
  <insert id="batchInsert" parameterType="list">
    INSERT INTO users (username, email, password, first_name, last_name, status, created_at, updated_at)
    VALUES
    <foreach collection="list" item="user" separator=",">
      (#{user.username}, #{user.email}, #{user.password}, 
       #{user.firstName}, #{user.lastName}, #{user.status}, NOW(), NOW())
    </foreach>
  </insert>

  <!-- 更新用户 -->
  <update id="update" parameterType="User">
    UPDATE users SET
      <if test="firstName != null">first_name = #{firstName},</if>
      <if test="lastName != null">last_name = #{lastName},</if>
      <if test="status != null">status = #{status},</if>
      updated_at = NOW()
    WHERE id = #{id} AND deleted_at IS NULL
  </update>

  <!-- 软删除 -->
  <update id="softDelete" parameterType="int">
    UPDATE users SET deleted_at = NOW() 
    WHERE id = #{id} AND deleted_at IS NULL
  </update>

  <!-- 统计查询 -->
  <select id="countByStatus" parameterType="int" resultType="long">
    SELECT COUNT(*) FROM users 
    WHERE status = #{status} AND deleted_at IS NULL
  </select>

</mapper>
```

---

## 🔍 复杂查询示例

### 1. 控制器中使用MyBatis

```go
package controllers

import (
    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/framework/orm"
)

type UserController struct {
    mvc.BaseController
}

// 获取用户详情(包含关联数据)
func (c *UserController) GetDetail() {
    id, err := c.GetInt("id")
    if err != nil {
        c.ErrorJSON(400, "无效的用户ID")
        return
    }
    
    // 使用MyBatis引擎执行复杂查询
    mybatisSession := orm.GetMyBatisSession()
    
    result, err := mybatisSession.SelectOne("user.selectUserDetail", map[string]interface{}{
        "userId": id,
    })
    if err != nil {
        c.ErrorJSON(500, "查询失败")
        return
    }
    
    if result == nil {
        c.ErrorJSON(404, "用户不存在")
        return
    }
    
    c.JSON(result)
}

// 用户列表查询(支持分页和筛选)
func (c *UserController) GetList() {
    page := c.GetInt("page", 1)
    size := c.GetInt("size", 20)
    status := c.GetInt("status", -1)
    keyword := c.GetString("keyword")
    
    // 构建查询参数
    params := map[string]interface{}{
        "page": page,
        "size": size,
    }
    
    if status >= 0 {
        params["status"] = status
    }
    if keyword != "" {
        params["keyword"] = keyword
    }
    
    mybatisSession := orm.GetMyBatisSession()
    
    // 获取总数
    totalCount, err := mybatisSession.SelectOne("user.countByParams", params)
    if err != nil {
        c.ErrorJSON(500, "查询总数失败")
        return
    }
    
    // 获取列表数据
    users, err := mybatisSession.SelectList("user.selectPageList", params)
    if err != nil {
        c.ErrorJSON(500, "查询列表失败")
        return
    }
    
    c.JSON(map[string]interface{}{
        "users":    users,
        "total":    totalCount,
        "page":     page,
        "size":     size,
    })
}
```

### 2. 复杂统计查询

```xml
<!-- 订单统计查询 -->
<select id="selectOrderStats" parameterType="map" resultType="map">
  SELECT 
    DATE_FORMAT(o.created_at, '%Y-%m') as month,
    COUNT(o.id) as order_count,
    SUM(o.total_amount) as total_amount,
    AVG(o.total_amount) as avg_amount,
    COUNT(DISTINCT o.user_id) as unique_users,
    
    -- 按状态分组统计
    SUM(CASE WHEN o.status = 1 THEN 1 ELSE 0 END) as pending_orders,
    SUM(CASE WHEN o.status = 2 THEN 1 ELSE 0 END) as paid_orders,
    SUM(CASE WHEN o.status = 3 THEN 1 ELSE 0 END) as shipped_orders,
    SUM(CASE WHEN o.status = 4 THEN 1 ELSE 0 END) as completed_orders,
    
    -- 计算转化率
    ROUND(
      SUM(CASE WHEN o.status = 4 THEN 1 ELSE 0 END) * 100.0 / COUNT(o.id), 
      2
    ) as completion_rate
    
  FROM orders o
  LEFT JOIN users u ON o.user_id = u.id
  WHERE 1=1
    <if test="startDate != null">
      AND o.created_at >= #{startDate}
    </if>
    <if test="endDate != null">
      AND o.created_at <![CDATA[<=]]> #{endDate}
    </if>
    <if test="userType != null">
      AND u.user_type = #{userType}
    </if>
  GROUP BY DATE_FORMAT(o.created_at, '%Y-%m')
  ORDER BY month DESC
  <if test="limit != null">
    LIMIT #{limit}
  </if>
</select>
```

### 3. 多表关联复杂查询

```xml
<!-- 用户订单详情查询 -->
<select id="selectUserOrderDetail" parameterType="map" resultMap="UserOrderDetailResultMap">
  SELECT 
    u.id as user_id,
    u.username,
    u.email,
    u.first_name,
    u.last_name,
    
    o.id as order_id,
    o.order_no,
    o.total_amount,
    o.status as order_status,
    o.created_at as order_date,
    
    oi.id as item_id,
    oi.quantity,
    oi.price as item_price,
    
    p.id as product_id,
    p.name as product_name,
    p.description as product_desc,
    
    c.id as category_id,
    c.name as category_name
    
  FROM users u
  INNER JOIN orders o ON u.id = o.user_id
  INNER JOIN order_items oi ON o.id = oi.order_id
  INNER JOIN products p ON oi.product_id = p.id
  LEFT JOIN categories c ON p.category_id = c.id
  
  WHERE u.id = #{userId}
    <if test="orderId != null">
      AND o.id = #{orderId}
    </if>
    <if test="status != null">
      AND o.status = #{status}
    </if>
    <if test="startDate != null">
      AND o.created_at >= #{startDate}
    </if>
    <if test="endDate != null">
      AND o.created_at <![CDATA[<=]]> #{endDate}
    </if>
  
  ORDER BY o.created_at DESC, oi.id ASC
</select>

<!-- 对应的ResultMap -->
<resultMap id="UserOrderDetailResultMap" type="map">
  <id property="userId" column="user_id"/>
  <result property="username" column="username"/>
  <result property="email" column="email"/>
  <result property="firstName" column="first_name"/>
  <result property="lastName" column="last_name"/>
  
  <!-- 订单集合 -->
  <collection property="orders" ofType="map">
    <id property="orderId" column="order_id"/>
    <result property="orderNo" column="order_no"/>
    <result property="totalAmount" column="total_amount"/>
    <result property="status" column="order_status"/>
    <result property="orderDate" column="order_date"/>
    
    <!-- 订单项集合 -->
    <collection property="items" ofType="map">
      <id property="itemId" column="item_id"/>
      <result property="quantity" column="quantity"/>
      <result property="price" column="item_price"/>
      
      <!-- 产品信息 -->
      <association property="product" javaType="map">
        <id property="productId" column="product_id"/>
        <result property="name" column="product_name"/>
        <result property="description" column="product_desc"/>
        
        <!-- 分类信息 -->
        <association property="category" javaType="map">
          <id property="categoryId" column="category_id"/>
          <result property="name" column="category_name"/>
        </association>
      </association>
    </collection>
  </collection>
</resultMap>
```

---

## 🔄 动态SQL使用

### 1. 动态查询条件

```xml
<!-- 动态条件查询 -->
<select id="selectByDynamicConditions" parameterType="map" resultMap="UserResultMap">
  SELECT * FROM users
  <where>
    deleted_at IS NULL
    <if test="id != null">
      AND id = #{id}
    </if>
    <if test="username != null and username != ''">
      AND username = #{username}
    </if>
    <if test="email != null and email != ''">
      AND email = #{email}
    </if>
    <if test="status != null">
      AND status = #{status}
    </if>
    <if test="startDate != null">
      AND created_at >= #{startDate}
    </if>
    <if test="endDate != null">
      AND created_at <![CDATA[<=]]> #{endDate}
    </if>
    <if test="userIds != null and userIds.size() > 0">
      AND id IN
      <foreach collection="userIds" item="id" open="(" separator="," close=")">
        #{id}
      </foreach>
    </if>
  </where>
  
  <!-- 动态排序 -->
  <if test="orderBy != null and orderBy != ''">
    ORDER BY ${orderBy}
    <if test="orderDirection != null and orderDirection != ''">
      ${orderDirection}
    </if>
  </if>
  <if test="orderBy == null or orderBy == ''">
    ORDER BY created_at DESC
  </if>
  
  <!-- 分页 -->
  <if test="limit != null">
    LIMIT #{limit}
    <if test="offset != null">
      OFFSET #{offset}
    </if>
  </if>
</select>
```

### 2. 动态更新字段

```xml
<!-- 动态更新 -->
<update id="updateSelective" parameterType="User">
  UPDATE users
  <set>
    <if test="username != null">username = #{username},</if>
    <if test="email != null">email = #{email},</if>
    <if test="firstName != null">first_name = #{firstName},</if>
    <if test="lastName != null">last_name = #{lastName},</if>
    <if test="status != null">status = #{status},</if>
    updated_at = NOW()
  </set>
  WHERE id = #{id} AND deleted_at IS NULL
</update>
```

### 3. choose-when-otherwise 条件分支

```xml
<!-- 条件分支查询 -->
<select id="selectByRole" parameterType="map" resultMap="UserResultMap">
  SELECT u.* FROM users u
  LEFT JOIN user_roles ur ON u.id = ur.user_id
  LEFT JOIN roles r ON ur.role_id = r.id
  WHERE u.deleted_at IS NULL
  
  <choose>
    <when test="roleType == 'admin'">
      AND r.name = 'admin'
    </when>
    <when test="roleType == 'user'">
      AND r.name = 'user'
    </when>
    <when test="roleType == 'guest'">
      AND u.id NOT IN (SELECT user_id FROM user_roles WHERE user_id IS NOT NULL)
    </when>
    <otherwise>
      AND 1=1
    </otherwise>
  </choose>
  
  ORDER BY u.created_at DESC
</select>
```

---

## 📞 存储过程调用

### 1. 简单存储过程

```sql
-- 创建存储过程
DELIMITER //
CREATE PROCEDURE GetUserStats(
    IN p_start_date DATE,
    IN p_end_date DATE,
    OUT p_total_users INT,
    OUT p_active_users INT,
    OUT p_new_users INT
)
BEGIN
    -- 获取总用户数
    SELECT COUNT(*) INTO p_total_users 
    FROM users 
    WHERE deleted_at IS NULL;
    
    -- 获取活跃用户数
    SELECT COUNT(*) INTO p_active_users 
    FROM users 
    WHERE status = 1 AND deleted_at IS NULL;
    
    -- 获取新增用户数
    SELECT COUNT(*) INTO p_new_users 
    FROM users 
    WHERE created_at BETWEEN p_start_date AND p_end_date 
    AND deleted_at IS NULL;
END //
DELIMITER ;
```

```xml
<!-- 调用存储过程 -->
<select id="getUserStats" statementType="CALLABLE" parameterType="map">
  {CALL GetUserStats(
    #{startDate,mode=IN,jdbcType=DATE},
    #{endDate,mode=IN,jdbcType=DATE},
    #{totalUsers,mode=OUT,jdbcType=INTEGER},
    #{activeUsers,mode=OUT,jdbcType=INTEGER},
    #{newUsers,mode=OUT,jdbcType=INTEGER}
  )}
</select>
```

### 2. 复杂存储过程

```sql
-- 用户数据分析存储过程
DELIMITER //
CREATE PROCEDURE AnalyzeUserBehavior(
    IN p_user_id INT,
    IN p_days INT
)
BEGIN
    DECLARE done INT DEFAULT FALSE;
    DECLARE temp_order_count INT;
    DECLARE temp_total_amount DECIMAL(10,2);
    
    -- 临时表存储结果
    CREATE TEMPORARY TABLE temp_user_analysis (
        metric_name VARCHAR(50),
        metric_value VARCHAR(100),
        metric_date DATE
    );
    
    -- 订单统计
    SELECT COUNT(*), COALESCE(SUM(total_amount), 0)
    INTO temp_order_count, temp_total_amount
    FROM orders 
    WHERE user_id = p_user_id 
    AND created_at >= DATE_SUB(CURDATE(), INTERVAL p_days DAY);
    
    INSERT INTO temp_user_analysis VALUES 
    ('order_count', temp_order_count, CURDATE()),
    ('total_spent', temp_total_amount, CURDATE());
    
    -- 返回结果
    SELECT * FROM temp_user_analysis;
    
    DROP TEMPORARY TABLE temp_user_analysis;
END //
DELIMITER ;
```

```xml
<!-- 调用复杂存储过程 -->
<select id="analyzeUserBehavior" statementType="CALLABLE" parameterType="map" resultType="map">
  {CALL AnalyzeUserBehavior(#{userId,mode=IN,jdbcType=INTEGER}, #{days,mode=IN,jdbcType=INTEGER})}
</select>
```

---

## 🚀 批量操作优化

### 1. 批量插入优化

```xml
<!-- 批量插入优化版本 -->
<insert id="batchInsertOptimized" parameterType="list">
  INSERT INTO users (username, email, password, first_name, last_name, status, created_at, updated_at)
  VALUES
  <foreach collection="list" item="user" separator="," index="index">
    (#{user.username}, #{user.email}, #{user.password}, 
     #{user.firstName}, #{user.lastName}, #{user.status}, NOW(), NOW())
    <if test="index != 0 and index % 1000 == 999">
      ; INSERT INTO users (username, email, password, first_name, last_name, status, created_at, updated_at) VALUES
    </if>
  </foreach>
</insert>
```

### 2. 批量更新

```xml
<!-- 批量更新 - 使用CASE WHEN -->
<update id="batchUpdate" parameterType="list">
  UPDATE users SET 
    status = CASE id
      <foreach collection="list" item="user">
        WHEN #{user.id} THEN #{user.status}
      </foreach>
    END,
    updated_at = NOW()
  WHERE id IN
  <foreach collection="list" item="user" open="(" separator="," close=")">
    #{user.id}
  </foreach>
</update>
```

### 3. 分页批量处理

```go
// 批量处理示例
func (c *UserController) PostBatchProcess() {
    var request struct {
        UserIDs []int `json:"user_ids"`
        Action  string `json:"action"`
        Status  int    `json:"status,omitempty"`
    }
    
    if err := c.BindJSON(&request); err != nil {
        c.ErrorJSON(400, "参数错误")
        return
    }
    
    mybatisSession := orm.GetMyBatisSession()
    
    // 分批处理，每批1000条
    batchSize := 1000
    totalProcessed := 0
    
    for i := 0; i < len(request.UserIDs); i += batchSize {
        end := i + batchSize
        if end > len(request.UserIDs) {
            end = len(request.UserIDs)
        }
        
        batchIDs := request.UserIDs[i:end]
        
        switch request.Action {
        case "activate":
            affected, err := mybatisSession.Update("user.batchUpdateStatus", map[string]interface{}{
                "userIds": batchIDs,
                "status":  1,
            })
            if err != nil {
                c.ErrorJSON(500, "批量激活失败")
                return
            }
            totalProcessed += affected
            
        case "deactivate":
            affected, err := mybatisSession.Update("user.batchUpdateStatus", map[string]interface{}{
                "userIds": batchIDs,
                "status":  0,
            })
            if err != nil {
                c.ErrorJSON(500, "批量停用失败")
                return
            }
            totalProcessed += affected
            
        case "delete":
            affected, err := mybatisSession.Update("user.batchSoftDelete", map[string]interface{}{
                "userIds": batchIDs,
            })
            if err != nil {
                c.ErrorJSON(500, "批量删除失败")
                return
            }
            totalProcessed += affected
        }
    }
    
    c.JSON(map[string]interface{}{
        "message":  "批量处理完成",
        "processed": totalProcessed,
    })
}
```

---

## 🤖 智能引擎选择

### 1. 智能选择器配置

```go
package services

import (
    "github.com/zsy619/yyhertz/framework/orm"
)

// 用户服务 - 演示智能引擎选择
type UserService struct {
    selector orm.SmartSelector
}

func NewUserService() *UserService {
    return &UserService{
        selector: orm.NewSmartSelector(),
    }
}

// 简单查询 - 自动选择GORM
func (s *UserService) GetUserByID(id uint) (*User, error) {
    var user User
    // 智能选择器会选择GORM引擎
    err := s.selector.Find(&user, id)
    return &user, err
}

// 复杂查询 - 自动选择MyBatis
func (s *UserService) GetUserOrderStats(userID int, days int) (map[string]interface{}, error) {
    // 智能选择器检测到复杂查询，会选择MyBatis引擎
    result, err := s.selector.ExecuteComplexQuery("user.analyzeUserBehavior", map[string]interface{}{
        "userId": userID,
        "days":   days,
    })
    
    if err != nil {
        return nil, err
    }
    
    if len(result) == 0 {
        return nil, nil
    }
    
    return result[0], nil
}

// 手动指定使用MyBatis引擎
func (s *UserService) GetComplexUserReport(params map[string]interface{}) ([]map[string]interface{}, error) {
    // 强制使用MyBatis引擎
    mybatisEngine := orm.GetEngine("mybatis")
    session := mybatisEngine.OpenSession()
    defer session.Close()
    
    return session.SelectList("user.complexUserReport", params)
}
```

### 2. 性能对比示例

```go
// 性能测试控制器
type BenchmarkController struct {
    mvc.BaseController
}

func (c *BenchmarkController) GetPerformanceComparison() {
    userID := c.GetInt("user_id", 1)
    
    // GORM引擎查询
    start := time.Now()
    gormResult, err := c.queryWithGORM(userID)
    gormDuration := time.Since(start)
    
    // MyBatis引擎查询
    start = time.Now()
    mybatisResult, err := c.queryWithMyBatis(userID)
    mybatisDuration := time.Since(start)
    
    // 智能选择器查询
    start = time.Now()
    smartResult, err := c.queryWithSmartSelector(userID)
    smartDuration := time.Since(start)
    
    c.JSON(map[string]interface{}{
        "user_id": userID,
        "performance": map[string]interface{}{
            "gorm": map[string]interface{}{
                "duration": gormDuration.Milliseconds(),
                "result_count": len(gormResult),
            },
            "mybatis": map[string]interface{}{
                "duration": mybatisDuration.Milliseconds(),
                "result_count": len(mybatisResult),
            },
            "smart_selector": map[string]interface{}{
                "duration": smartDuration.Milliseconds(),
                "result_count": len(smartResult),
            },
        },
        "recommendation": c.getEngineRecommendation(gormDuration, mybatisDuration, smartDuration),
    })
}

func (c *BenchmarkController) getEngineRecommendation(gormDur, mybatisDur, smartDur time.Duration) string {
    if smartDur <= gormDur && smartDur <= mybatisDur {
        return "智能选择器提供了最佳性能"
    } else if gormDur <= mybatisDur {
        return "GORM引擎在此场景下性能更好"
    } else {
        return "MyBatis引擎在此场景下性能更好"
    }
}
```

### 3. 引擎切换策略

```go
// 引擎切换策略配置
func setupEngineStrategy() {
    // 配置智能选择策略
    orm.ConfigureSmartSelector(&orm.SelectorConfig{
        // 简单查询阈值
        SimpleQueryThreshold: &orm.QueryThreshold{
            TableCount:    1,           // 单表查询
            JoinCount:     0,           // 无关联
            ConditionCount: 3,          // 条件数量 <= 3
            FunctionCount:  0,          // 无复杂函数
            SubQueryCount:  0,          // 无子查询
        },
        
        // 复杂查询阈值
        ComplexQueryThreshold: &orm.QueryThreshold{
            TableCount:    3,           // 多表查询
            JoinCount:     2,           // 有关联
            ConditionCount: 5,          // 条件数量 > 5
            FunctionCount:  2,          // 有复杂函数
            SubQueryCount:  1,          // 有子查询
        },
        
        // 默认引擎偏好
        DefaultEngine: "gorm",
        
        // 强制使用MyBatis的查询类型
        ForceMyBatisQueries: []string{
            "STATISTICAL",
            "ANALYTICAL", 
            "STORED_PROCEDURE",
            "BULK_OPERATION",
        },
        
        // 性能监控
        EnablePerformanceMonitoring: true,
        PerformanceLogThreshold:     100, // 超过100ms记录性能日志
    })
}
```

---

## 🔧 MyBatis最佳实践

### 1. SQL优化建议

```xml
<!-- ✅ 好的实践 -->
<select id="selectUsersOptimized" resultMap="UserResultMap">
  SELECT 
    id, username, email, first_name, last_name, status, created_at
  FROM users 
  WHERE status = #{status} 
    AND created_at >= #{startDate}
  ORDER BY created_at DESC
  LIMIT #{limit}
</select>

<!-- ❌ 避免的做法 -->
<select id="selectUsersBad" resultMap="UserResultMap">
  SELECT * FROM users  -- 避免使用 SELECT *
  WHERE 1=1           -- 避免恒真条件
    AND UPPER(username) LIKE UPPER(CONCAT('%', #{keyword}, '%'))  -- 避免函数包装索引字段
  ORDER BY id         -- 考虑是否需要排序
</select>
```

### 2. 缓存配置

```xml
<!-- 启用二级缓存 -->
<cache 
  eviction="LRU"
  flushInterval="60000" 
  size="512"
  readOnly="true"/>

<!-- 缓存引用 -->
<cache-ref namespace="user"/>

<!-- 查询缓存控制 -->
<select id="selectById" resultMap="UserResultMap" useCache="true">
  SELECT * FROM users WHERE id = #{id}
</select>

<!-- 不缓存的查询 -->
<select id="selectRealTimeStats" resultType="map" useCache="false">
  SELECT COUNT(*) as online_users 
  FROM user_sessions 
  WHERE last_activity > DATE_SUB(NOW(), INTERVAL 5 MINUTE)
</select>
```

### 3. 错误处理

```go
func (c *UserController) handleMyBatisError(err error) {
    if err == nil {
        return
    }
    
    // 根据错误类型进行处理
    switch {
    case strings.Contains(err.Error(), "Duplicate entry"):
        c.ErrorJSON(400, "数据已存在，请检查输入")
    case strings.Contains(err.Error(), "Data too long"):
        c.ErrorJSON(400, "数据长度超出限制")
    case strings.Contains(err.Error(), "Connection refused"):
        c.ErrorJSON(500, "数据库连接失败，请稍后重试")
    case strings.Contains(err.Error(), "Timeout"):
        c.ErrorJSON(500, "查询超时，请优化查询条件")
    default:
        // 记录详细错误到日志
        c.Logger.Error("MyBatis query error", "error", err.Error())
        c.ErrorJSON(500, "操作失败，请联系系统管理员")
    }
}
```

---

<div align="center">

**🔧 MyBatis引擎适用于复杂查询场景，与GORM互补使用！**

**智能选择器让你无需关心引擎切换，专注业务逻辑开发！🚀**

</div>