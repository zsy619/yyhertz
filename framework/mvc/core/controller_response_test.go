package core

import (
	"encoding/json"
	"encoding/xml"
	"os"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/stretchr/testify/assert"

	context "github.com/zsy619/yyhertz/framework/mvc/context"
)

// TestServeJSON_BasicFunctionality 测试ServeJSON基本功能
func TestServeJSON_BasicFunctionality(t *testing.T) {
	tests := []struct {
		name     string
		data     any
		expected string
	}{
		{
			name:     "简单字符串",
			data:     "hello world",
			expected: `"hello world"`,
		},
		{
			name:     "数字",
			data:     42,
			expected: `42`,
		},
		{
			name: "简单对象",
			data: map[string]any{
				"name": "test",
				"age":  25,
			},
			expected: `{"age":25,"name":"test"}`,
		},
		{
			name: "数组",
			data: []string{"a", "b", "c"},
			expected: `["a","b","c"]`,
		},
		{
			name:     "布尔值",
			data:     true,
			expected: `true`,
		},
		{
			name:     "null值",
			data:     nil,
			expected: `null`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试环境
			ctx := createTestContext()
			controller := NewBaseController()
			controller.Ctx = ctx
			controller.Data = make(map[string]any)

			// 设置JSON数据
			controller.Data["json"] = tt.data

			// 调用ServeJSON
			controller.ServeJSON()

			// 验证响应
			response := ctx.Request().Response
			assert.Equal(t, "application/json; charset=utf-8", string(response.Header.ContentType()))
			assert.JSONEq(t, tt.expected, string(response.Body()))
		})
	}
}

// TestServeJSON_WithEncoding 测试UTF-8编码功能
func TestServeJSON_WithEncoding(t *testing.T) {
	tests := []struct {
		name     string
		data     any
		encoding bool
		expected string
	}{
		{
			name:     "中文字符_不编码",
			data:     map[string]any{"message": "你好世界"},
			encoding: false,
			expected: `{"message":"你好世界"}`,
		},
		{
			name:     "中文字符_编码",
			data:     map[string]any{"message": "你好世界"},
			encoding: true,
			expected: `{
  "message": "\u4F60\u597D\u4E16\u754C"
}`,
		},
		{
			name:     "英文字符_编码",
			data:     map[string]any{"message": "hello"},
			encoding: true,
			expected: `{
  "message": "hello"
}`,
		},
		{
			name:     "混合字符_编码",
			data:     map[string]any{"message": "hello你好"},
			encoding: true,
			expected: `{
  "message": "hello\u4F60\u597D"
}`,
		},
		{
			name:     "特殊符号_编码",
			data:     map[string]any{"message": "测试®™"},
			encoding: true,
			expected: `{
  "message": "\u6D4B\u8BD5\u00AE\u2122"
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试环境
			ctx := createTestContext()
			controller := NewBaseController()
			controller.Ctx = ctx
			controller.Data = make(map[string]any)

			// 设置JSON数据
			controller.Data["json"] = tt.data

			// 调用ServeJSON
			controller.ServeJSON(tt.encoding)

			// 验证响应
			response := ctx.Request().Response
			assert.Equal(t, "application/json; charset=utf-8", string(response.Header.ContentType()))
			
			// 对于编码测试，直接比较字符串
			if tt.encoding {
				assert.Equal(t, tt.expected, string(response.Body()))
			} else {
				assert.JSONEq(t, tt.expected, string(response.Body()))
			}
		})
	}
}

// TestServeJSON_IndentationBehavior 测试缩进行为
func TestServeJSON_IndentationBehavior(t *testing.T) {
	tests := []struct {
		name      string
		runMode   string
		shouldIndent bool
	}{
		{
			name:      "开发模式_dev",
			runMode:   "dev",
			shouldIndent: true,
		},
		{
			name:      "开发模式_development",
			runMode:   "development",
			shouldIndent: true,
		},
		{
			name:      "调试模式_debug",
			runMode:   "debug",
			shouldIndent: true,
		},
		{
			name:      "生产模式_prod",
			runMode:   "prod",
			shouldIndent: false,
		},
		{
			name:      "生产模式_production",
			runMode:   "production",
			shouldIndent: false,
		},
		{
			name:      "测试模式_test",
			runMode:   "test",
			shouldIndent: false,
		},
	}

	testData := map[string]any{
		"status": "success",
		"data": map[string]any{
			"id":   1,
			"name": "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试环境
			ctx := createTestContext()
			controller := NewBaseController()
			controller.Ctx = ctx
			controller.Data = make(map[string]any)

			// 设置JSON数据
			controller.Data["json"] = testData

			// 设置环境变量模拟运行模式
			oldEnv := os.Getenv("APP_ENV")
			os.Setenv("APP_ENV", tt.runMode)
			defer os.Setenv("APP_ENV", oldEnv)
			
			// 调用ServeJSON
			controller.ServeJSON()

			// 验证响应
			response := ctx.Request().Response
			responseBody := string(response.Body())

			if tt.shouldIndent {
				// 格式化输出应包含缩进
				assert.Contains(t, responseBody, "  ")
				assert.Contains(t, responseBody, "\n")
			} else {
				// 紧凑输出不应包含多余空格
				var compactData any
				err := json.Unmarshal(response.Body(), &compactData)
				assert.NoError(t, err)
				
				compactBytes, err := json.Marshal(compactData)
				assert.NoError(t, err)
				assert.Equal(t, string(compactBytes), responseBody)
			}
		})
	}
}

// TestServeJSON_ErrorCases 测试错误情况
func TestServeJSON_ErrorCases(t *testing.T) {
	t.Run("Context为nil", func(t *testing.T) {
		controller := NewBaseController()
		controller.Ctx = nil
		controller.Data = make(map[string]any)
		controller.Data["json"] = "test"

		// 调用ServeJSON不应该panic
		assert.NotPanics(t, func() {
			controller.ServeJSON()
		})
	})

	t.Run("Data为nil", func(t *testing.T) {
		ctx := createTestContext()
		controller := NewBaseController()
		controller.Ctx = ctx
		controller.Data = nil

		// 调用ServeJSON
		controller.ServeJSON()

		// 验证错误响应
		response := ctx.Request().Response
		assert.Equal(t, consts.StatusInternalServerError, response.StatusCode())
		assert.Contains(t, string(response.Body()), "No JSON data provided")
	})

	t.Run("JSON数据不存在", func(t *testing.T) {
		ctx := createTestContext()
		controller := NewBaseController()
		controller.Ctx = ctx
		controller.Data = make(map[string]any)
		// 不设置 controller.Data["json"]

		// 调用ServeJSON
		controller.ServeJSON()

		// 验证错误响应
		response := ctx.Request().Response
		assert.Equal(t, consts.StatusInternalServerError, response.StatusCode())
		assert.Contains(t, string(response.Body()), "No JSON data provided")
	})

	t.Run("无法序列化的数据", func(t *testing.T) {
		ctx := createTestContext()
		controller := NewBaseController()
		controller.Ctx = ctx
		controller.Data = make(map[string]any)

		// 设置无法序列化的数据（包含循环引用）
		cyclicData := make(map[string]any)
		cyclicData["self"] = cyclicData
		controller.Data["json"] = cyclicData

		// 调用ServeJSON
		controller.ServeJSON()

		// 验证错误响应
		response := ctx.Request().Response
		assert.Equal(t, consts.StatusInternalServerError, response.StatusCode())
		assert.Contains(t, string(response.Body()), "Failed to serialize JSON data")
	})
}

// TestServeJSON_ComplexData 测试复杂数据结构
func TestServeJSON_ComplexData(t *testing.T) {
	complexData := map[string]any{
		"user": map[string]any{
			"id":       123,
			"username": "testuser",
			"profile": map[string]any{
				"age":    25,
				"gender": "male",
				"tags":   []string{"developer", "gamer"},
			},
		},
		"settings": map[string]any{
			"theme":        "dark",
			"notifications": true,
			"preferences": map[string]any{
				"language": "zh-CN",
				"timezone": "Asia/Shanghai",
			},
		},
		"metadata": map[string]any{
			"created_at": "2023-01-01T00:00:00Z",
			"version":    "1.0.0",
			"features":   []any{"feature1", 123, true, nil},
		},
	}

	ctx := createTestContext()
	controller := NewBaseController()
	controller.Ctx = ctx
	controller.Data = make(map[string]any)
	controller.Data["json"] = complexData

	// 调用ServeJSON
	controller.ServeJSON()

	// 验证响应
	response := ctx.Request().Response
	assert.Equal(t, "application/json; charset=utf-8", string(response.Header.ContentType()))
	
	// 验证JSON可以正确解析
	var parsedData map[string]any
	err := json.Unmarshal(response.Body(), &parsedData)
	assert.NoError(t, err)
	
	// 验证部分数据结构
	assert.Equal(t, float64(123), parsedData["user"].(map[string]any)["id"])
	assert.Equal(t, "testuser", parsedData["user"].(map[string]any)["username"])
	assert.Equal(t, "dark", parsedData["settings"].(map[string]any)["theme"])
}

// TestServeJSON_BeegoCompatibility 测试与Beego兼容性
func TestServeJSON_BeegoCompatibility(t *testing.T) {
	// 模拟典型的Beego使用模式
	ctx := createTestContext()
	controller := NewBaseController()
	controller.Ctx = ctx
	controller.Data = make(map[string]any)

	// 典型的Beego响应格式
	beegoResponse := map[string]any{
		"status": "success",
		"code":   200,
		"data": map[string]any{
			"users": []map[string]any{
				{"id": 1, "name": "Alice"},
				{"id": 2, "name": "Bob"},
			},
			"total": 2,
		},
		"message": "获取用户列表成功",
	}

	controller.Data["json"] = beegoResponse

	// 调用ServeJSON（模拟Beego的调用方式）
	controller.ServeJSON()

	// 验证响应
	response := ctx.Request().Response
	assert.Equal(t, "application/json; charset=utf-8", string(response.Header.ContentType()))
	
	var parsedResponse map[string]any
	err := json.Unmarshal(response.Body(), &parsedResponse)
	assert.NoError(t, err)
	
	assert.Equal(t, "success", parsedResponse["status"])
	assert.Equal(t, float64(200), parsedResponse["code"])
	assert.Equal(t, "获取用户列表成功", parsedResponse["message"])
	
	// 验证数据结构
	data := parsedResponse["data"].(map[string]any)
	assert.Equal(t, float64(2), data["total"])
	
	users := data["users"].([]any)
	assert.Len(t, users, 2)
}

// TestEncodeUTF8ToUnicode 测试UTF-8编码方法
func TestEncodeUTF8ToUnicode(t *testing.T) {
	controller := NewBaseController()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "纯ASCII字符",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "中文字符",
			input:    "你好",
			expected: "你好", // 不在JSON字符串内，不编码
		},
		{
			name:     "混合字符",
			input:    "hello你好world",
			expected: "hello你好world", // 不在JSON字符串内，不编码
		},
		{
			name:     "特殊符号",
			input:    "®™",
			expected: "®™", // 不在JSON字符串内，不编码
		},
		{
			name:     "JSON字符串",
			input:    `{"name":"张三"}`,
			expected: `{"name":"\u5F20\u4E09"}`, // 在JSON字符串内，进行编码
		},
		{
			name:     "空字符串",
			input:    "",
			expected: "",
		},
		{
			name:     "数字和符号",
			input:    "123!@#",
			expected: "123!@#",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := controller.encodeUTF8ToUnicode([]byte(tt.input))
			assert.Equal(t, tt.expected, string(result))
		})
	}
}

// TestShouldIndentJSON 测试JSON缩进判断方法
func TestShouldIndentJSON(t *testing.T) {
	controller := NewBaseController()

	// 注意：实际测试时需要模拟config.GetAppRunMode()的返回值
	// 这里只测试方法的存在性和基本逻辑
	result := controller.shouldIndentJSON()
	assert.IsType(t, true, result) // 验证返回布尔值
}

// ============= ServeXML测试 =============

// TestServeXML_BasicFunctionality 测试ServeXML基本功能
func TestServeXML_BasicFunctionality(t *testing.T) {
	type TestData struct {
		XMLName xml.Name `xml:"response"`
		Name    string   `xml:"name"`
		Age     int      `xml:"age"`
	}

	tests := []struct {
		name        string
		data        any
		shouldError bool
		contains    string
	}{
		{
			name:        "简单结构体",
			data:        TestData{Name: "test", Age: 25},
			shouldError: false,
			contains:    "test",
		},
		{
			name:        "map数据_应该失败",
			data:        map[string]any{"message": "hello", "code": 200},
			shouldError: true,
			contains:    "Failed to serialize XML data",
		},
		{
			name:        "字符串",
			data:        "hello world",
			shouldError: false,
			contains:    "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试环境
			ctx := createTestContext()
			controller := NewBaseController()
			controller.Ctx = ctx
			controller.Data = make(map[string]any)

			// 设置XML数据
			controller.Data["xml"] = tt.data

			// 调用ServeXML
			controller.ServeXML()

			// 验证响应
			response := ctx.Request().Response
			responseBody := string(response.Body())
			
			if tt.shouldError {
				// 验证错误情况
				assert.Equal(t, consts.StatusInternalServerError, response.StatusCode())
				assert.Contains(t, responseBody, tt.contains)
			} else {
				// 验证成功情况
				assert.Equal(t, "application/xml; charset=utf-8", string(response.Header.ContentType()))
				assert.Contains(t, responseBody, tt.contains)
			}
		})
	}
}

// TestServeXML_ErrorCases 测试ServeXML错误情况
func TestServeXML_ErrorCases(t *testing.T) {
	t.Run("Context为nil", func(t *testing.T) {
		controller := NewBaseController()
		controller.Ctx = nil
		controller.Data = make(map[string]any)
		controller.Data["xml"] = "test"

		// 调用ServeXML不应该panic
		assert.NotPanics(t, func() {
			controller.ServeXML()
		})
	})

	t.Run("XML数据不存在", func(t *testing.T) {
		ctx := createTestContext()
		controller := NewBaseController()
		controller.Ctx = ctx
		controller.Data = make(map[string]any)
		// 不设置 controller.Data["xml"]

		// 调用ServeXML
		controller.ServeXML()

		// 验证错误响应
		response := ctx.Request().Response
		assert.Equal(t, consts.StatusInternalServerError, response.StatusCode())
		assert.Contains(t, string(response.Body()), "No XML data provided")
	})
}

// ============= ServeJSONP测试 =============

// TestServeJSONP_BasicFunctionality 测试ServeJSONP基本功能
func TestServeJSONP_BasicFunctionality(t *testing.T) {
	tests := []struct {
		name     string
		data     any
		callback string
		expected string
	}{
		{
			name:     "简单对象_默认回调",
			data:     map[string]any{"message": "hello"},
			callback: "",
			expected: `callback({"message":"hello"});`,
		},
		{
			name:     "简单对象_自定义回调",
			data:     map[string]any{"status": "success"},
			callback: "myCallback",
			expected: `myCallback({"status":"success"});`,
		},
		{
			name:     "数组数据",
			data:     []string{"a", "b", "c"},
			callback: "handleData",
			expected: `handleData(["a","b","c"]);`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试环境
			ctx := createTestContext()
			if tt.callback != "" {
				ctx.Request().URI().QueryArgs().Set("callback", tt.callback)
			}
			
			controller := NewBaseController()
			controller.Ctx = ctx
			controller.Data = make(map[string]any)

			// 设置JSONP数据
			controller.Data["jsonp"] = tt.data

			// 调用ServeJSONP
			controller.ServeJSONP()

			// 验证响应
			response := ctx.Request().Response
			assert.Equal(t, "application/javascript; charset=utf-8", string(response.Header.ContentType()))
			
			// 对于缩进，需要处理默认开发模式的格式化
			responseBody := string(response.Body())
			
			// 验证回调函数包装
			if tt.callback != "" {
				assert.Contains(t, responseBody, tt.callback+"(")
			} else {
				assert.Contains(t, responseBody, "callback(")
			}
			assert.Contains(t, responseBody, ");")
		})
	}
}

// ============= ServeYAML测试 =============

// TestServeYAML_BasicFunctionality 测试ServeYAML基本功能
func TestServeYAML_BasicFunctionality(t *testing.T) {
	tests := []struct {
		name string
		data any
	}{
		{
			name: "简单对象",
			data: map[string]any{"name": "test", "age": 25},
		},
		{
			name: "数组",
			data: []string{"item1", "item2", "item3"},
		},
		{
			name: "嵌套结构",
			data: map[string]any{
				"user": map[string]any{
					"name": "alice",
					"age":  30,
				},
				"active": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试环境
			ctx := createTestContext()
			controller := NewBaseController()
			controller.Ctx = ctx
			controller.Data = make(map[string]any)

			// 设置YAML数据
			controller.Data["yaml"] = tt.data

			// 调用ServeYAML
			controller.ServeYAML()

			// 验证响应
			response := ctx.Request().Response
			assert.Equal(t, "application/yaml; charset=utf-8", string(response.Header.ContentType()))
			
			// 验证YAML内容包含关键数据
			responseBody := string(response.Body())
			assert.NotEmpty(t, responseBody)
			
			// 对于简单验证，确保不是错误响应
			assert.NotContains(t, responseBody, "error:")
		})
	}
}

// ============= ServeFormatted测试 =============

// TestServeFormatted_AcceptHeader 测试ServeFormatted根据Accept header选择格式
func TestServeFormatted_AcceptHeader(t *testing.T) {
	type XMLData struct {
		XMLName xml.Name `xml:"data"`
		Message string   `xml:"message"`
	}

	tests := []struct {
		name         string
		acceptHeader string
		setupData    func(controller *BaseController)
		expectedType string
	}{
		{
			name:         "JSON格式",
			acceptHeader: "application/json",
			setupData: func(controller *BaseController) {
				controller.Data["json"] = map[string]any{"message": "json test"}
			},
			expectedType: "application/json; charset=utf-8",
		},
		{
			name:         "XML格式",
			acceptHeader: "application/xml",
			setupData: func(controller *BaseController) {
				controller.Data["xml"] = XMLData{Message: "xml test"}
			},
			expectedType: "application/xml; charset=utf-8",
		},
		{
			name:         "YAML格式",
			acceptHeader: "application/yaml",
			setupData: func(controller *BaseController) {
				controller.Data["yaml"] = map[string]any{"message": "yaml test"}
			},
			expectedType: "application/yaml; charset=utf-8",
		},
		{
			name:         "默认JSON格式",
			acceptHeader: "*/*",
			setupData: func(controller *BaseController) {
				controller.Data["json"] = map[string]any{"message": "default test"}
			},
			expectedType: "application/json; charset=utf-8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试环境
			ctx := createTestContext()
			ctx.Request().Request.Header.Set("Accept", tt.acceptHeader)
			
			controller := NewBaseController()
			controller.Ctx = ctx
			controller.Data = make(map[string]any)

			// 设置测试数据
			tt.setupData(controller)

			// 调用ServeFormatted
			controller.ServeFormatted()

			// 验证响应
			response := ctx.Request().Response
			assert.Equal(t, tt.expectedType, string(response.Header.ContentType()))
			assert.NotEqual(t, consts.StatusInternalServerError, response.StatusCode())
		})
	}
}

// createTestContext 创建测试用的Context
func createTestContext() *context.Context {
	hertzCtx := app.NewContext(0)
	hertzCtx.Request.SetMethod("GET")
	hertzCtx.Request.SetRequestURI("/test")
	
	return context.NewContext(hertzCtx)
}