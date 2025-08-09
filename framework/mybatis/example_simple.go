// Package mybatis 使用示例
//
// 展示简化版MyBatis的使用方法，体现Go语言的简洁性
package mybatis

import (
	"context"
	"log"
	"time"

	"gorm.io/gorm"
)

// UserService 用户服务示例
type UserService struct {
	session SimpleSession
	txSession *TransactionAwareSession
}

// User 用户模型
type User struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	CreateAt time.Time `json:"create_at"`
}

// UserQuery 用户查询参数
type UserQuery struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	PageNum  int    `json:"pageNum"`
	PageSize int    `json:"pageSize"`
}

// NewUserService 创建用户服务
func NewUserService(db *gorm.DB) *UserService {
	// 创建简单会话，配置钩子
	session := NewSimpleSession(db).
		Debug(true).  // 启用调试模式
		AddBeforeHook(AuditHook()).  // 添加审计钩子
		AddBeforeHook(SecurityHook())  // 添加安全检查钩子
	
	// 添加性能监控钩子
	beforeHook, afterHook := PerformanceHook(100 * time.Millisecond)
	session = session.AddBeforeHook(beforeHook).AddAfterHook(afterHook)
	
	// 添加调试钩子
	debugBefore, debugAfter := DebugHook()
	session = session.AddBeforeHook(debugBefore).AddAfterHook(debugAfter)
	
	return &UserService{
		session:   session,
		txSession: NewTransactionAwareSession(db),
	}
}

// FindUserByID 根据ID查找用户
func (s *UserService) FindUserByID(ctx context.Context, id int64) (*User, error) {
	sql := "SELECT id, name, email, create_at FROM users WHERE id = ?"
	
	result, err := s.session.SelectOne(ctx, sql, id)
	if err != nil {
		return nil, err
	}
	
	if result == nil {
		return nil, nil
	}
	
	// 简单的结果映射
	if userMap, ok := result.(map[string]interface{}); ok {
		return mapToUser(userMap), nil
	}
	
	return nil, nil
}

// FindUsersByCondition 根据条件查找用户（分页）
func (s *UserService) FindUsersByCondition(ctx context.Context, query UserQuery) (*PageResult, error) {
	sql := "SELECT id, name, email, create_at FROM users WHERE 1=1"
	var args []interface{}
	
	// 动态构建WHERE条件
	if query.Name != "" {
		sql += " AND name LIKE ?"
		args = append(args, "%"+query.Name+"%")
	}
	
	if query.Email != "" {
		sql += " AND email = ?"
		args = append(args, query.Email)
	}
	
	sql += " ORDER BY id DESC"
	
	// 分页查询
	pageReq := PageRequest{
		Page: query.PageNum,
		Size: query.PageSize,
	}
	
	return s.session.SelectPage(ctx, sql, pageReq, args...)
}

// CreateUser 创建用户
func (s *UserService) CreateUser(ctx context.Context, user *User) error {
	sql := "INSERT INTO users (name, email, create_at) VALUES (?, ?, ?)"
	
	_, err := s.session.Insert(ctx, sql, user.Name, user.Email, time.Now())
	return err
}

// UpdateUser 更新用户
func (s *UserService) UpdateUser(ctx context.Context, user *User) error {
	sql := "UPDATE users SET name = ?, email = ? WHERE id = ?"
	
	_, err := s.session.Update(ctx, sql, user.Name, user.Email, user.ID)
	return err
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(ctx context.Context, id int64) error {
	sql := "DELETE FROM users WHERE id = ?"
	
	_, err := s.session.Delete(ctx, sql, id)
	return err
}

// CreateUserWithPosts 创建用户及其文章（事务示例）
func (s *UserService) CreateUserWithPosts(ctx context.Context, user *User, posts []Post) error {
	userID := getContextValue(ctx, UserIDKey, "system").(string)
	
	return s.txSession.ExecuteInTransaction(ctx, userID, func(txCtx context.Context, session SimpleSession) error {
		// 1. 创建用户
		userSQL := "INSERT INTO users (name, email, create_at) VALUES (?, ?, ?)"
		result, err := session.Insert(txCtx, userSQL, user.Name, user.Email, time.Now())
		if err != nil {
			return err
		}
		
		log.Printf("Created user, affected rows: %d", result)
		
		// 2. 创建文章
		for _, post := range posts {
			postSQL := "INSERT INTO posts (user_id, title, content, create_at) VALUES (?, ?, ?, ?)"
			_, err = session.Insert(txCtx, postSQL, user.ID, post.Title, post.Content, time.Now())
			if err != nil {
				return err
			}
		}
		
		return nil
	})
}

// DryRunExample DryRun模式示例
func (s *UserService) DryRunExample(ctx context.Context) error {
	// 启用DryRun模式
	dryRunSession := NewSimpleSession(nil).DryRun(true)
	
	// 这些操作只会打印SQL，不会实际执行
	_, err := dryRunSession.SelectOne(ctx, "SELECT * FROM users WHERE id = ?", 1)
	if err != nil {
		return err
	}
	
	_, err = dryRunSession.Insert(ctx, "INSERT INTO users (name, email) VALUES (?, ?)", "test", "test@example.com")
	if err != nil {
		return err
	}
	
	// 分页查询也支持DryRun
	_, err = dryRunSession.SelectPage(ctx, "SELECT * FROM users", PageRequest{Page: 1, Size: 10})
	return err
}

// Post 文章模型
type Post struct {
	ID       int64  `json:"id"`
	UserID   int64  `json:"user_id"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	CreateAt time.Time `json:"create_at"`
}

// mapToUser 将map转换为User对象
func mapToUser(m map[string]interface{}) *User {
	user := &User{}
	
	if id, ok := m["id"].(int64); ok {
		user.ID = id
	}
	
	if name, ok := m["name"].(string); ok {
		user.Name = name
	}
	
	if email, ok := m["email"].(string); ok {
		user.Email = email
	}
	
	if createAt, ok := m["create_at"].(time.Time); ok {
		user.CreateAt = createAt
	}
	
	return user
}

// 使用示例函数

// ExampleBasicUsage 基本使用示例
func ExampleBasicUsage(db *gorm.DB) {
	// 创建用户服务
	userService := NewUserService(db)
	
	// 创建带用户信息的context
	ctx := context.WithValue(context.Background(), UserIDKey, "admin")
	ctx = context.WithValue(ctx, RequestIDKey, "req_123")
	
	// 查找用户
	user, err := userService.FindUserByID(ctx, 1)
	if err != nil {
		log.Printf("Error finding user: %v", err)
		return
	}
	
	if user != nil {
		log.Printf("Found user: %+v", user)
	}
	
	// 分页查询
	query := UserQuery{
		Name:     "john",
		PageNum:  1,
		PageSize: 10,
	}
	
	pageResult, err := userService.FindUsersByCondition(ctx, query)
	if err != nil {
		log.Printf("Error finding users: %v", err)
		return
	}
	
	log.Printf("Found %d users, total: %d", len(pageResult.Items), pageResult.Total)
}

// ExampleTransactionUsage 事务使用示例
func ExampleTransactionUsage(db *gorm.DB) {
	userService := NewUserService(db)
	
	ctx := context.WithValue(context.Background(), UserIDKey, "admin")
	
	user := &User{
		Name:  "John Doe",
		Email: "john@example.com",
	}
	
	posts := []Post{
		{Title: "First Post", Content: "Hello World"},
		{Title: "Second Post", Content: "MyBatis is awesome"},
	}
	
	// 在事务中创建用户和文章
	err := userService.CreateUserWithPosts(ctx, user, posts)
	if err != nil {
		log.Printf("Transaction failed: %v", err)
	} else {
		log.Println("Transaction completed successfully")
	}
}

// ExampleDryRunUsage DryRun使用示例
func ExampleDryRunUsage() {
	userService := NewUserService(nil) // db为nil，因为DryRun不会实际执行
	
	ctx := context.WithValue(context.Background(), UserIDKey, "tester")
	
	// DryRun模式：只打印SQL，不执行
	err := userService.DryRunExample(ctx)
	if err != nil {
		log.Printf("DryRun error: %v", err)
	}
	
	log.Println("DryRun completed - check logs for SQL output")
}

// ExampleHookUsage 钩子使用示例
func ExampleHookUsage(db *gorm.DB) {
	// 创建指标收集器
	metrics := NewSimpleMetricsCollector()
	
	// 创建缓存
	cache := NewSimpleCache()
	
	// 创建会话并添加各种钩子
	session := NewSimpleSession(db).
		Debug(true).
		AddBeforeHook(AuditHook()).
		AddBeforeHook(SecurityHook())
	
	// 添加指标收集钩子
	metricsBefore, metricsAfter := MetricsHook(metrics)
	session = session.AddBeforeHook(metricsBefore).AddAfterHook(metricsAfter)
	
	// 添加缓存钩子
	cacheBefore, cacheAfter := CacheHook(cache)
	session = session.AddBeforeHook(cacheBefore).AddAfterHook(cacheAfter)
	
	ctx := context.WithValue(context.Background(), UserIDKey, "hook_user")
	
	// 执行一些查询
	_, _ = session.SelectOne(ctx, "SELECT COUNT(*) FROM users")
	_, _ = session.SelectList(ctx, "SELECT * FROM users LIMIT 5")
	
	// 查看指标
	stats := metrics.GetStats()
	log.Printf("Metrics: %+v", stats)
}