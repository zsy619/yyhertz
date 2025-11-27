# YYHertz 模板函数参考

<div align="center">

🎨 **150+模板函数完整说明** | 让模板开发更高效

</div>

---

## 📋 函数分类索引

- [字符串处理函数](#字符串处理函数) (25个)
- [日期时间函数](#日期时间函数) (15个)
- [数学计算函数](#数学计算函数) (12个)
- [逻辑判断函数](#逻辑判断函数) (10个)
- [集合操作函数](#集合操作函数) (18个)
- [类型转换函数](#类型转换函数) (8个)
- [URL和编码函数](#url和编码函数) (12个)
- [模板包含函数](#模板包含函数) (15个)
- [格式化函数](#格式化函数) (10个)
- [其他实用函数](#其他实用函数) (25个)

---

## 🔤 字符串处理函数

### 基础字符串操作

#### `substr` / `Substring`
截取子字符串，支持中文字符
```html
{{substr "Hello World" 0 5}}        <!-- "Hello" -->
{{substr "你好世界" 0 2}}             <!-- "你好" -->
{{Substring "programming" 3 4}}     <!-- "gram" -->
```

#### `truncate`
截断字符串并添加省略号
```html
{{truncate "这是一个很长的字符串内容" 10}}  <!-- "这是一个很长的字..." -->
{{truncate "Short text" 20}}              <!-- "Short text" -->
```

#### `upper` / `lower` / `title`
大小写转换
```html
{{upper "hello"}}          <!-- "HELLO" -->
{{lower "WORLD"}}          <!-- "world" -->
{{title "hello world"}}    <!-- "Hello World" -->
{{tocapital "john doe"}}   <!-- "John doe" -->
```

#### `trim` / `trimprefix` / `trimsuffix`
字符串修剪
```html
{{trim "  hello world  "}}           <!-- "hello world" -->
{{trimprefix "https://example.com" "https://"}}  <!-- "example.com" -->
{{trimsuffix "filename.txt" ".txt"}} <!-- "filename" -->
```

#### `replace`
字符串替换
```html
{{replace "hello world" "world" "YYHertz"}}  <!-- "hello YYHertz" -->
{{replace "a-b-c" "-" "_"}}                  <!-- "a_b_c" -->
```

### 高级字符串处理

#### `contains` / `hasPrefix` / `hasSuffix`
字符串包含检查
```html
{{if contains .Email "@"}}这是一个邮箱地址{{end}}
{{if hasPrefix .URL "https://"}}安全链接{{end}}
{{if hasSuffix .Filename ".pdf"}}PDF文档{{end}}
```

#### `split` / `join`
字符串分割和连接
```html
{{$tags := split .TagString ","}}
{{range $tags}}<span class="tag">{{.}}</span>{{end}}

{{$words := split "apple,banana,orange" ","}}
{{join $words " | "}}  <!-- "apple | banana | orange" -->
```

#### `nl2br`
换行符转HTML换行
```html
{{nl2br "第一行\n第二行\n第三行"}}
<!-- 输出: 第一行<br>第二行<br>第三行 -->
```

#### `markdown`
简单Markdown渲染
```html
{{markdown "**粗体** *斜体* [链接](http://example.com)"}}
<!-- 输出: <strong>粗体</strong> <em>斜体</em> <a href="http://example.com">链接</a> -->
```

#### `striphtml`
去除HTML标签
```html
{{striphtml "<p>Hello <strong>World</strong></p>"}}  <!-- "Hello World" -->
```

---

## 📅 日期时间函数

### 日期格式化

#### `dateformat`
格式化日期时间
```html
{{dateformat .CreatedAt "2006-01-02"}}           <!-- "2024-09-05" -->
{{dateformat .UpdatedAt "2006-01-02 15:04:05"}}  <!-- "2024-09-05 14:30:25" -->
{{dateformat now "Y-m-d H:i:s"}}                 <!-- "2024-09-05 14:30:25" -->
```

#### `date`
获取当前日期
```html
{{date "2006-01-02"}}        <!-- 当前日期 -->
{{date "15:04:05"}}          <!-- 当前时间 -->
{{date "2006年01月02日"}}     <!-- "2024年09月05日" -->
```

### 相对时间

#### `timeago` / `timeSince`
显示相对时间（多久之前）
```html
{{timeago .CreatedAt}}       <!-- "2小时前" -->
{{timeSince .UpdatedAt}}     <!-- "5分钟前" -->
{{timeago .LoginAt}}         <!-- "3天前" -->
```

#### `timeuntil`
显示到某时间还有多久
```html
{{timeuntil .DeadlineAt}}    <!-- "3天后" -->
{{timeuntil .EventTime}}     <!-- "2小时后" -->
```

### 时间工具

#### `now`
获取当前时间对象
```html
{{$currentTime := now}}
当前时间：{{dateformat $currentTime "2006-01-02 15:04:05"}}
```

#### `timestamp`
获取当前时间戳
```html
{{timestamp}}  <!-- 1725529825 -->
```

#### `compare` / `comparenot`
比较时间
```html
{{if compare .StartTime .EndTime}}开始时间晚于结束时间{{end}}
{{if comparenot .CreatedAt now}}不是刚刚创建{{end}}
```

---

## 🔢 数学计算函数

### 基础运算

#### `add` / `sub` / `mul` / `div` / `mod`
基础数学运算
```html
{{add 10 20}}          <!-- 30 -->
{{sub 50 20}}          <!-- 30 -->
{{mul 6 7}}            <!-- 42 -->
{{div 100 4}}          <!-- 25 -->
{{mod 10 3}}           <!-- 1 -->

<!-- 多个数值运算 -->
{{add 10 20 30}}       <!-- 60 -->
{{mul 2 3 4}}          <!-- 24 -->
```

#### `round` / `ceil` / `floor`
数值取整
```html
{{round 3.14159}}      <!-- 3 -->
{{round 3.6}}          <!-- 4 -->
{{ceil 3.1}}           <!-- 4 -->
{{floor 3.9}}          <!-- 3 -->
```

#### `abs`
绝对值
```html
{{abs -15}}            <!-- 15 -->
{{abs 25}}             <!-- 25 -->
```

### 数值比较

#### `eq` / `ne` / `lt` / `le` / `gt` / `ge`
**注意：这些函数已重命名以避免Go内置冲突**
```html
<!-- 使用安全别名 -->
{{if equal .Status "active"}}激活状态{{end}}
{{if notEqual .Type "admin"}}非管理员{{end}}
{{if lessThan .Age 18}}未成年人{{end}}
{{if greaterThan .Score 90}}优秀{{end}}

<!-- 数值比较 -->
{{if lessThanOrEqual .Price 100}}价格合理{{end}}
{{if greaterThanOrEqual .Rating 4.5}}高评分{{end}}
```

---

## 🧠 逻辑判断函数

### 逻辑运算

#### `and` / `or` / `not`
逻辑运算符
```html
{{if and .IsLogin .IsAdmin}}管理员已登录{{end}}
{{if or .IsMember .IsVIP}}享有特权{{end}}
{{if not .IsDeleted}}未删除{{end}}
```

#### `empty` / `notnil`
空值检查
```html
{{if empty .Description}}暂无描述{{end}}
{{if notnil .Avatar}}
    <img src="{{.Avatar}}" alt="头像">
{{end}}
```

#### `default`
默认值设置
```html
{{.Username | default "匿名用户"}}
{{.Avatar | default "/static/images/default-avatar.png"}}
```

### 条件判断

#### `in`
检查元素是否在集合中
```html
{{if in .UserRole (slice "admin" "moderator")}}有管理权限{{end}}
{{if in .Status (slice "active" "pending")}}状态正常{{end}}
```

---

## 📚 集合操作函数

### 集合基础

#### `len`
获取长度（重命名为`length`避免冲突）
```html
{{length .Users}}个用户
{{length .Comments}}条评论
共{{length .Items}}项
```

#### `first` / `last`
获取首尾元素
```html
{{$firstUser := first .Users}}
{{$lastPost := last .Posts}}

{{if first .Items}}第一项：{{(first .Items).Name}}{{end}}
```

#### `index`
获取指定位置元素（重命名为`getIndex`）
```html
{{$secondItem := getIndex .Items 1}}
{{if $secondItem}}{{$secondItem.Name}}{{end}}
```

### 集合操作

#### `slice`
切片操作（重命名为`sliceArray`）
```html
{{$recentPosts := sliceArray .Posts 0 5}}
{{range $recentPosts}}
    <div>{{.Title}}</div>
{{end}}
```

#### `append`
追加元素
```html
{{$newList := append .ExistingList "新元素"}}
{{range $newList}}<li>{{.}}</li>{{end}}
```

#### `reverse`
反转集合
```html
{{$reversedList := reverse .Items}}
{{range $reversedList}}<div>{{.}}</div>{{end}}
```

#### `sort`
排序集合
```html
{{$sortedItems := sort .Items}}
{{range $sortedItems}}<option>{{.}}</option>{{end}}
```

### 高级集合操作

#### `unique`
去重
```html
{{$uniqueTags := unique .AllTags}}
{{range $uniqueTags}}<span class="tag">{{.}}</span>{{end}}
```

#### `compact`
移除空值
```html
{{$cleanList := compact .MaybeEmptyList}}
{{range $cleanList}}<p>{{.}}</p>{{end}}
```

#### `flatten`
扁平化嵌套数组
```html
{{$flatList := flatten .NestedArray}}
{{join $flatList ", "}}
```

---

## 🔄 类型转换函数

### 基础类型转换

#### `string` / `int` / `float` / `bool`
类型转换
```html
{{string 123}}              <!-- "123" -->
{{int "456"}}               <!-- 456 -->
{{float "3.14"}}            <!-- 3.14 -->
{{bool "true"}}             <!-- true -->

<!-- 在条件中使用 -->
{{if bool .IsActiveString}}激活状态{{end}}
```

#### `int64`
64位整数转换
```html
{{$timestamp := int64 .TimestampString}}
{{dateformat $timestamp "2006-01-02"}}
```

---

## 🌐 URL和编码函数

### URL处理

#### `urlencode` / `urldecode`
URL编码和解码
```html
{{urlencode "hello world"}}     <!-- "hello%20world" -->
{{urldecode "hello%20world"}}   <!-- "hello world" -->

<!-- 构建URL -->
<a href="/search?q={{urlencode .SearchTerm}}">搜索</a>
```

#### `urlfor`
生成URL
```html
{{urlfor "/api/users" "page" .CurrentPage "size" 20}}
<!-- 输出: /api/users?page=1&size=20 -->
```

### 编码函数

#### `base64enc` / `base64dec`
Base64编码和解码
```html
{{base64enc "hello world"}}     <!-- "aGVsbG8gd29ybGQ=" -->
{{base64dec "aGVsbG8gd29ybGQ="}} <!-- "hello world" -->
```

#### `md5`
MD5哈希
```html
{{md5 "password123"}}           <!-- "482c811da5d5b4bc6d497ffa98491e38" -->
```

#### `html2str` / `str2html`
HTML转换
```html
{{html2str "<p>Hello <strong>World</strong></p>"}}  <!-- "Hello World" -->
{{str2html "Hello <script>alert('xss')</script>"}}  <!-- 安全的HTML输出 -->
```

#### `htmlquote` / `htmlunquote`
HTML实体编码
```html
{{htmlquote "<script>"}}        <!-- "&lt;script&gt;" -->
{{htmlunquote "&lt;script&gt;"}} <!-- "<script>" -->
```

### 安全输出

#### `safejs` / `safehtml` / `raw`
安全输出函数
```html
{{safejs .JsonData}}            <!-- 安全的JavaScript输出 -->
{{safehtml .TrustedHTML}}       <!-- 安全的HTML输出 -->
{{raw .RawContent}}             <!-- 原始内容输出 -->
```

---

## 📄 模板包含函数

### 基础包含

#### `include` / `templatefunc` / `partial`
模板包含函数
```html
<!-- 包含头部模板 -->
{{include "partials/header.html" .}}

<!-- 使用templatefunc（推荐，避免冲突） -->
{{templatefunc "components/user-card.html" .User}}

<!-- 包含部分模板 -->
{{partial "layouts/sidebar.html" (makedict "menu" .MainMenu)}}
```

#### `component`
组件模板
```html
{{component "button" (makedict "text" "提交" "type" "primary" "onclick" "submitForm()")}}
{{component "modal" (makedict "title" "确认删除" "content" .ConfirmMessage)}}
```

#### `render`
渲染模板
```html
{{render "email/welcome.html" (makedict "user" .User "company" .Company)}}
```

### 高级包含

#### `prop` / `slot`
组件属性和插槽
```html
<!-- 在组件模板中使用 -->
<button class="{{prop "class" "btn-default"}}" onclick="{{prop "onclick" ""}}">
    {{slot "content" (prop "text" "按钮")}}
</button>

<!-- 调用组件时传递属性 -->
{{component "button" (makedict "text" "删除" "class" "btn-danger" "onclick" "deleteItem()")}}
```

---

## 📊 格式化函数

### 数值格式化

#### `number` / `percent`
数值和百分比格式化
```html
{{number 1234.5678 2}}         <!-- "1234.57" -->
{{percent 0.85 1}}             <!-- "85.0%" -->
```

#### `currency`
货币格式化
```html
{{currency 1234.56 "CNY"}}     <!-- "¥1234.56" -->
{{currency 999.99 "USD"}}      <!-- "$999.99" -->
```

#### `formatsize`
文件大小格式化
```html
{{formatsize 1048576}}         <!-- "1.00 MB" -->
{{formatsize 2048}}            <!-- "2.00 KB" -->
```

### 字符串格式化

#### `printf` / `sprintf`
格式化字符串
```html
{{printf "用户 %s 获得了 %d 积分" .Username .Points}}
{{sprintf "订单号：%08d" .OrderID}}  <!-- "订单号：00001234" -->
```

---

## 🛠️ 其他实用函数

### 数据构造

#### `makedict` / `makeslice` / `makerange`
构造数据结构
```html
<!-- 构造字典 -->
{{$data := makedict "name" "张三" "age" 25 "email" "zhangsan@example.com"}}
{{$data.name}}  <!-- "张三" -->

<!-- 构造切片 -->
{{$tags := makeslice "Go" "Web" "Framework"}}
{{range $tags}}<span>{{.}}</span>{{end}}

<!-- 构造数字范围 -->
{{range makerange 1 5}}
    <option value="{{.}}">第{{.}}页</option>
{{end}}
```

#### `makeseq`
构造序列
```html
{{range makeseq 3}}
    <div class="item-{{.}}">项目 {{add . 1}}</div>
{{end}}
```

### 随机和唯一值

#### `uuid` / `random`
生成UUID和随机字符串
```html
{{uuid}}                       <!-- "550e8400-e29b-41d4-a716-446655440000" -->
{{random 8}}                   <!-- "aB3xY9Qm" -->
```

#### `shuffle`
随机打乱集合
```html
{{$randomItems := shuffle .Items}}
{{range $randomItems}}<div>{{.}}</div>{{end}}
```

### 配置和环境

#### `config`
获取配置值
```html
{{config "String" "app.name" "默认应用名"}}
{{config "Int" "app.port" 8080}}
{{config "Bool" "app.debug" false}}
```

#### `map_get`
从map中获取值
```html
{{map_get .UserData "profile" "avatar"}}  <!-- 获取 .UserData["profile"]["avatar"] -->
```

### 表单和资源

#### `renderform`
渲染表单
```html
{{renderform (makedict
    "method" "POST"
    "action" "/users"
    "fields" (makeslice
        (makedict "type" "text" "name" "username" "label" "用户名" "required" true)
        (makedict "type" "email" "name" "email" "label" "邮箱" "required" true)
        (makedict "type" "password" "name" "password" "label" "密码" "required" true)
    )
)}}
```

#### `assets_js` / `assets_css`
资源引入
```html
{{assets_js "/static/js/app.js" "/static/js/utils.js"}}
{{assets_css "/static/css/app.css" "/static/css/theme.css"}}
```

### CSRF保护

#### `csrf_token`
获取CSRF令牌
```html
<form method="POST">
    <input type="hidden" name="csrf_token" value="{{csrf_token}}">
    <!-- 表单字段 -->
</form>
```

### 国际化

#### `i18n` / `trans`
国际化翻译
```html
{{i18n "welcome.message" .Username}}       <!-- "欢迎 张三" -->
{{trans "error.not_found"}}                <!-- "未找到资源" -->
```

---

## 💡 使用技巧和最佳实践

### 1. 函数链式调用
```html
{{.Content | truncate 100 | nl2br}}
{{.Price | mul 0.8 | currency "CNY"}}
{{.Tags | join ", " | upper}}
```

### 2. 条件渲染
```html
{{if and .User (equal .User.Role "admin")}}
    <div class="admin-panel">管理员面板</div>
{{end}}

{{if or (empty .Posts) (lt (len .Posts) 1)}}
    <div class="no-content">暂无内容</div>
{{end}}
```

### 3. 循环中使用函数
```html
{{range $index, $user := .Users}}
    <tr class="{{if even $index}}even{{else}}odd{{end}}">
        <td>{{add $index 1}}</td>
        <td>{{$user.Name | default "未知用户"}}</td>
        <td>{{timeago $user.CreatedAt}}</td>
        <td>
            {{if equal $user.Status "active"}}
                <span class="status-active">激活</span>
            {{else}}
                <span class="status-inactive">未激活</span>
            {{end}}
        </td>
    </tr>
{{end}}
```

### 4. 复杂数据操作
```html
{{$activeUsers := slice .Users}}
{{range .Users}}
    {{if equal .Status "active"}}
        {{$activeUsers = append $activeUsers .}}
    {{end}}
{{end}}

<p>活跃用户：{{len $activeUsers}}人</p>
```

### 5. 组件化开发
```html
<!-- 定义可复用的用户卡片组件 -->
{{define "user-card"}}
<div class="user-card">
    <img src="{{.avatar | default "/static/images/default-avatar.png"}}" alt="头像">
    <h3>{{.name | default "匿名用户"}}</h3>
    <p>{{.bio | truncate 100}}</p>
    <span class="join-date">{{timeago .created_at}}</span>
</div>
{{end}}

<!-- 使用组件 -->
{{range .Users}}
    {{templatefunc "user-card" .}}
{{end}}
```

---

<div align="center">

**🎨 YYHertz提供了150+强大的模板函数，让前端开发更高效！**

**善用这些函数，可以大大简化模板逻辑，提升开发体验 ✨**

</div>