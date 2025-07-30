package mvc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBaseController(t *testing.T) {
	controller := NewBaseController()

	assert.NotNil(t, controller)
	assert.NotNil(t, controller.Data)
	assert.Equal(t, "views", controller.ViewPath)
	assert.Equal(t, "views/layout", controller.LayoutPath)
	// logger字段已移除，现在使用单例日志系统
	assert.Empty(t, controller.Data)
}

func TestBaseControllerInit(t *testing.T) {
	controller := &BaseController{}
	controller.Init()

	assert.NotNil(t, controller.Data)
	// logger字段已移除，现在使用单例日志系统
}

func TestBaseControllerDataManagement(t *testing.T) {
	controller := NewBaseController()

	// 测试SetData
	controller.SetData("key1", "value1")
	assert.Equal(t, "value1", controller.Data["key1"])

	// 测试SetDatas
	data := map[string]any{
		"key2": "value2",
		"key3": 123,
	}
	controller.SetDatas(data)
	assert.Equal(t, "value2", controller.Data["key2"])
	assert.Equal(t, 123, controller.Data["key3"])
}

func TestBaseControllerTimeHelpers(t *testing.T) {
	controller := NewBaseController()

	// 测试时间生成函数
	createTime := controller.CreateTime()
	assert.Greater(t, createTime, int64(20240101000000))

	createDate := controller.CreateDate()
	assert.Greater(t, createDate, 20240101)

	updateTime := controller.UpdateTime()
	assert.Greater(t, updateTime, int64(20240101000000))

	updateDate := controller.UpdateDate()
	assert.Greater(t, updateDate, 20240101)
}

func TestBaseControllerGetPageInfo(t *testing.T) {
	// 这个测试需要RequestContext，暂时跳过具体的HTTP参数测试
	// 只测试逻辑方法
	t.Run("GetPageInfoByParam边界情况", func(t *testing.T) {
		controller := NewBaseController()

		// 测试参数为空的情况
		page, pageSize := controller.GetPageInfoByParam("", "", 30)
		assert.Equal(t, 1, page)      // 默认页码
		assert.Equal(t, 30, pageSize) // 自定义默认页大小
	})
}

func TestBaseControllerLifecycle(t *testing.T) {
	controller := NewBaseController()

	// 这些方法应该不会panic
	assert.NotPanics(t, func() {
		controller.Init()
		controller.Prepare()
		controller.Finish()
	})
}

// 基准测试
func BenchmarkNewBaseController(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewBaseController()
	}
}

func BenchmarkBaseControllerSetData(b *testing.B) {
	controller := NewBaseController()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		controller.SetData("test_key", "test_value")
	}
}

func BenchmarkBaseControllerCreateTime(b *testing.B) {
	controller := NewBaseController()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		controller.CreateTime()
	}
}
