package validation

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/zsy619/yyhertz/framework/config"
)

// ============= 示例数据结构 =============

// User 用户结构体
type User struct {
	ID       uint      `json:"id" form:"id"`
	Username string    `json:"username" form:"username" validate:"required|minLength:3|maxLength:20|alphaNum"`
	Email    string    `json:"email" form:"email" validate:"required|email|maxLength:255"`
	Password string    `json:"password" form:"password" validate:"required|minLength:6|maxLength:100"`
	Mobile   string    `json:"mobile" form:"mobile" validate:"required|mobile"`
	Age      int       `json:"age" form:"age" validate:"required|integer|min:18|max:120"`
	Gender   string    `json:"gender" form:"gender" validate:"required|in:male,female,other"`
	Website  string    `json:"website" form:"website" validate:"url"`
	Bio      string    `json:"bio" form:"bio" validate:"maxLength:500"`
	Birthday time.Time `json:"birthday" form:"birthday" validate:"date"`
	Status   string    `json:"status" form:"status" validate:"required|in:active,inactive,banned"`
}

// Product 产品结构体
type Product struct {
	Name        string   `json:"name" form:"name" validate:"required|minLength:2|maxLength:100"`
	Price       float64  `json:"price" form:"price" validate:"required|decimal|positive"`
	Description string   `json:"description" form:"description" validate:"maxLength:1000"`
	Category    string   `json:"category" form:"category" validate:"required|in:electronics,clothing,books,home"`
	SKU         string   `json:"sku" form:"sku" validate:"required|regex:^[A-Z]{2}-\d{6}$"`
	Weight      float64  `json:"weight" form:"weight" validate:"positive"`
	InStock     bool     `json:"in_stock" form:"in_stock"`
	Tags        []string `json:"tags" form:"tags" validate:"maxLength:5"`
}

// Order 订单结构体
type Order struct {
	OrderID      string    `json:"order_id" form:"order_id" validate:"required|uuid"`
	CustomerID   uint      `json:"customer_id" form:"customer_id" validate:"required|positive"`
	Amount       float64   `json:"amount" form:"amount" validate:"required|decimal|positive"`
	Currency     string    `json:"currency" form:"currency" validate:"required|in:CNY,USD,EUR"`
	Status       string    `json:"status" form:"status" validate:"required|in:pending,paid,shipped,delivered,cancelled"`
	CreatedAt    time.Time `json:"created_at" form:"created_at" validate:"required|datetime"`
	ShippingAddr Address   `json:"shipping_address" form:"shipping_address"`
}

// Address 地址结构体
type Address struct {
	Street   string `json:"street" form:"street" validate:"required|maxLength:200"`
	City     string `json:"city" form:"city" validate:"required|maxLength:50"`
	Province string `json:"province" form:"province" validate:"required|maxLength:50"`
	ZipCode  string `json:"zip_code" form:"zip_code" validate:"required|zipCode"`
	Country  string `json:"country" form:"country" validate:"required|in:CN,US,UK,JP"`
}

// ============= 验证器使用示例 =============

// RunValidationExamples 运行验证器示例
func RunValidationExamples() error {
	config.Info("Starting validation examples...")

	// 1. 基本验证示例
	config.Info("=== Basic Validation Examples ===")
	if err := basicValidationExamples(); err != nil {
		return err
	}

	// 2. 结构体验证示例
	config.Info("=== Struct Validation Examples ===")
	if err := structValidationExamples(); err != nil {
		return err
	}

	// 3. 表单验证示例
	config.Info("=== Form Validation Examples ===")
	if err := formValidationExamples(); err != nil {
		return err
	}

	// 4. JSON绑定验证示例
	config.Info("=== JSON Binding Examples ===")
	if err := jsonBindingExamples(); err != nil {
		return err
	}

	// 5. 自定义验证器示例
	config.Info("=== Custom Validator Examples ===")
	if err := customValidatorExamples(); err != nil {
		return err
	}

	// 6. 消息国际化示例
	config.Info("=== Message Internationalization Examples ===")
	if err := i18nExamples(); err != nil {
		return err
	}

	// 7. 规则构建器示例
	config.Info("=== Rule Builder Examples ===")
	if err := ruleBuilderExamples(); err != nil {
		return err
	}

	config.Info("Validation examples completed successfully!")
	return nil
}

// basicValidationExamples 基本验证示例
func basicValidationExamples() error {
	validator := GetDefaultValidator()

	// 验证邮箱
	result := validator.Validate("test@example.com", "required|email")
	config.Infof("Email validation (valid): %t", result.Valid)

	result = validator.Validate("invalid-email", "required|email")
	config.Infof("Email validation (invalid): %t, errors: %v", result.Valid, result.Errors)

	// 验证数字范围
	result = validator.Validate(25, "required|min:18|max:65")
	config.Infof("Age validation (valid): %t", result.Valid)

	result = validator.Validate(15, "required|min:18|max:65")
	config.Infof("Age validation (invalid): %t, errors: %v", result.Valid, result.Errors)

	// 验证字符串长度
	result = validator.Validate("hello", "required|minLength:3|maxLength:10")
	config.Infof("String length validation (valid): %t", result.Valid)

	result = validator.Validate("hi", "required|minLength:3|maxLength:10")
	config.Infof("String length validation (invalid): %t, errors: %v", result.Valid, result.Errors)

	// 验证中国手机号
	result = validator.Validate("13812345678", "required|mobile")
	config.Infof("Mobile validation (valid): %t", result.Valid)

	result = validator.Validate("12812345678", "required|mobile")
	config.Infof("Mobile validation (invalid): %t, errors: %v", result.Valid, result.Errors)

	// 验证枚举值
	result = validator.Validate("active", "required|in:active,inactive,banned")
	config.Infof("Enum validation (valid): %t", result.Valid)

	result = validator.Validate("unknown", "required|in:active,inactive,banned")
	config.Infof("Enum validation (invalid): %t, errors: %v", result.Valid, result.Errors)

	return nil
}

// structValidationExamples 结构体验证示例
func structValidationExamples() error {
	validator := GetDefaultValidator()

	// 有效的用户数据
	validUser := User{
		Username: "john_doe",
		Email:    "john@example.com",
		Password: "password123",
		Mobile:   "13812345678",
		Age:      25,
		Gender:   "male",
		Website:  "https://johndoe.com",
		Bio:      "Software developer",
		Status:   "active",
	}

	result := validator.ValidateStruct(validUser)
	config.Infof("Valid user validation: %t", result.Valid)

	// 无效的用户数据
	invalidUser := User{
		Username: "jo",            // 太短
		Email:    "invalid-email", // 无效邮箱
		Password: "123",           // 太短
		Mobile:   "12812345678",   // 无效手机号
		Age:      15,              // 太小
		Gender:   "unknown",       // 无效枚举值
		Website:  "not-a-url",     // 无效URL
		Status:   "invalid",       // 无效状态
	}

	result = validator.ValidateStruct(invalidUser)
	config.Infof("Invalid user validation: %t", result.Valid)
	if !result.Valid {
		config.Infof("User validation errors:")
		for _, err := range result.Errors {
			config.Infof("  - %s: %s", err.Field, err.Message)
		}
	}

	// 产品验证示例
	product := Product{
		Name:        "iPhone 13",
		Price:       6999.00,
		Description: "Latest iPhone model",
		Category:    "electronics",
		SKU:         "IP-123456",
		Weight:      0.174,
		InStock:     true,
		Tags:        []string{"phone", "apple", "5g"},
	}

	result = validator.ValidateStruct(product)
	config.Infof("Product validation: %t", result.Valid)

	return nil
}

// formValidationExamples 表单验证示例
func formValidationExamples() error {
	// 模拟表单数据
	formData := MapFormData{
		"username": "john_doe",
		"email":    "john@example.com",
		"password": "password123",
		"mobile":   "13812345678",
		"age":      "25",
		"gender":   "male",
		"status":   "active",
	}

	// 定义验证规则
	rules := map[string]string{
		"username": "required|minLength:3|maxLength:20|alphaNum",
		"email":    "required|email|maxLength:255",
		"password": "required|minLength:6|maxLength:100",
		"mobile":   "required|mobile",
		"age":      "required|integer|min:18|max:120",
		"gender":   "required|in:male,female,other",
		"status":   "required|in:active,inactive,banned",
	}

	// 验证表单
	fv, err := ValidateForm(formData, rules)
	if err != nil {
		config.Infof("Form validation failed: %v", err)
		config.Infof("Form errors: %v", fv.GetErrors())
	} else {
		config.Info("Form validation passed")
	}

	// 绑定到结构体
	var user User
	if err := BindForm(formData, &user); err != nil {
		config.Infof("Form binding failed: %v", err)
	} else {
		config.Infof("Form binding successful: %+v", user)
	}

	return nil
}

// jsonBindingExamples JSON绑定示例
func jsonBindingExamples() error {
	// 有效的JSON数据
	validJSON := `{
		"username": "john_doe",
		"email": "john@example.com",
		"password": "password123",
		"mobile": "13812345678",
		"age": 25,
		"gender": "male",
		"status": "active"
	}`

	var user User
	if err := BindJSON([]byte(validJSON), &user); err != nil {
		config.Infof("JSON binding failed: %v", err)
	} else {
		config.Infof("JSON binding successful: %+v", user)
	}

	// 无效的JSON数据
	invalidJSON := `{
		"username": "jo",
		"email": "invalid-email",
		"password": "123",
		"mobile": "12812345678",
		"age": 15,
		"gender": "unknown",
		"status": "invalid"
	}`

	var invalidUser User
	if err := BindJSON([]byte(invalidJSON), &invalidUser); err != nil {
		config.Infof("JSON binding validation failed: %v", err)
	}

	return nil
}

// customValidatorExamples 自定义验证器示例
func customValidatorExamples() error {
	validator := GetDefaultValidator()

	// 注册自定义验证器：强密码验证
	validator.RegisterValidator("strongPassword", func(value any, param string) bool {
		str := fmt.Sprintf("%v", value)
		if len(str) < 8 {
			return false
		}

		hasUpper := false
		hasLower := false
		hasDigit := false
		hasSpecial := false

		for _, char := range str {
			switch {
			case char >= 'A' && char <= 'Z':
				hasUpper = true
			case char >= 'a' && char <= 'z':
				hasLower = true
			case char >= '0' && char <= '9':
				hasDigit = true
			case char >= 33 && char <= 126: // 可见ASCII字符
				if !((char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9')) {
					hasSpecial = true
				}
			}
		}

		return hasUpper && hasLower && hasDigit && hasSpecial
	})

	// 设置自定义消息
	validator.SetMessage("strongPassword", "密码必须包含大写字母、小写字母、数字和特殊字符")

	// 测试强密码验证
	result := validator.Validate("Password123!", "strongPassword")
	config.Infof("Strong password validation (valid): %t", result.Valid)

	result = validator.Validate("password", "strongPassword")
	config.Infof("Strong password validation (invalid): %t, errors: %v", result.Valid, result.Errors)

	// 注册自定义验证器：中国姓名验证
	validator.RegisterValidator("chineseName", func(value any, param string) bool {
		str := fmt.Sprintf("%v", value)
		if len(str) < 2 || len(str) > 10 {
			return false
		}

		// 简单的中文字符检查
		for _, char := range str {
			if char < 0x4e00 || char > 0x9fff {
				return false
			}
		}
		return true
	})

	validator.SetMessage("chineseName", "请输入有效的中文姓名")

	result = validator.Validate("张三", "chineseName")
	config.Infof("Chinese name validation (valid): %t", result.Valid)

	result = validator.Validate("John", "chineseName")
	config.Infof("Chinese name validation (invalid): %t, errors: %v", result.Valid, result.Errors)

	return nil
}

// i18nExamples 国际化示例
func i18nExamples() error {
	validator := GetDefaultValidator()

	// 测试中文消息
	validator.SetLocale("zh-CN")
	result := validator.ValidateField("用户名", "", "required")
	config.Infof("Chinese message: %s", result.Errors[0].Message)

	// 切换到英文
	validator.SetLocale("en-US")
	result = validator.ValidateField("username", "", "required")
	config.Infof("English message: %s", result.Errors[0].Message)

	// 自定义消息
	customMessages := map[string]string{
		"username.required": "Username cannot be empty",
		"email.required":    "Email address is required",
		"password.required": "Password is mandatory",
	}

	for tag, message := range customMessages {
		validator.SetMessage(tag, message)
	}

	result = validator.ValidateField("username", "", "required")
	config.Infof("Custom message: %s", result.Errors[0].Message)

	return nil
}

// ruleBuilderExamples 规则构建器示例
func ruleBuilderExamples() error {
	// 使用规则构建器创建复杂验证规则
	userRules := Rules().
		Required().
		MinLength(3).
		MaxLength(20).
		Regex("^[a-zA-Z0-9_]+$").
		Build()

	config.Infof("Built user rules: %s", userRules)

	emailRules := Rules().
		Required().
		Email().
		MaxLength(255).
		Build()

	config.Infof("Built email rules: %s", emailRules)

	ageRules := Rules().
		Required().
		Range(18, 120).
		Build()

	config.Infof("Built age rules: %s", ageRules)

	// 测试构建的规则
	validator := GetDefaultValidator()

	result := validator.ValidateField("username", "john_doe", userRules)
	config.Infof("Username validation with built rules: %t", result.Valid)

	result = validator.ValidateField("email", "john@example.com", emailRules)
	config.Infof("Email validation with built rules: %t", result.Valid)

	result = validator.ValidateField("age", 25, ageRules)
	config.Infof("Age validation with built rules: %t", result.Valid)

	return nil
}

// ============= 性能测试示例 =============

// BenchmarkValidation 验证性能测试
func BenchmarkValidation() {
	config.Info("=== Validation Performance Benchmark ===")

	validator := GetDefaultValidator()
	user := User{
		Username: "john_doe",
		Email:    "john@example.com",
		Password: "password123",
		Mobile:   "13812345678",
		Age:      25,
		Gender:   "male",
		Status:   "active",
	}

	// 预热
	for i := 0; i < 100; i++ {
		validator.ValidateStruct(user)
	}

	// 性能测试
	iterations := 10000
	start := time.Now()

	for i := 0; i < iterations; i++ {
		validator.ValidateStruct(user)
	}

	duration := time.Since(start)
	config.Infof("Validated %d structs in %v", iterations, duration)
	config.Infof("Average time per validation: %v", duration/time.Duration(iterations))
	config.Infof("Validations per second: %.2f", float64(iterations)/duration.Seconds())
}

// ============= 错误处理示例 =============

// ErrorHandlingExamples 错误处理示例
func ErrorHandlingExamples() {
	config.Info("=== Error Handling Examples ===")

	validator := GetDefaultValidator()

	// 创建一个有多个错误的用户
	user := User{
		Username: "jo",      // 太短
		Email:    "invalid", // 无效邮箱
		Password: "123",     // 太短
		Age:      15,        // 太小
		Gender:   "unknown", // 无效枚举
		Status:   "invalid", // 无效状态
	}

	result := validator.ValidateStruct(user)
	if !result.Valid {
		errors := ValidationErrors(result.Errors)

		// 显示所有错误
		config.Infof("All errors: %s", errors.Error())

		// 获取第一个错误
		if firstError := errors.First(); firstError != nil {
			config.Infof("First error: %s", firstError.Message)
		}

		// 按字段获取错误
		usernameErrors := errors.ByField("username")
		config.Infof("Username errors: %d", len(usernameErrors))

		// 转换为映射
		errorMap := errors.ToMap()
		config.Infof("Error map: %v", errorMap)

		// JSON序列化
		if jsonData, err := json.MarshalIndent(result, "", "  "); err == nil {
			config.Infof("Validation result JSON:\n%s", string(jsonData))
		}
	}
}

// ============= 便捷函数 =============

// CreateSampleUser 创建示例用户
func CreateSampleUser() User {
	return User{
		Username: "john_doe",
		Email:    "john@example.com",
		Password: "password123",
		Mobile:   "13812345678",
		Age:      25,
		Gender:   "male",
		Website:  "https://johndoe.com",
		Bio:      "Software developer",
		Status:   "active",
	}
}

// CreateSampleProduct 创建示例产品
func CreateSampleProduct() Product {
	return Product{
		Name:        "iPhone 13",
		Price:       6999.00,
		Description: "Latest iPhone model",
		Category:    "electronics",
		SKU:         "IP-123456",
		Weight:      0.174,
		InStock:     true,
		Tags:        []string{"phone", "apple", "5g"},
	}
}

// CreateSampleFormData 创建示例表单数据
func CreateSampleFormData() MapFormData {
	return MapFormData{
		"username": "john_doe",
		"email":    "john@example.com",
		"password": "password123",
		"mobile":   "13812345678",
		"age":      "25",
		"gender":   "male",
		"status":   "active",
	}
}
