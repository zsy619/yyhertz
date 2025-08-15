# 🎨 YYHertz 错误页面定制指南

本文档将详细介绍如何定制和创建个性化的错误页面，包括HTML模板定制、CSS样式调整和JavaScript交互功能开发。

## 🏗️ 错误页面架构

### DefaultErrorController 模板系统

YYHertz的`DefaultErrorController`使用内嵌的HTML模板系统，具有以下特性：

```go
// DefaultErrorController 核心结构
type DefaultErrorController struct {
    CustomTitle    string                 // 自定义页面标题
    CustomCSS      string                 // 自定义CSS样式
    CustomJS       string                 // 自定义JavaScript
    TemplateEngine *template.Template     // 模板引擎
    Config         *ErrorPageConfig       // 页面配置
}

// ErrorPageConfig 错误页面配置
type ErrorPageConfig struct {
    ShowStackTrace    bool     `json:"show_stack_trace"`    // 显示堆栈跟踪
    ShowRequestInfo   bool     `json:"show_request_info"`   // 显示请求信息
    ShowTimestamp     bool     `json:"show_timestamp"`      // 显示时间戳
    Theme            string    `json:"theme"`               // 主题：light/dark
    Language         string    `json:"language"`            // 语言：zh/en
    EnableReporting  bool      `json:"enable_reporting"`    // 启用错误报告
    ContactInfo      string    `json:"contact_info"`        // 联系信息
    CompanyName      string    `json:"company_name"`        // 公司名称
    SupportURL       string    `json:"support_url"`         // 支持链接
}
```

### 模板变量系统

错误页面模板支持以下变量：

| 变量名 | 类型 | 描述 | 示例值 |
|--------|------|------|--------|
| `{{.StatusCode}}` | int | HTTP状态码 | 404 |
| `{{.StatusText}}` | string | 状态文本 | Not Found |
| `{{.RequestPath}}` | string | 请求路径 | /api/user/123 |
| `{{.RequestMethod}}` | string | 请求方法 | GET |
| `{{.UserAgent}}` | string | 用户代理 | Chrome/120.0 |
| `{{.Timestamp}}` | string | 错误时间 | 2024-03-15 14:30:00 |
| `{{.ErrorMessage}}` | string | 错误信息 | 用户不存在 |
| `{{.StackTrace}}` | string | 堆栈跟踪 | main.go:123 |
| `{{.CustomTitle}}` | string | 自定义标题 | My App Error |

## 🎨 自定义错误页面样式

### 1. CSS样式定制

#### 基础样式覆盖

```go
func NewCustomErrorController() *DefaultErrorController {
    controller := NewDefaultErrorController()
    
    // 自定义CSS样式
    controller.CustomCSS = `
    :root {
        --primary-color: #667eea;
        --secondary-color: #764ba2;
        --error-color: #e53e3e;
        --warning-color: #dd6b20;
        --success-color: #38a169;
        --text-color: #2d3748;
        --bg-gradient: linear-gradient(135deg, var(--primary-color), var(--secondary-color));
    }
    
    body {
        font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
        background: var(--bg-gradient);
        min-height: 100vh;
        margin: 0;
        padding: 20px;
    }
    
    .error-container {
        max-width: 800px;
        margin: 0 auto;
        background: rgba(255, 255, 255, 0.95);
        backdrop-filter: blur(10px);
        border-radius: 16px;
        box-shadow: 0 20px 40px rgba(0, 0, 0, 0.1);
        overflow: hidden;
        animation: slideInUp 0.6s ease-out;
    }
    
    @keyframes slideInUp {
        from {
            opacity: 0;
            transform: translateY(30px);
        }
        to {
            opacity: 1;
            transform: translateY(0);
        }
    }
    
    .error-header {
        background: linear-gradient(45deg, #ff6b6b, #ee5a24);
        color: white;
        padding: 40px;
        text-align: center;
        position: relative;
        overflow: hidden;
    }
    
    .error-header::before {
        content: '';
        position: absolute;
        top: -50%;
        left: -50%;
        width: 200%;
        height: 200%;
        background: url('data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><circle cx="50" cy="50" r="2" fill="rgba(255,255,255,0.1)"/></svg>');
        animation: float 20s linear infinite;
    }
    
    @keyframes float {
        from { transform: translateX(-50px) translateY(-50px) rotate(0deg); }
        to { transform: translateX(-50px) translateY(-50px) rotate(360deg); }
    }
    
    .error-code {
        font-size: 6rem;
        font-weight: 800;
        margin: 0;
        text-shadow: 0 4px 8px rgba(0, 0, 0, 0.3);
        position: relative;
        z-index: 1;
    }
    
    .error-title {
        font-size: 2rem;
        margin: 10px 0 0 0;
        font-weight: 600;
        position: relative;
        z-index: 1;
    }
    
    .error-body {
        padding: 40px;
        line-height: 1.6;
    }
    
    .error-message {
        font-size: 1.1rem;
        color: var(--text-color);
        margin-bottom: 30px;
        padding: 20px;
        background: #f8f9ff;
        border-left: 4px solid var(--primary-color);
        border-radius: 4px;
    }
    
    .error-actions {
        display: flex;
        gap: 15px;
        flex-wrap: wrap;
        margin: 30px 0;
    }
    
    .btn {
        display: inline-flex;
        align-items: center;
        padding: 12px 24px;
        border-radius: 8px;
        text-decoration: none;
        font-weight: 500;
        transition: all 0.3s ease;
        cursor: pointer;
        border: none;
        font-size: 1rem;
    }
    
    .btn-primary {
        background: var(--bg-gradient);
        color: white;
        box-shadow: 0 4px 15px rgba(102, 126, 234, 0.4);
    }
    
    .btn-primary:hover {
        transform: translateY(-2px);
        box-shadow: 0 6px 20px rgba(102, 126, 234, 0.6);
    }
    
    .btn-secondary {
        background: white;
        color: var(--text-color);
        border: 2px solid #e2e8f0;
    }
    
    .btn-secondary:hover {
        background: #f7fafc;
        border-color: var(--primary-color);
    }
    
    .error-details {
        background: #f8f9fa;
        border-radius: 8px;
        padding: 20px;
        margin-top: 20px;
        border: 1px solid #e9ecef;
    }
    
    .detail-item {
        display: flex;
        justify-content: space-between;
        padding: 8px 0;
        border-bottom: 1px solid #e9ecef;
    }
    
    .detail-item:last-child {
        border-bottom: none;
    }
    
    .detail-label {
        font-weight: 600;
        color: var(--text-color);
        min-width: 120px;
    }
    
    .detail-value {
        color: #666;
        word-break: break-all;
    }
    
    /* 深色主题 */
    [data-theme="dark"] {
        --text-color: #e2e8f0;
        --bg-gradient: linear-gradient(135deg, #1a202c, #2d3748);
    }
    
    [data-theme="dark"] .error-container {
        background: rgba(26, 32, 44, 0.95);
    }
    
    [data-theme="dark"] .error-details {
        background: #2d3748;
        border-color: #4a5568;
    }
    
    [data-theme="dark"] .detail-item {
        border-color: #4a5568;
    }
    
    /* 响应式设计 */
    @media (max-width: 768px) {
        .error-container {
            margin: 10px;
            border-radius: 12px;
        }
        
        .error-header {
            padding: 30px 20px;
        }
        
        .error-code {
            font-size: 4rem;
        }
        
        .error-title {
            font-size: 1.5rem;
        }
        
        .error-body {
            padding: 30px 20px;
        }
        
        .error-actions {
            flex-direction: column;
        }
        
        .btn {
            justify-content: center;
        }
        
        .detail-item {
            flex-direction: column;
            gap: 4px;
        }
    }
    `
    
    return controller
}
```

#### 主题切换功能

```go
// 添加主题切换的JavaScript
controller.CustomJS = `
function initTheme() {
    const themeToggle = document.getElementById('theme-toggle');
    const currentTheme = localStorage.getItem('error-page-theme') || 'light';
    
    document.documentElement.setAttribute('data-theme', currentTheme);
    updateThemeToggle(currentTheme);
    
    if (themeToggle) {
        themeToggle.addEventListener('click', function() {
            const currentTheme = document.documentElement.getAttribute('data-theme');
            const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
            
            document.documentElement.setAttribute('data-theme', newTheme);
            localStorage.setItem('error-page-theme', newTheme);
            updateThemeToggle(newTheme);
        });
    }
}

function updateThemeToggle(theme) {
    const themeToggle = document.getElementById('theme-toggle');
    if (themeToggle) {
        themeToggle.innerHTML = theme === 'dark' ? 
            '<i class="fas fa-sun"></i>' : 
            '<i class="fas fa-moon"></i>';
        themeToggle.title = theme === 'dark' ? '切换到浅色主题' : '切换到深色主题';
    }
}

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', function() {
    initTheme();
});
`
```

### 2. 创建自定义模板

#### 替换默认模板

```go
func NewBrandedErrorController(brand *BrandConfig) *DefaultErrorController {
    controller := &DefaultErrorController{
        CustomTitle: brand.CompanyName + " - 页面错误",
        Config: &ErrorPageConfig{
            CompanyName: brand.CompanyName,
            SupportURL:  brand.SupportURL,
            ContactInfo: brand.ContactInfo,
            Theme:      "light",
            Language:   "zh",
        },
    }
    
    // 自定义HTML模板
    customTemplate := `
    <!DOCTYPE html>
    <html lang="{{.Language}}">
    <head>
        <meta charset="UTF-8">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <title>{{.CustomTitle}}</title>
        <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css" rel="stylesheet">
        <style>{{.CustomCSS}}</style>
    </head>
    <body>
        <div class="error-container">
            <!-- 品牌头部 -->
            <div class="brand-header">
                <img src="{{.BrandLogo}}" alt="{{.CompanyName}}" class="brand-logo">
                <h1 class="brand-name">{{.CompanyName}}</h1>
            </div>
            
            <!-- 错误内容 -->
            <div class="error-content">
                <div class="error-icon">
                    <i class="{{.StatusIcon}}" aria-hidden="true"></i>
                </div>
                
                <h2 class="error-code">{{.StatusCode}}</h2>
                <h3 class="error-title">{{.StatusText}}</h3>
                
                {{if .ErrorMessage}}
                <div class="error-message">
                    <i class="fas fa-exclamation-circle"></i>
                    <span>{{.ErrorMessage}}</span>
                </div>
                {{end}}
                
                <!-- 操作按钮 -->
                <div class="error-actions">
                    <a href="/" class="btn btn-primary">
                        <i class="fas fa-home"></i>
                        返回首页
                    </a>
                    <button onclick="history.back()" class="btn btn-secondary">
                        <i class="fas fa-arrow-left"></i>
                        返回上页
                    </button>
                    <button onclick="location.reload()" class="btn btn-secondary">
                        <i class="fas fa-refresh"></i>
                        刷新页面
                    </button>
                    {{if .SupportURL}}
                    <a href="{{.SupportURL}}" class="btn btn-outline">
                        <i class="fas fa-life-ring"></i>
                        获取帮助
                    </a>
                    {{end}}
                </div>
                
                <!-- 建议信息 -->
                <div class="suggestions">
                    <h4><i class="fas fa-lightbulb"></i> 可能的解决方案：</h4>
                    <ul class="suggestion-list">
                        {{range .Suggestions}}
                        <li>{{.}}</li>
                        {{end}}
                    </ul>
                </div>
                
                <!-- 详细信息（开发环境） -->
                {{if .ShowRequestInfo}}
                <details class="error-details">
                    <summary>技术详情 <i class="fas fa-chevron-down"></i></summary>
                    <div class="detail-content">
                        <div class="detail-item">
                            <span class="detail-label">请求路径：</span>
                            <span class="detail-value">{{.RequestPath}}</span>
                        </div>
                        <div class="detail-item">
                            <span class="detail-label">请求方法：</span>
                            <span class="detail-value">{{.RequestMethod}}</span>
                        </div>
                        <div class="detail-item">
                            <span class="detail-label">用户代理：</span>
                            <span class="detail-value">{{.UserAgent}}</span>
                        </div>
                        <div class="detail-item">
                            <span class="detail-label">时间戳：</span>
                            <span class="detail-value">{{.Timestamp}}</span>
                        </div>
                        {{if .ShowStackTrace}}
                        <div class="detail-item">
                            <span class="detail-label">堆栈跟踪：</span>
                            <pre class="detail-value stack-trace">{{.StackTrace}}</pre>
                        </div>
                        {{end}}
                    </div>
                </details>
                {{end}}
                
                <!-- 错误报告 -->
                {{if .EnableReporting}}
                <div class="error-reporting">
                    <h4><i class="fas fa-bug"></i> 帮助我们改进</h4>
                    <p>如果您认为这是一个bug，请点击下面的按钮报告问题：</p>
                    <button onclick="reportError()" class="btn btn-warning">
                        <i class="fas fa-flag"></i>
                        报告问题
                    </button>
                </div>
                {{end}}
            </div>
            
            <!-- 页脚 -->
            <div class="error-footer">
                <p>&copy; {{.Year}} {{.CompanyName}}. All rights reserved.</p>
                {{if .ContactInfo}}
                <p>联系我们：{{.ContactInfo}}</p>
                {{end}}
                <div class="theme-controls">
                    <button id="theme-toggle" class="theme-toggle" title="切换主题">
                        <i class="fas fa-moon"></i>
                    </button>
                </div>
            </div>
        </div>
        
        <script>{{.CustomJS}}</script>
    </body>
    </html>
    `
    
    // 编译模板
    tmpl, err := template.New("error").Parse(customTemplate)
    if err != nil {
        log.Printf("Template parsing error: %v", err)
    } else {
        controller.TemplateEngine = tmpl
    }
    
    return controller
}
```

### 3. 交互功能开发

#### 错误报告功能

```javascript
// 错误报告JavaScript功能
function reportError() {
    const errorData = {
        statusCode: parseInt(document.querySelector('.error-code').textContent),
        statusText: document.querySelector('.error-title').textContent,
        path: getMetaValue('request-path'),
        method: getMetaValue('request-method'),
        userAgent: navigator.userAgent,
        timestamp: new Date().toISOString(),
        viewport: {
            width: window.innerWidth,
            height: window.innerHeight
        },
        url: window.location.href,
        referrer: document.referrer,
        language: navigator.language
    };
    
    // 显示报告对话框
    showReportDialog(errorData);
}

function showReportDialog(errorData) {
    const dialog = createReportDialog();
    document.body.appendChild(dialog);
    
    // 填充错误数据
    document.getElementById('report-data').textContent = 
        JSON.stringify(errorData, null, 2);
    
    // 显示对话框
    dialog.style.display = 'block';
    dialog.classList.add('show');
}

function createReportDialog() {
    const dialog = document.createElement('div');
    dialog.className = 'report-dialog';
    dialog.innerHTML = `
        <div class="report-content">
            <div class="report-header">
                <h3><i class="fas fa-bug"></i> 错误报告</h3>
                <button onclick="closeReportDialog()" class="close-btn">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="report-body">
                <p>请描述您遇到的问题（可选）：</p>
                <textarea id="user-description" rows="4" 
                    placeholder="请简要描述您在做什么时遇到了这个错误..."></textarea>
                
                <p>以下技术信息将被发送：</p>
                <pre id="report-data" class="report-data"></pre>
                
                <div class="report-actions">
                    <button onclick="submitReport()" class="btn btn-primary">
                        <i class="fas fa-paper-plane"></i>
                        发送报告
                    </button>
                    <button onclick="closeReportDialog()" class="btn btn-secondary">
                        取消
                    </button>
                </div>
            </div>
        </div>
    `;
    
    return dialog;
}

async function submitReport() {
    const description = document.getElementById('user-description').value;
    const reportData = JSON.parse(document.getElementById('report-data').textContent);
    
    reportData.userDescription = description;
    
    try {
        showLoading(true);
        
        const response = await fetch('/api/error-report', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(reportData)
        });
        
        if (response.ok) {
            showMessage('报告已发送，感谢您的反馈！', 'success');
            closeReportDialog();
        } else {
            showMessage('发送失败，请稍后再试', 'error');
        }
    } catch (error) {
        showMessage('网络错误，请检查连接后重试', 'error');
    } finally {
        showLoading(false);
    }
}

function closeReportDialog() {
    const dialog = document.querySelector('.report-dialog');
    if (dialog) {
        dialog.classList.remove('show');
        setTimeout(() => dialog.remove(), 300);
    }
}

function showLoading(show) {
    const btn = document.querySelector('.report-actions .btn-primary');
    if (show) {
        btn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> 发送中...';
        btn.disabled = true;
    } else {
        btn.innerHTML = '<i class="fas fa-paper-plane"></i> 发送报告';
        btn.disabled = false;
    }
}

function showMessage(message, type) {
    const messageEl = document.createElement('div');
    messageEl.className = `message message-${type}`;
    messageEl.innerHTML = `
        <i class="fas fa-${type === 'success' ? 'check-circle' : 'exclamation-circle'}"></i>
        <span>${message}</span>
    `;
    
    document.body.appendChild(messageEl);
    
    setTimeout(() => {
        messageEl.classList.add('show');
    }, 100);
    
    setTimeout(() => {
        messageEl.classList.remove('show');
        setTimeout(() => messageEl.remove(), 300);
    }, 3000);
}
```

#### 自动刷新和重试机制

```javascript
// 自动刷新和重试功能
class ErrorPageEnhancer {
    constructor() {
        this.retryCount = 0;
        this.maxRetries = 3;
        this.retryDelay = 5000; // 5秒
        this.autoRefreshEnabled = false;
        
        this.init();
    }
    
    init() {
        this.setupRetryMechanism();
        this.setupAutoRefresh();
        this.setupKeyboardShortcuts();
        this.setupServiceWorker();
    }
    
    setupRetryMechanism() {
        const retryBtn = document.getElementById('retry-btn');
        if (retryBtn) {
            retryBtn.addEventListener('click', () => this.retry());
        }
        
        // 对于网络错误，自动重试
        if (this.isNetworkError()) {
            this.startAutoRetry();
        }
    }
    
    setupAutoRefresh() {
        const toggleBtn = document.getElementById('auto-refresh-toggle');
        if (toggleBtn) {
            toggleBtn.addEventListener('change', (e) => {
                this.autoRefreshEnabled = e.target.checked;
                if (this.autoRefreshEnabled) {
                    this.startAutoRefresh();
                } else {
                    this.stopAutoRefresh();
                }
            });
        }
    }
    
    setupKeyboardShortcuts() {
        document.addEventListener('keydown', (e) => {
            switch (e.key) {
                case 'r':
                case 'R':
                    if (e.ctrlKey || e.metaKey) {
                        e.preventDefault();
                        this.retry();
                    }
                    break;
                case 'h':
                case 'H':
                    if (e.ctrlKey || e.metaKey) {
                        e.preventDefault();
                        window.location.href = '/';
                    }
                    break;
                case 'b':
                case 'B':
                    if (e.ctrlKey || e.metaKey) {
                        e.preventDefault();
                        history.back();
                    }
                    break;
            }
        });
    }
    
    async setupServiceWorker() {
        if ('serviceWorker' in navigator) {
            try {
                await navigator.serviceWorker.register('/sw-error-handler.js');
                this.listenForNetworkChanges();
            } catch (error) {
                console.log('Service Worker registration failed');
            }
        }
    }
    
    retry() {
        this.showRetryProgress();
        
        // 延迟重试，给用户反馈
        setTimeout(() => {
            window.location.reload();
        }, 1000);
    }
    
    async startAutoRetry() {
        if (this.retryCount >= this.maxRetries) {
            this.showMaxRetriesReached();
            return;
        }
        
        this.retryCount++;
        this.showRetryCountdown();
        
        setTimeout(async () => {
            const isOnline = await this.checkConnectivity();
            if (isOnline) {
                this.retry();
            } else {
                this.startAutoRetry();
            }
        }, this.retryDelay);
    }
    
    startAutoRefresh() {
        this.autoRefreshTimer = setInterval(() => {
            this.retry();
        }, 30000); // 每30秒刷新一次
    }
    
    stopAutoRefresh() {
        if (this.autoRefreshTimer) {
            clearInterval(this.autoRefreshTimer);
        }
    }
    
    async checkConnectivity() {
        try {
            const response = await fetch('/api/health', {
                method: 'HEAD',
                cache: 'no-cache'
            });
            return response.ok;
        } catch {
            return false;
        }
    }
    
    isNetworkError() {
        const statusCode = parseInt(
            document.querySelector('.error-code').textContent
        );
        return statusCode >= 500 && statusCode <= 599;
    }
    
    listenForNetworkChanges() {
        window.addEventListener('online', () => {
            this.showMessage('网络连接已恢复', 'success');
            if (this.isNetworkError()) {
                this.retry();
            }
        });
        
        window.addEventListener('offline', () => {
            this.showMessage('网络连接中断', 'warning');
        });
    }
    
    showRetryProgress() {
        const retryBtn = document.getElementById('retry-btn');
        if (retryBtn) {
            retryBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> 重试中...';
            retryBtn.disabled = true;
        }
    }
    
    showRetryCountdown() {
        let countdown = this.retryDelay / 1000;
        const countdownEl = document.createElement('div');
        countdownEl.className = 'retry-countdown';
        countdownEl.innerHTML = `<i class="fas fa-clock"></i> ${countdown}秒后自动重试 (${this.retryCount}/${this.maxRetries})`;
        
        document.querySelector('.error-actions').appendChild(countdownEl);
        
        const timer = setInterval(() => {
            countdown--;
            countdownEl.innerHTML = `<i class="fas fa-clock"></i> ${countdown}秒后自动重试 (${this.retryCount}/${this.maxRetries})`;
            
            if (countdown <= 0) {
                clearInterval(timer);
                countdownEl.remove();
            }
        }, 1000);
    }
    
    showMaxRetriesReached() {
        const message = document.createElement('div');
        message.className = 'max-retries-message';
        message.innerHTML = `
            <i class="fas fa-exclamation-triangle"></i>
            已达到最大重试次数，请检查网络连接或联系技术支持。
        `;
        
        document.querySelector('.error-content').appendChild(message);
    }
    
    showMessage(text, type) {
        // 复用之前定义的showMessage函数
        showMessage(text, type);
    }
}

// 页面加载后初始化增强功能
document.addEventListener('DOMContentLoaded', () => {
    new ErrorPageEnhancer();
});
```

## 🌍 多语言支持

### 语言包系统

```go
// LanguagePack 语言包结构
type LanguagePack struct {
    Code        string                 `json:"code"`         // 语言代码
    Name        string                 `json:"name"`         // 语言名称
    Messages    map[string]string      `json:"messages"`     // 消息文本
    StatusTexts map[int]string         `json:"status_texts"` // 状态码文本
    Suggestions map[int][]string       `json:"suggestions"`  // 建议信息
}

// 预定义语言包
var languagePacks = map[string]*LanguagePack{
    "zh": {
        Code: "zh",
        Name: "中文",
        Messages: map[string]string{
            "back_home":        "返回首页",
            "back_previous":    "返回上页",
            "refresh_page":     "刷新页面",
            "get_help":         "获取帮助",
            "report_problem":   "报告问题",
            "technical_details": "技术详情",
            "suggestions":      "可能的解决方案",
            "contact_us":       "联系我们",
            "retry":           "重试",
            "auto_refresh":     "自动刷新",
        },
        StatusTexts: map[int]string{
            400: "请求错误",
            401: "未授权访问",
            403: "访问被禁止",
            404: "页面未找到",
            500: "服务器内部错误",
            502: "网关错误",
            503: "服务不可用",
        },
        Suggestions: map[int][]string{
            404: {
                "检查URL是否正确",
                "确认页面是否存在",
                "尝试从首页重新导航",
                "清除浏览器缓存后重试",
            },
            500: {
                "稍后重试",
                "检查网络连接",
                "联系技术支持",
                "查看系统状态页面",
            },
        },
    },
    "en": {
        Code: "en",
        Name: "English",
        Messages: map[string]string{
            "back_home":        "Back to Home",
            "back_previous":    "Go Back",
            "refresh_page":     "Refresh Page",
            "get_help":         "Get Help",
            "report_problem":   "Report Problem",
            "technical_details": "Technical Details",
            "suggestions":      "Possible Solutions",
            "contact_us":       "Contact Us",
            "retry":           "Retry",
            "auto_refresh":     "Auto Refresh",
        },
        StatusTexts: map[int]string{
            400: "Bad Request",
            401: "Unauthorized",
            403: "Forbidden",
            404: "Page Not Found",
            500: "Internal Server Error",
            502: "Bad Gateway",
            503: "Service Unavailable",
        },
        Suggestions: map[int][]string{
            404: {
                "Check if the URL is correct",
                "Verify the page exists",
                "Try navigating from home page",
                "Clear browser cache and retry",
            },
            500: {
                "Try again later",
                "Check network connection",
                "Contact technical support",
                "Check system status page",
            },
        },
    },
}

// GetLanguagePack 获取语言包
func GetLanguagePack(code string) *LanguagePack {
    if pack, exists := languagePacks[code]; exists {
        return pack
    }
    return languagePacks["zh"] // 默认中文
}

// 在错误控制器中使用语言包
func (c *DefaultErrorController) generateMultiLanguagePage(ctx *Context, statusCode int, err error) error {
    // 检测用户语言偏好
    lang := c.detectLanguage(ctx)
    pack := GetLanguagePack(lang)
    
    // 构建模板数据
    data := map[string]any{
        "StatusCode":        statusCode,
        "StatusText":        pack.StatusTexts[statusCode],
        "Language":         lang,
        "Messages":         pack.Messages,
        "Suggestions":      pack.Suggestions[statusCode],
        "CustomTitle":      c.CustomTitle,
        "CustomCSS":        c.CustomCSS,
        "CustomJS":         c.CustomJS,
        // ... 其他数据
    }
    
    return c.TemplateEngine.Execute(ctx.Writer, data)
}

func (c *DefaultErrorController) detectLanguage(ctx *Context) string {
    // 1. URL参数优先级最高
    if lang := ctx.Query("lang"); lang != "" {
        return lang
    }
    
    // 2. Cookie中的语言设置
    if lang := ctx.Cookie("language"); lang != "" {
        return lang
    }
    
    // 3. Accept-Language头
    acceptLang := ctx.GetHeader("Accept-Language")
    if strings.HasPrefix(acceptLang, "en") {
        return "en"
    }
    
    // 4. 默认中文
    return "zh"
}
```

## 📱 响应式设计优化

### 移动端适配

```css
/* 移动端特殊优化 */
@media (max-width: 480px) {
    .error-container {
        margin: 5px;
        border-radius: 8px;
        box-shadow: 0 10px 20px rgba(0, 0, 0, 0.1);
    }
    
    .error-header {
        padding: 20px 15px;
    }
    
    .error-code {
        font-size: 3rem;
        line-height: 1;
    }
    
    .error-title {
        font-size: 1.2rem;
    }
    
    .error-body {
        padding: 20px 15px;
    }
    
    .error-message {
        font-size: 1rem;
        padding: 15px;
    }
    
    .btn {
        padding: 10px 16px;
        font-size: 0.9rem;
    }
    
    .error-actions {
        gap: 10px;
    }
    
    .detail-item {
        font-size: 0.85rem;
        padding: 6px 0;
    }
    
    .detail-label {
        min-width: 80px;
        font-size: 0.8rem;
    }
    
    .suggestions {
        margin-top: 15px;
    }
    
    .suggestion-list li {
        font-size: 0.9rem;
        line-height: 1.4;
    }
    
    .report-dialog .report-content {
        margin: 10px;
        max-height: calc(100vh - 20px);
        border-radius: 8px;
    }
    
    .report-body {
        padding: 15px;
    }
    
    #user-description {
        font-size: 16px; /* 防止iOS缩放 */
    }
}

/* 平板适配 */
@media (min-width: 481px) and (max-width: 768px) {
    .error-container {
        max-width: 600px;
    }
    
    .error-actions {
        justify-content: center;
    }
    
    .btn {
        min-width: 140px;
    }
}

/* 超小屏设备 */
@media (max-width: 320px) {
    .error-code {
        font-size: 2.5rem;
    }
    
    .error-title {
        font-size: 1.1rem;
    }
    
    .btn {
        padding: 8px 12px;
        font-size: 0.85rem;
    }
}
```

### 触摸优化

```css
/* 触摸优化 */
.btn, .theme-toggle, .close-btn, summary {
    touch-action: manipulation;
    min-height: 44px; /* iOS建议的最小触摸目标 */
    min-width: 44px;
}

/* 提高触摸目标的可见性 */
@media (hover: none) {
    .btn:hover {
        transform: none;
        box-shadow: 0 2px 8px rgba(102, 126, 234, 0.3);
    }
    
    .btn:active {
        transform: scale(0.98);
    }
}

/* 改善滚动性能 */
.error-container {
    -webkit-overflow-scrolling: touch;
    will-change: scroll-position;
}

/* 优化文本选择 */
.error-code, .error-title {
    -webkit-user-select: none;
    user-select: none;
}

.detail-value, .stack-trace {
    -webkit-user-select: text;
    user-select: text;
}
```

## 🚀 部署和优化

### 生产环境配置

```go
func setupProductionErrorPages() {
    controller := NewBrandedErrorController(&BrandConfig{
        CompanyName: "YYHertz Technology",
        SupportURL:  "https://support.example.com",
        ContactInfo: "support@example.com",
        BrandLogo:   "/assets/logo.svg",
    })
    
    // 生产环境配置
    controller.Config = &ErrorPageConfig{
        ShowStackTrace:    false,    // 生产环境不显示堆栈
        ShowRequestInfo:   false,    // 不显示详细请求信息
        ShowTimestamp:     true,     // 显示时间戳便于排查
        Theme:            "light",   // 默认浅色主题
        Language:         "zh",      // 默认中文
        EnableReporting:  true,      // 启用错误报告
        ContactInfo:      "400-888-8888",
        CompanyName:      "YYHertz Technology",
        SupportURL:       "https://support.example.com",
    }
    
    // 注册错误处理器
    errors.Register(404, controller)
    errors.Register(500, controller)
}
```

### 性能优化

```go
// 缓存错误页面模板
var templateCache sync.Map

func (c *DefaultErrorController) getCachedTemplate(templateKey string) (*template.Template, error) {
    if tmpl, ok := templateCache.Load(templateKey); ok {
        return tmpl.(*template.Template), nil
    }
    
    // 编译和缓存模板
    tmpl, err := c.compileTemplate()
    if err != nil {
        return nil, err
    }
    
    templateCache.Store(templateKey, tmpl)
    return tmpl, nil
}

// 启用Gzip压缩
func (c *DefaultErrorController) Handle(ctx *Context, statusCode int, err error) error {
    // 设置压缩头
    ctx.Header("Content-Encoding", "gzip")
    ctx.Header("Cache-Control", "public, max-age=300") // 缓存5分钟
    
    // 压缩响应
    return c.renderCompressed(ctx, statusCode, err)
}

// CDN和静态资源优化
func optimizeStaticAssets() {
    // 使用CDN加载外部资源
    cdnResources := map[string]string{
        "bootstrap":   "https://cdn.bootcdn.net/ajax/libs/bootstrap/5.3.0/css/bootstrap.min.css",
        "fontawesome": "https://cdn.bootcdn.net/ajax/libs/font-awesome/6.4.0/css/all.min.css",
        "jquery":      "https://cdn.bootcdn.net/ajax/libs/jquery/3.7.0/jquery.min.js",
    }
    
    // 预加载关键资源
    for name, url := range cdnResources {
        fmt.Printf(`<link rel="preload" href="%s" as="%s" crossorigin>`, url, getResourceType(name))
    }
}
```

## 📚 相关文档

- **[快速开始](quick-start.md)** - 了解错误处理系统基础用法
- **[默认处理器详解](default-handlers.md)** - 深入理解默认错误处理器
- **[自定义处理器](custom-handlers.md)** - 开发自定义错误处理逻辑
- **[错误监控](monitoring.md)** - 建立完善的监控体系

---

> 💡 **提示**: 良好的错误页面设计能显著提升用户体验。建议根据品牌特色定制页面样式，并充分测试在不同设备和浏览器上的显示效果。