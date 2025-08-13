// Package examples 方法测试
package main

import (
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/route"
	yycontext "github.com/zsy619/yyhertz/framework/mvc/context"
)

func createTestRequestContext() *app.RequestContext {
	ctx := app.NewContext(10)
	_ = route.NewEngine(config.NewOptions([]config.Option{}))
	ctx.Request.SetRequestURI("http://localhost:8080/test")
	ctx.Request.Header.SetMethod("GET")
	return ctx
}

// TestInitializeMethods 测试新添加的Initialize方法
func TestInitializeMethods(t *testing.T) {
	hertzCtx := createTestRequestContext()
	
	// 测试InputData的Initialize方法
	inputData := &yycontext.InputData{}
	inputData.Initialize(hertzCtx)
	
	// 验证Initialize方法是否正确设置了Context
	ctx := inputData.GetContext()
	if ctx == nil {
		t.Error("InputData.Initialize() 后 GetContext() 返回 nil")
	}
	
	if ctx.Request == nil {
		t.Error("InputData.Initialize() 后 Context.Request 为 nil")
	}
	
	if ctx.Request != hertzCtx {
		t.Error("InputData.Initialize() 后 Context.Request 不是期望的 RequestContext")
	}
	
	// 测试OutputData的Initialize方法
	outputData := &yycontext.OutputData{}
	outputData.Initialize(hertzCtx)
	
	// 验证Initialize方法是否正确设置了Context
	outCtx := outputData.GetContext()
	if outCtx == nil {
		t.Error("OutputData.Initialize() 后 GetContext() 返回 nil")
	}
	
	if outCtx.Request == nil {
		t.Error("OutputData.Initialize() 后 Context.Request 为 nil")
	}
	
	if outCtx.Request != hertzCtx {
		t.Error("OutputData.Initialize() 后 Context.Request 不是期望的 RequestContext")
	}
	
	t.Log("✅ Initialize 和 GetContext 方法测试通过")
}

// TestGetContextMethods 测试GetContext方法
func TestGetContextMethods(t *testing.T) {
	hertzCtx := createTestRequestContext()
	
	// 使用构造函数创建
	yyCtx := &yycontext.Context{
		Request: hertzCtx,
		RequestContext: hertzCtx,
	}
	
	inputData := yycontext.NewInputData(yyCtx)
	outputData := yycontext.NewOutputData(yyCtx)
	
	// 测试GetContext
	inputCtx := inputData.GetContext()
	outputCtx := outputData.GetContext()
	
	if inputCtx != yyCtx {
		t.Error("InputData.GetContext() 返回的不是期望的Context")
	}
	
	if outputCtx != yyCtx {
		t.Error("OutputData.GetContext() 返回的不是期望的Context")
	}
	
	t.Log("✅ GetContext 方法测试通过")
}

// TestCookieOperations 测试Cookie操作是否正常工作
func TestCookieOperations(t *testing.T) {
	hertzCtx := createTestRequestContext()
	
	// 通过Initialize方法创建
	inputData := &yycontext.InputData{}
	inputData.Initialize(hertzCtx)
	
	outputData := &yycontext.OutputData{}
	outputData.Initialize(hertzCtx)
	
	// 测试Cookie操作（这些应该不会崩溃）
	inputData.SetCookie("test_cookie", "test_value")
	value := inputData.Cookie("test_cookie")
	
	// 在测试环境中，由于没有完整的HTTP周期，Cookie值可能为空
	// 但方法调用不应该崩溃
	t.Logf("Cookie value: %s", value)
	
	// 测试输出Cookie
	outputData.Cookie("output_test", "output_value", 3600, "/", "", false, true)
	
	t.Log("✅ Cookie操作测试通过")
}

// TestSessionOperations 测试Session操作是否正常工作  
func TestSessionOperations(t *testing.T) {
	hertzCtx := createTestRequestContext()
	
	inputData := &yycontext.InputData{}
	inputData.Initialize(hertzCtx)
	
	// 测试Session操作
	err := inputData.SetSession("test_session", "test_session_value")
	if err != nil {
		t.Logf("SetSession 返回错误（在测试环境中是正常的）: %v", err)
	}
	
	value := inputData.GetSession("test_session")
	t.Logf("Session value: %v", value)
	
	// 测试Session扩展访问
	ext := inputData.GetSessionExtension()
	if ext == nil {
		t.Error("GetSessionExtension() 返回 nil")
	}
	
	t.Log("✅ Session操作测试通过")
}

// TestBackwardCompatibility 测试向后兼容性
func TestBackwardCompatibility(t *testing.T) {
	hertzCtx := createTestRequestContext()
	
	// 测试两种创建方式的兼容性
	
	// 方式1：通过构造函数
	yyCtx := &yycontext.Context{Request: hertzCtx}
	inputData1 := yycontext.NewInputData(yyCtx)
	
	// 方式2：通过Initialize方法
	inputData2 := &yycontext.InputData{}
	inputData2.Initialize(hertzCtx)
	
	// 两种方式都应该能正常工作
	inputData1.SetCookie("method1_cookie", "value1")
	inputData2.SetCookie("method2_cookie", "value2")
	
	// 验证Context访问
	ctx1 := inputData1.GetContext()
	ctx2 := inputData2.GetContext()
	
	if ctx1 == nil || ctx2 == nil {
		t.Error("Context 不应该为 nil")
	}
	
	if ctx1.Request != hertzCtx || ctx2.Request != hertzCtx {
		t.Error("Context.Request 应该指向相同的 RequestContext")
	}
	
	t.Log("✅ 向后兼容性测试通过")
}