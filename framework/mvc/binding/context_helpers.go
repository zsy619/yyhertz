package binding

import (
	"encoding/json"
	"mime/multipart"

	"github.com/zsy619/yyhertz/framework/mvc/context"
)

// ContextHelpers 上下文帮助函数，为binding模块提供Context接口适配
type ContextHelpers struct{}

// Query 获取查询参数
func (ch *ContextHelpers) Query(ctx *context.Context, key string) string {
	if ctx.Request != nil {
		return string(ctx.Request.QueryArgs().Peek(key))
	}
	return ""
}

// Param 获取路径参数  
func (ch *ContextHelpers) Param(ctx *context.Context, key string) string {
	return ctx.Params.ByName(key)
}

// FormValue 获取表单值
func (ch *ContextHelpers) FormValue(ctx *context.Context, key string) string {
	if ctx.Request != nil {
		return string(ctx.Request.PostArgs().Peek(key))
	}
	return ""
}

// GetRawData 获取原始请求体数据
func (ch *ContextHelpers) GetRawData(ctx *context.Context) ([]byte, error) {
	if ctx.Request != nil {
		body, err := ctx.Request.Body()
		return body, err
	}
	return nil, nil
}

// GetHeader 获取请求头
func (ch *ContextHelpers) GetHeader(ctx *context.Context, key string) string {
	return ctx.GetHeader(key)
}

// Cookie 获取Cookie
func (ch *ContextHelpers) Cookie(ctx *context.Context, key string) string {
	if ctx.Request != nil {
		return string(ctx.Request.Cookie(key))
	}
	return ""
}

// FormFile 获取上传文件
func (ch *ContextHelpers) FormFile(ctx *context.Context, key string) (*multipart.FileHeader, error) {
	if ctx.Request != nil {
		return ctx.Request.FormFile(key)
	}
	return nil, nil
}

// JSON 响应JSON数据
func (ch *ContextHelpers) JSON(ctx *context.Context, code int, data interface{}) {
	if ctx.Request != nil {
		ctx.Request.JSON(code, data)
	}
}

// ContextAdapter Context适配器，为binding提供统一接口
type ContextAdapter struct {
	*ContextHelpers
	ctx *context.Context
}

// NewContextAdapter 创建Context适配器
func NewContextAdapter(ctx *context.Context) *ContextAdapter {
	return &ContextAdapter{
		ContextHelpers: &ContextHelpers{},
		ctx:           ctx,
	}
}

// Query 获取查询参数
func (ca *ContextAdapter) Query(key string) string {
	return ca.ContextHelpers.Query(ca.ctx, key)
}

// FormValue 获取表单值
func (ca *ContextAdapter) FormValue(key string) string {
	return ca.ContextHelpers.FormValue(ca.ctx, key)
}

// GetRawData 获取原始请求体数据
func (ca *ContextAdapter) GetRawData() ([]byte, error) {
	return ca.ContextHelpers.GetRawData(ca.ctx)
}

// FormFile 获取上传文件
func (ca *ContextAdapter) FormFile(key string) (*multipart.FileHeader, error) {
	return ca.ContextHelpers.FormFile(ca.ctx, key)
}

// Param 获取路径参数
func (ca *ContextAdapter) Param(key string) string {
	return ca.ContextHelpers.Param(ca.ctx, key)
}

// Cookie 获取Cookie
func (ca *ContextAdapter) Cookie(key string) string {
	return ca.ContextHelpers.Cookie(ca.ctx, key)
}

// ShouldBindJSON 便利方法：从JSON绑定结构体
func (ca *ContextAdapter) ShouldBindJSON(target interface{}) error {
	data, err := ca.GetRawData()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}