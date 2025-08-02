package orm

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel 基础模型，包含常用字段
type BaseModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TimestampModel 时间戳模型，只包含时间字段
type TimestampModel struct {
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// SoftDeleteModel 软删除模型
type SoftDeleteModel struct {
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// UserTrackingModel 用户追踪模型，记录创建和更新用户
type UserTrackingModel struct {
	BaseModel
	CreatedBy uint `json:"created_by" gorm:"index"`
	UpdatedBy uint `json:"updated_by" gorm:"index"`
}

// ModelInterface 模型接口
type ModelInterface interface {
	TableName() string
	BeforeCreate(tx *gorm.DB) error
	AfterCreate(tx *gorm.DB) error
	BeforeUpdate(tx *gorm.DB) error
	AfterUpdate(tx *gorm.DB) error
	BeforeDelete(tx *gorm.DB) error
	AfterDelete(tx *gorm.DB) error
}

// Repository 通用仓库接口
type Repository[T any] interface {
	Create(model *T) error
	Update(model *T) error
	Delete(id uint) error
	FindByID(id uint) (*T, error)
	FindAll() ([]*T, error)
	FindWhere(condition string, args ...any) ([]*T, error)
	Count() (int64, error)
	Paginate(page, pageSize int) ([]*T, int64, error)
}

// BaseRepository 基础仓库实现
type BaseRepository[T any] struct {
	db *gorm.DB
}

// NewBaseRepository 创建基础仓库
func NewBaseRepository[T any](db *gorm.DB) *BaseRepository[T] {
	if db == nil {
		db = GetDefaultORM().DB()
	}
	return &BaseRepository[T]{db: db}
}

// Create 创建记录
func (r *BaseRepository[T]) Create(model *T) error {
	return r.db.Create(model).Error
}

// Update 更新记录
func (r *BaseRepository[T]) Update(model *T) error {
	return r.db.Save(model).Error
}

// Delete 删除记录（软删除）
func (r *BaseRepository[T]) Delete(id uint) error {
	var model T
	return r.db.Delete(&model, id).Error
}

// HardDelete 硬删除记录
func (r *BaseRepository[T]) HardDelete(id uint) error {
	var model T
	return r.db.Unscoped().Delete(&model, id).Error
}

// FindByID 根据ID查找记录
func (r *BaseRepository[T]) FindByID(id uint) (*T, error) {
	var model T
	err := r.db.First(&model, id).Error
	if err != nil {
		return nil, err
	}
	return &model, nil
}

// FindAll 查找所有记录
func (r *BaseRepository[T]) FindAll() ([]*T, error) {
	var models []*T
	err := r.db.Find(&models).Error
	return models, err
}

// FindWhere 根据条件查找记录
func (r *BaseRepository[T]) FindWhere(condition string, args ...any) ([]*T, error) {
	var models []*T
	err := r.db.Where(condition, args...).Find(&models).Error
	return models, err
}

// Count 统计记录数量
func (r *BaseRepository[T]) Count() (int64, error) {
	var count int64
	var model T
	err := r.db.Model(&model).Count(&count).Error
	return count, err
}

// CountWhere 根据条件统计记录数量
func (r *BaseRepository[T]) CountWhere(condition string, args ...any) (int64, error) {
	var count int64
	var model T
	err := r.db.Model(&model).Where(condition, args...).Count(&count).Error
	return count, err
}

// Paginate 分页查询
func (r *BaseRepository[T]) Paginate(page, pageSize int) ([]*T, int64, error) {
	var models []*T
	var total int64
	var model T

	// 计算总数
	if err := r.db.Model(&model).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := r.db.Offset(offset).Limit(pageSize).Find(&models).Error
	return models, total, err
}

// PaginateWhere 根据条件分页查询
func (r *BaseRepository[T]) PaginateWhere(condition string, page, pageSize int, args ...any) ([]*T, int64, error) {
	var models []*T
	var total int64
	var model T

	// 计算总数
	if err := r.db.Model(&model).Where(condition, args...).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := r.db.Where(condition, args...).Offset(offset).Limit(pageSize).Find(&models).Error
	return models, total, err
}

// Exists 检查记录是否存在
func (r *BaseRepository[T]) Exists(id uint) (bool, error) {
	var count int64
	var model T
	err := r.db.Model(&model).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// ExistsWhere 根据条件检查记录是否存在
func (r *BaseRepository[T]) ExistsWhere(condition string, args ...any) (bool, error) {
	var count int64
	var model T
	err := r.db.Model(&model).Where(condition, args...).Count(&count).Error
	return count > 0, err
}

// FirstOrCreate 查找或创建记录
func (r *BaseRepository[T]) FirstOrCreate(model *T, conditions ...any) error {
	return r.db.FirstOrCreate(model, conditions...).Error
}

// UpdateColumns 更新指定列
func (r *BaseRepository[T]) UpdateColumns(id uint, columns map[string]any) error {
	var model T
	return r.db.Model(&model).Where("id = ?", id).Updates(columns).Error
}

// UpdateWhere 根据条件更新记录
func (r *BaseRepository[T]) UpdateWhere(condition string, updates map[string]any, args ...any) error {
	var model T
	return r.db.Model(&model).Where(condition, args...).Updates(updates).Error
}

// DeleteWhere 根据条件删除记录
func (r *BaseRepository[T]) DeleteWhere(condition string, args ...any) error {
	var model T
	return r.db.Where(condition, args...).Delete(&model).Error
}

// Restore 恢复软删除的记录
func (r *BaseRepository[T]) Restore(id uint) error {
	var model T
	return r.db.Unscoped().Model(&model).Where("id = ?", id).Update("deleted_at", nil).Error
}

// WithTransaction 在事务中执行操作
func (r *BaseRepository[T]) WithTransaction(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(fn)
}

// GetDB 获取数据库实例
func (r *BaseRepository[T]) GetDB() *gorm.DB {
	return r.db
}

// ============= 查询构建器 =============

// QueryBuilder 查询构建器
type QueryBuilder[T any] struct {
	db *gorm.DB
}

// NewQueryBuilder 创建查询构建器
func NewQueryBuilder[T any](db *gorm.DB) *QueryBuilder[T] {
	if db == nil {
		db = GetDefaultORM().DB()
	}
	var model T
	return &QueryBuilder[T]{db: db.Model(&model)}
}

// Where 添加WHERE条件
func (q *QueryBuilder[T]) Where(condition string, args ...any) *QueryBuilder[T] {
	q.db = q.db.Where(condition, args...)
	return q
}

// Or 添加OR条件
func (q *QueryBuilder[T]) Or(condition string, args ...any) *QueryBuilder[T] {
	q.db = q.db.Or(condition, args...)
	return q
}

// Order 添加排序
func (q *QueryBuilder[T]) Order(order string) *QueryBuilder[T] {
	q.db = q.db.Order(order)
	return q
}

// Limit 限制结果数量
func (q *QueryBuilder[T]) Limit(limit int) *QueryBuilder[T] {
	q.db = q.db.Limit(limit)
	return q
}

// Offset 设置偏移量
func (q *QueryBuilder[T]) Offset(offset int) *QueryBuilder[T] {
	q.db = q.db.Offset(offset)
	return q
}

// Group 添加分组
func (q *QueryBuilder[T]) Group(group string) *QueryBuilder[T] {
	q.db = q.db.Group(group)
	return q
}

// Having 添加HAVING条件
func (q *QueryBuilder[T]) Having(condition string, args ...any) *QueryBuilder[T] {
	q.db = q.db.Having(condition, args...)
	return q
}

// Joins 添加连接
func (q *QueryBuilder[T]) Joins(join string, args ...any) *QueryBuilder[T] {
	q.db = q.db.Joins(join, args...)
	return q
}

// Preload 预加载关联
func (q *QueryBuilder[T]) Preload(associations ...string) *QueryBuilder[T] {
	for _, assoc := range associations {
		q.db = q.db.Preload(assoc)
	}
	return q
}

// Find 执行查询
func (q *QueryBuilder[T]) Find() ([]*T, error) {
	var models []*T
	err := q.db.Find(&models).Error
	return models, err
}

// First 查询第一条记录
func (q *QueryBuilder[T]) First() (*T, error) {
	var model T
	err := q.db.First(&model).Error
	if err != nil {
		return nil, err
	}
	return &model, nil
}

// Count 统计数量
func (q *QueryBuilder[T]) Count() (int64, error) {
	var count int64
	err := q.db.Count(&count).Error
	return count, err
}

// Paginate 分页查询
func (q *QueryBuilder[T]) Paginate(page, pageSize int) ([]*T, int64, error) {
	// 先统计总数
	total, err := q.Count()
	if err != nil {
		return nil, 0, err
	}

	// 再查询数据
	offset := (page - 1) * pageSize
	models, err := q.Offset(offset).Limit(pageSize).Find()
	return models, total, err
}

// ============= 便捷函数 =============

// GetRepository 获取仓库实例
func GetRepository[T any]() *BaseRepository[T] {
	return NewBaseRepository[T](GetDefaultORM().DB())
}

// GetQueryBuilder 获取查询构建器
func GetQueryBuilder[T any]() *QueryBuilder[T] {
	return NewQueryBuilder[T](GetDefaultORM().DB())
}
