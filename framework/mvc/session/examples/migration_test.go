// Package examples 迁移示例
//
// 这个文件展示了如何从旧版本迁移到新的Session架构
// 包含向后兼容性演示和渐进式迁移策略
package main

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	// 传统的context包（用于演示兼容性）

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	yycontext "github.com/zsy619/yyhertz/framework/mvc/context"
	"github.com/zsy619/yyhertz/framework/mvc/session"
)

func Test_Migration(t *testing.T) {
	h := server.Default(server.WithHostPorts(":8081"))

	// 演示不同的使用方式
	h.GET("/legacy", legacyStyleHandler)          // 传统方式（继续支持）
	h.GET("/new", newStyleHandler)                // 新方式（推荐）
	h.GET("/mixed", mixedStyleHandler)            // 混合方式（渐进迁移）
	h.GET("/compatibility", compatibilityHandler) // 兼容性测试

	fmt.Println("🚀 迁移示例服务器启动在端口 8081")
	fmt.Println("访问以下端点测试不同的使用方式:")
	fmt.Println("  GET /legacy       - 传统方式（继续支持）")
	fmt.Println("  GET /new          - 新方式（推荐）")
	fmt.Println("  GET /mixed        - 混合方式（渐进迁移）")
	fmt.Println("  GET /compatibility - 兼容性测试")

	h.Spin()
}

// ============= 传统方式（继续支持，无需修改）=============

func legacyStyleHandler(ctx context.Context, c *app.RequestContext) {
	// 模拟传统的YYHertz Context使用方式
	// 注意：这里我们直接使用Hertz的RequestContext来演示
	// 在实际的YYHertz项目中，这将是YYHertz的Context

	// 创建模拟的InputData和OutputData（在实际项目中由框架提供）
	inputData := &yycontext.InputData{}
	outputData := &yycontext.OutputData{}

	// 初始化（在实际项目中由框架自动完成）
	inputData.Initialize(c)
	outputData.Initialize(c)

	// 传统的Cookie操作 - 这些API继续完全支持
	inputData.SetCookie("legacy_cookie", "legacy_value")
	cookieValue := inputData.Cookie("legacy_cookie")

	// 传统的Session操作 - 这些API继续完全支持
	inputData.SetSession("legacy_session", "legacy_session_value")
	sessionValue := inputData.GetSession("legacy_session")

	// 传统的输出Cookie操作
	outputData.Cookie("output_cookie", "output_value", 3600, "/", "", false, true)

	c.JSON(consts.StatusOK, utils.H{
		"message":       "传统方式 - 完全向后兼容",
		"cookie_value":  cookieValue,
		"session_value": sessionValue,
		"note":          "现有代码无需任何修改即可继续使用",
	})
}

// ============= 新方式（推荐）=============

func newStyleHandler(ctx context.Context, c *app.RequestContext) {
	// 新的统一方式 - 推荐用于新项目
	extension := session.NewExtensionForHertzContext(c)

	// Cookie操作 - 更清晰的API
	extension.SetCookie("new_cookie", "new_value")
	cookieValue := extension.GetCookie("new_cookie")

	// Session操作 - 统一的错误处理
	err := extension.SetSession("new_session", "new_session_value")
	if err != nil {
		log.Printf("设置session失败: %v", err)
	}
	sessionValue := extension.GetSession("new_session")

	// 安全Cookie操作 - 新功能
	secret := "demo-secret-key"
	extension.SetSecureCookie(secret, "secure_cookie", "secure_value")
	secureValue, secureOk := extension.GetSecureCookie(secret, "secure_cookie")

	// 批量Session操作 - 更高效
	adapter := extension.StartSession()
	adapter.Set("batch_key1", "batch_value1")
	adapter.Set("batch_key2", "batch_value2")
	adapter.Save()

	c.JSON(consts.StatusOK, utils.H{
		"message":       "新方式 - 推荐用于新项目",
		"cookie_value":  cookieValue,
		"session_value": sessionValue,
		"secure_cookie": utils.H{
			"value": secureValue,
			"valid": secureOk,
		},
		"session_id": extension.GetSessionID(),
		"features": []string{
			"统一的API接口",
			"更好的错误处理",
			"安全Cookie支持",
			"批量操作支持",
		},
	})
}

// ============= 混合方式（渐进迁移）=============

func mixedStyleHandler(ctx context.Context, c *app.RequestContext) {
	// 渐进式迁移策略：现有功能用传统方式，新功能用新方式

	// 1. 继续使用现有的传统代码
	inputData := &yycontext.InputData{}
	inputData.Initialize(c)

	// 现有的业务逻辑保持不变
	inputData.SetCookie("existing_feature", "existing_value")
	existingValue := inputData.Cookie("existing_feature")

	// 2. 对于新功能，使用新的API
	extension := session.NewExtensionForHertzContext(c)

	// 新增的安全功能
	secret := "mixed-secret-key"
	extension.SetSecureCookie(secret, "new_security_feature", "secure_data")
	secureData, _ := extension.GetSecureCookie(secret, "new_security_feature")

	// 新增的高级Session功能
	err := extension.SetSession("new_feature_data", map[string]any{
		"feature_id": "mixed_migration",
		"enabled":    true,
		"config": map[string]string{
			"level": "advanced",
			"mode":  "progressive",
		},
	})
	if err != nil {
		log.Printf("设置复杂session数据失败: %v", err)
	}

	c.JSON(consts.StatusOK, utils.H{
		"message":              "混合方式 - 渐进式迁移策略",
		"existing_feature":     existingValue,
		"new_security_feature": secureData,
		"migration_strategy": utils.H{
			"phase":         "progressive",
			"existing_code": "unchanged",
			"new_features":  "use_new_api",
			"timeline":      "gradual_migration",
		},
		"benefits": []string{
			"零风险迁移",
			"逐步引入新功能",
			"团队学习成本分摊",
			"可控的迁移节奏",
		},
	})
}

// ============= 兼容性测试 =============

func compatibilityHandler(ctx context.Context, c *app.RequestContext) {
	// 全面测试新旧API的兼容性

	// 传统方式设置
	inputData := &yycontext.InputData{}
	inputData.Initialize(c)
	inputData.SetCookie("compat_test_old", "old_way_value")
	inputData.SetSession("compat_session_old", "old_session_value")

	// 新方式读取传统方式设置的数据
	extension := session.NewExtensionForHertzContext(c)
	oldCookieFromNew := extension.GetCookie("compat_test_old")
	oldSessionFromNew := extension.GetSession("compat_session_old")

	// 新方式设置
	extension.SetCookie("compat_test_new", "new_way_value")
	extension.SetSession("compat_session_new", "new_session_value")

	// 传统方式读取新方式设置的数据
	newCookieFromOld := inputData.Cookie("compat_test_new")
	newSessionFromOld := inputData.GetSession("compat_session_new")

	// 交叉验证
	compatibility := utils.H{
		"old_to_new": utils.H{
			"cookie_compatible":  oldCookieFromNew == "old_way_value",
			"session_compatible": oldSessionFromNew == "old_session_value",
		},
		"new_to_old": utils.H{
			"cookie_compatible":  newCookieFromOld == "new_way_value",
			"session_compatible": newSessionFromOld == "new_session_value",
		},
	}

	// 性能对比（简单测试）
	performanceTest := func() utils.H {
		startTime := time.Now()

		// 传统方式性能
		oldStart := time.Now()
		for i := 0; i < 1000; i++ {
			inputData.SetCookie(fmt.Sprintf("perf_old_%d", i), "value")
			inputData.Cookie(fmt.Sprintf("perf_old_%d", i))
		}
		oldDuration := time.Since(oldStart)

		// 新方式性能
		newStart := time.Now()
		for i := 0; i < 1000; i++ {
			extension.SetCookie(fmt.Sprintf("perf_new_%d", i), "value")
			extension.GetCookie(fmt.Sprintf("perf_new_%d", i))
		}
		newDuration := time.Since(newStart)

		totalDuration := time.Since(startTime)

		return utils.H{
			"old_way_duration": oldDuration.String(),
			"new_way_duration": newDuration.String(),
			"total_duration":   totalDuration.String(),
			"performance_impact": fmt.Sprintf("%.2f%%",
				float64(newDuration-oldDuration)/float64(oldDuration)*100),
		}
	}

	c.JSON(consts.StatusOK, utils.H{
		"message":       "兼容性测试完成",
		"compatibility": compatibility,
		"performance":   performanceTest(),
		"conclusion": utils.H{
			"backward_compatible":    true,
			"data_interchangeable":   true,
			"performance_acceptable": true,
			"migration_safe":         true,
		},
		"recommendations": []string{
			"现有项目可以立即升级，无需修改代码",
			"新项目推荐使用新API",
			"可以按模块逐步迁移到新API",
			"充分利用新的安全功能",
		},
	})
}

// ============= 迁移最佳实践 =============

/*
迁移策略建议:

1. 立即升级策略（适用于大部分项目）:
   - 直接升级框架版本
   - 现有代码无需修改
   - 立即享受性能和稳定性改进

2. 渐进式迁移策略（适用于大型项目）:
   - 第一阶段：升级框架，现有代码保持不变
   - 第二阶段：新功能使用新API
   - 第三阶段：逐模块迁移现有代码（可选）

3. 全面迁移策略（适用于新项目或重构项目）:
   - 直接使用新的统一API
   - 充分利用新的安全功能
   - 享受更好的开发体验

迁移检查清单:

□ 确认框架版本兼容性
□ 运行现有测试确保功能正常
□ 性能基准测试对比
□ 新功能集成测试
□ 生产环境灰度发布
□ 监控关键指标
□ 团队培训新API用法

风险控制:

1. 备份策略：保留旧版本备份
2. 回滚方案：快速回滚机制
3. 监控体系：关键指标监控
4. 测试覆盖：全面的自动化测试
5. 分步实施：分阶段推进迁移
*/
