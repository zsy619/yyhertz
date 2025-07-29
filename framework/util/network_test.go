package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLocalIPv4Address(t *testing.T) {
	t.Run("获取本机IPv4地址", func(t *testing.T) {
		ip, err := GetLocalIPv4Address()
		
		// 在某些环境下可能没有非回环接口，所以允许错误
		if err == nil {
			assert.NotEmpty(t, ip)
			assert.True(t, IsValidIP(ip))
			// 应该不是回环地址
			assert.NotEqual(t, "127.0.0.1", ip)
		}
	})
}

func TestGetFreePort(t *testing.T) {
	t.Run("获取可用端口", func(t *testing.T) {
		port, err := GetFreePort()
		
		assert.NoError(t, err)
		assert.Greater(t, port, 0)
		assert.Less(t, port, 65536)
		
		// 获取两个端口应该不同
		port2, err2 := GetFreePort()
		assert.NoError(t, err2)
		assert.NotEqual(t, port, port2)
	})
}

func TestNetworkUtilsIntegration(t *testing.T) {
	t.Run("网络工具集成测试", func(t *testing.T) {
		// 测试获取本地IP列表
		ips := GetLocalIPs()
		
		if len(ips) > 0 {
			// 验证获取的IP地址有效性
			for _, ip := range ips {
				assert.True(t, IsValidIP(ip))
				// 本地局域网IP应该被IsLocalIP识别，但127.0.0.1已被排除在GetLocalIPs之外
			}
			
			// 测试IPv4地址获取
			ipv4, err := GetLocalIPv4Address()
			if err == nil {
				assert.Contains(t, ips, ipv4)
			}
		}
		
		// 测试端口获取
		ports := make([]int, 5)
		for i := 0; i < 5; i++ {
			port, err := GetFreePort()
			assert.NoError(t, err)
			ports[i] = port
		}
		
		// 验证获取的端口都不相同
		for i := 0; i < len(ports); i++ {
			for j := i + 1; j < len(ports); j++ {
				assert.NotEqual(t, ports[i], ports[j])
			}
		}
	})
}