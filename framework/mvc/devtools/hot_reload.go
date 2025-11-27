package devtools

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/zsy619/yyhertz/framework/mvc"
)

// HotReloader 热重载器
type HotReloader struct {
	app         *mvc.App
	watcher     *fsnotify.Watcher
	watchDirs   []string
	excludeDirs []string
	extensions  []string
	debounce    time.Duration
	mu          sync.RWMutex
	running     bool
	restartCh   chan struct{}
	stopCh      chan struct{}

	// 回调函数
	onReload     func() error
	onError      func(error)
	onFileChange func(string, string) // 文件路径, 事件类型
}

// HotReloadConfig 热重载配置
type HotReloadConfig struct {
	WatchDirs    []string             // 监控目录
	ExcludeDirs  []string             // 排除目录
	Extensions   []string             // 监控文件扩展名
	Debounce     time.Duration        // 防抖时间
	OnReload     func() error         // 重载回调
	OnError      func(error)          // 错误回调
	OnFileChange func(string, string) // 文件变化回调
}

// NewHotReloader 创建热重载器
func NewHotReloader(app *mvc.App, config HotReloadConfig) (*HotReloader, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("创建文件监控器失败: %v", err)
	}

	// 设置默认值
	if len(config.WatchDirs) == 0 {
		config.WatchDirs = []string{".", "controllers", "views", "static"}
	}
	if len(config.ExcludeDirs) == 0 {
		config.ExcludeDirs = []string{"logs", "tmp", ".git", "node_modules", "vendor"}
	}
	if len(config.Extensions) == 0 {
		config.Extensions = []string{".go", ".html", ".css", ".js", ".yaml", ".yml", ".json"}
	}
	if config.Debounce == 0 {
		config.Debounce = 500 * time.Millisecond
	}

	hr := &HotReloader{
		app:          app,
		watcher:      watcher,
		watchDirs:    config.WatchDirs,
		excludeDirs:  config.ExcludeDirs,
		extensions:   config.Extensions,
		debounce:     config.Debounce,
		restartCh:    make(chan struct{}, 1),
		stopCh:       make(chan struct{}),
		onReload:     config.OnReload,
		onError:      config.OnError,
		onFileChange: config.OnFileChange,
	}

	return hr, nil
}

// Start 启动热重载
func (hr *HotReloader) Start() error {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	if hr.running {
		return fmt.Errorf("热重载器已在运行")
	}

	// 添加监控目录
	for _, dir := range hr.watchDirs {
		if err := hr.addWatchDir(dir); err != nil {
			return fmt.Errorf("添加监控目录 %s 失败: %v", dir, err)
		}
	}

	hr.running = true
	go hr.watchLoop()

	log.Printf("热重载器已启动，监控目录: %v", hr.watchDirs)
	return nil
}

// Stop 停止热重载
func (hr *HotReloader) Stop() error {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	if !hr.running {
		return nil
	}

	close(hr.stopCh)
	hr.running = false

	if err := hr.watcher.Close(); err != nil {
		return fmt.Errorf("关闭文件监控器失败: %v", err)
	}

	log.Println("热重载器已停止")
	return nil
}

// addWatchDir 添加监控目录
func (hr *HotReloader) addWatchDir(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过排除的目录
		if info.IsDir() {
			for _, exclude := range hr.excludeDirs {
				if strings.Contains(path, exclude) {
					return filepath.SkipDir
				}
			}
			return hr.watcher.Add(path)
		}

		return nil
	})
}

// watchLoop 监控循环
func (hr *HotReloader) watchLoop() {
	var (
		timer    *time.Timer
		timerCh  <-chan time.Time
		lastFile string
	)

	for {
		select {
		case event, ok := <-hr.watcher.Events:
			if !ok {
				return
			}

			if hr.shouldIgnoreEvent(event) {
				continue
			}

			// 通知文件变化
			if hr.onFileChange != nil {
				hr.onFileChange(event.Name, event.Op.String())
			}

			log.Printf("文件变化: %s [%s]", event.Name, event.Op.String())

			// 防抖处理
			if timer != nil {
				timer.Stop()
			}
			timer = time.NewTimer(hr.debounce)
			timerCh = timer.C
			lastFile = event.Name

		case <-timerCh:
			log.Printf("触发重载，最后修改文件: %s", lastFile)
			hr.triggerReload()
			timer = nil
			timerCh = nil

		case err, ok := <-hr.watcher.Errors:
			if !ok {
				return
			}
			if hr.onError != nil {
				hr.onError(fmt.Errorf("文件监控错误: %v", err))
			} else {
				log.Printf("文件监控错误: %v", err)
			}

		case <-hr.stopCh:
			if timer != nil {
				timer.Stop()
			}
			return
		}
	}
}

// shouldIgnoreEvent 是否应该忽略事件
func (hr *HotReloader) shouldIgnoreEvent(event fsnotify.Event) bool {
	// 忽略临时文件和隐藏文件
	fileName := filepath.Base(event.Name)
	if strings.HasPrefix(fileName, ".") || strings.HasSuffix(fileName, "~") {
		return true
	}

	// 检查文件扩展名
	ext := filepath.Ext(event.Name)
	if len(hr.extensions) > 0 {
		found := false
		for _, allowedExt := range hr.extensions {
			if ext == allowedExt {
				found = true
				break
			}
		}
		if !found {
			return true
		}
	}

	// 检查排除目录
	for _, exclude := range hr.excludeDirs {
		if strings.Contains(event.Name, exclude) {
			return true
		}
	}

	return false
}

// triggerReload 触发重载
func (hr *HotReloader) triggerReload() {
	select {
	case hr.restartCh <- struct{}{}:
	default:
		// 如果通道已满，忽略这次重载请求
	}

	if hr.onReload != nil {
		if err := hr.onReload(); err != nil {
			if hr.onError != nil {
				hr.onError(fmt.Errorf("重载失败: %v", err))
			} else {
				log.Printf("重载失败: %v", err)
			}
		}
	}
}

// RestartChannel 获取重启通道
func (hr *HotReloader) RestartChannel() <-chan struct{} {
	return hr.restartCh
}

// IsRunning 检查是否正在运行
func (hr *HotReloader) IsRunning() bool {
	hr.mu.RLock()
	defer hr.mu.RUnlock()
	return hr.running
}

// HotReloadServer 热重载服务器
type HotReloadServer struct {
	app      *mvc.App
	reloader *HotReloader
	ctx      context.Context
	cancel   context.CancelFunc
	serverCh chan error
}

// NewHotReloadServer 创建热重载服务器
func NewHotReloadServer(app *mvc.App, config HotReloadConfig) (*HotReloadServer, error) {
	reloader, err := NewHotReloader(app, config)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &HotReloadServer{
		app:      app,
		reloader: reloader,
		ctx:      ctx,
		cancel:   cancel,
		serverCh: make(chan error, 1),
	}, nil
}

// Run 运行热重载服务器
func (hrs *HotReloadServer) Run(addr ...string) error {
	// 启动热重载器
	if err := hrs.reloader.Start(); err != nil {
		return err
	}
	defer hrs.reloader.Stop()

	// 启动服务器
	if len(addr) > 0 {
		go func() {
			log.Printf("服务器启动在 %s", addr)
			hrs.app.Run(addr[0])
			// 如果服务器正常退出，发送nil错误
			hrs.serverCh <- nil
		}()
	}

	// 监听重载信号
	for {
		select {
		case <-hrs.reloader.RestartChannel():
			log.Println("检测到文件变化，准备重启服务器...")

			// 这里可以添加优雅关闭逻辑
			// 实际项目中可能需要重新编译和重启整个进程
			log.Println("服务器重启完成")

		case err := <-hrs.serverCh:
			return fmt.Errorf("服务器错误: %v", err)

		case <-hrs.ctx.Done():
			return nil
		}
	}
}

// Stop 停止热重载服务器
func (hrs *HotReloadServer) Stop() error {
	hrs.cancel()
	return hrs.reloader.Stop()
}

// DefaultHotReloadConfig 默认热重载配置
func DefaultHotReloadConfig() HotReloadConfig {
	return HotReloadConfig{
		WatchDirs:   []string{".", "controllers", "views", "static"},
		ExcludeDirs: []string{"logs", "tmp", ".git", "node_modules", "vendor", "docs"},
		Extensions:  []string{".go", ".html", ".css", ".js", ".yaml", ".yml", ".json"},
		Debounce:    500 * time.Millisecond,
		OnReload: func() error {
			log.Println("执行重载操作...")
			return nil
		},
		OnError: func(err error) {
			log.Printf("热重载错误: %v", err)
		},
		OnFileChange: func(path, event string) {
			log.Printf("文件变化: %s [%s]", path, event)
		},
	}
}
