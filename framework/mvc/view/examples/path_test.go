package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestAllPath(t *testing.T) {
	t.Log("Go路径获取示例")
	t.Log("==============")

	// 1. 获取当前工作目录
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatal("获取当前工作目录失败:", err)
	}
	fmt.Printf("1. 当前工作目录: %s\n", currentDir)

	// 2. 获取可执行文件路径
	execPath, err := os.Executable()
	if err != nil {
		log.Fatal("获取可执行文件路径失败:", err)
	}
	t.Logf("2. 可执行文件路径: %s\n", execPath)

	// 3. 获取可执行文件所在目录
	execDir := filepath.Dir(execPath)
	t.Logf("3. 可执行文件所在目录: %s\n", execDir)

	// 4. 获取用户主目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatal("获取用户主目录失败:", err)
	}
	t.Logf("4. 用户主目录: %s\n", homeDir)

	// 5. 获取用户缓存目录
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		t.Fatal("获取用户缓存目录失败:", err)
	}
	t.Logf("5. 用户缓存目录: %s\n", cacheDir)

	// 6. 获取用户配置目录
	configDir, err := os.UserConfigDir()
	if err != nil {
		t.Error("获取用户配置目录失败:", err)
	}
	t.Logf("6. 用户配置目录: %s\n", configDir)

	// 7. 获取临时目录
	tempDir := os.TempDir()
	t.Logf("7. 临时目录: %s\n", tempDir)

	// 8. 获取当前源代码文件名和行号
	_, file, line, ok := runtime.Caller(0)
	if ok {
		t.Logf("8. 当前源代码位置: %s:%d\n", file, line)
	}

	// 9. 获取GOROOT（Go安装目录）
	goRoot := runtime.GOROOT()
	t.Logf("9. GOROOT: %s\n", goRoot)

	// 10. 使用filepath包处理路径示例
	examplePath := filepath.Join("dir", "subdir", "file.txt")
	t.Logf("10. 使用filepath.Join处理的路径: %s\n", examplePath)

	// 11. 获取路径的绝对路径
	absPath, err := filepath.Abs("./somefile.txt")
	if err != nil {
		t.Fatal("获取绝对路径失败:", err)
	}
	t.Logf("11. 绝对路径示例: %s\n", absPath)

	// 12. 获取路径的基名（最后一部分）
	baseName := filepath.Base(examplePath)
	t.Logf("12. 路径基名: %s\n", baseName)

	// 13. 获取路径的目录部分
	dirName := filepath.Dir(examplePath)
	t.Logf("13. 路径目录部分: %s\n", dirName)

	// 14. 获取路径的扩展名
	ext := filepath.Ext(examplePath)
	t.Logf("14. 路径扩展名: %s\n", ext)

	// 15. 使用环境变量获取路径
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		goPath = "未设置"
	}
	fmt.Printf("15. GOPATH: %s\n", goPath)

	// 16. 获取当前进程的PID
	pid := os.Getpid()
	t.Logf("16. 当前进程PID: %d\n", pid)

	t.Log("\n路径测试完成!")
}

func TestAbsolatePath(t *testing.T) {
	fmt.Println("filepath.IsAbs 测试用例")
	fmt.Println("=====================")

	// 测试不同操作系统下的绝对路径判断
	testCases := []struct {
		path     string
		expected bool
		desc     string
	}{
		// Unix/Linux/macOS 风格的绝对路径
		{"/usr/local", true, "Unix绝对路径"},
		{"/home/user/file.txt", true, "Unix绝对文件路径"},

		// Windows 风格的绝对路径
		{"C:\\Windows\\System32", true, "Windows绝对路径（反斜杠）"},
		{"C:/Windows/System32", true, "Windows绝对路径（正斜杠）"},
		{"\\\\Server\\Share", true, "Windows UNC路径"},

		// 相对路径
		{".", false, "当前目录"},
		{"..", false, "父目录"},
		{"../file.txt", false, "相对文件路径"},
		{"dir/file.txt", false, "相对子目录文件路径"},
		{"file.txt", false, "当前目录下的文件"},

		// 空路径和特殊路径
		{"", false, "空路径"},
		{"/", true, "根目录"},
		{"~", false, "家目录符号（相对路径）"},
	}

	fmt.Printf("当前操作系统: %s\n", runtime.GOOS)
	fmt.Println("")

	for _, tc := range testCases {
		result := filepath.IsAbs(tc.path)
		status := "✓"
		if result != tc.expected {
			status = "✗"
		}

		fmt.Printf("%s 路径: %-30s 是否为绝对路径: %-5v 期望: %-5v %s\n",
			status,
			fmt.Sprintf("\"%s\"", tc.path),
			result,
			tc.expected,
			tc.desc)
	}

	// 演示跨平台路径处理
	fmt.Println("\n跨平台路径处理演示:")
	demoCrossPlatformPaths()
}

func demoCrossPlatformPaths() {
	// 演示在不同操作系统上处理路径
	paths := []string{
		"/home/user/file.txt",
		"C:\\Windows\\System32",
		"../relative/path",
		"file.txt",
	}

	for _, path := range paths {
		abs := filepath.IsAbs(path)
		fmt.Printf("路径: %-20s 绝对路径: %v\n", path, abs)

		// 转换为当前系统的绝对路径
		if !abs {
			absPath, err := filepath.Abs(path)
			if err == nil {
				fmt.Printf("  转换为绝对路径: %s\n", absPath)
			}
		}
	}
}
