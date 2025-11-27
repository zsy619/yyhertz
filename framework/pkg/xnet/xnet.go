// Package xnet 提供了网络操作的工具函数
//
// 主要功能：
//   - IP地址处理：IP验证、IP段检查、地理位置查询等
//   - URL处理：URL解析、拼接、编码解码、参数操作等
//   - HTTP客户端：HTTP请求、响应处理、重试机制、超时控制等
//   - 网络检测：端口检测、网络连通性、延迟测试等
//   - 域名解析：DNS查询、域名验证、子域名检测等
//
// 基础用法：
//
//	import "your-project/framework/pkg/xnet"
//
//	// IP地址操作
//	isValid := xnet.IsValidIP("192.168.1.1")
//	isPrivate := xnet.IsPrivateIP("10.0.0.1")
//	ipInfo := xnet.GetIPInfo("8.8.8.8")
//
//	// URL操作
//	parsed, err := xnet.ParseURL("https://example.com/path?key=value")
//	encoded := xnet.URLEncode("hello world")
//	params := xnet.ParseURLParams("key1=value1&key2=value2")
//
//	// HTTP请求
//	resp, err := xnet.HTTPGet("https://api.example.com/data")
//	data, err := xnet.HTTPPost("https://api.example.com", jsonData)
//	client := xnet.NewHTTPClient(30 * time.Second)
//
// 高级用法：
//
//	// 网络检测
//	isOpen := xnet.IsPortOpen("example.com", 80)
//	latency := xnet.PingHost("google.com")
//	isConnected := xnet.CheckConnection("https://example.com")
//
//	// 域名操作
//	ips, err := xnet.LookupIP("example.com")
//	isValidDomain := xnet.IsValidDomain("example.com")
//	mx, err := xnet.LookupMX("gmail.com")
//
//	// 高级HTTP操作
//	req := xnet.NewHTTPRequest("POST", "https://api.example.com")
//	req.SetHeader("Authorization", "Bearer token")
//	req.SetTimeout(60 * time.Second)
//	req.SetRetry(3)
//	resp, err := req.Execute()
//
//	// 网络工具
//	localIP := xnet.GetLocalIP()
//	publicIP := xnet.GetPublicIP()
//	mac := xnet.GetMACAddress()
package xnet
