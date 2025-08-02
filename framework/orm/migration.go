package orm

import (
	"fmt"
	"reflect"
	"time"

	"gorm.io/gorm"

	"github.com/zsy619/yyhertz/framework/config"
)

// Migration 数据库迁移接口
type Migration interface {
	Up(db *gorm.DB) error   // 执行迁移
	Down(db *gorm.DB) error // 回滚迁移
	GetVersion() string     // 获取版本号
	GetDescription() string // 获取描述
}

// MigrationInfo 迁移信息
type MigrationInfo struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	Version     string    `gorm:"uniqueIndex;size:255" json:"version"`
	Description string    `gorm:"size:500" json:"description"`
	ExecutedAt  time.Time `json:"executed_at"`
	CreatedAt   time.Time `json:"created_at"`
}

// TableName 表名
func (MigrationInfo) TableName() string {
	return "migrations"
}

// BaseMigration 基础迁移结构
type BaseMigration struct {
	version     string
	description string
}

// NewBaseMigration 创建基础迁移
func NewBaseMigration(version, description string) *BaseMigration {
	return &BaseMigration{
		version:     version,
		description: description,
	}
}

// GetVersion 获取版本号
func (bm *BaseMigration) GetVersion() string {
	return bm.version
}

// GetDescription 获取描述
func (bm *BaseMigration) GetDescription() string {
	return bm.description
}

// Up 默认实现（需要子类重写）
func (bm *BaseMigration) Up(db *gorm.DB) error {
	return fmt.Errorf("Up method not implemented for migration %s", bm.version)
}

// Down 默认实现（需要子类重写）
func (bm *BaseMigration) Down(db *gorm.DB) error {
	return fmt.Errorf("Down method not implemented for migration %s", bm.version)
}

// ModelMigration 模型迁移，用于自动迁移模型
type ModelMigration struct {
	*BaseMigration
	models []any
}

// NewModelMigration 创建模型迁移
func NewModelMigration(version, description string, models ...any) *ModelMigration {
	return &ModelMigration{
		BaseMigration: NewBaseMigration(version, description),
		models:        models,
	}
}

// Up 执行模型自动迁移
func (mm *ModelMigration) Up(db *gorm.DB) error {
	config.Infof("Executing model migration %s: %s", mm.version, mm.description)

	for _, model := range mm.models {
		modelName := reflect.TypeOf(model).String()
		config.Debugf("Auto-migrating model: %s", modelName)

		if err := db.AutoMigrate(model); err != nil {
			return fmt.Errorf("failed to migrate model %s: %w", modelName, err)
		}
	}

	config.Infof("Model migration %s completed successfully", mm.version)
	return nil
}

// Down 模型迁移回滚（删除表）
func (mm *ModelMigration) Down(db *gorm.DB) error {
	config.Warnf("Rolling back model migration %s: %s", mm.version, mm.description)

	for i := len(mm.models) - 1; i >= 0; i-- {
		model := mm.models[i]
		modelName := reflect.TypeOf(model).String()
		config.Debugf("Dropping table for model: %s", modelName)

		if err := db.Migrator().DropTable(model); err != nil {
			config.Errorf("Failed to drop table for model %s: %v", modelName, err)
			return fmt.Errorf("failed to drop table for model %s: %w", modelName, err)
		}
	}

	config.Infof("Model migration %s rolled back successfully", mm.version)
	return nil
}

// SQLMigration SQL迁移，执行原生SQL
type SQLMigration struct {
	*BaseMigration
	upSQL   string
	downSQL string
}

// NewSQLMigration 创建SQL迁移
func NewSQLMigration(version, description, upSQL, downSQL string) *SQLMigration {
	return &SQLMigration{
		BaseMigration: NewBaseMigration(version, description),
		upSQL:         upSQL,
		downSQL:       downSQL,
	}
}

// Up 执行SQL迁移
func (sm *SQLMigration) Up(db *gorm.DB) error {
	config.Infof("Executing SQL migration %s: %s", sm.version, sm.description)

	if sm.upSQL == "" {
		return fmt.Errorf("no up SQL provided for migration %s", sm.version)
	}

	if err := db.Exec(sm.upSQL).Error; err != nil {
		return fmt.Errorf("failed to execute up SQL for migration %s: %w", sm.version, err)
	}

	config.Infof("SQL migration %s completed successfully", sm.version)
	return nil
}

// Down 回滚SQL迁移
func (sm *SQLMigration) Down(db *gorm.DB) error {
	config.Warnf("Rolling back SQL migration %s: %s", sm.version, sm.description)

	if sm.downSQL == "" {
		return fmt.Errorf("no down SQL provided for migration %s", sm.version)
	}

	if err := db.Exec(sm.downSQL).Error; err != nil {
		return fmt.Errorf("failed to execute down SQL for migration %s: %w", sm.version, err)
	}

	config.Infof("SQL migration %s rolled back successfully", sm.version)
	return nil
}

// MigrationManager 迁移管理器
type MigrationManager struct {
	db         *gorm.DB
	migrations []Migration
}

// NewMigrationManager 创建迁移管理器
func NewMigrationManager(db *gorm.DB) *MigrationManager {
	if db == nil {
		db = GetDefaultORM().DB()
	}
	return &MigrationManager{
		db:         db,
		migrations: make([]Migration, 0),
	}
}

// GetDefaultMigrationManager 获取默认迁移管理器
func GetDefaultMigrationManager() *MigrationManager {
	return NewMigrationManager(GetDefaultORM().DB())
}

// Add 添加迁移
func (mm *MigrationManager) Add(migration Migration) *MigrationManager {
	mm.migrations = append(mm.migrations, migration)
	return mm
}

// AddModel 添加模型迁移
func (mm *MigrationManager) AddModel(version, description string, models ...any) *MigrationManager {
	return mm.Add(NewModelMigration(version, description, models...))
}

// AddSQL 添加SQL迁移
func (mm *MigrationManager) AddSQL(version, description, upSQL, downSQL string) *MigrationManager {
	return mm.Add(NewSQLMigration(version, description, upSQL, downSQL))
}

// Initialize 初始化迁移表
func (mm *MigrationManager) Initialize() error {
	config.Info("Initializing migration system...")

	// 创建迁移信息表
	if err := mm.db.AutoMigrate(&MigrationInfo{}); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	config.Info("Migration system initialized successfully")
	return nil
}

// GetExecutedMigrations 获取已执行的迁移
func (mm *MigrationManager) GetExecutedMigrations() (map[string]*MigrationInfo, error) {
	var migrations []*MigrationInfo
	if err := mm.db.Find(&migrations).Error; err != nil {
		return nil, fmt.Errorf("failed to get executed migrations: %w", err)
	}

	result := make(map[string]*MigrationInfo)
	for _, migration := range migrations {
		result[migration.Version] = migration
	}

	return result, nil
}

// Migrate 执行所有未执行的迁移
func (mm *MigrationManager) Migrate() error {
	if err := mm.Initialize(); err != nil {
		return err
	}

	config.Info("Starting database migration...")

	executed, err := mm.GetExecutedMigrations()
	if err != nil {
		return err
	}

	pendingCount := 0
	for _, migration := range mm.migrations {
		if _, exists := executed[migration.GetVersion()]; !exists {
			pendingCount++
		}
	}

	if pendingCount == 0 {
		config.Info("No pending migrations found")
		return nil
	}

	config.Infof("Found %d pending migrations", pendingCount)

	for _, migration := range mm.migrations {
		version := migration.GetVersion()

		// 检查是否已执行
		if _, exists := executed[version]; exists {
			config.Debugf("Migration %s already executed, skipping", version)
			continue
		}

		config.Infof("Executing migration %s: %s", version, migration.GetDescription())

		// 在事务中执行迁移
		err := mm.db.Transaction(func(tx *gorm.DB) error {
			// 执行迁移
			if err := migration.Up(tx); err != nil {
				return err
			}

			// 记录迁移信息
			migrationInfo := &MigrationInfo{
				Version:     version,
				Description: migration.GetDescription(),
				ExecutedAt:  time.Now(),
				CreatedAt:   time.Now(),
			}

			return tx.Create(migrationInfo).Error
		})

		if err != nil {
			config.Errorf("Migration %s failed: %v", version, err)
			return fmt.Errorf("migration %s failed: %w", version, err)
		}

		config.Infof("Migration %s completed successfully", version)
	}

	config.Info("All migrations completed successfully")
	return nil
}

// Rollback 回滚指定数量的迁移
func (mm *MigrationManager) Rollback(steps int) error {
	if err := mm.Initialize(); err != nil {
		return err
	}

	config.Infof("Starting rollback of %d migrations...", steps)

	// 获取已执行的迁移，按执行时间倒序
	var executedMigrations []*MigrationInfo
	if err := mm.db.Order("executed_at DESC").Limit(steps).Find(&executedMigrations).Error; err != nil {
		return fmt.Errorf("failed to get executed migrations for rollback: %w", err)
	}

	if len(executedMigrations) == 0 {
		config.Info("No migrations to rollback")
		return nil
	}

	// 创建迁移映射
	migrationMap := make(map[string]Migration)
	for _, migration := range mm.migrations {
		migrationMap[migration.GetVersion()] = migration
	}

	for _, migrationInfo := range executedMigrations {
		version := migrationInfo.Version
		migration, exists := migrationMap[version]

		if !exists {
			config.Warnf("Migration %s not found in registered migrations, skipping rollback", version)
			continue
		}

		config.Infof("Rolling back migration %s: %s", version, migration.GetDescription())

		// 在事务中执行回滚
		err := mm.db.Transaction(func(tx *gorm.DB) error {
			// 执行回滚
			if err := migration.Down(tx); err != nil {
				return err
			}

			// 删除迁移记录
			return tx.Delete(&MigrationInfo{}, "version = ?", version).Error
		})

		if err != nil {
			config.Errorf("Rollback of migration %s failed: %v", version, err)
			return fmt.Errorf("rollback of migration %s failed: %w", version, err)
		}

		config.Infof("Migration %s rolled back successfully", version)
	}

	config.Info("All rollbacks completed successfully")
	return nil
}

// RollbackTo 回滚到指定版本
func (mm *MigrationManager) RollbackTo(targetVersion string) error {
	if err := mm.Initialize(); err != nil {
		return err
	}

	config.Infof("Rolling back to migration version: %s", targetVersion)

	// 获取目标版本之后的所有迁移
	var migrationsToRollback []*MigrationInfo
	if err := mm.db.Where("version > ?", targetVersion).Order("executed_at DESC").Find(&migrationsToRollback).Error; err != nil {
		return fmt.Errorf("failed to get migrations to rollback: %w", err)
	}

	if len(migrationsToRollback) == 0 {
		config.Info("No migrations to rollback")
		return nil
	}

	// 创建迁移映射
	migrationMap := make(map[string]Migration)
	for _, migration := range mm.migrations {
		migrationMap[migration.GetVersion()] = migration
	}

	for _, migrationInfo := range migrationsToRollback {
		version := migrationInfo.Version
		migration, exists := migrationMap[version]

		if !exists {
			config.Warnf("Migration %s not found in registered migrations, skipping rollback", version)
			continue
		}

		config.Infof("Rolling back migration %s: %s", version, migration.GetDescription())

		// 在事务中执行回滚
		err := mm.db.Transaction(func(tx *gorm.DB) error {
			// 执行回滚
			if err := migration.Down(tx); err != nil {
				return err
			}

			// 删除迁移记录
			return tx.Delete(&MigrationInfo{}, "version = ?", version).Error
		})

		if err != nil {
			config.Errorf("Rollback of migration %s failed: %v", version, err)
			return fmt.Errorf("rollback of migration %s failed: %w", version, err)
		}

		config.Infof("Migration %s rolled back successfully", version)
	}

	config.Infof("Successfully rolled back to version: %s", targetVersion)
	return nil
}

// Status 获取迁移状态
func (mm *MigrationManager) Status() (*MigrationStatus, error) {
	if err := mm.Initialize(); err != nil {
		return nil, err
	}

	executed, err := mm.GetExecutedMigrations()
	if err != nil {
		return nil, err
	}

	status := &MigrationStatus{
		TotalMigrations:    len(mm.migrations),
		ExecutedMigrations: len(executed),
		PendingMigrations:  0,
		Migrations:         make([]*MigrationStatusItem, 0),
	}

	for _, migration := range mm.migrations {
		version := migration.GetVersion()
		item := &MigrationStatusItem{
			Version:     version,
			Description: migration.GetDescription(),
			Executed:    false,
		}

		if migrationInfo, exists := executed[version]; exists {
			item.Executed = true
			item.ExecutedAt = &migrationInfo.ExecutedAt
		} else {
			status.PendingMigrations++
		}

		status.Migrations = append(status.Migrations, item)
	}

	return status, nil
}

// MigrationStatus 迁移状态
type MigrationStatus struct {
	TotalMigrations    int                    `json:"total_migrations"`
	ExecutedMigrations int                    `json:"executed_migrations"`
	PendingMigrations  int                    `json:"pending_migrations"`
	Migrations         []*MigrationStatusItem `json:"migrations"`
}

// MigrationStatusItem 迁移状态项
type MigrationStatusItem struct {
	Version     string     `json:"version"`
	Description string     `json:"description"`
	Executed    bool       `json:"executed"`
	ExecutedAt  *time.Time `json:"executed_at,omitempty"`
}

// ============= 便捷函数 =============

// Migrate 使用默认管理器执行迁移
func Migrate() error {
	return GetDefaultMigrationManager().Migrate()
}

// Rollback 使用默认管理器回滚迁移
func Rollback(steps int) error {
	return GetDefaultMigrationManager().Rollback(steps)
}

// RollbackTo 使用默认管理器回滚到指定版本
func RollbackTo(version string) error {
	return GetDefaultMigrationManager().RollbackTo(version)
}

// GetMigrationStatus 获取迁移状态
func GetMigrationStatus() (*MigrationStatus, error) {
	return GetDefaultMigrationManager().Status()
}
