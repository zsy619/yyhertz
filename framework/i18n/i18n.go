// Package i18n 提供国际化支持
package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/zsy619/yyhertz/framework/config"
)

// I18n 国际化管理器
type I18n struct {
	defaultLocale string
	currentLocale string
	messages      map[string]map[string]string // locale -> key -> message
	mutex         sync.RWMutex
}

// NewI18n 创建国际化管理器
func NewI18n(defaultLocale string) *I18n {
	return &I18n{
		defaultLocale: defaultLocale,
		currentLocale: defaultLocale,
		messages:      make(map[string]map[string]string),
	}
}

// LoadMessages 加载消息文件
func (i *I18n) LoadMessages(locale, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var messages map[string]string
	if err := json.Unmarshal(data, &messages); err != nil {
		return err
	}

	i.mutex.Lock()
	defer i.mutex.Unlock()

	if i.messages[locale] == nil {
		i.messages[locale] = make(map[string]string)
	}

	for key, value := range messages {
		i.messages[locale][key] = value
	}

	return nil
}

// T 翻译函数
func (i *I18n) T(key string, args ...any) string {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	// 尝试当前语言
	if messages, exists := i.messages[i.currentLocale]; exists {
		if msg, exists := messages[key]; exists {
			if len(args) > 0 {
				return fmt.Sprintf(msg, args...)
			}
			return msg
		}
	}

	// 尝试默认语言
	if i.currentLocale != i.defaultLocale {
		if messages, exists := i.messages[i.defaultLocale]; exists {
			if msg, exists := messages[key]; exists {
				if len(args) > 0 {
					return fmt.Sprintf(msg, args...)
				}
				return msg
			}
		}
	}

	// 返回键名作为后备
	return key
}

// SetLocale 设置当前语言
func (i *I18n) SetLocale(locale string) {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	i.currentLocale = locale
}

// GetLocale 获取当前语言
func (i *I18n) GetLocale() string {
	i.mutex.RLock()
	defer i.mutex.RUnlock()
	return i.currentLocale
}

// 全局实例
var globalI18n = NewI18n("en")

// T 全局翻译函数
func T(key string, args ...any) string {
	return globalI18n.T(key, args...)
}

// SetLocale 设置全局语言
func SetLocale(locale string) {
	globalI18n.SetLocale(locale)
}

// LoadMessages 加载全局消息
func LoadMessages(locale, filePath string) error {
	return globalI18n.LoadMessages(locale, filePath)
}

// LoadMessagesFromDir 从目录加载所有消息文件
func LoadMessagesFromDir(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".json") {
			locale := strings.TrimSuffix(info.Name(), ".json")
			if err := LoadMessages(locale, path); err != nil {
				config.Errorf("Failed to load messages for locale %s: %v", locale, err)
			} else {
				config.Infof("Loaded messages for locale: %s", locale)
			}
		}
		return nil
	})
}
