// Package core 提供 App 的 WebSocket 路由注册功能
//
// 本文件实现了 App 的 WebSocket 实例方法，支持：
// 1. HandlerWs - 直接注册 WebSocket 处理函数
// 2. RouterWs - 注册控制器的 WebSocket 方法
// 3. RouterWsWithUpgrader - 使用自定义升级器注册 WebSocket 路由
package core

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hertz-contrib/websocket"

	mvcContext "github.com/zsy619/yyhertz/framework/mvc/context"
	"github.com/zsy619/yyhertz/framework/mvc/define"
)

// ============= WebSocket 路由实例方法 =============

// HandlerWs 注册 WebSocket 处理器到应用
//
// 该方法将一个 WebSocket 处理函数注册为应用路由。
// 内部会创建一个 WebSocket 控制器来处理协议升级和连接管理。
//
// 参数：
//   - path: string - WebSocket 路径，如 "/ws/echo"
//   - handler: func(*define.WsConn) - WebSocket 连接处理函数
//
// 返回值：
//   - *App: 应用实例，支持链式调用
//
// 实现原理：
//  1. 创建默认的 WebSocket 升级器
//  2. 使用升级器和处理函数创建 WebSocket 控制器
//  3. 将控制器的 HandleWebSocket 方法注册为 GET 路由
//
// 使用示例：
//
//	app := mvc.NewApp()
//	app.HandlerWs("/ws/echo", func(conn *define.WsConn) {
//		for {
//			messageType, message, err := conn.ReadMessage()
//			if err != nil {
//				break
//			}
//			conn.WriteMessage(messageType, message)
//		}
//	})
func (app *App) HandlerWs(path string, handler func(*define.WsConn)) *App {
	// 创建默认的 WebSocket 升级器
	upgrader := websocket.HertzUpgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	// 创建 WebSocket 控制器
	wsController := createWebSocketController(handler, upgrader).(*webSocketControllerWrapper)

	// 创建处理器函数
	handlerFunc := define.HandlerFunc(func(ctx context.Context, c *define.RequestContext) {
		// 调用 WebSocket 处理方法
		wsController.HandleWebSocket(c)
	})

	// 注册 GET 路由（WebSocket 升级使用 GET 请求）
	app.registerRoute("GET", path, handlerFunc)

	return app
}

// RouterWs 注册控制器的 WebSocket 方法到应用
//
// 该方法通过反射调用控制器的指定方法来处理 WebSocket 连接。
// 控制器方法必须具有特定的签名：func(conn *define.WsConn)
// 现在使用新的包装器来确保控制器被正确初始化。
//
// 参数：
//   - path: string - WebSocket 路径
//   - ctrl: IController - 控制器实例
//   - method: string - 控制器方法名
//
// 返回值：
//   - *App: 应用实例，支持链式调用
//
// 方法签名要求：
//
//	控制器方法必须具有以下签名之一：
//	- func (c *YourController) MethodName(conn *define.WsConn)
//	- func (c *YourController) MethodName(conn *define.WsConn) error
//
// 使用示例：
//
//	type ChatController struct {
//		core.BaseController
//	}
//
//	func (c *ChatController) HandleChat(conn *define.WsConn) {
//		// WebSocket 处理逻辑
//	}
//
//	app.RouterWs("/ws/chat", &ChatController{}, "HandleChat")
func (app *App) RouterWs(path string, ctrl IController, method string) *App {
	// 验证控制器和方法
	ctrlValue := reflect.ValueOf(ctrl)
	ctrlType := reflect.TypeOf(ctrl)

	// 检查控制器是否为指针类型
	if ctrlType.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("RouterWs: 控制器必须是指针类型"))
	}

	// 获取指定的方法
	methodValue := ctrlValue.MethodByName(method)
	if !methodValue.IsValid() {
		panic(fmt.Sprintf("RouterWs: 控制器方法 '%s' 不存在", method))
	}

	// 验证方法签名
	if err := validateWebSocketMethodSignature(methodValue.Type(), method); err != nil {
		panic(fmt.Sprintf("RouterWs: %v", err))
	}

	// 创建专门的 WebSocket 控制器方法包装器
	wsController := &webSocketControllerMethodWrapper{
		targetController: ctrl,
		methodName:       method,
		method:           methodValue,
		upgrader:         websocket.HertzUpgrader{ReadBufferSize: 1024, WriteBufferSize: 1024},
	}

	// 创建处理器函数
	handlerFunc := define.HandlerFunc(func(ctx context.Context, c *define.RequestContext) {
		// 调用 WebSocket 处理方法
		wsController.HandleWebSocket(c)
	})

	// 注册 GET 路由（WebSocket 升级使用 GET 请求）
	app.registerRoute("GET", path, handlerFunc)

	return app
}

// RouterWsWithUpgrader 注册带自定义升级器的 WebSocket 控制器路由
//
// 该方法允许指定自定义的 WebSocket 升级器配置，
// 适用于需要特殊配置的 WebSocket 连接。
// 现在使用新的包装器来确保控制器被正确初始化。
//
// 参数：
//   - path: string - WebSocket 路径
//   - ctrl: IController - 控制器实例
//   - method: string - 控制器方法名
//   - upgrader: websocket.HertzUpgrader - 自定义升级器配置
//
// 返回值：
//   - *App: 应用实例，支持链式调用
//
// 升级器配置示例：
//
//	upgrader := websocket.HertzUpgrader{
//		ReadBufferSize:   2048,
//		WriteBufferSize:  2048,
//		HandshakeTimeout: 10 * time.Second,
//		CheckOrigin: func(ctx *app.RequestContext) bool {
//			// 自定义来源检查
//			return true
//		},
//	}
//
//	app.RouterWsWithUpgrader("/ws/game", &GameController{}, "HandleGame", upgrader)
func (app *App) RouterWsWithUpgrader(path string, ctrl IController, method string, upgrader define.WsHertzUpgrader) *App {
	// 验证控制器和方法
	ctrlValue := reflect.ValueOf(ctrl)
	ctrlType := reflect.TypeOf(ctrl)

	// 检查控制器是否为指针类型
	if ctrlType.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("RouterWsWithUpgrader: 控制器必须是指针类型"))
	}

	// 获取指定的方法
	methodValue := ctrlValue.MethodByName(method)
	if !methodValue.IsValid() {
		panic(fmt.Sprintf("RouterWsWithUpgrader: 控制器方法 '%s' 不存在", method))
	}

	// 验证方法签名
	if err := validateWebSocketMethodSignature(methodValue.Type(), method); err != nil {
		panic(fmt.Sprintf("RouterWsWithUpgrader: %v", err))
	}

	// 创建专门的 WebSocket 控制器方法包装器
	wsController := &webSocketControllerMethodWrapper{
		targetController: ctrl,
		methodName:       method,
		method:           methodValue,
		upgrader:         upgrader,
	}

	// 创建处理器函数
	handlerFunc := define.HandlerFunc(func(ctx context.Context, c *define.RequestContext) {
		// 调用 WebSocket 处理方法
		wsController.HandleWebSocket(c)
	})

	// 注册 GET 路由
	app.registerRoute("GET", path, handlerFunc)

	return app
}

// ============= 辅助方法 =============

// createWebSocketHandler 创建控制器方法的 WebSocket 处理器
//
// 该方法通过反射获取控制器方法，并创建相应的 WebSocket 处理器函数。
// 支持方法签名验证和错误处理。现在使用新的包装器来确保控制器被正确初始化。
//
// 参数：
//   - ctrl: IController - 控制器实例
//   - methodName: string - 方法名
//
// 返回值：
//   - func(*define.WsConn): WebSocket 处理器函数
//   - error: 错误信息
func (app *App) createWebSocketHandler(ctrl IController, methodName string) (func(*define.WsConn), error) {
	// 获取控制器的反射值和类型
	ctrlValue := reflect.ValueOf(ctrl)
	ctrlType := reflect.TypeOf(ctrl)

	// 检查控制器是否为指针类型
	if ctrlType.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("控制器必须是指针类型")
	}

	// 获取指定的方法
	method := ctrlValue.MethodByName(methodName)
	if !method.IsValid() {
		return nil, fmt.Errorf("控制器方法 '%s' 不存在", methodName)
	}

	// 获取方法类型
	methodType := method.Type()

	// 验证方法签名
	if err := validateWebSocketMethodSignature(methodType, methodName); err != nil {
		return nil, err
	}

	// 创建 WebSocket 处理器函数
	// 注意：这个函数是一个占位符，实际的处理逻辑在 webSocketControllerMethodWrapper 中
	handler := func(conn *define.WsConn) {
		// 这个函数不应该被直接调用，因为我们使用了新的包装器
		fmt.Printf("警告：createWebSocketHandler 创建的处理器被直接调用，这可能导致控制器未初始化\n")
		
		// 准备方法调用参数
		args := []reflect.Value{
			reflect.ValueOf(conn),
		}

		// 调用控制器方法
		results := method.Call(args)

		// 处理方法返回值（如果有错误返回值）
		if len(results) > 0 {
			if err, ok := results[0].Interface().(error); ok && err != nil {
				// 记录错误（这里可以添加日志记录）
				fmt.Printf("WebSocket 处理器执行错误: %v\n", err)
			}
		}
	}

	return handler, nil
}

// validateWebSocketMethodSignature 验证 WebSocket 方法签名
//
// 检查方法是否符合 WebSocket 处理器的要求。
//
// 支持的签名：
//   - func(conn *define.WsConn)
//   - func(conn *define.WsConn) error
//
// 参数：
//   - methodType: reflect.Type - 方法类型
//   - methodName: string - 方法名（用于错误消息）
//
// 返回值：
//   - error: 验证错误，nil 表示验证通过
func validateWebSocketMethodSignature(methodType reflect.Type, methodName string) error {
	// 检查参数数量
	if methodType.NumIn() != 1 {
		return fmt.Errorf("方法 '%s' 参数数量错误，期望 1 个参数，实际 %d 个", methodName, methodType.NumIn())
	}

	// 检查第一个参数类型
	paramType := methodType.In(0)
	expectedParamType := reflect.TypeOf((*define.WsConn)(nil))
	if paramType != expectedParamType {
		return fmt.Errorf("方法 '%s' 第一个参数类型错误，期望 *define.WsConn，实际 %s", methodName, paramType)
	}

	// 检查返回值数量（可以是 0 或 1）
	numOut := methodType.NumOut()
	if numOut > 1 {
		return fmt.Errorf("方法 '%s' 返回值数量错误，期望 0 或 1 个返回值，实际 %d 个", methodName, numOut)
	}

	// 如果有返回值，检查返回值类型
	if numOut == 1 {
		returnType := methodType.Out(0)
		expectedReturnType := reflect.TypeOf((*error)(nil)).Elem()
		if returnType != expectedReturnType {
			return fmt.Errorf("方法 '%s' 返回值类型错误，期望 error，实际 %s", methodName, returnType)
		}
	}

	return nil
}

// createWebSocketController 创建 WebSocket 控制器的辅助函数
//
// 该函数创建一个 WebSocket 控制器实例，用于处理 WebSocket 连接。
//
// 参数：
//   - handler: func(*define.WsConn) - WebSocket 处理函数
//   - upgrader: websocket.HertzUpgrader - WebSocket 升级器
//
// 返回值：
//   - IController: WebSocket 控制器实例
func createWebSocketController(handler func(*define.WsConn), upgrader websocket.HertzUpgrader) IController {
	return &webSocketControllerWrapper{
		handler:  handler,
		upgrader: upgrader,
	}
}

// webSocketControllerWrapper WebSocket 控制器包装器
//
// 该控制器包装 WebSocket 处理器，提供标准的控制器接口。
type webSocketControllerWrapper struct {
	BaseController
	handler  func(*define.WsConn)
	upgrader websocket.HertzUpgrader
}

// webSocketControllerMethodWrapper WebSocket 控制器方法包装器
//
// 专门用于处理控制器方法的 WebSocket 连接，确保控制器被正确初始化。
type webSocketControllerMethodWrapper struct {
	BaseController
	targetController IController // 目标控制器实例
	methodName       string      // 方法名
	method           reflect.Value // 反射方法
	upgrader         websocket.HertzUpgrader
}

// HandleWebSocket 处理 WebSocket 连接
func (w *webSocketControllerWrapper) HandleWebSocket(ctx *define.RequestContext) {
	// 创建MVC上下文包装
	mvcCtx := mvcContext.NewContext(ctx)
	// 设置请求上下文
	w.Ctx = mvcCtx

	// 检查是否为 WebSocket 请求
	if !w.Ctx.IsWebsocket() {
		w.Ctx.String(400, "Bad Request: Not a WebSocket request")
		return
	}

	// 升级连接
	err := w.upgrader.Upgrade(w.Ctx.Request(), func(conn *websocket.Conn) {
		// 直接调用处理器，define.WsConn 是 websocket.Conn 的别名
		w.handler((*define.WsConn)(conn))
	})

	if err != nil {
		fmt.Printf("WebSocket 升级失败: %v\n", err)
		w.Ctx.String(500, "Internal Server Error: WebSocket upgrade failed")
	}
}

// HandleWebSocket 处理控制器方法的 WebSocket 连接
func (w *webSocketControllerMethodWrapper) HandleWebSocket(ctx *define.RequestContext) {
	// 创建MVC上下文包装
	mvcCtx := mvcContext.NewContext(ctx)
	// 设置请求上下文
	w.Ctx = mvcCtx

	// 检查是否为 WebSocket 请求
	if !w.Ctx.IsWebsocket() {
		w.Ctx.String(400, "Bad Request: Not a WebSocket request")
		return
	}

	// 升级连接
	err := w.upgrader.Upgrade(w.Ctx.Request(), func(conn *websocket.Conn) {
		// 添加 recover 机制来捕获 ErrAbort panic（兼容beego）
		defer func() {
			if r := recover(); r != nil {
				if r == ErrAbort {
					// 捕获到 ErrAbort，正常停止执行，不输出任何错误
					// 这是用户主动调用 StopRun() 或 ServeJSON() 的预期行为
					return
				}
				// 其他类型的 panic，重新抛出
				panic(r)
			}
		}()
		
		// 重要：在这里初始化目标控制器的 Context
		// 获取控制器名称和方法名称用于初始化
		controllerName := reflect.TypeOf(w.targetController).Elem().Name()
		
		// 调用目标控制器的初始化方法
		w.targetController.Init(mvcCtx, controllerName, w.methodName, w.targetController)
		
		// 准备方法调用参数
		args := []reflect.Value{
			reflect.ValueOf((*define.WsConn)(conn)),
		}

		// 调用控制器方法
		results := w.method.Call(args)

		// 处理方法返回值（如果有错误返回值）
		if len(results) > 0 {
			if err, ok := results[0].Interface().(error); ok && err != nil {
				fmt.Printf("WebSocket 处理器执行错误: %v\n", err)
			}
		}
	})

	if err != nil {
		fmt.Printf("WebSocket 升级失败: %v\n", err)
		w.Ctx.String(500, "Internal Server Error: WebSocket upgrade failed")
	}
}
