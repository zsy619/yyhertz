# 🛣️ 路由管理系统

YYHertz控制器提供了强大的路由管理功能，支持动态路由、URL生成和路由参数处理。

## 🎯 路由管理方法

### 核心路由方法 (14个)

| 方法 | 说明 | 示例 |
|------|------|------|
| `GetParam(key)` | 获取路由参数 | `id := c.GetParam("id")` |
| `GetAllParams()` | 获取所有路由参数 | `params := c.GetAllParams()` |
| `URLFor(route, params...)` | 生成URL | `url := c.URLFor("user.show", 123)` |
| `Redirect(url, code...)` | 重定向 | `c.Redirect("/login")` |
| `RedirectToRoute(route, params...)` | 路由重定向 | `c.RedirectToRoute("user.list")` |

## 🔗 动态路由参数

### 路由参数获取

```go
// 路由定义: /users/:id/posts/:postId
func (c *UserController) GetUserPost() {
    // 📋 获取路径参数
    userID := c.GetParam("id")
    postID := c.GetParam("postId")
    
    // 📊 获取所有参数
    allParams := c.GetAllParams()
    c.LogInfo("路由参数", allParams)
    
    // 🔍 参数验证
    if userID == "" || postID == "" {
        c.JSONBadRequest("缺少必要参数")
        return
    }
    
    // 💾 业务逻辑
    post, err := c.postService.GetUserPost(userID, postID)
    if err != nil {
        c.JSONNotFound("文章不存在")
        return
    }
    
    c.JSONSuccess("获取成功", post)
}
```

## 🔧 URL生成工具

### 动态URL生成

```go
func (c *BaseController) generateURLs() {
    // 🔗 生成用户资料URL
    userURL := c.URLFor("user.profile", 123)
    // 结果: /users/123/profile
    
    // 🔗 生成文章编辑URL
    editURL := c.URLFor("post.edit", map[string]any{
        "id": 456,
        "action": "edit",
    })
    // 结果: /posts/456/edit?action=edit
    
    // 🔗 生成API URL
    apiURL := c.URLFor("api.users.list", map[string]any{
        "page": 1,
        "limit": 20,
        "sort": "created_at",
    })
    // 结果: /api/users?page=1&limit=20&sort=created_at
    
    // 📊 设置到模板数据
    c.SetData("UserURL", userURL)
    c.SetData("EditURL", editURL)
    c.SetData("APIURL", apiURL)
}
```

## 🚀 重定向处理

### 智能重定向

```go
func (c *AuthController) PostLogin() {
    username := c.GetString("username")
    password := c.GetString("password")
    
    // 🔐 验证用户身份
    user, err := c.authService.Login(username, password)
    if err != nil {
        c.JSONError("登录失败: " + err.Error())
        return
    }
    
    // 📝 保存用户信息
    c.SetSession("user_id", user.ID)
    c.SetSession("username", user.Username)
    
    // 🔄 智能重定向
    redirectURL := c.GetString("redirect_url")
    if redirectURL == "" {
        // 根据用户角色决定默认页面
        switch user.Role {
        case "admin":
            c.RedirectToRoute("admin.dashboard")
        case "user":
            c.RedirectToRoute("user.dashboard")
        default:
            c.Redirect("/")
        }
    } else {
        // 安全检查重定向URL
        if c.isSafeRedirectURL(redirectURL) {
            c.Redirect(redirectURL)
        } else {
            c.Redirect("/")
        }
    }
}

// 安全重定向URL检查
func (c *BaseController) isSafeRedirectURL(url string) bool {
    // 🛡️ 防止开放重定向攻击
    if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
        return false // 拒绝外部URL
    }
    
    if strings.Contains(url, "..") {
        return false // 拒绝路径遍历
    }
    
    return strings.HasPrefix(url, "/")
}
```

## 📱 RESTful路由设计

### 资源路由模式

```go
type PostController struct {
    core.BaseController
    postService *PostService
}

// GET /posts - 文章列表
func (c *PostController) GetIndex() {
    posts, err := c.postService.GetPosts()
    if err != nil {
        c.JSONError("获取列表失败")
        return
    }
    c.JSONSuccess("获取成功", posts)
}

// GET /posts/:id - 文章详情
func (c *PostController) GetShow() {
    id := c.GetParam("id")
    post, err := c.postService.GetPost(id)
    if err != nil {
        c.JSONNotFound("文章不存在")
        return
    }
    c.JSONSuccess("获取成功", post)
}

// POST /posts - 创建文章
func (c *PostController) PostCreate() {
    var req CreatePostRequest
    if err := c.BindJSON(&req); err != nil {
        c.JSONBadRequest("参数错误")
        return
    }
    
    post, err := c.postService.CreatePost(&req)
    if err != nil {
        c.JSONError("创建失败")
        return
    }
    c.JSONSuccess("创建成功", post)
}

// PUT /posts/:id - 更新文章
func (c *PostController) PutUpdate() {
    id := c.GetParam("id")
    var req UpdatePostRequest
    if err := c.BindJSON(&req); err != nil {
        c.JSONBadRequest("参数错误")
        return
    }
    
    err := c.postService.UpdatePost(id, &req)
    if err != nil {
        c.JSONError("更新失败")
        return
    }
    c.JSONSuccess("更新成功", nil)
}

// DELETE /posts/:id - 删除文章
func (c *PostController) DeleteDestroy() {
    id := c.GetParam("id")
    err := c.postService.DeletePost(id)
    if err != nil {
        c.JSONError("删除失败")
        return
    }
    c.JSONSuccess("删除成功", nil)
}
```

## 🎨 嵌套路由处理

### 复杂路由结构

```go
// 路由: /users/:userId/posts/:postId/comments/:commentId
func (c *CommentController) GetComment() {
    // 📋 获取嵌套参数
    userID := c.GetParam("userId")
    postID := c.GetParam("postId")
    commentID := c.GetParam("commentId")
    
    // 🔍 逐级验证
    if !c.userService.UserExists(userID) {
        c.JSONNotFound("用户不存在")
        return
    }
    
    if !c.postService.PostBelongsToUser(postID, userID) {
        c.JSONNotFound("文章不存在")
        return
    }
    
    comment, err := c.commentService.GetComment(commentID)
    if err != nil || comment.PostID != postID {
        c.JSONNotFound("评论不存在")
        return
    }
    
    c.JSONSuccess("获取成功", comment)
}

// 批量路由操作
func (c *PostController) PostBatchAction() {
    action := c.GetString("action")
    ids := c.GetStringSlice("ids")
    
    if len(ids) == 0 {
        c.JSONBadRequest("请选择要操作的项目")
        return
    }
    
    var err error
    switch action {
    case "delete":
        err = c.postService.BatchDelete(ids)
    case "publish":
        err = c.postService.BatchPublish(ids)
    case "archive":
        err = c.postService.BatchArchive(ids)
    default:
        c.JSONBadRequest("不支持的操作")
        return
    }
    
    if err != nil {
        c.JSONError("批量操作失败: " + err.Error())
        return
    }
    
    c.JSONSuccess("操作成功", nil)
}
```

## 🔧 路由中间件

### 路由级权限控制

```go
func (c *BaseController) RequireAuth() {
    if !c.IsAuthenticated() {
        if c.IsAjax() {
            c.JSONUnauthorized("请先登录")
        } else {
            // 保存当前URL用于登录后重定向
            currentURL := c.Ctx.Request.URL.String()
            loginURL := c.URLFor("auth.login", map[string]string{
                "redirect_url": currentURL,
            })
            c.Redirect(loginURL)
        }
        c.StopRun()
        return
    }
}

func (c *BaseController) RequireRole(role string) {
    c.RequireAuth() // 先确保已登录
    
    userRole := c.GetSession("role")
    if userRole == nil || userRole.(string) != role {
        c.JSONForbidden("权限不足")
        c.StopRun()
        return
    }
}

// 使用示例
func (c *AdminController) Prepare() {
    c.BaseController.Prepare()
    c.RequireRole("admin") // 需要管理员权限
}
```

## 📊 路由分组处理

```go
// API版本化路由
type APIv1Controller struct {
    core.BaseController
}

func (c *APIv1Controller) handleAPIRouting() {
    // 📋 版本信息
    c.SetHeader("API-Version", "v1")
    c.SetHeader("Content-Type", "application/json")
    
    // 🔍 路径分析
    path := c.Ctx.Path()
    resource := c.extractResource(path) // 如: "users", "posts"
    
    // 📊 统一处理
    switch resource {
    case "users":
        c.handleUserAPI()
    case "posts":
        c.handlePostAPI()
    case "comments":
        c.handleCommentAPI()
    default:
        c.JSONNotFound("API接口不存在")
    }
}
```

## ❓ 常见问题解答

**Q: 如何处理可选路由参数？**
A: 使用多个路由规则或在控制器中判断参数是否存在。

**Q: 路由参数中的特殊字符如何处理？**
A: 使用URL编码/解码，注意安全过滤。

**Q: 如何实现路由缓存优化？**
A: 在生产环境启用路由缓存，避免重复解析。

**Q: 动态路由如何实现？**
A: 使用正则表达式路由或自定义路由匹配逻辑。