package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/zsy619/yyhertz/framework/mvc/codegen"
)

func main() {
	var (
		projectRoot = flag.String("root", ".", "项目根目录")
		genType     = flag.String("type", "all", "生成类型: all, routes, docs, client")
		help        = flag.Bool("help", false, "显示帮助信息")
	)
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	// 获取绝对路径
	absRoot, err := filepath.Abs(*projectRoot)
	if err != nil {
		fmt.Printf("错误: 无法获取项目根目录的绝对路径: %v\n", err)
		os.Exit(1)
	}

	// 检查项目根目录是否存在
	if _, err := os.Stat(absRoot); os.IsNotExist(err) {
		fmt.Printf("错误: 项目根目录不存在: %s\n", absRoot)
		os.Exit(1)
	}

	fmt.Printf("项目根目录: %s\n", absRoot)
	fmt.Printf("生成类型: %s\n", *genType)

	// 创建代码生成器
	generator := codegen.NewCodeGenerator(absRoot)

	// 根据类型生成代码
	switch *genType {
	case "all":
		err = generator.GenerateAll()
	case "routes":
		err = generator.GenerateRoutes()
	case "docs":
		err = generator.GenerateDocs()
	case "client":
		err = generator.GenerateClient()
	default:
		fmt.Printf("错误: 不支持的生成类型: %s\n", *genType)
		showHelp()
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("代码生成成功！")
}

func showHelp() {
	fmt.Println("YYHertz 代码生成工具")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  codegen [选项]")
	fmt.Println()
	fmt.Println("选项:")
	fmt.Println("  -root string")
	fmt.Println("        项目根目录 (默认: \".\")")
	fmt.Println("  -type string")
	fmt.Println("        生成类型: all, routes, docs, client (默认: \"all\")")
	fmt.Println("  -help")
	fmt.Println("        显示帮助信息")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  codegen -root ./myproject -type all")
	fmt.Println("  codegen -type routes")
	fmt.Println("  codegen -type docs")
	fmt.Println("  codegen -type client")
	fmt.Println()
	fmt.Println("生成类型说明:")
	fmt.Println("  all     - 生成所有代码（路由、文档、客户端）")
	fmt.Println("  routes  - 仅生成路由注册代码")
	fmt.Println("  docs    - 仅生成API文档")
	fmt.Println("  client  - 仅生成客户端SDK代码")
}