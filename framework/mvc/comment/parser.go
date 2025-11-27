package comment

import (
	"go/ast"
	"go/parser" 
	"go/token"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"
)

// Comment-based annotation system using Go comments
// 基于注释的注解系统

// ControllerInfo 基于注释的控制器信息
type ControllerInfo struct {
	PackageName      string            // 包名
	TypeName         string            // 类型名
	IsRestController bool              // 是否为REST控制器
	IsController     bool              // 是否为MVC控制器
	BasePath         string            // 基础路径
	Description      string            // 描述
	Tags             map[string]string // 其他标签
}

// MethodInfo 基于注释的方法信息
type MethodInfo struct {
	PackageName string          // 包名
	TypeName    string          // 类型名
	MethodName  string          // 方法名
	HTTPMethod  string          // HTTP方法
	Path        string          // 路径
	Description string          // 描述
	Params      []*ParamInfo    // 参数信息
	Middlewares []string        // 中间件
	Tags        map[string]string // 标签
}

// ParamInfo 基于注释的参数信息
type ParamInfo struct {
	Name         string      // 参数名
	Source       ParamSource // 参数来源
	Required     bool        // 是否必需
	DefaultValue string      // 默认值
	Description  string      // 描述
}

// ParamSource 参数来源
type ParamSource string

const (
	ParamSourcePath   ParamSource = "path"
	ParamSourceQuery  ParamSource = "query"
	ParamSourceBody   ParamSource = "body"
	ParamSourceHeader ParamSource = "header"
	ParamSourceCookie ParamSource = "cookie"
	ParamSourceForm   ParamSource = "form"
)

// AnnotationParser 注释注解解析器
type AnnotationParser struct {
	ControllerInfos map[string]*ControllerInfo // key: 包名.类型名
	MethodInfos     map[string]*MethodInfo     // key: 包名.类型名.方法名
}

var globalParser = &AnnotationParser{
	ControllerInfos: make(map[string]*ControllerInfo),
	MethodInfos:     make(map[string]*MethodInfo),
}

// GetGlobalParser 获取全局注解解析器实例
func GetGlobalParser() *AnnotationParser {
	return globalParser
}

// NewAnnotationParser 创建新的注解解析器
func NewAnnotationParser() *AnnotationParser {
	return &AnnotationParser{
		ControllerInfos: make(map[string]*ControllerInfo),
		MethodInfos:     make(map[string]*MethodInfo),
	}
}

// ParseSourceFile 解析源文件中的注释注解
func (ap *AnnotationParser) ParseSourceFile(filename string) error {
	fset := token.NewFileSet()
	
	// 解析Go源文件
	file, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return err
	}
	
	packageName := file.Name.Name
	
	// 遍历所有声明
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			// 处理类型声明（struct）
			if d.Tok == token.TYPE {
				ap.parseTypeDecl(d, packageName)
			}
		case *ast.FuncDecl:
			// 处理方法声明
			if d.Recv != nil { // 是方法而不是函数
				ap.parseMethodDecl(d, packageName)
			}
		}
	}
	
	return nil
}

// parseTypeDecl 解析类型声明的注释
func (ap *AnnotationParser) parseTypeDecl(decl *ast.GenDecl, packageName string) {
	for _, spec := range decl.Specs {
		if typeSpec, ok := spec.(*ast.TypeSpec); ok {
			if structType, ok := typeSpec.Type.(*ast.StructType); ok {
				// 解析struct的注释
				ap.parseStructComments(typeSpec.Name.Name, decl.Doc, packageName, structType)
			}
		}
	}
}

// parseStructComments 解析struct注释
func (ap *AnnotationParser) parseStructComments(typeName string, doc *ast.CommentGroup, packageName string, structType *ast.StructType) {
	if doc == nil {
		return
	}
	
	info := &ControllerInfo{
		PackageName: packageName,
		TypeName:    typeName,
		Tags:        make(map[string]string),
	}
	
	// 解析注释文本
	commentText := doc.Text()
	lines := strings.Split(commentText, "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// 解析各种注解
		if matched := parseAnnotation(line, `@RestController`); matched {
			info.IsRestController = true
		} else if matched := parseAnnotation(line, `@Controller`); matched {
			info.IsController = true
		} else if path := parseAnnotationWithValue(line, `@RequestMapping`); path != "" {
			info.BasePath = normalizePath(path)
		} else if desc := parseAnnotationWithValue(line, `@Description`); desc != "" {
			info.Description = desc
		} else if tag := parseTagAnnotation(line); tag != nil {
			info.Tags[tag.Key] = tag.Value
		}
	}
	
	// 只有标记为控制器的才注册
	if info.IsRestController || info.IsController {
		key := packageName + "." + typeName
		ap.ControllerInfos[key] = info
	}
}

// parseMethodDecl 解析方法声明的注释
func (ap *AnnotationParser) parseMethodDecl(decl *ast.FuncDecl, packageName string) {
	if decl.Doc == nil || decl.Recv == nil || len(decl.Recv.List) == 0 {
		return
	}
	
	// 获取接收者类型名
	var typeName string
	switch expr := decl.Recv.List[0].Type.(type) {
	case *ast.StarExpr:
		if ident, ok := expr.X.(*ast.Ident); ok {
			typeName = ident.Name
		}
	case *ast.Ident:
		typeName = expr.Name
	}
	
	if typeName == "" {
		return
	}
	
	// 检查是否是控制器方法
	controllerKey := packageName + "." + typeName
	if _, exists := ap.ControllerInfos[controllerKey]; !exists {
		return
	}
	
	methodName := decl.Name.Name
	
	info := &MethodInfo{
		PackageName: packageName,
		TypeName:    typeName,
		MethodName:  methodName,
		Tags:        make(map[string]string),
		Params:      make([]*ParamInfo, 0),
		Middlewares: make([]string, 0),
	}
	
	// 解析方法注释
	commentText := decl.Doc.Text()
	lines := strings.Split(commentText, "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// 解析HTTP方法注解
		if path := parseAnnotationWithValue(line, `@GetMapping`); path != "" {
			info.HTTPMethod = "GET"
			info.Path = normalizePath(path)
		} else if path := parseAnnotationWithValue(line, `@PostMapping`); path != "" {
			info.HTTPMethod = "POST"
			info.Path = normalizePath(path)
		} else if path := parseAnnotationWithValue(line, `@PutMapping`); path != "" {
			info.HTTPMethod = "PUT"
			info.Path = normalizePath(path)
		} else if path := parseAnnotationWithValue(line, `@DeleteMapping`); path != "" {
			info.HTTPMethod = "DELETE"
			info.Path = normalizePath(path)
		} else if path := parseAnnotationWithValue(line, `@PatchMapping`); path != "" {
			info.HTTPMethod = "PATCH"
			info.Path = normalizePath(path)
		} else if path := parseAnnotationWithValue(line, `@RequestMapping`); path != "" {
			// 如果没有指定HTTP方法，默认为GET
			if info.HTTPMethod == "" {
				info.HTTPMethod = "GET"
			}
			info.Path = normalizePath(path)
		} else if desc := parseAnnotationWithValue(line, `@Description`); desc != "" {
			info.Description = desc
		} else if param := parseParamAnnotation(line); param != nil {
			info.Params = append(info.Params, param)
		} else if middlewares := parseMiddlewareAnnotation(line); len(middlewares) > 0 {
			info.Middlewares = append(info.Middlewares, middlewares...)
		}
	}
	
	// 如果有HTTP方法映射才注册
	if info.HTTPMethod != "" {
		key := packageName + "." + typeName + "." + methodName
		ap.MethodInfos[key] = info
	}
}

// parseAnnotation 解析简单注解（无参数）
func parseAnnotation(line, annotation string) bool {
	pattern := `^\s*` + regexp.QuoteMeta(annotation) + `\s*$`
	matched, _ := regexp.MatchString(pattern, line)
	return matched
}

// parseAnnotationWithValue 解析带值的注解
func parseAnnotationWithValue(line, annotation string) string {
	// 匹配 @Annotation("value") 或 @Annotation(value) 格式
	patterns := []string{
		`^\s*` + regexp.QuoteMeta(annotation) + `\s*\(\s*"([^"]+)"\s*\)\s*$`,
		`^\s*` + regexp.QuoteMeta(annotation) + `\s*\(\s*([^)]+)\s*\)\s*$`,
		`^\s*` + regexp.QuoteMeta(annotation) + `\s*:\s*"([^"]+)"\s*$`,
		`^\s*` + regexp.QuoteMeta(annotation) + `\s*:\s*([^\s]+)\s*$`,
	}
	
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(line); len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}
	
	return ""
}

// TagInfo 标签信息
type TagInfo struct {
	Key   string
	Value string
}

// parseTagAnnotation 解析标签注解
func parseTagAnnotation(line string) *TagInfo {
	// 匹配 @Tag(key="value") 格式
	re := regexp.MustCompile(`^\s*@Tag\s*\(\s*(\w+)\s*=\s*"([^"]+)"\s*\)\s*$`)
	if matches := re.FindStringSubmatch(line); len(matches) > 2 {
		return &TagInfo{
			Key:   matches[1],
			Value: matches[2],
		}
	}
	return nil
}

// parseParamAnnotation 解析参数注解
func parseParamAnnotation(line string) *ParamInfo {
	patterns := map[string]ParamSource{
		`@PathVariable`:   ParamSourcePath,
		`@RequestParam`:   ParamSourceQuery,
		`@RequestBody`:    ParamSourceBody,
		`@RequestHeader`:  ParamSourceHeader,
		`@CookieValue`:    ParamSourceCookie,
	}
	
	for annotation, source := range patterns {
		if value := parseAnnotationWithValue(line, annotation); value != "" {
			param := &ParamInfo{
				Name:   value,
				Source: source,
			}
			
			// 解析额外属性 @RequestParam(name="id", required=true, defaultValue="1")
			if source == ParamSourceQuery || source == ParamSourceHeader || source == ParamSourceCookie {
				param.Required = strings.Contains(line, "required=true")
				if defaultVal := extractDefaultValue(line); defaultVal != "" {
					param.DefaultValue = defaultVal
				}
			} else if source == ParamSourcePath || source == ParamSourceBody {
				param.Required = true // 路径参数和请求体通常是必需的
			}
			
			return param
		}
	}
	
	return nil
}

// extractDefaultValue 提取默认值
func extractDefaultValue(line string) string {
	re := regexp.MustCompile(`defaultValue\s*=\s*"([^"]+)"`)
	if matches := re.FindStringSubmatch(line); len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// parseMiddlewareAnnotation 解析中间件注解
func parseMiddlewareAnnotation(line string) []string {
	// 匹配 @Middleware("auth", "ratelimit") 格式
	re := regexp.MustCompile(`@Middleware\s*\(\s*(.+)\s*\)`)
	if matches := re.FindStringSubmatch(line); len(matches) > 1 {
		middlewareStr := matches[1]
		var middlewares []string
		
		// 分割多个中间件
		for _, middleware := range strings.Split(middlewareStr, ",") {
			middleware = strings.TrimSpace(middleware)
			middleware = strings.Trim(middleware, `"'`)
			if middleware != "" {
				middlewares = append(middlewares, middleware)
			}
		}
		
		return middlewares
	}
	return nil
}

// ScanPackage 扫描包中的所有Go文件
func (ap *AnnotationParser) ScanPackage(packagePath string) error {
	matches, err := filepath.Glob(filepath.Join(packagePath, "*.go"))
	if err != nil {
		return err
	}
	
	for _, file := range matches {
		err := ap.ParseSourceFile(file)
		if err != nil {
			return err
		}
	}
	
	return nil
}

// ScanControllerType 根据类型扫描控制器
func (ap *AnnotationParser) ScanControllerType(controllerType reflect.Type) error {
	// 获取类型的源码位置
	if controllerType.Kind() == reflect.Ptr {
		controllerType = controllerType.Elem()
	}
	
	// 通过runtime获取类型定义的文件路径
	pkgPath := controllerType.PkgPath()
	if pkgPath == "" {
		return nil
	}
	
	// 尝试找到源文件路径
	// 这是一个简化的实现，实际可能需要更复杂的逻辑
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		return nil
	}
	
	dir := filepath.Dir(filename)
	return ap.ScanPackage(dir)
}

// GetControllerInfo 获取控制器信息
func (ap *AnnotationParser) GetControllerInfo(packageName, typeName string) *ControllerInfo {
	key := packageName + "." + typeName
	return ap.ControllerInfos[key]
}

// GetMethodInfo 获取方法信息
func (ap *AnnotationParser) GetMethodInfo(packageName, typeName, methodName string) *MethodInfo {
	key := packageName + "." + typeName + "." + methodName
	return ap.MethodInfos[key]
}

// GetAllControllerInfos 获取所有控制器信息
func (ap *AnnotationParser) GetAllControllerInfos() []*ControllerInfo {
	var infos []*ControllerInfo
	for _, info := range ap.ControllerInfos {
		infos = append(infos, info)
	}
	return infos
}

// GetControllerMethods 获取控制器的所有方法
func (ap *AnnotationParser) GetControllerMethods(packageName, typeName string) []*MethodInfo {
	var methods []*MethodInfo
	prefix := packageName + "." + typeName + "."
	
	for key, method := range ap.MethodInfos {
		if strings.HasPrefix(key, prefix) {
			methods = append(methods, method)
		}
	}
	
	return methods
}

// normalizePath 规范化路径
func normalizePath(path string) string {
	if path == "" {
		return ""
	}
	
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	
	if strings.HasSuffix(path, "/") && path != "/" {
		path = strings.TrimSuffix(path, "/")
	}
	
	return path
}

// CombinePath 组合路径
func CombinePath(basePath, methodPath string) string {
	basePath = normalizePath(basePath)
	methodPath = normalizePath(methodPath)
	
	if basePath == "" {
		return methodPath
	}
	
	if methodPath == "" {
		return basePath
	}
	
	return basePath + methodPath
}