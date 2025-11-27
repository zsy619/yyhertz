package core

import (
	"strings"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/stretchr/testify/assert"

	context "github.com/zsy619/yyhertz/framework/mvc/context"
)

// TestUser 测试用户结构体
type TestUser struct {
	Name  string `json:"name" yaml:"name" form:"name"`
	Email string `json:"email" yaml:"email" form:"email"`
	Age   int    `json:"age" yaml:"age" form:"age"`
}

// TestBindYAML 测试BindYAML方法
func TestBindYAML(t *testing.T) {
	// 创建测试用的Hertz RequestContext
	ctx := &app.RequestContext{}
	
	// 设置YAML请求体
	yamlData := `
name: John Doe
email: john@example.com
age: 30
`
	ctx.Request.SetBodyString(yamlData)
	ctx.Request.Header.Set("Content-Type", "application/yaml")
	
	// 创建BaseController
	controller := NewBaseController()
	enhancedCtx := context.NewContext(ctx)
	controller.Ctx = enhancedCtx
	
	// 测试BindYAML
	var user TestUser
	err := controller.BindYAML(&user)
	
	assert.NoError(t, err)
	assert.Equal(t, "John Doe", user.Name)
	assert.Equal(t, "john@example.com", user.Email)
	assert.Equal(t, 30, user.Age)
}

// TestBindForm 测试BindForm方法
func TestBindForm(t *testing.T) {
	// 创建测试用的Hertz RequestContext
	ctx := &app.RequestContext{}
	
	// 设置表单请求体
	formData := "name=John+Doe&email=john@example.com&age=30"
	ctx.Request.SetBodyString(formData)
	ctx.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	
	// 创建BaseController
	controller := NewBaseController()
	enhancedCtx := context.NewContext(ctx)
	controller.Ctx = enhancedCtx
	
	// 测试BindForm
	var user TestUser
	err := controller.BindForm(&user)
	
	assert.NoError(t, err)
	assert.Equal(t, "John Doe", user.Name)
	assert.Equal(t, "john@example.com", user.Email)
	assert.Equal(t, 30, user.Age)
}

// TestBindProtobuf 测试BindProtobuf方法
func TestBindProtobuf(t *testing.T) {
	// 创建测试用的Hertz RequestContext
	ctx := &app.RequestContext{}
	
	// 设置Protobuf请求体（这里用空数据作为示例）
	ctx.Request.SetBody([]byte{})
	ctx.Request.Header.Set("Content-Type", "application/x-protobuf")
	
	// 创建BaseController
	controller := NewBaseController()
	enhancedCtx := context.NewContext(ctx)
	controller.Ctx = enhancedCtx
	
	// 测试BindProtobuf（由于没有实际的protobuf消息，这里只测试方法存在性）
	var data interface{}
	err := controller.BindProtobuf(&data)
	
	// 预期会有错误，因为data不实现proto.Message接口
	assert.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "proto.message")
}

// TestBindMethodsExist 测试所有Bind方法都存在
func TestBindMethodsExist(t *testing.T) {
	controller := NewBaseController()
	
	// 检查方法是否存在（通过反射）
	assert.NotNil(t, controller.Bind)
	assert.NotNil(t, controller.BindJSON)
	assert.NotNil(t, controller.BindYAML)
	assert.NotNil(t, controller.BindForm)
	assert.NotNil(t, controller.BindProtobuf)
}

// TestBindWithNilContext 测试空Context的处理
func TestBindWithNilContext(t *testing.T) {
	controller := NewBaseController()
	controller.Ctx = nil
	
	var user TestUser
	
	// 测试所有Bind方法都能正确处理nil context
	err1 := controller.BindYAML(&user)
	assert.Error(t, err1)
	assert.Contains(t, err1.Error(), "context is nil")
	
	err2 := controller.BindForm(&user)
	assert.Error(t, err2)
	assert.Contains(t, err2.Error(), "context is nil")
	
	err3 := controller.BindProtobuf(&user)
	assert.Error(t, err3)
	assert.Contains(t, err3.Error(), "context is nil")
}