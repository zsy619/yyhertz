package util

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
)

// IsWindows 判断当前操作系统是否为Windows
func IsWindows() bool {
	return runtime.GOOS == "windows"
}

// IsLinux 判断当前操作系统是否为Linux
func IsLinux() bool {
	return runtime.GOOS == "linux"
}

// IsDarwin 判断当前操作系统是否为Darwin(macOS)
func IsDarwin() bool {
	return runtime.GOOS == "darwin"
}

// IsMacOS 判断当前操作系统是否为macOS
func IsMacOS() bool {
	return IsDarwin()
}

// GetOS 获取操作系统名称
func GetOS() string {
	return runtime.GOOS
}

// GetArch 获取系统架构
func GetArch() string {
	return runtime.GOARCH
}

// GetOSInfo 获取操作系统信息
func GetOSInfo() map[string]string {
	return map[string]string{
		"os":      runtime.GOOS,
		"arch":    runtime.GOARCH,
		"version": runtime.Version(),
		"numCPU":  string(rune(runtime.NumCPU() + '0')),
	}
}

// GetCurrentWorkingDir 获取当前工作目录
func GetCurrentWorkingDir() (string, error) {
	return os.Getwd()
}

// GetExecutableDir 获取可执行文件所在目录
func GetExecutableDir() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(execPath), nil
}

// Pwd 获取当前工作目录路径(兼容原函数)
func Pwd() string {
	if dir, err := GetCurrentWorkingDir(); err == nil {
		return dir
	}
	
	// 备用方法
	file, _ := exec.LookPath(os.Args[0])
	pwd, _ := filepath.Abs(file)
	return filepath.Dir(pwd)
}

// Home 获取当前用户的主目录
func Home() (string, error) {
	currentUser, err := user.Current()
	if err == nil && currentUser.HomeDir != "" {
		return currentUser.HomeDir, nil
	}

	// 跨平台兼容处理
	if IsWindows() {
		return homeWindows()
	}

	// Unix-like系统
	return homeUnix()
}

// homeUnix 获取Unix系统的主目录
func homeUnix() (string, error) {
	// 首选HOME环境变量
	if home := os.Getenv("HOME"); home != "" {
		return home, nil
	}

	// 如果失败，尝试使用shell命令
	var stdout bytes.Buffer
	cmd := exec.Command("sh", "-c", "eval echo ~$USER")
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return "", err
	}

	result := strings.TrimSpace(stdout.String())
	if result == "" {
		return "", errors.New("blank output when reading home directory")
	}

	return result, nil
}

// homeWindows 获取Windows系统的主目录
func homeWindows() (string, error) {
	drive := os.Getenv("HOMEDRIVE")
	path := os.Getenv("HOMEPATH")
	home := drive + path
	
	if drive == "" || path == "" {
		home = os.Getenv("USERPROFILE")
	}
	
	if home == "" {
		return "", errors.New("HOMEDRIVE, HOMEPATH, and USERPROFILE are blank")
	}

	return home, nil
}

// GetEnv 获取环境变量，如果不存在则返回默认值
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// SetEnv 设置环境变量
func SetEnv(key, value string) error {
	return os.Setenv(key, value)
}

// GetAllEnv 获取所有环境变量
func GetAllEnv() []string {
	return os.Environ()
}

// GetEnvMap 获取环境变量映射
func GetEnvMap() map[string]string {
	envMap := make(map[string]string)
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if len(pair) == 2 {
			envMap[pair[0]] = pair[1]
		}
	}
	return envMap
}

// PathExists 检查路径是否存在
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsDir 检查路径是否为目录
func IsDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// IsFile 检查路径是否为文件
func IsFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// CreateDir 创建目录
func CreateDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// CreateFile 创建文件
func CreateFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	return file.Close()
}

// RemovePath 删除路径（文件或目录）
func RemovePath(path string) error {
	return os.RemoveAll(path)
}

// CopyFile 复制文件
func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// MoveFile 移动文件
func MoveFile(src, dst string) error {
	return os.Rename(src, dst)
}

// GetFileSize 获取文件大小
func GetFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// GetFileInfo 获取文件信息
func GetFileInfo(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

// ListDir 列出目录内容
func ListDir(path string) ([]os.FileInfo, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return file.Readdir(-1)
}

// ListDirNames 列出目录中的文件名
func ListDirNames(path string) ([]string, error) {
	files, err := ListDir(path)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(files))
	for i, file := range files {
		names[i] = file.Name()
	}
	return names, nil
}

// WalkDir 递归遍历目录
func WalkDir(root string, fn filepath.WalkFunc) error {
	return filepath.Walk(root, fn)
}

// GetTempDir 获取临时目录
func GetTempDir() string {
	return os.TempDir()
}

// CreateTempFile 创建临时文件
func CreateTempFile(pattern string) (*os.File, error) {
	return os.CreateTemp("", pattern)
}

// CreateTempDir 创建临时目录
func CreateTempDir(pattern string) (string, error) {
	return os.MkdirTemp("", pattern)
}

// GetHostname 获取主机名
func GetHostname() (string, error) {
	return os.Hostname()
}

// GetUserInfo 获取当前用户信息
func GetUserInfo() (*user.User, error) {
	return user.Current()
}

// GetUID 获取用户ID
func GetUID() string {
	if currentUser, err := user.Current(); err == nil {
		return currentUser.Uid
	}
	return ""
}

// GetGID 获取组ID
func GetGID() string {
	if currentUser, err := user.Current(); err == nil {
		return currentUser.Gid
	}
	return ""
}

// GetUsername 获取用户名
func GetUsername() string {
	if currentUser, err := user.Current(); err == nil {
		return currentUser.Username
	}
	return ""
}

// ExecCommand 执行系统命令
func ExecCommand(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	return cmd.Output()
}

// ExecCommandWithDir 在指定目录执行系统命令
func ExecCommandWithDir(dir, name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	return cmd.Output()
}

// DetailedSystemInfo 详细系统信息结构
type DetailedSystemInfo struct {
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Version  string `json:"version"`
	NumCPU   int    `json:"numCpu"`
	Hostname string `json:"hostname"`
	Username string `json:"username"`
	HomeDir  string `json:"homeDir"`
	WorkDir  string `json:"workDir"`
}

// GetDetailedSystemInfo 获取详细系统信息
func GetDetailedSystemInfo() *DetailedSystemInfo {
	hostname, _ := GetHostname()
	homeDir, _ := Home()
	workDir, _ := GetCurrentWorkingDir()
	
	return &DetailedSystemInfo{
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		Version:  runtime.Version(),
		NumCPU:   runtime.NumCPU(),
		Hostname: hostname,
		Username: GetUsername(),
		HomeDir:  homeDir,
		WorkDir:  workDir,
	}
}