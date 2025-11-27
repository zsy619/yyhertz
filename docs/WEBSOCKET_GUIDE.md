# YYHertz WebSocket 使用指南

YYHertz 框架现已完全支持 WebSocket 功能，提供了灵活且强大的实时通信能力。支持命名空间路由、嵌套命名空间、控制器方法集成等高级功能。

> **最新更新**: 已修复嵌套命名空间 WebSocket 路由问题，现在支持任意深度的命名空间嵌套。

## 🚀 快速开始

### 1. 函数式 WebSocket 路由

使用命名空间注册 WebSocket 路由：

```go
package main

import (
    "github.com/hertz-contrib/websocket"
    "github.com/zsy619/yyhertz/framework/mvc"
)

func main() {
    // 获取应用实例
    app := mvc.HertzApp
    
    // 创建 WebSocket 命名空间
    ns := mvc.NewNamespace("/api",
        mvc.NSWebSocket("/echo", func(conn *websocket.Conn) {
            for {
                messageType, message, err := conn.ReadMessage()
                if err != nil {
                    break
                }
                // Echo 消息
                conn.WriteMessage(messageType, message)
            }
        }),
    )
    
    // 注册命名空间
    mvc.AddNamespace(ns)
    
    app.Run()
}
```

### 2. 控制器中使用 WebSocket

```go
import (
    "fmt"
    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/framework/mvc/core"
    "github.com/zsy619/yyhertz/framework/mvc/define"
)

type ChatController struct {
    core.BaseController
}

// Index WebSocket 处理方法
// 注意：WebSocket 控制器方法必须使用 *define.WsConn 类型参数
func (c *ChatController) Index(conn *define.WsConn) {
    fmt.Printf("WebSocket 连接建立: %s\n", conn.RemoteAddr())
    
    // 发送欢迎消息
    conn.WriteMessage(1, []byte("欢迎使用 WebSocket 服务"))
    
    // WebSocket 消息处理循环
    for {
        _, message, err := conn.ReadMessage()
        if err != nil {
            fmt.Printf("连接断开: %v\n", err)
            break
        }
        
        // 处理消息
        response := fmt.Sprintf("服务器收到: %s", string(message))
        err = conn.WriteMessage(1, []byte(response))
        if err != nil {
            fmt.Printf("发送消息失败: %v\n", err)
            break
        }
    }
}

func main() {
    app := mvc.HertzApp
    
    // 使用控制器方法注册 WebSocket 路由
    ns := mvc.NewNamespace("/chat",
        mvc.NSRouterWs("/room", &ChatController{}, "Index"),
    )
    
    mvc.AddNamespace(ns)
    app.Run()
}
```

## 📋 功能特性

### 1. 命名空间 WebSocket 路由

#### 基本路由
```go
mvc.NSWebSocket("/ws", handler)
```

#### 带路径参数的路由
```go
mvc.NSWebSocket("/chat/:room", func(conn *websocket.Conn) {
    // 处理聊天室连接
})
```

#### 自定义升级器
```go
upgrader := websocket.HertzUpgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(ctx *websocket.Context) bool {
        return true // 自定义源检查
    },
}

mvc.NSWebSocketWithUpgrader("/ws", handler, upgrader)
```

### 2. 中间件支持

WebSocket 路由会继承命名空间的中间件：

```go
ns := mvc.NewNamespace("/api",
    mvc.NSMiddleware(authMiddleware),  // 认证中间件
    mvc.NSMiddleware(corsMiddleware),  // CORS 中间件
    mvc.NSWebSocket("/ws", handler),   // WebSocket 路由继承上述中间件
)
```

### 3. 嵌套命名空间（✨ 已修复）

支持任意深度的嵌套命名空间，非常适合复杂的路由组织：

```go
// 基本嵌套命名空间
parentNS := mvc.NewNamespace("/api",
    mvc.NSNamespace("/v1",
        mvc.NSWebSocket("/ws", v1Handler),
    ),
    mvc.NSNamespace("/v2", 
        mvc.NSWebSocket("/ws", v2Handler),
    ),
)
// 生成路由: /api/v1/ws 和 /api/v2/ws

// 实际业务场景示例：管理后台心跳检测
adminNS := mvc.NewNamespace("/admin",
    mvc.NSNamespace("/online",
        mvc.NSRouterWs("/heartbeat", &OnlineController{}, "Index"),
    ),
    mvc.NSNamespace("/monitor",
        mvc.NSRouterWs("/system", &MonitorController{}, "SystemStatus"),
        mvc.NSRouterWs("/logs", &MonitorController{}, "LogStream"),
    ),
)
// 生成路由: 
// - /admin/online/heartbeat
// - /admin/monitor/system  
// - /admin/monitor/logs

mvc.AddNamespace(adminNS)
```

#### 深度嵌套支持

```go
// 三级嵌套命名空间
deepNS := mvc.NewNamespace("/api",
    mvc.NSNamespace("/v1",
        mvc.NSNamespace("/modules",
            mvc.NSNamespace("/chat",
                mvc.NSRouterWs("/room/:id", &ChatController{}, "HandleRoom"),
                mvc.NSRouterWs("/private/:user", &ChatController{}, "HandlePrivate"),
            ),
            mvc.NSNamespace("/game",
                mvc.NSRouterWs("/match", &GameController{}, "HandleMatch"),
                mvc.NSRouterWs("/spectate/:gameId", &GameController{}, "HandleSpectate"),
            ),
        ),
    ),
)
// 生成路由:
// - /api/v1/modules/chat/room/:id
// - /api/v1/modules/chat/private/:user
// - /api/v1/modules/game/match
// - /api/v1/modules/game/spectate/:gameId
```

## 🏗️ 高级功能

### 1. 连接管理

使用全局 WebSocket 管理器：

```go
func chatRoomHandler(conn *websocket.Conn) {
    manager := mvc.GetGlobalWebSocketManager()
    
    connID := "user_123"
    roomID := "room_456"
    
    // 添加连接
    manager.AddConnection(connID, conn)
    manager.JoinRoom(connID, roomID)
    
    defer func() {
        manager.LeaveRoom(connID, roomID)
        manager.RemoveConnection(connID)
    }()
    
    // 广播消息到房间
    manager.BroadcastToRoom(roomID, websocket.TextMessage, []byte("欢迎加入聊天室"))
    
    // 消息处理循环...
}
```

### 2. JSON 消息处理

在控制器中使用 JSON 处理器：

```go
func (c *APIController) JSONWebSocket() {
    handler := c.CreateWebSocketJSONHandler(func(data map[string]interface{}) map[string]interface{} {
        // 处理 JSON 消息
        response := map[string]interface{}{
            "echo": data,
            "timestamp": time.Now().Unix(),
        }
        return response
    })
    
    c.HandleWebSocket(handler, nil)
}
```

### 3. 连接池和心跳

```go
// 创建连接池
pool := mvc.NewWebSocketPool(1000, time.Minute*5)

// 创建心跳管理器
heartbeat := mvc.NewWebSocketHeartbeat(pool, time.Second*30, time.Second*10)
heartbeat.Start()

// 使用增强的处理器
handler := mvc.NewEnhancedWebSocketHandler()
defer handler.Close()
```

### 4. 限流和错误处理

```go
// 创建限流器
rateLimiter := mvc.NewWebSocketRateLimiter(100, time.Minute)

func protectedHandler(conn *websocket.Conn) {
    connID := generateConnectionID()
    
    for {
        messageType, message, err := conn.ReadMessage()
        if err != nil {
            break
        }
        
        // 检查限流
        if !rateLimiter.CheckRate(connID) {
            conn.WriteMessage(websocket.TextMessage, []byte("速率限制"))
            continue
        }
        
        // 处理消息...
    }
}
```

## 📝 完整示例

### 聊天室应用

```go
package main

import (
    "encoding/json"
    "log"
    "time"
    
    "github.com/hertz-contrib/websocket"
    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/framework/mvc/core"
)

type ChatMessage struct {
    Type     string `json:"type"`
    Room     string `json:"room"`
    User     string `json:"user"`
    Message  string `json:"message"`
    Time     int64  `json:"time"`
}

func main() {
    app := core.NewApp()
    manager := mvc.GetGlobalWebSocketManager()
    
    // 聊天室 WebSocket
    chatNS := mvc.NewNamespace("/chat",
        mvc.NSWebSocket("/room/:roomId", func(conn *websocket.Conn) {
            // 这里应该从路径参数获取 roomId
            roomID := "general" // 简化处理
            userID := fmt.Sprintf("user_%d", time.Now().UnixNano())
            
            // 连接管理
            manager.AddConnection(userID, conn)
            manager.JoinRoom(userID, roomID)
            defer func() {
                // 发送离开消息
                leaveMsg := ChatMessage{
                    Type:    "leave",
                    Room:    roomID,
                    User:    userID,
                    Message: "离开了聊天室",
                    Time:    time.Now().Unix(),
                }
                data, _ := json.Marshal(leaveMsg)
                manager.BroadcastToRoom(roomID, websocket.TextMessage, data)
                
                manager.LeaveRoom(userID, roomID)
                manager.RemoveConnection(userID)
            }()
            
            // 发送加入消息
            joinMsg := ChatMessage{
                Type:    "join",
                Room:    roomID,
                User:    userID,
                Message: "加入了聊天室",
                Time:    time.Now().Unix(),
            }
            data, _ := json.Marshal(joinMsg)
            manager.BroadcastToRoom(roomID, websocket.TextMessage, data)
            
            // 消息处理循环
            for {
                var msg ChatMessage
                err := conn.ReadJSON(&msg)
                if err != nil {
                    log.Printf("读取消息错误: %v", err)
                    break
                }
                
                // 设置消息信息
                msg.Type = "message"
                msg.Room = roomID
                msg.User = userID
                msg.Time = time.Now().Unix()
                
                // 广播消息
                data, _ := json.Marshal(msg)
                manager.BroadcastToRoom(roomID, websocket.TextMessage, data)
            }
        }),
    )
    
    // 注册命名空间
    chatNS.Register(app)
    
    // 静态文件服务
    app.Static("/", "./static")
    
    log.Println("聊天室服务器启动在 :8080")
    log.Println("WebSocket 端点: ws://localhost:8080/chat/room/general")
    log.Println("测试页面: http://localhost:8080/websocket_test.html")
    
    app.Run(":8080")
}
```

### 实时数据推送

```go
// 实时数据推送
streamNS := mvc.NewNamespace("/stream",
    mvc.NSWebSocket("/data", func(conn *websocket.Conn) {
        ticker := time.NewTicker(time.Second * 2)
        defer ticker.Stop()
        
        for {
            select {
            case <-ticker.C:
                data := map[string]interface{}{
                    "timestamp": time.Now().Unix(),
                    "cpu":       rand.Float64() * 100,
                    "memory":    rand.Float64() * 100,
                    "disk":      rand.Float64() * 100,
                }
                
                if err := conn.WriteJSON(data); err != nil {
                    return
                }
                
            default:
                // 检查连接状态
                if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                    return
                }
                time.Sleep(time.Millisecond * 100)
            }
        }
    }),
)
```

## 🧪 测试

### 使用提供的测试页面

框架提供了完整的测试页面用于验证 WebSocket 功能：

1. 启动 WebSocket 服务器
```bash
go run websocket_server.go
```

2. 在浏览器中访问测试页面：
   - `http://localhost:8888/static/nested_websocket_test.html` - **嵌套命名空间专用测试页面**
   - `http://localhost:8888/static/websocket_test.html` - 通用 WebSocket 测试页面

3. 测试不同的 WebSocket 端点：
   - ✅ `/admin/online/heartbeat` (嵌套命名空间)
   - ✅ `/ws/simple` (简单路由)
   - ✅ `/api/v1/websocket/test` (单级命名空间)
   - ✅ `/chat/room/general` (聊天室嵌套)

### 使用命令行客户端

使用提供的 Go 客户端测试工具：

```bash
go run websocket_client.go
```

### 使用 JavaScript 客户端

```javascript
// 测试嵌套命名空间路由
const ws = new WebSocket('ws://localhost:8888/admin/online/heartbeat');

ws.onopen = function() {
    console.log('✅ 嵌套命名空间连接成功');
    ws.send('test'); // 发送测试消息
    ws.send('ping'); // 发送心跳
};

ws.onmessage = function(event) {
    console.log('📨 收到消息:', event.data);
};

ws.onclose = function() {
    console.log('❌ 连接已关闭');
};

ws.onerror = function(error) {
    console.error('⚠️ 连接错误:', error);
};
```

### 手动测试验证

使用浏览器开发者工具验证连接：

1. 打开浏览器开发者工具 (F12)
2. 进入 Console 标签
3. 运行以下代码：

```javascript
// 创建 WebSocket 连接
const testConnection = (url, description) => {
    console.log(`🔄 测试 ${description}: ${url}`);
    const ws = new WebSocket(url);
    
    ws.onopen = () => {
        console.log(`✅ ${description} 连接成功`);
        ws.send('Hello from browser!');
    };
    
    ws.onmessage = (event) => {
        console.log(`📨 ${description} 收到:`, event.data);
        ws.close();
    };
    
    ws.onerror = (error) => {
        console.error(`❌ ${description} 连接失败:`, error);
    };
};

// 测试所有端点
testConnection('ws://localhost:8888/admin/online/heartbeat', '嵌套命名空间');
testConnection('ws://localhost:8888/ws/simple', '简单路由'); 
testConnection('ws://localhost:8888/api/v1/websocket/test', '单级命名空间');
testConnection('ws://localhost:8888/chat/room/general', '聊天室');
```

## 🔧 WebSocket 路由类型对比

### 函数式路由 vs 控制器路由

| 特性 | NSWebSocket | NSRouterWs |
|-----|-------------|-----------|
| 实现方式 | 直接函数 | 控制器方法 |
| 参数类型 | `*websocket.Conn` | `*define.WsConn` |
| 代码组织 | 适合简单逻辑 | 适合复杂业务逻辑 |
| 可测试性 | 较低 | 高（可以 mock 控制器） |
| 中间件支持 | 支持 | 支持 |
| 推荐场景 | 原型开发、简单功能 | 生产环境、复杂应用 |

```go
// 函数式路由 - 适合快速原型
mvc.NSWebSocket("/echo", func(conn *websocket.Conn) {
    // 简单逻辑
})

// 控制器路由 - 适合生产环境
mvc.NSRouterWs("/chat", &ChatController{}, "HandleChat")
```

## 🐛 常见问题排查

### 问题 1: WebSocket 连接失败 `{"isTrusted": true}`

**症状**: 浏览器显示连接错误，服务端返回 HTTP 200 而不是 101

**原因**: 
- 嵌套命名空间路由注册问题（v1.4.0 之前）
- 路径拼接错误

**解决方案**:
```go
// ❌ 错误的做法（旧版本问题）
// 系统内部会错误地使用 RouterPrefix 而不是 RouterWs

// ✅ 正确的做法（已修复）
nsOnline := mvc.NewNamespace("/admin",
    mvc.NSNamespace("/online",
        mvc.NSRouterWs("/heartbeat", &OnlineController{}, "Index"),
    ),
)
mvc.AddNamespace(nsOnline)
// 现在可以正常连接到: ws://localhost:8888/admin/online/heartbeat
```

### 问题 2: 控制器方法签名错误

**症状**: 运行时 panic，提示方法签名不匹配

**原因**: WebSocket 控制器方法必须使用特定的参数类型

**解决方案**:
```go
// ❌ 错误的签名
func (c *Controller) WebSocket(conn *websocket.Conn) {} // 参数类型错误

// ✅ 正确的签名
func (c *Controller) WebSocket(conn *define.WsConn) {} // 使用正确的类型
```

### 问题 3: 路径参数获取

**症状**: 无法在 WebSocket 处理器中获取路径参数

**解决方案**:
```go
// 对于函数式路由，需要通过其他方式传递参数
mvc.NSWebSocket("/room/:id", func(conn *websocket.Conn) {
    // conn 没有内置获取路径参数的方法
    // 可以通过查询参数或自定义头部传递
})

// 控制器路由可以在方法中通过上下文获取
func (c *ChatController) HandleRoom(conn *define.WsConn) {
    // 可以通过控制器上下文获取路径参数
    roomID := c.Ctx.Param("id")
    // ...
}
```

## 💡 最佳实践

### 1. 路由组织

```go
// ✅ 推荐: 按功能模块组织
chatNS := mvc.NewNamespace("/chat",
    mvc.NSRouterWs("/room/:id", &ChatController{}, "HandleRoom"),
    mvc.NSRouterWs("/private/:user", &ChatController{}, "HandlePrivate"),
)

// ✅ 推荐: 按 API 版本组织
apiV1 := mvc.NewNamespace("/api/v1",
    mvc.NSNamespace("/realtime",
        mvc.NSRouterWs("/notifications", &NotificationController{}, "HandleNotifications"),
        mvc.NSRouterWs("/updates", &UpdateController{}, "HandleUpdates"),
    ),
)
```

### 2. 错误处理

```go
func (c *ChatController) HandleChat(conn *define.WsConn) {
    defer func() {
        if r := recover(); r != nil {
            fmt.Printf("WebSocket 处理器 panic: %v\n", r)
            conn.WriteMessage(websocket.CloseMessage, []byte("服务器内部错误"))
            conn.Close()
        }
    }()
    
    for {
        _, message, err := conn.ReadMessage()
        if err != nil {
            if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
                fmt.Printf("WebSocket 意外断开: %v\n", err)
            }
            break
        }
        
        // 消息处理...
    }
}
```

### 3. 性能优化

```go
// 使用消息缓冲
type MessageBuffer struct {
    messages chan []byte
    done     chan struct{}
}

func (c *ChatController) HandleChatWithBuffer(conn *define.WsConn) {
    buffer := &MessageBuffer{
        messages: make(chan []byte, 100), // 缓冲 100 条消息
        done:     make(chan struct{}),
    }
    
    // 启动发送协程
    go c.messageSender(conn, buffer)
    
    // 主接收循环
    for {
        _, message, err := conn.ReadMessage()
        if err != nil {
            close(buffer.done)
            break
        }
        
        select {
        case buffer.messages <- message:
        case <-buffer.done:
            return
        }
    }
}
```

## ⚠️ 注意事项

1. **连接管理**: WebSocket 是长连接，注意内存和资源管理
2. **错误处理**: 适当处理连接断开、网络错误等异常情况  
3. **安全性**: 在生产环境中实现适当的认证和授权机制
4. **性能**: 对于大量连接，考虑使用连接池和负载均衡
5. **消息格式**: 统一消息格式，便于客户端和服务端处理
6. **路径拼接**: 确保使用最新版本，避免嵌套命名空间路径问题

## 🔧 配置选项

### WebSocket 升级器配置

```go
upgrader := websocket.HertzUpgrader{
    ReadBufferSize:   1024,
    WriteBufferSize:  1024,
    HandshakeTimeout: 10 * time.Second,
    CheckOrigin: func(ctx *websocket.Context) bool {
        // 检查请求来源
        origin := ctx.GetHeader("Origin")
        return isAllowedOrigin(string(origin))
    },
}
```

### 连接池配置

```go
pool := mvc.NewWebSocketPool(
    1000,                    // 最大连接数
    time.Minute * 5,        // 清理间隔
)
```

### 心跳配置

```go
heartbeat := mvc.NewWebSocketHeartbeat(
    pool,
    time.Second * 30,       // 心跳间隔
    time.Second * 10,       // 超时时间
)
```

## 📚 更多资源

### 官方文档
- [WebSocket RFC 6455](https://tools.ietf.org/html/rfc6455)
- [Hertz WebSocket 文档](https://www.cloudwego.io/docs/hertz/tutorials/basic-feature/protocol/websocket/)
- [YYHertz 框架主页](https://github.com/zsy619/yyhertz)

### 示例代码
- [完整服务器示例](../websocket_server.go) - 包含所有路由类型的完整服务器
- [客户端测试工具](../websocket_client.go) - 命令行测试客户端
- [WebSocket 控制器示例](../example/websocket/) - 各种控制器实现示例
- [聊天室示例](../example/websocket/simple/simple_websocket.go) - 完整的聊天室应用

### 测试页面
- [嵌套命名空间测试页面](../static/nested_websocket_test.html) - **推荐使用**
- [通用测试页面](../static/websocket_test.html) - 基础功能测试

### 问题反馈
如果遇到 WebSocket 相关问题，请在 GitHub 仓库提交 Issue，并提供：
1. YYHertz 框架版本
2. 完整的路由注册代码  
3. 客户端连接代码
4. 错误信息或异常日志

---

## 🎉 总结

通过本指南，你已经掌握了 YYHertz 框架的 WebSocket 功能：

✅ **基础功能**：函数式路由和控制器路由  
✅ **高级功能**：嵌套命名空间、中间件支持、连接管理  
✅ **生产就绪**：错误处理、性能优化、安全考虑  
✅ **问题排查**：常见问题解决方案和最佳实践  

现在可以在 YYHertz 框架中轻松实现各种 WebSocket 应用场景，从简单的实时通信到复杂的多人协作应用。
