package main

import (
	"testing"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/zsy619/yyhertz/framework/mvc/session"
)

// Test_FilterSessionScenario 测试过滤器场景下的Session获取
func Test_FilterSessionScenario(t *testing.T) {
	t.Log("🔄 测试过滤器场景下的Session获取")

	// 模拟一个完整的请求场景
	sessionID := "filter_test_session_123"

	// ============= 第一步：模拟控制器设置Session =============
	t.Log("步骤1: 模拟控制器设置Session数据")

	// 直接使用全局存储池设置Session数据（模拟真实情况）
	store := session.GetOrCreateMemoryStore(sessionID)
	store.Set("adminId", int64(12345))
	store.Set("username", "admin_user")

	t.Log("✓ 控制器成功设置Session数据")

	// ============= 第二步：模拟过滤器获取Session =============
	t.Log("步骤2: 模拟过滤器获取Session数据")

	// 模拟过滤器中从全局存储池获取Session数据
	filterStore := session.GetOrCreateMemoryStore(sessionID)

	// 尝试获取Session数据
	adminId := filterStore.Get("adminId")
	username := filterStore.Get("username")

	// ============= 第三步：验证结果 =============
	t.Log("步骤3: 验证过滤器能否获取到Session数据")

	if adminId == nil {
		t.Error("❌ 过滤器中获取不到adminId Session数据")
		return
	}

	if adminIdValue, ok := adminId.(int64); !ok || adminIdValue != 12345 {
		t.Errorf("❌ 过滤器中获取的adminId数据不正确，期望: 12345, 实际: %v", adminId)
		return
	}

	if username == nil {
		t.Error("❌ 过滤器中获取不到username Session数据")
		return
	}

	if usernameValue, ok := username.(string); !ok || usernameValue != "admin_user" {
		t.Errorf("❌ 过滤器中获取的username数据不正确，期望: admin_user, 实际: %v", username)
		return
	}

	t.Log("✅ 过滤器成功获取到Session数据:")
	t.Logf("  - adminId: %v", adminId)
	t.Logf("  - username: %v", username)
	t.Log("✅ 过滤器场景测试通过，问题已修复！")
}

// Test_SessionManagerScenario 测试使用SessionManager的场景
func Test_SessionManagerScenario(t *testing.T) {
	t.Log("🔄 测试使用SessionManager的过滤器场景")

	sessionID := "manager_filter_test_session_456"

	// ============= 第一步：使用SessionManager设置Session =============
	t.Log("步骤1: 通过SessionManager设置Session数据")

	manager := session.NewSessionManager(nil)

	// 创建Session
	adapter1 := manager.CreateSession(nil)
	if adapter1 == nil {
		t.Error("创建Session失败")
		return
	}

	// 设置具体的SessionID（模拟从Cookie获取的情况）
	adapter1.Store.(*session.MemoryStore).SessionID = sessionID
	session.GetSessionStorePool().Stores.Store(sessionID, adapter1.Store)

	// 设置数据
	err := adapter1.Set("role", "administrator")
	if err != nil {
		t.Errorf("设置Session数据失败: %v", err)
		return
	}

	err = adapter1.Set("permissions", []string{"read", "write", "delete"})
	if err != nil {
		t.Errorf("设置Session数据失败: %v", err)
		return
	}

	t.Log("✓ 通过SessionManager成功设置Session数据")

	// ============= 第二步：模拟过滤器获取Session =============
	t.Log("步骤2: 通过SessionManager获取Session数据")

	// 模拟过滤器中获取现有Session
	adapter2 := manager.GetSession(nil, sessionID)
	if adapter2 == nil {
		t.Error("❌ 过滤器中获取不到Session")
		return
	}

	// 尝试获取数据
	role := adapter2.Get("role")
	permissions := adapter2.Get("permissions")

	// ============= 第三步：验证结果 =============
	t.Log("步骤3: 验证通过SessionManager获取的Session数据")

	if role == nil {
		t.Error("❌ 过滤器中获取不到role Session数据")
		return
	}

	if roleValue, ok := role.(string); !ok || roleValue != "administrator" {
		t.Errorf("❌ 获取的role数据不正确，期望: administrator, 实际: %v", role)
		return
	}

	if permissions == nil {
		t.Error("❌ 过滤器中获取不到permissions Session数据")
		return
	}

	if permissionsValue, ok := permissions.([]string); !ok || len(permissionsValue) != 3 {
		t.Errorf("❌ 获取的permissions数据不正确，期望: [read write delete], 实际: %v", permissions)
		return
	}

	t.Log("✅ 通过SessionManager成功获取到Session数据:")
	t.Logf("  - role: %v", role)
	t.Logf("  - permissions: %v", permissions)
	t.Log("✅ SessionManager场景测试通过！")
}

// Test_SessionConsistencyAcrossRequests 测试跨请求的Session一致性
func Test_SessionConsistencyAcrossRequests(t *testing.T) {
	t.Log("🔄 测试跨请求的Session一致性")

	sessionID := "consistency_test_session_789"

	// 请求1：设置初始数据
	hertzCtx1 := &app.RequestContext{}
	hertzCtx1.Request.Header.SetCookie("GOSESSIONID", sessionID)
	extension1 := session.NewExtensionForHertzContext(hertzCtx1)

	extension1.SetSession("counter", 1)
	extension1.SetSession("data", "initial")

	// 请求2：修改数据
	hertzCtx2 := &app.RequestContext{}
	hertzCtx2.Request.Header.SetCookie("GOSESSIONID", sessionID)
	extension2 := session.NewExtensionForHertzContext(hertzCtx2)

	counter := extension2.GetSession("counter")
	if counter != 1 {
		t.Errorf("请求2中获取的counter不正确，期望: 1, 实际: %v", counter)
		return
	}

	extension2.SetSession("counter", 2)
	extension2.SetSession("data", "modified")

	// 请求3：验证修改结果
	hertzCtx3 := &app.RequestContext{}
	hertzCtx3.Request.Header.SetCookie("GOSESSIONID", sessionID)
	extension3 := session.NewExtensionForHertzContext(hertzCtx3)

	finalCounter := extension3.GetSession("counter")
	finalData := extension3.GetSession("data")

	if finalCounter != 2 {
		t.Errorf("请求3中获取的counter不正确，期望: 2, 实际: %v", finalCounter)
		return
	}

	if finalData != "modified" {
		t.Errorf("请求3中获取的data不正确，期望: modified, 实际: %v", finalData)
		return
	}

	t.Log("✅ 跨请求Session一致性测试通过")
}

// Test_SessionIsolation 测试不同Session之间的隔离性
func Test_SessionIsolation(t *testing.T) {
	t.Log("🔄 测试不同Session之间的隔离性")

	sessionID1 := "isolation_test_session_1"
	sessionID2 := "isolation_test_session_2"

	// Session 1 设置数据
	hertzCtx1 := &app.RequestContext{}
	hertzCtx1.Request.Header.SetCookie("GOSESSIONID", sessionID1)
	extension1 := session.NewExtensionForHertzContext(hertzCtx1)
	extension1.SetSession("user", "user1")
	extension1.SetSession("role", "admin")

	// Session 2 设置不同数据
	hertzCtx2 := &app.RequestContext{}
	hertzCtx2.Request.Header.SetCookie("GOSESSIONID", sessionID2)
	extension2 := session.NewExtensionForHertzContext(hertzCtx2)
	extension2.SetSession("user", "user2")
	extension2.SetSession("role", "guest")

	// 验证Session隔离
	user1 := extension1.GetSession("user")
	role1 := extension1.GetSession("role")
	user2 := extension2.GetSession("user")
	role2 := extension2.GetSession("role")

	if user1 != "user1" || role1 != "admin" {
		t.Errorf("Session1数据不正确，user: %v, role: %v", user1, role1)
		return
	}

	if user2 != "user2" || role2 != "guest" {
		t.Errorf("Session2数据不正确，user: %v, role: %v", user2, role2)
		return
	}

	t.Log("✅ Session隔离性测试通过")
}
