package util

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"
)

// GetClientIP 获取客户端真实IP地址
// 解析 X-Real-IP 和 X-Forwarded-For 以便于反向代理（nginx 或 haproxy）可以正常工作
func GetClientIP(r *http.Request) string {
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	ip := strings.TrimSpace(strings.Split(xForwardedFor, ",")[0])
	if ip != "" && ip != "unknown" {
		return ip
	}
	
	ip = strings.TrimSpace(r.Header.Get("X-Real-IP"))
	if ip != "" && ip != "unknown" {
		return ip
	}
	
	ip = strings.TrimSpace(r.Header.Get("X-Original-Forwarded-For"))
	if ip != "" && ip != "unknown" {
		return ip
	}
	
	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		return ip
	}
	
	return ""
}

// IsLocalIP 检查是否为本地IP
func IsLocalIP(ip string) bool {
	if ip == "" {
		return false
	}
	return ip == "::1" || ip == "127.0.0.1" || strings.HasPrefix(ip, "192.168.") || 
		   strings.HasPrefix(ip, "10.") || strings.HasPrefix(ip, "172.")
}

// IsValidIP 验证IP地址格式
func IsValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// GetIPLocation 根据IP获取地理位置信息
func GetIPLocation(ip string) map[string]string {
	result := map[string]string{
		"ip":       ip,
		"country":  "",
		"province": "",
		"city":     "",
		"isp":      "",
	}
	
	if ip == "" {
		return result
	}
	
	if IsLocalIP(ip) {
		result["country"] = "中国"
		result["province"] = "内网"
		result["city"] = "内网IP"
		result["isp"] = "内网"
		return result
	}
	
	// 尝试从多个API获取IP信息
	if info := getIPInfoFromAPI1(ip); info != nil {
		if country, ok := info["country"].(string); ok {
			result["country"] = country
		}
		if province, ok := info["province"].(string); ok {
			result["province"] = province
		}
		if city, ok := info["city"].(string); ok {
			result["city"] = city
		}
		if isp, ok := info["isp"].(string); ok {
			result["isp"] = isp
		}
	}
	
	return result
}

// getIPInfoFromAPI1 从第一个API获取IP信息
func getIPInfoFromAPI1(ip string) map[string]any {
	url := "http://ip-api.com/json/" + ip + "?lang=zh-CN"
	
	client := &http.Client{}
	resp, err := client.Get(url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return nil
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	
	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		return nil
	}
	
	// 转换字段名
	info := make(map[string]any)
	if status, ok := result["status"].(string); ok && status == "success" {
		if country, ok := result["country"].(string); ok {
			info["country"] = country
		}
		if region, ok := result["regionName"].(string); ok {
			info["province"] = region
		}
		if city, ok := result["city"].(string); ok {
			info["city"] = city
		}
		if isp, ok := result["isp"].(string); ok {
			info["isp"] = isp
		}
	}
	
	return info
}

// GetCityByIP 根据IP获取城市信息(兼容原函数)
func GetCityByIP(ip string) string {
	if ip == "" {
		return "未知"
	}
	
	if IsLocalIP(ip) {
		return "内网IP"
	}
	
	location := GetIPLocation(ip)
	if location["city"] != "" {
		return location["city"]
	}
	
	return "未知"
}

// ValidateIPRange 验证IP是否在指定范围内
func ValidateIPRange(ip string, cidr string) bool {
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}
	
	ipAddr := net.ParseIP(ip)
	if ipAddr == nil {
		return false
	}
	
	return network.Contains(ipAddr)
}

// GetLocalIPs 获取本机所有IP地址
func GetLocalIPs() []string {
	var ips []string
	
	interfaces, err := net.Interfaces()
	if err != nil {
		return ips
	}
	
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			ips = append(ips, ip.String())
		}
	}
	
	return ips
}

// GetLocalIPv4Address 获取本机IPv4地址(来自FreeCar项目)
func GetLocalIPv4Address() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipv4 := ipnet.IP.To4(); ipv4 != nil {
				return ipv4.String(), nil
			}
		}
	}
	return "", net.ErrWriteToConnected
}

// GetFreePort 获取可用端口号(来自FreeCar项目)
func GetFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}
	
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	
	return listener.Addr().(*net.TCPAddr).Port, nil
}