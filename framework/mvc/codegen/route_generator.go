package codegen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// RouteGenerator 路由生成器
type RouteGenerator struct {
	ProjectRoot   string
	OutputFile    string
	ControllerDir string
	PackageName   string
}

// ControllerInfo 控制器信息
type ControllerInfo struct {
	Name       string
	Package    string
	Methods    []MethodInfo
	Middleware []string
	Prefix     string
}

// MethodInfo 方法信息
type MethodInfo struct {
	Name       string
	HTTPMethod string
	Path       string
	Params     []ParamInfo
	Returns    []string
	Comment    string
}

// ParamInfo 参数信息
type ParamInfo struct {
	Name string
	Type string
	Tag  string
}

// NewRouteGenerator 创建路由生成器
func NewRouteGenerator(projectRoot, controllerDir string) *RouteGenerator {
	return &RouteGenerator{
		ProjectRoot:   projectRoot,
		ControllerDir: controllerDir,
		OutputFile:    "routes_generated.go",
		PackageName:   "routes",
	}
}

// Generate 生成路由代码
func (rg *RouteGenerator) Generate() error {
	controllers, err := rg.scanControllers()
	if err != nil {
		return fmt.Errorf("扫描控制器失败: %v", err)
	}

	return rg.generateRouteFile(controllers)
}

// scanControllers 扫描控制器
func (rg *RouteGenerator) scanControllers() ([]ControllerInfo, error) {
	var controllers []ControllerInfo

	err := filepath.Walk(rg.ControllerDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		ctrl, err := rg.parseController(path)
		if err != nil {
			return err
		}

		if ctrl != nil {
			controllers = append(controllers, *ctrl)
		}

		return nil
	})

	return controllers, err
}

// parseController 解析控制器文件
func (rg *RouteGenerator) parseController(filePath string) (*ControllerInfo, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var ctrl *ControllerInfo

	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.TypeSpec:
			if rg.isController(x) {
				ctrl = &ControllerInfo{
					Name:    x.Name.Name,
					Package: node.Name.Name,
					Methods: []MethodInfo{},
				}

				// 解析控制器注解
				if x.Doc != nil {
					ctrl.Middleware, ctrl.Prefix = rg.parseControllerAnnotations(x.Doc.Text())
				}
			}
		case *ast.FuncDecl:
			if ctrl != nil && rg.isControllerMethod(x) {
				method := rg.parseMethod(x)
				if method != nil {
					ctrl.Methods = append(ctrl.Methods, *method)
				}
			}
		}
		return true
	})

	return ctrl, nil
}

// isController 判断是否为控制器
func (rg *RouteGenerator) isController(ts *ast.TypeSpec) bool {
	if structType, ok := ts.Type.(*ast.StructType); ok {
		for _, field := range structType.Fields.List {
			if len(field.Names) == 0 {
				// 匿名字段，检查是否嵌入BaseController
				if ident, ok := field.Type.(*ast.SelectorExpr); ok {
					if x, ok := ident.X.(*ast.Ident); ok && x.Name == "mvc" {
						if ident.Sel.Name == "BaseController" {
							return true
						}
					}
				}
			}
		}
	}
	return false
}

// isControllerMethod 判断是否为控制器方法
func (rg *RouteGenerator) isControllerMethod(fn *ast.FuncDecl) bool {
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return false
	}

	// 检查方法是否为公开方法
	return fn.Name.IsExported()
}

// parseMethod 解析方法
func (rg *RouteGenerator) parseMethod(fn *ast.FuncDecl) *MethodInfo {
	method := &MethodInfo{
		Name:    fn.Name.Name,
		Params:  []ParamInfo{},
		Returns: []string{},
	}

	// 解析注释中的路由信息
	if fn.Doc != nil {
		method.HTTPMethod, method.Path, method.Comment = rg.parseMethodAnnotations(fn.Doc.Text())
	}

	// 如果没有注解，使用默认规则
	if method.HTTPMethod == "" {
		method.HTTPMethod, method.Path = rg.inferRouteFromMethod(method.Name)
	}

	// 解析参数
	if fn.Type.Params != nil {
		for _, param := range fn.Type.Params.List {
			for _, name := range param.Names {
				paramInfo := ParamInfo{
					Name: name.Name,
					Type: rg.typeToString(param.Type),
				}
				method.Params = append(method.Params, paramInfo)
			}
		}
	}

	return method
}

// parseControllerAnnotations 解析控制器注解
func (rg *RouteGenerator) parseControllerAnnotations(comment string) ([]string, string) {
	var middleware []string
	var prefix string

	lines := strings.Split(comment, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "@Middleware") {
			mw := strings.TrimPrefix(line, "@Middleware")
			mw = strings.Trim(mw, " ()")
			middleware = strings.Split(mw, ",")
		} else if strings.HasPrefix(line, "@Prefix") {
			prefix = strings.TrimPrefix(line, "@Prefix")
			prefix = strings.Trim(prefix, " ()")
		}
	}

	return middleware, prefix
}

// parseMethodAnnotations 解析方法注解
func (rg *RouteGenerator) parseMethodAnnotations(comment string) (string, string, string) {
	var httpMethod, path, desc string

	lines := strings.Split(comment, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "@Route") {
			route := strings.TrimPrefix(line, "@Route")
			route = strings.Trim(route, " ()")
			parts := strings.SplitN(route, " ", 2)
			if len(parts) == 2 {
				httpMethod = parts[0]
				path = parts[1]
			}
		} else if strings.HasPrefix(line, "//") {
			if desc == "" {
				desc = strings.TrimPrefix(line, "//")
				desc = strings.TrimSpace(desc)
			}
		}
	}

	return httpMethod, path, desc
}

// inferRouteFromMethod 从方法名推断路由
func (rg *RouteGenerator) inferRouteFromMethod(methodName string) (string, string) {
	// RESTful 约定
	switch {
	case strings.HasPrefix(methodName, "Get"):
		return "GET", "/" + strings.ToLower(strings.TrimPrefix(methodName, "Get"))
	case strings.HasPrefix(methodName, "Post"):
		return "POST", "/" + strings.ToLower(strings.TrimPrefix(methodName, "Post"))
	case strings.HasPrefix(methodName, "Put"):
		return "PUT", "/" + strings.ToLower(strings.TrimPrefix(methodName, "Put"))
	case strings.HasPrefix(methodName, "Delete"):
		return "DELETE", "/" + strings.ToLower(strings.TrimPrefix(methodName, "Delete"))
	case methodName == "Index":
		return "GET", "/"
	default:
		return "GET", "/" + strings.ToLower(methodName)
	}
}

// typeToString 类型转字符串
func (rg *RouteGenerator) typeToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return rg.typeToString(t.X) + "." + t.Sel.Name
	case *ast.StarExpr:
		return "*" + rg.typeToString(t.X)
	default:
		return "interface{}"
	}
}

// generateRouteFile 生成路由文件
func (rg *RouteGenerator) generateRouteFile(controllers []ControllerInfo) error {
	tmpl := `// Code generated by RouteGenerator. DO NOT EDIT.
package {{.PackageName}}

import (
	"github.com/zsy619/yyhertz/framework/mvc"
	"github.com/zsy619/yyhertz/framework/mvc/register"
)

// RegisterRoutes 注册所有路由
func RegisterRoutes(app *mvc.App) {
{{range .Controllers}}
	// {{.Name}} 路由
	{{.Name|lower}}Ctrl := &{{.Package}}.{{.Name}}{}
	{{if .Prefix}}
	app.RegisterControllerWithPrefix("{{.Prefix}}", {{.Name|lower}}Ctrl)
	{{else}}
	app.RegisterController({{.Name|lower}}Ctrl)
	{{end}}
	
	{{range .Methods}}
	// {{.Comment}}
	app.MapRoutes({{$.Name|lower}}Ctrl, "{{.Name}}", "{{.HTTPMethod}}:{{.Path}}")
	{{end}}
{{end}}
}

// GetRouteInfo 获取路由信息
func GetRouteInfo() map[string]interface{} {
	return map[string]interface{}{
		"controllers": []map[string]interface{}{
{{range .Controllers}}
			{
				"name": "{{.Name}}",
				"package": "{{.Package}}",
				"prefix": "{{.Prefix}}",
				"middleware": []string{ {{range .Middleware}}"{{.}}",{{end}} },
				"methods": []map[string]interface{}{
{{range .Methods}}
					{
						"name": "{{.Name}}",
						"method": "{{.HTTPMethod}}",
						"path": "{{.Path}}",
						"comment": "{{.Comment}}",
					},
{{end}}
				},
			},
{{end}}
		},
	}
}
`

	funcMap := template.FuncMap{
		"lower": strings.ToLower,
	}

	t, err := template.New("routes").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return err
	}

	outputPath := filepath.Join(rg.ProjectRoot, rg.OutputFile)
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	data := struct {
		PackageName string
		Controllers []ControllerInfo
	}{
		PackageName: rg.PackageName,
		Controllers: controllers,
	}

	return t.Execute(file, data)
}
