package context

import (
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
)

// TestContextEnhancedFeatures 测试Context增强功能
func TestContextEnhancedFeatures(t *testing.T) {
	// 创建测试用的Context
	reqCtx := &app.RequestContext{}
	ctx := NewContext(reqCtx)
	defer ctx.Release()

	t.Run("测试带过期时间的键值对", func(t *testing.T) {
		// 设置带过期时间的值
		ctx.SetWithExpiry("test_key", "test_value", 100*time.Millisecond)
		
		// 立即获取应该成功
		value, exists := ctx.GetWithExpiry("test_key")
		if !exists || value.(string) != "test_value" {
			t.Error("Expected to get test_value immediately")
		}
		
		// 等待过期
		time.Sleep(150 * time.Millisecond)
		
		// 过期后获取应该失败
		_, exists = ctx.GetWithExpiry("test_key")
		if exists {
			t.Error("Expected key to be expired")
		}
	})

	t.Run("测试计数器功能", func(t *testing.T) {
		// 测试递增计数器
		count1 := ctx.IncrementCounter("counter")
		if count1 != 1 {
			t.Errorf("Expected counter to be 1, got %d", count1)
		}
		
		count2 := ctx.IncrementCounter("counter")
		if count2 != 2 {
			t.Errorf("Expected counter to be 2, got %d", count2)
		}
		
		// 测试递减计数器
		count3 := ctx.DecrementCounter("counter")
		if count3 != 1 {
			t.Errorf("Expected counter to be 1, got %d", count3)
		}
		
		// 测试获取计数器
		currentCount := ctx.GetCounter("counter")
		if currentCount != 1 {
			t.Errorf("Expected counter to be 1, got %d", currentCount)
		}
	})

	t.Run("测试链式操作", func(t *testing.T) {
		chain := ctx.NewChain()
		success, err := chain.
			Set("key1", "value1").
			Set("key2", "value2").
			SetIf(true, "key3", "value3").
			SetIf(false, "key4", "value4").
			Result()
		
		if !success {
			t.Errorf("Chain operation failed: %v", err)
		}
		
		// 验证值是否正确设置
		if val, _ := ctx.Get("key1"); val.(string) != "value1" {
			t.Error("key1 not set correctly")
		}
		if val, _ := ctx.Get("key2"); val.(string) != "value2" {
			t.Error("key2 not set correctly")
		}
		if val, _ := ctx.Get("key3"); val.(string) != "value3" {
			t.Error("key3 not set correctly")
		}
		if _, exists := ctx.Get("key4"); exists {
			t.Error("key4 should not exist")
		}
	})

	t.Run("测试事务操作", func(t *testing.T) {
		// 设置初始值
		ctx.Set("tx_key1", "initial1")
		ctx.Set("tx_key2", "initial2")
		
		// 开始事务
		tx := ctx.BeginTransaction()
		
		// 在事务中修改值
		tx.Set("tx_key1", "modified1")
		tx.Set("tx_key3", "new_value")
		
		// 回滚事务
		tx.Rollback()
		
		// 验证值被回滚
		if val, _ := ctx.Get("tx_key1"); val.(string) != "initial1" {
			t.Error("Transaction rollback failed for tx_key1")
		}
		if val, _ := ctx.Get("tx_key2"); val.(string) != "initial2" {
			t.Error("Transaction rollback failed for tx_key2")
		}
		if _, exists := ctx.Get("tx_key3"); exists {
			t.Error("tx_key3 should not exist after rollback")
		}
	})

	t.Run("测试批量操作", func(t *testing.T) {
		batch := ctx.NewBatch()
		batch.
			AddSet("batch_key1", "batch_value1").
			AddSet("batch_key2", "batch_value2").
			AddDelete("tx_key1") // 删除之前测试中的键
		
		batch.Execute()
		
		// 验证批量操作结果
		if val, _ := ctx.Get("batch_key1"); val.(string) != "batch_value1" {
			t.Error("Batch set failed for batch_key1")
		}
		if val, _ := ctx.Get("batch_key2"); val.(string) != "batch_value2" {
			t.Error("Batch set failed for batch_key2")
		}
		if _, exists := ctx.Get("tx_key1"); exists {
			t.Error("Batch delete failed for tx_key1")
		}
	})

	t.Run("测试快照功能", func(t *testing.T) {
		// 设置一些测试数据
		ctx.Set("snapshot_key1", "snapshot_value1")
		ctx.Set("snapshot_key2", "snapshot_value2")
		
		// 创建快照
		snapshot := ctx.TakeSnapshot()
		
		// 验证快照内容
		if snapshot.Timestamp.IsZero() {
			t.Error("Snapshot timestamp should not be zero")
		}
		
		if len(snapshot.Keys) < 2 {
			t.Errorf("Expected at least 2 keys in snapshot, got %d", len(snapshot.Keys))
		}
		
		if snapshot.Keys["snapshot_key1"] != "snapshot_value1" {
			t.Error("Snapshot failed to capture snapshot_key1")
		}
		
		if snapshot.KeysCount != ctx.KeysCount() {
			t.Error("Snapshot keys count mismatch")
		}
	})
}

// TestObserverPattern 测试观察者模式
func TestObserverPattern(t *testing.T) {
	reqCtx := &app.RequestContext{}
	ctx := NewContext(reqCtx)
	defer ctx.Release()

	// 创建测试观察者
	changeCount := 0
	deleteCount := 0
	
	observer := &testObserver{
		onChanged: func(ctx *Context, key string, oldValue, newValue any) {
			changeCount++
		},
		onDeleted: func(ctx *Context, key string, oldValue any) {
			deleteCount++
		},
	}
	
	// 添加观察者
	ctx.AddObserver("observed_key", observer)
	
	// 触发变更事件
	ctx.SetWithNotify("observed_key", "value1")
	ctx.SetWithNotify("observed_key", "value2")
	
	// 触发删除事件
	ctx.DeleteWithNotify("observed_key")
	
	// 验证观察者被正确调用
	if changeCount != 2 {
		t.Errorf("Expected 2 change events, got %d", changeCount)
	}
	
	if deleteCount != 1 {
		t.Errorf("Expected 1 delete event, got %d", deleteCount)
	}
}

// testObserver 测试用观察者实现
type testObserver struct {
	onChanged func(ctx *Context, key string, oldValue, newValue any)
	onDeleted func(ctx *Context, key string, oldValue any)
}

func (o *testObserver) OnKeyChanged(ctx *Context, key string, oldValue, newValue any) {
	if o.onChanged != nil {
		o.onChanged(ctx, key, oldValue, newValue)
	}
}

func (o *testObserver) OnKeyDeleted(ctx *Context, key string, oldValue any) {
	if o.onDeleted != nil {
		o.onDeleted(ctx, key, oldValue)
	}
}

// TestCacheStrategy 测试缓存策略
func TestCacheStrategy(t *testing.T) {
	reqCtx := &app.RequestContext{}
	ctx := NewContext(reqCtx)
	defer ctx.Release()

	// 创建简单缓存策略
	strategy := &SimpleCacheStrategy{
		DefaultTTL: 50 * time.Millisecond,
		MaxSize:    100,
	}
	
	// 使用缓存策略设置值
	ctx.SetWithCache("cache_key", "cache_value", strategy)
	
	// 立即获取应该成功
	value, exists := ctx.GetWithExpiry("cache_key")
	if !exists || value.(string) != "cache_value" {
		t.Error("Cache strategy failed to set value")
	}
	
	// 等待过期
	time.Sleep(80 * time.Millisecond)
	
	// 过期后获取应该失败
	_, exists = ctx.GetWithExpiry("cache_key")
	if exists {
		t.Error("Cached value should have expired")
	}
}