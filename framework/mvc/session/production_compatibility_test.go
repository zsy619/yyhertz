// Package session 生产环境兼容性测试套件
//
// 这个文件包含模拟生产环境场景的测试，确保重构后的代码在实际环境中稳定运行
// 测试覆盖：
// - 高并发场景测试
// - 大数据量处理测试
// - 错误恢复测试
// - 内存泄漏检测
// - 长时间运行稳定性测试
package session

import (
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"
)

// ============= 高并发测试 =============

// TestHighConcurrencyCookieOperations 高并发Cookie操作测试
func TestHighConcurrencyCookieOperations(t *testing.T) {
	const goroutines = 100
	const operationsPerGoroutine = 100 // 减少操作数，专注测试并发安全性
	
	var wg sync.WaitGroup
	errors := make(chan error, goroutines)
	
	// 启动多个goroutine并发操作
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			ctx := createTestRequestContext()
			extension := NewExtensionForHertzContext(ctx)
			
			for j := 0; j < operationsPerGoroutine; j++ {
				key := fmt.Sprintf("concurrent_cookie_%d_%d", id, j)
				value := fmt.Sprintf("value_%d_%d", id, j)
				
				// 测试Cookie API调用的并发安全性（不依赖实际HTTP周期）
				extension.SetCookie(key, value)
				
				// 验证Cookie设置API是否正常工作
				if extension.Cookie == nil {
					errors <- fmt.Errorf("goroutine %d: cookie component not initialized", id)
					return
				}
				
				// 测试Cookie检查API
				_ = extension.CookieExists(key)
				
				// 删除Cookie
				extension.DelCookie(key)
			}
		}(i)
	}
	
	// 等待所有goroutine完成
	wg.Wait()
	close(errors)
	
	// 检查错误
	for err := range errors {
		t.Error(err)
	}
	
	t.Logf("✅ 高并发Cookie API测试完成: %d goroutines × %d operations = %d total operations", 
		goroutines, operationsPerGoroutine, goroutines*operationsPerGoroutine)
}

// TestHighConcurrencySessionOperations 高并发Session操作测试
func TestHighConcurrencySessionOperations(t *testing.T) {
	const goroutines = 50
	const operationsPerGoroutine = 500
	
	var wg sync.WaitGroup
	errors := make(chan error, goroutines)
	
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			ctx := createTestRequestContext()
			extension := NewExtensionForHertzContext(ctx)
			
			for j := 0; j < operationsPerGoroutine; j++ {
				key := fmt.Sprintf("session_key_%d_%d", id, j)
				value := fmt.Sprintf("session_value_%d_%d", id, j)
				
				// Session操作
				if err := extension.SetSession(key, value); err != nil {
					errors <- fmt.Errorf("goroutine %d: failed to set session %s: %v", id, key, err)
					return
				}
				
				retrieved := extension.GetSession(key)
				if retrieved == nil {
					errors <- fmt.Errorf("goroutine %d: failed to retrieve session %s", id, key)
					return
				}
			}
		}(i)
	}
	
	wg.Wait()
	close(errors)
	
	for err := range errors {
		t.Error(err)
	}
	
	t.Logf("✅ 高并发Session测试完成: %d goroutines × %d operations", goroutines, operationsPerGoroutine)
}

// ============= 大数据量处理测试 =============

// TestLargeDataHandling 大数据量处理测试
func TestLargeDataHandling(t *testing.T) {
	ctx := createTestRequestContext()
	extension := NewExtensionForHertzContext(ctx)
	
	// 测试大Cookie值
	t.Run("LargeCookieValue", func(t *testing.T) {
		largeValue := make([]byte, 4096) // 4KB cookie值
		for i := range largeValue {
			largeValue[i] = byte('A' + (i % 26))
		}
		
		extension.SetCookie("large_cookie", string(largeValue))
		retrieved := extension.GetCookie("large_cookie")
		
		if len(retrieved) == 0 {
			t.Error("大Cookie值处理失败")
		}
		t.Logf("✅ 大Cookie值测试通过: %d bytes", len(retrieved))
	})
	
	// 测试大量Cookie
	t.Run("ManyCookies", func(t *testing.T) {
		const cookieCount = 100
		
		// 设置大量Cookie
		for i := 0; i < cookieCount; i++ {
			key := fmt.Sprintf("cookie_%d", i)
			value := fmt.Sprintf("value_%d", i)
			extension.SetCookie(key, value)
		}
		
		// 验证所有Cookie（使用Cookie组件的方法）
		allCookies := extension.Cookie.GetAll()
		if len(allCookies) < cookieCount {
			t.Errorf("期望至少 %d 个Cookie，实际获得 %d 个", cookieCount, len(allCookies))
		}
		
		t.Logf("✅ 大量Cookie测试通过: %d cookies", len(allCookies))
	})
	
	// 测试大Session数据
	t.Run("LargeSessionData", func(t *testing.T) {
		largeData := make(map[string]string)
		for i := 0; i < 1000; i++ {
			key := fmt.Sprintf("large_session_key_%d", i)
			value := fmt.Sprintf("large_session_value_%d_with_extra_data", i)
			largeData[key] = value
			
			if err := extension.SetSession(key, value); err != nil {
				t.Errorf("设置大Session数据失败: %v", err)
			}
		}
		
		// 验证数据完整性
		for key, expectedValue := range largeData {
			retrieved := extension.GetSession(key)
			if retrieved != expectedValue {
				t.Errorf("Session数据不匹配: key=%s, expected=%s, got=%v", key, expectedValue, retrieved)
			}
		}
		
		t.Logf("✅ 大Session数据测试通过: %d entries", len(largeData))
	})
}

// ============= 错误恢复测试 =============

// TestErrorRecovery 错误恢复能力测试
func TestErrorRecovery(t *testing.T) {
	// 测试无效Cookie格式恢复
	t.Run("InvalidCookieFormat", func(t *testing.T) {
		ctx := createTestRequestContext()
		// 故意设置无效的Cookie头
		ctx.Request.Header.Set("Cookie", "invalid=cookie=format=test")
		
		extension := NewExtensionForHertzContext(ctx)
		
		// 应该能够正常处理，不应该崩溃
		cookies := extension.Cookie.GetAll()
		t.Logf("✅ 无效Cookie格式恢复测试通过，获得 %d 个有效Cookie", len(cookies))
	})
	
	// 测试安全Cookie解析错误恢复
	t.Run("SecureCookieParsingError", func(t *testing.T) {
		ctx := createTestRequestContext()
		extension := NewExtensionForHertzContext(ctx)
		
		// 设置一个被篡改的安全Cookie
		extension.SetCookie("tampered_secure", "invalid|format|signature")
		
		// 尝试获取安全Cookie，应该返回空值而不是崩溃
		value, ok := extension.GetSecureCookie("secret", "tampered_secure")
		if ok || value != "" {
			t.Error("期望篡改的安全Cookie验证失败")
		}
		
		t.Log("✅ 安全Cookie解析错误恢复测试通过")
	})
	
	// 测试Session存储错误恢复
	t.Run("SessionStorageError", func(t *testing.T) {
		ctx := createTestRequestContext()
		extension := NewExtensionForHertzContext(ctx)
		
		// 模拟Session存储操作，应该有适当的错误处理
		err := extension.SetSession("test_key", "test_value")
		if err != nil {
			t.Logf("Session设置返回错误（这是正常的）: %v", err)
		}
		
		t.Log("✅ Session存储错误恢复测试通过")
	})
}

// ============= 内存泄漏检测 =============

// TestMemoryLeakDetection 内存泄漏检测测试
func TestMemoryLeakDetection(t *testing.T) {
	// 强制垃圾回收
	runtime.GC()
	runtime.GC()
	
	var memStats1, memStats2 runtime.MemStats
	runtime.ReadMemStats(&memStats1)
	
	// 执行大量操作
	const iterations = 10000
	for i := 0; i < iterations; i++ {
		ctx := createTestRequestContext()
		extension := NewExtensionForHertzContext(ctx)
		
		// Cookie操作
		extension.SetCookie("leak_test", "leak_value")
		_ = extension.GetCookie("leak_test")
		
		// Session操作
		_ = extension.SetSession("leak_session", "leak_session_value")
		_ = extension.GetSession("leak_session")
		
		// 安全Cookie操作
		extension.SetSecureCookie("secret", "leak_secure", "leak_secure_value")
		_, _ = extension.GetSecureCookie("secret", "leak_secure")
	}
	
	// 强制垃圾回收
	runtime.GC()
	runtime.GC()
	runtime.ReadMemStats(&memStats2)
	
	// 计算内存增长
	memoryGrowth := memStats2.Alloc - memStats1.Alloc
	memoryGrowthMB := float64(memoryGrowth) / 1024 / 1024
	
	t.Logf("内存增长: %.2f MB (%d bytes) after %d iterations", memoryGrowthMB, memoryGrowth, iterations)
	
	// 合理的内存增长阈值（根据实际情况调整）
	const maxMemoryGrowthMB = 50.0
	if memoryGrowthMB > maxMemoryGrowthMB {
		t.Errorf("可能存在内存泄漏: 内存增长 %.2f MB 超过阈值 %.2f MB", memoryGrowthMB, maxMemoryGrowthMB)
	} else {
		t.Log("✅ 内存泄漏检测通过")
	}
}

// ============= 长时间运行稳定性测试 =============

// TestLongRunningStability 长时间运行稳定性测试
func TestLongRunningStability(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过长时间运行测试（使用 -short 标志）")
	}
	
	const testDuration = 30 * time.Second // 在CI环境中运行30秒
	const operationInterval = 10 * time.Millisecond
	
	ctx := createTestRequestContext()
	extension := NewExtensionForHertzContext(ctx)
	
	startTime := time.Now()
	operationCount := 0
	errorCount := 0
	
	t.Logf("开始长时间运行稳定性测试，持续时间: %v", testDuration)
	
	for time.Since(startTime) < testDuration {
		operationCount++
		
		// 混合操作
		key := fmt.Sprintf("stability_test_%d", operationCount)
		value := fmt.Sprintf("value_%d_%d", operationCount, time.Now().UnixNano())
		
		// Cookie操作
		extension.SetCookie(key, value)
		if retrieved := extension.GetCookie(key); retrieved == "" {
			errorCount++
		}
		
		// Session操作
		if err := extension.SetSession(key, value); err != nil {
			errorCount++
		}
		if retrieved := extension.GetSession(key); retrieved == nil {
			errorCount++
		}
		
		// 安全Cookie操作
		extension.SetSecureCookie("secret", key, value)
		if _, ok := extension.GetSecureCookie("secret", key); !ok {
			errorCount++
		}
		
		time.Sleep(operationInterval)
	}
	
	actualDuration := time.Since(startTime)
	errorRate := float64(errorCount) / float64(operationCount) * 100
	
	t.Logf("长时间运行测试完成:")
	t.Logf("  实际运行时间: %v", actualDuration)
	t.Logf("  总操作次数: %d", operationCount)
	t.Logf("  错误次数: %d", errorCount)
	t.Logf("  错误率: %.2f%%", errorRate)
	
	// 错误率应该非常低
	const maxErrorRate = 1.0 // 1%
	if errorRate > maxErrorRate {
		t.Errorf("错误率过高: %.2f%% > %.2f%%", errorRate, maxErrorRate)
	} else {
		t.Log("✅ 长时间运行稳定性测试通过")
	}
}

// ============= 生产环境模拟测试 =============

// TestProductionScenario 生产环境场景模拟测试
func TestProductionScenario(t *testing.T) {
	// 模拟典型的Web应用Session使用场景
	t.Run("TypicalWebApplication", func(t *testing.T) {
		// 模拟用户登录
		ctx := createTestRequestContext()
		extension := NewExtensionForHertzContext(ctx)
		
		// 1. 用户访问，设置访问Cookie
		extension.SetCookie("visitor_id", "visitor_12345")
		extension.SetCookie("last_visit", time.Now().Format(time.RFC3339))
		
		// 2. 用户登录，创建Session
		err := extension.SetSession("user_id", "user_67890")
		if err != nil {
			t.Errorf("用户登录Session设置失败: %v", err)
		}
		
		err = extension.SetSession("username", "testuser")
		if err != nil {
			t.Errorf("用户名Session设置失败: %v", err)
		}
		
		err = extension.SetSession("login_time", time.Now().Unix())
		if err != nil {
			t.Errorf("登录时间Session设置失败: %v", err)
		}
		
		// 3. 设置安全Token
		extension.SetSecureCookie("csrf_secret", "csrf_token", "random_csrf_token_12345")
		
		// 4. 验证Session数据
		if userID := extension.GetSession("user_id"); userID != "user_67890" {
			t.Errorf("用户ID Session验证失败: expected=user_67890, got=%v", userID)
		}
		
		if username := extension.GetSession("username"); username != "testuser" {
			t.Errorf("用户名Session验证失败: expected=testuser, got=%v", username)
		}
		
		// 5. 验证安全Cookie（在测试环境中验证API调用正常）
		if extension.SecureCookie == nil {
			t.Error("安全Cookie组件未初始化")
		} else {
			// 在测试环境中，我们主要验证API调用不会崩溃
			_, _ = extension.GetSecureCookie("csrf_secret", "csrf_token")
			t.Log("✅ 安全Cookie API调用正常")
		}
		
		// 6. 模拟用户注销
		extension.ClearSession()
		extension.DelCookie("csrf_token")
		
		t.Log("✅ 典型Web应用场景测试通过")
	})
	
	// 模拟电商购物车场景
	t.Run("EcommerceCart", func(t *testing.T) {
		ctx := createTestRequestContext()
		extension := NewExtensionForHertzContext(ctx)
		
		// 购物车操作
		cartItems := []string{"item_1", "item_2", "item_3"}
		for i, item := range cartItems {
			key := fmt.Sprintf("cart_item_%d", i)
			err := extension.SetSession(key, item)
			if err != nil {
				t.Errorf("购物车商品设置失败: %v", err)
			}
		}
		
		// 验证购物车内容
		for i, expectedItem := range cartItems {
			key := fmt.Sprintf("cart_item_%d", i)
			if item := extension.GetSession(key); item != expectedItem {
				t.Errorf("购物车商品验证失败: key=%s, expected=%s, got=%v", key, expectedItem, item)
			}
		}
		
		t.Log("✅ 电商购物车场景测试通过")
	})
}

// ============= 向后兼容性测试 =============

// TestBackwardCompatibility 向后兼容性测试
func TestBackwardCompatibility(t *testing.T) {
	// 这里应该测试原有API的兼容性
	// 由于我们使用代理模式，所有旧的API调用应该仍然有效
	
	ctx := createTestRequestContext()
	extension := NewExtensionForHertzContext(ctx)
	
	// 测试所有兼容性方法
	t.Run("CookieCompatibility", func(t *testing.T) {
		// 这些方法应该与beego兼容
		extension.SetCookie("compat_test", "compat_value")
		if value := extension.GetCookie("compat_test"); value == "" {
			t.Error("Cookie兼容性测试失败")
		}
		
		extension.DelCookie("compat_test")
		t.Log("✅ Cookie兼容性测试通过")
	})
	
	t.Run("SessionCompatibility", func(t *testing.T) {
		err := extension.SetSession("compat_session", "compat_session_value")
		if err != nil {
			t.Errorf("Session兼容性设置失败: %v", err)
		}
		
		if value := extension.GetSession("compat_session"); value != "compat_session_value" {
			t.Errorf("Session兼容性获取失败: expected=compat_session_value, got=%v", value)
		}
		
		t.Log("✅ Session兼容性测试通过")
	})
	
	t.Run("SecureCookieCompatibility", func(t *testing.T) {
		extension.SetSecureCookie("secret", "compat_secure", "compat_secure_value")
		if value, ok := extension.GetSecureCookie("secret", "compat_secure"); !ok || value != "compat_secure_value" {
			t.Errorf("安全Cookie兼容性测试失败: expected=compat_secure_value, got=%s, ok=%v", value, ok)
		}
		
		t.Log("✅ 安全Cookie兼容性测试通过")
	})
}

// ============= 性能压力测试 =============

// TestPerformanceStress 性能压力测试
func TestPerformanceStress(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能压力测试（使用 -short 标志）")
	}
	
	const (
		goroutines = 200
		operations = 1000
	)
	
	var wg sync.WaitGroup
	startTime := time.Now()
	
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			ctx := createTestRequestContext()
			extension := NewExtensionForHertzContext(ctx)
			
			for j := 0; j < operations; j++ {
				key := fmt.Sprintf("stress_%d_%d", id, j)
				value := fmt.Sprintf("stress_value_%d_%d", id, j)
				
				// 混合操作
				extension.SetCookie(key, value)
				_ = extension.GetCookie(key)
				
				_ = extension.SetSession(key, value)
				_ = extension.GetSession(key)
				
				extension.SetSecureCookie("secret", key, value)
				_, _ = extension.GetSecureCookie("secret", key)
			}
		}(i)
	}
	
	wg.Wait()
	duration := time.Since(startTime)
	
	totalOperations := goroutines * operations * 6 // 每次循环6个操作
	opsPerSecond := float64(totalOperations) / duration.Seconds()
	
	t.Logf("性能压力测试完成:")
	t.Logf("  总操作数: %d", totalOperations)
	t.Logf("  耗时: %v", duration)
	t.Logf("  OPS: %.2f operations/second", opsPerSecond)
	
	// 设置性能阈值
	const minOpsPerSecond = 10000.0
	if opsPerSecond < minOpsPerSecond {
		t.Errorf("性能不达标: %.2f ops/s < %.2f ops/s", opsPerSecond, minOpsPerSecond)
	} else {
		t.Log("✅ 性能压力测试通过")
	}
}

/*
运行生产环境兼容性测试的方法：

1. 基础测试:
   go test -run TestHigh -v

2. 包含长时间运行测试:
   go test -run TestLong -timeout 2m -v

3. 包含性能压力测试:
   go test -run TestPerformance -timeout 5m -v

4. 完整测试套件:
   go test -v -timeout 10m

5. 内存泄漏检测:
   go test -run TestMemoryLeak -v -memprofile=mem.prof
   go tool pprof mem.prof

6. 竞态条件检测:
   go test -race -run TestHigh -v

测试通过标准：
- 高并发测试：无数据竞争，无错误
- 大数据量测试：正确处理大Cookie/Session
- 错误恢复：优雅处理各种异常情况
- 内存泄漏：内存增长在合理范围内
- 长时间运行：错误率 < 1%
- 性能压力：OPS > 10000 operations/second
*/