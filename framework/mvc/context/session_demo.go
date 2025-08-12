package context

import (
	"fmt"
)

// SessionCompatExample 演示InputData的标准框架session兼容功能
func SessionCompatExample() {
	fmt.Println("🔧 YYHertz InputData Session 标准框架兼容性演示")
	fmt.Println("==========================================")
	
	// 模拟创建一个Context（在实际应用中由框架创建）
	ctx := &Context{
		Keys: make(map[string]interface{}),
	}
	
	// 初始化InputData
	ctx.Input = &InputData{ctx: ctx}
	
	fmt.Println("\n📋 功能1: 启动Session")
	fmt.Println("--------------------------------------")
	
	// 启动session - 返回标准框架兼容的session store
	sessionStore := ctx.Input.StartSession()
	if sessionStore != nil {
		fmt.Printf("✅ Session已启动，ID: %s\n", ctx.Input.GetSessionID())
	}
	
	fmt.Println("\n📋 功能2: 标准MVC风格的Session操作")
	fmt.Println("--------------------------------------")
	
	// 使用标准MVC风格的InputData方法
	ctx.Input.SetSession("username", "张三")
	ctx.Input.SetSession("user_id", 12345)
	ctx.Input.SetSession("role", "admin")
	
	fmt.Printf("✅ 设置session数据: username=%v, user_id=%v, role=%v\n", 
		ctx.Input.GetSession("username"),
		ctx.Input.GetSession("user_id"),
		ctx.Input.GetSession("role"))
	
	fmt.Println("\n📋 功能3: 使用标准框架Store接口")
	fmt.Println("--------------------------------------")
	
	// 使用YYHertz Context替代标准库context
	// 这是新的增强功能！现在可以直接使用我们的Context！
	
	// 标准框架 Store.Set/Get方法 (现在使用YYHertz Context)
	sessionStore.Set(ctx, "compat_key", "compat_value")
	compatValue := sessionStore.Get(ctx, "compat_key")
	fmt.Printf("✅ 标准框架Store接口: compat_key=%v\n", compatValue)
	
	// 支持[]byte类型key (标准框架特性)
	sessionStore.Set(ctx, []byte("byte_key"), "byte_value")
	byteValue := sessionStore.Get(ctx, []byte("byte_key"))
	fmt.Printf("✅ []byte key支持: byte_key=%v\n", byteValue)
	
	// 获取session ID
	sessionID := sessionStore.SessionID(ctx)
	fmt.Printf("✅ Session ID: %s\n", sessionID)
	
	fmt.Println("\n📋 功能4: Session生命周期管理")
	fmt.Println("--------------------------------------")
	
	// 检查session状态
	fmt.Printf("✅ Session是否已启动: %t\n", ctx.Input.IsSessionStarted())
	
	// 保存session (标准MVC兼容)
	ctx.Input.SaveSession()
	fmt.Println("✅ Session已保存")
	
	// 删除特定键
	ctx.Input.DelSession("role")
	fmt.Printf("✅ 删除role后: role=%v (应为nil)\n", ctx.Input.GetSession("role"))
	
	// 清空session
	ctx.Input.ClearSession()
	fmt.Printf("✅ 清空session后: username=%v (应为nil)\n", ctx.Input.GetSession("username"))
	
	fmt.Println("\n📋 功能5: 高级Session操作")  
	fmt.Println("--------------------------------------")
	
	// 重新设置数据用于演示
	ctx.Input.SetSession("demo_key1", "demo_value1")
	ctx.Input.SetSession("demo_key2", "demo_value2")
	
	// 使用标准框架 Store的Flush方法清空所有数据
	sessionStore.Flush(ctx)
	fmt.Printf("✅ Flush后: demo_key1=%v (应为nil)\n", sessionStore.Get(ctx, "demo_key1"))
	
	// SessionRelease (标准框架接口) - 现在有增强功能！
	sessionStore.Set(ctx, "release_test", "release_value")
	sessionStore.SessionRelease(ctx, nil) // 释放并保存，会设置增强标记
	fmt.Printf("✅ SessionRelease后数据保持: release_test=%v\n", sessionStore.Get(ctx, "release_test"))
	fmt.Printf("✅ SessionRelease增强标记: session_released=%v\n", ctx.Keys["session_released"])
	
	// 销毁session
	ctx.Input.DestroySession()
	fmt.Printf("✅ 销毁session后状态: IsStarted=%t\n", ctx.Input.IsSessionStarted())
	
	fmt.Println("\n🎯 兼容性总结")
	fmt.Println("==========================================")
	fmt.Println("✅ 支持所有标准MVC InputData session方法:")
	fmt.Println("   - StartSession() *SessionStore")
	fmt.Println("   - SetSession(key, value) error")
	fmt.Println("   - GetSession(key) interface{}")
	fmt.Println("   - DelSession(key) error")
	fmt.Println("   - GetSessionID() string")
	fmt.Println("   - IsSessionStarted() bool")
	fmt.Println("   - SaveSession() error")
	fmt.Println("   - ClearSession()")
	fmt.Println("   - DestroySession()")
	fmt.Println("")
	fmt.Println("✅ 支持所有标准框架 Store接口方法:")
	fmt.Println("   - Set(ctx, key, value) error")
	fmt.Println("   - Get(ctx, key) interface{}")
	fmt.Println("   - Delete(ctx, key) error")
	fmt.Println("   - SessionID(ctx) string")
	fmt.Println("   - SessionRelease(ctx, w)")
	fmt.Println("   - SessionReleaseIfPresent(ctx, w)")
	fmt.Println("   - Flush(ctx) error")
	fmt.Println("")
	fmt.Println("🚀 完美的标准框架到YYHertz迁移兼容性!")
}

// LegacyMigrationExample 展示从标准框架迁移到YYHertz的代码对比
func LegacyMigrationExample() {
	fmt.Println("\n📖 标准框架迁移代码对比")
	fmt.Println("==========================================")
	
	fmt.Println("# 标准MVC框架原始代码:")
	fmt.Println(`
// 标准MVC Controller
func (c *MainController) Login() {
    // 启动session
    sess := c.StartSession()
    defer sess.SessionRelease(c.Ctx.ResponseWriter)
    
    // 设置session数据
    sess.Set("username", "admin") 
    sess.Set("user_id", 123)
    
    // 获取session数据
    username := sess.Get("username")
    
    // 使用InputData
    userInput := c.Ctx.Input
    userInput.SetSession("role", "admin")
    role := userInput.GetSession("role")
}`)
	
	fmt.Println("\n# YYHertz原生代码 (更简洁高效!):")
	fmt.Println(`
// YYHertz Controller 
func (c *BaseController) Login() {
    // 启动session - 使用YYHertz原生Context!
    sess := c.Ctx.Input.StartSession()
    defer sess.SessionRelease(c.Ctx, c.Ctx.Response)
    
    // 设置session数据 - 直接使用YYHertz Context!
    sess.Set(c.Ctx, "username", "admin")
    sess.Set(c.Ctx, "user_id", 123)
    
    // 获取session数据 - 性能更优!
    username := sess.Get(c.Ctx, "username")
    
    // 使用InputData - 完全兼容且功能增强!
    userInput := c.Ctx.Input
    userInput.SetSession("role", "admin")
    role := userInput.GetSession("role")
}`)
	
	fmt.Println("\n✨ 迁移优势:")
	fmt.Println("✅ 零破坏性迁移 - 保持API兼容性")
	fmt.Println("✅ 性能大幅提升 - 基于YYHertz高性能架构")
	fmt.Println("✅ 功能全面增强 - 支持更多存储后端和特性")
	fmt.Println("✅ 类型系统优化 - 更好的Go类型安全支持")
	fmt.Println("✅ 现代化设计 - 统一的Context架构模式")
	fmt.Println("✅ 原生集成 - 深度融合YYHertz生态系统")
}