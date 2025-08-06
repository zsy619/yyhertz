// Package orm 提供基于GORM的数据库ORM集成
package orm

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

// PaginationType 分页类型
type PaginationType string

const (
	// PaginationTypeOffset 偏移分页
	PaginationTypeOffset PaginationType = "offset"
	// PaginationTypeCursor 游标分页
	PaginationTypeCursor PaginationType = "cursor"
	// PaginationTypeHybrid 混合分页
	PaginationTypeHybrid PaginationType = "hybrid"
)

// PaginationConfig 分页配置
type PaginationConfig struct {
	// 分页类型
	Type PaginationType `json:"type" yaml:"type"`
	// 默认页大小
	DefaultPageSize int `json:"default_page_size" yaml:"default_page_size"`
	// 最大页大小
	MaxPageSize int `json:"max_page_size" yaml:"max_page_size"`
	// 是否返回总数
	IncludeTotal bool `json:"include_total" yaml:"include_total"`
	// 游标字段
	CursorField string `json:"cursor_field" yaml:"cursor_field"`
	// 游标方向
	CursorDirection string `json:"cursor_direction" yaml:"cursor_direction"`
}

// DefaultPaginationConfig 默认分页配置
func DefaultPaginationConfig() *PaginationConfig {
	return &PaginationConfig{
		Type:            PaginationTypeOffset,
		DefaultPageSize: 20,
		MaxPageSize:     100,
		IncludeTotal:    true,
		CursorField:     "id",
		CursorDirection: "ASC",
	}
}

// PaginationResult 分页结果
type PaginationResult struct {
	// 数据
	Data interface{} `json:"data"`
	// 分页信息
	Pagination *PaginationInfo `json:"pagination"`
}

// PaginationInfo 分页信息
type PaginationInfo struct {
	// 当前页
	CurrentPage int `json:"current_page,omitempty"`
	// 页大小
	PageSize int `json:"page_size"`
	// 总数
	Total int64 `json:"total,omitempty"`
	// 总页数
	TotalPages int `json:"total_pages,omitempty"`
	// 是否有下一页
	HasNext bool `json:"has_next"`
	// 是否有上一页
	HasPrev bool `json:"has_prev"`
	// 下一页游标
	NextCursor string `json:"next_cursor,omitempty"`
	// 上一页游标
	PrevCursor string `json:"prev_cursor,omitempty"`
}

// OffsetPaginator 偏移分页器
type OffsetPaginator struct {
	db     *gorm.DB
	config *PaginationConfig
}

// NewOffsetPaginator 创建偏移分页器
func NewOffsetPaginator(db *gorm.DB, config *PaginationConfig) *OffsetPaginator {
	if config == nil {
		config = DefaultPaginationConfig()
	}
	return &OffsetPaginator{
		db:     db,
		config: config,
	}
}

// Paginate 执行偏移分页
func (p *OffsetPaginator) Paginate(page, pageSize int, dest interface{}) (*PaginationResult, error) {
	// 验证参数
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = p.config.DefaultPageSize
	}
	if pageSize > p.config.MaxPageSize {
		pageSize = p.config.MaxPageSize
	}

	// 计算偏移量
	offset := (page - 1) * pageSize

	// 获取总数
	var total int64
	var err error
	if p.config.IncludeTotal {
		err = p.db.Count(&total).Error
		if err != nil {
			return nil, fmt.Errorf("获取总数失败: %w", err)
		}
	}

	// 执行查询
	err = p.db.Limit(pageSize).Offset(offset).Find(dest).Error
	if err != nil {
		return nil, fmt.Errorf("查询数据失败: %w", err)
	}

	// 构建分页信息
	info := &PaginationInfo{
		CurrentPage: page,
		PageSize:    pageSize,
		HasNext:     false,
		HasPrev:     page > 1,
	}

	if p.config.IncludeTotal {
		info.Total = total
		info.TotalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
		info.HasNext = page < info.TotalPages
	} else {
		// 不获取总数时，通过查询下一页来判断是否有更多数据
		var nextCount int64
		err = p.db.Limit(1).Offset(offset + pageSize).Count(&nextCount).Error
		if err == nil {
			info.HasNext = nextCount > 0
		}
	}

	return &PaginationResult{
		Data:       dest,
		Pagination: info,
	}, nil
}

// CursorPaginator 游标分页器
type CursorPaginator struct {
	db     *gorm.DB
	config *PaginationConfig
}

// NewCursorPaginator 创建游标分页器
func NewCursorPaginator(db *gorm.DB, config *PaginationConfig) *CursorPaginator {
	if config == nil {
		config = DefaultPaginationConfig()
		config.Type = PaginationTypeCursor
	}
	return &CursorPaginator{
		db:     db,
		config: config,
	}
}

// CursorInfo 游标信息
type CursorInfo struct {
	Field     string      `json:"field"`
	Value     interface{} `json:"value"`
	Direction string      `json:"direction"`
	Timestamp time.Time   `json:"timestamp"`
}

// EncodeCursor 编码游标
func (p *CursorPaginator) EncodeCursor(info *CursorInfo) string {
	data, err := json.Marshal(info)
	if err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(data)
}

// DecodeCursor 解码游标
func (p *CursorPaginator) DecodeCursor(cursor string) (*CursorInfo, error) {
	if cursor == "" {
		return nil, nil
	}

	data, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return nil, fmt.Errorf("解码游标失败: %w", err)
	}

	var info CursorInfo
	err = json.Unmarshal(data, &info)
	if err != nil {
		return nil, fmt.Errorf("解析游标失败: %w", err)
	}

	return &info, nil
}

// Paginate 执行游标分页
func (p *CursorPaginator) Paginate(cursor string, pageSize int, dest interface{}) (*PaginationResult, error) {
	// 验证参数
	if pageSize < 1 {
		pageSize = p.config.DefaultPageSize
	}
	if pageSize > p.config.MaxPageSize {
		pageSize = p.config.MaxPageSize
	}

	// 解码游标
	cursorInfo, err := p.DecodeCursor(cursor)
	if err != nil {
		return nil, err
	}

	// 构建查询
	query := p.db
	if cursorInfo != nil {
		operator := ">"
		if strings.ToUpper(p.config.CursorDirection) == "DESC" {
			operator = "<"
		}
		query = query.Where(fmt.Sprintf("%s %s ?", p.config.CursorField, operator), cursorInfo.Value)
	}

	// 添加排序
	query = query.Order(fmt.Sprintf("%s %s", p.config.CursorField, p.config.CursorDirection))

	// 查询多一条记录来判断是否有下一页
	err = query.Limit(pageSize + 1).Find(dest).Error
	if err != nil {
		return nil, fmt.Errorf("查询数据失败: %w", err)
	}

	// 检查结果
	results := dest
	hasNext := false

	// 这里需要根据实际的数据类型来处理
	// 简化处理，假设dest是slice类型
	if slice, ok := results.([]interface{}); ok {
		if len(slice) > pageSize {
			hasNext = true
			// 移除多余的记录
			slice = slice[:pageSize]
			dest = slice
		}
	}

	// 构建下一页游标
	var nextCursor string
	if hasNext {
		// 这里需要从最后一条记录中提取游标值
		// 简化处理
		nextCursor = p.EncodeCursor(&CursorInfo{
			Field:     p.config.CursorField,
			Value:     "next_value", // 实际应该从最后一条记录中获取
			Direction: p.config.CursorDirection,
			Timestamp: time.Now(),
		})
	}

	// 构建分页信息
	info := &PaginationInfo{
		PageSize:   pageSize,
		HasNext:    hasNext,
		HasPrev:    cursorInfo != nil,
		NextCursor: nextCursor,
		PrevCursor: cursor,
	}

	return &PaginationResult{
		Data:       dest,
		Pagination: info,
	}, nil
}

// HybridPaginator 混合分页器
type HybridPaginator struct {
	offsetPaginator *OffsetPaginator
	cursorPaginator *CursorPaginator
	config          *PaginationConfig
}

// NewHybridPaginator 创建混合分页器
func NewHybridPaginator(db *gorm.DB, config *PaginationConfig) *HybridPaginator {
	if config == nil {
		config = DefaultPaginationConfig()
		config.Type = PaginationTypeHybrid
	}

	return &HybridPaginator{
		offsetPaginator: NewOffsetPaginator(db, config),
		cursorPaginator: NewCursorPaginator(db, config),
		config:          config,
	}
}

// Paginate 执行混合分页
func (p *HybridPaginator) Paginate(pageOrCursor interface{}, pageSize int, dest interface{}) (*PaginationResult, error) {
	switch v := pageOrCursor.(type) {
	case int:
		// 使用偏移分页
		return p.offsetPaginator.Paginate(v, pageSize, dest)
	case string:
		// 使用游标分页
		return p.cursorPaginator.Paginate(v, pageSize, dest)
	default:
		return nil, fmt.Errorf("不支持的分页参数类型: %T", pageOrCursor)
	}
}

// Paginator 分页器接口
type Paginator interface {
	Paginate(pageOrCursor interface{}, pageSize int, dest interface{}) (*PaginationResult, error)
}

// OffsetPaginatorWrapper 偏移分页器包装器
type OffsetPaginatorWrapper struct {
	*OffsetPaginator
}

// Paginate 实现Paginator接口
func (opw *OffsetPaginatorWrapper) Paginate(pageOrCursor interface{}, pageSize int, dest interface{}) (*PaginationResult, error) {
	page, ok := pageOrCursor.(int)
	if !ok {
		return nil, fmt.Errorf("偏移分页器需要整数页码")
	}
	return opw.OffsetPaginator.Paginate(page, pageSize, dest)
}

// CursorPaginatorWrapper 游标分页器包装器
type CursorPaginatorWrapper struct {
	*CursorPaginator
}

// Paginate 实现Paginator接口
func (cpw *CursorPaginatorWrapper) Paginate(pageOrCursor interface{}, pageSize int, dest interface{}) (*PaginationResult, error) {
	cursor, ok := pageOrCursor.(string)
	if !ok {
		return nil, fmt.Errorf("游标分页器需要字符串游标")
	}
	return cpw.CursorPaginator.Paginate(cursor, pageSize, dest)
}

// PaginatorManager 分页器管理器
type PaginatorManager struct {
	db     *gorm.DB
	config *PaginationConfig
}

// NewPaginatorManager 创建分页器管理器
func NewPaginatorManager(db *gorm.DB, config *PaginationConfig) *PaginatorManager {
	if config == nil {
		config = DefaultPaginationConfig()
	}
	return &PaginatorManager{
		db:     db,
		config: config,
	}
}

// GetPaginator 获取分页器
func (pm *PaginatorManager) GetPaginator() Paginator {
	switch pm.config.Type {
	case PaginationTypeOffset:
		return &OffsetPaginatorWrapper{NewOffsetPaginator(pm.db, pm.config)}
	case PaginationTypeCursor:
		return &CursorPaginatorWrapper{NewCursorPaginator(pm.db, pm.config)}
	case PaginationTypeHybrid:
		return NewHybridPaginator(pm.db, pm.config)
	default:
		return &OffsetPaginatorWrapper{NewOffsetPaginator(pm.db, pm.config)}
	}
}

// OffsetPaginate 偏移分页
func (pm *PaginatorManager) OffsetPaginate(page, pageSize int, dest interface{}) (*PaginationResult, error) {
	paginator := NewOffsetPaginator(pm.db, pm.config)
	return paginator.Paginate(page, pageSize, dest)
}

// CursorPaginate 游标分页
func (pm *PaginatorManager) CursorPaginate(cursor string, pageSize int, dest interface{}) (*PaginationResult, error) {
	paginator := NewCursorPaginator(pm.db, pm.config)
	return paginator.Paginate(cursor, pageSize, dest)
}

// AutoPaginate 自动分页
func (pm *PaginatorManager) AutoPaginate(pageOrCursor interface{}, pageSize int, dest interface{}) (*PaginationResult, error) {
	paginator := pm.GetPaginator()
	return paginator.Paginate(pageOrCursor, pageSize, dest)
}

// ============= 扩展方法 =============

// PaginateWithDB 使用指定数据库连接进行分页
func PaginateWithDB(db *gorm.DB, page, pageSize int, dest interface{}) (*PaginationResult, error) {
	manager := NewPaginatorManager(db, nil)
	return manager.OffsetPaginate(page, pageSize, dest)
}

// CursorPaginateWithDB 使用指定数据库连接进行游标分页
func CursorPaginateWithDB(db *gorm.DB, cursor string, pageSize int, dest interface{}) (*PaginationResult, error) {
	config := DefaultPaginationConfig()
	config.Type = PaginationTypeCursor
	manager := NewPaginatorManager(db, config)
	return manager.CursorPaginate(cursor, pageSize, dest)
}

// ============= 便捷函数 =============

// Paginate 使用默认ORM进行分页
func Paginate(page, pageSize int, dest interface{}) (*PaginationResult, error) {
	return PaginateWithDB(GetDefaultORM().DB(), page, pageSize, dest)
}

// CursorPaginate 使用默认ORM进行游标分页
func CursorPaginate(cursor string, pageSize int, dest interface{}) (*PaginationResult, error) {
	return CursorPaginateWithDB(GetDefaultORM().DB(), cursor, pageSize, dest)
}

// ============= 分页辅助函数 =============

// ParsePageParams 解析分页参数
func ParsePageParams(pageStr, pageSizeStr string) (int, int, error) {
	page := 1
	pageSize := 20

	if pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		if err != nil {
			return 0, 0, fmt.Errorf("无效的页码: %s", pageStr)
		}
		if p > 0 {
			page = p
		}
	}

	if pageSizeStr != "" {
		ps, err := strconv.Atoi(pageSizeStr)
		if err != nil {
			return 0, 0, fmt.Errorf("无效的页大小: %s", pageSizeStr)
		}
		if ps > 0 {
			pageSize = ps
		}
	}

	return page, pageSize, nil
}

// ValidatePageParams 验证分页参数
func ValidatePageParams(page, pageSize int, config *PaginationConfig) (int, int) {
	if config == nil {
		config = DefaultPaginationConfig()
	}

	if page < 1 {
		page = 1
	}

	if pageSize < 1 {
		pageSize = config.DefaultPageSize
	}

	if pageSize > config.MaxPageSize {
		pageSize = config.MaxPageSize
	}

	return page, pageSize
}

// BuildPaginationLinks 构建分页链接
func BuildPaginationLinks(baseURL string, currentPage, totalPages int) map[string]string {
	links := make(map[string]string)

	// 首页
	if currentPage > 1 {
		links["first"] = fmt.Sprintf("%s?page=1", baseURL)
		links["prev"] = fmt.Sprintf("%s?page=%d", baseURL, currentPage-1)
	}

	// 下一页和最后一页
	if currentPage < totalPages {
		links["next"] = fmt.Sprintf("%s?page=%d", baseURL, currentPage+1)
		links["last"] = fmt.Sprintf("%s?page=%d", baseURL, totalPages)
	}

	return links
}

// CalculatePageRange 计算页码范围
func CalculatePageRange(currentPage, totalPages, rangeSize int) (int, int) {
	if rangeSize < 1 {
		rangeSize = 5
	}

	half := rangeSize / 2
	start := currentPage - half
	end := currentPage + half

	if start < 1 {
		start = 1
		end = start + rangeSize - 1
	}

	if end > totalPages {
		end = totalPages
		start = end - rangeSize + 1
		if start < 1 {
			start = 1
		}
	}

	return start, end
}
