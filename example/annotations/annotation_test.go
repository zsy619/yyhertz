package main

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/test/assert"

	"github.com/zsy619/yyhertz/framework/mvc/annotation"
	"github.com/zsy619/yyhertz/framework/mvc/core"
)

// TestController 测试控制器
type TestController struct {
	core.BaseController `rest:"" mapping:"/test"`
}

func init() {
	testType := reflect.TypeOf((*TestController)(nil)).Elem()

	annotation.RegisterGetMethod(testType, "Get", "/").
		WithDescription("测试GET方法")

	annotation.RegisterPostMethod(testType, "Post", "/").
		WithDescription("测试POST方法").
		WithBodyParam(true)

	annotation.RegisterGetMethod(testType, "GetWithParam", "/{id}").
		WithDescription("测试路径参数").
		WithPathParam("id", true)

	annotation.RegisterGetMethod(testType, "GetWithQuery", "/search").
		WithDescription("测试查询参数").
		WithQueryParam("q", true, "").
		WithQueryParam("page", false, "1")
}

func (c *TestController) Get() (map[string]interface{}, error) {
	return map[string]interface{}{
		"message": "GET success",
		"method":  "GET",
	}, nil
}

func (c *TestController) Post(req *TestRequest) (*TestResponse, error) {
	return &TestResponse{
		Message: "POST success",
		Data:    req,
	}, nil
}

func (c *TestController) GetWithParam() (map[string]interface{}, error) {
	id := c.GetParam("id")
	return map[string]interface{}{
		"message": "GET with param success",
		"id":      id,
	}, nil
}

func (c *TestController) GetWithQuery() (map[string]interface{}, error) {
	q := c.GetQuery("q", "")
	page := c.GetQuery("page", "1")
	return map[string]interface{}{
		"message": "GET with query success",
		"q":       q,
		"page":    page,
	}, nil
}

type TestRequest struct {
	Name  string `json:"name" binding:"required"`
	Value int    `json:"value" binding:"min=0"`
}

type TestResponse struct {
	Message string       `json:"message"`
	Data    *TestRequest `json:"data"`
}

// 为测试添加缺失的结构体定义
type UserController struct {
	core.BaseController `rest:"" mapping:"/api/users"`
}

type ProductController struct {
	core.BaseController `controller:"" mapping:"/api/products"`
}

type AdminController struct {
	core.BaseController `rest:"" mapping:"/api/admin"`
}

type UserRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
	Age   int    `json:"age" binding:"min=0,max=120"`
}

type UserResponse struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

type ProductResponse struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

// 添加控制器方法实现
func (c *UserController) GetUsers() ([]*UserResponse, error) {
	return []*UserResponse{
		{ID: 1, Name: "张三", Email: "zhang@example.com", Age: 25},
	}, nil
}

func (c *ProductController) GetProducts() ([]*ProductResponse, error) {
	return []*ProductResponse{
		{ID: 1, Name: "iPhone 15", Price: 8999.0},
	}, nil
}

func (c *AdminController) GetDashboard() (map[string]interface{}, error) {
	return map[string]interface{}{
		"userCount":    1000,
		"productCount": 500,
		"orderCount":   2000,
	}, nil
}

// setupTestServer 设置测试服务器
func setupTestServer() *server.Hertz {
	h := core.NewApp()
	app := annotation.NewAnnotationWithApp(h)
	app.AutoRegister(&TestController{})
	return h.Hertz
}

func TestAnnotationRouteRegistration(t *testing.T) {
	h := core.NewApp()
	app := annotation.NewAnnotationWithApp(h)
	app.AutoRegister(&TestController{})

	routes := app.GetAnnotatedRoutes()

	// 验证路由数量
	expectedRoutes := 4
	assert.DeepEqual(t, expectedRoutes, len(routes))

	// 验证路由信息
	routeMap := make(map[string]*annotation.RouteInfo)
	for _, route := range routes {
		key := route.HTTPMethod + " " + route.Path
		routeMap[key] = route
	}

	// 验证GET /test
	getRoute := routeMap["GET /test"]
	assert.Assert(t, getRoute != nil)
	assert.DeepEqual(t, "Get", getRoute.MethodName)

	// 验证POST /test
	postRoute := routeMap["POST /test"]
	assert.Assert(t, postRoute != nil)
	assert.DeepEqual(t, "Post", postRoute.MethodName)

	// 验证GET /test/{id}
	getParamRoute := routeMap["GET /test/{id}"]
	assert.Assert(t, getParamRoute != nil)
	assert.DeepEqual(t, "GetWithParam", getParamRoute.MethodName)

	// 验证GET /test/search
	getQueryRoute := routeMap["GET /test/search"]
	assert.Assert(t, getQueryRoute != nil)
	assert.DeepEqual(t, "GetWithQuery", getQueryRoute.MethodName)
}

func TestStructTagParsing(t *testing.T) {
	controllerType := reflect.TypeOf((*TestController)(nil)).Elem()

	// 测试struct标签解析
	info, err := annotation.ParseControllerTags(controllerType)
	assert.Assert(t, err == nil)
	assert.Assert(t, info != nil)
	assert.DeepEqual(t, "TestController", info.Name)
	assert.DeepEqual(t, "/test", info.BasePath)
	assert.Assert(t, info.IsRest)
}

func TestMethodRegistryFunctionality(t *testing.T) {
	registry := annotation.GetRegistry()
	testType := reflect.TypeOf((*TestController)(nil)).Elem()

	// 测试方法映射获取
	mapping := registry.GetMethodMapping(testType, "Get")
	assert.Assert(t, mapping != nil)
	assert.DeepEqual(t, "GET", mapping.HTTPMethod)
	assert.DeepEqual(t, "/", mapping.Path)

	// 测试控制器所有映射获取
	mappings := registry.GetControllerMappings(testType)
	assert.Assert(t, len(mappings) >= 4)
}

func TestAnnotationBuilder(t *testing.T) {
	testType := reflect.TypeOf((*TestController)(nil)).Elem()

	// 测试链式构建
	builder := annotation.RegisterGetMethod(testType, "TestMethod", "/test-builder").
		WithDescription("测试构建器").
		WithQueryParam("param1", true, "").
		WithQueryParam("param2", false, "default").
		WithMiddleware("middleware1", "middleware2")

	mapping := builder.Build()

	assert.DeepEqual(t, "GET", mapping.HTTPMethod)
	assert.DeepEqual(t, "/test-builder", mapping.Path)
	assert.DeepEqual(t, "测试构建器", mapping.Description)
	assert.DeepEqual(t, 2, len(mapping.Params))
	assert.DeepEqual(t, 2, len(mapping.Middlewares))
}

func TestPathCombination(t *testing.T) {
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
		{"/", "/users", "//users"},
		{"/api", "/", "/api"},
	}

	for _, tc := range testCases {
		result := annotation.CombinePath(tc.basePath, tc.methodPath)
		assert.DeepEqual(t, tc.expected, result)
	}
}

func TestParameterMapping(t *testing.T) {
	testType := reflect.TypeOf((*TestController)(nil)).Elem()

	// 测试参数映射
	builder := annotation.RegisterPostMethod(testType, "TestParams", "/test-params").
		WithPathParam("id", true).
		WithQueryParam("search", false, "").
		WithBodyParam(true).
		WithHeaderParam("Authorization", true, "").
		WithCookieParam("session", false, "")

	mapping := builder.Build()

	assert.DeepEqual(t, 5, len(mapping.Params))

	// 验证参数类型
	paramSources := make(map[annotation.ParamSource]int)
	for _, param := range mapping.Params {
		paramSources[param.Source]++
	}

	assert.DeepEqual(t, 1, paramSources[annotation.ParamSourcePath])
	assert.DeepEqual(t, 1, paramSources[annotation.ParamSourceQuery])
	assert.DeepEqual(t, 1, paramSources[annotation.ParamSourceBody])
	assert.DeepEqual(t, 1, paramSources[annotation.ParamSourceHeader])
	assert.DeepEqual(t, 1, paramSources[annotation.ParamSourceCookie])
}

func TestMiddlewareRegistration(t *testing.T) {
	testType := reflect.TypeOf((*TestController)(nil)).Elem()

	builder := annotation.RegisterGetMethod(testType, "TestMiddleware", "/test-middleware").
		WithMiddleware("auth").
		WithMiddleware("ratelimit", "cors")

	mapping := builder.Build()

	expected := []string{"auth", "ratelimit", "cors"}
	assert.DeepEqual(t, expected, mapping.Middlewares)
}

func TestAnnotationAppInitialization(t *testing.T) {
	// 创建独立的应用实例，避免路由冲突
	h := core.NewApp()
	app := annotation.NewAnnotationWithApp(h)

	assert.Assert(t, app != nil)
	assert.Assert(t, app.GetAutoRouter() != nil)

	// 测试控制器注册
	app.AutoRegister(&TestController{})

	routes := app.GetAnnotatedRoutes()
	assert.Assert(t, len(routes) > 0)
}

// 集成测试
func TestIntegrationBasicRoutes(t *testing.T) {
	// 注意：这些是集成测试，需要启动完整的服务器
	// 在实际项目中，你可能需要使用testify或其他测试框架
	// 来进行更完整的HTTP请求测试

	h := setupTestServer()

	// 这里可以添加HTTP客户端测试
	// 例如使用 h.Test() 方法或创建测试HTTP客户端

	assert.Assert(t, h != nil)
}

func TestUserControllerMethods(t *testing.T) {
	// 测试UserController的方法
	controller := &UserController{}

	// 初始化BaseController
	controller.BaseController.Data = make(map[string]interface{})

	// 测试GetUsers方法
	users, err := controller.GetUsers()
	assert.Assert(t, err == nil)
	assert.Assert(t, len(users) > 0)
	assert.DeepEqual(t, "张三", users[0].Name)
}

func TestProductControllerMethods(t *testing.T) {
	// 测试ProductController的方法
	controller := &ProductController{}

	// 初始化BaseController
	controller.BaseController.Data = make(map[string]interface{})

	// 测试GetProducts方法
	products, err := controller.GetProducts()
	assert.Assert(t, err == nil)
	assert.Assert(t, len(products) > 0)
	assert.DeepEqual(t, "iPhone 15", products[0].Name)
}

func TestAdminControllerMethods(t *testing.T) {
	// 测试AdminController的方法
	controller := &AdminController{}

	// 初始化BaseController
	controller.BaseController.Data = make(map[string]interface{})

	// 测试GetDashboard方法
	dashboard, err := controller.GetDashboard()
	assert.Assert(t, err == nil)
	assert.Assert(t, dashboard["userCount"] != nil)
	assert.DeepEqual(t, 1000, dashboard["userCount"])
}

func TestRequestResponseSerialization(t *testing.T) {
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

// 基准测试
func BenchmarkRouteRegistration(b *testing.B) {
	testType := reflect.TypeOf((*TestController)(nil)).Elem()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		annotation.RegisterGetMethod(testType, "BenchmarkMethod", "/benchmark").
			WithDescription("基准测试方法").
			WithQueryParam("param", false, "default")
	}
}

func BenchmarkControllerRegistration(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h := core.NewApp() // 每次迭代创建新的app实例，避免路由冲突
		app := annotation.NewAnnotationWithApp(h)
		app.AutoRegister(&TestController{})
	}
}

// 运行所有测试的函数
func RunAllAnnotationTests() {
	// 这个函数可以用于手动运行所有测试
	// go test -v ./example/annotations/
}
