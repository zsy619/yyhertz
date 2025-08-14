package core

import (
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/zsy619/yyhertz/framework/mvc/context"
	"github.com/zsy619/yyhertz/framework/mvc/session"
)

func TestBaseController_DestroySession(t *testing.T) {
	// 创建测试控制器
	c := &app.RequestContext{}
	ctx := context.NewContext(c)
	
	controller := &BaseController{
		Ctx:           ctx,
		sessionHelper: session.NewManager(nil), // 使用默认配置
	}

	// 设置一些session数据
	controller.SetSession("key1", "value1")
	controller.SetSession("key2", "value2")
	controller.SetSession("key3", 123)

	// 验证数据已设置
	if controller.GetSession("key1") != "value1" {
		t.Error("Expected session key1 to be 'value1'")
	}
	if controller.GetSession("key2") != "value2" {
		t.Error("Expected session key2 to be 'value2'")
	}
	if controller.GetSession("key3") != 123 {
		t.Error("Expected session key3 to be 123")
	}

	// 调用DestroySession
	controller.DestroySession()

	// 验证所有数据已被清空
	if controller.GetSession("key1") != nil {
		t.Error("Expected session key1 to be nil after DestroySession")
	}
	if controller.GetSession("key2") != nil {
		t.Error("Expected session key2 to be nil after DestroySession")
	}
	if controller.GetSession("key3") != nil {
		t.Error("Expected session key3 to be nil after DestroySession")
	}

	// 验证session ID仍然存在（session存在但数据被清空）
	sessionID := controller.GetSessionID()
	if sessionID == "" {
		t.Error("Expected session ID to still exist after DestroySession")
	}
}

func TestBaseController_DestroySession_NilSession(t *testing.T) {
	// 创建没有session的控制器
	controller := &BaseController{
		Ctx: nil, // 模拟没有context的情况
	}

	// 调用DestroySession不应该崩溃
	controller.DestroySession()

	// 验证没有panic发生
	t.Log("DestroySession handled nil session gracefully")
}

func TestBaseController_DestroySession_AfterDelete(t *testing.T) {
	// 创建测试控制器
	c := &app.RequestContext{}
	ctx := context.NewContext(c)
	
	controller := &BaseController{
		Ctx:           ctx,
		sessionHelper: session.NewManager(nil), // 使用默认配置
	}

	// 设置session数据
	controller.SetSession("test1", "value1")
	controller.SetSession("test2", "value2")
	controller.SetSession("test3", "value3")

	// 先删除单个键
	controller.DeleteSession("test1")
	
	// 验证单个键被删除
	if controller.GetSession("test1") != nil {
		t.Error("Expected test1 to be deleted")
	}
	if controller.GetSession("test2") != "value2" {
		t.Error("Expected test2 to still exist")
	}

	// 调用DestroySession清空所有剩余数据
	controller.DestroySession()

	// 验证所有数据已被清空
	if controller.GetSession("test2") != nil {
		t.Error("Expected test2 to be nil after DestroySession")
	}
	if controller.GetSession("test3") != nil {
		t.Error("Expected test3 to be nil after DestroySession")
	}
}

func TestBaseController_DestroySession_MultipleTypes(t *testing.T) {
	// 创建测试控制器
	c := &app.RequestContext{}
	ctx := context.NewContext(c)
	
	controller := &BaseController{
		Ctx:           ctx,
		sessionHelper: session.NewManager(nil), // 使用默认配置
	}

	// 设置不同类型的session数据
	controller.SetSession("string", "test")
	controller.SetSession("int", 42)
	controller.SetSession("bool", true)
	controller.SetSession("slice", []string{"a", "b", "c"})
	controller.SetSession("map", map[string]int{"x": 1, "y": 2})

	// 验证数据已设置
	if controller.GetSession("string") != "test" {
		t.Error("String value not set correctly")
	}
	if controller.GetSession("int") != 42 {
		t.Error("Int value not set correctly")
	}
	if controller.GetSession("bool") != true {
		t.Error("Bool value not set correctly")
	}

	// 销毁所有session数据
	controller.DestroySession()

	// 验证所有不同类型的数据都被清空
	if controller.GetSession("string") != nil {
		t.Error("String value should be nil after DestroySession")
	}
	if controller.GetSession("int") != nil {
		t.Error("Int value should be nil after DestroySession")
	}
	if controller.GetSession("bool") != nil {
		t.Error("Bool value should be nil after DestroySession")
	}
	if controller.GetSession("slice") != nil {
		t.Error("Slice value should be nil after DestroySession")
	}
	if controller.GetSession("map") != nil {
		t.Error("Map value should be nil after DestroySession")
	}
}