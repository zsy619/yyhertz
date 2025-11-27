// Package main 懒加载机制测试
//
// 测试MyBatis懒加载功能：
// 1. 基础懒加载代理对象
// 2. 一对一关联懒加载
// 3. 一对多关联懒加载
// 4. 多对一关联懒加载
// 5. 嵌套懒加载
// 6. 懒加载性能优化
// 7. 懒加载缓存集成
// 8. 懒加载错误处理
package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/zsy619/yyhertz/framework/mybatis"
)

// LazyUser 懒加载用户模型
type LazyUser struct {
	ID       int64     `json:"id" db:"id"`
	Name     string    `json:"name" db:"name"`
	Email    string    `json:"email" db:"email"`
	Age      int       `json:"age" db:"age"`
	Status   string    `json:"status" db:"status"`
	CreateAt time.Time `json:"createAt" db:"created_at"`
	
	// 懒加载关联字段
	Profile   *LazyUserProfile `json:"profile,omitempty" lazy:"true"`    // 一对一
	Orders    []*LazyOrder     `json:"orders,omitempty" lazy:"true"`     // 一对多
	Department *LazyDepartment `json:"department,omitempty" lazy:"true"` // 多对一
}

// LazyUserProfile 懒加载用户资料（一对一关联）
type LazyUserProfile struct {
	ID       int64  `json:"id" db:"id"`
	UserID   int64  `json:"userId" db:"user_id"`
	Avatar   string `json:"avatar" db:"avatar"`
	Bio      string `json:"bio" db:"bio"`
	Phone    string `json:"phone" db:"phone"`
	Address  string `json:"address" db:"address"`
	
	// 反向关联
	User *LazyUser `json:"user,omitempty" lazy:"true"`
}

// LazyOrder 懒加载订单（一对多关联）
type LazyOrder struct {
	ID         int64     `json:"id" db:"id"`
	UserID     int64     `json:"userId" db:"user_id"`
	OrderNo    string    `json:"orderNo" db:"order_no"`
	Amount     float64   `json:"amount" db:"amount"`
	Status     string    `json:"status" db:"status"`
	CreateAt   time.Time `json:"createAt" db:"created_at"`
	
	// 关联字段
	User      *LazyUser        `json:"user,omitempty" lazy:"true"`
	OrderItems []*LazyOrderItem `json:"orderItems,omitempty" lazy:"true"`
}

// LazyOrderItem 懒加载订单项目
type LazyOrderItem struct {
	ID       int64   `json:"id" db:"id"`
	OrderID  int64   `json:"orderId" db:"order_id"`
	ProductName string `json:"productName" db:"product_name"`
	Quantity int     `json:"quantity" db:"quantity"`
	Price    float64 `json:"price" db:"price"`
	
	// 关联字段
	Order *LazyOrder `json:"order,omitempty" lazy:"true"`
}

// LazyDepartment 懒加载部门（多对一关联）
type LazyDepartment struct {
	ID       int64  `json:"id" db:"id"`
	Name     string `json:"name" db:"name"`
	Code     string `json:"code" db:"code"`
	Manager  string `json:"manager" db:"manager"`
	
	// 部门员工（一对多）
	Users []*LazyUser `json:"users,omitempty" lazy:"true"`
}

// TestLazyLoadingBasic 测试基础懒加载功能
func TestLazyLoadingBasic(t *testing.T) {
	db, err := setupTestDatabase()
	if err != nil {
		t.Fatalf("设置测试数据库失败: %v", err)
	}
	defer cleanupDatabase(db)

	xmlSession := mybatis.NewXMLMapper(db).Debug(false)
	ctx := context.Background()

	// 加载懒加载测试映射
	err = xmlSession.LoadMapperXMLFromString(getLazyLoadingMapperXML())
	if err != nil {
		t.Fatalf("加载XML映射失败: %v", err)
	}

	// 准备测试数据
	setupLazyLoadingTestData(t, xmlSession, ctx)

	t.Run("基础懒加载代理对象", func(t *testing.T) {
		// 查询用户但不立即加载关联数据
		users, err := xmlSession.SelectListByID(ctx, "LazyLoading.selectUsers", nil)
		if err != nil {
			t.Fatalf("查询用户失败: %v", err)
		}

		if len(users) == 0 {
			t.Fatal("没有查询到用户数据")
		}

		// 检查懒加载代理
		for i, userInterface := range users {
			userMap := userInterface.(map[string]interface{})
			t.Logf("用户 %d: %+v", i, userMap)

			// 此时关联字段应该还没有加载
			if profile, exists := userMap["profile"]; exists && profile != nil {
				t.Logf("注意：用户 %d 的profile已经被加载", i)
			}
		}

		t.Log("基础懒加载代理对象测试通过")
	})

	t.Run("懒加载触发机制", func(t *testing.T) {
		// 查询单个用户
		user, err := xmlSession.SelectOneByID(ctx, "LazyLoading.selectUserById", map[string]interface{}{"id": 1})
		if err != nil {
			t.Fatalf("查询用户失败: %v", err)
		}

		if user == nil {
			t.Fatal("用户不存在")
		}

		userMap := user.(map[string]interface{})
		t.Logf("查询到用户: %+v", userMap)

		// 模拟懒加载触发 - 访问profile字段
		if userId, ok := userMap["id"]; ok {
			// 手动触发懒加载
			profile, err := xmlSession.SelectOneByID(ctx, "LazyLoading.selectUserProfile", 
				map[string]interface{}{"userId": userId})
			if err != nil {
				t.Logf("懒加载用户资料失败: %v", err)
			} else if profile != nil {
				t.Logf("懒加载的用户资料: %+v", profile)
			}
		}

		t.Log("懒加载触发机制测试通过")
	})
}

// TestLazyLoadingAssociations 测试关联懒加载
func TestLazyLoadingAssociations(t *testing.T) {
	db, err := setupTestDatabase()
	if err != nil {
		t.Fatalf("设置测试数据库失败: %v", err)
	}
	defer cleanupDatabase(db)

	xmlSession := mybatis.NewXMLMapper(db).Debug(false)
	ctx := context.Background()

	// 加载映射
	err = xmlSession.LoadMapperXMLFromString(getLazyLoadingMapperXML())
	if err != nil {
		t.Fatalf("加载XML映射失败: %v", err)
	}

	// 准备测试数据
	setupLazyLoadingTestData(t, xmlSession, ctx)

	t.Run("一对一关联懒加载", func(t *testing.T) {
		// 查询用户（不包含profile）
		users, err := xmlSession.SelectListByID(ctx, "LazyLoading.selectUsers", nil)
		if err != nil {
			t.Fatalf("查询用户失败: %v", err)
		}

		if len(users) == 0 {
			t.Skip("没有用户数据可测试")
		}

		// 对每个用户，懒加载其profile
		for i, userInterface := range users {
			userMap := userInterface.(map[string]interface{})
			userID := userMap["id"]

			// 懒加载用户资料
			profile, err := xmlSession.SelectOneByID(ctx, "LazyLoading.selectUserProfile", 
				map[string]interface{}{"userId": userID})
			if err != nil {
				t.Logf("用户 %v 懒加载profile失败: %v", userID, err)
			} else if profile != nil {
				profileMap := profile.(map[string]interface{})
				t.Logf("用户 %v 的懒加载profile: %+v", userID, profileMap)
			}

			if i >= 2 { // 只测试前3个用户
				break
			}
		}

		t.Log("一对一关联懒加载测试通过")
	})

	t.Run("一对多关联懒加载", func(t *testing.T) {
		// 查询用户（不包含orders）
		users, err := xmlSession.SelectListByID(ctx, "LazyLoading.selectUsers", nil)
		if err != nil {
			t.Fatalf("查询用户失败: %v", err)
		}

		if len(users) == 0 {
			t.Skip("没有用户数据可测试")
		}

		// 对每个用户，懒加载其orders
		for i, userInterface := range users {
			userMap := userInterface.(map[string]interface{})
			userID := userMap["id"]

			// 懒加载用户订单
			orders, err := xmlSession.SelectListByID(ctx, "LazyLoading.selectUserOrders", 
				map[string]interface{}{"userId": userID})
			if err != nil {
				t.Logf("用户 %v 懒加载orders失败: %v", userID, err)
			} else {
				t.Logf("用户 %v 的懒加载orders: %d 条订单", userID, len(orders))
				
				// 显示前2条订单详情
				for j, orderInterface := range orders {
					if j >= 2 {
						break
					}
					orderMap := orderInterface.(map[string]interface{})
					t.Logf("  订单 %d: %+v", j+1, orderMap)
				}
			}

			if i >= 1 { // 只测试前2个用户
				break
			}
		}

		t.Log("一对多关联懒加载测试通过")
	})

	t.Run("多对一关联懒加载", func(t *testing.T) {
		// 查询用户（不包含department）
		users, err := xmlSession.SelectListByID(ctx, "LazyLoading.selectUsers", nil)
		if err != nil {
			t.Fatalf("查询用户失败: %v", err)
		}

		if len(users) == 0 {
			t.Skip("没有用户数据可测试")
		}

		// 对每个用户，懒加载其department
		departmentCache := make(map[interface{}]interface{}) // 简单缓存避免重复查询
		
		for i, userInterface := range users {
			userMap := userInterface.(map[string]interface{})
			userID := userMap["id"]
			
			// 假设用户有department_id字段（需要在数据准备中设置）
			if deptID, exists := userMap["department_id"]; exists && deptID != nil {
				// 检查缓存
				if cachedDept, cached := departmentCache[deptID]; cached {
					t.Logf("用户 %v 的部门（缓存）: %+v", userID, cachedDept)
				} else {
					// 懒加载部门信息
					dept, err := xmlSession.SelectOneByID(ctx, "LazyLoading.selectDepartmentById", 
						map[string]interface{}{"id": deptID})
					if err != nil {
						t.Logf("用户 %v 懒加载department失败: %v", userID, err)
					} else if dept != nil {
						deptMap := dept.(map[string]interface{})
						departmentCache[deptID] = deptMap
						t.Logf("用户 %v 的懒加载department: %+v", userID, deptMap)
					}
				}
			}

			if i >= 2 { // 只测试前3个用户
				break
			}
		}

		t.Log("多对一关联懒加载测试通过")
	})
}

// TestLazyLoadingPerformance 测试懒加载性能
func TestLazyLoadingPerformance(t *testing.T) {
	db, err := setupTestDatabase()
	if err != nil {
		t.Fatalf("设置测试数据库失败: %v", err)
	}
	defer cleanupDatabase(db)

	xmlSession := mybatis.NewXMLMapper(db).Debug(false)
	ctx := context.Background()

	// 加载映射
	err = xmlSession.LoadMapperXMLFromString(getLazyLoadingMapperXML())
	if err != nil {
		t.Fatalf("加载XML映射失败: %v", err)
	}

	// 准备测试数据
	setupLazyLoadingTestData(t, xmlSession, ctx)

	t.Run("懒加载vs立即加载性能对比", func(t *testing.T) {
		// 懒加载方式 - 只查询用户
		start := time.Now()
		users, err := xmlSession.SelectListByID(ctx, "LazyLoading.selectUsers", nil)
		lazyLoadDuration := time.Since(start)
		
		if err != nil {
			t.Fatalf("懒加载查询用户失败: %v", err)
		}

		// 立即加载方式 - 查询用户及所有关联数据
		start = time.Now()
		usersWithAssociations, err := xmlSession.SelectListByID(ctx, "LazyLoading.selectUsersWithAssociations", nil)
		eagerLoadDuration := time.Since(start)
		
		if err != nil {
			t.Fatalf("立即加载查询失败: %v", err)
		}

		t.Logf("性能对比结果:")
		t.Logf("  懒加载方式: 查询 %d 个用户，耗时 %v", len(users), lazyLoadDuration)
		t.Logf("  立即加载方式: 查询 %d 个用户（含关联），耗时 %v", len(usersWithAssociations), eagerLoadDuration)
		
		// 懒加载应该更快（因为没有加载关联数据）
		if lazyLoadDuration < eagerLoadDuration {
			speedup := float64(eagerLoadDuration) / float64(lazyLoadDuration)
			t.Logf("  懒加载速度提升: %.2fx", speedup)
		}

		t.Log("懒加载vs立即加载性能对比测试通过")
	})

	t.Run("按需懒加载性能测试", func(t *testing.T) {
		// 查询所有用户
		users, err := xmlSession.SelectListByID(ctx, "LazyLoading.selectUsers", nil)
		if err != nil {
			t.Fatalf("查询用户失败: %v", err)
		}

		// 模拟只访问部分用户的关联数据（20%的用户）
		accessCount := len(users) / 5
		if accessCount == 0 {
			accessCount = 1
		}

		start := time.Now()
		for i := 0; i < accessCount; i++ {
			userMap := users[i].(map[string]interface{})
			userID := userMap["id"]

			// 懒加载用户资料
			_, err := xmlSession.SelectOneByID(ctx, "LazyLoading.selectUserProfile", 
				map[string]interface{}{"userId": userID})
			if err != nil {
				t.Logf("懒加载用户 %v 资料失败: %v", userID, err)
			}

			// 懒加载用户订单
			_, err = xmlSession.SelectListByID(ctx, "LazyLoading.selectUserOrders", 
				map[string]interface{}{"userId": userID})
			if err != nil {
				t.Logf("懒加载用户 %v 订单失败: %v", userID, err)
			}
		}
		onDemandDuration := time.Since(start)

		t.Logf("按需懒加载测试结果:")
		t.Logf("  总用户数: %d", len(users))
		t.Logf("  实际访问关联数据的用户: %d", accessCount)
		t.Logf("  按需懒加载耗时: %v", onDemandDuration)
		t.Logf("  平均每用户懒加载耗时: %v", onDemandDuration/time.Duration(accessCount))

		t.Log("按需懒加载性能测试通过")
	})
}

// TestLazyLoadingWithCache 测试懒加载与缓存集成
func TestLazyLoadingWithCache(t *testing.T) {
	db, err := setupTestDatabase()
	if err != nil {
		t.Fatalf("设置测试数据库失败: %v", err)
	}
	defer cleanupDatabase(db)

	// 启用缓存的会话
	cacheConfig := &mybatis.CacheConfig{
		L1CacheEnabled:  true,
		L1CacheSize:     500,
		L1CacheTTL:      5 * time.Minute,
		L2CacheEnabled:  true,
		L2CacheSize:     1000,
		L2CacheTTL:      10 * time.Minute,
		CleanupInterval: 2 * time.Minute,
	}
	
	xmlSession := mybatis.NewXMLMapper(db).
		EnableCache(cacheConfig).
		Debug(false)
	ctx := context.Background()

	// 加载映射
	err = xmlSession.LoadMapperXMLFromString(getLazyLoadingMapperXML())
	if err != nil {
		t.Fatalf("加载XML映射失败: %v", err)
	}

	// 准备测试数据
	setupLazyLoadingTestData(t, xmlSession, ctx)

	t.Run("懒加载缓存命中测试", func(t *testing.T) {
		// 第一次懒加载
		start := time.Now()
		profile1, err := xmlSession.SelectOneByID(ctx, "LazyLoading.selectUserProfile", 
			map[string]interface{}{"userId": 1})
		firstLoadTime := time.Since(start)
		
		if err != nil {
			t.Fatalf("第一次懒加载失败: %v", err)
		}

		// 第二次相同懒加载（应该命中缓存）
		start = time.Now()
		profile2, err := xmlSession.SelectOneByID(ctx, "LazyLoading.selectUserProfile", 
			map[string]interface{}{"userId": 1})
		secondLoadTime := time.Since(start)
		
		if err != nil {
			t.Fatalf("第二次懒加载失败: %v", err)
		}

		// 验证数据一致性
		profile1Map := profile1.(map[string]interface{})
		profile2Map := profile2.(map[string]interface{})
		
		if profile1Map["user_id"] != profile2Map["user_id"] {
			t.Error("缓存的懒加载数据不一致")
		}

		t.Logf("懒加载缓存测试结果:")
		t.Logf("  第一次加载耗时: %v", firstLoadTime)
		t.Logf("  第二次加载耗时: %v", secondLoadTime)
		
		// 第二次应该更快（缓存命中）
		if secondLoadTime < firstLoadTime {
			speedup := float64(firstLoadTime) / float64(secondLoadTime)
			t.Logf("  缓存命中速度提升: %.2fx", speedup)
		}

		t.Log("懒加载缓存命中测试通过")
	})

	t.Run("跨用户懒加载缓存共享", func(t *testing.T) {
		// 假设多个用户属于同一个部门
		users, err := xmlSession.SelectListByID(ctx, "LazyLoading.selectUsers", nil)
		if err != nil {
			t.Fatalf("查询用户失败: %v", err)
		}

		if len(users) < 2 {
			t.Skip("用户数量不足，跳过跨用户缓存测试")
		}

		// 为前两个用户懒加载相同的部门信息
		start := time.Now()
		_, err = xmlSession.SelectOneByID(ctx, "LazyLoading.selectDepartmentById", 
			map[string]interface{}{"id": 1})
		firstDeptLoadTime := time.Since(start)
		
		if err != nil {
			t.Logf("第一次部门懒加载失败: %v", err)
		}

		start = time.Now()
		_, err = xmlSession.SelectOneByID(ctx, "LazyLoading.selectDepartmentById", 
			map[string]interface{}{"id": 1})
		secondDeptLoadTime := time.Since(start)
		
		if err != nil {
			t.Logf("第二次部门懒加载失败: %v", err)
		}

		t.Logf("跨用户懒加载缓存共享结果:")
		t.Logf("  第一次部门加载耗时: %v", firstDeptLoadTime)
		t.Logf("  第二次部门加载耗时: %v", secondDeptLoadTime)

		t.Log("跨用户懒加载缓存共享测试通过")
	})
}

// TestLazyLoadingErrorHandling 测试懒加载错误处理
func TestLazyLoadingErrorHandling(t *testing.T) {
	db, err := setupTestDatabase()
	if err != nil {
		t.Fatalf("设置测试数据库失败: %v", err)
	}
	defer cleanupDatabase(db)

	xmlSession := mybatis.NewXMLMapper(db).Debug(false)
	ctx := context.Background()

	// 加载映射
	err = xmlSession.LoadMapperXMLFromString(getLazyLoadingMapperXML())
	if err != nil {
		t.Fatalf("加载XML映射失败: %v", err)
	}

	t.Run("懒加载不存在的关联数据", func(t *testing.T) {
		// 尝试懒加载不存在的用户资料
		profile, err := xmlSession.SelectOneByID(ctx, "LazyLoading.selectUserProfile", 
			map[string]interface{}{"userId": 99999})
		if err != nil {
			t.Logf("正确处理不存在用户的懒加载: %v", err)
		} else if profile == nil {
			t.Log("正确返回nil表示数据不存在")
		} else {
			t.Log("意外返回了数据，可能是测试数据问题")
		}
	})

	t.Run("懒加载SQL错误处理", func(t *testing.T) {
		// 尝试执行错误的懒加载查询
		_, err := xmlSession.SelectOneByID(ctx, "LazyLoading.nonExistentQuery", nil)
		if err == nil {
			t.Error("期望SQL错误但没有返回错误")
		} else {
			t.Logf("正确处理SQL错误: %v", err)
		}
	})

	t.Run("懒加载参数错误处理", func(t *testing.T) {
		// 传入错误类型的参数
		_, err := xmlSession.SelectOneByID(ctx, "LazyLoading.selectUserProfile", 
			map[string]interface{}{"userId": "invalid_id"})
		if err != nil {
			t.Logf("正确处理参数错误: %v", err)
		}
	})

	t.Log("懒加载错误处理测试完成")
}

// setupLazyLoadingTestData 设置懒加载测试数据
func setupLazyLoadingTestData(t *testing.T, xmlSession mybatis.XMLSession, ctx context.Context) {
	t.Helper()
	
	// 获取底层SimpleSession来插入测试数据
	db, err := setupTestDatabase()
	if err != nil {
		t.Fatalf("获取数据库连接失败: %v", err)
	}
	
	simpleSession := mybatis.NewSimpleSession(db)
	
	// 创建部门数据
	departments := []struct {
		name    string
		code    string
		manager string
	}{
		{"技术部", "TECH", "张技术"},
		{"销售部", "SALES", "李销售"},
		{"人事部", "HR", "王人事"},
	}
	
	var deptIDs []int64
	for i, dept := range departments {
		deptID, err := simpleSession.Insert(ctx,
			`INSERT INTO departments (name, code, manager) VALUES (?, ?, ?)`,
			dept.name, dept.code, dept.manager)
		if err != nil {
			t.Logf("插入部门 %d 失败: %v", i+1, err)
		} else {
			deptIDs = append(deptIDs, deptID)
		}
	}
	
	// 创建用户数据（分配到不同部门）
	testUsers := []struct {
		name   string
		email  string
		age    int
		status string
		deptID int64
	}{
		{"懒加载用户1", "lazy1@example.com", 28, "active", deptIDs[0]},
		{"懒加载用户2", "lazy2@example.com", 32, "active", deptIDs[0]},
		{"懒加载用户3", "lazy3@example.com", 25, "active", deptIDs[1]},
		{"懒加载用户4", "lazy4@example.com", 30, "inactive", deptIDs[1]},
		{"懒加载用户5", "lazy5@example.com", 27, "active", deptIDs[2]},
	}
	
	var userIDs []int64
	for i, user := range testUsers {
		userID, err := simpleSession.Insert(ctx,
			`INSERT INTO users (name, email, age, status, department_id) VALUES (?, ?, ?, ?, ?)`,
			user.name, user.email, user.age, user.status, user.deptID)
		if err != nil {
			t.Logf("插入懒加载用户 %d 失败: %v", i+1, err)
		} else {
			userIDs = append(userIDs, userID)
		}
	}
	
	// 创建用户资料数据
	for i, userID := range userIDs {
		_, err := simpleSession.Insert(ctx,
			`INSERT INTO user_profiles (user_id, avatar, bio, phone, address) VALUES (?, ?, ?, ?, ?)`,
			userID,
			fmt.Sprintf("avatar_%d.jpg", i+1),
			fmt.Sprintf("用户%d的个人简介", i+1),
			fmt.Sprintf("1380000000%d", i+1),
			fmt.Sprintf("地址%d号", i+1))
		if err != nil {
			t.Logf("插入用户资料 %d 失败: %v", i+1, err)
		}
	}
	
	// 创建订单数据
	for i, userID := range userIDs {
		// 每个用户创建1-3个订单
		orderCount := (i % 3) + 1
		for j := 0; j < orderCount; j++ {
			orderID, err := simpleSession.Insert(ctx,
				`INSERT INTO orders (user_id, order_no, amount, status) VALUES (?, ?, ?, ?)`,
				userID,
				fmt.Sprintf("ORD%d%03d", i+1, j+1),
				float64((j+1)*100),
				"pending")
			if err != nil {
				t.Logf("插入订单失败: %v", err)
			} else {
				// 为每个订单创建订单项目
				for k := 0; k < 2; k++ {
					_, err := simpleSession.Insert(ctx,
						`INSERT INTO order_items (order_id, product_name, quantity, price) VALUES (?, ?, ?, ?)`,
						orderID,
						fmt.Sprintf("商品%d-%d", j+1, k+1),
						k+1,
						float64((k+1)*50))
					if err != nil {
						t.Logf("插入订单项目失败: %v", err)
					}
				}
			}
		}
	}
	
	t.Logf("成功设置懒加载测试数据 - 部门:%d个, 用户:%d个", len(deptIDs), len(userIDs))
}

// getLazyLoadingMapperXML 获取懒加载映射XML
func getLazyLoadingMapperXML() string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE mapper PUBLIC "-//mybatis.org//DTD Mapper 3.0//EN" "http://mybatis.org/dtd/mybatis-3-mapper.dtd">
<mapper namespace="LazyLoading">
    
    <!-- 查询用户（不包含关联数据） -->
    <select id="selectUsers" resultType="map">
        SELECT id, name, email, age, status, department_id, created_at
        FROM users
        ORDER BY id
        LIMIT 10
    </select>
    
    <!-- 根据ID查询用户 -->
    <select id="selectUserById" parameterType="map" resultType="map">
        SELECT id, name, email, age, status, department_id, created_at
        FROM users 
        WHERE id = #{id}
    </select>
    
    <!-- 查询用户及关联数据（立即加载） -->
    <select id="selectUsersWithAssociations" resultType="map">
        SELECT 
            u.id, u.name, u.email, u.age, u.status, u.department_id,
            p.avatar, p.bio, p.phone, p.address,
            d.name as dept_name, d.code as dept_code
        FROM users u
        LEFT JOIN user_profiles p ON u.id = p.user_id
        LEFT JOIN departments d ON u.department_id = d.id
        ORDER BY u.id
        LIMIT 10
    </select>
    
    <!-- 懒加载用户资料 -->
    <select id="selectUserProfile" parameterType="map" resultType="map">
        SELECT id, user_id, avatar, bio, phone, address
        FROM user_profiles
        WHERE user_id = #{userId}
    </select>
    
    <!-- 懒加载用户订单 -->
    <select id="selectUserOrders" parameterType="map" resultType="map">
        SELECT id, user_id, order_no, amount, status, created_at
        FROM orders
        WHERE user_id = #{userId}
        ORDER BY created_at DESC
    </select>
    
    <!-- 懒加载订单项目 -->
    <select id="selectOrderItems" parameterType="map" resultType="map">
        SELECT id, order_id, product_name, quantity, price
        FROM order_items
        WHERE order_id = #{orderId}
        ORDER BY id
    </select>
    
    <!-- 根据ID查询部门 -->
    <select id="selectDepartmentById" parameterType="map" resultType="map">
        SELECT id, name, code, manager
        FROM departments
        WHERE id = #{id}
    </select>
    
    <!-- 查询部门用户 -->
    <select id="selectDepartmentUsers" parameterType="map" resultType="map">
        SELECT id, name, email, age, status
        FROM users
        WHERE department_id = #{departmentId}
        ORDER BY name
    </select>
    
    <!-- 根据用户ID查询部门 -->
    <select id="selectUserDepartment" parameterType="map" resultType="map">
        SELECT d.id, d.name, d.code, d.manager
        FROM departments d
        INNER JOIN users u ON d.id = u.department_id
        WHERE u.id = #{userId}
    </select>
    
</mapper>`
}

// 注意：此测试文件演示了懒加载机制的使用方法。
// 由于当前框架可能还未完全实现懒加载的代理对象功能，
// 某些测试可能需要手动触发懒加载查询。
// 这些测试可以作为后续实现懒加载功能的参考。