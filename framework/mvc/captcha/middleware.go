package captcha

import (
	"context"
	"net/http"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
)

// Middleware 验证码中间件
type Middleware struct {
	generator *Generator
	config    *MiddlewareConfig
}

// MiddlewareConfig 中间件配置
type MiddlewareConfig struct {
	// SkipPaths 跳过验证的路径
	SkipPaths []string
	// ErrorHandler 错误处理函数
	ErrorHandler func(c context.Context, ctx *app.RequestContext, err error)
	// SuccessHandler 成功处理函数
	SuccessHandler func(c context.Context, ctx *app.RequestContext)
}

// NewMiddleware 创建验证码中间件
func NewMiddleware(generator *Generator, config *MiddlewareConfig) *Middleware {
	if config == nil {
		config = &MiddlewareConfig{}
	}
	
	// 设置默认错误处理
	if config.ErrorHandler == nil {
		config.ErrorHandler = defaultErrorHandler
	}
	
	// 设置默认成功处理
	if config.SuccessHandler == nil {
		config.SuccessHandler = defaultSuccessHandler
	}
	
	return &Middleware{
		generator: generator,
		config:    config,
	}
}

// Handler 中间件处理函数
func (m *Middleware) Handler() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		// 检查是否跳过验证
		if m.shouldSkip(string(ctx.Path())) {
			ctx.Next(c)
			return
		}
		
		// 验证验证码
		if err := m.verifyCaptcha(ctx); err != nil {
			m.config.ErrorHandler(c, ctx, err)
			ctx.Abort()
			return
		}
		
		m.config.SuccessHandler(c, ctx)
		ctx.Next(c)
	}
}

// shouldSkip 检查是否应该跳过验证
func (m *Middleware) shouldSkip(path string) bool {
	for _, skipPath := range m.config.SkipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// verifyCaptcha 验证验证码
func (m *Middleware) verifyCaptcha(ctx *app.RequestContext) error {
	captchaID := string(ctx.PostForm("captcha_id"))
	captchaCode := string(ctx.PostForm("captcha_code"))
	
	if captchaID == "" || captchaCode == "" {
		return &CaptchaError{
			Code:    ErrCodeMissingParams,
			Message: "验证码ID或验证码不能为空",
		}
	}
	
	if !m.generator.Verify(captchaID, captchaCode) {
		return &CaptchaError{
			Code:    ErrCodeInvalidCaptcha,
			Message: "验证码错误或已过期",
		}
	}
	
	return nil
}

// defaultErrorHandler 默认错误处理
func defaultErrorHandler(c context.Context, ctx *app.RequestContext, err error) {
	if captchaErr, ok := err.(*CaptchaError); ok {
		ctx.JSON(http.StatusBadRequest, utils.H{
			"code":    captchaErr.Code,
			"message": captchaErr.Message,
		})
	} else {
		ctx.JSON(http.StatusInternalServerError, utils.H{
			"code":    ErrCodeInternal,
			"message": "内部错误",
		})
	}
}

// defaultSuccessHandler 默认成功处理
func defaultSuccessHandler(c context.Context, ctx *app.RequestContext) {
	// 验证成功，不做额外处理
}

// CaptchaError 验证码错误
type CaptchaError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *CaptchaError) Error() string {
	return e.Message
}

// 错误码定义
const (
	ErrCodeMissingParams   = 4001 // 缺少参数
	ErrCodeInvalidCaptcha  = 4002 // 验证码错误
	ErrCodeExpiredCaptcha  = 4003 // 验证码过期
	ErrCodeInternal        = 5001 // 内部错误
)

// HandlerFunc 处理函数类型
type HandlerFunc func(c context.Context, ctx *app.RequestContext, generator *Generator)

// GenerateHandler 生成验证码的处理函数
func GenerateHandler(generator *Generator) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		captcha, err := generator.Generate()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, utils.H{
				"code":    ErrCodeInternal,
				"message": "生成验证码失败",
			})
			return
		}
		
		ctx.JSON(http.StatusOK, utils.H{
			"code":       0,
			"message":    "success",
			"captcha_id": captcha.ID,
		})
	}
}

// ImageHandler 获取验证码图片的处理函数
func ImageHandler(generator *Generator) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		id := ctx.Param("id")
		if id == "" {
			ctx.JSON(http.StatusBadRequest, utils.H{
				"code":    ErrCodeMissingParams,
				"message": "验证码ID不能为空",
			})
			return
		}
		
		imageBytes, err := generator.GetImage(id)
		if err != nil {
			ctx.JSON(http.StatusNotFound, utils.H{
				"code":    ErrCodeExpiredCaptcha,
				"message": "验证码不存在或已过期",
			})
			return
		}
		
		ctx.Header("Content-Type", "image/png")
		ctx.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		ctx.Header("Pragma", "no-cache")
		ctx.Header("Expires", "0")
		ctx.Write(imageBytes)
	}
}

// VerifyHandler 验证验证码的处理函数
func VerifyHandler(generator *Generator) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		captchaID := string(ctx.PostForm("captcha_id"))
		captchaCode := string(ctx.PostForm("captcha_code"))
		
		if captchaID == "" || captchaCode == "" {
			ctx.JSON(http.StatusBadRequest, utils.H{
				"code":    ErrCodeMissingParams,
				"message": "验证码ID或验证码不能为空",
			})
			return
		}
		
		if !generator.Verify(captchaID, captchaCode) {
			ctx.JSON(http.StatusBadRequest, utils.H{
				"code":    ErrCodeInvalidCaptcha,
				"message": "验证码错误或已过期",
			})
			return
		}
		
		ctx.JSON(http.StatusOK, utils.H{
			"code":    0,
			"message": "验证成功",
		})
	}
}
