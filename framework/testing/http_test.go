package testing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"

	"github.com/zsy619/yyhertz/framework/config"
)

// HTTPTestClient HTTP测试客户端
type HTTPTestClient struct {
	server   *httptest.Server
	client   *http.Client
	hertzApp *server.Hertz
	baseURL  string
	headers  map[string]string
	cookies  []*http.Cookie
	timeout  time.Duration
	t        *testing.T
}

// NewHTTPTestClient 创建HTTP测试客户端
func NewHTTPTestClient(t *testing.T, hertzApp *server.Hertz) *HTTPTestClient {
	// 创建测试服务器
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 将标准HTTP请求转换为Hertz请求
		// 这里需要一个适配器来处理
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Test server response"))
	}))

	client := &HTTPTestClient{
		server:   testServer,
		client:   &http.Client{Timeout: time.Second * 30},
		hertzApp: hertzApp,
		baseURL:  testServer.URL,
		headers:  make(map[string]string),
		cookies:  make([]*http.Cookie, 0),
		timeout:  time.Second * 30,
		t:        t,
	}

	return client
}

// NewHTTPTestClientWithHandler 使用处理器创建HTTP测试客户端
func NewHTTPTestClientWithHandler(t *testing.T, handler http.Handler) *HTTPTestClient {
	testServer := httptest.NewServer(handler)

	return &HTTPTestClient{
		server:  testServer,
		client:  &http.Client{Timeout: time.Second * 30},
		baseURL: testServer.URL,
		headers: make(map[string]string),
		cookies: make([]*http.Cookie, 0),
		timeout: time.Second * 30,
		t:       t,
	}
}

// SetTimeout 设置超时时间
func (htc *HTTPTestClient) SetTimeout(timeout time.Duration) *HTTPTestClient {
	htc.timeout = timeout
	htc.client.Timeout = timeout
	return htc
}

// SetHeader 设置请求头
func (htc *HTTPTestClient) SetHeader(key, value string) *HTTPTestClient {
	htc.headers[key] = value
	return htc
}

// SetHeaders 批量设置请求头
func (htc *HTTPTestClient) SetHeaders(headers map[string]string) *HTTPTestClient {
	for k, v := range headers {
		htc.headers[k] = v
	}
	return htc
}

// SetCookie 设置Cookie
func (htc *HTTPTestClient) SetCookie(cookie *http.Cookie) *HTTPTestClient {
	htc.cookies = append(htc.cookies, cookie)
	return htc
}

// SetAuth 设置基础认证
func (htc *HTTPTestClient) SetAuth(username, password string) *HTTPTestClient {
	htc.SetHeader("Authorization", "Basic "+basicAuth(username, password))
	return htc
}

// SetBearerToken 设置Bearer Token
func (htc *HTTPTestClient) SetBearerToken(token string) *HTTPTestClient {
	htc.SetHeader("Authorization", "Bearer "+token)
	return htc
}

// GET 发送GET请求
func (htc *HTTPTestClient) GET(path string, params ...map[string]string) *HTTPResponse {
	fullURL := htc.buildURL(path, params...)
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		htc.t.Fatalf("Failed to create GET request: %v", err)
	}

	return htc.doRequest(req)
}

// POST 发送POST请求
func (htc *HTTPTestClient) POST(path string, body any) *HTTPResponse {
	fullURL := htc.buildURL(path)

	var reqBody io.Reader
	var contentType string

	switch v := body.(type) {
	case string:
		reqBody = strings.NewReader(v)
		contentType = "text/plain"
	case []byte:
		reqBody = bytes.NewReader(v)
		contentType = "application/octet-stream"
	case map[string]any:
		jsonData, err := json.Marshal(v)
		if err != nil {
			htc.t.Fatalf("Failed to marshal JSON: %v", err)
		}
		reqBody = bytes.NewReader(jsonData)
		contentType = "application/json"
	case url.Values:
		reqBody = strings.NewReader(v.Encode())
		contentType = "application/x-www-form-urlencoded"
	default:
		jsonData, err := json.Marshal(v)
		if err != nil {
			htc.t.Fatalf("Failed to marshal body: %v", err)
		}
		reqBody = bytes.NewReader(jsonData)
		contentType = "application/json"
	}

	req, err := http.NewRequest("POST", fullURL, reqBody)
	if err != nil {
		htc.t.Fatalf("Failed to create POST request: %v", err)
	}

	if _, exists := htc.headers["Content-Type"]; !exists {
		req.Header.Set("Content-Type", contentType)
	}

	return htc.doRequest(req)
}

// PUT 发送PUT请求
func (htc *HTTPTestClient) PUT(path string, body any) *HTTPResponse {
	fullURL := htc.buildURL(path)

	var reqBody io.Reader
	var contentType string

	switch v := body.(type) {
	case string:
		reqBody = strings.NewReader(v)
		contentType = "text/plain"
	case []byte:
		reqBody = bytes.NewReader(v)
		contentType = "application/octet-stream"
	default:
		jsonData, err := json.Marshal(v)
		if err != nil {
			htc.t.Fatalf("Failed to marshal JSON: %v", err)
		}
		reqBody = bytes.NewReader(jsonData)
		contentType = "application/json"
	}

	req, err := http.NewRequest("PUT", fullURL, reqBody)
	if err != nil {
		htc.t.Fatalf("Failed to create PUT request: %v", err)
	}

	if _, exists := htc.headers["Content-Type"]; !exists {
		req.Header.Set("Content-Type", contentType)
	}

	return htc.doRequest(req)
}

// DELETE 发送DELETE请求
func (htc *HTTPTestClient) DELETE(path string) *HTTPResponse {
	fullURL := htc.buildURL(path)
	req, err := http.NewRequest("DELETE", fullURL, nil)
	if err != nil {
		htc.t.Fatalf("Failed to create DELETE request: %v", err)
	}

	return htc.doRequest(req)
}

// PATCH 发送PATCH请求
func (htc *HTTPTestClient) PATCH(path string, body any) *HTTPResponse {
	fullURL := htc.buildURL(path)

	var reqBody io.Reader
	var contentType string

	switch v := body.(type) {
	case string:
		reqBody = strings.NewReader(v)
		contentType = "text/plain"
	case []byte:
		reqBody = bytes.NewReader(v)
		contentType = "application/octet-stream"
	default:
		jsonData, err := json.Marshal(v)
		if err != nil {
			htc.t.Fatalf("Failed to marshal JSON: %v", err)
		}
		reqBody = bytes.NewReader(jsonData)
		contentType = "application/json"
	}

	req, err := http.NewRequest("PATCH", fullURL, reqBody)
	if err != nil {
		htc.t.Fatalf("Failed to create PATCH request: %v", err)
	}

	if _, exists := htc.headers["Content-Type"]; !exists {
		req.Header.Set("Content-Type", contentType)
	}

	return htc.doRequest(req)
}

// doRequest 执行HTTP请求
func (htc *HTTPTestClient) doRequest(req *http.Request) *HTTPResponse {
	// 设置请求头
	for k, v := range htc.headers {
		req.Header.Set(k, v)
	}

	// 设置Cookie
	for _, cookie := range htc.cookies {
		req.AddCookie(cookie)
	}

	// 设置超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), htc.timeout)
	defer cancel()
	req = req.WithContext(ctx)

	// 发送请求
	start := time.Now()
	resp, err := htc.client.Do(req)
	duration := time.Since(start)

	if err != nil {
		htc.t.Fatalf("HTTP request failed: %v", err)
	}

	return NewHTTPResponse(htc.t, resp, duration)
}

// buildURL 构建完整URL
func (htc *HTTPTestClient) buildURL(path string, params ...map[string]string) string {
	fullURL := htc.baseURL + path

	if len(params) > 0 && len(params[0]) > 0 {
		values := url.Values{}
		for k, v := range params[0] {
			values.Add(k, v)
		}
		if strings.Contains(fullURL, "?") {
			fullURL += "&" + values.Encode()
		} else {
			fullURL += "?" + values.Encode()
		}
	}

	return fullURL
}

// Close 关闭测试客户端
func (htc *HTTPTestClient) Close() {
	if htc.server != nil {
		htc.server.Close()
	}
}

// HTTPResponse HTTP响应包装器
type HTTPResponse struct {
	t        *testing.T
	response *http.Response
	body     []byte
	duration time.Duration
}

// NewHTTPResponse 创建HTTP响应包装器
func NewHTTPResponse(t *testing.T, resp *http.Response, duration time.Duration) *HTTPResponse {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	return &HTTPResponse{
		t:        t,
		response: resp,
		body:     body,
		duration: duration,
	}
}

// StatusCode 获取状态码
func (hr *HTTPResponse) StatusCode() int {
	return hr.response.StatusCode
}

// Header 获取响应头
func (hr *HTTPResponse) Header(key string) string {
	return hr.response.Header.Get(key)
}

// Headers 获取所有响应头
func (hr *HTTPResponse) Headers() http.Header {
	return hr.response.Header
}

// Body 获取响应体
func (hr *HTTPResponse) Body() []byte {
	return hr.body
}

// String 获取响应体字符串
func (hr *HTTPResponse) String() string {
	return string(hr.body)
}

// JSON 解析JSON响应
func (hr *HTTPResponse) JSON(v any) error {
	return json.Unmarshal(hr.body, v)
}

// Duration 获取请求耗时
func (hr *HTTPResponse) Duration() time.Duration {
	return hr.duration
}

// Cookies 获取响应中的Cookie
func (hr *HTTPResponse) Cookies() []*http.Cookie {
	return hr.response.Cookies()
}

// Cookie 获取指定名称的Cookie
func (hr *HTTPResponse) Cookie(name string) *http.Cookie {
	for _, cookie := range hr.response.Cookies() {
		if cookie.Name == name {
			return cookie
		}
	}
	return nil
}

// ============= 响应断言方法 =============

// AssertStatusCode 断言状态码
func (hr *HTTPResponse) AssertStatusCode(expected int) *HTTPResponse {
	if hr.response.StatusCode != expected {
		hr.t.Helper()
		hr.t.Errorf("Expected status code %d, got %d", expected, hr.response.StatusCode)
	}
	return hr
}

// AssertStatusOK 断言状态码为200
func (hr *HTTPResponse) AssertStatusOK() *HTTPResponse {
	return hr.AssertStatusCode(http.StatusOK)
}

// AssertStatusCreated 断言状态码为201
func (hr *HTTPResponse) AssertStatusCreated() *HTTPResponse {
	return hr.AssertStatusCode(http.StatusCreated)
}

// AssertStatusBadRequest 断言状态码为400
func (hr *HTTPResponse) AssertStatusBadRequest() *HTTPResponse {
	return hr.AssertStatusCode(http.StatusBadRequest)
}

// AssertStatusUnauthorized 断言状态码为401
func (hr *HTTPResponse) AssertStatusUnauthorized() *HTTPResponse {
	return hr.AssertStatusCode(http.StatusUnauthorized)
}

// AssertStatusNotFound 断言状态码为404
func (hr *HTTPResponse) AssertStatusNotFound() *HTTPResponse {
	return hr.AssertStatusCode(http.StatusNotFound)
}

// AssertStatusInternalServerError 断言状态码为500
func (hr *HTTPResponse) AssertStatusInternalServerError() *HTTPResponse {
	return hr.AssertStatusCode(http.StatusInternalServerError)
}

// AssertHeader 断言响应头
func (hr *HTTPResponse) AssertHeader(key, expected string) *HTTPResponse {
	actual := hr.response.Header.Get(key)
	if actual != expected {
		hr.t.Helper()
		hr.t.Errorf("Expected header %s to be '%s', got '%s'", key, expected, actual)
	}
	return hr
}

// AssertHeaderExists 断言响应头存在
func (hr *HTTPResponse) AssertHeaderExists(key string) *HTTPResponse {
	if hr.response.Header.Get(key) == "" {
		hr.t.Helper()
		hr.t.Errorf("Expected header %s to exist", key)
	}
	return hr
}

// AssertContentType 断言Content-Type
func (hr *HTTPResponse) AssertContentType(expected string) *HTTPResponse {
	return hr.AssertHeader("Content-Type", expected)
}

// AssertBodyContains 断言响应体包含指定内容
func (hr *HTTPResponse) AssertBodyContains(expected string) *HTTPResponse {
	body := string(hr.body)
	if !strings.Contains(body, expected) {
		hr.t.Helper()
		hr.t.Errorf("Expected response body to contain '%s', got '%s'", expected, body)
	}
	return hr
}

// AssertBodyEquals 断言响应体等于指定内容
func (hr *HTTPResponse) AssertBodyEquals(expected string) *HTTPResponse {
	body := string(hr.body)
	if body != expected {
		hr.t.Helper()
		hr.t.Errorf("Expected response body to be '%s', got '%s'", expected, body)
	}
	return hr
}

// AssertBodyNotEmpty 断言响应体不为空
func (hr *HTTPResponse) AssertBodyNotEmpty() *HTTPResponse {
	if len(hr.body) == 0 {
		hr.t.Helper()
		hr.t.Error("Expected response body to not be empty")
	}
	return hr
}

// AssertJSON 断言响应为有效JSON并解析到指定对象
func (hr *HTTPResponse) AssertJSON(v any) *HTTPResponse {
	if err := json.Unmarshal(hr.body, v); err != nil {
		hr.t.Helper()
		hr.t.Errorf("Failed to parse JSON response: %v", err)
	}
	return hr
}

// AssertJSONPath 断言JSON路径的值
func (hr *HTTPResponse) AssertJSONPath(path string, expected any) *HTTPResponse {
	var data any
	if err := json.Unmarshal(hr.body, &data); err != nil {
		hr.t.Helper()
		hr.t.Errorf("Failed to parse JSON response: %v", err)
		return hr
	}

	// 简单的JSON路径解析（只支持点表示法）
	value, err := getJSONValue(data, path)
	if err != nil {
		hr.t.Helper()
		hr.t.Errorf("Failed to get JSON path '%s': %v", path, err)
		return hr
	}

	if value != expected {
		hr.t.Helper()
		hr.t.Errorf("Expected JSON path '%s' to be %v, got %v", path, expected, value)
	}

	return hr
}

// AssertCookie 断言Cookie
func (hr *HTTPResponse) AssertCookie(name, expected string) *HTTPResponse {
	cookie := hr.Cookie(name)
	if cookie == nil {
		hr.t.Helper()
		hr.t.Errorf("Expected cookie '%s' to exist", name)
		return hr
	}

	if cookie.Value != expected {
		hr.t.Helper()
		hr.t.Errorf("Expected cookie '%s' to be '%s', got '%s'", name, expected, cookie.Value)
	}

	return hr
}

// AssertDurationLess 断言请求耗时小于指定时间
func (hr *HTTPResponse) AssertDurationLess(expected time.Duration) *HTTPResponse {
	if hr.duration >= expected {
		hr.t.Helper()
		hr.t.Errorf("Expected request duration to be less than %v, got %v", expected, hr.duration)
	}
	return hr
}

// ============= 辅助函数 =============

// basicAuth 生成基础认证字符串
func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64Encode([]byte(auth))
}

// base64Encode base64编码
func base64Encode(data []byte) string {
	const base64Table = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

	if len(data) == 0 {
		return ""
	}

	// 简单的base64编码实现
	var result strings.Builder
	for i := 0; i < len(data); i += 3 {
		b1, b2, b3 := data[i], uint8(0), uint8(0)
		if i+1 < len(data) {
			b2 = data[i+1]
		}
		if i+2 < len(data) {
			b3 = data[i+2]
		}

		result.WriteByte(base64Table[b1>>2])
		result.WriteByte(base64Table[((b1&0x03)<<4)|((b2&0xf0)>>4)])
		if i+1 < len(data) {
			result.WriteByte(base64Table[((b2&0x0f)<<2)|((b3&0xc0)>>6)])
		} else {
			result.WriteByte('=')
		}
		if i+2 < len(data) {
			result.WriteByte(base64Table[b3&0x3f])
		} else {
			result.WriteByte('=')
		}
	}

	return result.String()
}

// getJSONValue 从JSON数据中获取指定路径的值
func getJSONValue(data any, path string) (any, error) {
	parts := strings.Split(path, ".")
	current := data

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]any:
			val, exists := v[part]
			if !exists {
				return nil, fmt.Errorf("path '%s' not found", part)
			}
			current = val
		default:
			return nil, fmt.Errorf("cannot traverse path '%s' on non-object", part)
		}
	}

	return current, nil
}

// ============= Hertz特定的测试工具 =============

// HertzTestSuite Hertz测试套件
type HertzTestSuite struct {
	*BaseTestSuite
	App    *server.Hertz
	Client *HTTPTestClient
}

// NewHertzTestSuite 创建Hertz测试套件
func NewHertzTestSuite() *HertzTestSuite {
	return &HertzTestSuite{
		BaseTestSuite: NewBaseTestSuite(),
	}
}

// SetUp 设置Hertz测试环境
func (hts *HertzTestSuite) SetUp(t *testing.T) {
	hts.BaseTestSuite.SetUp(t)

	// 创建Hertz应用
	hts.App = server.Default()

	// 创建测试客户端
	hts.Client = NewHTTPTestClient(t, hts.App)

	config.Info("Hertz test suite initialized")
}

// TearDown 清理Hertz测试环境
func (hts *HertzTestSuite) TearDown(t *testing.T) {
	if hts.Client != nil {
		hts.Client.Close()
	}

	hts.BaseTestSuite.TearDown(t)
	config.Info("Hertz test suite cleaned up")
}

// RegisterRoute 注册测试路由
func (hts *HertzTestSuite) RegisterRoute(method, path string, handler app.HandlerFunc) {
	switch strings.ToUpper(method) {
	case "GET":
		hts.App.GET(path, handler)
	case "POST":
		hts.App.POST(path, handler)
	case "PUT":
		hts.App.PUT(path, handler)
	case "DELETE":
		hts.App.DELETE(path, handler)
	case "PATCH":
		hts.App.PATCH(path, handler)
	default:
		panic(fmt.Sprintf("Unsupported HTTP method: %s", method))
	}
}

// ============= 便捷函数 =============

// CreateTestHandler 创建测试处理器
func CreateTestHandler(statusCode int, response any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)

		switch v := response.(type) {
		case string:
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte(v))
		case []byte:
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(v)
		default:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(v)
		}
	}
}

// CreateJSONHandler 创建JSON响应处理器
func CreateJSONHandler(statusCode int, data any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(data)
	}
}

// CreateErrorHandler 创建错误处理器
func CreateErrorHandler(statusCode int, message string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, message, statusCode)
	}
}
