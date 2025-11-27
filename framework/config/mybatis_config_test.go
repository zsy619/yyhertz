package config

import (
	"testing"
)

// TestMyBatisConfig 测试MyBatis配置
func TestMyBatisConfig(t *testing.T) {
	// 测试获取MyBatis配置
	config, err := GetMyBatisConfig()
	if err != nil {
		t.Fatalf("获取MyBatis配置失败: %v", err)
	}

	// 验证默认配置值
	if !config.Basic.Enable {
		t.Error("MyBatis应该默认启用")
	}

	if config.Basic.ConfigFile != "./conf/mybatis-config.xml" {
		t.Errorf("期望配置文件路径为 './conf/mybatis-config.xml', 实际为 '%s'", config.Basic.ConfigFile)
	}

	if config.Logging.Level != "info" {
		t.Errorf("期望日志级别为 'info', 实际为 '%s'", config.Logging.Level)
	}

	t.Logf("MyBatis配置测试通过")
}

// TestMyBatisConfigManager 测试MyBatis配置管理器
func TestMyBatisConfigManager(t *testing.T) {
	// 获取配置管理器
	manager := GetMyBatisConfigManager()
	if manager == nil {
		t.Fatal("获取MyBatis配置管理器失败")
	}

	// 测试配置名称
	if manager.GetString("basic.config_file") != "./conf/mybatis-config.xml" {
		t.Error("配置文件路径不正确")
	}

	// 测试布尔值
	if !manager.GetBool("basic.enable") {
		t.Error("MyBatis应该默认启用")
	}

	// 测试整数值
	if manager.GetInt("pool.max_open_conns") != 100 {
		t.Error("最大打开连接数应该为100")
	}

	t.Logf("MyBatis配置管理器测试通过")
}

// TestMyBatisConfigName 测试配置名称常量
func TestMyBatisConfigName(t *testing.T) {
	config := MyBatisConfig{}
	if config.GetConfigName() != MyBatisConfigName {
		t.Errorf("期望配置名称为 '%s', 实际为 '%s'", MyBatisConfigName, config.GetConfigName())
	}

	if MyBatisConfigName != "mybatis" {
		t.Errorf("期望MyBatisConfigName为 'mybatis', 实际为 '%s'", MyBatisConfigName)
	}

	t.Logf("MyBatis配置名称测试通过")
}
