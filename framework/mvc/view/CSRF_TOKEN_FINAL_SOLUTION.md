# CSRF Token 模板访问问题最终解决方案

## 问题重述

用户在模板中使用 `{{.csrf_token}}` 时遇到错误：
```
template: Login.html:162:50: executing "Login.html" at <.csrf_token>: 
can't evaluate field csrf_token in type *view.RenderData
```

## 根本原因分析

这个错误**不是**代码bug，而是**Go语言的基本限制**：

1. **Go可见性规则**: 只有首字母大写的字段和方法才能被外部包访问
2. **结构体字段**: 必须首字母大写才能导出（如 `CsrfToken`，不能是 `csrf_token`）
3. **方法名称**: 必须首字母大写才能导出（如 `Csrf_token()`，不能是 `csrf_token()`）
4. **模板引擎**: 严格遵循Go的可见性规则

因此，**`{{.csrf_token}}` 在Go中永远不可能工作**。

## 最终解决方案

我们添加了 `CsrfToken` 字段到 `RenderData` 结构体：

```go
type RenderData struct {
    Data      any          `json:"data"`
    Meta      *MetaData    `json:"meta,omitempty"`
    Flash     *FlashData   `json:"flash,omitempty"`
    CSRF      string       `json:"csrf,omitempty"`
    CsrfToken string       `json:"csrf_token,omitempty"` // 新增字段
    Theme     string       `json:"theme,omitempty"`
    User      any          `json:"user,omitempty"`
    Request   *RequestData `json:"request,omitempty"`
}
```

## 用户需要做的修改

**将模板中的：**
```html
<input type="hidden" name="csrf_token" value="{{.csrf_token}}" />
```

**改为：**
```html
<input type="hidden" name="csrf_token" value="{{.CsrfToken}}" />
```

**注意**: 只需要将 `{{.csrf_token}}` 改为 `{{.CsrfToken}}`（C大写），其他都不用变。

## 自动字段同步

系统会自动确保 `CSRF` 和 `CsrfToken` 两个字段的值相同：

```go
// prepareRenderData 会自动同步两个字段
renderData.CSRF = csrfToken
renderData.CsrfToken = csrfToken  // 与CSRF字段保持同步
```

## 完整的可用方案

| 模板语法 | 类型 | 状态 | 说明 |
|----------|------|------|------|
| `{{.csrf_token}}` | 字段访问 | ❌ 不可用 | Go语言不支持 |
| `{{.CsrfToken}}` | 字段访问 | ✅ 可用 | **推荐方案** |
| `{{.CSRF}}` | 字段访问 | ✅ 可用 | 原有字段 |
| `{{.Csrf_token}}` | 方法调用 | ✅ 可用 | 调用方法 |
| `{{.GetCSRFToken}}` | 方法调用 | ✅ 可用 | 调用方法 |
| `{{csrf}}` | 模板函数 | ✅ 可用 | 使用函数 |
| `{{csrf_token}}` | 模板函数 | ✅ 可用 | 使用函数 |

## 迁移指南

### 步骤1: 查找所有使用 `.csrf_token` 的模板

```bash
find . -name "*.html" -exec grep -l "\.csrf_token" {} \;
```

### 步骤2: 批量替换

```bash
find . -name "*.html" -exec sed -i 's/{{\.csrf_token}}/{{.CsrfToken}}/g' {} \;
```

### 步骤3: 验证修改

确保所有模板中的 `{{.csrf_token}}` 都已被替换为 `{{.CsrfToken}}`。

## 测试验证

运行以下测试确认功能正常：

```bash
go test -v -run TestAllCsrfAccessMethods
go test -v -run TestLowercaseCsrfTokenAccess  
go test -v -run TestPracticalUsageExample
```

## 常见问题

**Q: 为什么不能直接支持 `.csrf_token`？**
A: Go语言的基本限制，无法绕过。所有导出的标识符必须首字母大写。

**Q: 有没有其他编程语言可以支持 `.csrf_token`？**  
A: 有，但Go不行。这是Go语言设计的一部分。

**Q: 我能用反射来绕过这个限制吗？**
A: 理论上可以，但html/template包不支持，且会破坏类型安全。

**Q: 我的旧模板很多，有批量修改的方法吗？**
A: 使用上面的迁移指南中的sed命令进行批量替换。

## 总结

这不是一个bug，而是Go语言的特性。解决方案是将 `{{.csrf_token}}` 改为 `{{.CsrfToken}}`。系统已经完全支持这种访问方式，并且会自动同步CSRF token的值。