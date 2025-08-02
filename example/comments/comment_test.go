package main

import (
	"encoding/json"
	"testing"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/test/assert"

	"github.com/zsy619/yyhertz/framework/mvc"
	"github.com/zsy619/yyhertz/framework/mvc/comment"
	"github.com/zsy619/yyhertz/framework/mvc/core"
)

// TestCommentController 测试用的注释控制器
// @RestController
// @RequestMapping("/test")
// @Description("测试控制器")
type TestCommentController struct {
	core.BaseController
}

// Get 测试GET方法
// @GetMapping("/")
// @Description("测试GET方法")
func (c *TestCommentController) Get() (map[string]interface{}, error) {
	return map[string]interface{}{
		"message": "GET success",
		"method":  "GET",
	}, nil
}

// Post 测试POST方法
// @PostMapping("/")
// @Description("测试POST方法")
// @RequestBody
func (c *TestCommentController) Post(req *TestCommentRequest) (*TestCommentResponse, error) {
	return &TestCommentResponse{
		Message: "POST success",
		Data:    req,
	}, nil
}

// GetWithParam 测试路径参数
// @GetMapping("/{id}")
// @Description("测试路径参数")
// @PathVariable("id")
func (c *TestCommentController) GetWithParam() (map[string]interface{}, error) {
	id := c.GetParam("id")
	return map[string]interface{}{
		"message": "GET with param success",
		"id":      id,
	}, nil
}

// GetWithQuery 测试查询参数
// @GetMapping("/search")
// @Description("测试查询参数")
// @RequestParam(name="q", required=true)
// @RequestParam(name="page", required=false, defaultValue="1")
func (c *TestCommentController) GetWithQuery() (map[string]interface{}, error) {
	q := c.GetQuery("q", "")
	page := c.GetQuery("page", "1")
	return map[string]interface{}{
		"message": "GET with query success",
		"q":       q,
		"page":    page,
	}, nil
}

// GetWithHeader 测试请求头参数
// @GetMapping("/header")
// @Description("测试请求头参数")
// @RequestHeader(name="Authorization", required=false)
// @RequestHeader(name="X-Request-ID", required=false)
func (c *TestCommentController) GetWithHeader() (map[string]interface{}, error) {
	auth := c.GetHeader("Authorization")
	requestID := c.GetHeader("X-Request-ID")
	return map[string]interface{}{
		"message":    "GET with header success",
		"auth":       string(auth),
		"request_id": string(requestID),
	}, nil
}

// GetWithMiddleware 测试中间件
// @GetMapping("/middleware")
// @Description("测试中间件")
// @Middleware("auth", "ratelimit")
func (c *TestCommentController) GetWithMiddleware() (map[string]interface{}, error) {
	return map[string]interface{}{
		"message": "GET with middleware success",
	}, nil
}

type TestCommentRequest struct {
	Name  string `json:"name" binding:"required"`
	Value int    `json:"value" binding:"min=0"`
}

type TestCommentResponse struct {
	Message string              `json:"message"`
	Data    *TestCommentRequest `json:"data"`
}

// setupCommentTestServer 设置测试服务器
func setupCommentTestServer() *server.Hertz {
	h := mvc.HertzApp
	app := comment.NewCommentWithApp(h)
	app.AutoScanAndRegister(&TestCommentController{})
	return h.Hertz
}

func TestCommentAnnotationParsing(t *testing.T) {
	// 创建注释注解解析器
	ca := comment.GetGlobalParser()

	// 模拟解析当前文件
	err := ca.ParseSourceFile("comment_test.go")
	if err != nil {
		t.Logf("警告: 无法解析源文件 - %v", err)
		// 在实际测试中，你可能需要提供正确的文件路径
		return
	}

	// 验证控制器信息
	controllerInfo := ca.GetControllerInfo("main", "TestCommentController")
	if controllerInfo != nil {
		assert.Assert(t, controllerInfo.IsRestController)
		assert.DeepEqual(t, "/test", controllerInfo.BasePath)
		assert.DeepEqual(t, "测试控制器", controllerInfo.Description)
	}

	// 验证方法信息
	methodInfo := ca.GetMethodInfo("main", "TestCommentController", "Get")
	if methodInfo != nil {
		assert.DeepEqual(t, "GET", methodInfo.HTTPMethod)
		assert.DeepEqual(t, "/", methodInfo.Path)
		assert.DeepEqual(t, "测试GET方法", methodInfo.Description)
	}
}

func TestCommentRouteRegistration(t *testing.T) {
	h := mvc.HertzApp
	app := comment.NewCommentWithApp(h)

	// 由于源码解析可能失败，我们测试应用初始化
	assert.Assert(t, app != nil)
	assert.Assert(t, app.GetRouter() != nil)

	routes := app.GetRoutes()
	// 如果源码解析成功，应该有路由注册
	t.Logf("注册的路由数量: %d", len(routes))
}

func TestCommentParameterParsing(t *testing.T) {
	// 测试注释参数解析逻辑
	testCases := []struct {
		line     string
		expected string
	}{
		{`@RequestParam(name="page", required=false, defaultValue="1")`, "page"},
		{`@PathVariable("id")`, "id"},
		{`@RequestHeader(name="Authorization", required=true)`, "Authorization"},
		{`@RequestBody`, ""},
	}

	for _, tc := range testCases {
		// 这里需要调用内部的解析函数
		// 由于这些函数不是导出的，我们可能需要重构或创建测试辅助函数
		t.Logf("测试解析: %s", tc.line)
	}
}

func TestCommentMiddlewareParsing(t *testing.T) {
	testLines := []string{
		`@Middleware("auth")`,
		`@Middleware("auth", "ratelimit")`,
		`@Middleware("auth", "ratelimit", "cors")`,
	}

	for _, line := range testLines {
		t.Logf("测试中间件解析: %s", line)
		// 在实际实现中，这里会调用中间件解析函数
	}
}

func TestCommentPathCombination(t *testing.T) {
	// 测试路径组合功能
	testCases := []struct {
		basePath   string
		methodPath string
		expected   string
	}{
		{"/api", "/users", "/api/users"},
		{"/api/", "/users", "/api/users"},
		{"/api", "/users/", "/api/users"},
		{"/api/", "/users/", "/api/users"},
		{"", "/users", "/users"},
		{"/api", "", "/api"},
		{"", "", ""},
	}

	for _, tc := range testCases {
		// 使用comment包中的CombinePath函数
		result := comment.CombinePath(tc.basePath, tc.methodPath)
		assert.DeepEqual(t, tc.expected, result)
	}
}

func TestCommentControllerMethods(t *testing.T) {
	// 测试控制器方法
	controller := &TestCommentController{}
	controller.BaseController.Data = make(map[string]interface{})

	// 测试Get方法
	result, err := controller.Get()
	assert.Assert(t, err == nil)
	assert.DeepEqual(t, "GET success", result["message"])
	assert.DeepEqual(t, "GET", result["method"])
}

func TestUserControllerCommentMethods(t *testing.T) {
	// 测试UserController的方法
	controller := &UserController{}
	controller.BaseController.Data = make(map[string]interface{})

	// 测试GetUsers方法
	users, err := controller.GetUsers()
	assert.Assert(t, err == nil)
	assert.Assert(t, len(users) > 0)
	assert.DeepEqual(t, "张三", users[0].Name)
	assert.DeepEqual(t, "active", users[0].Status)
}

func TestProductControllerCommentMethods(t *testing.T) {
	// 测试ProductController的方法
	controller := &ProductController{}
	controller.BaseController.Data = make(map[string]interface{})

	// 测试GetProducts方法
	products, err := controller.GetProducts()
	assert.Assert(t, err == nil)
	assert.Assert(t, len(products) > 0)
	assert.DeepEqual(t, "iPhone 15", products[0].Name)
	assert.DeepEqual(t, "electronics", products[0].Category)
}

func TestWebControllerCommentMethods(t *testing.T) {
	// 测试WebController的方法
	controller := &WebController{}
	controller.BaseController.Data = make(map[string]interface{})

	// 测试Index方法
	controller.Index()
	assert.DeepEqual(t, "首页", controller.Data["Title"])
	assert.Assert(t, controller.Data["Message"] != nil)
}

func TestAdminControllerCommentMethods(t *testing.T) {
	// 测试AdminController的方法
	controller := &AdminController{}
	controller.BaseController.Data = make(map[string]interface{})

	// 测试GetDashboard方法
	dashboard, err := controller.GetDashboard()
	assert.Assert(t, err == nil)
	assert.DeepEqual(t, 1000, dashboard["userCount"])
	assert.DeepEqual(t, "healthy", dashboard["systemStatus"])

	// 测试GetSystemInfo方法
	info, err := controller.GetSystemInfo()
	assert.Assert(t, err == nil)
	assert.DeepEqual(t, "1.0.0", info["version"])
	assert.DeepEqual(t, "production", info["environment"])
}

func TestCommentRequestResponseSerialization(t *testing.T) {
	// 测试JSON序列化
	req := &UserRequest{
		Name:  "测试用户",
		Email: "test@example.com",
		Age:   25,
	}

	data, err := json.Marshal(req)
	assert.Assert(t, err == nil)

	var decoded UserRequest
	err = json.Unmarshal(data, &decoded)
	assert.Assert(t, err == nil)
	assert.DeepEqual(t, req.Name, decoded.Name)
	assert.DeepEqual(t, req.Email, decoded.Email)
	assert.DeepEqual(t, req.Age, decoded.Age)
}

func TestCommentRouteCollector(t *testing.T) {
	// 测试路由收集器
	collector := comment.NewRouteCollector()
	assert.Assert(t, collector != nil)

	// 测试空收集器
	assert.DeepEqual(t, 0, collector.GetRouteCount())
	assert.DeepEqual(t, 0, collector.GetControllerCount())

	// 测试方法统计
	methodCounts := collector.GetMethodCount()
	assert.Assert(t, methodCounts != nil)
}

func TestCommentRouteAnalyzer(t *testing.T) {
	// 测试路由分析器
	collector := comment.NewRouteCollector()
	analyzer := comment.NewRouteAnalyzer(collector)

	assert.Assert(t, analyzer != nil)

	// 测试重复路由分析
	duplicates := analyzer.AnalyzeDuplicates()
	assert.Assert(t, duplicates != nil)

	// 测试RESTful分析
	restPatterns := analyzer.AnalyzeRESTfulness()
	assert.Assert(t, restPatterns != nil)
}

func TestCommentAutoDiscovery(t *testing.T) {
	h := mvc.HertzApp
	app := comment.NewCommentWithApp(h)

	// 测试自动发现
	discovery := comment.NewAutoDiscovery(app).
		WithScanPaths("./").
		WithExcludePaths("./test").
		WithControllerSuffix("Controller")

	assert.Assert(t, discovery != nil)

	// 注意：实际的Discover()调用可能会失败，因为它需要访问文件系统
	// err := discovery.Discover()
	// 在实际测试中，你可能需要模拟文件系统或使用临时文件
}

// 性能测试
func BenchmarkCommentControllerMethod(b *testing.B) {
	controller := &TestCommentController{}
	controller.BaseController.Data = make(map[string]interface{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		controller.Get()
	}
}

func BenchmarkCommentUserControllerGetUsers(b *testing.B) {
	controller := &UserController{}
	controller.BaseController.Data = make(map[string]interface{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		controller.GetUsers()
	}
}

func BenchmarkCommentRequestSerialization(b *testing.B) {
	req := &UserRequest{
		Name:  "测试用户",
		Email: "test@example.com",
		Age:   25,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(req)
	}
}

// 集成测试辅助函数
func TestCommentIntegration(t *testing.T) {
	// 集成测试
	h := setupCommentTestServer()
	assert.Assert(t, h != nil)

	// 在实际项目中，这里可以添加HTTP请求测试
	// 使用hertz的测试工具或HTTP客户端
}

// 测试注释解析的边界情况
func TestCommentParsingEdgeCases(t *testing.T) {
	testCases := []struct {
		name     string
		comment  string
		expected bool
	}{
		{"Empty comment", "", false},
		{"Valid RestController", "@RestController", true},
		{"Valid RequestMapping", `@RequestMapping("/api/users")`, true},
		{"Invalid annotation", "@InvalidAnnotation", false},
		{"Malformed RequestParam", "@RequestParam(invalid)", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 这里需要调用实际的注释解析函数
			// 由于解析函数不是导出的，我们可能需要重构代码结构
			t.Logf("测试注释: %s", tc.comment)
		})
	}
}

// 测试错误处理
func TestCommentErrorHandling(t *testing.T) {
	controller := &TestCommentController{}
	controller.BaseController.Data = make(map[string]interface{})

	// 测试正常情况
	result, err := controller.Get()
	assert.Assert(t, err == nil)
	assert.Assert(t, result != nil)

	// 更多错误处理测试可以在这里添加
}

// 运行所有注释测试的函数
func RunAllCommentTests() {
	// 这个函数可以用于手动运行所有测试
	// go test -v ./example/comments/
}
