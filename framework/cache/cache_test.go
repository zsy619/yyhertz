package cache

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCacheManager(t *testing.T) {
	t.Run("创建缓存管理器", func(t *testing.T) {
		cache := NewCacheManager[string]("test", "测试缓存")
		
		assert.NotNil(t, cache)
		assert.Equal(t, "test", cache.name)
		assert.Equal(t, "测试缓存", cache.desc)
		assert.NotNil(t, cache.items)
		assert.Equal(t, 0, cache.Count())
	})
}

func TestCacheSet(t *testing.T) {
	cache := NewCacheManager[string]("test", "测试缓存")
	
	t.Run("设置无过期时间的缓存", func(t *testing.T) {
		cache.Set("key1", "value1", 0)
		
		value, found := cache.Get("key1")
		assert.True(t, found)
		assert.Equal(t, "value1", value)
	})
	
	t.Run("设置有过期时间的缓存", func(t *testing.T) {
		cache.Set("key2", "value2", time.Minute)
		
		value, found := cache.Get("key2")
		assert.True(t, found)
		assert.Equal(t, "value2", value)
	})
	
	t.Run("覆盖已存在的缓存", func(t *testing.T) {
		cache.Set("key1", "new_value1", 0)
		
		value, found := cache.Get("key1")
		assert.True(t, found)
		assert.Equal(t, "new_value1", value)
	})
}

func TestCacheGet(t *testing.T) {
	cache := NewCacheManager[int]("test", "测试缓存")
	
	t.Run("获取存在的缓存", func(t *testing.T) {
		cache.Set("number", 42, 0)
		
		value, found := cache.Get("number")
		assert.True(t, found)
		assert.Equal(t, 42, value)
	})
	
	t.Run("获取不存在的缓存", func(t *testing.T) {
		value, found := cache.Get("nonexistent")
		assert.False(t, found)
		assert.Equal(t, 0, value) // int类型的零值
	})
	
	t.Run("获取已过期的缓存", func(t *testing.T) {
		cache.Set("expired", 100, time.Millisecond)
		
		// 等待过期
		time.Sleep(time.Millisecond * 2)
		
		value, found := cache.Get("expired")
		assert.False(t, found)
		assert.Equal(t, 0, value)
	})
}

func TestCacheDelete(t *testing.T) {
	cache := NewCacheManager[string]("test", "测试缓存")
	
	t.Run("删除存在的缓存", func(t *testing.T) {
		cache.Set("delete_me", "value", 0)
		
		// 确认存在
		_, found := cache.Get("delete_me")
		assert.True(t, found)
		
		// 删除
		cache.Delete("delete_me")
		
		// 确认已删除
		_, found = cache.Get("delete_me")
		assert.False(t, found)
	})
	
	t.Run("删除不存在的缓存", func(t *testing.T) {
		// 删除不存在的key不应该panic
		assert.NotPanics(t, func() {
			cache.Delete("nonexistent")
		})
	})
}

func TestCacheExists(t *testing.T) {
	cache := NewCacheManager[bool]("test", "测试缓存")
	
	t.Run("检查存在的缓存", func(t *testing.T) {
		cache.Set("exists", true, 0)
		assert.True(t, cache.Exists("exists"))
	})
	
	t.Run("检查不存在的缓存", func(t *testing.T) {
		assert.False(t, cache.Exists("nonexistent"))
	})
	
	t.Run("检查已过期的缓存", func(t *testing.T) {
		cache.Set("expired", false, time.Millisecond)
		time.Sleep(time.Millisecond * 2)
		assert.False(t, cache.Exists("expired"))
	})
}

func TestCacheClear(t *testing.T) {
	cache := NewCacheManager[string]("test", "测试缓存")
	
	t.Run("清空缓存", func(t *testing.T) {
		// 添加一些数据
		cache.Set("key1", "value1", 0)
		cache.Set("key2", "value2", 0)
		cache.Set("key3", "value3", 0)
		
		assert.Equal(t, 3, cache.Count())
		
		// 清空
		cache.Clear()
		
		assert.Equal(t, 0, cache.Count())
		assert.False(t, cache.Exists("key1"))
		assert.False(t, cache.Exists("key2"))
		assert.False(t, cache.Exists("key3"))
	})
}

func TestCacheKeys(t *testing.T) {
	cache := NewCacheManager[int]("test", "测试缓存")
	
	t.Run("获取所有键", func(t *testing.T) {
		cache.Set("a", 1, 0)
		cache.Set("b", 2, 0)
		cache.Set("c", 3, time.Millisecond) // 即将过期
		
		keys := cache.Keys()
		assert.Len(t, keys, 3)
		assert.Contains(t, keys, "a")
		assert.Contains(t, keys, "b")
		assert.Contains(t, keys, "c")
		
		// 等待c过期
		time.Sleep(time.Millisecond * 2)
		
		keys = cache.Keys()
		assert.Len(t, keys, 2)
		assert.Contains(t, keys, "a")
		assert.Contains(t, keys, "b")
		assert.NotContains(t, keys, "c")
	})
}

func TestCacheGetOrSet(t *testing.T) {
	cache := NewCacheManager[string]("test", "测试缓存")
	
	t.Run("获取不存在的值时设置默认值", func(t *testing.T) {
		value := cache.GetOrSet("new_key", "default_value", time.Minute)
		assert.Equal(t, "default_value", value)
		
		// 再次获取应该返回已存在的值
		value = cache.GetOrSet("new_key", "another_value", time.Minute)
		assert.Equal(t, "default_value", value) // 不应该是another_value
	})
	
	t.Run("获取已存在的值", func(t *testing.T) {
		cache.Set("existing", "existing_value", 0)
		
		value := cache.GetOrSet("existing", "default_value", time.Minute)
		assert.Equal(t, "existing_value", value)
	})
}

func TestCacheExpiration(t *testing.T) {
	cache := NewCacheManager[string]("test", "测试缓存")
	
	t.Run("过期缓存自动清理", func(t *testing.T) {
		// 设置短过期时间
		cache.Set("short_lived", "value", time.Millisecond*10)
		
		// 立即检查存在
		assert.True(t, cache.Exists("short_lived"))
		
		// 等待过期
		time.Sleep(time.Millisecond * 20)
		
		// 检查已过期
		assert.False(t, cache.Exists("short_lived"))
		_, found := cache.Get("short_lived")
		assert.False(t, found)
	})
	
	t.Run("零过期时间表示永不过期", func(t *testing.T) {
		cache.Set("permanent", "value", 0)
		
		// 等待一段时间
		time.Sleep(time.Millisecond * 10)
		
		// 应该仍然存在
		assert.True(t, cache.Exists("permanent"))
		value, found := cache.Get("permanent")
		assert.True(t, found)
		assert.Equal(t, "value", value)
	})
}

func TestCacheGetWithTTL(t *testing.T) {
	cache := NewCacheManager[string]("test", "测试缓存")
	
	t.Run("获取带TTL的缓存", func(t *testing.T) {
		cache.Set("ttl_test", "value", time.Hour)
		
		value, ttl, found := cache.GetWithTTL("ttl_test")
		assert.True(t, found)
		assert.Equal(t, "value", value)
		assert.True(t, ttl > 0)
		assert.True(t, ttl <= time.Hour)
	})
	
	t.Run("获取永不过期缓存的TTL", func(t *testing.T) {
		cache.Set("permanent", "value", 0)
		
		value, ttl, found := cache.GetWithTTL("permanent")
		assert.True(t, found)
		assert.Equal(t, "value", value)
		assert.Equal(t, time.Duration(-1), ttl) // -1表示永不过期
	})
	
	t.Run("获取不存在缓存的TTL", func(t *testing.T) {
		value, ttl, found := cache.GetWithTTL("nonexistent")
		assert.False(t, found)
		assert.Equal(t, "", value)
		assert.Equal(t, time.Duration(-1), ttl)
	})
}

func TestCacheCleanupExpired(t *testing.T) {
	cache := NewCacheManager[int]("test", "测试缓存")
	
	t.Run("清理过期缓存", func(t *testing.T) {
		// 添加一些缓存，部分设置为很短的过期时间
		cache.Set("keep1", 1, 0)                    // 永不过期
		cache.Set("keep2", 2, time.Hour)            // 长期缓存
		cache.Set("expire1", 3, time.Millisecond)   // 短期缓存
		cache.Set("expire2", 4, time.Millisecond*2) // 短期缓存
		
		// 等待短期缓存过期
		time.Sleep(time.Millisecond * 5)
		
		// 手动清理过期缓存
		cleaned := cache.CleanupExpired()
		
		assert.Equal(t, 2, cleaned) // 应该清理了2个过期缓存
		assert.Equal(t, 2, cache.Count())
		assert.True(t, cache.Exists("keep1"))
		assert.True(t, cache.Exists("keep2"))
		assert.False(t, cache.Exists("expire1"))
		assert.False(t, cache.Exists("expire2"))
	})
}

func TestCacheStatistics(t *testing.T) {
	cache := NewCacheManager[string]("statistics", "统计测试")
	
	t.Run("获取缓存统计信息", func(t *testing.T) {
		// 添加一些数据
		cache.Set("stat1", "value1", 0)
		cache.Set("stat2", "value2", time.Hour)
		cache.Set("stat3", "value3", time.Millisecond) // 即将过期
		
		stats := cache.GetStats()
		
		assert.Equal(t, "statistics", stats.Name)
		assert.Equal(t, "统计测试", stats.Description)
		assert.Equal(t, 3, stats.ItemCount)
		
		// 等待一个过期
		time.Sleep(time.Millisecond * 2)
		
		stats = cache.GetStats()
		assert.Equal(t, 2, stats.ItemCount) // 过期的应该在获取统计时被清理
	})
}

// 并发测试
func TestCacheConcurrency(t *testing.T) {
	cache := NewCacheManager[int]("concurrent", "并发测试")
	
	t.Run("并发读写测试", func(t *testing.T) {
		const numGoroutines = 100
		const numOperations = 100
		
		// 启动多个goroutine并发操作缓存
		done := make(chan bool, numGoroutines)
		
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer func() { done <- true }()
				
				for j := 0; j < numOperations; j++ {
					key := fmt.Sprintf("key_%d_%d", id, j)
					value := id*1000 + j
					
					// 写入
					cache.Set(key, value, 0)
					
					// 读取
					if readValue, found := cache.Get(key); found {
						assert.Equal(t, value, readValue)
					}
					
					// 删除一些
					if j%10 == 0 {
						cache.Delete(key)
					}
				}
			}(i)
		}
		
		// 等待所有goroutine完成
		for i := 0; i < numGoroutines; i++ {
			<-done
		}
		
		// 缓存应该仍然能正常工作
		cache.Set("final_test", 999, 0)
		value, found := cache.Get("final_test")
		assert.True(t, found)
		assert.Equal(t, 999, value)
	})
}

// 性能测试
func BenchmarkCacheSet(b *testing.B) {
	cache := NewCacheManager[string]("benchmark", "性能测试")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key_%d", i)
		cache.Set(key, "value", 0)
	}
}

func BenchmarkCacheGet(b *testing.B) {
	cache := NewCacheManager[string]("benchmark", "性能测试")
	
	// 预填充数据
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key_%d", i)
		cache.Set(key, "value", 0)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key_%d", i%1000)
		cache.Get(key)
	}
}

func BenchmarkCacheConcurrentReadWrite(b *testing.B) {
	cache := NewCacheManager[int]("benchmark", "并发性能测试")
	
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key_%d", i%100)
			if i%2 == 0 {
				cache.Set(key, i, 0)
			} else {
				cache.Get(key)
			}
			i++
		}
	})
}

// 测试不同类型的缓存
func TestCacheWithDifferentTypes(t *testing.T) {
	t.Run("结构体类型缓存", func(t *testing.T) {
		type User struct {
			ID   int
			Name string
		}
		
		cache := NewCacheManager[User]("users", "用户缓存")
		
		user := User{ID: 1, Name: "张三"}
		cache.Set("user_1", user, 0)
		
		retrieved, found := cache.Get("user_1")
		assert.True(t, found)
		assert.Equal(t, user.ID, retrieved.ID)
		assert.Equal(t, user.Name, retrieved.Name)
	})
	
	t.Run("切片类型缓存", func(t *testing.T) {
		cache := NewCacheManager[[]string]("lists", "列表缓存")
		
		list := []string{"a", "b", "c"}
		cache.Set("list_1", list, 0)
		
		retrieved, found := cache.Get("list_1")
		assert.True(t, found)
		assert.Equal(t, list, retrieved)
	})
	
	t.Run("映射类型缓存", func(t *testing.T) {
		cache := NewCacheManager[map[string]int]("maps", "映射缓存")
		
		data := map[string]int{"a": 1, "b": 2}
		cache.Set("map_1", data, 0)
		
		retrieved, found := cache.Get("map_1")
		assert.True(t, found)
		assert.Equal(t, data, retrieved)
	})
}