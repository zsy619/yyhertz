# 🔧 代码生成工具

YYHertz MVC框架提供了强大的代码生成工具，可以自动生成控制器、模型、路由等代码，大幅提升开发效率。

## 🌟 功能特性

### ✨ 代码生成能力
- **🏗️ MVC代码生成** - 自动生成控制器、模型、视图
- **🗄️ 数据库相关** - 生成模型、迁移文件、种子数据
- **🔌 中间件生成** - 自动生成中间件模板
- **📋 API文档生成** - 自动生成Swagger文档
- **🧪 测试代码生成** - 生成单元测试和集成测试

### 🎯 生成器类型
- **控制器生成器** - RESTful控制器模板
- **模型生成器** - GORM模型和关联关系
- **服务生成器** - 业务逻辑服务层
- **中间件生成器** - 自定义中间件模板
- **API生成器** - API路由和文档

## 🚀 安装和配置

### 1. 安装代码生成工具

```bash
# 从项目根目录安装
go install github.com/zsy619/yyhertz/tools/codegen@latest

# 或者使用项目内置工具
go run tools/codegen/main.go
```

### 2. 配置文件

```yaml
# config/codegen.yaml
codegen:
  # 输出目录配置
  output:
    controllers: "controllers"
    models: "models"
    services: "services"
    middleware: "middleware"
    tests: "tests"
    docs: "docs/api"
  
  # 模板配置
  templates:
    base_path: "tools/codegen/templates"
    controller: "controller.tmpl"
    model: "model.tmpl"
    service: "service.tmpl"
    middleware: "middleware.tmpl"
    test: "test.tmpl"
  
  # 数据库配置
  database:
    driver: "mysql"
    dsn: "user:password@tcp(localhost:3306)/dbname"
    tables: ["users", "posts", "categories"]
  
  # 代码风格配置
  style:
    package_comment: true
    method_comment: true
    json_tag: "json"
    gorm_tag: "gorm"
    validate_tag: "validate"
```

### 3. 初始化项目结构

```bash
# 初始化代码生成器配置
codegen init

# 创建项目目录结构
codegen scaffold myproject
```

## 🏗️ 控制器生成

### 1. 基础控制器生成

```bash
# 生成RESTful控制器
codegen controller User --rest

# 生成自定义控制器
codegen controller Admin --methods=index,show,create,update,delete

# 生成API控制器
codegen controller api/User --api --version=v1
```

### 2. 控制器模板

```go
// tools/codegen/templates/controller.tmpl
package {{.Package}}

import (
    "strconv"
    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/models"
    "gorm.io/gorm"
)

// {{.Name}}Controller {{.Comment}}
type {{.Name}}Controller struct {
    mvc.BaseController
}

{{if .HasIndex}}
// GetIndex 获取{{.Comment}}列表
func (c *{{.Name}}Controller) GetIndex() {
    // 获取查询参数
    page, _ := strconv.Atoi(c.GetQuery("page", "1"))
    pageSize, _ := strconv.Atoi(c.GetQuery("page_size", "10"))
    
    // 构建查询
    var {{.ModelVar}} []models.{{.Name}}
    var total int64
    
    db := c.GetDB().Model(&models.{{.Name}}{})
    db.Count(&total)
    
    err := db.Scopes(c.Paginate(page, pageSize)).Find(&{{.ModelVar}}).Error
    if err != nil {
        c.Error(500, "查询失败: "+err.Error())
        return
    }
    
    c.JSON(map[string]interface{}{
        "code": 0,
        "message": "success",
        "data": map[string]interface{}{
            "{{.ModelVar}}": {{.ModelVar}},
            "pagination": map[string]interface{}{
                "page":      page,
                "page_size": pageSize,
                "total":     total,
                "pages":     (total + int64(pageSize) - 1) / int64(pageSize),
            },
        },
    })
}
{{end}}

{{if .HasShow}}
// GetShow 获取{{.Comment}}详情
func (c *{{.Name}}Controller) GetShow() {
    id, err := strconv.ParseUint(c.GetParam("id"), 10, 32)
    if err != nil {
        c.Error(400, "无效的ID")
        return
    }
    
    var {{.ModelVar}} models.{{.Name}}
    err = c.GetDB().First(&{{.ModelVar}}, id).Error
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            c.Error(404, "{{.Comment}}不存在")
        } else {
            c.Error(500, "查询失败: "+err.Error())
        }
        return
    }
    
    c.JSON(map[string]interface{}{
        "code": 0,
        "message": "success",
        "data": {{.ModelVar}},
    })
}
{{end}}

{{if .HasCreate}}
// PostCreate 创建{{.Comment}}
func (c *{{.Name}}Controller) PostCreate() {
    var {{.ModelVar}} models.{{.Name}}
    if err := c.BindJSON(&{{.ModelVar}}); err != nil {
        c.Error(400, "参数错误: "+err.Error())
        return
    }
    
    // 验证数据
    if err := c.Validate(&{{.ModelVar}}); err != nil {
        c.Error(400, "数据验证失败: "+err.Error())
        return
    }
    
    // 创建记录
    if err := c.GetDB().Create(&{{.ModelVar}}).Error; err != nil {
        c.Error(500, "创建失败: "+err.Error())
        return
    }
    
    c.JSON(map[string]interface{}{
        "code": 0,
        "message": "创建成功",
        "data": {{.ModelVar}},
    })
}
{{end}}

{{if .HasUpdate}}
// PutUpdate 更新{{.Comment}}
func (c *{{.Name}}Controller) PutUpdate() {
    id, err := strconv.ParseUint(c.GetParam("id"), 10, 32)
    if err != nil {
        c.Error(400, "无效的ID")
        return
    }
    
    // 检查记录是否存在
    var {{.ModelVar}} models.{{.Name}}
    if err := c.GetDB().First(&{{.ModelVar}}, id).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            c.Error(404, "{{.Comment}}不存在")
        } else {
            c.Error(500, "查询失败: "+err.Error())
        }
        return
    }
    
    // 绑定更新数据
    var updateData models.{{.Name}}
    if err := c.BindJSON(&updateData); err != nil {
        c.Error(400, "参数错误: "+err.Error())
        return
    }
    
    // 验证数据
    if err := c.Validate(&updateData); err != nil {
        c.Error(400, "数据验证失败: "+err.Error())
        return
    }
    
    // 更新记录
    if err := c.GetDB().Model(&{{.ModelVar}}).Updates(&updateData).Error; err != nil {
        c.Error(500, "更新失败: "+err.Error())
        return
    }
    
    c.JSON(map[string]interface{}{
        "code": 0,
        "message": "更新成功",
        "data": {{.ModelVar}},
    })
}
{{end}}

{{if .HasDelete}}
// DeleteDestroy 删除{{.Comment}}
func (c *{{.Name}}Controller) DeleteDestroy() {
    id, err := strconv.ParseUint(c.GetParam("id"), 10, 32)
    if err != nil {
        c.Error(400, "无效的ID")
        return
    }
    
    // 检查记录是否存在
    var {{.ModelVar}} models.{{.Name}}
    if err := c.GetDB().First(&{{.ModelVar}}, id).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            c.Error(404, "{{.Comment}}不存在")
        } else {
            c.Error(500, "查询失败: "+err.Error())
        }
        return
    }
    
    // 删除记录
    if err := c.GetDB().Delete(&{{.ModelVar}}).Error; err != nil {
        c.Error(500, "删除失败: "+err.Error())
        return
    }
    
    c.JSON(map[string]interface{}{
        "code": 0,
        "message": "删除成功",
    })
}
{{end}}
```

## 🗄️ 模型生成

### 1. 从数据库生成模型

```bash
# 从数据库表生成模型
codegen model --from-db --table=users

# 生成所有表的模型
codegen model --from-db --all

# 生成模型和关联关系
codegen model User --relations=Profile,Posts,Roles
```

### 2. 模型模板

```go
// tools/codegen/templates/model.tmpl
package {{.Package}}

import (
    "time"
    "gorm.io/gorm"
    {{range .Imports}}
    "{{.}}"
    {{end}}
)

{{if .Comment}}
// {{.Name}} {{.Comment}}
{{end}}
type {{.Name}} struct {
    {{if .HasBaseModel}}
    BaseModel
    {{else}}
    ID        uint           `gorm:"primarykey" json:"id"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
    {{end}}
    
    {{range .Fields}}
    {{.Name}} {{.Type}} `gorm:"{{.GormTag}}" json:"{{.JsonTag}}"{{if .ValidateTag}} validate:"{{.ValidateTag}}"{{end}}`
    {{end}}
    
    {{if .Relations}}
    // 关联关系
    {{range .Relations}}
    {{.Name}} {{.Type}} `gorm:"{{.GormTag}}" json:"{{.JsonTag}},omitempty"`
    {{end}}
    {{end}}
}

{{if .TableName}}
// TableName 自定义表名
func ({{.Receiver}}) TableName() string {
    return "{{.TableName}}"
}
{{end}}

{{range .Hooks}}
// {{.Name}} GORM钩子 - {{.Comment}}
func ({{$.Receiver}} *{{$.Name}}) {{.Name}}(tx *gorm.DB) error {
    {{.Body}}
    return nil
}
{{end}}

{{range .Methods}}
// {{.Name}} {{.Comment}}
func ({{$.Receiver}} *{{$.Name}}) {{.Name}}({{.Params}}) {{.Returns}} {
    {{.Body}}
}
{{end}}
```

### 3. 模型生成示例

```bash
# 生成用户模型
codegen model User \
    --fields="Username:string:uniqueIndex;size:50;not null,Email:string:uniqueIndex;size:100,Password:string:size:255" \
    --relations="Profile:UserProfile:foreignKey:UserID,Posts:Post:foreignKey:AuthorID" \
    --hooks="BeforeCreate,AfterFind" \
    --methods="GetFullName,IsActive"
```

## 🔌 中间件生成

### 1. 中间件生成命令

```bash
# 生成基础中间件
codegen middleware Auth

# 生成带配置的中间件
codegen middleware RateLimit --with-config

# 生成测试中间件
codegen middleware Logger --with-tests
```

### 2. 中间件模板

```go
// tools/codegen/templates/middleware.tmpl
package {{.Package}}

import (
    "context"
    "github.com/cloudwego/hertz/pkg/app"
    {{range .Imports}}
    "{{.}}"
    {{end}}
)

{{if .WithConfig}}
// {{.Name}}Config {{.Name}}中间件配置
type {{.Name}}Config struct {
    {{range .ConfigFields}}
    {{.Name}} {{.Type}} `yaml:"{{.YamlTag}}" json:"{{.JsonTag}}"`
    {{end}}
}

// Default{{.Name}}Config 默认配置
func Default{{.Name}}Config() *{{.Name}}Config {
    return &{{.Name}}Config{
        {{range .ConfigFields}}
        {{.Name}}: {{.DefaultValue}},
        {{end}}
    }
}
{{end}}

{{if .Comment}}
// {{.Name}}Middleware {{.Comment}}
{{end}}
func {{.Name}}Middleware({{if .WithConfig}}config *{{.Name}}Config{{end}}) app.HandlerFunc {
    {{if .WithConfig}}
    if config == nil {
        config = Default{{.Name}}Config()
    }
    {{end}}
    
    return func(c context.Context, ctx *app.RequestContext) {
        // 前置处理
        {{.PreProcess}}
        
        // 执行下一个处理器
        ctx.Next(c)
        
        // 后置处理
        {{.PostProcess}}
    }
}

{{range .HelperFunctions}}
// {{.Name}} {{.Comment}}
func {{.Name}}({{.Params}}) {{.Returns}} {
    {{.Body}}
}
{{end}}
```

## 📋 API文档生成

### 1. Swagger文档生成

```bash
# 安装swag工具
go install github.com/swaggo/swag/cmd/swag@latest

# 生成API文档
codegen swagger --output docs/swagger

# 从注释生成文档
swag init --generalInfo main.go --output docs/swagger
```

### 2. API注释规范

```go
// @Summary 创建用户
// @Description 创建新用户账户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param user body models.User true "用户信息"
// @Success 201 {object} models.User "创建成功"
// @Failure 400 {object} ErrorResponse "参数错误"
// @Failure 500 {object} ErrorResponse "服务器错误"
// @Router /api/users [post]
func (c *UserController) PostCreate() {
    // 实现代码
}
```

### 3. 文档模板生成

```bash
# 生成API文档模板
codegen docs --controller=UserController --output=docs/api/user.md

# 生成完整API文档
codegen docs --all --format=markdown --output=docs/api/
```

## 🧪 测试代码生成

### 1. 单元测试生成

```bash
# 生成控制器测试
codegen test UserController --type=unit

# 生成模型测试
codegen test User --type=model

# 生成集成测试
codegen test api/User --type=integration
```

### 2. 测试模板

```go
// tools/codegen/templates/test.tmpl
package {{.Package}}_test

import (
    "testing"
    "net/http"
    "bytes"
    "encoding/json"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"
    "github.com/zsy619/yyhertz/controllers"
    "github.com/zsy619/yyhertz/models"
    "github.com/zsy619/yyhertz/database"
)

// {{.Name}}TestSuite {{.Name}}测试套件
type {{.Name}}TestSuite struct {
    suite.Suite
    controller *controllers.{{.Name}}Controller
}

// SetupSuite 设置测试套件
func (suite *{{.Name}}TestSuite) SetupSuite() {
    // 初始化测试数据库
    database.InitTestDB()
    suite.controller = &controllers.{{.Name}}Controller{}
}

// TearDownSuite 清理测试套件
func (suite *{{.Name}}TestSuite) TearDownSuite() {
    database.CleanupTestDB()
}

// SetupTest 设置单个测试
func (suite *{{.Name}}TestSuite) SetupTest() {
    // 清理测试数据
    database.TruncateTables()
}

{{range .TestMethods}}
// Test{{.Name}} 测试{{.Comment}}
func (suite *{{$.Name}}TestSuite) Test{{.Name}}() {
    {{.Body}}
}
{{end}}

// Test{{.Name}}Controller_GetIndex 测试获取列表
func (suite *{{.Name}}TestSuite) Test{{.Name}}Controller_GetIndex() {
    // 创建测试数据
    {{.ModelVar}} := &models.{{.Name}}{
        {{range .SampleFields}}
        {{.Name}}: {{.Value}},
        {{end}}
    }
    err := database.GetDB().Create({{.ModelVar}}).Error
    assert.NoError(suite.T(), err)
    
    // 执行请求
    resp := suite.performRequest("GET", "/{{.Route}}", nil)
    
    // 验证响应
    assert.Equal(suite.T(), http.StatusOK, resp.Code)
    
    var response map[string]interface{}
    err = json.Unmarshal(resp.Body.Bytes(), &response)
    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), float64(0), response["code"])
}

// Test{{.Name}}Controller_PostCreate 测试创建
func (suite *{{.Name}}TestSuite) Test{{.Name}}Controller_PostCreate() {
    // 准备测试数据
    data := map[string]interface{}{
        {{range .RequiredFields}}
        "{{.JsonTag}}": {{.TestValue}},
        {{end}}
    }
    
    jsonData, _ := json.Marshal(data)
    
    // 执行请求
    resp := suite.performRequest("POST", "/{{.Route}}", bytes.NewBuffer(jsonData))
    
    // 验证响应
    assert.Equal(suite.T(), http.StatusCreated, resp.Code)
    
    // 验证数据库
    var count int64
    database.GetDB().Model(&models.{{.Name}}{}).Count(&count)
    assert.Equal(suite.T(), int64(1), count)
}

// performRequest 执行HTTP请求
func (suite *{{.Name}}TestSuite) performRequest(method, url string, body *bytes.Buffer) *httptest.ResponseRecorder {
    // 实现HTTP请求测试逻辑
    // ...
}

// TestMain 测试入口
func TestMain(m *testing.M) {
    // 设置测试环境
    setupTestEnvironment()
    
    // 运行测试
    code := m.Run()
    
    // 清理测试环境
    teardownTestEnvironment()
    
    os.Exit(code)
}

// Test{{.Name}}TestSuite 运行测试套件
func Test{{.Name}}TestSuite(t *testing.T) {
    suite.Run(t, new({{.Name}}TestSuite))
}
```

## 🔄 批量生成

### 1. 批量生成脚本

```bash
#!/bin/bash
# scripts/generate_crud.sh

# 生成CRUD模块
generate_crud() {
    local model=$1
    local table=$2
    local comment=$3
    
    echo "Generating CRUD for $model..."
    
    # 生成模型
    codegen model $model --from-db --table=$table --comment="$comment"
    
    # 生成控制器
    codegen controller $model --rest --comment="$comment"
    
    # 生成服务层
    codegen service $model --comment="$comment"
    
    # 生成测试
    codegen test $model --type=all
    
    # 生成API文档
    codegen docs $model --format=swagger
    
    echo "Generated CRUD for $model successfully!"
}

# 批量生成
generate_crud "User" "users" "用户"
generate_crud "Post" "posts" "文章"
generate_crud "Category" "categories" "分类"
```

### 2. 配置驱动生成

```yaml
# config/generate.yaml
models:
  - name: "User"
    table: "users"
    comment: "用户"
    fields:
      - name: "Username"
        type: "string"
        gorm: "uniqueIndex;size:50;not null"
        validate: "required,min=3,max=50"
      - name: "Email"
        type: "string"
        gorm: "uniqueIndex;size:100;not null"
        validate: "required,email"
    relations:
      - name: "Profile"
        type: "UserProfile"
        relation: "has_one"
      - name: "Posts"
        type: "Post"
        relation: "has_many"
    
  - name: "Post"
    table: "posts"
    comment: "文章"
    fields:
      - name: "Title"
        type: "string"
        gorm: "size:200;not null"
        validate: "required,max=200"
      - name: "Content"
        type: "string"
        gorm: "type:longtext"
    relations:
      - name: "Author"
        type: "User"
        relation: "belongs_to"
```

```bash
# 从配置文件批量生成
codegen generate --config=config/generate.yaml --all
```

## 🎨 自定义模板

### 1. 创建自定义模板

```go
// tools/codegen/templates/custom_controller.tmpl
package {{.Package}}

import (
    "github.com/zsy619/yyhertz/framework/mvc"
    {{range .CustomImports}}
    "{{.}}"
    {{end}}
)

// {{.Name}}Controller {{.Comment}}
// 这是一个自定义生成的控制器模板
type {{.Name}}Controller struct {
    mvc.BaseController
    {{range .Dependencies}}
    {{.Name}} {{.Type}}
    {{end}}
}

// NewController 创建控制器实例
func New{{.Name}}Controller({{range .Dependencies}}{{.Param}} {{.Type}},{{end}}) *{{.Name}}Controller {
    return &{{.Name}}Controller{
        {{range .Dependencies}}
        {{.Name}}: {{.Param}},
        {{end}}
    }
}

{{range .CustomMethods}}
// {{.Name}} {{.Comment}}
func (c *{{$.Name}}Controller) {{.Name}}({{.Params}}) {{.Returns}} {
    {{.Body}}
}
{{end}}
```

### 2. 模板函数扩展

```go
// tools/codegen/template_funcs.go
package codegen

import (
    "strings"
    "text/template"
)

// GetTemplateFuncs 获取模板函数
func GetTemplateFuncs() template.FuncMap {
    return template.FuncMap{
        "lower":      strings.ToLower,
        "upper":      strings.ToUpper,
        "title":      strings.Title,
        "camelCase":  toCamelCase,
        "snakeCase":  toSnakeCase,
        "pluralize":  pluralize,
        "contains":   strings.Contains,
        "hasPrefix":  strings.HasPrefix,
        "hasSuffix":  strings.HasSuffix,
        "replace":    strings.Replace,
        "join":       strings.Join,
        "split":      strings.Split,
        "formatType": formatGoType,
        "formatTag":  formatStructTag,
    }
}

func toCamelCase(s string) string {
    // 实现驼峰命名转换
    words := strings.Split(s, "_")
    result := strings.ToLower(words[0])
    for i := 1; i < len(words); i++ {
        result += strings.Title(strings.ToLower(words[i]))
    }
    return result
}

func toSnakeCase(s string) string {
    // 实现下划线命名转换
    // 实现逻辑...
}

func pluralize(s string) string {
    // 实现复数形式转换
    // 简单实现，可以使用更复杂的规则
    if strings.HasSuffix(s, "y") {
        return strings.TrimSuffix(s, "y") + "ies"
    }
    if strings.HasSuffix(s, "s") {
        return s + "es"
    }
    return s + "s"
}
```

## 📊 生成报告

### 1. 生成统计

```bash
# 生成代码统计报告
codegen stats --output=reports/generation_stats.json

# 生成覆盖率报告
codegen coverage --output=reports/coverage.html
```

### 2. 代码质量检查

```bash
# 检查生成的代码质量
codegen lint --fix

# 格式化生成的代码
codegen format --all

# 优化导入
codegen imports --optimize
```

## 🔧 CLI工具开发

### 1. 命令行接口

```go
// tools/codegen/cmd/root.go
package cmd

import (
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
    Use:   "codegen",
    Short: "YYHertz MVC代码生成工具",
    Long:  `强大的代码生成工具，支持生成控制器、模型、中间件等`,
}

func Execute() error {
    return rootCmd.Execute()
}

func init() {
    cobra.OnInitialize(initConfig)
    
    rootCmd.PersistentFlags().StringP("config", "c", "", "配置文件路径")
    rootCmd.PersistentFlags().BoolP("verbose", "v", false, "详细输出")
    rootCmd.PersistentFlags().StringP("output", "o", ".", "输出目录")
    
    viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
    viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
    viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
}

func initConfig() {
    if cfgFile := viper.GetString("config"); cfgFile != "" {
        viper.SetConfigFile(cfgFile)
    } else {
        viper.SetConfigName("codegen")
        viper.SetConfigType("yaml")
        viper.AddConfigPath("./config")
        viper.AddConfigPath(".")
    }
    
    viper.AutomaticEnv()
    
    if err := viper.ReadInConfig(); err == nil {
        fmt.Println("Using config file:", viper.ConfigFileUsed())
    }
}
```

### 2. 子命令实现

```go
// tools/codegen/cmd/controller.go
package cmd

import (
    "github.com/spf13/cobra"
    "github.com/zsy619/yyhertz/tools/codegen/generator"
)

var controllerCmd = &cobra.Command{
    Use:   "controller [name]",
    Short: "生成控制器",
    Args:  cobra.ExactArgs(1),
    Run:   generateController,
}

func init() {
    rootCmd.AddCommand(controllerCmd)
    
    controllerCmd.Flags().Bool("rest", false, "生成RESTful控制器")
    controllerCmd.Flags().Bool("api", false, "生成API控制器")
    controllerCmd.Flags().StringSlice("methods", []string{}, "指定方法列表")
    controllerCmd.Flags().String("comment", "", "控制器注释")
    controllerCmd.Flags().String("package", "controllers", "包名")
}

func generateController(cmd *cobra.Command, args []string) {
    name := args[0]
    
    // 获取参数
    isREST, _ := cmd.Flags().GetBool("rest")
    isAPI, _ := cmd.Flags().GetBool("api")
    methods, _ := cmd.Flags().GetStringSlice("methods")
    comment, _ := cmd.Flags().GetString("comment")
    pkg, _ := cmd.Flags().GetString("package")
    
    // 创建生成器
    gen := generator.NewControllerGenerator(&generator.ControllerConfig{
        Name:    name,
        Package: pkg,
        Comment: comment,
        IsREST:  isREST,
        IsAPI:   isAPI,
        Methods: methods,
    })
    
    // 生成代码
    if err := gen.Generate(); err != nil {
        fmt.Printf("生成控制器失败: %v\n", err)
        os.Exit(1)
    }
    
    fmt.Printf("控制器 %s 生成成功!\n", name)
}
```

## 📚 最佳实践

### 1. 模板设计原则
- 保持模板简洁清晰
- 使用有意义的变量名
- 提供充分的注释
- 支持自定义配置

### 2. 代码生成规范
- 遵循Go语言规范
- 保持代码风格一致
- 生成完整的测试代码
- 包含必要的文档

### 3. 维护建议
- 定期更新模板
- 版本控制生成工具
- 收集用户反馈
- 持续优化性能

## 🔗 相关资源

- [热重载开发](./hot-reload.md)
- [性能监控工具](./performance.md)
- [测试工具集成](./testing.md)
- [项目结构指南](../getting-started/structure.md)

---

> 💡 **提示**: 代码生成工具可以大幅提升开发效率，但生成的代码仍需要根据具体需求进行调整和优化。
