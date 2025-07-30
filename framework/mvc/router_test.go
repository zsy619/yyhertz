package mvc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestController 测试控制器
type TestController struct {
	BaseController
}

func (c *TestController) GetIndex() {
	c.JSONWithStatus(200, map[string]string{"message": "index"})
}

func (c *TestController) PostCreate() {
	c.JSONWithStatus(201, map[string]string{"message": "created"})
}

func (c *TestController) GetProfile() {
	c.JSONWithStatus(200, map[string]string{"message": "profile"})
}

func TestRouterMethod(t *testing.T) {
	app := NewApp()
	controller := &TestController{}

	t.Run("正常路由注册", func(t *testing.T) {
		// 不应该panic
		assert.NotPanics(t, func() {
			app.Router("/test", controller,
				"GetIndex", "GET:/",
				"PostCreate", "POST:/create",
				"GetProfile", "GET:/profile",
			)
		})
	})

	t.Run("路由参数数量不匹配", func(t *testing.T) {
		assert.Panics(t, func() {
			app.Router("/test", controller,
				"GetIndex", "GET:/", // 缺少一个路由定义
				"PostCreate",
			)
		})
	})

	t.Run("无效路由模式", func(t *testing.T) {
		assert.Panics(t, func() {
			app.Router("/test", controller,
				"GetIndex", "invalid-pattern", // 缺少冒号分隔符
			)
		})
	})

	t.Run("方法不存在", func(t *testing.T) {
		assert.Panics(t, func() {
			app.Router("/test", controller,
				"NonExistentMethod", "GET:/nonexistent",
			)
		})
	})
}

func TestIncludeMethod(t *testing.T) {
	app := NewApp()
	controller := &TestController{}

	t.Run("自动注册控制器", func(t *testing.T) {
		assert.NotPanics(t, func() {
			app.Include(controller)
		})
	})

	t.Run("批量注册多个控制器", func(t *testing.T) {
		controller2 := &TestController{}
		assert.NotPanics(t, func() {
			app.Include(controller, controller2)
		})
	})
}

func TestRunMethod(t *testing.T) {
	app := NewApp()

	t.Run("设置默认地址", func(t *testing.T) {
		assert.Equal(t, ":8080", app.address)
	})

	t.Run("不带参数运行", func(t *testing.T) {
		// 这里不能实际调用Run，因为会阻塞测试
		// 只测试地址设置
		assert.NotNil(t, app.Hertz)
	})
}
