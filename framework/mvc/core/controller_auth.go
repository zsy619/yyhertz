package core

// IsAdminLogin 检查当前会话是否为管理员登录状态。
// 返回值：
//   - int64: 如果会话中存在有效的管理员ID，则返回该ID；否则返回0。
//
// 说明：
//   - 该方法通过检查会话中的`adminId`字段来判断管理员是否已登录。
//   - 如果`adminId`不存在或类型不为int64，则返回0。
func (c *BaseController) IsAdminLogin(adminId ...string) int64 {
	idKey := AdminIdKey
	if len(adminId) > 0 {
		idKey = adminId[0]
	}
	id := c.GetSession(idKey)
	if id == nil {
		return 0
	} else {
		switch id := id.(type) {
		case int64:
			return id
		default:
			return 0
		}
	}
}

// SetAdminId 设置管理员ID到会话中。
// 参数：
//   - id: 需要设置的管理员ID，类型为int64。
//   - adminId: 可选参数，用于指定会话中存储管理员ID的键名。如果未提供，默认使用"adminId"作为键名。
//
// 说明：
//   - 该方法会将管理员ID存储到当前会话中，键名由adminId参数指定或使用默认值。
//   - 适用于需要动态设置管理员ID的场景。
func (c *BaseController) SetAdminId(id int64, adminId ...string) {
	idKey := AdminIdKey
	if len(adminId) > 0 {
		idKey = adminId[0]
	}
	c.SetSession(idKey, id)
}
