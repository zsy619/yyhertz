package main

import (
	"testing"
	"time"

	"github.com/zsy619/yyhertz/framework/mvc/session"
)

// Test_SessionStorePool_Basic 测试Session存储池基础功能
func Test_SessionStorePool_Basic(t *testing.T) {
	t.Log("🔄 测试Session存储池基础功能")

	pool := session.GetSessionStorePool()

	// 测试创建Session
	sessionID1 := "test_session_001"
	store1 := pool.GetOrCreate(sessionID1)
	if store1 == nil {
		t.Error("创建Session存储失败")
		return
	}

	// 测试相同ID返回相同实例
	store1_again := pool.GetOrCreate(sessionID1)
	if store1 != store1_again {
		t.Error("相同SessionID应该返回相同的存储实例")
		return
	}

	// 设置数据
	store1.Set("adminId", int64(12345))
	store1.Set("username", "testuser")

	// 通过池再次获取并验证数据
	store1_verify := pool.GetOrCreate(sessionID1)
	adminId := store1_verify.Get("adminId")
	username := store1_verify.Get("username")

	if adminId != int64(12345) {
		t.Errorf("期望adminId为12345，实际为: %v", adminId)
	}

	if username != "testuser" {
		t.Errorf("期望username为testuser，实际为: %v", username)
	}

	t.Log("✅ Session存储池基础功能测试通过")
}

// Test_SessionStorePool_MultipleInstances 测试多个Session实例
func Test_SessionStorePool_MultipleInstances(t *testing.T) {
	t.Log("🔄 测试多个Session实例")

	pool := session.GetSessionStorePool()

	// 创建多个不同的Session
	sessionID1 := "test_session_001"
	sessionID2 := "test_session_002"

	store1 := pool.GetOrCreate(sessionID1)
	store2 := pool.GetOrCreate(sessionID2)

	if store1 == store2 {
		t.Error("不同SessionID应该返回不同的存储实例")
		return
	}

	// 设置不同的数据
	store1.Set("role", "admin")
	store2.Set("role", "user")

	// 验证数据隔离
	role1 := store1.Get("role")
	role2 := store2.Get("role")

	if role1 != "admin" {
		t.Errorf("Session1的role应该为admin，实际为: %v", role1)
	}

	if role2 != "user" {
		t.Errorf("Session2的role应该为user，实际为: %v", role2)
	}

	t.Log("✅ 多个Session实例测试通过")
}

// Test_SessionStorePool_Cleanup 测试Session清理
func Test_SessionStorePool_Cleanup(t *testing.T) {
	t.Log("🔄 测试Session清理功能")

	pool := session.GetSessionStorePool()

	// 创建一个Session并设置过期时间很短
	sessionID := "test_session_cleanup"
	store := pool.GetOrCreate(sessionID)
	store.Set("test", "value")

	// 手动设置创建时间为很久以前
	store.CreateTime = time.Now().Add(-2 * time.Hour)
	store.LastAccess = time.Now().Add(-2 * time.Hour)

	// 获取清理前的数量
	countBefore := pool.Count()

	// 执行清理（设置过期时间为1小时）
	cleaned := pool.Cleanup(1 * time.Hour)

	// 获取清理后的数量
	countAfter := pool.Count()

	if cleaned == 0 {
		t.Error("应该清理至少1个过期Session")
	}

	if countAfter >= countBefore {
		t.Error("清理后Session数量应该减少")
	}

	t.Logf("✅ Session清理测试通过，清理了%d个过期Session", cleaned)
}

// Test_SessionStorePool_PersistentData 测试数据持久性
func Test_SessionStorePool_PersistentData(t *testing.T) {
	t.Log("🔄 测试Session数据持久性（模拟过滤器场景）")

	pool := session.GetSessionStorePool()
	sessionID := "test_filter_session"

	// 模拟控制器设置Session数据
	store1 := pool.GetOrCreate(sessionID)
	store1.Set("adminId", int64(67890))
	store1.Set("loginTime", time.Now().Unix())

	// 模拟过滤器获取Session数据（新的获取操作）
	store2 := pool.GetOrCreate(sessionID)
	adminId, ok := store2.Get("adminId").(int64)
	if !ok || adminId != 67890 {
		t.Errorf("过滤器中获取不到Session数据，期望adminId=67890，实际: %v", adminId)
		return
	}

	loginTime := store2.Get("loginTime")
	if loginTime == nil {
		t.Error("过滤器中获取不到loginTime数据")
		return
	}

	t.Log("✅ Session数据持久性测试通过，过滤器能正确获取到Session数据")
}

// Test_NewMemoryStore_BackwardCompatibility 测试向后兼容性
func Test_NewMemoryStore_BackwardCompatibility(t *testing.T) {
	t.Log("🔄 测试NewMemoryStore向后兼容性")

	sessionID := "test_backward_compat"

	// 使用旧的NewMemoryStore函数
	store1 := session.NewMemoryStore(sessionID)
	store1.Set("test", "value1")

	// 使用新的GetOrCreateMemoryStore函数
	store2 := session.GetOrCreateMemoryStore(sessionID)

	// 应该是同一个实例
	if store1 != store2 {
		t.Error("NewMemoryStore和GetOrCreateMemoryStore应该返回相同的实例")
		return
	}

	// 验证数据
	value := store2.Get("test")
	if value != "value1" {
		t.Errorf("期望值为value1，实际为: %v", value)
	}

	t.Log("✅ 向后兼容性测试通过")
}

// Test_SessionStorePool_Concurrent 测试并发安全
func Test_SessionStorePool_Concurrent(t *testing.T) {
	t.Log("🔄 测试Session存储池并发安全")

	pool := session.GetSessionStorePool()
	sessionID := "test_concurrent"

	// 启动多个goroutine并发操作
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			store := pool.GetOrCreate(sessionID)
			store.Set("counter", id)
			_ = store.Get("counter")
			done <- true
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < 10; i++ {
		<-done
	}

	// 验证Session仍然正常
	store := pool.GetOrCreate(sessionID)
	if store == nil {
		t.Error("并发操作后Session存储丢失")
		return
	}

	t.Log("✅ 并发安全测试通过")
}
