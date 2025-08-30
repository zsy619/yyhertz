# YYHertz 模板引擎演示 - 静态资源

这个目录包含YYHertz模板引擎演示应用的静态资源文件。

## 目录结构

```
static/
├── css/
│   └── app.css          # 主样式文件
├── js/
│   └── app.js           # 主JavaScript文件
├── images/
│   └── favicon.ico      # 网站图标
└── README.md           # 本文件
```

## 样式文件 (CSS)

### app.css
主要样式文件，包含：
- CSS自定义属性（CSS变量）
- 通用组件样式（按钮、卡片、表单等）
- 响应式网格布局
- 工具类
- 深色主题支持

## JavaScript文件

### app.js
主要功能模块：
- **Utils**: 工具函数集合
- **CSRF**: CSRF令牌处理
- **TemplatePreview**: 模板预览功能
- **Performance**: 性能监控
- **Theme**: 主题切换
- **Search**: 搜索和过滤
- **Animation**: 动画效果

## 使用说明

在HTML模板中引用静态资源：

```html
<!-- CSS样式 -->
<link rel="stylesheet" href="/static/css/app.css">

<!-- JavaScript -->
<script src="/static/js/app.js"></script>
```

## 主题支持

应用支持浅色和深色主题切换：

```javascript
// 切换主题
YYHertz.Theme.toggle();

// 设置特定主题
YYHertz.Theme.apply('dark'); // 或 'light'
```

## CSRF保护

自动为表单和AJAX请求添加CSRF令牌：

```javascript
// 手动添加到表单
YYHertz.CSRF.addToForm(formElement);

// 手动添加到请求选项
const options = YYHertz.CSRF.addToRequest({
    method: 'POST',
    body: formData
});
```

## 工具函数示例

```javascript
// 格式化文件大小
YYHertz.Utils.formatFileSize(1024); // "1 KB"

// 显示消息提示
YYHertz.Utils.showMessage('操作成功', 'success');

// 复制到剪贴板
YYHertz.Utils.copyToClipboard('要复制的文本');

// AJAX请求
YYHertz.Utils.request('/api/data')
    .then(data => console.log(data))
    .catch(error => console.error(error));
```

## 浏览器兼容性

- 现代浏览器 (Chrome 60+, Firefox 60+, Safari 12+)
- 部分功能在旧浏览器中有降级处理
- 使用了 CSS Grid 和 Flexbox 布局
- 使用了 ES6+ 语法

## 开发建议

1. 使用CSS变量来保持样式的一致性
2. 利用提供的工具类来快速布局
3. 使用内置的动画函数来提升用户体验
4. 遵循响应式设计原则
5. 合理使用CSRF保护来增强安全性