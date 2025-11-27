/**
 * YYHertz Framework - Markdown Theme Switcher
 * 提供多种Markdown样式主题的切换功能
 */

class MarkdownThemeSwitcher {
    constructor() {
        this.themes = {
            'default': {
                name: '默认主题',
                file: '/static/css/markdown.css',
                description: '基础的Markdown样式'
            },
            'github': {
                name: 'GitHub风格',
                file: '/static/css/markdown-github.css',
                description: '仿GitHub的Markdown样式'
            },
            'juejin': {
                name: '掘金风格',
                file: '/static/css/markdown-juejin.css',
                description: '仿掘金社区的Markdown样式'
            },
            'docsify': {
                name: 'Docsify风格',
                file: '/static/css/markdown-docsify.css',
                description: '仿Docsify文档系统的样式'
            },
            'doclever': {
                name: 'DocLever风格',
                file: '/static/css/markdown-doclever.css',
                description: '仿DocLever接口文档的样式'
            },
            'hexo': {
                name: 'Hexo风格',
                file: '/static/css/markdown-hexo.css',
                description: '仿Hexo博客系统的样式'
            },
            'gitlab': {
                name: 'GitLab风格',
                file: '/static/css/markdown-gitlab.css',
                description: '仿GitLab平台的Markdown样式'
            },
            'bootstrap': {
                name: 'Bootstrap风格',
                file: '/static/css/markdown-bootstrap.css',
                description: '基于Bootstrap设计系统的样式'
            },
            'darkreader': {
                name: 'DarkReader暗色主题',
                file: '/static/css/markdown-darkreader.css',
                description: '暗色主题，保护眼睛'
            }
        };
        
        this.currentTheme = localStorage.getItem('markdown-theme') || 'default';
        this.linkElement = null;
        
        this.init();
    }
    
    init() {
        this.createStyleLink();
        this.loadTheme(this.currentTheme);
        this.createSwitcher();
    }
    
    createStyleLink() {
        // 创建或获取样式链接元素
        this.linkElement = document.getElementById('markdown-theme-link');
        if (!this.linkElement) {
            this.linkElement = document.createElement('link');
            this.linkElement.id = 'markdown-theme-link';
            this.linkElement.rel = 'stylesheet';
            this.linkElement.type = 'text/css';
            document.head.appendChild(this.linkElement);
        }
    }
    
    loadTheme(themeKey) {
        if (!this.themes[themeKey]) {
            console.warn(`主题 "${themeKey}" 不存在，使用默认主题`);
            themeKey = 'default';
        }
        
        const theme = this.themes[themeKey];
        this.linkElement.href = theme.file;
        this.currentTheme = themeKey;
        
        // 保存到本地存储
        localStorage.setItem('markdown-theme', themeKey);
        
        // 触发主题切换事件
        this.dispatchThemeChangeEvent(themeKey, theme);
        
        // 更新切换器UI
        this.updateSwitcherUI();
    }
    
    createSwitcher() {
        // 创建主题切换器容器
        const switcherContainer = document.createElement('div');
        switcherContainer.className = 'markdown-theme-switcher';
        switcherContainer.innerHTML = `
            <div class="theme-switcher-btn" id="theme-switcher-btn">
                <i class="fas fa-palette"></i>
                <span class="theme-name">${this.themes[this.currentTheme].name}</span>
                <i class="fas fa-chevron-down"></i>
            </div>
            <div class="theme-switcher-dropdown" id="theme-switcher-dropdown">
                ${this.createThemeOptions()}
            </div>
        `;
        
        // 添加样式
        this.addSwitcherStyles();
        
        // 添加到页面
        document.body.appendChild(switcherContainer);
        
        // 绑定事件
        this.bindEvents();
    }
    
    createThemeOptions() {
        return Object.entries(this.themes).map(([key, theme]) => `
            <div class="theme-option ${key === this.currentTheme ? 'active' : ''}" 
                 data-theme="${key}" 
                 title="${theme.description}">
                <span class="theme-preview"></span>
                <div class="theme-info">
                    <span class="theme-name">${theme.name}</span>
                    <span class="theme-description">${theme.description}</span>
                </div>
                ${key === this.currentTheme ? '<i class="fas fa-check"></i>' : ''}
            </div>
        `).join('');
    }
    
    addSwitcherStyles() {
        if (document.getElementById('theme-switcher-styles')) return;
        
        const styles = document.createElement('style');
        styles.id = 'theme-switcher-styles';
        styles.textContent = `
            .markdown-theme-switcher {
                position: fixed;
                top: 20px;
                right: 20px;
                z-index: 1000;
                font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            }
            
            .theme-switcher-btn {
                background: #fff;
                border: 1px solid #ddd;
                border-radius: 8px;
                padding: 10px 15px;
                cursor: pointer;
                display: flex;
                align-items: center;
                gap: 8px;
                box-shadow: 0 2px 8px rgba(0,0,0,0.1);
                transition: all 0.3s ease;
                min-width: 160px;
            }
            
            .theme-switcher-btn:hover {
                box-shadow: 0 4px 12px rgba(0,0,0,0.15);
                transform: translateY(-1px);
            }
            
            .theme-switcher-btn .fas:first-child {
                color: #667eea;
            }
            
            .theme-switcher-btn .theme-name {
                flex: 1;
                font-size: 13px;
                font-weight: 500;
            }
            
            .theme-switcher-btn .fa-chevron-down {
                font-size: 10px;
                color: #999;
                transition: transform 0.3s ease;
            }
            
            .theme-switcher-dropdown {
                position: absolute;
                top: 100%;
                left: 0;
                right: 0;
                background: #fff;
                border: 1px solid #ddd;
                border-radius: 8px;
                box-shadow: 0 4px 20px rgba(0,0,0,0.15);
                max-height: 400px;
                overflow-y: auto;
                opacity: 0;
                visibility: hidden;
                transform: translateY(-10px);
                transition: all 0.3s ease;
                margin-top: 5px;
            }
            
            .theme-switcher-dropdown.show {
                opacity: 1;
                visibility: visible;
                transform: translateY(0);
            }
            
            .theme-option {
                padding: 12px 15px;
                cursor: pointer;
                display: flex;
                align-items: center;
                gap: 12px;
                border-bottom: 1px solid #f0f0f0;
                transition: background-color 0.2s ease;
            }
            
            .theme-option:last-child {
                border-bottom: none;
            }
            
            .theme-option:hover {
                background-color: #f8f9fa;
            }
            
            .theme-option.active {
                background-color: #e7f3ff;
                border-left: 3px solid #667eea;
            }
            
            .theme-preview {
                width: 20px;
                height: 15px;
                border-radius: 3px;
                border: 1px solid #ddd;
                flex-shrink: 0;
            }
            
            .theme-option[data-theme="github"] .theme-preview {
                background: linear-gradient(45deg, #f6f8fa 50%, #0969da 50%);
            }
            
            .theme-option[data-theme="juejin"] .theme-preview {
                background: linear-gradient(45deg, #1e80ff 50%, #fff 50%);
            }
            
            .theme-option[data-theme="docsify"] .theme-preview {
                background: linear-gradient(45deg, #42b983 50%, #f8f8f8 50%);
            }
            
            .theme-option[data-theme="doclever"] .theme-preview {
                background: linear-gradient(45deg, #3498db 50%, #fff 50%);
            }
            
            .theme-option[data-theme="hexo"] .theme-preview {
                background: linear-gradient(45deg, #667eea 50%, #764ba2 50%);
            }
            
            .theme-option[data-theme="gitlab"] .theme-preview {
                background: linear-gradient(45deg, #fc6d26 50%, #f6f8fa 50%);
            }
            
            .theme-option[data-theme="bootstrap"] .theme-preview {
                background: linear-gradient(45deg, #0d6efd 50%, #fff 50%);
            }
            
            .theme-option[data-theme="darkreader"] .theme-preview {
                background: linear-gradient(45deg, #181a1b 50%, #4285f4 50%);
            }
            
            .theme-option[data-theme="default"] .theme-preview {
                background: linear-gradient(45deg, #667eea 50%, #f8f9fa 50%);
            }
            
            .theme-info {
                flex: 1;
                display: flex;
                flex-direction: column;
                gap: 2px;
            }
            
            .theme-info .theme-name {
                font-size: 13px;
                font-weight: 500;
                color: #333;
            }
            
            .theme-info .theme-description {
                font-size: 11px;
                color: #666;
            }
            
            .theme-option .fa-check {
                color: #667eea;
                font-size: 12px;
            }
            
            .markdown-theme-switcher.open .fa-chevron-down {
                transform: rotate(180deg);
            }
            
            @media (max-width: 768px) {
                .markdown-theme-switcher {
                    top: 10px;
                    right: 10px;
                }
                
                .theme-switcher-btn {
                    padding: 8px 12px;
                    min-width: 140px;
                }
                
                .theme-option {
                    padding: 10px 12px;
                }
            }
        `;
        
        document.head.appendChild(styles);
    }
    
    bindEvents() {
        const btn = document.getElementById('theme-switcher-btn');
        const dropdown = document.getElementById('theme-switcher-dropdown');
        const container = btn.parentElement;
        
        // 切换下拉菜单
        btn.addEventListener('click', (e) => {
            e.stopPropagation();
            container.classList.toggle('open');
            dropdown.classList.toggle('show');
        });
        
        // 点击外部关闭下拉菜单
        document.addEventListener('click', () => {
            container.classList.remove('open');
            dropdown.classList.remove('show');
        });
        
        // 主题选择
        dropdown.addEventListener('click', (e) => {
            const option = e.target.closest('.theme-option');
            if (option) {
                const themeKey = option.dataset.theme;
                this.loadTheme(themeKey);
                container.classList.remove('open');
                dropdown.classList.remove('show');
            }
        });
        
        // 键盘支持
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') {
                container.classList.remove('open');
                dropdown.classList.remove('show');
            }
        });
    }
    
    updateSwitcherUI() {
        const btn = document.getElementById('theme-switcher-btn');
        const dropdown = document.getElementById('theme-switcher-dropdown');
        
        if (btn) {
            const themeName = btn.querySelector('.theme-name');
            if (themeName) {
                themeName.textContent = this.themes[this.currentTheme].name;
            }
        }
        
        if (dropdown) {
            dropdown.innerHTML = this.createThemeOptions();
        }
    }
    
    dispatchThemeChangeEvent(themeKey, theme) {
        const event = new CustomEvent('themeChanged', {
            detail: { themeKey, theme }
        });
        document.dispatchEvent(event);
    }
    
    // 公共API
    switchTheme(themeKey) {
        this.loadTheme(themeKey);
    }
    
    getCurrentTheme() {
        return this.currentTheme;
    }
    
    getAvailableThemes() {
        return this.themes;
    }
}

// 自动初始化
document.addEventListener('DOMContentLoaded', () => {
    if (document.querySelector('.markdown-content')) {
        window.markdownThemeSwitcher = new MarkdownThemeSwitcher();
        
        // 添加主题切换事件监听器
        document.addEventListener('themeChanged', (e) => {
            console.log(`主题已切换为: ${e.detail.theme.name}`);
        });
    }
});

// 导出类供外部使用
if (typeof module !== 'undefined' && module.exports) {
    module.exports = MarkdownThemeSwitcher;
} else if (typeof window !== 'undefined') {
    window.MarkdownThemeSwitcher = MarkdownThemeSwitcher;
}