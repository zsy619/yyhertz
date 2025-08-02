package main

import (
	"log"
	"reflect"

	"github.com/zsy619/yyhertz/framework/mvc"
	"github.com/zsy619/yyhertz/framework/mvc/annotation"
	"github.com/zsy619/yyhertz/framework/mvc/comment"
	"github.com/zsy619/yyhertz/framework/mvc/core"
)

// HybridDemoController 混合演示控制器
// @RestController
// @RequestMapping("/api/demo")
type HybridDemoController struct {
	core.BaseController `rest:"" mapping:"/api/demo"`
}

// 在 init() 中注册方法映射
func init() {
	demoType := reflect.TypeOf((*HybridDemoController)(nil)).Elem()

	// 使用 annotation.RegisterGetMethod 注册
	annotation.RegisterGetMethod(demoType, "GetByInit", "/init").
		WithDescription("通过init()注册的方法").
		WithQueryParam("page", false, "1").
		WithQueryParam("size", false, "10").
		WithQueryParam("keyword", false, "")
}

// GetByInit 通过 init() 注册的方法 (annotation 方式)
func (c *HybridDemoController) GetByInit() (map[string]interface{}, error) {
	page := c.GetQuery("page", "1")
	size := c.GetQuery("size", "10")
	keyword := c.GetQuery("keyword", "")

	log.Printf("GetByInit调用: page=%s, size=%s, keyword=%s", page, size, keyword)

	return map[string]interface{}{
		"message":          "通过 annotation.RegisterGetMethod 注册",
		"method":           "GetByInit",
		"registrationType": "struct标签 + init()方法映射",
		"data": map[string]interface{}{
			"page":    page,
			"size":    size,
			"keyword": keyword,
		},
	}, nil
}

// GetByComment 通过 Go 注释注解的方法 (comment 方式)
// @GetMapping("/comment")
// @Description("通过Go注释注解的方法")
// @RequestParam(name="keyword", required=false, defaultValue="")
// @RequestParam(name="category", required=false, defaultValue="all")
func (c *HybridDemoController) GetByComment() (map[string]interface{}, error) {
	keyword := c.GetQuery("keyword", "")
	category := c.GetQuery("category", "all")

	log.Printf("GetByComment调用: keyword=%s, category=%s", keyword, category)

	return map[string]interface{}{
		"message":          "通过 Go 注释注解",
		"method":           "GetByComment",
		"registrationType": "struct标签 + Go注释解析",
		"data": map[string]interface{}{
			"keyword":  keyword,
			"category": category,
		},
	}, nil
}

// PostData 同时演示两种方式都可以处理的POST方法
// @PostMapping("/data")
// @Description("演示POST方法")
// @RequestBody
func (c *HybridDemoController) PostData(req *DemoRequest) (*DemoResponse, error) {
	log.Printf("PostData调用: %+v", req)

	return &DemoResponse{
		Message: "POST请求成功处理",
		Data:    req,
		Info:    "这个方法通过Go注释注解注册",
	}, nil
}

type DemoRequest struct {
	Name  string `json:"name" binding:"required"`
	Value string `json:"value"`
	Type  string `json:"type"`
}

type DemoResponse struct {
	Message string       `json:"message"`
	Data    *DemoRequest `json:"data"`
	Info    string       `json:"info"`
}

func main() {
	log.Println("🚀 启动混合注解演示...")

	// 创建两个应用实例来处理不同的注册方式

	// 1. annotation应用 - 处理通过 init() 注册的方法
	annotationApp := annotation.NewAnnotationWithApp(mvc.HertzApp)
	annotationApp.AutoRegister(&HybridDemoController{})

	// 2. comment应用 - 处理通过 Go 注释注解的方法
	commentApp := comment.NewCommentWithApp(mvc.HertzApp)
	commentApp.AutoScanAndRegister(&HybridDemoController{})

	// 显示注册的路由
	annotationRoutes := annotationApp.GetAnnotatedRoutes()
	commentRoutes := commentApp.GetRoutes()

	log.Printf("📊 注册统计:")
	log.Printf("  通过 init() 注册: %d 个路由", len(annotationRoutes))
	log.Printf("  通过 Go注释 注册: %d 个路由", len(commentRoutes))
	log.Printf("  总计: %d 个路由", len(annotationRoutes)+len(commentRoutes))

	log.Printf("\n📋 Annotation 路由 (init方式):")
	for _, route := range annotationRoutes {
		log.Printf("  %s %s -> %s.%s",
			route.HTTPMethod, route.Path, route.ControllerType.Name(), route.MethodName)
	}

	log.Printf("\n📋 Comment 路由 (注释方式):")
	for _, route := range commentRoutes {
		log.Printf("  %s %s -> %s.%s - %s",
			route.HTTPMethod, route.Path, "HybridDemoController", route.MethodName, route.Description)
	}

	log.Println("\n🌟 混合注解演示服务器启动在 :8889")
	log.Println("\n🔧 测试命令:")
	log.Println("\n# 测试 init() 注册的方法:")
	log.Println("curl -X GET 'http://localhost:8889/api/demo/init?page=1&size=20&keyword=test'")

	log.Println("\n# 测试 Go注释 注册的方法:")
	log.Println("curl -X GET 'http://localhost:8889/api/demo/comment?keyword=demo&category=api'")
	log.Println("curl -X POST http://localhost:8889/api/demo/data -H 'Content-Type: application/json' -d '{\"name\":\"演示\",\"value\":\"混合注解\",\"type\":\"hybrid\"}'")

	log.Println("\n✨ 这个示例展示了如何同时使用:")
	log.Println("  1. struct 标签注解 (`rest:\"\" mapping:\"/api/demo\"`)")
	log.Println("  2. init() 方法映射 (annotation.RegisterGetMethod)")
	log.Println("  3. Go 注释注解 (// @GetMapping, // @PostMapping)")

	mvc.HertzApp.Spin()
}
