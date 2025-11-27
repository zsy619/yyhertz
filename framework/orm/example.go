package orm

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/zsy619/yyhertz/framework/config"
)

// ============= 示例模型定义 =============

// User 用户模型示例
type User struct {
	BaseModel
	Username string     `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Email    string     `gorm:"uniqueIndex;size:100;not null" json:"email"`
	Password string     `gorm:"size:255;not null" json:"-"`
	Nickname string     `gorm:"size:100" json:"nickname"`
	Avatar   string     `gorm:"size:255" json:"avatar"`
	Status   int        `gorm:"default:1;comment:用户状态 1-正常 0-禁用" json:"status"`
	LastIP   string     `gorm:"size:45" json:"last_ip"`
	LoginAt  *time.Time `json:"login_at"`

	// 关联关系
	Profile *UserProfile `gorm:"foreignKey:UserID" json:"profile,omitempty"`
	Posts   []*Post      `gorm:"foreignKey:UserID" json:"posts,omitempty"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// UserProfile 用户资料模型
type UserProfile struct {
	BaseModel
	UserID   uint   `gorm:"uniqueIndex;not null" json:"user_id"`
	RealName string `gorm:"size:50" json:"real_name"`
	Phone    string `gorm:"size:20" json:"phone"`
	Address  string `gorm:"size:255" json:"address"`
	Bio      string `gorm:"type:text" json:"bio"`

	// 关联关系
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 指定表名
func (UserProfile) TableName() string {
	return "user_profiles"
}

// Post 文章模型示例
type Post struct {
	BaseModel
	UserID      uint       `gorm:"index;not null" json:"user_id"`
	Title       string     `gorm:"size:200;not null" json:"title"`
	Content     string     `gorm:"type:longtext" json:"content"`
	Summary     string     `gorm:"size:500" json:"summary"`
	Status      int        `gorm:"default:1;comment:文章状态 1-发布 0-草稿" json:"status"`
	ViewCount   int64      `gorm:"default:0" json:"view_count"`
	LikeCount   int64      `gorm:"default:0" json:"like_count"`
	PublishedAt *time.Time `json:"published_at"`

	// 关联关系
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 指定表名
func (Post) TableName() string {
	return "posts"
}

// ============= 用户仓库示例 =============

// UserRepository 用户仓库
type UserRepository struct {
	*BaseRepository[User]
}

// NewUserRepository 创建用户仓库
func NewUserRepository() *UserRepository {
	return &UserRepository{
		BaseRepository: GetRepository[User](),
	}
}

// FindByUsername 根据用户名查找用户
func (ur *UserRepository) FindByUsername(username string) (*User, error) {
	var user User
	err := ur.GetDB().Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByEmail 根据邮箱查找用户
func (ur *UserRepository) FindByEmail(email string) (*User, error) {
	var user User
	err := ur.GetDB().Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateWithProfile 创建用户并创建对应的用户资料
func (ur *UserRepository) CreateWithProfile(user *User, profile *UserProfile) error {
	return ur.WithTransaction(func(tx *gorm.DB) error {
		// 创建用户
		if err := tx.Create(user).Error; err != nil {
			return err
		}

		// 设置资料的用户ID
		profile.UserID = user.ID

		// 创建用户资料
		if err := tx.Create(profile).Error; err != nil {
			return err
		}

		return nil
	})
}

// GetActiveUsers 获取活跃用户列表
func (ur *UserRepository) GetActiveUsers(page, pageSize int) ([]*User, int64, error) {
	return ur.PaginateWhere("status = ?", page, pageSize, 1)
}

// UpdateLastLogin 更新最后登录信息
func (ur *UserRepository) UpdateLastLogin(userID uint, ip string) error {
	return ur.UpdateColumns(userID, map[string]any{
		"last_ip":  ip,
		"login_at": time.Now(),
	})
}

// ============= 文章仓库示例 =============

// PostRepository 文章仓库
type PostRepository struct {
	*BaseRepository[Post]
}

// NewPostRepository 创建文章仓库
func NewPostRepository() *PostRepository {
	return &PostRepository{
		BaseRepository: GetRepository[Post](),
	}
}

// GetPublishedPosts 获取已发布的文章列表
func (pr *PostRepository) GetPublishedPosts(page, pageSize int) ([]*Post, int64, error) {
	return pr.PaginateWhere("status = ?", page, pageSize, 1)
}

// GetUserPosts 获取用户的文章列表
func (pr *PostRepository) GetUserPosts(userID uint, page, pageSize int) ([]*Post, int64, error) {
	return pr.PaginateWhere("user_id = ? AND status = ?", page, pageSize, userID, 1)
}

// IncrementViewCount 增加阅读量
func (pr *PostRepository) IncrementViewCount(postID uint) error {
	return pr.GetDB().Model(&Post{}).Where("id = ?", postID).
		Update("view_count", gorm.Expr("view_count + ?", 1)).Error
}

// SearchPosts 搜索文章
func (pr *PostRepository) SearchPosts(keyword string, page, pageSize int) ([]*Post, int64, error) {
	condition := "status = ? AND (title LIKE ? OR content LIKE ?)"
	searchTerm := "%" + keyword + "%"
	return pr.PaginateWhere(condition, page, pageSize, 1, searchTerm, searchTerm)
}

// ============= 服务层示例 =============

// UserService 用户服务
type UserService struct {
	userRepo *UserRepository
}

// NewUserService 创建用户服务
func NewUserService() *UserService {
	return &UserService{
		userRepo: NewUserRepository(),
	}
}

// Register 用户注册
func (us *UserService) Register(username, email, password, nickname string) (*User, error) {
	// 检查用户名是否存在
	if exists, _ := us.userRepo.ExistsWhere("username = ?", username); exists {
		return nil, fmt.Errorf("username already exists")
	}

	// 检查邮箱是否存在
	if exists, _ := us.userRepo.ExistsWhere("email = ?", email); exists {
		return nil, fmt.Errorf("email already exists")
	}

	// 创建用户
	user := &User{
		Username: username,
		Email:    email,
		Password: password, // 实际应用中需要加密
		Nickname: nickname,
		Status:   1,
	}

	// 创建用户资料
	profile := &UserProfile{
		RealName: nickname,
	}

	// 在事务中创建用户和资料
	if err := us.userRepo.CreateWithProfile(user, profile); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	config.Infof("User registered successfully: %s (ID: %d)", username, user.ID)
	return user, nil
}

// Login 用户登录
func (us *UserService) Login(username, password, ip string) (*User, error) {
	// 查找用户
	user, err := us.userRepo.FindByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// 验证密码（实际应用中需要验证加密密码）
	if user.Password != password {
		return nil, fmt.Errorf("invalid password")
	}

	// 检查用户状态
	if user.Status != 1 {
		return nil, fmt.Errorf("user account disabled")
	}

	// 更新最后登录信息
	if err := us.userRepo.UpdateLastLogin(user.ID, ip); err != nil {
		config.Warnf("Failed to update last login for user %d: %v", user.ID, err)
	}

	config.Infof("User logged in successfully: %s (ID: %d)", username, user.ID)
	return user, nil
}

// GetUserWithProfile 获取用户及其资料
func (us *UserService) GetUserWithProfile(userID uint) (*User, error) {
	var user User
	err := us.userRepo.GetDB().Preload("Profile").First(&user, userID).Error
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return &user, nil
}

// ============= 完整示例 =============

// RunExample 运行ORM示例
func RunExample() error {
	config.Info("Starting ORM example...")

	// 1. 获取ORM实例
	orm := GetDefaultORM()

	// 2. 自动迁移模型
	config.Info("Running auto migration...")
	if err := orm.AutoMigrate(&User{}, &UserProfile{}, &Post{}); err != nil {
		return fmt.Errorf("auto migration failed: %w", err)
	}

	// 3. 创建服务实例
	userService := NewUserService()
	postRepo := NewPostRepository()

	// 4. 用户注册示例
	config.Info("Creating sample users...")
	user1, err := userService.Register("john_doe", "john@example.com", "password123", "John Doe")
	if err != nil {
		config.Warnf("Failed to create user1 (might already exist): %v", err)
		// 尝试查找现有用户
		if user1, err = userService.userRepo.FindByUsername("john_doe"); err != nil {
			return fmt.Errorf("failed to find existing user: %w", err)
		}
	}

	user2, err := userService.Register("jane_smith", "jane@example.com", "password456", "Jane Smith")
	if err != nil {
		config.Warnf("Failed to create user2 (might already exist): %v", err)
		if user2, err = userService.userRepo.FindByUsername("jane_smith"); err != nil {
			return fmt.Errorf("failed to find existing user: %w", err)
		}
	}

	// 5. 创建文章示例
	config.Info("Creating sample posts...")
	posts := []*Post{
		{
			UserID:      user1.ID,
			Title:       "Go语言最佳实践",
			Content:     "本文介绍Go语言开发中的最佳实践...",
			Summary:     "Go语言开发最佳实践指南",
			Status:      1,
			PublishedAt: &time.Time{},
		},
		{
			UserID:      user2.ID,
			Title:       "微服务架构设计",
			Content:     "微服务架构是现代应用开发的重要模式...",
			Summary:     "微服务架构设计原则与实践",
			Status:      1,
			PublishedAt: &time.Time{},
		},
	}

	for _, post := range posts {
		now := time.Now()
		post.PublishedAt = &now
		if err := postRepo.Create(post); err != nil {
			config.Warnf("Failed to create post (might already exist): %v", err)
		}
	}

	// 6. 查询示例
	config.Info("Running query examples...")

	// 分页查询用户
	users, total, err := userService.userRepo.GetActiveUsers(1, 10)
	if err != nil {
		return fmt.Errorf("failed to get active users: %w", err)
	}
	config.Infof("Found %d active users (total: %d)", len(users), total)

	// 查询用户详情
	userWithProfile, err := userService.GetUserWithProfile(user1.ID)
	if err != nil {
		return fmt.Errorf("failed to get user with profile: %w", err)
	}
	config.Infof("User with profile: %s (%s)", userWithProfile.Username, userWithProfile.Profile.RealName)

	// 查询文章
	publishedPosts, totalPosts, err := postRepo.GetPublishedPosts(1, 10)
	if err != nil {
		return fmt.Errorf("failed to get published posts: %w", err)
	}
	config.Infof("Found %d published posts (total: %d)", len(publishedPosts), totalPosts)

	// 7. 事务示例
	config.Info("Running transaction example...")
	err = WithTransaction(func(tx *gorm.DB) error {
		// 在事务中增加文章阅读量和点赞数
		for _, post := range publishedPosts {
			if err := tx.Model(post).Updates(map[string]any{
				"view_count": gorm.Expr("view_count + ?", 10),
				"like_count": gorm.Expr("like_count + ?", 5),
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("transaction example failed: %w", err)
	}

	// 8. 查询构建器示例
	config.Info("Running query builder example...")
	queryBuilder := GetQueryBuilder[Post]()
	recentPosts, err := queryBuilder.
		Where("status = ?", 1).
		Where("created_at > ?", time.Now().AddDate(0, 0, -30)).
		Order("created_at DESC").
		Limit(5).
		Preload("User").
		Find()

	if err != nil {
		return fmt.Errorf("query builder example failed: %w", err)
	}
	config.Infof("Found %d recent posts", len(recentPosts))

	// 9. 批量操作示例
	config.Info("Running batch operation example...")
	batchManager := NewBatchTransactionManager(orm.DB())

	err = batchManager.
		AddOperationFunc("Update user status", func(tx *gorm.DB) error {
			return tx.Model(&User{}).Where("status = ?", 1).Update("status", 1).Error
		}).
		AddOperationFunc("Update post view counts", func(tx *gorm.DB) error {
			return tx.Model(&Post{}).Where("status = ?", 1).
				Update("view_count", gorm.Expr("view_count + ?", 1)).Error
		}).
		Execute()

	if err != nil {
		return fmt.Errorf("batch operation example failed: %w", err)
	}

	// 10. 获取数据库统计信息
	stats := orm.GetStats()
	config.Infof("Database stats: %+v", stats)

	config.Info("ORM example completed successfully!")
	return nil
}

// RunExampleWithContext 带上下文的示例
func RunExampleWithContext(ctx context.Context) error {
	config.Info("Starting ORM example with context...")

	// 使用上下文的数据库操作
	db := GetDBFromContext(ctx)

	var users []*User
	if err := db.WithContext(ctx).Limit(5).Find(&users).Error; err != nil {
		return fmt.Errorf("context query failed: %w", err)
	}

	config.Infof("Found %d users using context", len(users))
	return nil
}

// RunMigrationExample 迁移示例
func RunMigrationExample() error {
	config.Info("Starting migration example...")

	// 创建迁移管理器
	migrationManager := GetDefaultMigrationManager()

	// 添加模型迁移
	migrationManager.
		AddModel("001_create_users_table", "Create users and user_profiles tables", &User{}, &UserProfile{}).
		AddModel("002_create_posts_table", "Create posts table", &Post{}).
		AddSQL("003_create_indexes", "Create performance indexes",
			`CREATE INDEX IF NOT EXISTS idx_posts_user_status ON posts(user_id, status);
			 CREATE INDEX IF NOT EXISTS idx_users_status_created ON users(status, created_at);`,
			`DROP INDEX IF EXISTS idx_posts_user_status;
			 DROP INDEX IF EXISTS idx_users_status_created;`)

	// 执行迁移
	if err := migrationManager.Migrate(); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	// 获取迁移状态
	status, err := migrationManager.Status()
	if err != nil {
		return fmt.Errorf("failed to get migration status: %w", err)
	}

	config.Infof("Migration status: %d total, %d executed, %d pending",
		status.TotalMigrations, status.ExecutedMigrations, status.PendingMigrations)

	return nil
}
