package middleware

import (	
	mvccontext "github.com/zsy619/yyhertz/framework/mvc/context"
)

// ContextAdapter Context适配器，实现两套系统Context的转换和统一接口

// UnifiedContext 统一的Context接口
type UnifiedContext interface {
	// 基础操作
	Next()
	Abort()
	IsAborted() bool
	
	// 数据存取
	Set(key string, value any)
	Get(key string) (any, bool)
	GetString(key string) string
	GetBool(key string) bool
	GetInt(key string) int
	GetInt64(key string) int64
	GetFloat64(key string) float64
	
	// 错误处理
	AddError(err error) error
	GetErrors() []error
	HasErrors() bool
	LastError() error
	
	// HTTP相关
	Header(key, value string)
	JSON(code int, obj any)
	SetStatusCode(code int)
	ClientIP() string
	Method() []byte
	Path() []byte
	
	// 状态控制
	AbortWithStatus(code int)
	AbortWithError(code int, err error) error
}

// MVCContextWrapper MVC Context包装器
type MVCContextWrapper struct {
	ctx *mvccontext.EnhancedContext
}

// NewMVCContextWrapper 创建MVC Context包装器
func NewMVCContextWrapper(ctx *mvccontext.EnhancedContext) *MVCContextWrapper {
	return &MVCContextWrapper{ctx: ctx}
}

// 实现UnifiedContext接口
func (w *MVCContextWrapper) Next() {
	w.ctx.Next()
}

func (w *MVCContextWrapper) Abort() {
	w.ctx.Abort()
}

func (w *MVCContextWrapper) IsAborted() bool {
	return w.ctx.IsAborted()
}

func (w *MVCContextWrapper) Set(key string, value any) {
	w.ctx.Set(key, value)
}

func (w *MVCContextWrapper) Get(key string) (any, bool) {
	return w.ctx.Get(key)
}

func (w *MVCContextWrapper) GetString(key string) string {
	if val, exists := w.ctx.Get(key); exists {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func (w *MVCContextWrapper) GetBool(key string) bool {
	if val, exists := w.ctx.Get(key); exists {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

func (w *MVCContextWrapper) GetInt(key string) int {
	if val, exists := w.ctx.Get(key); exists {
		if i, ok := val.(int); ok {
			return i
		}
	}
	return 0
}

func (w *MVCContextWrapper) GetInt64(key string) int64 {
	if val, exists := w.ctx.Get(key); exists {
		if i, ok := val.(int64); ok {
			return i
		}
	}
	return 0
}

func (w *MVCContextWrapper) GetFloat64(key string) float64 {
	if val, exists := w.ctx.Get(key); exists {
		if f, ok := val.(float64); ok {
			return f
		}
	}
	return 0.0
}

func (w *MVCContextWrapper) AddError(err error) error {
	w.ctx.AddError(err)
	return err
}

func (w *MVCContextWrapper) GetErrors() []error {
	return w.ctx.GetErrors()
}

func (w *MVCContextWrapper) HasErrors() bool {
	return len(w.ctx.GetErrors()) > 0
}

func (w *MVCContextWrapper) LastError() error {
	errors := w.ctx.GetErrors()
	if len(errors) == 0 {
		return nil
	}
	return errors[len(errors)-1]
}

func (w *MVCContextWrapper) Header(key, value string) {
	w.ctx.Request.Response.Header.Set(key, value)
}

func (w *MVCContextWrapper) JSON(code int, obj any) {
	w.ctx.JSON(code, obj)
}

func (w *MVCContextWrapper) SetStatusCode(code int) {
	w.ctx.Request.SetStatusCode(code)
}

func (w *MVCContextWrapper) ClientIP() string {
	return string(w.ctx.Request.ClientIP())
}

func (w *MVCContextWrapper) Method() []byte {
	return w.ctx.Request.Method()
}

func (w *MVCContextWrapper) Path() []byte {
	return w.ctx.Request.Path()
}

func (w *MVCContextWrapper) AbortWithStatus(code int) {
	w.ctx.JSON(code, map[string]string{"error": "Request aborted"})
	w.ctx.Abort()
}

func (w *MVCContextWrapper) AbortWithError(code int, err error) error {
	w.SetStatusCode(code)
	w.Abort()
	return w.AddError(err)
}

// GetMVCContext 获取原始MVC Context
func (w *MVCContextWrapper) GetMVCContext() *mvccontext.EnhancedContext {
	return w.ctx
}

// BasicContextWrapper 基础Context包装器
type BasicContextWrapper struct {
	ctx *Context
}

// NewBasicContextWrapper 创建基础Context包装器
func NewBasicContextWrapper(ctx *Context) *BasicContextWrapper {
	return &BasicContextWrapper{ctx: ctx}
}

// 实现UnifiedContext接口
func (w *BasicContextWrapper) Next() {
	w.ctx.Next()
}

func (w *BasicContextWrapper) Abort() {
	w.ctx.Abort()
}

func (w *BasicContextWrapper) IsAborted() bool {
	return w.ctx.IsAborted()
}

func (w *BasicContextWrapper) Set(key string, value any) {
	w.ctx.Set(key, value)
}

func (w *BasicContextWrapper) Get(key string) (any, bool) {
	return w.ctx.Get(key)
}

func (w *BasicContextWrapper) GetString(key string) string {
	return w.ctx.GetString(key)
}

func (w *BasicContextWrapper) GetBool(key string) bool {
	return w.ctx.GetBool(key)
}

func (w *BasicContextWrapper) GetInt(key string) int {
	return w.ctx.GetInt(key)
}

func (w *BasicContextWrapper) GetInt64(key string) int64 {
	return w.ctx.GetInt64(key)
}

func (w *BasicContextWrapper) GetFloat64(key string) float64 {
	return w.ctx.GetFloat64(key)
}

func (w *BasicContextWrapper) AddError(err error) error {
	return w.ctx.AddError(err)
}

func (w *BasicContextWrapper) GetErrors() []error {
	return w.ctx.GetErrors()
}

func (w *BasicContextWrapper) HasErrors() bool {
	return w.ctx.HasErrors()
}

func (w *BasicContextWrapper) LastError() error {
	return w.ctx.LastError()
}

func (w *BasicContextWrapper) Header(key, value string) {
	w.ctx.Header(key, value)
}

func (w *BasicContextWrapper) JSON(code int, obj any) {
	w.ctx.JSON(code, obj)
}

func (w *BasicContextWrapper) SetStatusCode(code int) {
	w.ctx.SetStatusCode(code)
}

func (w *BasicContextWrapper) ClientIP() string {
	return w.ctx.ClientIP()
}

func (w *BasicContextWrapper) Method() []byte {
	return w.ctx.Method()
}

func (w *BasicContextWrapper) Path() []byte {
	return w.ctx.URI().Path()
}

func (w *BasicContextWrapper) AbortWithStatus(code int) {
	w.ctx.AbortWithStatus(code)
}

func (w *BasicContextWrapper) AbortWithError(code int, err error) error {
	return w.ctx.AbortWithError(code, err)
}

// GetBasicContext 获取原始基础Context
func (w *BasicContextWrapper) GetBasicContext() *Context {
	return w.ctx
}

// ContextConverter Context转换器
type ContextConverter struct{}

// NewContextConverter 创建Context转换器
func NewContextConverter() *ContextConverter {
	return &ContextConverter{}
}

// ToUnified 转换为统一Context接口
func (c *ContextConverter) ToUnified(ctx interface{}) UnifiedContext {
	switch v := ctx.(type) {
	case *mvccontext.EnhancedContext:
		return NewMVCContextWrapper(v)
	case *Context:
		return NewBasicContextWrapper(v)
	default:
		return nil
	}
}

// MVCToBasicContext MVC Context转换为基础Context
func (c *ContextConverter) MVCToBasicContext(mvcCtx *mvccontext.EnhancedContext) *Context {
	return CreateBasicContext(mvcCtx)
}

// BasicToMVCContext 基础Context转换为MVC Context
func (c *ContextConverter) BasicToMVCContext(basicCtx *Context) *mvccontext.EnhancedContext {
	// 获取底层的Hertz RequestContext
	hertzCtx := basicCtx.RequestContext
	
	// 创建MVC增强上下文
	mvcCtx := mvccontext.NewContext(hertzCtx)
	
	// 同步数据
	for key, value := range basicCtx.Keys {
		mvcCtx.Set(key, value)
	}
	
	// 同步错误
	for _, err := range basicCtx.Errors {
		mvcCtx.AddError(err)
	}
	
	// 同步状态
	if basicCtx.IsAborted() {
		mvcCtx.Abort()
	}
	
	return mvcCtx
}

// CreateCompatibleHandler 创建兼容的处理器
func (c *ContextConverter) CreateCompatibleHandler(handler interface{}) interface{} {
	switch h := handler.(type) {
	case MiddlewareFunc:
		// MVC中间件转换为基础中间件
		return func(ctx *Context) {
			mvcCtx := c.BasicToMVCContext(ctx)
			h(mvcCtx)
			// 同步状态回基础Context
			SyncContextState(ctx, mvcCtx)
		}
	case HandlerFunc:
		// 基础中间件转换为MVC中间件
		return func(mvcCtx *mvccontext.EnhancedContext) {
			basicCtx := c.MVCToBasicContext(mvcCtx)
			h(basicCtx)
			// 同步状态回MVC Context
			SyncContextState(basicCtx, mvcCtx)
		}
	default:
		return nil
	}
}

// 全局转换器实例
var globalConverter = NewContextConverter()

// GetGlobalConverter 获取全局转换器
func GetGlobalConverter() *ContextConverter {
	return globalConverter
}

// ToUnified 便捷函数 - 转换为统一Context
func ToUnified(ctx interface{}) UnifiedContext {
	return globalConverter.ToUnified(ctx)
}

// CreateCompatible 便捷函数 - 创建兼容处理器
func CreateCompatible(handler interface{}) interface{} {
	return globalConverter.CreateCompatibleHandler(handler)
}