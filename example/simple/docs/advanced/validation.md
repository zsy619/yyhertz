# 验证系统

YYHertz 框架提供了强大而灵活的数据验证系统，支持结构体标签验证、自定义验证规则、国际化错误消息等功能，确保应用程序的数据完整性和安全性。

## 概述

数据验证是 Web 应用程序的重要组成部分。YYHertz 的验证系统基于流行的 `validator` 包，并进行了扩展和优化，提供：

- 声明式验证规则
- 自定义验证器
- 条件验证
- 错误消息国际化
- 复合验证
- 异步验证

## 基本使用

### 结构体标签验证

```go
package models

import (
    "time"
    "github.com/zsy619/yyhertz/framework/mvc/validation"
)

// 用户注册请求
type UserRegisterRequest struct {
    Username string `json:"username" validate:"required,min=3,max=20,alphanum"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8,containsany=!@#$%^&*"`
    Age      int    `json:"age" validate:"gte=18,lte=120"`
    Phone    string `json:"phone" validate:"omitempty,e164"`
    Website  string `json:"website" validate:"omitempty,url"`
}

// 产品创建请求
type ProductCreateRequest struct {
    Name        string    `json:"name" validate:"required,min=1,max=100"`
    Description string    `json:"description" validate:"max=1000"`
    Price       float64   `json:"price" validate:"required,gt=0"`
    CategoryID  int       `json:"category_id" validate:"required,min=1"`
    Tags        []string  `json:"tags" validate:"max=10,dive,min=1,max=50"`
    LaunchDate  time.Time `json:"launch_date" validate:"omitempty,gtefield=CreatedAt"`
    CreatedAt   time.Time `json:"-"`
}
```

### 在控制器中使用验证

```go
package controllers

import (
    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/framework/mvc/validation"
)

type UserController struct {
    mvc.Controller
}

func (c *UserController) Register() {
    var req UserRegisterRequest
    
    // 绑定并验证请求数据
    if err := c.BindAndValidate(&req); err != nil {
        // 返回验证错误
        c.JSON(400, map[string]interface{}{
            "error": "Validation failed",
            "details": err.Details(),
        })
        return
    }
    
    // 验证通过，处理注册逻辑
    user, err := c.userService.Register(req)
    if err != nil {
        c.JSON(500, map[string]string{"error": "Registration failed"})
        return
    }
    
    c.JSON(201, user)
}

// 手动验证
func (c *UserController) UpdateProfile() {
    var req UserUpdateRequest
    
    if err := c.BindJSON(&req); err != nil {
        c.JSON(400, map[string]string{"error": "Invalid JSON"})
        return
    }
    
    // 使用验证器验证
    validator := validation.GetValidator()
    if err := validator.Validate(req); err != nil {
        c.JSON(400, map[string]interface{}{
            "error": "Validation failed",
            "details": validation.FormatErrors(err),
        })
        return
    }
    
    // 处理更新逻辑
    // ...
}
```

## 验证规则

### 内置验证规则

```go
// 字符串验证
type StringValidation struct {
    Value string `validate:"required"`           // 必填
    Value string `validate:"min=3,max=20"`       // 长度范围
    Value string `validate:"len=10"`             // 固定长度
    Value string `validate:"alpha"`              // 只包含字母
    Value string `validate:"alphanum"`           // 字母和数字
    Value string `validate:"alphanumspace"`      // 字母、数字和空格
    Value string `validate:"ascii"`              // ASCII 字符
    Value string `validate:"printascii"`         // 可打印 ASCII
    Value string `validate:"email"`              // 邮箱格式
    Value string `validate:"url"`                // URL 格式
    Value string `validate:"uri"`                // URI 格式
    Value string `validate:"base64"`             // Base64 编码
    Value string `validate:"contains=substring"` // 包含子字符串
    Value string `validate:"containsany=!@#"`    // 包含任意指定字符
    Value string `validate:"excludes=test"`      // 不包含指定字符串
    Value string `validate:"startswith=prefix"`  // 以指定字符串开头
    Value string `validate:"endswith=suffix"`    // 以指定字符串结尾
}

// 数字验证
type NumberValidation struct {
    Value int     `validate:"min=0,max=100"`      // 范围验证
    Value int     `validate:"gte=18,lte=65"`      // 大于等于、小于等于
    Value int     `validate:"gt=0,lt=1000"`       // 大于、小于
    Value float64 `validate:"gte=0.0"`            // 浮点数验证
    Value int     `validate:"oneof=1 2 3"`        // 枚举值
}

// 时间验证
type TimeValidation struct {
    Value time.Time `validate:"required"`         // 必填时间
    Value time.Time `validate:"gtefield=StartTime"` // 大于某个字段
    Value time.Time `validate:"ltefield=EndTime"`   // 小于某个字段
}

// 数组/切片验证
type SliceValidation struct {
    Tags  []string `validate:"required,min=1,max=5"`          // 数组长度
    Items []Item   `validate:"required,dive,required"`        // 深度验证
    IDs   []int    `validate:"unique"`                        // 唯一性验证
}
```

### 正则表达式验证

```go
type RegexValidation struct {
    Phone     string `validate:"regexp=^[0-9]{10,11}$"`         // 手机号
    IDCard    string `validate:"regexp=^[0-9]{17}[0-9Xx]$"`     // 身份证号
    Username  string `validate:"regexp=^[a-zA-Z][a-zA-Z0-9_]{2,19}$"` // 用户名
    Password  string `validate:"regexp=^(?=.*[a-z])(?=.*[A-Z])(?=.*\\d)[a-zA-Z\\d@$!%*?&]{8,}$"` // 密码强度
}
```

## 自定义验证器

### 注册自定义验证规则

```go
package validation

import (
    "regexp"
    "unicode/utf8"
    "github.com/go-playground/validator/v10"
)

// 初始化自定义验证器
func init() {
    v := GetValidator()
    
    // 注册中文字符验证
    v.RegisterValidation("chinese", validateChinese)
    
    // 注册用户名唯一性验证
    v.RegisterValidation("unique_username", validateUniqueUsername)
    
    // 注册强密码验证
    v.RegisterValidation("strong_password", validateStrongPassword)
    
    // 注册手机号验证
    v.RegisterValidation("mobile", validateMobile)
}

// 验证中文字符
func validateChinese(fl validator.FieldLevel) bool {
    value := fl.Field().String()
    if value == "" {
        return true // 空值由 required 处理
    }
    
    for _, r := range value {
        if !isChinese(r) {
            return false
        }
    }
    return true
}

func isChinese(r rune) bool {
    return r >= 0x4e00 && r <= 0x9fff
}

// 验证用户名唯一性
func validateUniqueUsername(fl validator.FieldLevel) bool {
    username := fl.Field().String()
    // 这里需要访问数据库，实际项目中可能需要依赖注入
    return !userService.UsernameExists(username)
}

// 验证强密码
func validateStrongPassword(fl validator.FieldLevel) bool {
    password := fl.Field().String()
    
    if len(password) < 8 {
        return false
    }
    
    var (
        hasUpper   = false
        hasLower   = false
        hasNumber  = false
        hasSpecial = false
    )
    
    for _, char := range password {
        switch {
        case char >= 'A' && char <= 'Z':
            hasUpper = true
        case char >= 'a' && char <= 'z':
            hasLower = true
        case char >= '0' && char <= '9':
            hasNumber = true
        case char >= '!' && char <= '/':
            hasSpecial = true
        }
    }
    
    return hasUpper && hasLower && hasNumber && hasSpecial
}

// 验证手机号
func validateMobile(fl validator.FieldLevel) bool {
    mobile := fl.Field().String()
    matched, _ := regexp.MatchString(`^1[3-9]\d{9}$`, mobile)
    return matched
}
```

### 使用自定义验证规则

```go
type UserProfile struct {
    Name     string `json:"name" validate:"required,chinese,min=2,max=10"`
    Username string `json:"username" validate:"required,unique_username,min=3,max=20"`
    Password string `json:"password" validate:"required,strong_password"`
    Mobile   string `json:"mobile" validate:"required,mobile"`
}
```

## 条件验证

### 基于字段的条件验证

```go
type ConditionalValidation struct {
    Type      string `json:"type" validate:"required,oneof=personal business"`
    
    // 个人用户字段
    FirstName string `json:"first_name" validate:"required_if=Type personal"`
    LastName  string `json:"last_name" validate:"required_if=Type personal"`
    IDCard    string `json:"id_card" validate:"required_if=Type personal,omitempty,len=18"`
    
    // 企业用户字段
    CompanyName string `json:"company_name" validate:"required_if=Type business"`
    TaxNumber   string `json:"tax_number" validate:"required_if=Type business,omitempty,len=15"`
    
    // 地址信息（当提供了城市时，省份是必需的）
    Province string `json:"province" validate:"required_with=City"`
    City     string `json:"city"`
    Address  string `json:"address" validate:"required_with=Province City"`
    
    // 年龄验证（当提供生日时验证年龄）
    Birthday *time.Time `json:"birthday"`
    Age      int        `json:"age" validate:"required_without=Birthday,gte=18"`
}
```

### 自定义条件验证

```go
// 自定义条件验证器
func validateBusinessFields(fl validator.FieldLevel) bool {
    // 获取当前结构体
    current := fl.Top().Interface()
    
    if user, ok := current.(UserProfile); ok {
        if user.Type == "business" {
            // 企业用户必须提供公司名称
            return user.CompanyName != ""
        }
    }
    
    return true
}

// 注册条件验证器
func init() {
    v := GetValidator()
    v.RegisterValidation("business_required", validateBusinessFields)
}
```

## 错误消息国际化

### 自定义错误消息

```go
package validation

import (
    "fmt"
    "github.com/go-playground/validator/v10"
)

// 错误消息映射
var errorMessages = map[string]string{
    "required":        "{{.Field}}是必填字段",
    "min":            "{{.Field}}最小长度为{{.Param}}",
    "max":            "{{.Field}}最大长度为{{.Param}}",
    "email":          "{{.Field}}必须是有效的邮箱地址",
    "gte":            "{{.Field}}必须大于或等于{{.Param}}",
    "lte":            "{{.Field}}必须小于或等于{{.Param}}",
    "chinese":        "{{.Field}}只能包含中文字符",
    "unique_username": "{{.Field}}已被使用",
    "strong_password": "{{.Field}}必须包含大小写字母、数字和特殊字符",
    "mobile":         "{{.Field}}必须是有效的手机号码",
}

// 字段名称映射
var fieldNames = map[string]string{
    "Username": "用户名",
    "Email":    "邮箱",
    "Password": "密码",
    "Age":      "年龄",
    "Name":     "姓名",
    "Mobile":   "手机号",
}

// 格式化验证错误
func FormatErrors(err error) []ValidationError {
    var errors []ValidationError
    
    if validationErrors, ok := err.(validator.ValidationErrors); ok {
        for _, validationError := range validationErrors {
            fieldName := getFieldName(validationError.Field())
            message := getErrorMessage(validationError)
            
            errors = append(errors, ValidationError{
                Field:   validationError.Field(),
                Value:   validationError.Value(),
                Tag:     validationError.Tag(),
                Param:   validationError.Param(),
                Message: message,
                FieldName: fieldName,
            })
        }
    }
    
    return errors
}

func getFieldName(field string) string {
    if name, exists := fieldNames[field]; exists {
        return name
    }
    return field
}

func getErrorMessage(err validator.FieldError) string {
    template, exists := errorMessages[err.Tag()]
    if !exists {
        return fmt.Sprintf("%s验证失败", err.Field())
    }
    
    // 简单的模板替换
    message := template
    message = strings.ReplaceAll(message, "{{.Field}}", getFieldName(err.Field()))
    message = strings.ReplaceAll(message, "{{.Param}}", err.Param())
    
    return message
}

type ValidationError struct {
    Field     string      `json:"field"`
    Value     interface{} `json:"value"`
    Tag       string      `json:"tag"`
    Param     string      `json:"param"`
    Message   string      `json:"message"`
    FieldName string      `json:"field_name"`
}
```

### 多语言支持

```go
// 国际化验证错误
type I18nValidator struct {
    validator *validator.Validate
    messages  map[string]map[string]string // [language][tag]message
}

func NewI18nValidator() *I18nValidator {
    return &I18nValidator{
        validator: validator.New(),
        messages:  make(map[string]map[string]string),
    }
}

func (v *I18nValidator) SetMessages(lang string, messages map[string]string) {
    v.messages[lang] = messages
}

func (v *I18nValidator) ValidateWithLang(lang string, s interface{}) error {
    err := v.validator.Struct(s)
    if err == nil {
        return nil
    }
    
    return v.formatErrorsWithLang(lang, err)
}

// 加载多语言消息
func loadI18nMessages() {
    validator := GetI18nValidator()
    
    // 中文消息
    validator.SetMessages("zh", map[string]string{
        "required": "{{.Field}}是必填字段",
        "min":      "{{.Field}}最小长度为{{.Param}}",
        "max":      "{{.Field}}最大长度为{{.Param}}",
        "email":    "{{.Field}}必须是有效的邮箱地址",
    })
    
    // 英文消息
    validator.SetMessages("en", map[string]string{
        "required": "{{.Field}} is required",
        "min":      "{{.Field}} must be at least {{.Param}} characters",
        "max":      "{{.Field}} must be at most {{.Param}} characters",
        "email":    "{{.Field}} must be a valid email address",
    })
}
```

## 复合验证

### 组合验证规则

```go
// 定义验证组
type ValidationGroups struct {
    Basic    []string `json:"basic"`
    Advanced []string `json:"advanced"`
    Admin    []string `json:"admin"`
}

type User struct {
    Username string `json:"username" validate:"required,min=3" groups:"basic"`
    Email    string `json:"email" validate:"required,email" groups:"basic"`
    Password string `json:"password" validate:"required,strong_password" groups:"basic,advanced"`
    Role     string `json:"role" validate:"required,oneof=user admin" groups:"admin"`
    Profile  Profile `json:"profile" validate:"required" groups:"advanced"`
}

// 按组验证
func ValidateByGroup(data interface{}, group string) error {
    // 根据组标签进行验证
    // 实现略...
}
```

### 嵌套结构验证

```go
type Address struct {
    Street   string `json:"street" validate:"required,min=5"`
    City     string `json:"city" validate:"required"`
    Province string `json:"province" validate:"required"`
    ZipCode  string `json:"zip_code" validate:"required,len=6,numeric"`
    Country  string `json:"country" validate:"required,oneof=CN US UK"`
}

type Company struct {
    Name    string   `json:"name" validate:"required,min=2,max=100"`
    Address Address  `json:"address" validate:"required,dive"`
    Phones  []string `json:"phones" validate:"required,min=1,max=3,dive,mobile"`
    Email   string   `json:"email" validate:"required,email"`
}

type Employee struct {
    Name      string   `json:"name" validate:"required,chinese,min=2,max=10"`
    Age       int      `json:"age" validate:"required,gte=18,lte=65"`
    Company   Company  `json:"company" validate:"required,dive"`
    Addresses []Address `json:"addresses" validate:"omitempty,max=2,dive,required"`
}
```

## 异步验证

### 异步验证框架

```go
type AsyncValidator struct {
    validator *validator.Validate
    timeout   time.Duration
}

func NewAsyncValidator(timeout time.Duration) *AsyncValidator {
    return &AsyncValidator{
        validator: validator.New(),
        timeout:   timeout,
    }
}

// 异步验证方法
func (av *AsyncValidator) ValidateAsync(data interface{}) <-chan ValidationResult {
    result := make(chan ValidationResult, 1)
    
    go func() {
        defer close(result)
        
        ctx, cancel := context.WithTimeout(context.Background(), av.timeout)
        defer cancel()
        
        done := make(chan error, 1)
        go func() {
            done <- av.validator.Struct(data)
        }()
        
        select {
        case err := <-done:
            result <- ValidationResult{Error: err}
        case <-ctx.Done():
            result <- ValidationResult{Error: errors.New("validation timeout")}
        }
    }()
    
    return result
}

type ValidationResult struct {
    Error error
}

// 使用示例
func (c *UserController) RegisterAsync() {
    var req UserRegisterRequest
    if err := c.BindJSON(&req); err != nil {
        c.JSON(400, map[string]string{"error": "Invalid JSON"})
        return
    }
    
    validator := NewAsyncValidator(5 * time.Second)
    resultChan := validator.ValidateAsync(req)
    
    select {
    case result := <-resultChan:
        if result.Error != nil {
            c.JSON(400, map[string]interface{}{
                "error": "Validation failed",
                "details": FormatErrors(result.Error),
            })
            return
        }
        
        // 验证通过，继续处理
        // ...
    }
}
```

## 验证中间件

### 全局验证中间件

```go
// 验证中间件
func ValidationMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 只对 POST/PUT/PATCH 请求进行验证
        if c.Request.Method == "GET" || c.Request.Method == "DELETE" {
            c.Next()
            return
        }
        
        // 获取请求体
        body, err := io.ReadAll(c.Request.Body)
        if err != nil {
            c.JSON(400, map[string]string{"error": "Failed to read request body"})
            c.Abort()
            return
        }
        
        // 重置请求体
        c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
        
        // 根据路由确定验证规则
        validationRules := getValidationRules(c.FullPath())
        if validationRules != nil {
            if err := validateRequest(body, validationRules); err != nil {
                c.JSON(400, map[string]interface{}{
                    "error": "Validation failed",
                    "details": FormatErrors(err),
                })
                c.Abort()
                return
            }
        }
        
        c.Next()
    }
}
```

## 最佳实践

### 1. 验证规则组织

```go
// 将验证规则组织到单独的文件中
package validators

// 用户相关验证
var UserValidationRules = map[string]interface{}{
    "register": UserRegisterRequest{},
    "login":    UserLoginRequest{},
    "update":   UserUpdateRequest{},
}

// 产品相关验证
var ProductValidationRules = map[string]interface{}{
    "create": ProductCreateRequest{},
    "update": ProductUpdateRequest{},
}
```

### 2. 验证错误统一处理

```go
// 全局错误处理器
func HandleValidationError(c *gin.Context, err error) {
    if validationErrors, ok := err.(validator.ValidationErrors); ok {
        errors := FormatErrors(validationErrors)
        c.JSON(400, gin.H{
            "code":    40001,
            "message": "Validation failed",
            "errors":  errors,
        })
    } else {
        c.JSON(400, gin.H{
            "code":    40000,
            "message": "Bad request",
        })
    }
}
```

### 3. 验证性能优化

```go
// 使用单例验证器
var (
    validatorInstance *validator.Validate
    validatorOnce     sync.Once
)

func GetValidator() *validator.Validate {
    validatorOnce.Do(func() {
        validatorInstance = validator.New()
        
        // 注册自定义验证器
        registerCustomValidators(validatorInstance)
        
        // 注册标签名称函数
        validatorInstance.RegisterTagNameFunc(func(fld reflect.StructField) string {
            name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
            if name == "-" {
                return ""
            }
            return name
        })
    })
    
    return validatorInstance
}
```

### 4. 测试验证规则

```go
package validation_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestUserValidation(t *testing.T) {
    validator := GetValidator()
    
    tests := []struct {
        name    string
        user    UserRegisterRequest
        wantErr bool
    }{
        {
            name: "valid user",
            user: UserRegisterRequest{
                Username: "testuser",
                Email:    "test@example.com",
                Password: "StrongPass123!",
                Age:      25,
            },
            wantErr: false,
        },
        {
            name: "invalid email",
            user: UserRegisterRequest{
                Username: "testuser",
                Email:    "invalid-email",
                Password: "StrongPass123!",
                Age:      25,
            },
            wantErr: true,
        },
        {
            name: "weak password",
            user: UserRegisterRequest{
                Username: "testuser",
                Email:    "test@example.com",
                Password: "weak",
                Age:      25,
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validator.Struct(tt.user)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

YYHertz 的验证系统提供了全面而灵活的数据验证功能，从简单的标签验证到复杂的自定义验证器，能够满足各种应用场景的需求，确保数据的完整性和安全性。
