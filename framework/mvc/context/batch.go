package context

// BatchContexts 批量Context处理器
type BatchContexts struct {
	contexts []*Context
	size     int
}

// NewBatchContexts 创建批量处理器
func NewBatchContexts(size int) *BatchContexts {
	return &BatchContexts{
		contexts: make([]*Context, size),
		size:     0,
	}
}

// Add 添加Context到批处理
func (batch *BatchContexts) Add(ctx *Context) {
	if batch.size < len(batch.contexts) {
		batch.contexts[batch.size] = ctx
		batch.size++
	}
}

// Release 批量释放Context
func (batch *BatchContexts) Release() {
	for i := 0; i < batch.size; i++ {
		if ctx := batch.contexts[i]; ctx != nil {
			ctx.Release()
			batch.contexts[i] = nil
		}
	}
	batch.size = 0
}

// ForEach 遍历所有Context
func (batch *BatchContexts) ForEach(fn func(*Context)) {
	for i := 0; i < batch.size; i++ {
		if ctx := batch.contexts[i]; ctx != nil {
			fn(ctx)
		}
	}
}