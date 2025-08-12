package context

import (
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/zsy619/yyhertz/framework/mvc/session"
)

func TestInputDataSessionBasic(t *testing.T) {
	// 创建测试上下文
	ctx := &Context{
		Keys: make(map[string]interface{}),
	}
	
	// 模拟Request上下文
	mockRequest := &app.RequestContext{}
	ctx.Request = mockRequest
	ctx.RequestContext = mockRequest
	
	// 初始化InputData
	ctx.Input = &InputData{ctx: ctx}
	
	// 测试启动session
	sessionStore := ctx.Input.StartSession()
	if sessionStore == nil {
		t.Fatal("Expected sessionStore to be not nil")
	}
	
	if !ctx.Input.IsSessionStarted() {
		t.Fatal("Expected session to be started")
	}
	
	sessionID := ctx.Input.GetSessionID()
	if sessionID == "" {
		t.Fatal("Expected sessionID to be not empty")
	}
	
	// 测试设置和获取session数据
	err := ctx.Input.SetSession("username", "test_user")
	if err != nil {
		t.Fatalf("SetSession error: %v", err)
	}
	
	username := ctx.Input.GetSession("username")
	if username != "test_user" {
		t.Fatalf("Expected 'test_user', got %v", username)
	}
	
	// 测试删除session数据
	ctx.Input.SetSession("temp_key", "temp_value")
	ctx.Input.DelSession("temp_key")
	
	deleted := ctx.Input.GetSession("temp_key")
	if deleted != nil {
		t.Fatalf("Expected nil after deletion, got %v", deleted)
	}
	
	t.Logf("✅ Basic session functionality works correctly")
}

func TestSessionStoreBasic(t *testing.T) {
	// 创建YYHertz原生session store
	yyStore := session.NewMemoryStore("test_session_id")
	
	// 创建测试上下文
	testCtx := &Context{
		Keys: make(map[string]interface{}),
	}
	
	// 创建session适配器
	sessionStore := NewSessionStore(yyStore, testCtx)
	
	// 测试Set/Get方法
	err := sessionStore.Set(testCtx, "test_key", "test_value")
	if err != nil {
		t.Fatalf("Set error: %v", err)
	}
	
	value := sessionStore.Get(testCtx, "test_key")
	if value != "test_value" {
		t.Fatalf("Expected 'test_value', got %v", value)
	}
	
	// 测试SessionID
	sessionID := sessionStore.SessionID(testCtx)
	if sessionID != "test_session_id" {
		t.Fatalf("Expected 'test_session_id', got %v", sessionID)
	}
	
	// 测试Delete
	sessionStore.Delete(testCtx, "test_key")
	deletedValue := sessionStore.Get(testCtx, "test_key")
	if deletedValue != nil {
		t.Fatalf("Expected nil after delete, got %v", deletedValue)
	}
	
	// 测试SessionRelease的增强功能
	sessionStore.Set(testCtx, "release_test", "release_value")
	sessionStore.SessionRelease(testCtx, nil)
	
	// 验证SessionRelease设置的标记
	if released, exists := testCtx.Keys["session_released"]; !exists || released != true {
		t.Fatalf("Expected session_released flag to be set to true")
	}
	
	t.Logf("✅ session adapter works correctly")
}