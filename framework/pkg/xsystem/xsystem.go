// Package xsystem 提供了系统操作的工具函数
//
// 主要功能：
//   - 系统信息：CPU、内存、磁盘、操作系统等信息获取
//   - 进程管理：进程查询、启动、停止、监控等
//   - 环境变量：环境变量读取、设置、管理等
//   - 命令执行：系统命令执行、输出捕获、异步执行等
//   - 性能监控：系统负载、资源使用率、性能指标等
//
// 基础用法：
//
//	import "your-project/framework/pkg/xsystem"
//
//	// 系统信息获取
//	cpuCount := xsystem.CPUCount()
//	memInfo := xsystem.MemoryInfo()
//	diskInfo := xsystem.DiskInfo("/")
//	osInfo := xsystem.OSInfo()
//
//	// 环境变量操作
//	path := xsystem.GetEnv("PATH")
//	xsystem.SetEnv("MY_VAR", "value")
//	envs := xsystem.GetAllEnv()
//
//	// 命令执行
//	output, err := xsystem.Exec("ls", "-la")
//	result := xsystem.ExecCommand("python", "script.py")
//	xsystem.ExecAsync("long-running-command")
//
// 高级用法：
//
//	// 系统性能监控
//	cpuUsage := xsystem.CPUUsage()
//	memUsage := xsystem.MemoryUsage()
//	loadAvg := xsystem.LoadAverage()
//
//	// 进程管理
//	processes := xsystem.ListProcesses()
//	process := xsystem.FindProcess("nginx")
//	xsystem.KillProcess(1234)
//	isRunning := xsystem.IsProcessRunning("mysqld")
//
//	// 用户和权限
//	currentUser := xsystem.CurrentUser()
//	userHome := xsystem.UserHomeDir()
//	hasPermission := xsystem.HasPermission("/etc/passwd", "r")
//
//	// 网络接口信息
//	interfaces := xsystem.NetworkInterfaces()
//	defaultGateway := xsystem.DefaultGateway()
//	dnsServers := xsystem.DNSServers()
//
//	// 硬件信息
//	cpuInfo := xsystem.CPUInfo()
//	gpuInfo := xsystem.GPUInfo()
//	usbDevices := xsystem.USBDevices()
//
// 系统监控示例：
//
//	// 实时系统监控
//	monitor := xsystem.NewSystemMonitor()
//	monitor.SetInterval(time.Second * 5)
//	monitor.OnCPUAlert(80.0, func(usage float64) {
//		log.Printf("CPU使用率过高: %.2f%%", usage)
//	})
//	monitor.OnMemoryAlert(90.0, func(usage float64) {
//		log.Printf("内存使用率过高: %.2f%%", usage)
//	})
//	monitor.Start()
//
//	// 服务管理
//	service := xsystem.NewService("nginx")
//	service.Start()
//	service.Stop()
//	service.Restart()
//	status := service.Status()
//
package xsystem