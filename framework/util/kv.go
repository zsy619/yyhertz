package util

// Package util 提供若干小型通用工具类型，当前文件包含通用的键值接口及其简单实现。

// KV 表示一个通用的键值对接口。
// 实现者需提供 GetKey 和 GetValue，两者返回类型为 any。
type KV interface {
	GetKey() any
	GetValue() any
}

// SimpleKV 是 KV 的简单实现，直接持有 Key 与 Value 字段。
type SimpleKV struct {
	Key   any
	Value any
}

var _ KV = new(SimpleKV)

func (s *SimpleKV) GetKey() any   { return s.Key }
func (s *SimpleKV) GetValue() any { return s.Value }

// KVs 表示键到值的查找容器，提供带默认值的读取、存在性检查及条件回调。
type KVs interface {
	// GetValueOr 返回 key 对应的值；若不存在则返回 defValue。
	GetValueOr(key interface{}, defValue any) any

	// Contains 返回 key 是否存在于容器中。
	Contains(key interface{}) bool

	// IfContains 若 key 存在则执行 action，并返回自身以支持链式调用。
	IfContains(key interface{}, action func(value any)) KVs
}

// SimpleKVs 使用内置 map 作为后端存储的 KVs 实现。
type SimpleKVs struct {
	kvs map[interface{}]any
}

var _ KVs = new(SimpleKVs)

// GetValueOr 返回 key 的值，若不存在则返回 defValue。
func (kvs *SimpleKVs) GetValueOr(key any, defValue any) any {
	if v, ok := kvs.kvs[key]; ok {
		return v
	}
	return defValue
}

// Contains 检查 key 是否存在。
func (kvs *SimpleKVs) Contains(key any) bool {
	_, ok := kvs.kvs[key]
	return ok
}

// IfContains 当 key 存在时执行 action，并返回自身以便链式调用。
func (kvs *SimpleKVs) IfContains(key any, action func(value any)) KVs {
	if v, ok := kvs.kvs[key]; ok {
		action(v)
	}
	return kvs
}

// NewKVs 根据给定的 KV 列表构造并返回一个 KVs 实例。
func NewKVs(kvs ...KV) KVs {
	res := &SimpleKVs{
		kvs: make(map[any]any, len(kvs)),
	}
	for _, kv := range kvs {
		res.kvs[kv.GetKey()] = kv.GetValue()
	}
	return res
}
