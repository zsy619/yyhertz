# 热重载

YYHertz 框架提供了强大的热重载功能，支持代码修改后自动重启应用程序，大大提升了开发效率。热重载功能监视文件变化，自动编译和重启，让开发者专注于业务逻辑的实现。

## 概述

热重载（Hot Reload）是现代开发工具的重要功能。YYHertz 的热重载系统提供：

- 文件变化监控
- 智能编译重启
- 配置文件热更新
- 静态资源自动刷新
- 依赖变化检测
- 多项目支持
- 自定义监控规则

## 基本使用

### 启用热重载

```go
package main

import (
    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/framework/mvc/devtools"
)

func main() {
    app := mvc.HertzApp
    
    // 启用开发工具（包含热重载）
    if err := devtools.SetupDevTools(app); err != nil {
        panic(err)
    }
    
    // 或者单独启用热重载
    if devtools.IsDevMode() {
        hotReloader := devtools.NewHotReloader(devtools.HotReloadConfig{
            WatchPaths: []string{".", "./controllers", "./models", "./views"},
            IgnorePatterns: []string{"*.log", "*.tmp", ".git/*"},
            BuildCommand: "go build -o app .",
            RunCommand: "./app",
        })
        
        hotReloader.Start()
        defer hotReloader.Stop()
    }
    
    app.Run()
}
```

### 命令行工具

```bash
# 使用 YYHertz CLI 启动热重载
yyhertz dev

# 指定配置文件
yyhertz dev --config=dev.yaml

# 指定监控路径
yyhertz dev --watch="./controllers,./models" 

# 指定忽略模式
yyhertz dev --ignore="*.log,*.tmp"

# 自定义构建命令
yyhertz dev --build="go build -tags dev -o app"

# 启用详细日志
yyhertz dev --verbose
```

## 配置选项

### 热重载配置

```go
type HotReloadConfig struct {
    // 是否启用热重载
    Enabled bool `yaml:"enabled" json:"enabled"`
    
    // 监控路径
    WatchPaths []string `yaml:"watch_paths" json:"watch_paths"`
    
    // 忽略的文件模式
    IgnorePatterns []string `yaml:"ignore_patterns" json:"ignore_patterns"`
    
    // 文件扩展名过滤
    Extensions []string `yaml:"extensions" json:"extensions"`
    
    // 构建命令
    BuildCommand string `yaml:"build_command" json:"build_command"`
    
    // 运行命令
    RunCommand string `yaml:"run_command" json:"run_command"`
    
    // 重启延迟（避免频繁重启）
    RestartDelay time.Duration `yaml:"restart_delay" json:"restart_delay"`
    
    // 构建超时
    BuildTimeout time.Duration `yaml:"build_timeout" json:"build_timeout"`
    
    // 是否在构建失败时保持运行
    KeepRunningOnFailure bool `yaml:"keep_running_on_failure" json:"keep_running_on_failure"`
    
    // 环境变量
    Env map[string]string `yaml:"env" json:"env"`
    
    // 预处理钩子
    PreBuildHooks []string `yaml:"pre_build_hooks" json:"pre_build_hooks"`
    
    // 后处理钩子
    PostBuildHooks []string `yaml:"post_build_hooks" json:"post_build_hooks"`
    
    // 通知配置
    Notifications NotificationConfig `yaml:"notifications" json:"notifications"`
}

type NotificationConfig struct {
    Enabled bool   `yaml:"enabled" json:"enabled"`
    Sound   bool   `yaml:"sound" json:"sound"`
    Desktop bool   `yaml:"desktop" json:"desktop"`
    Webhook string `yaml:"webhook" json:"webhook"`
}
```

### 配置文件示例

```yaml
# .yyhertz/hotreload.yaml
hot_reload:
  enabled: true
  watch_paths:
    - "."
    - "./controllers"
    - "./models" 
    - "./services"
    - "./views"
    - "./static"
  ignore_patterns:
    - "*.log"
    - "*.tmp"
    - "*.swp"
    - ".git/*"
    - "node_modules/*"
    - "vendor/*"
    - "build/*"
    - "dist/*"
  extensions:
    - ".go"
    - ".html"
    - ".css"
    - ".js"
    - ".yaml"
    - ".json"
  build_command: "go build -tags dev -o app ."
  run_command: "./app"
  restart_delay: "1s"
  build_timeout: "30s"
  keep_running_on_failure: true
  env:
    APP_ENV: "development"
    DEBUG: "true"
  pre_build_hooks:
    - "go mod tidy"
    - "go generate ./..."
  post_build_hooks:
    - "echo 'Build completed successfully'"
  notifications:
    enabled: true
    sound: true
    desktop: true
    webhook: ""
```

## 智能监控

### 文件变化检测

```go
package devtools

import (
    "path/filepath"
    "strings"
    "time"
    "github.com/fsnotify/fsnotify"
)

type FileWatcher struct {
    watcher   *fsnotify.Watcher
    config    HotReloadConfig
    debouncer *Debouncer
    onChange  func([]string)
}

func NewFileWatcher(config HotReloadConfig) *FileWatcher {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        panic(err)
    }
    
    return &FileWatcher{
        watcher:   watcher,
        config:    config,
        debouncer: NewDebouncer(config.RestartDelay),
    }
}

func (fw *FileWatcher) Start() error {
    // 添加监控路径
    for _, path := range fw.config.WatchPaths {
        if err := fw.addPath(path); err != nil {
            return err
        }
    }
    
    go fw.watchLoop()
    return nil
}

func (fw *FileWatcher) watchLoop() {
    for {
        select {
        case event, ok := <-fw.watcher.Events:
            if !ok {
                return
            }
            
            if fw.shouldIgnore(event.Name) {
                continue
            }
            
            if fw.isValidEvent(event) {
                fw.debouncer.Call(func() {
                    fw.onChange([]string{event.Name})
                })
            }
            
        case err, ok := <-fw.watcher.Errors:
            if !ok {
                return
            }
            log.Printf("Watcher error: %v", err)
        }
    }
}

func (fw *FileWatcher) shouldIgnore(path string) bool {
    for _, pattern := range fw.config.IgnorePatterns {
        if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
            return true
        }
        
        if strings.Contains(path, strings.TrimSuffix(pattern, "/*")) {
            return true
        }
    }
    return false
}

func (fw *FileWatcher) isValidEvent(event fsnotify.Event) bool {
    // 只关心写入和创建事件
    if event.Op&fsnotify.Write == fsnotify.Write ||
       event.Op&fsnotify.Create == fsnotify.Create {
        
        // 检查文件扩展名
        if len(fw.config.Extensions) > 0 {
            ext := filepath.Ext(event.Name)
            for _, validExt := range fw.config.Extensions {
                if ext == validExt {
                    return true
                }
            }
            return false
        }
        
        return true
    }
    
    return false
}
```

### 防抖动机制

```go
// 防抖动器，避免短时间内频繁重启
type Debouncer struct {
    delay time.Duration
    timer *time.Timer
    mutex sync.Mutex
}

func NewDebouncer(delay time.Duration) *Debouncer {
    return &Debouncer{
        delay: delay,
    }
}

func (d *Debouncer) Call(fn func()) {
    d.mutex.Lock()
    defer d.mutex.Unlock()
    
    if d.timer != nil {
        d.timer.Stop()
    }
    
    d.timer = time.AfterFunc(d.delay, fn)
}

func (d *Debouncer) Stop() {
    d.mutex.Lock()
    defer d.mutex.Unlock()
    
    if d.timer != nil {
        d.timer.Stop()
        d.timer = nil
    }
}
```

## 构建和重启管理

### 构建管理器

```go
type BuildManager struct {
    config  HotReloadConfig
    process *ProcessManager
    logger  Logger
}

func NewBuildManager(config HotReloadConfig) *BuildManager {
    return &BuildManager{
        config:  config,
        process: NewProcessManager(),
        logger:  GetLogger("build"),
    }
}

func (bm *BuildManager) Rebuild() error {
    bm.logger.Info("Starting rebuild...")
    
    // 执行预构建钩子
    if err := bm.executeHooks(bm.config.PreBuildHooks); err != nil {
        return fmt.Errorf("pre-build hook failed: %w", err)
    }
    
    // 停止当前进程
    if err := bm.process.Stop(); err != nil {
        bm.logger.Warn("Failed to stop process gracefully", "error", err)
    }
    
    // 执行构建
    buildStart := time.Now()
    if err := bm.build(); err != nil {
        bm.notifyBuildFailure(err)
        
        if bm.config.KeepRunningOnFailure {
            bm.logger.Warn("Build failed, keeping old version running", "error", err)
            return bm.process.Start(bm.config.RunCommand)
        }
        
        return err
    }
    
    buildDuration := time.Since(buildStart)
    bm.logger.Info("Build completed", "duration", buildDuration)
    
    // 执行后构建钩子
    if err := bm.executeHooks(bm.config.PostBuildHooks); err != nil {
        bm.logger.Warn("Post-build hook failed", "error", err)
    }
    
    // 启动新进程
    if err := bm.process.Start(bm.config.RunCommand); err != nil {
        return fmt.Errorf("failed to start process: %w", err)
    }
    
    bm.notifyBuildSuccess(buildDuration)
    return nil
}

func (bm *BuildManager) build() error {
    ctx, cancel := context.WithTimeout(context.Background(), bm.config.BuildTimeout)
    defer cancel()
    
    cmd := exec.CommandContext(ctx, "sh", "-c", bm.config.BuildCommand)
    
    // 设置环境变量
    cmd.Env = append(os.Environ(), bm.formatEnv()...)
    
    output, err := cmd.CombinedOutput()
    if err != nil {
        bm.logger.Error("Build failed", "output", string(output), "error", err)
        return fmt.Errorf("build command failed: %w", err)
    }
    
    if len(output) > 0 {
        bm.logger.Debug("Build output", "output", string(output))
    }
    
    return nil
}

func (bm *BuildManager) executeHooks(hooks []string) error {
    for _, hook := range hooks {
        if err := bm.executeHook(hook); err != nil {
            return err
        }
    }
    return nil
}

func (bm *BuildManager) executeHook(hook string) error {
    cmd := exec.Command("sh", "-c", hook)
    cmd.Env = append(os.Environ(), bm.formatEnv()...)
    
    output, err := cmd.CombinedOutput()
    if err != nil {
        bm.logger.Error("Hook failed", "hook", hook, "output", string(output), "error", err)
        return err
    }
    
    if len(output) > 0 {
        bm.logger.Debug("Hook output", "hook", hook, "output", string(output))
    }
    
    return nil
}

func (bm *BuildManager) formatEnv() []string {
    var env []string
    for k, v := range bm.config.Env {
        env = append(env, fmt.Sprintf("%s=%s", k, v))
    }
    return env
}
```

### 进程管理器

```go
type ProcessManager struct {
    cmd    *exec.Cmd
    mutex  sync.Mutex
    logger Logger
}

func NewProcessManager() *ProcessManager {
    return &ProcessManager{
        logger: GetLogger("process"),
    }
}

func (pm *ProcessManager) Start(command string) error {
    pm.mutex.Lock()
    defer pm.mutex.Unlock()
    
    if pm.cmd != nil && pm.cmd.Process != nil {
        return fmt.Errorf("process already running")
    }
    
    pm.cmd = exec.Command("sh", "-c", command)
    
    // 设置进程组，方便后续终止子进程
    pm.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
    
    // 重定向输出
    pm.cmd.Stdout = os.Stdout
    pm.cmd.Stderr = os.Stderr
    
    if err := pm.cmd.Start(); err != nil {
        return fmt.Errorf("failed to start process: %w", err)
    }
    
    pm.logger.Info("Process started", "pid", pm.cmd.Process.Pid)
    
    // 监控进程退出
    go pm.waitForExit()
    
    return nil
}

func (pm *ProcessManager) Stop() error {
    pm.mutex.Lock()
    defer pm.mutex.Unlock()
    
    if pm.cmd == nil || pm.cmd.Process == nil {
        return nil
    }
    
    pm.logger.Info("Stopping process", "pid", pm.cmd.Process.Pid)
    
    // 发送 SIGTERM 信号
    if err := pm.cmd.Process.Signal(syscall.SIGTERM); err != nil {
        pm.logger.Warn("Failed to send SIGTERM", "error", err)
    }
    
    // 等待进程优雅退出
    done := make(chan error, 1)
    go func() {
        done <- pm.cmd.Wait()
    }()
    
    select {
    case err := <-done:
        pm.logger.Info("Process stopped gracefully")
        pm.cmd = nil
        return err
    case <-time.After(10 * time.Second):
        // 强制终止
        pm.logger.Warn("Force killing process")
        if err := pm.cmd.Process.Kill(); err != nil {
            return fmt.Errorf("failed to kill process: %w", err)
        }
        pm.cmd = nil
        return nil
    }
}

func (pm *ProcessManager) waitForExit() {
    if pm.cmd == nil {
        return
    }
    
    err := pm.cmd.Wait()
    
    pm.mutex.Lock()
    if pm.cmd != nil {
        if err != nil {
            pm.logger.Error("Process exited with error", "error", err)
        } else {
            pm.logger.Info("Process exited normally")
        }
        pm.cmd = nil
    }
    pm.mutex.Unlock()
}

func (pm *ProcessManager) IsRunning() bool {
    pm.mutex.Lock()
    defer pm.mutex.Unlock()
    
    return pm.cmd != nil && pm.cmd.Process != nil
}
```

## 通知系统

### 构建通知

```go
type NotificationManager struct {
    config NotificationConfig
    logger Logger
}

func NewNotificationManager(config NotificationConfig) *NotificationManager {
    return &NotificationManager{
        config: config,
        logger: GetLogger("notification"),
    }
}

func (nm *NotificationManager) NotifyBuildSuccess(duration time.Duration) {
    if !nm.config.Enabled {
        return
    }
    
    message := fmt.Sprintf("✅ Build completed successfully in %v", duration)
    
    if nm.config.Desktop {
        nm.sendDesktopNotification("Build Success", message, "success")
    }
    
    if nm.config.Sound {
        nm.playNotificationSound("success")
    }
    
    if nm.config.Webhook != "" {
        nm.sendWebhookNotification("build_success", message)
    }
    
    nm.logger.Info(message)
}

func (nm *NotificationManager) NotifyBuildFailure(err error) {
    if !nm.config.Enabled {
        return
    }
    
    message := fmt.Sprintf("❌ Build failed: %v", err)
    
    if nm.config.Desktop {
        nm.sendDesktopNotification("Build Failed", message, "error")
    }
    
    if nm.config.Sound {
        nm.playNotificationSound("error")
    }
    
    if nm.config.Webhook != "" {
        nm.sendWebhookNotification("build_failure", message)
    }
    
    nm.logger.Error(message)
}

func (nm *NotificationManager) sendDesktopNotification(title, message, level string) {
    // 根据操作系统发送桌面通知
    switch runtime.GOOS {
    case "darwin":
        nm.sendMacOSNotification(title, message)
    case "linux":
        nm.sendLinuxNotification(title, message)
    case "windows":
        nm.sendWindowsNotification(title, message)
    }
}

func (nm *NotificationManager) sendMacOSNotification(title, message string) {
    script := fmt.Sprintf(`display notification "%s" with title "%s"`, message, title)
    exec.Command("osascript", "-e", script).Run()
}

func (nm *NotificationManager) sendLinuxNotification(title, message string) {
    exec.Command("notify-send", title, message).Run()
}

func (nm *NotificationManager) sendWindowsNotification(title, message string) {
    // Windows 通知实现
    // 可以使用 PowerShell 或第三方库
}

func (nm *NotificationManager) playNotificationSound(soundType string) {
    switch runtime.GOOS {
    case "darwin":
        if soundType == "success" {
            exec.Command("afplay", "/System/Library/Sounds/Glass.aiff").Run()
        } else {
            exec.Command("afplay", "/System/Library/Sounds/Sosumi.aiff").Run()
        }
    case "linux":
        if soundType == "success" {
            exec.Command("paplay", "/usr/share/sounds/alsa/Front_Left.wav").Run()
        } else {
            exec.Command("paplay", "/usr/share/sounds/alsa/Side_Left.wav").Run()
        }
    }
}

func (nm *NotificationManager) sendWebhookNotification(event, message string) {
    payload := map[string]interface{}{
        "event":     event,
        "message":   message,
        "timestamp": time.Now().Unix(),
        "hostname":  getHostname(),
    }
    
    jsonData, _ := json.Marshal(payload)
    
    go func() {
        resp, err := http.Post(nm.config.Webhook, "application/json", bytes.NewBuffer(jsonData))
        if err != nil {
            nm.logger.Warn("Failed to send webhook notification", "error", err)
            return
        }
        defer resp.Body.Close()
        
        if resp.StatusCode >= 400 {
            nm.logger.Warn("Webhook notification failed", "status", resp.StatusCode)
        }
    }()
}
```

## 静态资源监控

### 前端资源热更新

```go
type AssetWatcher struct {
    config    AssetWatchConfig
    server    *http.Server
    clients   map[string]chan string
    clientsMu sync.RWMutex
}

type AssetWatchConfig struct {
    Enabled     bool     `yaml:"enabled"`
    WatchPaths  []string `yaml:"watch_paths"`
    Extensions  []string `yaml:"extensions"`
    Port        int      `yaml:"port"`
    LiveReload  bool     `yaml:"live_reload"`
}

func NewAssetWatcher(config AssetWatchConfig) *AssetWatcher {
    return &AssetWatcher{
        config:  config,
        clients: make(map[string]chan string),
    }
}

func (aw *AssetWatcher) Start() error {
    if !aw.config.Enabled {
        return nil
    }
    
    // 启动 WebSocket 服务器用于浏览器通信
    mux := http.NewServeMux()
    mux.HandleFunc("/livereload", aw.handleWebSocket)
    
    aw.server = &http.Server{
        Addr:    fmt.Sprintf(":%d", aw.config.Port),
        Handler: mux,
    }
    
    go func() {
        if err := aw.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Printf("Asset watcher server error: %v", err)
        }
    }()
    
    // 监控静态资源变化
    go aw.watchAssets()
    
    return nil
}

func (aw *AssetWatcher) handleWebSocket(w http.ResponseWriter, r *http.Request) {
    upgrader := websocket.Upgrader{
        CheckOrigin: func(r *http.Request) bool { return true },
    }
    
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        return
    }
    defer conn.Close()
    
    clientID := generateClientID()
    client := make(chan string, 10)
    
    aw.clientsMu.Lock()
    aw.clients[clientID] = client
    aw.clientsMu.Unlock()
    
    defer func() {
        aw.clientsMu.Lock()
        delete(aw.clients, clientID)
        close(client)
        aw.clientsMu.Unlock()
    }()
    
    for message := range client {
        if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
            break
        }
    }
}

func (aw *AssetWatcher) watchAssets() {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        return
    }
    defer watcher.Close()
    
    for _, path := range aw.config.WatchPaths {
        watcher.Add(path)
    }
    
    debouncer := NewDebouncer(500 * time.Millisecond)
    defer debouncer.Stop()
    
    for {
        select {
        case event, ok := <-watcher.Events:
            if !ok {
                return
            }
            
            if aw.isAssetFile(event.Name) {
                debouncer.Call(func() {
                    aw.notifyClients(event.Name)
                })
            }
        case err, ok := <-watcher.Errors:
            if !ok {
                return
            }
            log.Printf("Asset watcher error: %v", err)
        }
    }
}

func (aw *AssetWatcher) isAssetFile(path string) bool {
    ext := filepath.Ext(path)
    for _, validExt := range aw.config.Extensions {
        if ext == validExt {
            return true
        }
    }
    return false
}

func (aw *AssetWatcher) notifyClients(changedFile string) {
    message := fmt.Sprintf(`{"type":"reload","file":"%s"}`, changedFile)
    
    aw.clientsMu.RLock()
    for _, client := range aw.clients {
        select {
        case client <- message:
        default:
            // 客户端缓冲区满，跳过
        }
    }
    aw.clientsMu.RUnlock()
}

// 在 HTML 中注入的 JavaScript
const liveReloadScript = `
<script>
(function() {
    var ws = new WebSocket('ws://localhost:35729/livereload');
    ws.onmessage = function(event) {
        var data = JSON.parse(event.data);
        if (data.type === 'reload') {
            console.log('Reloading due to', data.file);
            window.location.reload();
        }
    };
    ws.onopen = function() {
        console.log('Live reload connected');
    };
    ws.onclose = function() {
        console.log('Live reload disconnected');
        // 尝试重连
        setTimeout(arguments.callee, 1000);
    };
})();
</script>
`
```

## 性能优化

### 智能编译

```go
type SmartBuilder struct {
    lastBuildTime time.Time
    dependencies  map[string]time.Time
    checksumCache map[string]string
}

func (sb *SmartBuilder) NeedsBuild(changedFiles []string) bool {
    // 检查是否有 Go 文件变化
    for _, file := range changedFiles {
        if strings.HasSuffix(file, ".go") {
            return true
        }
    }
    
    // 检查依赖文件
    if sb.dependenciesChanged() {
        return true
    }
    
    // 检查配置文件
    if sb.configChanged(changedFiles) {
        return true
    }
    
    return false
}

func (sb *SmartBuilder) dependenciesChanged() bool {
    // 检查 go.mod 和 go.sum
    files := []string{"go.mod", "go.sum"}
    
    for _, file := range files {
        if stat, err := os.Stat(file); err == nil {
            if lastMod := stat.ModTime(); lastMod.After(sb.lastBuildTime) {
                return true
            }
        }
    }
    
    return false
}

func (sb *SmartBuilder) configChanged(changedFiles []string) bool {
    configExts := []string{".yaml", ".yml", ".json", ".toml"}
    
    for _, file := range changedFiles {
        ext := filepath.Ext(file)
        for _, configExt := range configExts {
            if ext == configExt {
                return true
            }
        }
    }
    
    return false
}
```

### 增量编译

```go
type IncrementalBuilder struct {
    buildCache map[string]BuildInfo
    mutex      sync.RWMutex
}

type BuildInfo struct {
    Hash      string
    BuildTime time.Time
    Success   bool
}

func (ib *IncrementalBuilder) Build(files []string) error {
    ib.mutex.Lock()
    defer ib.mutex.Unlock()
    
    // 计算文件哈希
    currentHash := ib.calculateHash(files)
    
    // 检查缓存
    if info, exists := ib.buildCache["main"]; exists {
        if info.Hash == currentHash && info.Success {
            return nil // 无需重新构建
        }
    }
    
    // 执行构建
    startTime := time.Now()
    err := ib.doBuild()
    
    // 更新缓存
    ib.buildCache["main"] = BuildInfo{
        Hash:      currentHash,
        BuildTime: startTime,
        Success:   err == nil,
    }
    
    return err
}

func (ib *IncrementalBuilder) calculateHash(files []string) string {
    h := sha256.New()
    
    for _, file := range files {
        if data, err := os.ReadFile(file); err == nil {
            h.Write(data)
        }
    }
    
    return fmt.Sprintf("%x", h.Sum(nil))
}
```

## 最佳实践

### 1. 配置优化

```yaml
# 生产环境禁用热重载
hot_reload:
  enabled: false  # 在生产环境中关闭

# 开发环境优化配置
hot_reload:
  enabled: true
  watch_paths: ["./controllers", "./models", "./services"]  # 只监控必要路径
  ignore_patterns: 
    - "*.log"
    - "*.tmp"
    - ".git/*"
    - "vendor/*"  # 忽略第三方代码
  restart_delay: "500ms"  # 适中的延迟
  build_timeout: "30s"    # 合理的构建超时
```

### 2. 监控规则

```go
// 自定义监控规则
func setupWatchRules() HotReloadConfig {
    return HotReloadConfig{
        WatchPaths: []string{
            "./controllers",  // 控制器变化需要重启
            "./models",       // 模型变化需要重启
            "./services",     // 服务变化需要重启
            "./config",       // 配置变化需要重启
        },
        IgnorePatterns: []string{
            "*.log",          // 忽略日志文件
            "*.tmp",          // 忽略临时文件
            "*_test.go",      // 忽略测试文件（可选）
            ".git/*",         // 忽略 Git 文件
            "vendor/*",       // 忽略依赖包
        },
        Extensions: []string{
            ".go",            // Go 源码
            ".yaml", ".yml",  // 配置文件
            ".json",          // JSON 配置
        },
    }
}
```

### 3. 性能监控

```go
type HotReloadMetrics struct {
    BuildCount    int64         `json:"build_count"`
    SuccessCount  int64         `json:"success_count"`
    FailureCount  int64         `json:"failure_count"`
    AvgBuildTime  time.Duration `json:"avg_build_time"`
    LastBuildTime time.Time     `json:"last_build_time"`
}

func (hr *HotReloader) GetMetrics() HotReloadMetrics {
    return hr.metrics
}

// 监控 API
func (c *DevToolsController) GetHotReloadMetrics() {
    metrics := hotReloader.GetMetrics()
    c.JSON(200, metrics)
}
```

### 4. 故障排除

```go
// 诊断工具
func (hr *HotReloader) Diagnose() DiagnosticInfo {
    return DiagnosticInfo{
        WatcherStatus:   hr.watcher.IsRunning(),
        ProcessStatus:   hr.process.IsRunning(),
        LastError:      hr.lastError,
        WatchedPaths:   hr.config.WatchPaths,
        IgnoredPattern: hr.config.IgnorePatterns,
        BuildHistory:   hr.buildHistory,
    }
}

// 健康检查
func (c *DevToolsController) HealthCheck() {
    diagnosis := hotReloader.Diagnose()
    
    status := "healthy"
    if !diagnosis.WatcherStatus || !diagnosis.ProcessStatus {
        status = "unhealthy"
    }
    
    c.JSON(200, gin.H{
        "status":    status,
        "diagnosis": diagnosis,
    })
}
```

YYHertz 的热重载功能为开发者提供了高效的开发体验，通过智能的文件监控、快速的构建重启和友好的通知系统，让开发过程更加流畅和高效。
