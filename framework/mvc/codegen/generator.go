package codegen

import (
	"fmt"
	"path/filepath"
)

// CodeGenerator 代码生成器主入口
type CodeGenerator struct {
	ProjectRoot   string
	ControllerDir string
	OutputDir     string
}

// NewCodeGenerator 创建代码生成器
func NewCodeGenerator(projectRoot string) *CodeGenerator {
	return &CodeGenerator{
		ProjectRoot:   projectRoot,
		ControllerDir: filepath.Join(projectRoot, "controller"),
		OutputDir:     filepath.Join(projectRoot, "generated"),
	}
}

// GenerateAll 生成所有代码
func (cg *CodeGenerator) GenerateAll() error {
	fmt.Println("开始扫描控制器...")

	// 扫描控制器
	routeGen := NewRouteGenerator(cg.ProjectRoot, cg.ControllerDir)
	controllers, err := routeGen.scanControllers()
	if err != nil {
		return fmt.Errorf("扫描控制器失败: %v", err)
	}

	fmt.Printf("发现 %d 个控制器\n", len(controllers))

	// 生成路由代码
	fmt.Println("生成路由代码...")
	if err := routeGen.Generate(); err != nil {
		return fmt.Errorf("生成路由代码失败: %v", err)
	}

	// 生成API文档
	fmt.Println("生成API文档...")
	docGen := NewDocGenerator(cg.ProjectRoot)
	if err := docGen.Generate(controllers); err != nil {
		return fmt.Errorf("生成API文档失败: %v", err)
	}

	// 生成客户端代码
	fmt.Println("生成客户端代码...")
	clientGen := NewClientGenerator(cg.ProjectRoot)
	if err := clientGen.Generate(controllers); err != nil {
		return fmt.Errorf("生成客户端代码失败: %v", err)
	}

	fmt.Println("代码生成完成！")
	return nil
}

// GenerateRoutes 仅生成路由代码
func (cg *CodeGenerator) GenerateRoutes() error {
	routeGen := NewRouteGenerator(cg.ProjectRoot, cg.ControllerDir)
	return routeGen.Generate()
}

// GenerateDocs 仅生成API文档
func (cg *CodeGenerator) GenerateDocs() error {
	routeGen := NewRouteGenerator(cg.ProjectRoot, cg.ControllerDir)
	controllers, err := routeGen.scanControllers()
	if err != nil {
		return err
	}

	docGen := NewDocGenerator(cg.ProjectRoot)
	return docGen.Generate(controllers)
}

// GenerateClient 仅生成客户端代码
func (cg *CodeGenerator) GenerateClient() error {
	routeGen := NewRouteGenerator(cg.ProjectRoot, cg.ControllerDir)
	controllers, err := routeGen.scanControllers()
	if err != nil {
		return err
	}

	clientGen := NewClientGenerator(cg.ProjectRoot)
	return clientGen.Generate(controllers)
}
