// Package xvalidation 提供了数据验证的工具函数
//
// 主要功能：
//   - 基础验证：必填、长度、范围、格式等基本验证
//   - 字符串验证：邮箱、手机号、身份证、URL、IP等格式验证
//   - 数字验证：整数、浮点数、正数、负数、范围验证等
//   - 日期验证：日期格式、日期范围、年龄验证等
//   - 自定义验证：自定义验证规则、组合验证、条件验证等
//
// 基础用法：
//
//	import "your-project/framework/pkg/xvalidation"
//
//	// 基础验证
//	isRequired := xvalidation.Required("value")
//	isLengthValid := xvalidation.Length("hello", 3, 10)
//	isInRange := xvalidation.Range(25, 18, 65)
//
//	// 字符串格式验证
//	isEmail := xvalidation.Email("test@example.com")
//	isPhone := xvalidation.Phone("13800138000")
//	isURL := xvalidation.URL("https://example.com")
//	isIP := xvalidation.IP("192.168.1.1")
//
//	// 数字验证
//	isInt := xvalidation.IsInteger("123")
//	isFloat := xvalidation.IsFloat("123.45")
//	isPositive := xvalidation.Positive(10)
//	isInNumRange := xvalidation.NumberRange(5, 1, 10)
//
// 高级用法：
//
//	// 结构体验证
//	type User struct {
//		Name  string `validate:"required,length:2-20"`
//		Email string `validate:"required,email"`
//		Age   int    `validate:"required,range:18-100"`
//	}
//
//	user := User{Name: "John", Email: "john@example.com", Age: 25}
//	errors := xvalidation.ValidateStruct(user)
//
//	// 自定义验证器
//	validator := xvalidation.NewValidator()
//	validator.RegisterRule("password", func(value interface{}) bool {
//		// 自定义密码验证逻辑
//		return len(value.(string)) >= 8
//	})
//
//	// 条件验证
//	rules := xvalidation.Rules{
//		"username": []string{"required", "length:3-20", "alpha_num"},
//		"email":    []string{"required", "email"},
//		"age":      []string{"required", "integer", "range:1-120"},
//	}
//
//	data := map[string]interface{}{
//		"username": "john_doe",
//		"email":    "john@example.com",
//		"age":      25,
//	}
//
//	errors := xvalidation.ValidateMap(data, rules)
//
// 验证规则：
//
//	// 基础规则
//	required     - 必填
//	length:min-max - 长度范围
//	range:min-max  - 数值范围
//	in:val1,val2   - 枚举值
//
//	// 字符串规则
//	email        - 邮箱格式
//	phone        - 手机号格式
//	url          - URL格式
//	ip           - IP地址格式
//	alpha        - 只包含字母
//	alpha_num    - 只包含字母和数字
//	numeric      - 只包含数字
//
//	// 数字规则
//	integer      - 整数
//	float        - 浮点数
//	positive     - 正数
//	negative     - 负数
//	min:val      - 最小值
//	max:val      - 最大值
//
//	// 日期规则
//	date         - 日期格式
//	datetime     - 日期时间格式
//	before:date  - 早于指定日期
//	after:date   - 晚于指定日期
//
package xvalidation