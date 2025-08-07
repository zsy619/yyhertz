package codegen

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

// DocGenerator API文档生成器
type DocGenerator struct {
	ProjectRoot string
	OutputDir   string
	Title       string
	Version     string
	BaseURL     string
}

// APIDoc API文档结构
type APIDoc struct {
	Info        APIInfo             `json:"info"`
	Servers     []APIServer         `json:"servers"`
	Paths       map[string]PathItem `json:"paths"`
	Components  Components          `json:"components"`
	GeneratedAt time.Time           `json:"generated_at"`
}

// APIInfo API信息
type APIInfo struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
}

// APIServer 服务器信息
type APIServer struct {
	URL         string `json:"url"`
	Description string `json:"description"`
}

// PathItem 路径项
type PathItem struct {
	Get    *Operation `json:"get,omitempty"`
	Post   *Operation `json:"post,omitempty"`
	Put    *Operation `json:"put,omitempty"`
	Delete *Operation `json:"delete,omitempty"`
}

// Operation 操作
type Operation struct {
	Summary     string              `json:"summary"`
	Description string              `json:"description"`
	Tags        []string            `json:"tags"`
	Parameters  []Parameter         `json:"parameters,omitempty"`
	RequestBody *RequestBody        `json:"requestBody,omitempty"`
	Responses   map[string]Response `json:"responses"`
}

// Parameter 参数
type Parameter struct {
	Name        string `json:"name"`
	In          string `json:"in"` // query, path, header
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Schema      Schema `json:"schema"`
}

// RequestBody 请求体
type RequestBody struct {
	Description string               `json:"description"`
	Content     map[string]MediaType `json:"content"`
	Required    bool                 `json:"required"`
}

// Response 响应
type Response struct {
	Description string               `json:"description"`
	Content     map[string]MediaType `json:"content,omitempty"`
}

// MediaType 媒体类型
type MediaType struct {
	Schema Schema `json:"schema"`
}

// Schema 模式
type Schema struct {
	Type       string            `json:"type"`
	Properties map[string]Schema `json:"properties,omitempty"`
	Items      *Schema           `json:"items,omitempty"`
	Example    interface{}       `json:"example,omitempty"`
}

// Components 组件
type Components struct {
	Schemas map[string]Schema `json:"schemas"`
}

// NewDocGenerator 创建文档生成器
func NewDocGenerator(projectRoot string) *DocGenerator {
	return &DocGenerator{
		ProjectRoot: projectRoot,
		OutputDir:   "docs/api",
		Title:       "YYHertz API Documentation",
		Version:     "1.0.0",
		BaseURL:     "http://localhost:8080",
	}
}

// Generate 生成API文档
func (dg *DocGenerator) Generate(controllers []ControllerInfo) error {
	doc := dg.buildAPIDoc(controllers)

	// 生成JSON格式
	if err := dg.generateJSON(doc); err != nil {
		return err
	}

	// 生成HTML格式
	if err := dg.generateHTML(doc); err != nil {
		return err
	}

	// 生成Markdown格式
	return dg.generateMarkdown(doc)
}

// buildAPIDoc 构建API文档
func (dg *DocGenerator) buildAPIDoc(controllers []ControllerInfo) *APIDoc {
	doc := &APIDoc{
		Info: APIInfo{
			Title:       dg.Title,
			Description: "基于 YYHertz 框架的 API 文档",
			Version:     dg.Version,
		},
		Servers: []APIServer{
			{
				URL:         dg.BaseURL,
				Description: "开发服务器",
			},
		},
		Paths:       make(map[string]PathItem),
		Components:  Components{Schemas: make(map[string]Schema)},
		GeneratedAt: time.Now(),
	}

	for _, ctrl := range controllers {
		for _, method := range ctrl.Methods {
			path := dg.buildPath(ctrl.Prefix, method.Path)

			if _, exists := doc.Paths[path]; !exists {
				doc.Paths[path] = PathItem{}
			}

			pathItem := doc.Paths[path]
			operation := dg.buildOperation(ctrl, method)

			switch strings.ToUpper(method.HTTPMethod) {
			case "GET":
				pathItem.Get = operation
			case "POST":
				pathItem.Post = operation
			case "PUT":
				pathItem.Put = operation
			case "DELETE":
				pathItem.Delete = operation
			}

			doc.Paths[path] = pathItem
		}
	}

	return doc
}

// buildPath 构建路径
func (dg *DocGenerator) buildPath(prefix, path string) string {
	if prefix != "" {
		return "/" + strings.Trim(prefix, "/") + "/" + strings.Trim(path, "/")
	}
	return path
}

// buildOperation 构建操作
func (dg *DocGenerator) buildOperation(ctrl ControllerInfo, method MethodInfo) *Operation {
	operation := &Operation{
		Summary:     method.Comment,
		Description: fmt.Sprintf("%s.%s", ctrl.Name, method.Name),
		Tags:        []string{ctrl.Name},
		Parameters:  []Parameter{},
		Responses: map[string]Response{
			"200": {
				Description: "成功",
				Content: map[string]MediaType{
					"application/json": {
						Schema: Schema{
							Type: "object",
							Properties: map[string]Schema{
								"code":    {Type: "integer", Example: 0},
								"message": {Type: "string", Example: "success"},
								"data":    {Type: "object"},
							},
						},
					},
				},
			},
			"400": {
				Description: "请求错误",
			},
			"500": {
				Description: "服务器错误",
			},
		},
	}

	// 解析参数
	for _, param := range method.Params {
		if param.Name != "ctx" && param.Name != "c" {
			parameter := Parameter{
				Name:        param.Name,
				Description: fmt.Sprintf("%s 参数", param.Name),
				Required:    true,
				Schema:      dg.typeToSchema(param.Type),
			}

			// 根据HTTP方法确定参数位置
			if method.HTTPMethod == "GET" || method.HTTPMethod == "DELETE" {
				parameter.In = "query"
			} else {
				// POST/PUT 参数通常在请求体中
				if operation.RequestBody == nil {
					operation.RequestBody = &RequestBody{
						Description: "请求参数",
						Required:    true,
						Content: map[string]MediaType{
							"application/json": {
								Schema: Schema{
									Type:       "object",
									Properties: make(map[string]Schema),
								},
							},
						},
					}
				}
				operation.RequestBody.Content["application/json"].Schema.Properties[param.Name] = dg.typeToSchema(param.Type)
				continue
			}

			operation.Parameters = append(operation.Parameters, parameter)
		}
	}

	return operation
}

// typeToSchema 类型转Schema
func (dg *DocGenerator) typeToSchema(typeName string) Schema {
	switch typeName {
	case "string":
		return Schema{Type: "string"}
	case "int", "int32", "int64":
		return Schema{Type: "integer"}
	case "float32", "float64":
		return Schema{Type: "number"}
	case "bool":
		return Schema{Type: "boolean"}
	default:
		return Schema{Type: "object"}
	}
}

// generateJSON 生成JSON文档
func (dg *DocGenerator) generateJSON(doc *APIDoc) error {
	outputDir := filepath.Join(dg.ProjectRoot, dg.OutputDir)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	file, err := os.Create(filepath.Join(outputDir, "api.json"))
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(doc)
}

// generateHTML 生成HTML文档
func (dg *DocGenerator) generateHTML(doc *APIDoc) error {
	tmpl := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Info.Title}}</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; }
        .header { background: #f5f5f5; padding: 20px; border-radius: 5px; margin-bottom: 20px; }
        .path { margin-bottom: 30px; border: 1px solid #ddd; border-radius: 5px; }
        .path-header { background: #f8f9fa; padding: 15px; border-bottom: 1px solid #ddd; }
        .method { display: inline-block; padding: 5px 10px; border-radius: 3px; color: white; font-weight: bold; }
        .get { background: #28a745; }
        .post { background: #007bff; }
        .put { background: #ffc107; color: black; }
        .delete { background: #dc3545; }
        .operation { padding: 15px; }
        .parameters { margin-top: 15px; }
        .parameter { background: #f8f9fa; padding: 10px; margin: 5px 0; border-radius: 3px; }
    </style>
</head>
<body>
    <div class="header">
        <h1>{{.Info.Title}}</h1>
        <p>{{.Info.Description}}</p>
        <p>版本: {{.Info.Version}} | 生成时间: {{.GeneratedAt.Format "2006-01-02 15:04:05"}}</p>
    </div>

    {{range $path, $pathItem := .Paths}}
    <div class="path">
        <div class="path-header">
            <h3>{{$path}}</h3>
        </div>
        
        {{if $pathItem.Get}}
        <div class="operation">
            <span class="method get">GET</span>
            <strong>{{$pathItem.Get.Summary}}</strong>
            <p>{{$pathItem.Get.Description}}</p>
            {{if $pathItem.Get.Parameters}}
            <div class="parameters">
                <h4>参数:</h4>
                {{range $pathItem.Get.Parameters}}
                <div class="parameter">
                    <strong>{{.Name}}</strong> ({{.In}}) - {{.Description}}
                    {{if .Required}}<span style="color: red;">*</span>{{end}}
                </div>
                {{end}}
            </div>
            {{end}}
        </div>
        {{end}}
        
        {{if $pathItem.Post}}
        <div class="operation">
            <span class="method post">POST</span>
            <strong>{{$pathItem.Post.Summary}}</strong>
            <p>{{$pathItem.Post.Description}}</p>
        </div>
        {{end}}
        
        {{if $pathItem.Put}}
        <div class="operation">
            <span class="method put">PUT</span>
            <strong>{{$pathItem.Put.Summary}}</strong>
            <p>{{$pathItem.Put.Description}}</p>
        </div>
        {{end}}
        
        {{if $pathItem.Delete}}
        <div class="operation">
            <span class="method delete">DELETE</span>
            <strong>{{$pathItem.Delete.Summary}}</strong>
            <p>{{$pathItem.Delete.Description}}</p>
        </div>
        {{end}}
    </div>
    {{end}}
</body>
</html>`

	t, err := template.New("html").Parse(tmpl)
	if err != nil {
		return err
	}

	outputDir := filepath.Join(dg.ProjectRoot, dg.OutputDir)
	file, err := os.Create(filepath.Join(outputDir, "api.html"))
	if err != nil {
		return err
	}
	defer file.Close()

	return t.Execute(file, doc)
}

// generateMarkdown 生成Markdown文档
func (dg *DocGenerator) generateMarkdown(doc *APIDoc) error {
	tmpl := `# {{.Info.Title}}

{{.Info.Description}}

**版本:** {{.Info.Version}}  
**生成时间:** {{.GeneratedAt.Format "2006-01-02 15:04:05"}}

## 服务器

{{range .Servers}}
- {{.URL}} - {{.Description}}
{{end}}

## API 接口

{{range $path, $pathItem := .Paths}}
### {{$path}}

{{if $pathItem.Get}}
#### GET {{$path}}

{{$pathItem.Get.Summary}}

{{$pathItem.Get.Description}}

{{if $pathItem.Get.Parameters}}
**参数:**

| 名称 | 位置 | 类型 | 必需 | 描述 |
|------|------|------|------|------|
{{range $pathItem.Get.Parameters}}| {{.Name}} | {{.In}} | {{.Schema.Type}} | {{if .Required}}是{{else}}否{{end}} | {{.Description}} |
{{end}}
{{end}}

{{end}}

{{if $pathItem.Post}}
#### POST {{$path}}

{{$pathItem.Post.Summary}}

{{$pathItem.Post.Description}}

{{end}}

{{if $pathItem.Put}}
#### PUT {{$path}}

{{$pathItem.Put.Summary}}

{{$pathItem.Put.Description}}

{{end}}

{{if $pathItem.Delete}}
#### DELETE {{$path}}

{{$pathItem.Delete.Summary}}

{{$pathItem.Delete.Description}}

{{end}}

---

{{end}}
`

	t, err := template.New("markdown").Parse(tmpl)
	if err != nil {
		return err
	}

	outputDir := filepath.Join(dg.ProjectRoot, dg.OutputDir)
	file, err := os.Create(filepath.Join(outputDir, "api.md"))
	if err != nil {
		return err
	}
	defer file.Close()

	return t.Execute(file, doc)
}
