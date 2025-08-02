// Package mybatis_tests 提供MyBatis测试用例的数据模型
//
// 定义测试中使用的实体类和查询参数
package mybat

import (
	"time"
)

// User 用户实体
type User struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement" db:"id"`
	Name      string    `json:"name" gorm:"size:100;not null" db:"name" validate:"required,min=2,max=50"`
	Email     string    `json:"email" gorm:"size:100;uniqueIndex;not null" db:"email" validate:"required,email"`
	Age       int       `json:"age" gorm:"not null" db:"age" validate:"min=0,max=120"`
	Status    string    `json:"status" gorm:"size:20;default:active" db:"status" validate:"oneof=active inactive banned"`
	Avatar    string    `json:"avatar" gorm:"size:255" db:"avatar"`
	Phone     string    `json:"phone" gorm:"size:20" db:"phone"`
	Birthday  *time.Time `json:"birthday" db:"birthday"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" gorm:"index" db:"deleted_at"`
}

// UserQuery 用户查询参数
type UserQuery struct {
	// 基本查询参数
	Name     string `json:"name" form:"name"`
	Email    string `json:"email" form:"email"`
	Status   string `json:"status" form:"status"`
	Phone    string `json:"phone" form:"phone"`
	
	// 年龄范围查询
	AgeMin   int    `json:"age_min" form:"age_min"`
	AgeMax   int    `json:"age_max" form:"age_max"`
	
	// 时间范围查询
	CreatedAfter  *time.Time `json:"created_after" form:"created_after"`
	CreatedBefore *time.Time `json:"created_before" form:"created_before"`
	
	// 分页参数
	Page     int `json:"page" form:"page" validate:"min=1"`
	PageSize int `json:"page_size" form:"page_size" validate:"min=1,max=100"`
	Offset   int `json:"offset" form:"offset" validate:"min=0"`
	
	// 排序参数
	OrderBy   string `json:"order_by" form:"order_by"`
	OrderDesc bool   `json:"order_desc" form:"order_desc"`
	
	// 搜索参数
	Keyword   string   `json:"keyword" form:"keyword"`
	Tags      []string `json:"tags" form:"tags"`
	
	// 高级过滤
	IncludeDeleted bool `json:"include_deleted" form:"include_deleted"`
	OnlyActive     bool `json:"only_active" form:"only_active"`
}

// UserStats 用户统计信息
type UserStats struct {
	TotalUsers   int64            `json:"total_users"`
	ActiveUsers  int64            `json:"active_users"`
	StatusCounts map[string]int64 `json:"status_counts"`
	AgeGroups    map[string]int64 `json:"age_groups"`
	RecentUsers  int64            `json:"recent_users"`
}

// UserProfile 用户档案（关联数据）
type UserProfile struct {
	UserID      int64  `json:"user_id" gorm:"primaryKey" db:"user_id"`
	Bio         string `json:"bio" gorm:"type:text" db:"bio"`
	Website     string `json:"website" gorm:"size:255" db:"website"`
	Location    string `json:"location" gorm:"size:100" db:"location"`
	Company     string `json:"company" gorm:"size:100" db:"company"`
	Occupation  string `json:"occupation" gorm:"size:100" db:"occupation"`
	Education   string `json:"education" gorm:"size:100" db:"education"`
	Skills      string `json:"skills" gorm:"type:text" db:"skills"` // JSON格式存储技能列表
	Preferences string `json:"preferences" gorm:"type:text" db:"preferences"` // JSON格式存储偏好设置
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime" db:"updated_at"`
}

// UserRole 用户角色
type UserRole struct {
	ID          int64     `json:"id" gorm:"primaryKey;autoIncrement" db:"id"`
	UserID      int64     `json:"user_id" gorm:"not null;index" db:"user_id"`
	RoleName    string    `json:"role_name" gorm:"size:50;not null" db:"role_name"`
	Permissions string    `json:"permissions" gorm:"type:text" db:"permissions"` // JSON格式
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime" db:"updated_at"`
}

// Article 文章实体（用于复杂查询测试）
type Article struct {
	ID          int64     `json:"id" gorm:"primaryKey;autoIncrement" db:"id"`
	Title       string    `json:"title" gorm:"size:200;not null" db:"title"`
	Content     string    `json:"content" gorm:"type:longtext" db:"content"`
	Summary     string    `json:"summary" gorm:"size:500" db:"summary"`
	AuthorID    int64     `json:"author_id" gorm:"not null;index" db:"author_id"`
	CategoryID  int64     `json:"category_id" gorm:"index" db:"category_id"`
	Tags        string    `json:"tags" gorm:"size:255" db:"tags"` // 逗号分隔的标签
	Status      string    `json:"status" gorm:"size:20;default:draft" db:"status"`
	ViewCount   int64     `json:"view_count" gorm:"default:0" db:"view_count"`
	LikeCount   int64     `json:"like_count" gorm:"default:0" db:"like_count"`
	CommentCount int64    `json:"comment_count" gorm:"default:0" db:"comment_count"`
	PublishedAt *time.Time `json:"published_at" db:"published_at"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" gorm:"index" db:"deleted_at"`
}

// ArticleQuery 文章查询参数
type ArticleQuery struct {
	Title      string     `json:"title" form:"title"`
	AuthorID   int64      `json:"author_id" form:"author_id"`
	CategoryID int64      `json:"category_id" form:"category_id"`
	Status     string     `json:"status" form:"status"`
	Tags       []string   `json:"tags" form:"tags"`
	Keyword    string     `json:"keyword" form:"keyword"`
	
	// 时间范围
	PublishedAfter  *time.Time `json:"published_after" form:"published_after"`
	PublishedBefore *time.Time `json:"published_before" form:"published_before"`
	
	// 统计范围
	MinViews int64 `json:"min_views" form:"min_views"`
	MaxViews int64 `json:"max_views" form:"max_views"`
	MinLikes int64 `json:"min_likes" form:"min_likes"`
	
	// 分页和排序
	Page      int    `json:"page" form:"page"`
	PageSize  int    `json:"page_size" form:"page_size"`
	OrderBy   string `json:"order_by" form:"order_by"`
	OrderDesc bool   `json:"order_desc" form:"order_desc"`
}

// Category 分类实体
type Category struct {
	ID          int64     `json:"id" gorm:"primaryKey;autoIncrement" db:"id"`
	Name        string    `json:"name" gorm:"size:100;not null;uniqueIndex" db:"name"`
	Slug        string    `json:"slug" gorm:"size:100;not null;uniqueIndex" db:"slug"`
	Description string    `json:"description" gorm:"size:500" db:"description"`
	ParentID    *int64    `json:"parent_id" gorm:"index" db:"parent_id"`
	SortOrder   int       `json:"sort_order" gorm:"default:0" db:"sort_order"`
	IsActive    bool      `json:"is_active" gorm:"default:true" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime" db:"updated_at"`
}

// UserArticleView 用户文章浏览记录（用于复杂连接查询）
type UserArticleView struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement" db:"id"`
	UserID    int64     `json:"user_id" gorm:"not null;index" db:"user_id"`
	ArticleID int64     `json:"article_id" gorm:"not null;index" db:"article_id"`
	ViewTime  time.Time `json:"view_time" gorm:"not null;index" db:"view_time"`
	Duration  int       `json:"duration" gorm:"default:0" db:"duration"` // 阅读时长（秒）
	Device    string    `json:"device" gorm:"size:50" db:"device"`
	UserAgent string    `json:"user_agent" gorm:"size:255" db:"user_agent"`
}

// BatchInsertRequest 批量插入请求
type BatchInsertRequest struct {
	Users    []*User    `json:"users"`
	Articles []*Article `json:"articles"`
}

// BatchUpdateRequest 批量更新请求
type BatchUpdateRequest struct {
	UserIDs []int64           `json:"user_ids"`
	Updates map[string]any    `json:"updates"`
}

// ComplexQueryResult 复杂查询结果
type ComplexQueryResult struct {
	User     *User     `json:"user"`
	Profile  *UserProfile `json:"profile,omitempty"`
	Articles []*Article   `json:"articles,omitempty"`
	Roles    []*UserRole  `json:"roles,omitempty"`
}

// AggregationResult 聚合查询结果
type AggregationResult struct {
	Field string `json:"field"`
	Value any    `json:"value"`
	Count int64  `json:"count"`
}

// PaginationResult 分页结果
type PaginationResult struct {
	Data       any   `json:"data"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// TableName 指定表名
func (u *User) TableName() string { return "users" }
func (p *UserProfile) TableName() string { return "user_profiles" }
func (r *UserRole) TableName() string { return "user_roles" }
func (a *Article) TableName() string { return "articles" }
func (c *Category) TableName() string { return "categories" }
func (v *UserArticleView) TableName() string { return "user_article_views" }

// 辅助方法
func (u *User) IsActive() bool {
	return u.Status == "active"
}

func (u *User) IsAdult() bool {
	return u.Age >= 18
}

func (a *Article) IsPublished() bool {
	return a.Status == "published" && a.PublishedAt != nil
}

func (q *UserQuery) HasAgeFilter() bool {
	return q.AgeMin > 0 || q.AgeMax > 0
}

func (q *UserQuery) HasTimeFilter() bool {
	return q.CreatedAfter != nil || q.CreatedBefore != nil
}

func (q *UserQuery) GetOffset() int {
	if q.Offset > 0 {
		return q.Offset
	}
	if q.Page > 0 && q.PageSize > 0 {
		return (q.Page - 1) * q.PageSize
	}
	return 0
}

func (q *UserQuery) GetLimit() int {
	if q.PageSize > 0 {
		return q.PageSize
	}
	return 10 // 默认限制
}