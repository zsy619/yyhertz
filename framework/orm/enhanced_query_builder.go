// Package orm 提供基于GORM的数据库ORM集成
package orm

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// EnhancedQueryBuilder 增强查询构造器
type EnhancedQueryBuilder struct {
	db        *gorm.DB
	tableName string
	fields    []string
	wheres    []whereCondition
	orders    []orderCondition
	groups    []string
	havings   []string
	joins     []joinCondition
	limit     int
	offset    int
	distinct  bool
	forUpdate bool
	lock      string
}

// whereCondition WHERE条件
type whereCondition struct {
	Type     string                // 条件类型：where, orWhere, whereIn, whereNotIn, whereBetween, whereNull, whereNotNull
	Column   string                // 列名
	Operator string                // 操作符
	Value    interface{}           // 值
	Boolean  string                // 布尔连接符：AND, OR
	SubQuery *EnhancedQueryBuilder // 子查询
}

// orderCondition 排序条件
type orderCondition struct {
	Column    string // 列名
	Direction string // 排序方向：ASC, DESC
}

// joinCondition JOIN条件
type joinCondition struct {
	Type      string // JOIN类型：INNER, LEFT, RIGHT, FULL
	Table     string // 表名
	Condition string // 连接条件
}

// NewEnhancedQueryBuilder 创建增强查询构造器
func NewEnhancedQueryBuilder(db *gorm.DB) *EnhancedQueryBuilder {
	return &EnhancedQueryBuilder{
		db:      db,
		fields:  []string{"*"},
		wheres:  make([]whereCondition, 0),
		orders:  make([]orderCondition, 0),
		groups:  make([]string, 0),
		havings: make([]string, 0),
		joins:   make([]joinCondition, 0),
		limit:   -1,
		offset:  -1,
	}
}

// Table 设置表名
func (qb *EnhancedQueryBuilder) Table(tableName string) *EnhancedQueryBuilder {
	qb.tableName = tableName
	return qb
}

// Select 设置查询字段
func (qb *EnhancedQueryBuilder) Select(fields ...string) *EnhancedQueryBuilder {
	qb.fields = fields
	return qb
}

// Distinct 设置去重
func (qb *EnhancedQueryBuilder) Distinct() *EnhancedQueryBuilder {
	qb.distinct = true
	return qb
}

// Where 添加WHERE条件
func (qb *EnhancedQueryBuilder) Where(column string, operator string, value interface{}) *EnhancedQueryBuilder {
	qb.wheres = append(qb.wheres, whereCondition{
		Type:     "where",
		Column:   column,
		Operator: operator,
		Value:    value,
		Boolean:  "AND",
	})
	return qb
}

// OrWhere 添加OR WHERE条件
func (qb *EnhancedQueryBuilder) OrWhere(column string, operator string, value interface{}) *EnhancedQueryBuilder {
	qb.wheres = append(qb.wheres, whereCondition{
		Type:     "where",
		Column:   column,
		Operator: operator,
		Value:    value,
		Boolean:  "OR",
	})
	return qb
}

// WhereIn 添加WHERE IN条件
func (qb *EnhancedQueryBuilder) WhereIn(column string, values interface{}) *EnhancedQueryBuilder {
	qb.wheres = append(qb.wheres, whereCondition{
		Type:    "whereIn",
		Column:  column,
		Value:   values,
		Boolean: "AND",
	})
	return qb
}

// WhereNotIn 添加WHERE NOT IN条件
func (qb *EnhancedQueryBuilder) WhereNotIn(column string, values interface{}) *EnhancedQueryBuilder {
	qb.wheres = append(qb.wheres, whereCondition{
		Type:    "whereNotIn",
		Column:  column,
		Value:   values,
		Boolean: "AND",
	})
	return qb
}

// WhereBetween 添加WHERE BETWEEN条件
func (qb *EnhancedQueryBuilder) WhereBetween(column string, min, max interface{}) *EnhancedQueryBuilder {
	qb.wheres = append(qb.wheres, whereCondition{
		Type:    "whereBetween",
		Column:  column,
		Value:   []interface{}{min, max},
		Boolean: "AND",
	})
	return qb
}

// WhereNotBetween 添加WHERE NOT BETWEEN条件
func (qb *EnhancedQueryBuilder) WhereNotBetween(column string, min, max interface{}) *EnhancedQueryBuilder {
	qb.wheres = append(qb.wheres, whereCondition{
		Type:    "whereNotBetween",
		Column:  column,
		Value:   []interface{}{min, max},
		Boolean: "AND",
	})
	return qb
}

// WhereNull 添加WHERE NULL条件
func (qb *EnhancedQueryBuilder) WhereNull(column string) *EnhancedQueryBuilder {
	qb.wheres = append(qb.wheres, whereCondition{
		Type:    "whereNull",
		Column:  column,
		Boolean: "AND",
	})
	return qb
}

// WhereNotNull 添加WHERE NOT NULL条件
func (qb *EnhancedQueryBuilder) WhereNotNull(column string) *EnhancedQueryBuilder {
	qb.wheres = append(qb.wheres, whereCondition{
		Type:    "whereNotNull",
		Column:  column,
		Boolean: "AND",
	})
	return qb
}

// WhereLike 添加WHERE LIKE条件
func (qb *EnhancedQueryBuilder) WhereLike(column string, pattern string) *EnhancedQueryBuilder {
	qb.wheres = append(qb.wheres, whereCondition{
		Type:     "where",
		Column:   column,
		Operator: "LIKE",
		Value:    pattern,
		Boolean:  "AND",
	})
	return qb
}

// WhereExists 添加WHERE EXISTS条件
func (qb *EnhancedQueryBuilder) WhereExists(subQuery *EnhancedQueryBuilder) *EnhancedQueryBuilder {
	qb.wheres = append(qb.wheres, whereCondition{
		Type:     "whereExists",
		SubQuery: subQuery,
		Boolean:  "AND",
	})
	return qb
}

// WhereNotExists 添加WHERE NOT EXISTS条件
func (qb *EnhancedQueryBuilder) WhereNotExists(subQuery *EnhancedQueryBuilder) *EnhancedQueryBuilder {
	qb.wheres = append(qb.wheres, whereCondition{
		Type:     "whereNotExists",
		SubQuery: subQuery,
		Boolean:  "AND",
	})
	return qb
}

// OrderBy 添加排序条件
func (qb *EnhancedQueryBuilder) OrderBy(column string, direction ...string) *EnhancedQueryBuilder {
	dir := "ASC"
	if len(direction) > 0 && strings.ToUpper(direction[0]) == "DESC" {
		dir = "DESC"
	}

	qb.orders = append(qb.orders, orderCondition{
		Column:    column,
		Direction: dir,
	})
	return qb
}

// OrderByDesc 添加降序排序条件
func (qb *EnhancedQueryBuilder) OrderByDesc(column string) *EnhancedQueryBuilder {
	return qb.OrderBy(column, "DESC")
}

// GroupBy 添加分组条件
func (qb *EnhancedQueryBuilder) GroupBy(columns ...string) *EnhancedQueryBuilder {
	qb.groups = append(qb.groups, columns...)
	return qb
}

// Having 添加HAVING条件
func (qb *EnhancedQueryBuilder) Having(condition string) *EnhancedQueryBuilder {
	qb.havings = append(qb.havings, condition)
	return qb
}

// Join 添加INNER JOIN
func (qb *EnhancedQueryBuilder) Join(table string, condition string) *EnhancedQueryBuilder {
	qb.joins = append(qb.joins, joinCondition{
		Type:      "INNER",
		Table:     table,
		Condition: condition,
	})
	return qb
}

// LeftJoin 添加LEFT JOIN
func (qb *EnhancedQueryBuilder) LeftJoin(table string, condition string) *EnhancedQueryBuilder {
	qb.joins = append(qb.joins, joinCondition{
		Type:      "LEFT",
		Table:     table,
		Condition: condition,
	})
	return qb
}

// RightJoin 添加RIGHT JOIN
func (qb *EnhancedQueryBuilder) RightJoin(table string, condition string) *EnhancedQueryBuilder {
	qb.joins = append(qb.joins, joinCondition{
		Type:      "RIGHT",
		Table:     table,
		Condition: condition,
	})
	return qb
}

// Limit 设置限制条数
func (qb *EnhancedQueryBuilder) Limit(limit int) *EnhancedQueryBuilder {
	qb.limit = limit
	return qb
}

// Offset 设置偏移量
func (qb *EnhancedQueryBuilder) Offset(offset int) *EnhancedQueryBuilder {
	qb.offset = offset
	return qb
}

// ForUpdate 设置FOR UPDATE锁
func (qb *EnhancedQueryBuilder) ForUpdate() *EnhancedQueryBuilder {
	qb.forUpdate = true
	return qb
}

// Lock 设置自定义锁
func (qb *EnhancedQueryBuilder) Lock(lockType string) *EnhancedQueryBuilder {
	qb.lock = lockType
	return qb
}

// ToSQL 生成SQL语句
func (qb *EnhancedQueryBuilder) ToSQL() (string, []interface{}) {
	var sql strings.Builder
	var args []interface{}

	// SELECT子句
	sql.WriteString("SELECT ")
	if qb.distinct {
		sql.WriteString("DISTINCT ")
	}
	sql.WriteString(strings.Join(qb.fields, ", "))

	// FROM子句
	if qb.tableName != "" {
		sql.WriteString(" FROM ")
		sql.WriteString(qb.tableName)
	}

	// JOIN子句
	for _, join := range qb.joins {
		sql.WriteString(fmt.Sprintf(" %s JOIN %s ON %s", join.Type, join.Table, join.Condition))
	}

	// WHERE子句
	if len(qb.wheres) > 0 {
		sql.WriteString(" WHERE ")
		for i, where := range qb.wheres {
			if i > 0 {
				sql.WriteString(fmt.Sprintf(" %s ", where.Boolean))
			}

			switch where.Type {
			case "where":
				sql.WriteString(fmt.Sprintf("%s %s ?", where.Column, where.Operator))
				args = append(args, where.Value)
			case "whereIn":
				sql.WriteString(fmt.Sprintf("%s IN (?)", where.Column))
				args = append(args, where.Value)
			case "whereNotIn":
				sql.WriteString(fmt.Sprintf("%s NOT IN (?)", where.Column))
				args = append(args, where.Value)
			case "whereBetween":
				sql.WriteString(fmt.Sprintf("%s BETWEEN ? AND ?", where.Column))
				values := where.Value.([]interface{})
				args = append(args, values[0], values[1])
			case "whereNotBetween":
				sql.WriteString(fmt.Sprintf("%s NOT BETWEEN ? AND ?", where.Column))
				values := where.Value.([]interface{})
				args = append(args, values[0], values[1])
			case "whereNull":
				sql.WriteString(fmt.Sprintf("%s IS NULL", where.Column))
			case "whereNotNull":
				sql.WriteString(fmt.Sprintf("%s IS NOT NULL", where.Column))
			case "whereExists":
				subSQL, subArgs := where.SubQuery.ToSQL()
				sql.WriteString(fmt.Sprintf("EXISTS (%s)", subSQL))
				args = append(args, subArgs...)
			case "whereNotExists":
				subSQL, subArgs := where.SubQuery.ToSQL()
				sql.WriteString(fmt.Sprintf("NOT EXISTS (%s)", subSQL))
				args = append(args, subArgs...)
			}
		}
	}

	// GROUP BY子句
	if len(qb.groups) > 0 {
		sql.WriteString(" GROUP BY ")
		sql.WriteString(strings.Join(qb.groups, ", "))
	}

	// HAVING子句
	if len(qb.havings) > 0 {
		sql.WriteString(" HAVING ")
		sql.WriteString(strings.Join(qb.havings, " AND "))
	}

	// ORDER BY子句
	if len(qb.orders) > 0 {
		sql.WriteString(" ORDER BY ")
		orderParts := make([]string, len(qb.orders))
		for i, order := range qb.orders {
			orderParts[i] = fmt.Sprintf("%s %s", order.Column, order.Direction)
		}
		sql.WriteString(strings.Join(orderParts, ", "))
	}

	// LIMIT子句
	if qb.limit > 0 {
		sql.WriteString(fmt.Sprintf(" LIMIT %d", qb.limit))
	}

	// OFFSET子句
	if qb.offset > 0 {
		sql.WriteString(fmt.Sprintf(" OFFSET %d", qb.offset))
	}

	// 锁子句
	if qb.forUpdate {
		sql.WriteString(" FOR UPDATE")
	} else if qb.lock != "" {
		sql.WriteString(fmt.Sprintf(" %s", qb.lock))
	}

	return sql.String(), args
}

// ToGormDB 转换为GORM DB对象
func (qb *EnhancedQueryBuilder) ToGormDB() *gorm.DB {
	db := qb.db

	// 设置表名
	if qb.tableName != "" {
		db = db.Table(qb.tableName)
	}

	// 设置字段
	if len(qb.fields) > 0 && qb.fields[0] != "*" {
		db = db.Select(qb.fields)
	}

	// 设置去重
	if qb.distinct {
		db = db.Distinct()
	}

	// 设置WHERE条件
	for _, where := range qb.wheres {
		switch where.Type {
		case "where":
			if where.Boolean == "OR" {
				db = db.Or(fmt.Sprintf("%s %s ?", where.Column, where.Operator), where.Value)
			} else {
				db = db.Where(fmt.Sprintf("%s %s ?", where.Column, where.Operator), where.Value)
			}
		case "whereIn":
			if where.Boolean == "OR" {
				db = db.Or(fmt.Sprintf("%s IN (?)", where.Column), where.Value)
			} else {
				db = db.Where(fmt.Sprintf("%s IN (?)", where.Column), where.Value)
			}
		case "whereNotIn":
			if where.Boolean == "OR" {
				db = db.Or(fmt.Sprintf("%s NOT IN (?)", where.Column), where.Value)
			} else {
				db = db.Where(fmt.Sprintf("%s NOT IN (?)", where.Column), where.Value)
			}
		case "whereBetween":
			values := where.Value.([]interface{})
			if where.Boolean == "OR" {
				db = db.Or(fmt.Sprintf("%s BETWEEN ? AND ?", where.Column), values[0], values[1])
			} else {
				db = db.Where(fmt.Sprintf("%s BETWEEN ? AND ?", where.Column), values[0], values[1])
			}
		case "whereNotBetween":
			values := where.Value.([]interface{})
			if where.Boolean == "OR" {
				db = db.Or(fmt.Sprintf("%s NOT BETWEEN ? AND ?", where.Column), values[0], values[1])
			} else {
				db = db.Where(fmt.Sprintf("%s NOT BETWEEN ? AND ?", where.Column), values[0], values[1])
			}
		case "whereNull":
			if where.Boolean == "OR" {
				db = db.Or(fmt.Sprintf("%s IS NULL", where.Column))
			} else {
				db = db.Where(fmt.Sprintf("%s IS NULL", where.Column))
			}
		case "whereNotNull":
			if where.Boolean == "OR" {
				db = db.Or(fmt.Sprintf("%s IS NOT NULL", where.Column))
			} else {
				db = db.Where(fmt.Sprintf("%s IS NOT NULL", where.Column))
			}
		}
	}

	// 设置JOIN
	for _, join := range qb.joins {
		switch strings.ToUpper(join.Type) {
		case "INNER":
			db = db.Joins(fmt.Sprintf("INNER JOIN %s ON %s", join.Table, join.Condition))
		case "LEFT":
			db = db.Joins(fmt.Sprintf("LEFT JOIN %s ON %s", join.Table, join.Condition))
		case "RIGHT":
			db = db.Joins(fmt.Sprintf("RIGHT JOIN %s ON %s", join.Table, join.Condition))
		}
	}

	// 设置GROUP BY
	if len(qb.groups) > 0 {
		db = db.Group(strings.Join(qb.groups, ", "))
	}

	// 设置HAVING
	for _, having := range qb.havings {
		db = db.Having(having)
	}

	// 设置ORDER BY
	for _, order := range qb.orders {
		db = db.Order(fmt.Sprintf("%s %s", order.Column, order.Direction))
	}

	// 设置LIMIT
	if qb.limit > 0 {
		db = db.Limit(qb.limit)
	}

	// 设置OFFSET
	if qb.offset > 0 {
		db = db.Offset(qb.offset)
	}

	return db
}

// Get 执行查询并返回结果
func (qb *EnhancedQueryBuilder) Get(dest interface{}) error {
	return qb.ToGormDB().Find(dest).Error
}

// First 执行查询并返回第一条记录
func (qb *EnhancedQueryBuilder) First(dest interface{}) error {
	return qb.ToGormDB().First(dest).Error
}

// Count 执行计数查询
func (qb *EnhancedQueryBuilder) Count() (int64, error) {
	var count int64
	err := qb.ToGormDB().Count(&count).Error
	return count, err
}

// Exists 检查记录是否存在
func (qb *EnhancedQueryBuilder) Exists() (bool, error) {
	count, err := qb.Count()
	return count > 0, err
}

// Paginate 分页查询
func (qb *EnhancedQueryBuilder) Paginate(page, pageSize int, dest interface{}) (int64, error) {
	// 计算总数
	total, err := qb.Count()
	if err != nil {
		return 0, err
	}

	// 计算偏移量
	offset := (page - 1) * pageSize

	// 执行分页查询
	err = qb.Limit(pageSize).Offset(offset).Get(dest)
	return total, err
}

// Clone 克隆查询构造器
func (qb *EnhancedQueryBuilder) Clone() *EnhancedQueryBuilder {
	newQB := &EnhancedQueryBuilder{
		db:        qb.db,
		tableName: qb.tableName,
		fields:    make([]string, len(qb.fields)),
		wheres:    make([]whereCondition, len(qb.wheres)),
		orders:    make([]orderCondition, len(qb.orders)),
		groups:    make([]string, len(qb.groups)),
		havings:   make([]string, len(qb.havings)),
		joins:     make([]joinCondition, len(qb.joins)),
		limit:     qb.limit,
		offset:    qb.offset,
		distinct:  qb.distinct,
		forUpdate: qb.forUpdate,
		lock:      qb.lock,
	}

	copy(newQB.fields, qb.fields)
	copy(newQB.wheres, qb.wheres)
	copy(newQB.orders, qb.orders)
	copy(newQB.groups, qb.groups)
	copy(newQB.havings, qb.havings)
	copy(newQB.joins, qb.joins)

	return newQB
}

// ============= 便捷函数 =============

// NewEnhancedQuery 创建新的增强查询构造器
func NewEnhancedQuery(db *gorm.DB) *EnhancedQueryBuilder {
	return NewEnhancedQueryBuilder(db)
}

// EnhancedQuery 使用默认ORM创建增强查询构造器
func EnhancedQuery() *EnhancedQueryBuilder {
	return NewEnhancedQueryBuilder(GetDefaultORM().DB())
}
