# 🚨 Go模板内置动作冲突检测报告

## 📊 风险统计

- 🔴 **Critical风险**: 11个
- 🟡 **High风险**: 2个
- 🟢 **Medium风险**: 0个
- **总冲突数**: 13个

## 🔍 详细冲突分析

### 🔴 Critical风险

#### `or` → `Or`
**问题描述**: 与Go模板逻辑OR运算符冲突，可能导致逻辑表达式解析错误

**修复建议**:
- `logicOr` (推荐)
- `orOp`
- `logical_or`
- `boolOr`

#### `not` → `Not`
**问题描述**: 与Go模板逻辑NOT运算符冲突，可能导致逻辑表达式解析错误

**修复建议**:
- `logicNot` (推荐)
- `notOp`
- `logical_not`
- `boolNot`

#### `ge` → `Ge`
**问题描述**: 与Go模板大于等于比较运算符冲突，可能导致条件判断失效

**修复建议**:
- `greaterEqual` (推荐)
- `greaterOrEqual`
- `compare_ge`
- `gte`

#### `index` → `Index`
**问题描述**: 与Go模板内置index函数冲突，可能导致索引访问错误

**修复建议**:
- `getIndex` (推荐)
- `arrayIndex`
- `sliceIndex`
- `indexOf`

#### `le` → `Le`
**问题描述**: 与Go模板小于等于比较运算符冲突，可能导致条件判断失效

**修复建议**:
- `lessEqual` (推荐)
- `lessOrEqual`
- `compare_le`
- `lte`

#### `lt` → `Lt`
**问题描述**: 与Go模板小于比较运算符冲突，可能导致条件判断失效

**修复建议**:
- `lessThan` (推荐)
- `isLess`
- `compare_lt`
- `less`

#### `ne` → `Ne`
**问题描述**: 与Go模板不等于比较运算符冲突，可能导致条件判断失效

**修复建议**:
- `notEqual` (推荐)
- `notEquals`
- `isNotEqual`
- `compare_ne`

#### `len` → `Len`
**问题描述**: 与Go模板内置len函数冲突，可能导致长度计算错误

**修复建议**:
- `length` (推荐)
- `size`
- `count`
- `getLen`

#### `and` → `And`
**问题描述**: 与Go模板逻辑AND运算符冲突，可能导致逻辑表达式解析错误

**修复建议**:
- `logicAnd` (推荐)
- `andOp`
- `logical_and`
- `boolAnd`

#### `eq` → `Eq`
**问题描述**: 与Go模板等于比较运算符冲突，可能导致条件判断失效

**修复建议**:
- `equal` (推荐)
- `equals`
- `isEqual`
- `compare_eq`

#### `gt` → `Gt`
**问题描述**: 与Go模板大于比较运算符冲突，可能导致条件判断失效

**修复建议**:
- `greaterThan` (推荐)
- `isGreater`
- `compare_gt`
- `greater`

### 🟡 High风险

#### `printf` → `fmt.Sprintf`
**问题描述**: 与Go模板内置printf函数冲突，可能导致格式化输出错误

**修复建议**:
- `sprintf` (推荐)
- `format`
- `formatStr`
- `strFormat`

#### `slice` → `Slice`
**问题描述**: 与Go模板内置slice函数冲突，可能导致切片操作错误

**修复建议**:
- `subSlice` (推荐)
- `arraySlice`
- `getSlice`
- `slicePart`

## 🔧 代码修复示例

```go
// 🔧 Go内置动作冲突修复建议
// 以下函数名与Go模板内置动作冲突，建议重命名：

// ❌ 冲突函数: or → Or (Critical风险)
// 说明: 与Go模板逻辑OR运算符冲突，可能导致逻辑表达式解析错误
// 建议的安全别名：
"logicOr": Or, // 推荐使用 (替代冲突的'or')
"orOp": Or, // 备选方案1
"logical_or": Or, // 备选方案2
"boolOr": Or, // 备选方案3

// ❌ 冲突函数: not → Not (Critical风险)
// 说明: 与Go模板逻辑NOT运算符冲突，可能导致逻辑表达式解析错误
// 建议的安全别名：
"logicNot": Not, // 推荐使用 (替代冲突的'not')
"notOp": Not, // 备选方案1
"logical_not": Not, // 备选方案2
"boolNot": Not, // 备选方案3

// ❌ 冲突函数: ge → Ge (Critical风险)
// 说明: 与Go模板大于等于比较运算符冲突，可能导致条件判断失效
// 建议的安全别名：
"greaterEqual": Ge, // 推荐使用 (替代冲突的'ge')
"greaterOrEqual": Ge, // 备选方案1
"compare_ge": Ge, // 备选方案2
"gte": Ge, // 备选方案3

// ❌ 冲突函数: index → Index (Critical风险)
// 说明: 与Go模板内置index函数冲突，可能导致索引访问错误
// 建议的安全别名：
"getIndex": Index, // 推荐使用 (替代冲突的'index')
"arrayIndex": Index, // 备选方案1
"sliceIndex": Index, // 备选方案2
"indexOf": Index, // 备选方案3

// ❌ 冲突函数: le → Le (Critical风险)
// 说明: 与Go模板小于等于比较运算符冲突，可能导致条件判断失效
// 建议的安全别名：
"lessEqual": Le, // 推荐使用 (替代冲突的'le')
"lessOrEqual": Le, // 备选方案1
"compare_le": Le, // 备选方案2
"lte": Le, // 备选方案3

// ❌ 冲突函数: lt → Lt (Critical风险)
// 说明: 与Go模板小于比较运算符冲突，可能导致条件判断失效
// 建议的安全别名：
"lessThan": Lt, // 推荐使用 (替代冲突的'lt')
"isLess": Lt, // 备选方案1
"compare_lt": Lt, // 备选方案2
"less": Lt, // 备选方案3

// ❌ 冲突函数: ne → Ne (Critical风险)
// 说明: 与Go模板不等于比较运算符冲突，可能导致条件判断失效
// 建议的安全别名：
"notEqual": Ne, // 推荐使用 (替代冲突的'ne')
"notEquals": Ne, // 备选方案1
"isNotEqual": Ne, // 备选方案2
"compare_ne": Ne, // 备选方案3

// ❌ 冲突函数: len → Len (Critical风险)
// 说明: 与Go模板内置len函数冲突，可能导致长度计算错误
// 建议的安全别名：
"length": Len, // 推荐使用 (替代冲突的'len')
"size": Len, // 备选方案1
"count": Len, // 备选方案2
"getLen": Len, // 备选方案3

// ❌ 冲突函数: and → And (Critical风险)
// 说明: 与Go模板逻辑AND运算符冲突，可能导致逻辑表达式解析错误
// 建议的安全别名：
"logicAnd": And, // 推荐使用 (替代冲突的'and')
"andOp": And, // 备选方案1
"logical_and": And, // 备选方案2
"boolAnd": And, // 备选方案3

// ❌ 冲突函数: eq → Eq (Critical风险)
// 说明: 与Go模板等于比较运算符冲突，可能导致条件判断失效
// 建议的安全别名：
"equal": Eq, // 推荐使用 (替代冲突的'eq')
"equals": Eq, // 备选方案1
"isEqual": Eq, // 备选方案2
"compare_eq": Eq, // 备选方案3

// ❌ 冲突函数: gt → Gt (Critical风险)
// 说明: 与Go模板大于比较运算符冲突，可能导致条件判断失效
// 建议的安全别名：
"greaterThan": Gt, // 推荐使用 (替代冲突的'gt')
"isGreater": Gt, // 备选方案1
"compare_gt": Gt, // 备选方案2
"greater": Gt, // 备选方案3

// ❌ 冲突函数: printf → fmt.Sprintf (High风险)
// 说明: 与Go模板内置printf函数冲突，可能导致格式化输出错误
// 建议的安全别名：
"sprintf": fmt.Sprintf, // 推荐使用 (替代冲突的'printf')
"format": fmt.Sprintf, // 备选方案1
"formatStr": fmt.Sprintf, // 备选方案2
"strFormat": fmt.Sprintf, // 备选方案3

// ❌ 冲突函数: slice → Slice (High风险)
// 说明: 与Go模板内置slice函数冲突，可能导致切片操作错误
// 建议的安全别名：
"subSlice": Slice, // 推荐使用 (替代冲突的'slice')
"arraySlice": Slice, // 备选方案1
"getSlice": Slice, // 备选方案2
"slicePart": Slice, // 备选方案3

```

## 🎯 修复优先级建议

1. **立即修复Critical风险**: 这些冲突会导致模板解析完全失败
2. **尽快修复High风险**: 这些冲突会导致功能异常但不会完全阻塞
3. **计划修复Medium风险**: 这些冲突在特定场景下可能出现问题

## 📋 修复检查清单

- [ ] **Critical**: 重命名 `or` 为 `logicOr`
- [ ] **Critical**: 重命名 `not` 为 `logicNot`
- [ ] **Critical**: 重命名 `ge` 为 `greaterEqual`
- [ ] **Critical**: 重命名 `index` 为 `getIndex`
- [ ] **Critical**: 重命名 `le` 为 `lessEqual`
- [ ] **Critical**: 重命名 `lt` 为 `lessThan`
- [ ] ... (更多项目请参考详细列表)
