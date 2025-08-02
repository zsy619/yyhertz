package core

// ============= 数据操作方法 =============

// SetData 设置模板数据（防御性编程 + 自动初始化）
func (c *BaseController) SetData(key string, value any) {
	// 确保Data已初始化
	if c.Data == nil {
		c.Data = make(map[string]any)
	}

	// 如果控制器信息未初始化，则自动初始化
	c.ensureInitialized()

	c.Data[key] = value
}

// GetData 获取模板数据（Beego兼容）
func (c *BaseController) GetData(key string) any {
	if c.Data == nil {
		return nil
	}
	return c.Data[key]
}

// DelData 删除模板数据
func (c *BaseController) DelData(key string) {
	if c.Data != nil {
		delete(c.Data, key)
	}
}