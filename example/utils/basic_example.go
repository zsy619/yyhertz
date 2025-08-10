package main

import (
	"fmt"
	"log"
	"time"

	"github.com/zsy619/yyhertz/framework/util/mapper"
)

// User 源结构体 - 模拟数据库模型
type User struct {
	ID          int                    `json:"id"`
	Name        string                 `json:"name"`
	Email       string                 `json:"email"`
	Age         int                    `json:"age"`
	Score       float64                `json:"score"`
	IsActive    bool                   `json:"is_active"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   *time.Time             `json:"updated_at"`
	Tags        []string               `json:"tags"`
	Metadata    map[string]interface{} `json:"metadata"`
	Profile     UserProfile            `json:"profile"`
}

type UserProfile struct {
	Avatar      string   `json:"avatar"`
	Bio         string   `json:"bio"`
	SocialLinks []string `json:"social_links"`
}

// UserDTO 目标结构体 - 模拟API响应
type UserDTO struct {
	UserID       int                    `json:"id"`
	FullName     string                 `json:"name"`
	EmailAddress string                 `json:"email"`
	Years        int                    `json:"age"`
	Rating       float64                `json:"score"`
	Active       bool                   `json:"is_active"`
	Created      time.Time              `json:"created_at"`
	Modified     *time.Time             `json:"updated_at"`
	Labels       []string               `json:"tags"`
	ExtraData    map[string]interface{} `json:"metadata"`
	UserProfile  UserProfile            `json:"profile"`
}

// SimpleUser 简单结构体用于基础测试
type SimpleUser struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type SimpleUserDTO struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func main() {
	fmt.Println("🚀 YYHertz 高性能对象映射器示例")
	fmt.Println("=====================================")
	fmt.Println()

	// 创建测试数据
	now := time.Now()
	updateTime := now.Add(-1 * time.Hour)

	user := &User{
		ID:        12345,
		Name:      "张三",
		Email:     "zhangsan@example.com",
		Age:       28,
		Score:     95.8,
		IsActive:  true,
		CreatedAt: now.Add(-30 * 24 * time.Hour),
		UpdatedAt: &updateTime,
		Tags:      []string{"开发者", "Go", "微服务"},
		Metadata: map[string]interface{}{
			"department": "技术部",
			"level":      "高级工程师",
			"remote":     true,
		},
		Profile: UserProfile{
			Avatar: "https://example.com/avatar.jpg",
			Bio:    "全栈开发者，专注于Go和云原生技术",
			SocialLinks: []string{
				"https://github.com/zhangsan",
				"https://twitter.com/zhangsan",
			},
		},
	}

	// 运行所有示例
	basicMappingExample(user)
	strategyComparisonExample(user)
	batchMappingExample()
	customConfigExample(user)
	performanceStatsExample()
	errorHandlingExample()
	advancedFeaturesExample()

	fmt.Println("✅ 所有示例运行完成！")
}

// 1. 基础映射示例
func basicMappingExample(user *User) {
	fmt.Println("1️⃣ 基础映射示例")
	fmt.Println("===============")

	// 创建映射器
	m := mapper.NewMapper()

	// 执行映射
	var userDTO UserDTO
	start := time.Now()
	err := m.Map(user, &userDTO)
	duration := time.Since(start)

	if err != nil {
		log.Printf("❌ 映射失败: %v", err)
		return
	}

	fmt.Printf("⏱️  映射耗时: %v\n", duration)
	fmt.Printf("✅ 映射成功: ID=%d, 姓名=%s, 邮箱=%s, 活跃=%t\n", 
		userDTO.UserID, userDTO.FullName, userDTO.EmailAddress, userDTO.Active)
	fmt.Printf("📊 嵌套数据: Profile.Bio=%s\n", userDTO.UserProfile.Bio)
	fmt.Printf("🏷️  标签数量: %d\n", len(userDTO.Labels))
	fmt.Println()
}

// 2. 不同策略性能对比
func strategyComparisonExample(user *User) {
	fmt.Println("2️⃣ 映射策略性能对比")
	fmt.Println("==================")

	strategies := []struct {
		name     string
		strategy mapper.MappingStrategy
		emoji    string
	}{
		{"自动选择", mapper.StrategyAuto, "🤖"},
		{"反射映射", mapper.StrategyReflection, "🪞"},
		{"JSON映射", mapper.StrategyJSON, "📄"},
		{"代码生成", mapper.StrategyCodegen, "⚡"},
	}

	// 预热
	fmt.Println("🔄 预热各策略...")
	for _, s := range strategies {
		m := mapper.NewMapper(mapper.WithStrategy(s.strategy))
		var dto UserDTO
		m.Map(user, &dto)
	}

	// 性能测试
	iterations := 1000
	fmt.Printf("🏃 运行 %d 次映射操作...\n\n", iterations)

	for _, s := range strategies {
		m := mapper.NewMapper(mapper.WithStrategy(s.strategy))

		start := time.Now()
		successCount := 0
		for i := 0; i < iterations; i++ {
			var dto UserDTO
			err := m.Map(user, &dto)
			if err == nil {
				successCount++
			} else if s.strategy != mapper.StrategyCodegen {
				// 代码生成策略可能失败，其他策略失败需要记录
				log.Printf("⚠️  %s 策略执行失败: %v", s.name, err)
			}
		}
		duration := time.Since(start)

		avgTime := duration / time.Duration(successCount)
		opsPerSec := int64(successCount) * int64(time.Second) / int64(duration)

		fmt.Printf("%s %-8s: 成功率=%d%%, 平均耗时=%v, 速度=%d ops/sec\n",
			s.emoji, s.name, successCount*100/iterations, avgTime, opsPerSec)
	}
	fmt.Println()
}

// 3. 批量映射示例
func batchMappingExample() {
	fmt.Println("3️⃣ 批量映射示例")
	fmt.Println("================")

	// 创建测试数据
	users := make([]*User, 100)
	for i := range users {
		users[i] = &User{
			ID:        i + 1,
			Name:      fmt.Sprintf("用户%d", i+1),
			Email:     fmt.Sprintf("user%d@example.com", i+1),
			Age:       20 + i%40,
			Score:     80.0 + float64(i%20),
			IsActive:  i%2 == 0,
			CreatedAt: time.Now().Add(-time.Duration(i) * time.Hour),
		}
	}

	m := mapper.NewMapper()

	start := time.Now()
	var dtos []UserDTO
	err := m.MapSlice(users, &dtos)
	duration := time.Since(start)

	if err != nil {
		log.Printf("❌ 批量映射失败: %v", err)
		return
	}

	recordsPerSec := int64(len(users)) * int64(time.Second) / int64(duration)
	fmt.Printf("✅ 批量映射完成: 处理%d条记录, 耗时%v, 速度=%d records/sec\n",
		len(users), duration, recordsPerSec)

	fmt.Printf("📋 前3条结果:\n")
	for i := 0; i < 3 && i < len(dtos); i++ {
		fmt.Printf("   %d. ID=%d, 姓名=%s, 邮箱=%s, 活跃=%t\n",
			i+1, dtos[i].UserID, dtos[i].FullName, dtos[i].EmailAddress, dtos[i].Active)
	}
	fmt.Println()
}

// 4. 自定义配置示例
func customConfigExample(user *User) {
	fmt.Println("4️⃣ 自定义配置示例")
	fmt.Println("=================")

	// 配置1：忽略某些字段
	fmt.Println("🔒 配置1: 忽略敏感字段")
	mapper1 := mapper.NewMapper(
		mapper.WithIgnoreFields("Email", "Score"),
		mapper.WithTagName("json"),
	)

	var dto1 UserDTO
	err := mapper1.Map(user, &dto1)
	if err != nil {
		log.Printf("❌ 映射失败: %v", err)
	} else {
		fmt.Printf("   忽略Email和Score后: 邮箱='%s' (应为空), 评分=%f (应为0)\n", 
			dto1.EmailAddress, dto1.Rating)
	}

	// 配置2：使用不同的标签
	fmt.Println("🏷️  配置2: 自定义标签映射")
	mapper2 := mapper.NewMapper(
		mapper.WithTagName("json"),
		mapper.WithCaseSensitive(false),
	)

	var dto2 UserDTO
	err = mapper2.Map(user, &dto2)
	if err != nil {
		log.Printf("❌ 映射失败: %v", err)
	} else {
		fmt.Printf("   自定义标签映射成功: ID=%d, 姓名=%s\n", dto2.UserID, dto2.FullName)
	}

	// 配置3：深拷贝模式
	fmt.Println("🔄 配置3: 深拷贝vs浅拷贝")
	mapper3 := mapper.NewMapper(
		mapper.WithDeepCopy(true),
		mapper.WithMaxDepth(10),
	)

	var dto3 UserDTO
	err = mapper3.Map(user, &dto3)
	if err != nil {
		log.Printf("❌ 映射失败: %v", err)
	} else {
		fmt.Printf("   深拷贝映射成功: 标签数=%d, Metadata=%v\n", 
			len(dto3.Labels), len(dto3.ExtraData))
	}
	fmt.Println()
}

// 5. 性能统计示例
func performanceStatsExample() {
	fmt.Println("5️⃣ 性能统计示例")
	fmt.Println("===============")

	m := mapper.NewMapper()

	// 执行一些映射操作
	user := &User{ID: 1, Name: "测试用户", Email: "test@example.com", Age: 25}
	
	fmt.Println("🔄 执行测试映射...")
	for i := 0; i < 10; i++ {
		var dto UserDTO
		err := m.Map(user, &dto)
		if err != nil {
			log.Printf("⚠️  第%d次映射失败: %v", i+1, err)
		}
	}

	// 获取统计信息
	stats := m.GetStats()

	fmt.Printf("📊 映射统计报告:\n")
	fmt.Printf("   📈 总映射次数: %d\n", stats.TotalMaps)
	fmt.Printf("   ✅ 成功次数: %d\n", stats.SuccessfulMaps)
	fmt.Printf("   ❌ 失败次数: %d\n", stats.FailedMaps)
	fmt.Printf("   💾 缓存命中: %d\n", stats.CacheHits)
	fmt.Printf("   🔍 缓存未命中: %d\n", stats.CacheMisses)

	if stats.TotalMaps > 0 {
		successRate := float64(stats.SuccessfulMaps) / float64(stats.TotalMaps) * 100
		fmt.Printf("   📊 成功率: %.1f%%\n", successRate)
	}

	if stats.CacheHits+stats.CacheMisses > 0 {
		hitRate := float64(stats.CacheHits) / float64(stats.CacheHits+stats.CacheMisses) * 100
		fmt.Printf("   🎯 缓存命中率: %.1f%%\n", hitRate)
	}
	fmt.Println()
}

// 6. 错误处理示例
func errorHandlingExample() {
	fmt.Println("6️⃣ 错误处理示例")
	fmt.Println("===============")

	m := mapper.NewMapper()
	user := &User{ID: 1, Name: "测试用户"}

	fmt.Println("🧪 测试各种错误情况...")

	// 测试1：nil源对象
	fmt.Println("   1. nil源对象测试:")
	var dto1 UserDTO
	err := m.Map(nil, &dto1)
	if err != nil {
		fmt.Printf("      ❌ 预期错误: %v\n", err)
	}

	// 测试2：非指针目标
	fmt.Println("   2. 非指针目标测试:")
	var dto2 UserDTO
	err = m.Map(user, dto2) // 注意：这里没有&
	if err != nil {
		fmt.Printf("      ❌ 预期错误: %v\n", err)
	}

	// 测试3：类型兼容性测试
	fmt.Println("   3. 兼容类型映射测试:")
	simpleUser := &SimpleUser{ID: 999, Name: "简单用户", Age: 30}
	var simpleDTO SimpleUserDTO
	err = m.Map(simpleUser, &simpleDTO)
	if err != nil {
		fmt.Printf("      ❌ 映射失败: %v\n", err)
	} else {
		fmt.Printf("      ✅ 简单类型映射成功: %+v\n", simpleDTO)
	}

	fmt.Println("   4. Map类型映射测试:")
	mapData := make(map[string]interface{})
	err = m.Map(user, &mapData)
	if err != nil {
		fmt.Printf("      ⚠️  Map映射可能失败: %v\n", err)
	} else {
		fmt.Printf("      ✅ Map映射成功，键数量: %d\n", len(mapData))
	}
	fmt.Println()
}

// 7. 高级特性示例
func advancedFeaturesExample() {
	fmt.Println("7️⃣ 高级特性示例")
	fmt.Println("===============")

	// 特性1：链式配置
	fmt.Println("🔗 特性1: 链式配置构建")
	config := mapper.DefaultMapConfig().
		WithStrategy(mapper.StrategyReflection).
		WithTagName("json").
		WithCaseSensitive(false).
		WithMaxDepth(5)

	_ = mapper.NewMapper(mapper.WithConfig(config))
	fmt.Printf("   ✅ 链式配置完成: 策略=%d, 最大深度=%d\n", 
		config.Strategy, config.MaxDepth)

	// 特性2：全局映射函数
	fmt.Println("🌐 特性2: 全局映射函数")
	src := &SimpleUser{ID: 1, Name: "全局测试", Age: 25}
	var dst SimpleUserDTO
	
	err := mapper.Map(src, &dst)
	if err != nil {
		fmt.Printf("   ❌ 全局映射失败: %v\n", err)
	} else {
		fmt.Printf("   ✅ 全局映射成功: %+v\n", dst)
	}

	// 特性3：映射器设置策略
	fmt.Println("⚙️  特性3: 动态设置策略")
	m2 := mapper.NewMapper()
	fmt.Println("   🔄 切换到JSON策略...")
	m2.SetStrategy(mapper.StrategyJSON)
	
	var dst2 SimpleUserDTO
	err = m2.Map(src, &dst2)
	if err != nil {
		fmt.Printf("   ❌ 策略切换后映射失败: %v\n", err)
	} else {
		fmt.Printf("   ✅ 策略切换后映射成功: %+v\n", dst2)
	}

	fmt.Println()
}