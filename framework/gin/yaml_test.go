// Package gin - YAML功能测试
package gin

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestYAMLRendering 测试YAML渲染功能
func TestYAMLRendering(t *testing.T) {
	// 创建引擎
	r := New()
	
	// 设置YAML响应路由
	r.GET("/yaml", func(c *Context) {
		data := H{
			"message": "Hello YAML",
			"status":  200,
			"config": H{
				"host": "localhost",
				"port": 8080,
			},
		}
		c.YAML(200, data)
	})
	
	// 创建测试请求
	req, _ := http.NewRequest("GET", "/yaml", nil)
	w := httptest.NewRecorder()
	
	// 执行请求
	r.ServeHTTP(w, req)
	
	// 验证响应
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	// 验证Content-Type
	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/x-yaml") {
		t.Errorf("Expected Content-Type to contain 'application/x-yaml', got %s", contentType)
	}
	
	// 验证YAML内容
	body := w.Body.String()
	expectedStrings := []string{
		"message: Hello YAML",
		"status: 200",
		"config:",
		"host: localhost",
		"port: 8080",
	}
	
	for _, expected := range expectedStrings {
		if !strings.Contains(body, expected) {
			t.Errorf("Expected YAML to contain '%s', but got:\n%s", expected, body)
		}
	}
	
	t.Logf("YAML rendering test passed. Response:\n%s", body)
}

// TestYAMLBinding 测试YAML绑定功能
func TestYAMLBinding(t *testing.T) {
	// 定义测试结构体
	type Config struct {
		Name    string            `yaml:"name" binding:"required"`
		Port    int               `yaml:"port" binding:"min=1,max=65535"`
		Enabled bool              `yaml:"enabled"`
		Tags    []string          `yaml:"tags"`
		Meta    map[string]string `yaml:"meta"`
	}
	
	// 创建引擎
	r := New()
	
	var boundConfig Config
	
	// 设置YAML绑定路由
	r.POST("/bind-yaml", func(c *Context) {
		if err := c.ShouldBindYAML(&boundConfig); err != nil {
			c.JSON(400, H{"error": err.Error()})
			return
		}
		
		c.JSON(200, H{
			"message": "YAML binding successful",
			"config":  boundConfig,
		})
	})
	
	// 测试数据
	yamlData := `
name: test-service
port: 8080
enabled: true
tags:
  - web
  - api
  - production
meta:
  version: "1.0.0"
  author: "test"
`
	
	// 创建测试请求
	req, _ := http.NewRequest("POST", "/bind-yaml", strings.NewReader(yamlData))
	req.Header.Set("Content-Type", "application/x-yaml")
	w := httptest.NewRecorder()
	
	// 执行请求
	r.ServeHTTP(w, req)
	
	// 验证响应
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d. Response: %s", w.Code, w.Body.String())
	}
	
	// 验证绑定结果
	if boundConfig.Name != "test-service" {
		t.Errorf("Expected name 'test-service', got '%s'", boundConfig.Name)
	}
	
	if boundConfig.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", boundConfig.Port)
	}
	
	if !boundConfig.Enabled {
		t.Error("Expected enabled to be true")
	}
	
	if len(boundConfig.Tags) != 3 {
		t.Errorf("Expected 3 tags, got %d", len(boundConfig.Tags))
	}
	
	if boundConfig.Meta["version"] != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", boundConfig.Meta["version"])
	}
	
	t.Logf("YAML binding test passed. Bound config: %+v", boundConfig)
}

// TestBindYAML 测试BindYAML方法（强制绑定）
func TestBindYAML(t *testing.T) {
	type User struct {
		Name string `yaml:"name" binding:"required"`
		Age  int    `yaml:"age" binding:"min=1"`
	}
	
	r := New()
	
	r.POST("/bind-yaml-must", func(c *Context) {
		var user User
		if err := c.BindYAML(&user); err != nil {
			// BindYAML会自动返回错误，这里不应该到达
			t.Error("BindYAML should handle errors automatically")
			return
		}
		
		c.YAML(200, H{
			"message": "User bound successfully",
			"user":    user,
		})
	})
	
	// 测试有效数据
	validYAML := `
name: "John Doe"
age: 25
`
	
	req, _ := http.NewRequest("POST", "/bind-yaml-must", strings.NewReader(validYAML))
	req.Header.Set("Content-Type", "application/x-yaml")
	w := httptest.NewRecorder()
	
	r.ServeHTTP(w, req)
	
	if w.Code != 200 {
		t.Errorf("Expected status 200 for valid YAML, got %d. Response: %s", w.Code, w.Body.String())
	}
	
	// 测试无效数据
	invalidYAML := `
age: 25
# 缺少required字段name
`
	
	req, _ = http.NewRequest("POST", "/bind-yaml-must", strings.NewReader(invalidYAML))
	req.Header.Set("Content-Type", "application/x-yaml")
	w = httptest.NewRecorder()
	
	r.ServeHTTP(w, req)
	
	if w.Code != 400 {
		t.Errorf("Expected status 400 for invalid YAML, got %d. Response: %s", w.Code, w.Body.String())
	}
	
	t.Log("BindYAML test passed")
}

// TestShouldBindBodyWithYAML 测试从请求体绑定YAML
func TestShouldBindBodyWithYAML(t *testing.T) {
	type Settings struct {
		Theme    string          `yaml:"theme"`
		Features map[string]bool `yaml:"features"`
	}
	
	r := New()
	
	r.POST("/bind-body-yaml", func(c *Context) {
		var settings Settings
		if err := c.ShouldBindBodyWithYAML(&settings); err != nil {
			c.YAML(422, H{
				"error":   "Invalid YAML format",
				"details": err.Error(),
			})
			return
		}
		
		c.YAML(200, H{
			"message":  "Settings bound successfully",
			"settings": settings,
		})
	})
	
	yamlData := `
theme: dark
features:
  notifications: true
  dark_mode: true
  auto_save: false
`
	
	req, _ := http.NewRequest("POST", "/bind-body-yaml", strings.NewReader(yamlData))
	req.Header.Set("Content-Type", "application/x-yaml")
	w := httptest.NewRecorder()
	
	r.ServeHTTP(w, req)
	
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d. Response: %s", w.Code, w.Body.String())
	}
	
	t.Log("ShouldBindBodyWithYAML test passed")
}

// TestYAMLBindingValidation 测试YAML绑定的验证功能
func TestYAMLBindingValidation(t *testing.T) {
	type ValidatedConfig struct {
		Email    string `yaml:"email" binding:"required,email"`
		Port     int    `yaml:"port" binding:"required,min=1000,max=65535"`
		URL      string `yaml:"url" binding:"required,url"`
		Required string `yaml:"required" binding:"required"`
	}
	
	r := New()
	
	r.POST("/validate-yaml", func(c *Context) {
		var config ValidatedConfig
		if err := c.ShouldBindYAML(&config); err != nil {
			c.JSON(400, H{"validation_error": err.Error()})
			return
		}
		
		c.JSON(200, H{"message": "Validation passed", "config": config})
	})
	
	// 测试无效数据
	invalidYAML := `
email: "not-an-email"
port: 99999999
url: "not-a-url"
# required字段缺失
`
	
	req, _ := http.NewRequest("POST", "/validate-yaml", strings.NewReader(invalidYAML))
	req.Header.Set("Content-Type", "application/x-yaml")
	w := httptest.NewRecorder()
	
	r.ServeHTTP(w, req)
	
	if w.Code != 400 {
		t.Errorf("Expected validation to fail with status 400, got %d", w.Code)
	}
	
	// 验证错误消息包含验证信息
	responseBody := w.Body.String()
	if !strings.Contains(responseBody, "validation_error") {
		t.Errorf("Expected response to contain validation error, got: %s", responseBody)
	}
	
	t.Log("YAML validation test passed")
}

// BenchmarkYAMLRendering YAML渲染性能测试
func BenchmarkYAMLRendering(b *testing.B) {
	r := New()
	
	data := H{
		"message": "Benchmark test",
		"data": H{
			"items": []H{
				{"id": 1, "name": "item1"},
				{"id": 2, "name": "item2"},
				{"id": 3, "name": "item3"},
			},
		},
	}
	
	r.GET("/benchmark", func(c *Context) {
		c.YAML(200, data)
	})
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/benchmark", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

// TestYAMLWithComplexData 测试复杂数据结构的YAML处理
func TestYAMLWithComplexData(t *testing.T) {
	type ComplexData struct {
		ID       int                    `yaml:"id"`
		Name     string                 `yaml:"name"`
		Tags     []string               `yaml:"tags"`
		Metadata map[string]interface{} `yaml:"metadata"`
		Config   struct {
			Host string `yaml:"host"`
			Port int    `yaml:"port"`
		} `yaml:"config"`
	}
	
	r := New()
	
	// 测试渲染复杂数据
	r.GET("/complex", func(c *Context) {
		data := ComplexData{
			ID:   1,
			Name: "complex-service",
			Tags: []string{"api", "service", "production"},
			Metadata: map[string]interface{}{
				"version":     "1.0.0",
				"maintainer":  "team@example.com",
				"last_update": "2025-08-17",
			},
		}
		data.Config.Host = "localhost"
		data.Config.Port = 8080
		
		c.YAML(200, data)
	})
	
	req, _ := http.NewRequest("GET", "/complex", nil)
	w := httptest.NewRecorder()
	
	r.ServeHTTP(w, req)
	
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	body := w.Body.String()
	expectedStrings := []string{
		"id: 1",
		"name: complex-service",
		"tags:",
		"- api",
		"metadata:",
		"version: 1.0.0",
		"config:",
		"host: localhost",
		"port: 8080",
	}
	
	for _, expected := range expectedStrings {
		if !strings.Contains(body, expected) {
			t.Errorf("Expected YAML to contain '%s', but got:\n%s", expected, body)
		}
	}
	
	t.Logf("Complex YAML test passed. Response:\n%s", body)
}