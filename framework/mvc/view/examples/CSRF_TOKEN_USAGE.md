# CSRF Token 模板访问方式

本文档介绍了在模板中访问CSRF token的多种方式。

## 问题描述

模板中使用 `{{.csrf_token}}` 时会报错：
```
template: Login.html:162:50: executing "Login.html" at <.csrf_token>: can't evaluate field csrf_token in type *view.RenderData
```

**根本原因**: Go语言的限制导致无法直接访问 `.csrf_token`（全小写）：
1. Go结构体的导出字段必须首字母大写
2. Go的导出方法必须首字母大写  
3. 模板引擎严格遵循Go的可见性规则
4. 因此无法创建名为 `csrf_token` 的字段或方法

## 解决方案

现在支持以下6种方式在模板中访问CSRF token：

### 1. 使用CsrfToken字段（推荐）
```html
<input type="hidden" name="csrf_token" value="{{.CsrfToken}}" />
<p>CSRF Token: {{.CsrfToken}}</p>
```
**说明**: 这是最接近原始需求的方案，使用驼峰命名的字段。

### 2. 直接访问CSRF字段
```html
<input type="hidden" name="csrf_token" value="{{.CSRF}}" />
<p>CSRF Token: {{.CSRF}}</p>
```

### 3. 使用Csrf_token方法
```html
<input type="hidden" name="csrf_token" value="{{.Csrf_token}}" />
<p>CSRF Token: {{.Csrf_token}}</p>
```

### 4. 使用GetCSRFToken方法
```html
<input type="hidden" name="csrf_token" value="{{.GetCSRFToken}}" />
<p>CSRF Token: {{.GetCSRFToken}}</p>
```

### 5. 使用csrf模板函数
```html
<input type="hidden" name="csrf_token" value="{{csrf}}" />
<p>CSRF Token: {{csrf}}</p>
```

### 6. 使用csrf_token模板函数
```html
<input type="hidden" name="csrf_token" value="{{csrf_token}}" />
<p>CSRF Token: {{csrf_token}}</p>
```

## 实际示例

以下是一个完整的登录表单示例：

```html
<form method="POST" action="/login">
    <!-- 使用.CsrfToken字段（推荐） -->
    <input type="hidden" name="csrf_token" value="{{.CsrfToken}}" />
    
    <div>
        <label for="username">用户名:</label>
        <input type="text" name="username" required>
    </div>
    
    <div>
        <label for="password">密码:</label>
        <input type="password" name="password" required>
    </div>
    
    <button type="submit">登录</button>
</form>

<!-- 调试信息 -->
<div style="display: none;">
    <p>CsrfToken字段: {{.CsrfToken}}</p>
    <p>CSRF字段: {{.CSRF}}</p>
    <p>Csrf_token方法: {{.Csrf_token}}</p>
    <p>GetCSRFToken方法: {{.GetCSRFToken}}</p>
    <p>csrf函数: {{csrf}}</p>
    <p>csrf_token函数: {{csrf_token}}</p>
</div>
```

## 使用建议

1. **首选方式**: 使用 `{{.CsrfToken}}` 字段，这是最接近原始 `.csrf_token` 需求的解决方案。

2. **替代方案**: 如果不想修改模板，可以使用 `{{csrf_token}}` 函数（不是 `.csrf_token`）。

3. **兼容性**: 所有6种方式都可以同时使用，选择最适合你项目的方式。

4. **调试**: 在开发环境中，可以在模板中显示CSRF token来确认其正确性。

## 重要提醒

⚠️ **无法使用 `{{.csrf_token}}`（全小写）**: 由于Go语言的限制，这种语法永远不会工作。

✅ **正确的替代方式**:
- `{{.CsrfToken}}` - 使用字段（推荐）
- `{{csrf_token}}` - 使用函数

## 自动CSRF处理

模板引擎现在会自动处理CSRF token：

- 如果传入普通数据（非RenderData），会自动创建RenderData并设置CSRF token
- 如果传入的RenderData的CSRF字段为空，会自动填充默认值
- 如果传入的RenderData已有CSRF值，会保持原值不变

## 测试验证

可以运行以下测试来验证CSRF token功能：

```bash
go test -v -run TestCSRFTokenAccess
go test -v -run TestRealTemplateCSRFAccess
go test -v -run TestCSRFTokenErrorRecovery
```

这些测试验证了：
- 模板中不同访问方式的正确性
- 真实模板渲染的功能
- 错误情况下的自动恢复机制