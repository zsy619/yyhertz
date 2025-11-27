// YYHertz 模板引擎演示应用 JavaScript

// 工具函数
const Utils = {
    // 格式化日期
    formatDate: function(date, format = 'YYYY-MM-DD HH:mm:ss') {
        const d = new Date(date);
        const year = d.getFullYear();
        const month = String(d.getMonth() + 1).padStart(2, '0');
        const day = String(d.getDate()).padStart(2, '0');
        const hour = String(d.getHours()).padStart(2, '0');
        const minute = String(d.getMinutes()).padStart(2, '0');
        const second = String(d.getSeconds()).padStart(2, '0');
        
        return format
            .replace('YYYY', year)
            .replace('MM', month)
            .replace('DD', day)
            .replace('HH', hour)
            .replace('mm', minute)
            .replace('ss', second);
    },
    
    // 格式化文件大小
    formatFileSize: function(bytes) {
        if (bytes === 0) return '0 B';
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    },
    
    // 复制到剪贴板
    copyToClipboard: function(text) {
        if (navigator.clipboard) {
            return navigator.clipboard.writeText(text);
        } else {
            // 兼容旧浏览器
            const textArea = document.createElement('textarea');
            textArea.value = text;
            document.body.appendChild(textArea);
            textArea.select();
            document.execCommand('copy');
            document.body.removeChild(textArea);
            return Promise.resolve();
        }
    },
    
    // 显示提示消息
    showMessage: function(message, type = 'info', duration = 3000) {
        const messageContainer = document.getElementById('messageContainer') || this.createMessageContainer();
        
        const messageElement = document.createElement('div');
        messageElement.className = `alert alert-${type} message-item`;
        messageElement.innerHTML = `
            <span>${message}</span>
            <button type="button" class="message-close" onclick="this.parentElement.remove()">×</button>
        `;
        
        messageContainer.appendChild(messageElement);
        
        // 自动移除
        if (duration > 0) {
            setTimeout(() => {
                if (messageElement.parentNode) {
                    messageElement.remove();
                }
            }, duration);
        }
    },
    
    // 创建消息容器
    createMessageContainer: function() {
        const container = document.createElement('div');
        container.id = 'messageContainer';
        container.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            z-index: 9999;
            max-width: 400px;
        `;
        document.body.appendChild(container);
        return container;
    },
    
    // AJAX 请求封装
    request: function(url, options = {}) {
        const defaultOptions = {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json'
            }
        };
        
        const config = Object.assign(defaultOptions, options);
        
        return fetch(url, config)
            .then(response => {
                if (!response.ok) {
                    throw new Error(`HTTP error! status: ${response.status}`);
                }
                return response.json();
            });
    }
};

// CSRF 处理
const CSRF = {
    token: null,
    
    // 初始化 CSRF token
    init: function() {
        const metaToken = document.querySelector('meta[name="csrf-token"]');
        if (metaToken) {
            this.token = metaToken.getAttribute('content');
        }
    },
    
    // 获取 CSRF token
    getToken: function() {
        return this.token;
    },
    
    // 设置 CSRF token
    setToken: function(token) {
        this.token = token;
    },
    
    // 为表单添加 CSRF token
    addToForm: function(form) {
        if (!this.token) return;
        
        let tokenInput = form.querySelector('input[name="csrf_token"]');
        if (!tokenInput) {
            tokenInput = document.createElement('input');
            tokenInput.type = 'hidden';
            tokenInput.name = 'csrf_token';
            form.appendChild(tokenInput);
        }
        tokenInput.value = this.token;
    },
    
    // 为 AJAX 请求添加 CSRF token
    addToRequest: function(options) {
        if (!this.token) return options;
        
        if (options.method && options.method.toUpperCase() !== 'GET') {
            if (options.body instanceof FormData) {
                options.body.append('csrf_token', this.token);
            } else if (typeof options.body === 'object') {
                options.body.csrf_token = this.token;
            }
        }
        
        return options;
    }
};

// 模板预览功能
const TemplatePreview = {
    // 预览模板
    preview: function(templateName, content) {
        const formData = new FormData();
        formData.append('template', templateName);
        formData.append('content', content || '');
        
        const options = CSRF.addToRequest({
            method: 'POST',
            body: formData
        });
        
        return Utils.request('/templates/preview', options)
            .then(data => {
                if (data.status === 'success') {
                    this.showPreviewWindow(templateName, data.result);
                } else {
                    throw new Error(data.error || '预览失败');
                }
            })
            .catch(error => {
                Utils.showMessage('预览失败: ' + error.message, 'danger');
            });
    },
    
    // 显示预览窗口
    showPreviewWindow: function(templateName, content) {
        const previewWindow = window.open('', '_blank', 'width=800,height=600,scrollbars=yes');
        
        previewWindow.document.write(`
            <!DOCTYPE html>
            <html lang="zh-CN">
            <head>
                <meta charset="UTF-8">
                <meta name="viewport" content="width=device-width, initial-scale=1.0">
                <title>模板预览 - ${templateName}</title>
                <link rel="stylesheet" href="/static/css/app.css">
                <style>
                    .preview-header {
                        background: #f8f9fa;
                        padding: 15px;
                        border: 1px solid #dee2e6;
                        margin-bottom: 20px;
                        border-radius: 8px;
                    }
                    .preview-content {
                        padding: 20px;
                    }
                </style>
            </head>
            <body>
                <div class="container">
                    <div class="preview-header">
                        <h3>📄 模板预览: ${templateName}</h3>
                        <p class="text-muted">预览时间: ${Utils.formatDate(new Date())}</p>
                        <button class="btn btn-secondary btn-sm" onclick="window.close()">关闭预览</button>
                    </div>
                    <div class="preview-content">
                        ${content}
                    </div>
                </div>
            </body>
            </html>
        `);
        
        previewWindow.document.close();
    }
};

// 性能监控
const Performance = {
    startTime: Date.now(),
    
    // 记录性能指标
    mark: function(name) {
        if (window.performance && window.performance.mark) {
            window.performance.mark(name);
        }
    },
    
    // 测量性能
    measure: function(name, startMark, endMark) {
        if (window.performance && window.performance.measure) {
            try {
                window.performance.measure(name, startMark, endMark);
                const measure = window.performance.getEntriesByName(name)[0];
                return measure.duration;
            } catch (e) {
                console.warn('Performance measurement failed:', e);
            }
        }
        return null;
    },
    
    // 获取页面加载时间
    getLoadTime: function() {
        return Date.now() - this.startTime;
    },
    
    // 监控资源加载
    monitorResources: function() {
        if (window.performance && window.performance.getEntriesByType) {
            const resources = window.performance.getEntriesByType('resource');
            return resources.map(resource => ({
                name: resource.name,
                duration: resource.duration,
                size: resource.transferSize
            }));
        }
        return [];
    }
};

// 主题切换
const Theme = {
    current: 'light',
    
    // 初始化主题
    init: function() {
        const savedTheme = localStorage.getItem('theme') || 'light';
        this.apply(savedTheme);
    },
    
    // 应用主题
    apply: function(theme) {
        this.current = theme;
        document.documentElement.setAttribute('data-theme', theme);
        localStorage.setItem('theme', theme);
        
        // 更新主题切换按钮状态
        const themeToggle = document.getElementById('themeToggle');
        if (themeToggle) {
            themeToggle.textContent = theme === 'light' ? '🌙' : '☀️';
            themeToggle.title = theme === 'light' ? '切换到深色模式' : '切换到浅色模式';
        }
    },
    
    // 切换主题
    toggle: function() {
        const newTheme = this.current === 'light' ? 'dark' : 'light';
        this.apply(newTheme);
    }
};

// 搜索功能
const Search = {
    // 过滤元素
    filter: function(searchTerm, elements, searchAttribute = 'data-search') {
        const term = searchTerm.toLowerCase();
        
        elements.forEach(element => {
            const searchText = element.getAttribute(searchAttribute) || element.textContent || '';
            const matches = searchText.toLowerCase().includes(term);
            
            element.style.display = matches ? '' : 'none';
            
            // 添加高亮
            if (matches && term) {
                this.highlight(element, term);
            } else {
                this.removeHighlight(element);
            }
        });
    },
    
    // 高亮文本
    highlight: function(element, term) {
        // 简单的高亮实现
        const regex = new RegExp(`(${term})`, 'gi');
        const textNodes = this.getTextNodes(element);
        
        textNodes.forEach(node => {
            if (node.textContent.toLowerCase().includes(term.toLowerCase())) {
                const highlightedText = node.textContent.replace(regex, '<mark>$1</mark>');
                const span = document.createElement('span');
                span.innerHTML = highlightedText;
                node.parentNode.replaceChild(span, node);
            }
        });
    },
    
    // 移除高亮
    removeHighlight: function(element) {
        const marks = element.querySelectorAll('mark');
        marks.forEach(mark => {
            mark.outerHTML = mark.innerHTML;
        });
    },
    
    // 获取文本节点
    getTextNodes: function(element) {
        const textNodes = [];
        const walker = document.createTreeWalker(
            element,
            NodeFilter.SHOW_TEXT,
            null,
            false
        );
        
        let node;
        while (node = walker.nextNode()) {
            if (node.textContent.trim()) {
                textNodes.push(node);
            }
        }
        
        return textNodes;
    }
};

// 动画工具
const Animation = {
    // 淡入效果
    fadeIn: function(element, duration = 300) {
        element.style.opacity = '0';
        element.style.display = 'block';
        
        const start = performance.now();
        
        const animate = (timestamp) => {
            const elapsed = timestamp - start;
            const progress = Math.min(elapsed / duration, 1);
            
            element.style.opacity = progress.toString();
            
            if (progress < 1) {
                requestAnimationFrame(animate);
            }
        };
        
        requestAnimationFrame(animate);
    },
    
    // 淡出效果
    fadeOut: function(element, duration = 300) {
        const start = performance.now();
        const startOpacity = parseFloat(window.getComputedStyle(element).opacity);
        
        const animate = (timestamp) => {
            const elapsed = timestamp - start;
            const progress = Math.min(elapsed / duration, 1);
            
            element.style.opacity = (startOpacity * (1 - progress)).toString();
            
            if (progress >= 1) {
                element.style.display = 'none';
            } else {
                requestAnimationFrame(animate);
            }
        };
        
        requestAnimationFrame(animate);
    },
    
    // 滑动显示
    slideDown: function(element, duration = 300) {
        element.style.height = '0';
        element.style.overflow = 'hidden';
        element.style.display = 'block';
        
        const targetHeight = element.scrollHeight;
        const start = performance.now();
        
        const animate = (timestamp) => {
            const elapsed = timestamp - start;
            const progress = Math.min(elapsed / duration, 1);
            
            element.style.height = (targetHeight * progress) + 'px';
            
            if (progress >= 1) {
                element.style.height = 'auto';
                element.style.overflow = 'visible';
            } else {
                requestAnimationFrame(animate);
            }
        };
        
        requestAnimationFrame(animate);
    }
};

// 页面初始化
document.addEventListener('DOMContentLoaded', function() {
    // 初始化模块
    CSRF.init();
    Theme.init();
    Performance.mark('app-init');
    
    // 性能监控
    console.log('Page load time:', Performance.getLoadTime() + 'ms');
    
    // 绑定全局事件
    bindGlobalEvents();
    
    // 标记应用初始化完成
    Performance.mark('app-ready');
});

// 绑定全局事件
function bindGlobalEvents() {
    // 主题切换按钮
    const themeToggle = document.getElementById('themeToggle');
    if (themeToggle) {
        themeToggle.addEventListener('click', () => Theme.toggle());
    }
    
    // 返回顶部按钮
    const backToTop = document.getElementById('backToTop');
    if (backToTop) {
        window.addEventListener('scroll', () => {
            if (window.pageYOffset > 300) {
                backToTop.style.display = 'block';
            } else {
                backToTop.style.display = 'none';
            }
        });
        
        backToTop.addEventListener('click', () => {
            window.scrollTo({ top: 0, behavior: 'smooth' });
        });
    }
    
    // 为所有表单自动添加 CSRF token
    document.querySelectorAll('form').forEach(form => {
        CSRF.addToForm(form);
    });
    
    // 自动绑定复制按钮
    document.querySelectorAll('[data-copy]').forEach(button => {
        button.addEventListener('click', function() {
            const target = document.querySelector(this.dataset.copy);
            if (target) {
                Utils.copyToClipboard(target.textContent)
                    .then(() => Utils.showMessage('已复制到剪贴板', 'success'))
                    .catch(() => Utils.showMessage('复制失败', 'danger'));
            }
        });
    });
}

// 导出全局对象
window.YYHertz = {
    Utils,
    CSRF,
    TemplatePreview,
    Performance,
    Theme,
    Search,
    Animation
};