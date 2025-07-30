package util

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/zsy619/yyhertz/framework/config"
	"github.com/zsy619/yyhertz/framework/types"
)

// HTTPGenericRequest 通用HTTP请求结构
type HTTPGenericRequest[T any] struct {
	Data      T             `json:"data"`
	RequestID string        `json:"request_id,omitempty"`
	UserID    string        `json:"user_id,omitempty"`
	ClientIP  string        `json:"client_ip,omitempty"`
	UserAgent string        `json:"user_agent,omitempty"`
	Headers   map[string]string `json:"headers,omitempty"`
	Meta      map[string]any    `json:"meta,omitempty"`
}

// HTTPGenericResponse 通用HTTP响应结构
type HTTPGenericResponse[T any] struct {
	Data      T             `json:"data"`
	Success   bool          `json:"success"`
	Message   string        `json:"message"`
	Code      int           `json:"code"`
	RequestID string        `json:"request_id,omitempty"`
	Timestamp int64         `json:"timestamp"`
	Meta      map[string]any `json:"meta,omitempty"`
}

// HTTPGenericHandler HTTP泛型处理器接口
type HTTPGenericHandler[T any, R any] interface {
	// HandleHTTP 处理HTTP请求
	HandleHTTP(ctx *app.RequestContext) (*HTTPGenericResponse[R], error)
	
	// ParseRequest 解析请求
	ParseRequest(ctx *app.RequestContext) (*HTTPGenericRequest[T], error)
	
	// BuildResponse 构建响应
	BuildResponse(result *GenericResult[R], requestID string) *HTTPGenericResponse[R]
	
	// HandleError 处理错误
	HandleError(ctx *app.RequestContext, err error, requestID string)
}

// BaseHTTPGenericHandler 基础HTTP泛型处理器
type BaseHTTPGenericHandler[T any, R any] struct {
	*BaseGenericHandler[T, R]
	
	// HTTP特定配置
	RequireAuth   bool
	RateLimit     *RateLimitConfig
	CacheConfig   *CacheConfig
	ValidateJSON  bool
	LogRequest    bool
	LogResponse   bool
}

// RateLimitConfig 速率限制配置
type RateLimitConfig struct {
	MaxRequests int
	Duration    time.Duration
	KeyFunc     func(*app.RequestContext) string
}

// CacheConfig 缓存配置
type CacheConfig struct {
	TTL     time.Duration
	KeyFunc func(any) string
}

// NewBaseHTTPGenericHandler 创建基础HTTP泛型处理器
func NewBaseHTTPGenericHandler[T any, R any](name string) *BaseHTTPGenericHandler[T, R] {
	return &BaseHTTPGenericHandler[T, R]{
		BaseGenericHandler: NewBaseGenericHandler[T, R](name),
		ValidateJSON:       true,
		LogRequest:         true,
		LogResponse:        true,
	}
}

// WithAuth 设置认证要求
func (h *BaseHTTPGenericHandler[T, R]) WithAuth(require bool) *BaseHTTPGenericHandler[T, R] {
	h.RequireAuth = require
	return h
}

// WithRateLimit 设置速率限制
func (h *BaseHTTPGenericHandler[T, R]) WithRateLimit(maxRequests int, duration time.Duration) *BaseHTTPGenericHandler[T, R] {
	h.RateLimit = &RateLimitConfig{
		MaxRequests: maxRequests,
		Duration:    duration,
		KeyFunc: func(ctx *app.RequestContext) string {
			return ctx.ClientIP()
		},
	}
	return h
}

// WithCache 设置缓存
func (h *BaseHTTPGenericHandler[T, R]) WithCache(ttl time.Duration) *BaseHTTPGenericHandler[T, R] {
	h.CacheConfig = &CacheConfig{
		TTL: ttl,
		KeyFunc: func(req any) string {
			if httpReq, ok := req.(*HTTPGenericRequest[T]); ok {
				data, _ := json.Marshal(httpReq.Data)
				return fmt.Sprintf("%s:%s", h.Name, string(data))
			}
			return fmt.Sprintf("%s:unknown", h.Name)
		},
	}
	return h
}

// HandleHTTP 实现HTTP请求处理
func (h *BaseHTTPGenericHandler[T, R]) HandleHTTP(ctx *app.RequestContext) (*HTTPGenericResponse[R], error) {
	start := time.Now()
	requestID := string(ctx.GetHeader("X-Request-ID"))
	if requestID == "" {
		requestID = ShortID()
		ctx.Set("request_id", requestID)
	}

	// 记录请求开始
	if h.LogRequest {
		config.WithFields(map[string]any{
			"handler":    h.Name,
			"method":     string(ctx.Method()),
			"path":       string(ctx.Path()),
			"client_ip":  ctx.ClientIP(),
			"user_agent": string(ctx.UserAgent()),
			"request_id": requestID,
		}).Info("HTTP generic handler processing started")
	}

	// 1. 认证检查
	if h.RequireAuth {
		if err := h.checkAuth(ctx); err != nil {
			h.HandleError(ctx, err, requestID)
			return nil, err
		}
	}

	// 2. 速率限制检查
	if h.RateLimit != nil {
		if err := h.checkRateLimit(ctx); err != nil {
			h.HandleError(ctx, err, requestID)
			return nil, err
		}
	}

	// 3. 解析请求
	request, err := h.ParseRequest(ctx)
	if err != nil {
		h.HandleError(ctx, err, requestID)
		return nil, err
	}
	request.RequestID = requestID

	// 4. 缓存检查
	if h.CacheConfig != nil {
		if cached, found := h.checkCache(request); found {
			response := h.BuildResponse(cached, requestID)
			if h.LogResponse {
				config.WithFields(map[string]any{
					"handler":    h.Name,
					"request_id": requestID,
					"duration":   time.Since(start).String(),
					"cached":     true,
				}).Info("HTTP generic handler returned cached result")
			}
			return response, nil
		}
	}

	// 5. 核心处理
	result, err := h.BaseGenericHandler.Handle(context.Background(), request.Data)
	if err != nil {
		h.HandleError(ctx, err, requestID)
		return nil, err
	}

	// 6. 缓存结果
	if h.CacheConfig != nil {
		h.setCache(request, result)
	}

	// 7. 构建响应
	response := h.BuildResponse(result, requestID)
	
	// 记录处理完成
	if h.LogResponse {
		duration := time.Since(start)
		config.WithFields(map[string]any{
			"handler":     h.Name,
			"request_id":  requestID,
			"duration":    duration.String(),
			"duration_ms": duration.Milliseconds(),
			"success":     result.Success,
			"code":        result.Code,
		}).Info("HTTP generic handler processing completed")
	}

	return response, nil
}

// ParseRequest 解析HTTP请求
func (h *BaseHTTPGenericHandler[T, R]) ParseRequest(ctx *app.RequestContext) (*HTTPGenericRequest[T], error) {
	request := &HTTPGenericRequest[T]{
		ClientIP:  ctx.ClientIP(),
		UserAgent: string(ctx.UserAgent()),
		Headers:   make(map[string]string),
		Meta:      make(map[string]any),
	}

	// 提取用户ID（如果存在）
	if userID := ctx.GetString("user_id"); userID != "" {
		request.UserID = userID
	}

	// 提取常用请求头
	commonHeaders := []string{"Content-Type", "Authorization", "Accept", "Accept-Language"}
	for _, header := range commonHeaders {
		if value := string(ctx.GetHeader(header)); value != "" {
			request.Headers[header] = value
		}
	}

	// 根据请求方法解析数据
	method := string(ctx.Method())
	switch method {
	case "GET", "DELETE":
		// 从查询参数解析
		if err := h.parseQueryParams(ctx, &request.Data); err != nil {
			return nil, fmt.Errorf("failed to parse query params: %w", err)
		}
	case "POST", "PUT", "PATCH":
		// 从请求体解析
		if h.ValidateJSON {
			if err := ctx.BindJSON(&request.Data); err != nil {
				return nil, fmt.Errorf("failed to bind JSON: %w", err)
			}
		} else {
			// 尝试多种格式
			if err := h.parseRequestBody(ctx, &request.Data); err != nil {
				return nil, fmt.Errorf("failed to parse request body: %w", err)
			}
		}
	}

	return request, nil
}

// parseQueryParams 解析查询参数到结构体
func (h *BaseHTTPGenericHandler[T, R]) parseQueryParams(ctx *app.RequestContext, data *T) error {
	// 使用反射将查询参数映射到结构体
	v := reflect.ValueOf(data).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)
		
		// 跳过未导出的字段
		if !field.CanSet() {
			continue
		}

		// 获取字段名（支持json tag）
		fieldName := fieldType.Name
		if jsonTag := fieldType.Tag.Get("json"); jsonTag != "" && jsonTag != "-" {
			if commaIdx := fmt.Sprintf("%s", jsonTag); commaIdx != "" {
				fieldName = jsonTag[:len(jsonTag)]
			}
		}

		// 从查询参数获取值
		paramValue := string(ctx.Query(fieldName))
		if paramValue == "" {
			continue
		}

		// 根据字段类型设置值
		if err := h.setFieldValue(field, paramValue); err != nil {
			return fmt.Errorf("failed to set field %s: %w", fieldName, err)
		}
	}

	return nil
}

// parseRequestBody 解析请求体
func (h *BaseHTTPGenericHandler[T, R]) parseRequestBody(ctx *app.RequestContext, data *T) error {
	contentType := string(ctx.GetHeader("Content-Type"))
	
	switch {
	case contentType == "application/json" || contentType == "":
		return ctx.BindJSON(data)
	case contentType == "application/x-www-form-urlencoded":
		return ctx.BindForm(data)
	case contentType == "multipart/form-data":
		return ctx.BindForm(data) // Hertz使用BindForm处理multipart
	default:
		// 尝试JSON解析
		return ctx.BindJSON(data)
	}
}

// setFieldValue 设置字段值
func (h *BaseHTTPGenericHandler[T, R]) setFieldValue(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(intVal)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(uintVal)
	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		field.SetFloat(floatVal)
	case reflect.Bool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(boolVal)
	default:
		return fmt.Errorf("unsupported field type: %s", field.Kind())
	}
	return nil
}

// BuildResponse 构建HTTP响应
func (h *BaseHTTPGenericHandler[T, R]) BuildResponse(result *GenericResult[R], requestID string) *HTTPGenericResponse[R] {
	return &HTTPGenericResponse[R]{
		Data:      result.Data,
		Success:   result.Success,
		Message:   result.Message,
		Code:      result.Code,
		RequestID: requestID,
		Timestamp: result.Timestamp,
		Meta:      result.Meta,
	}
}

// HandleError 处理错误
func (h *BaseHTTPGenericHandler[T, R]) HandleError(ctx *app.RequestContext, err error, requestID string) {
	config.WithFields(map[string]any{
		"handler":    h.Name,
		"request_id": requestID,
		"error":      err.Error(),
		"path":       string(ctx.Path()),
		"method":     string(ctx.Method()),
		"client_ip":  ctx.ClientIP(),
	}).Error("HTTP generic handler error")

	// 根据错误类型设置HTTP状态码
	statusCode := 500
	message := "Internal Server Error"
	
	if errNo, ok := err.(*types.ErrNo); ok {
		statusCode = int(errNo.ErrCode)
		message = errNo.ErrMsg
	}

	response := &HTTPGenericResponse[R]{
		Success:   false,
		Message:   message,
		Code:      statusCode,
		RequestID: requestID,
		Timestamp: time.Now().Unix(),
		Meta:      map[string]any{"error": err.Error()},
	}

	ctx.JSON(statusCode, response)
}

// 认证检查
func (h *BaseHTTPGenericHandler[T, R]) checkAuth(ctx *app.RequestContext) error {
	token := string(ctx.GetHeader("Authorization"))
	if token == "" {
		return errors.New("Missing authorization token")
	}
	
	// 这里应该实现实际的认证逻辑
	// 简化版本，实际应用中需要JWT验证等
	if token != "Bearer valid-token" {
		return errors.New("Invalid authorization token")
	}
	
	return nil
}

// 速率限制检查
func (h *BaseHTTPGenericHandler[T, R]) checkRateLimit(ctx *app.RequestContext) error {
	// 这里应该实现实际的速率限制逻辑
	// 可以使用Redis或内存缓存
	key := h.RateLimit.KeyFunc(ctx)
	
	// 简化实现，实际应用中需要更复杂的逻辑
	_ = key
	
	return nil
}

// 缓存检查
func (h *BaseHTTPGenericHandler[T, R]) checkCache(request *HTTPGenericRequest[T]) (*GenericResult[R], bool) {
	// 这里应该实现实际的缓存逻辑
	// 可以使用Redis、Memcached等
	return nil, false
}

// 设置缓存
func (h *BaseHTTPGenericHandler[T, R]) setCache(request *HTTPGenericRequest[T], result *GenericResult[R]) {
	// 这里应该实现实际的缓存设置逻辑
}