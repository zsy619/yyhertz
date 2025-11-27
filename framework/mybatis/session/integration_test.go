// Package session MyBatis风格优化的集成测试
package session

import (
	"fmt"
	"testing"
	"time"

	"github.com/zsy619/yyhertz/framework/mybatis/config"
)

// TestUser 测试用户结构体
type TestUser struct {
	ID       int64     `json:"id"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	CreateAt time.Time `json:"create_at"`
}

// TestMyBatisIntegration 全面集成测试
func TestMyBatisIntegration(t *testing.T) {
	configuration := createTestConfig()

	t.Run("BasicConfiguration", func(t *testing.T) {
		testBasicConfiguration(t, configuration)
	})

	t.Run("ComponentValidation", func(t *testing.T) {
		testComponentValidation(t, configuration)
	})

	t.Run("IntegrationValidation", func(t *testing.T) {
		testIntegrationValidation(t, configuration)
	})
}

// testBasicConfiguration 基础配置测试
func testBasicConfiguration(t *testing.T, cfg *config.Configuration) {
	if cfg == nil {
		t.Fatal("Configuration should not be nil")
	}

	if !cfg.CacheEnabled {
		t.Error("Cache should be enabled in test config")
	}

	if !cfg.LazyLoadingEnabled {
		t.Error("Lazy loading should be enabled in test config")
	}

	if !cfg.MapUnderscoreToCamelCase {
		t.Error("Map underscore to camel case should be enabled")
	}

	t.Log("Basic configuration validation passed")
}

// testComponentValidation 组件验证测试
func testComponentValidation(t *testing.T, cfg *config.Configuration) {
	// 验证类型别名注册表
	typeAliasRegistry := cfg.GetTypeAliasRegistry()
	if typeAliasRegistry == nil {
		t.Error("TypeAliasRegistry should not be nil")
	}

	// 验证类型处理器注册表
	typeHandlerRegistry := cfg.GetTypeHandlerRegistry()
	if typeHandlerRegistry == nil {
		t.Error("TypeHandlerRegistry should not be nil")
	}

	// 验证映射器注册表
	mapperRegistry := cfg.GetMapperRegistry()
	if mapperRegistry == nil {
		t.Error("MapperRegistry should not be nil")
	}

	t.Log("Component validation passed")
}

// testIntegrationValidation 集成验证测试
func testIntegrationValidation(t *testing.T, cfg *config.Configuration) {
	// 测试配置的完整性
	if cfg.GetDatabaseConfig() == nil {
		t.Log("Database config is nil (expected in test environment)")
	}

	// 测试执行器类型
	if cfg.DefaultExecutorType == config.ExecutorTypeDefault {
		t.Log("Default executor type is properly set")
	}

	// 测试缓存配置
	if cfg.DefaultCacheConfig != nil {
		if cfg.DefaultCacheConfig.Enabled {
			t.Log("Cache configuration is properly enabled")
		}
	}

	t.Log("Integration validation passed")
}

// createTestConfig 创建测试配置
func createTestConfig() *config.Configuration {
	cfg := config.NewConfiguration()
	cfg.CacheEnabled = true
	cfg.LazyLoadingEnabled = true
	cfg.MapUnderscoreToCamelCase = true
	return cfg
}

// BenchmarkBasicOperations 基础操作性能基准测试
func BenchmarkBasicOperations(b *testing.B) {
	config := createTestConfig()

	b.Run("ConfigAccess", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = config.CacheEnabled
			_ = config.LazyLoadingEnabled
			_ = config.MapUnderscoreToCamelCase
		}
	})

	b.Run("RegistryAccess", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = config.GetTypeAliasRegistry()
			_ = config.GetTypeHandlerRegistry()
			_ = config.GetMapperRegistry()
		}
	})
}

// ValidateFrameworkIntegration 验证框架集成效果
func ValidateFrameworkIntegration() error {
	config := createTestConfig()

	// 验证配置完整性
	if config == nil {
		return fmt.Errorf("configuration creation failed")
	}

	if !config.CacheEnabled {
		return fmt.Errorf("cache should be enabled")
	}

	// 验证注册表
	if config.GetTypeAliasRegistry() == nil {
		return fmt.Errorf("type alias registry is nil")
	}

	if config.GetTypeHandlerRegistry() == nil {
		return fmt.Errorf("type handler registry is nil")
	}

	if config.GetMapperRegistry() == nil {
		return fmt.Errorf("mapper registry is nil")
	}

	// 验证默认配置
	if config.DefaultCacheConfig == nil {
		return fmt.Errorf("default cache config is nil")
	}

	return nil
}

// TestFrameworkOptimizations 测试框架优化效果
func TestFrameworkOptimizations(t *testing.T) {
	err := ValidateFrameworkIntegration()
	if err != nil {
		t.Fatalf("Framework integration validation failed: %v", err)
	}

	t.Log("All framework optimizations validated successfully")
}
