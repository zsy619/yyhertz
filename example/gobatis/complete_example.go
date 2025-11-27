// Package main GoBatis完整使用示例
//
// 展示了MyBatis-Go框架的各种功能：
// 1. 简化版Session使用
// 2. XML映射器使用
// 3. 钩子系统
// 4. DryRun调试模式
// 5. 分页查询
// 6. 事务管理
// 7. 性能监控
package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/zsy619/yyhertz/framework/mybatis"
)

// 使用models.go中定义的结构体，这里不再重复定义

// ExampleService 示例服务
type ExampleService struct {
	simpleSession mybatis.SimpleSession
	xmlSession    mybatis.XMLSession
	db            *gorm.DB
}

// NewExampleService 创建示例服务
func NewExampleService() (*ExampleService, error) {
	// 创建数据库连接
	db, err := gorm.Open(sqlite.Open("example.db"), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold: 200 * time.Millisecond,
				LogLevel:      logger.Warn,
				Colorful:      true,
			},
		),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 自动迁移数据库表
	err = db.AutoMigrate(&User{}, &UserProfile{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	// 创建简化版Session
	simpleSession := mybatis.NewSimpleSession(db).
		AddBeforeHook(auditHook()).
		AddAfterHook(performanceHook(100 * time.Millisecond))

	// 创建XML Session
	xmlSession := mybatis.NewXMLMapper(db)
	xmlSession.AddBeforeHook(auditHook())
	xmlSession.AddAfterHook(performanceHook(100 * time.Millisecond))

	// 加载XML映射
	err = xmlSession.LoadMapperXMLFromString(getUserMapperXML())
	if err != nil {
		return nil, fmt.Errorf("failed to load XML mapper: %w", err)
	}

	return &ExampleService{
		simpleSession: simpleSession,
		xmlSession:    xmlSession,
		db:            db,
	}, nil
}

func main() {
	service, err := NewExampleService()
	if err != nil {
		log.Fatal("Failed to create service:", err)
	}

	ctx := context.Background()

	// 运行示例
	examples := []struct {
		name string
		fn   func(*ExampleService, context.Context) error
	}{
		{"基础CRUD操作", (*ExampleService).basicCRUDExample},
		{"DryRun调试模式", (*ExampleService).dryRunExample},
		{"分页查询示例", (*ExampleService).paginationExample},
		{"XML映射器示例", (*ExampleService).xmlMapperExample},
		{"批量操作示例", (*ExampleService).batchOperationExample},
		{"复杂查询示例", (*ExampleService).complexQueryExample},
		{"事务管理示例", (*ExampleService).transactionExample},
		{"钩子系统示例", (*ExampleService).hooksExample},
		{"性能监控示例", (*ExampleService).performanceMonitoringExample},
	}

	for _, example := range examples {
		fmt.Printf("\n=== %s ===\n", example.name)
		if err := example.fn(service, ctx); err != nil {
			log.Printf("示例 '%s' 执行失败: %v", example.name, err)
		} else {
			fmt.Printf("✅ %s 执行成功\n", example.name)
		}
	}
}

// basicCRUDExample 基础CRUD操作示例
func (s *ExampleService) basicCRUDExample(ctx context.Context) error {
	fmt.Println("📝 执行基础CRUD操作...")

	// 1. 插入用户
	insertSQL := `
		INSERT INTO users (name, email, age, status, phone, birthday) 
		VALUES (?, ?, ?, ?, ?, ?)
	`
	
	birthday := time.Date(1995, 6, 15, 0, 0, 0, 0, time.UTC)
	userID, err := s.simpleSession.Insert(ctx, insertSQL,
		"张三", "zhangsan@example.com", 28, "active", "13800138001", birthday)
	if err != nil {
		return fmt.Errorf("插入用户失败: %w", err)
	}
	fmt.Printf("插入用户成功，ID: %d\n", userID)

	// 2. 查询单个用户
	selectSQL := "SELECT * FROM users WHERE id = ?"
	user, err := s.simpleSession.SelectOne(ctx, selectSQL, userID)
	if err != nil {
		return fmt.Errorf("查询用户失败: %w", err)
	}
	fmt.Printf("查询用户成功: %+v\n", user)

	// 3. 更新用户
	updateSQL := "UPDATE users SET age = ?, status = ? WHERE id = ?"
	affected, err := s.simpleSession.Update(ctx, updateSQL, 29, "updated", userID)
	if err != nil {
		return fmt.Errorf("更新用户失败: %w", err)
	}
	fmt.Printf("更新用户成功，影响行数: %d\n", affected)

	// 4. 查询多个用户
	listSQL := "SELECT * FROM users WHERE status = ? LIMIT 5"
	users, err := s.simpleSession.SelectList(ctx, listSQL, "active")
	if err != nil {
		return fmt.Errorf("查询用户列表失败: %w", err)
	}
	fmt.Printf("查询用户列表成功，共 %d 条记录\n", len(users))

	return nil
}

// dryRunExample DryRun调试模式示例
func (s *ExampleService) dryRunExample(ctx context.Context) error {
	fmt.Println("🔍 DryRun调试模式演示...")

	// 创建DryRun会话
	dryRunSession := mybatis.NewSimpleSession(s.db).
		DryRun(true).
		Debug(true)

	fmt.Println("以下SQL将只预览，不实际执行：")

	// 查询操作
	_, err := dryRunSession.SelectOne(ctx, "SELECT * FROM users WHERE email = ?", "debug@example.com")
	if err != nil {
		return err
	}

	// 插入操作
	_, err = dryRunSession.Insert(ctx,
		"INSERT INTO users (name, email, age) VALUES (?, ?, ?)",
		"调试用户", "debug@example.com", 25)
	if err != nil {
		return err
	}

	// 更新操作
	_, err = dryRunSession.Update(ctx, "UPDATE users SET status = ? WHERE email = ?", "debug", "debug@example.com")
	if err != nil {
		return err
	}

	// 删除操作
	_, err = dryRunSession.Delete(ctx, "DELETE FROM users WHERE email = ?", "debug@example.com")
	if err != nil {
		return err
	}

	return nil
}

// paginationExample 分页查询示例
func (s *ExampleService) paginationExample(ctx context.Context) error {
	fmt.Println("📄 分页查询演示...")

	// 先插入一些测试数据
	for i := 1; i <= 50; i++ {
		_, err := s.simpleSession.Insert(ctx,
			"INSERT IGNORE INTO users (name, email, age, status) VALUES (?, ?, ?, ?)",
			fmt.Sprintf("分页用户%d", i),
			fmt.Sprintf("page_%d@example.com", i),
			rand.Intn(50)+20,
			[]string{"active", "inactive", "pending"}[rand.Intn(3)])
		if err != nil {
			log.Printf("插入测试数据失败: %v", err)
		}
	}

	// 执行分页查询
	pageSQL := "SELECT * FROM users WHERE status = 'active' ORDER BY id"
	pageRequest := mybatis.PageRequest{
		Page: 1,
		Size: 10,
	}

	pageResult, err := s.simpleSession.SelectPage(ctx, pageSQL, pageRequest)
	if err != nil {
		return fmt.Errorf("分页查询失败: %w", err)
	}

	fmt.Printf("分页查询结果:\n")
	fmt.Printf("  总记录数: %d\n", pageResult.Total)
	fmt.Printf("  当前页: %d\n", pageResult.Page)
	fmt.Printf("  每页大小: %d\n", pageResult.Size)
	fmt.Printf("  总页数: %d\n", pageResult.TotalPages)
	fmt.Printf("  当前页数据: %d 条\n", len(pageResult.Items))

	return nil
}

// xmlMapperExample XML映射器示例
func (s *ExampleService) xmlMapperExample(ctx context.Context) error {
	fmt.Println("🗺️ XML映射器演示...")

	// 使用XML映射器查询用户
	user, err := s.xmlSession.SelectOneByID(ctx, "UserMapper.selectById", 1)
	if err != nil {
		return fmt.Errorf("XML查询用户失败: %w", err)
	}
	fmt.Printf("XML查询用户成功: %+v\n", user)

	// 动态SQL查询
	query := UserQuery{
		Status: "active",
		AgeMin: 25,
		AgeMax: 35,
	}
	users, err := s.xmlSession.SelectListByID(ctx, "UserMapper.selectByCondition", query)
	if err != nil {
		return fmt.Errorf("XML动态查询失败: %w", err)
	}
	fmt.Printf("XML动态查询成功，共 %d 条记录\n", len(users))

	// XML分页查询
	pageResult, err := s.xmlSession.SelectPageByID(ctx, "UserMapper.selectByCondition", query, mybatis.PageRequest{
		Page: 1,
		Size: 5,
	})
	if err != nil {
		return fmt.Errorf("XML分页查询失败: %w", err)
	}
	fmt.Printf("XML分页查询成功，总记录数: %d，当前页: %d 条\n", pageResult.Total, len(pageResult.Items))

	return nil
}

// batchOperationExample 批量操作示例
func (s *ExampleService) batchOperationExample(ctx context.Context) error {
	fmt.Println("📦 批量操作演示...")

	// 批量插入
	batchSize := 20
	fmt.Printf("批量插入 %d 条记录...\n", batchSize)

	start := time.Now()
	for i := 1; i <= batchSize; i++ {
		_, err := s.simpleSession.Insert(ctx,
			"INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)",
			fmt.Sprintf("批量用户%d", i),
			fmt.Sprintf("batch_%d@example.com", i),
			rand.Intn(40)+20,
			"active")
		if err != nil {
			log.Printf("批量插入第 %d 条记录失败: %v", i, err)
		}
	}
	fmt.Printf("批量插入完成，耗时: %v\n", time.Since(start))

	// 批量更新
	updateStart := time.Now()
	affected, err := s.simpleSession.Update(ctx,
		"UPDATE users SET status = 'batch_updated' WHERE name LIKE 'batch_%'")
	if err != nil {
		return fmt.Errorf("批量更新失败: %w", err)
	}
	fmt.Printf("批量更新完成，影响 %d 行，耗时: %v\n", affected, time.Since(updateStart))

	return nil
}

// complexQueryExample 复杂查询示例
func (s *ExampleService) complexQueryExample(ctx context.Context) error {
	fmt.Println("🔍 复杂查询演示...")

	// 聚合查询
	countResult, err := s.simpleSession.SelectOne(ctx,
		"SELECT COUNT(*) as total, AVG(age) as avg_age, MIN(age) as min_age, MAX(age) as max_age FROM users WHERE status = ?",
		"active")
	if err != nil {
		return fmt.Errorf("聚合查询失败: %w", err)
	}
	fmt.Printf("聚合查询结果: %+v\n", countResult)

	// 分组查询
	groupResult, err := s.simpleSession.SelectList(ctx,
		"SELECT status, COUNT(*) as count, AVG(age) as avg_age FROM users GROUP BY status ORDER BY count DESC")
	if err != nil {
		return fmt.Errorf("分组查询失败: %w", err)
	}
	fmt.Printf("分组查询结果: %+v\n", groupResult)

	// 范围查询
	rangeResult, err := s.simpleSession.SelectList(ctx,
		"SELECT * FROM users WHERE age BETWEEN ? AND ? AND status = ? ORDER BY age LIMIT 5",
		25, 35, "active")
	if err != nil {
		return fmt.Errorf("范围查询失败: %w", err)
	}
	fmt.Printf("范围查询结果: %d 条记录\n", len(rangeResult))

	return nil
}

// transactionExample 事务管理示例
func (s *ExampleService) transactionExample(ctx context.Context) error {
	fmt.Println("💳 事务管理演示...")

	// 开始事务
	tx := s.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("开始事务失败: %w", tx.Error)
	}

	// 创建事务会话
	txSession := mybatis.NewSimpleSession(tx)

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			fmt.Printf("事务回滚: %v\n", r)
		}
	}()

	// 在事务中执行操作
	userID, err := txSession.Insert(ctx,
		"INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)",
		"事务用户", "tx@example.com", 30, "active")
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("事务插入失败: %w", err)
	}
	fmt.Printf("事务中插入用户，ID: %d\n", userID)

	// 插入用户档案
	_, err = txSession.Insert(ctx,
		"INSERT INTO user_profiles (user_id, bio, location) VALUES (?, ?, ?)",
		userID, "这是一个事务用户", "事务城市")
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("事务插入档案失败: %w", err)
	}
	fmt.Printf("事务中插入用户档案\n")

	// 模拟业务逻辑
	if userID > 0 {
		// 提交事务
		if err := tx.Commit().Error; err != nil {
			return fmt.Errorf("事务提交失败: %w", err)
		}
		fmt.Printf("事务提交成功\n")
	} else {
		// 回滚事务
		tx.Rollback()
		fmt.Printf("事务回滚\n")
	}

	return nil
}

// hooksExample 钩子系统示例
func (s *ExampleService) hooksExample(ctx context.Context) error {
	fmt.Println("🎣 钩子系统演示...")

	// 创建带多个钩子的会话
	hookSession := mybatis.NewSimpleSession(s.db).
		AddBeforeHook(func(ctx context.Context, sql string, args []interface{}) error {
			fmt.Printf("🔵 Before Hook 1: SQL长度 = %d\n", len(sql))
			return nil
		}).
		AddBeforeHook(func(ctx context.Context, sql string, args []interface{}) error {
			fmt.Printf("🔵 Before Hook 2: 参数数量 = %d\n", len(args))
			return nil
		}).
		AddAfterHook(func(ctx context.Context, result interface{}, duration time.Duration, err error) {
			if err != nil {
				fmt.Printf("🔴 After Hook: 执行失败，耗时 %v，错误: %v\n", duration, err)
			} else {
				fmt.Printf("🟢 After Hook: 执行成功，耗时 %v\n", duration)
			}
		})

	// 执行一个查询触发钩子
	_, err := hookSession.SelectOne(ctx, "SELECT * FROM users LIMIT 1")
	if err != nil {
		return fmt.Errorf("钩子测试查询失败: %w", err)
	}

	// 执行一个插入触发钩子
	_, err = hookSession.Insert(ctx,
		"INSERT INTO users (name, email, age) VALUES (?, ?, ?)",
		"钩子用户", "hooks@example.com", 25)
	if err != nil {
		log.Printf("钩子测试插入失败: %v", err)
	}

	return nil
}

// performanceMonitoringExample 性能监控示例
func (s *ExampleService) performanceMonitoringExample(ctx context.Context) error {
	fmt.Println("📊 性能监控演示...")

	// 统计变量
	var totalQueries int
	var totalDuration time.Duration
	var slowQueries int

	// 创建带性能监控的会话
	monitorSession := mybatis.NewSimpleSession(s.db).
		AddAfterHook(func(ctx context.Context, result interface{}, duration time.Duration, err error) {
			totalQueries++
			totalDuration += duration
			
			if duration > 50*time.Millisecond {
				slowQueries++
				fmt.Printf("⚠️ 慢查询检测: 耗时 %v\n", duration)
			}
		}).
		Debug(false) // 关闭调试输出，专注性能数据

	// 执行多个查询进行性能测试
	queries := []struct {
		name string
		sql  string
		args []interface{}
	}{
		{"快速查询", "SELECT COUNT(*) FROM users", nil},
		{"索引查询", "SELECT * FROM users WHERE id = ?", []interface{}{1}},
		{"范围查询", "SELECT * FROM users WHERE age BETWEEN ? AND ?", []interface{}{25, 35}},
		{"排序查询", "SELECT * FROM users ORDER BY created_at DESC LIMIT 10", nil},
		{"复杂查询", "SELECT status, COUNT(*) FROM users GROUP BY status", nil},
	}

	fmt.Printf("执行 %d 个测试查询...\n", len(queries))
	start := time.Now()

	for _, query := range queries {
		queryStart := time.Now()
		_, err := monitorSession.SelectList(ctx, query.sql, query.args...)
		queryDuration := time.Since(queryStart)
		
		if err != nil {
			fmt.Printf("❌ %s 失败: %v\n", query.name, err)
		} else {
			fmt.Printf("✅ %s 完成，耗时: %v\n", query.name, queryDuration)
		}
	}

	totalTestTime := time.Since(start)

	// 输出性能统计
	fmt.Printf("\n📈 性能统计:\n")
	fmt.Printf("  总查询数: %d\n", totalQueries)
	fmt.Printf("  总耗时: %v\n", totalDuration)
	fmt.Printf("  平均耗时: %v\n", totalDuration/time.Duration(totalQueries))
	fmt.Printf("  测试总时间: %v\n", totalTestTime)
	fmt.Printf("  慢查询数: %d (%.1f%%)\n", slowQueries, float64(slowQueries)/float64(totalQueries)*100)

	return nil
}

// auditHook 审计钩子
func auditHook() mybatis.BeforeHook {
	return func(ctx context.Context, sql string, args []interface{}) error {
		// 简化的审计日志
		if len(sql) > 100 {
			log.Printf("[AUDIT] SQL: %s... (参数: %d个)", sql[:100], len(args))
		} else {
			log.Printf("[AUDIT] SQL: %s (参数: %d个)", sql, len(args))
		}
		return nil
	}
}

// performanceHook 性能监控钩子
func performanceHook(slowThreshold time.Duration) mybatis.AfterHook {
	return func(ctx context.Context, result interface{}, duration time.Duration, err error) {
		if duration > slowThreshold {
			log.Printf("[PERF] 慢查询检测: %v (阈值: %v)", duration, slowThreshold)
		}
		
		if err != nil {
			log.Printf("[PERF] 查询失败: %v, 耗时: %v", err, duration)
		}
	}
}

// getUserMapperXML 获取用户映射XML
func getUserMapperXML() string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE mapper PUBLIC "-//mybatis.org//DTD Mapper 3.0//EN" "http://mybatis.org/dtd/mybatis-3-mapper.dtd">
<mapper namespace="UserMapper">
    <select id="selectById" parameterType="int" resultType="map">
        SELECT * FROM users WHERE id = #{id}
    </select>
    
    <select id="selectByStatus" parameterType="string" resultType="map">
        SELECT * FROM users WHERE status = #{status} LIMIT 10
    </select>
    
    <insert id="insertUser" parameterType="User">
        INSERT INTO users (name, email, age, status) 
        VALUES (#{name}, #{email}, #{age}, #{status})
    </insert>
</mapper>`
}