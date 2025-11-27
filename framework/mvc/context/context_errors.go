package context

// ============= 错误处理方法（细粒度锁优化） =============

// AddError 添加错误 - 使用专用锁优化
func (ctx *Context) AddError(err error) error {
	if err != nil {
		ctx.errMu.Lock()
		ctx.errors = append(ctx.errors, err)
		ctx.errMu.Unlock()
		return err
	}
	return nil
}

// GetErrors 获取所有错误
func (ctx *Context) GetErrors() []error {
	ctx.errMu.Lock()
	errors := make([]error, len(ctx.errors))
	copy(errors, ctx.errors)
	ctx.errMu.Unlock()
	return errors
}

// HasErrors 是否有错误
func (ctx *Context) HasErrors() bool {
	ctx.errMu.Lock()
	hasErr := len(ctx.errors) > 0
	ctx.errMu.Unlock()
	return hasErr
}

// ClearErrors 清除所有错误
func (ctx *Context) ClearErrors() {
	ctx.errMu.Lock()
	ctx.errors = ctx.errors[:0]
	ctx.errMu.Unlock()
}

// LastError 获取最后一个错误
func (ctx *Context) LastError() error {
	ctx.errMu.Lock()
	defer ctx.errMu.Unlock()

	if len(ctx.errors) == 0 {
		return nil
	}
	return ctx.errors[len(ctx.errors)-1]
}