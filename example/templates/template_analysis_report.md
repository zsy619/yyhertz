# 🔍 YYHertz 模板函数命名分析报告

## 📊 基本统计

- **总函数数量**: 136
- **Go内置动作数量**: 28
- **潜在冲突数量**: 13

## ⚠️ Go内置动作冲突

以下函数名与Go模板内置动作冲突：

- `and` → `And` ⚠️ **冲突风险高**
- `eq` → `Eq` ⚠️ **冲突风险高**
- `ge` → `Ge` ⚠️ **冲突风险高**
- `gt` → `Gt` ⚠️ **冲突风险高**
- `index` → `Index` ⚠️ **冲突风险高**
- `le` → `Le` ⚠️ **冲突风险高**
- `len` → `Len` ⚠️ **冲突风险高**
- `lt` → `Lt` ⚠️ **冲突风险高**
- `ne` → `Ne` ⚠️ **冲突风险高**
- `not` → `Not` ⚠️ **冲突风险高**
- `or` → `Or` ⚠️ **冲突风险高**
- `printf` → `fmt.Sprintf` ⚠️ **冲突风险高**
- `slice` → `Slice` ⚠️ **冲突风险高**

## 🏷️ 函数别名模式分析

发现 24 个函数有多个别名：

### `NotNil`
别名: `notNil`, `notNull`, `notnil`, `notnull`

### `xstring.Substr`
别名: `Substring`, `substr`

### `TemplateInclude`
别名: `templatefunc`, `templateinclude`

### `GetCSRFTokenFromContext`
别名: `csrf`, `csrf_token`

### `fmt.Sprintf`
别名: `printf`, `sprintf`

### `SafeJS`
别名: `safeJs`, `safejs`

### `TimeSince`
别名: `timeSince`, `timesince`

### `URLDecode`
别名: `urlDecode`, `urldecode`

### `CreateRange`
别名: `makeRange`, `makerange`

### `I18n`
别名: `i18n`, `trans`

### `CreateSequence`
别名: `makeSeq`, `makeseq`

### `Include`
别名: `include`, `includeTmpl`, `tmpl_include`

### `TimeAgo`
别名: `timeAgo`, `timeago`

### `URLEncode`
别名: `urlEncode`, `urlencode`

### `MakeSlice`
别名: `makeSlice`, `makeslice`

### `Base64Decode`
别名: `base64Dec`, `base64dec`

### `SafeHTML`
别名: `safeHtml`, `safehtml`

### `RenderTemplate`
别名: `render`, `renderPartial`

### `MakeDict`
别名: `makeDict`, `makedict`

### `TimeUntil`
别名: `timeUntil`, `timeuntil`

### `RawHTML`
别名: `raw`, `unescaped`

### `CompareNot`
别名: `compareNot`, `comparenot`

### `Base64Encode`
别名: `base64Enc`, `base64enc`

### `FmtByte`
别名: `formatFileSize`, `formatSize`, `formatfilesize`, `formatsize`

## 🔧 建议添加的标准别名

### `TemplateInclude`
现有别名: `templatefunc`, `templateinclude`
建议添加: `tEmplateinclude`, `tEmplatefunc`

### `GetCSRFTokenFromContext`
现有别名: `csrf`, `csrf_token`
建议添加: `cSrf`

### `RawHTML`
现有别名: `raw`, `unescaped`
建议添加: `uNescaped`, `rAw`

### `xstring.Substr`
现有别名: `Substring`, `substr`
建议添加: `sUbstr`

### `fmt.Sprintf`
现有别名: `printf`, `sprintf`
建议添加: `sPrintf`, `pRintf`

### `I18n`
现有别名: `i18n`, `trans`
建议添加: `tRans`

## 💡 命名规范建议

### 标准化原则
1. **避免冲突**: 不使用Go内置动作名称
2. **多样性**: 提供驼峰、小写、下划线等多种风格
3. **一致性**: 同类功能函数使用相似的命名模式
4. **简洁性**: 优先选择简短易记的主要名称

### 推荐的函数别名模式
```
主函数名: actualFunction
驼峰别名: camelCaseAlias (如: timeAgo)
小写别名: lowercase (如: timeago)
下划线别名: under_score (如: time_ago)
简化别名: short (如: ago)
```

