package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/zsy619/yyhertz/framework/orm"
)

// BenchmarkUser 基准测试用户模型
type BenchmarkUser struct {
	orm.BaseModel
	Name   string `gorm:"size:100" json:"name"`
	Email  string `gorm:"size:100;index" json:"email"`
	Age    int    `json:"age"`
	Status string `gorm:"size:20;index" json:"status"`
}

// BenchmarkProfile 基准测试资料模型
type BenchmarkProfile struct {
	orm.BaseModel
	UserID uint   `gorm:"index" json:"user_id"`
	Bio    string `gorm:"type:text" json:"bio"`
	Avatar string `gorm:"size:255" json:"avatar"`
}

// setupBenchmarkDB 设置基准测试数据库
func setupBenchmarkDB(b *testing.B) *orm.ORM {
	config := &orm.DatabaseConfig{
		Type:         "sqlite",
		Database:     ":memory:", // 使用内存数据库提高测试速度
		MaxIdleConns: 10,
		MaxOpenConns: 100,
		LogLevel:     "silent", // 关闭日志提高性能
	}

	ormInstance, err := orm.NewORM(config)
	if err != nil {
		b.Fatalf("创建ORM失败: %v", err)
	}

	// 自动迁移
	if err := ormInstance.AutoMigrate(&BenchmarkUser{}, &BenchmarkProfile{}); err != nil {
		b.Fatalf("自动迁移失败: %v", err)
	}

	return ormInstance
}

// generateTestUsers 生成测试用户数据
func generateTestUsers(count int) []*BenchmarkUser {
	users := make([]*BenchmarkUser, count)
	statuses := []string{"active", "inactive", "pending", "banned"}

	for i := 0; i < count; i++ {
		users[i] = &BenchmarkUser{
			Name:   fmt.Sprintf("User%d", i+1),
			Email:  fmt.Sprintf("user%d@example.com", i+1),
			Age:    rand.Intn(50) + 18, // 18-68岁
			Status: statuses[rand.Intn(len(statuses))],
		}
	}

	return users
}

// ==================== 基础CRUD基准测试 ====================

// BenchmarkCRUDCreate 创建操作基准测试
func BenchmarkCRUDCreate(b *testing.B) {
	ormInstance := setupBenchmarkDB(b)
	crud := orm.NewSimpleCRUD(ormInstance.DB())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		user := &BenchmarkUser{
			Name:   fmt.Sprintf("BenchUser%d", i),
			Email:  fmt.Sprintf("bench%d@example.com", i),
			Age:    25,
			Status: "active",
		}
		if err := crud.Create(user); err != nil {
			b.Fatalf("创建用户失败: %v", err)
		}
	}
}

// BenchmarkCRUDBatchCreate 批量创建基准测试
func BenchmarkCRUDBatchCreate(b *testing.B) {
	ormInstance := setupBenchmarkDB(b)
	crud := orm.NewSimpleCRUD(ormInstance.DB())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		users := generateTestUsers(100)
		if err := crud.BatchCreate(users, 50); err != nil {
			b.Fatalf("批量创建失败: %v", err)
		}
	}
}

// BenchmarkCRUDRead 读取操作基准测试
func BenchmarkCRUDRead(b *testing.B) {
	setupBenchmarkDB(b)
	crud := orm.GetSimpleCRUD()

	// 预先插入测试数据
	users := generateTestUsers(1000)
	crud.BatchCreate(users, 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var user BenchmarkUser
		if err := crud.First(&user, "id = ?", rand.Intn(1000)+1); err != nil {
			// 忽略未找到的错误
			if err.Error() != "record not found" {
				b.Fatalf("查询用户失败: %v", err)
			}
		}
	}
}

// BenchmarkCRUDUpdate 更新操作基准测试
func BenchmarkCRUDUpdate(b *testing.B) {
	setupBenchmarkDB(b)
	crud := orm.GetSimpleCRUD()

	// 预先插入测试数据
	users := generateTestUsers(1000)
	crud.BatchCreate(users, 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		user := &BenchmarkUser{}
		user.ID = uint(rand.Intn(1000) + 1)
		if err := crud.Update(user, "age", rand.Intn(50)+18); err != nil {
			b.Fatalf("更新用户失败: %v", err)
		}
	}
}

// BenchmarkCRUDDelete 删除操作基准测试
func BenchmarkCRUDDelete(b *testing.B) {
	setupBenchmarkDB(b)
	crud := orm.GetSimpleCRUD()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 每次测试前创建用户
		user := &BenchmarkUser{
			Name:   fmt.Sprintf("DeleteUser%d", i),
			Email:  fmt.Sprintf("delete%d@example.com", i),
			Age:    25,
			Status: "active",
		}
		crud.Create(user)

		// 删除用户
		if err := crud.Delete(user); err != nil {
			b.Fatalf("删除用户失败: %v", err)
		}
	}
}

// ==================== 查询基准测试 ====================

// BenchmarkQueryWhere 条件查询基准测试
func BenchmarkQueryWhere(b *testing.B) {
	setupBenchmarkDB(b)
	crud := orm.GetSimpleCRUD()

	// 预先插入测试数据
	users := generateTestUsers(10000)
	crud.BatchCreate(users, 500)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var users []BenchmarkUser
		status := []string{"active", "inactive", "pending"}[i%3]
		if err := crud.Find(&users, "status = ?", status); err != nil {
			b.Fatalf("条件查询失败: %v", err)
		}
	}
}

// BenchmarkQueryChain 链式查询基准测试
func BenchmarkQueryChain(b *testing.B) {
	setupBenchmarkDB(b)
	crud := orm.GetSimpleCRUD()

	// 预先插入测试数据
	users := generateTestUsers(10000)
	crud.BatchCreate(users, 500)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var users []BenchmarkUser
		err := crud.Where("age > ?", 25).
			WhereIn("status", []string{"active", "pending"}).
			OrderBy("created_at", "desc").
			Limit(100).
			Find(&users)
		if err != nil {
			b.Fatalf("链式查询失败: %v", err)
		}
	}
}

// BenchmarkQueryPagination 分页查询基准测试
func BenchmarkQueryPagination(b *testing.B) {
	setupBenchmarkDB(b)
	crud := orm.GetSimpleCRUD()

	// 预先插入测试数据
	users := generateTestUsers(10000)
	crud.BatchCreate(users, 500)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var users []BenchmarkUser
		page := (i % 100) + 1 // 1-100页
		_, err := crud.Paginate(&BenchmarkUser{}, page, 50, &users)
		if err != nil {
			b.Fatalf("分页查询失败: %v", err)
		}
	}
}

// ==================== 仓库模式基准测试 ====================

// BenchmarkRepositoryFind 仓库查询基准测试
func BenchmarkRepositoryFind(b *testing.B) {
	setupBenchmarkDB(b)
	repo := orm.GetRepository[BenchmarkUser]()

	// 预先插入测试数据
	users := generateTestUsers(5000)
	for _, user := range users {
		repo.Create(user)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		users, err := repo.FindWhere("age > ?", rand.Intn(30)+20)
		if err != nil {
			b.Fatalf("仓库查询失败: %v", err)
		}
		_ = users
	}
}

// BenchmarkRepositoryPaginate 仓库分页基准测试
func BenchmarkRepositoryPaginate(b *testing.B) {
	setupBenchmarkDB(b)
	repo := orm.GetRepository[BenchmarkUser]()

	// 预先插入测试数据
	users := generateTestUsers(10000)
	for i := 0; i < len(users); i += 100 {
		end := i + 100
		if end > len(users) {
			end = len(users)
		}
		for j := i; j < end; j++ {
			repo.Create(users[j])
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		page := (i % 200) + 1
		users, total, err := repo.Paginate(page, 50)
		if err != nil {
			b.Fatalf("仓库分页失败: %v", err)
		}
		_, _ = users, total
	}
}

// ==================== 事务基准测试 ====================

// BenchmarkTransaction 事务操作基准测试
func BenchmarkTransaction(b *testing.B) {
	setupBenchmarkDB(b)
	_ = orm.GetSimpleCRUD() // 避免未使用警告

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := orm.QuickTransaction(func(txCrud *orm.SimpleCRUD) error {
			// 在事务中创建用户和资料
			user := &BenchmarkUser{
				Name:   fmt.Sprintf("TxUser%d", i),
				Email:  fmt.Sprintf("tx%d@example.com", i),
				Age:    25,
				Status: "active",
			}
			if err := txCrud.Create(user); err != nil {
				return err
			}

			profile := &BenchmarkProfile{
				UserID: user.ID,
				Bio:    fmt.Sprintf("Bio for user %d", i),
				Avatar: fmt.Sprintf("avatar%d.jpg", i),
			}
			return txCrud.Create(profile)
		})
		if err != nil {
			b.Fatalf("事务操作失败: %v", err)
		}
	}
}

// ==================== 连接池基准测试 ====================

// BenchmarkConnectionPool 连接池基准测试
func BenchmarkConnectionPool(b *testing.B) {
	config := &orm.DatabaseConfig{
		Type:         "sqlite",
		Database:     ":memory:",
		MaxIdleConns: 20,
		MaxOpenConns: 50,
		LogLevel:     "silent",
	}

	ormInstance, err := orm.NewORM(config)
	if err != nil {
		b.Fatalf("创建ORM失败: %v", err)
	}

	if err := ormInstance.AutoMigrate(&BenchmarkUser{}); err != nil {
		b.Fatalf("自动迁移失败: %v", err)
	}

	crud := orm.GetSimpleCRUD()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			user := &BenchmarkUser{
				Name:   fmt.Sprintf("ParallelUser%d", i),
				Email:  fmt.Sprintf("parallel%d@example.com", i),
				Age:    rand.Intn(50) + 18,
				Status: "active",
			}
			if err := crud.Create(user); err != nil {
				b.Fatalf("并发创建用户失败: %v", err)
			}
			i++
		}
	})
}

// ==================== 缓存基准测试 ====================

// BenchmarkCachedQuery 缓存查询基准测试
func BenchmarkCachedQuery(b *testing.B) {
	setupBenchmarkDB(b)
	crud := orm.GetSimpleCRUD()

	// 预先插入测试数据
	users := generateTestUsers(1000)
	crud.BatchCreate(users, 100)

	db := orm.GetDefaultORM().DB()
	cachedQuery := orm.NewCachedQuery(db).
		WithCacheKey("benchmark_active_users").
		WithTTL(1 * time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var users []BenchmarkUser
		if err := cachedQuery.Find(&users, "status = ?", "active"); err != nil {
			b.Fatalf("缓存查询失败: %v", err)
		}
	}
}

// ==================== 性能对比测试 ====================

// BenchmarkCompareCreateMethods 创建方法性能对比
func BenchmarkCompareCreateMethods(b *testing.B) {
	setupBenchmarkDB(b)

	b.Run("SimpleCRUD/Create", func(b *testing.B) {
		crud := orm.GetSimpleCRUD()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			user := &BenchmarkUser{
				Name:   fmt.Sprintf("SimpleCRUD%d", i),
				Email:  fmt.Sprintf("simple%d@example.com", i),
				Age:    25,
				Status: "active",
			}
			if err := crud.Create(user); err != nil {
				b.Fatalf("SimpleCRUD创建失败: %v", err)
			}
		}
	})

	b.Run("Repository/Create", func(b *testing.B) {
		repo := orm.GetRepository[BenchmarkUser]()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			user := &BenchmarkUser{
				Name:   fmt.Sprintf("Repository%d", i),
				Email:  fmt.Sprintf("repo%d@example.com", i),
				Age:    25,
				Status: "active",
			}
			if err := repo.Create(user); err != nil {
				b.Fatalf("Repository创建失败: %v", err)
			}
		}
	})

	b.Run("QuickCreate", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			user := &BenchmarkUser{
				Name:   fmt.Sprintf("QuickCreate%d", i),
				Email:  fmt.Sprintf("quick%d@example.com", i),
				Age:    25,
				Status: "active",
			}
			if err := orm.QuickCreate(user); err != nil {
				b.Fatalf("QuickCreate创建失败: %v", err)
			}
		}
	})
}

// BenchmarkCompareQueryMethods 查询方法性能对比
func BenchmarkCompareQueryMethods(b *testing.B) {
	setupBenchmarkDB(b)
	crud := orm.GetSimpleCRUD()

	// 预先插入测试数据
	users := generateTestUsers(5000)
	crud.BatchCreate(users, 250)

	b.Run("SimpleCRUD/Find", func(b *testing.B) {
		crud := orm.GetSimpleCRUD()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var users []BenchmarkUser
			if err := crud.Find(&users, "status = ?", "active"); err != nil {
				b.Fatalf("SimpleCRUD查询失败: %v", err)
			}
		}
	})

	b.Run("Repository/FindWhere", func(b *testing.B) {
		repo := orm.GetRepository[BenchmarkUser]()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, err := repo.FindWhere("status = ?", "active"); err != nil {
				b.Fatalf("Repository查询失败: %v", err)
			}
		}
	})

	b.Run("QueryBuilder/Find", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			query := orm.GetQueryBuilder[BenchmarkUser]()
			if _, err := query.Where("status = ?", "active").Find(); err != nil {
				b.Fatalf("QueryBuilder查询失败: %v", err)
			}
		}
	})
}

// ==================== 压力测试辅助函数 ====================

// runStressTest 运行压力测试
func runStressTest(b *testing.B, name string, concurrency int, testFunc func()) {
	b.Run(fmt.Sprintf("%s/Concurrency-%d", name, concurrency), func(b *testing.B) {
		b.SetParallelism(concurrency)
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				testFunc()
			}
		})
	})
}

// BenchmarkStressTest 压力测试
func BenchmarkStressTest(b *testing.B) {
	setupBenchmarkDB(b)
	crud := orm.GetSimpleCRUD()

	// 预先插入测试数据
	users := generateTestUsers(1000)
	crud.BatchCreate(users, 100)

	// 创建压力测试
	createTest := func() {
		user := &BenchmarkUser{
			Name:   "StressUser" + strconv.Itoa(rand.Int()),
			Email:  "stress" + strconv.Itoa(rand.Int()) + "@example.com",
			Age:    rand.Intn(50) + 18,
			Status: "active",
		}
		crud.Create(user)
	}

	// 查询压力测试
	queryTest := func() {
		var users []BenchmarkUser
		crud.Find(&users, "status = ?", "active")
	}

	// 更新压力测试
	updateTest := func() {
		user := &BenchmarkUser{}
		user.ID = uint(rand.Intn(1000) + 1)
		crud.Update(user, "age", rand.Intn(50)+18)
	}

	// 运行不同并发级别的压力测试
	concurrencies := []int{1, 5, 10, 20, 50}

	for _, concurrency := range concurrencies {
		runStressTest(b, "Create", concurrency, createTest)
		runStressTest(b, "Query", concurrency, queryTest)
		runStressTest(b, "Update", concurrency, updateTest)
	}
}

// ==================== 内存使用基准测试 ====================

// BenchmarkMemoryUsage 内存使用基准测试
func BenchmarkMemoryUsage(b *testing.B) {
	setupBenchmarkDB(b)
	crud := orm.GetSimpleCRUD()

	b.ReportAllocs() // 报告内存分配

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 创建大量用户测试内存使用
		users := generateTestUsers(100)
		for _, user := range users {
			if err := crud.Create(user); err != nil {
				b.Fatalf("创建用户失败: %v", err)
			}
		}

		// 查询用户测试内存使用
		var queryUsers []BenchmarkUser
		if err := crud.Find(&queryUsers); err != nil {
			b.Fatalf("查询用户失败: %v", err)
		}
	}
}

// init 初始化随机数种子
func init() {
	rand.Seed(time.Now().UnixNano())
}