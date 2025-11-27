// Package testing 提供单元测试工具集
//
// 这个包提供了一套完整的测试工具，包括HTTP测试、数据库测试、模拟工具等
// 类似于Beego的testbench功能，但更加现代化和强大
package testing

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/zsy619/yyhertz/framework/config"
)

// TestSuite 测试套件接口
type TestSuite interface {
	// SetUp 测试前设置
	SetUp(t *testing.T)
	// TearDown 测试后清理
	TearDown(t *testing.T)
	// BeforeEach 每个测试前执行
	BeforeEach(t *testing.T)
	// AfterEach 每个测试后执行
	AfterEach(t *testing.T)
}

// BaseTestSuite 基础测试套件
type BaseTestSuite struct {
	TestData map[string]any
	mutex    sync.RWMutex
}

// NewBaseTestSuite 创建基础测试套件
func NewBaseTestSuite() *BaseTestSuite {
	return &BaseTestSuite{
		TestData: make(map[string]any),
	}
}

// SetUp 测试前设置
func (bts *BaseTestSuite) SetUp(t *testing.T) {
	bts.mutex.Lock()
	defer bts.mutex.Unlock()

	config.Infof("Setting up test suite: %s", t.Name())

	// 初始化测试数据
	bts.TestData = make(map[string]any)

	// 设置测试环境变量
	os.Setenv("TESTING", "true")
	os.Setenv("TEST_NAME", t.Name())
}

// TearDown 测试后清理
func (bts *BaseTestSuite) TearDown(t *testing.T) {
	bts.mutex.Lock()
	defer bts.mutex.Unlock()

	config.Infof("Tearing down test suite: %s", t.Name())

	// 清理测试数据
	bts.TestData = nil

	// 清理环境变量
	os.Unsetenv("TESTING")
	os.Unsetenv("TEST_NAME")
}

// BeforeEach 每个测试前执行
func (bts *BaseTestSuite) BeforeEach(t *testing.T) {
	config.Infof("Before test: %s", t.Name())
}

// AfterEach 每个测试后执行
func (bts *BaseTestSuite) AfterEach(t *testing.T) {
	config.Infof("After test: %s", t.Name())
}

// SetTestData 设置测试数据
func (bts *BaseTestSuite) SetTestData(key string, value any) {
	bts.mutex.Lock()
	defer bts.mutex.Unlock()
	bts.TestData[key] = value
}

// GetTestData 获取测试数据
func (bts *BaseTestSuite) GetTestData(key string) (any, bool) {
	bts.mutex.RLock()
	defer bts.mutex.RUnlock()
	value, exists := bts.TestData[key]
	return value, exists
}

// TestHelper 测试辅助工具
type TestHelper struct {
	t       *testing.T
	context context.Context
}

// NewTestHelper 创建测试辅助工具
func NewTestHelper(t *testing.T) *TestHelper {
	return &TestHelper{
		t:       t,
		context: context.Background(),
	}
}

// WithContext 设置上下文
func (th *TestHelper) WithContext(ctx context.Context) *TestHelper {
	th.context = ctx
	return th
}

// GetContext 获取上下文
func (th *TestHelper) GetContext() context.Context {
	return th.context
}

// AssertEqual 断言相等
func (th *TestHelper) AssertEqual(expected, actual any, msgAndArgs ...any) {
	if expected != actual {
		th.t.Helper()
		msg := th.formatMessage("Expected %v, but got %v", msgAndArgs, expected, actual)
		th.t.Errorf("%s", msg)
	}
}

// AssertNotEqual 断言不相等
func (th *TestHelper) AssertNotEqual(expected, actual any, msgAndArgs ...any) {
	if expected == actual {
		th.t.Helper()
		msg := th.formatMessage("Expected %v to not equal %v", msgAndArgs, expected, actual)
		th.t.Errorf("%s", msg)
	}
}

// AssertNil 断言为nil
func (th *TestHelper) AssertNil(value any, msgAndArgs ...any) {
	if value != nil {
		th.t.Helper()
		msg := th.formatMessage("Expected nil, but got %v", msgAndArgs, value)
		th.t.Errorf("%s", msg)
	}
}

// AssertNotNil 断言不为nil
func (th *TestHelper) AssertNotNil(value any, msgAndArgs ...any) {
	if value == nil {
		th.t.Helper()
		msg := th.formatMessage("Expected non-nil value", msgAndArgs)
		th.t.Errorf("%s", msg)
	}
}

// AssertTrue 断言为真
func (th *TestHelper) AssertTrue(condition bool, msgAndArgs ...any) {
	if !condition {
		th.t.Helper()
		msg := th.formatMessage("Expected true", msgAndArgs)
		th.t.Errorf("%s", msg)
	}
}

// AssertFalse 断言为假
func (th *TestHelper) AssertFalse(condition bool, msgAndArgs ...any) {
	if condition {
		th.t.Helper()
		msg := th.formatMessage("Expected false", msgAndArgs)
		th.t.Errorf("%s", msg)
	}
}

// AssertContains 断言包含
func (th *TestHelper) AssertContains(haystack, needle string, msgAndArgs ...any) {
	if !strings.Contains(haystack, needle) {
		th.t.Helper()
		msg := th.formatMessage("Expected '%s' to contain '%s'", msgAndArgs, haystack, needle)
		th.t.Errorf("%s", msg)
	}
}

// AssertNotContains 断言不包含
func (th *TestHelper) AssertNotContains(haystack, needle string, msgAndArgs ...any) {
	if strings.Contains(haystack, needle) {
		th.t.Helper()
		msg := th.formatMessage("Expected '%s' to not contain '%s'", msgAndArgs, haystack, needle)
		th.t.Errorf("%s", msg)
	}
}

// AssertPanic 断言会panic
func (th *TestHelper) AssertPanic(fn func(), msgAndArgs ...any) {
	defer func() {
		if r := recover(); r == nil {
			th.t.Helper()
			msg := th.formatMessage("Expected panic", msgAndArgs)
			th.t.Errorf("%s", msg)
		}
	}()
	fn()
}

// AssertNoPanic 断言不会panic
func (th *TestHelper) AssertNoPanic(fn func(), msgAndArgs ...any) {
	defer func() {
		if r := recover(); r != nil {
			th.t.Helper()
			msg := th.formatMessage("Unexpected panic: %v", msgAndArgs, r)
			th.t.Errorf("%s", msg)
		}
	}()
	fn()
}

// formatMessage 格式化消息
func (th *TestHelper) formatMessage(format string, msgAndArgs []any, values ...any) string {
	var message string

	// 如果有自定义消息
	if len(msgAndArgs) > 0 {
		if msg, ok := msgAndArgs[0].(string); ok {
			if len(msgAndArgs) > 1 {
				message = fmt.Sprintf(msg, msgAndArgs[1:]...) + " - "
			} else {
				message = msg + " - "
			}
		}
	}

	// 格式化主要消息
	return message + fmt.Sprintf(format, values...)
}

// TestConfig 测试配置
type TestConfig struct {
	// 测试数据库配置
	TestDB struct {
		Driver   string `json:"driver" yaml:"driver"`
		Host     string `json:"host" yaml:"host"`
		Port     int    `json:"port" yaml:"port"`
		Username string `json:"username" yaml:"username"`
		Password string `json:"password" yaml:"password"`
		Database string `json:"database" yaml:"database"`
	} `json:"test_db" yaml:"test_db"`

	// 测试服务器配置
	TestServer struct {
		Host string `json:"host" yaml:"host"`
		Port int    `json:"port" yaml:"port"`
	} `json:"test_server" yaml:"test_server"`

	// 测试选项
	Options struct {
		Parallel     bool          `json:"parallel" yaml:"parallel"`
		Timeout      time.Duration `json:"timeout" yaml:"timeout"`
		Cleanup      bool          `json:"cleanup" yaml:"cleanup"`
		Verbose      bool          `json:"verbose" yaml:"verbose"`
		KeepTestData bool          `json:"keep_test_data" yaml:"keep_test_data"`
	} `json:"options" yaml:"options"`
}

// DefaultTestConfig 默认测试配置
func DefaultTestConfig() *TestConfig {
	return &TestConfig{
		TestDB: struct {
			Driver   string `json:"driver" yaml:"driver"`
			Host     string `json:"host" yaml:"host"`
			Port     int    `json:"port" yaml:"port"`
			Username string `json:"username" yaml:"username"`
			Password string `json:"password" yaml:"password"`
			Database string `json:"database" yaml:"database"`
		}{
			Driver:   "sqlite",
			Host:     "localhost",
			Port:     0,
			Username: "",
			Password: "",
			Database: ":memory:",
		},
		TestServer: struct {
			Host string `json:"host" yaml:"host"`
			Port int    `json:"port" yaml:"port"`
		}{
			Host: "localhost",
			Port: 0, // 随机端口
		},
		Options: struct {
			Parallel     bool          `json:"parallel" yaml:"parallel"`
			Timeout      time.Duration `json:"timeout" yaml:"timeout"`
			Cleanup      bool          `json:"cleanup" yaml:"cleanup"`
			Verbose      bool          `json:"verbose" yaml:"verbose"`
			KeepTestData bool          `json:"keep_test_data" yaml:"keep_test_data"`
		}{
			Parallel:     false,
			Timeout:      time.Minute * 5,
			Cleanup:      true,
			Verbose:      false,
			KeepTestData: false,
		},
	}
}

// TestRunner 测试运行器
type TestRunner struct {
	config    *TestConfig
	suites    []TestSuite
	cleanupFn []func()
	mutex     sync.Mutex
}

// NewTestRunner 创建测试运行器
func NewTestRunner(config *TestConfig) *TestRunner {
	if config == nil {
		config = DefaultTestConfig()
	}

	return &TestRunner{
		config:    config,
		suites:    make([]TestSuite, 0),
		cleanupFn: make([]func(), 0),
	}
}

// AddSuite 添加测试套件
func (tr *TestRunner) AddSuite(suite TestSuite) {
	tr.mutex.Lock()
	defer tr.mutex.Unlock()
	tr.suites = append(tr.suites, suite)
}

// AddCleanup 添加清理函数
func (tr *TestRunner) AddCleanup(fn func()) {
	tr.mutex.Lock()
	defer tr.mutex.Unlock()
	tr.cleanupFn = append(tr.cleanupFn, fn)
}

// RunTest 运行单个测试
func (tr *TestRunner) RunTest(t *testing.T, testFunc func(*testing.T)) {
	// 设置超时
	if tr.config.Options.Timeout > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), tr.config.Options.Timeout)
		defer cancel()

		done := make(chan bool)
		go func() {
			tr.runTestWithSuites(t, testFunc)
			done <- true
		}()

		select {
		case <-done:
			// 测试完成
		case <-ctx.Done():
			t.Fatalf("Test timeout after %v", tr.config.Options.Timeout)
		}
	} else {
		tr.runTestWithSuites(t, testFunc)
	}
}

// runTestWithSuites 使用套件运行测试
func (tr *TestRunner) runTestWithSuites(t *testing.T, testFunc func(*testing.T)) {
	// 执行所有套件的SetUp
	for _, suite := range tr.suites {
		suite.SetUp(t)
	}

	// 添加清理函数
	defer func() {
		// 执行所有套件的TearDown
		for i := len(tr.suites) - 1; i >= 0; i-- {
			tr.suites[i].TearDown(t)
		}

		// 执行清理函数
		if tr.config.Options.Cleanup {
			tr.runCleanup()
		}
	}()

	// 执行BeforeEach
	for _, suite := range tr.suites {
		suite.BeforeEach(t)
	}

	// 添加AfterEach清理
	defer func() {
		for i := len(tr.suites) - 1; i >= 0; i-- {
			tr.suites[i].AfterEach(t)
		}
	}()

	// 运行测试
	testFunc(t)
}

// runCleanup 运行清理函数
func (tr *TestRunner) runCleanup() {
	for i := len(tr.cleanupFn) - 1; i >= 0; i-- {
		func() {
			defer func() {
				if r := recover(); r != nil {
					config.Errorf("Cleanup function panicked: %v", r)
				}
			}()
			tr.cleanupFn[i]()
		}()
	}
}

// FileHelper 文件测试辅助工具
type FileHelper struct {
	tempDir string
	files   []string
	mutex   sync.Mutex
}

// NewFileHelper 创建文件辅助工具
func NewFileHelper() *FileHelper {
	tempDir, _ := os.MkdirTemp("", "yyhertz_test_*")
	return &FileHelper{
		tempDir: tempDir,
		files:   make([]string, 0),
	}
}

// CreateTempFile 创建临时文件
func (fh *FileHelper) CreateTempFile(content string) (string, error) {
	fh.mutex.Lock()
	defer fh.mutex.Unlock()

	file, err := os.CreateTemp(fh.tempDir, "test_*.tmp")
	if err != nil {
		return "", err
	}
	defer file.Close()

	if content != "" {
		if _, err := file.WriteString(content); err != nil {
			return "", err
		}
	}

	filepath := file.Name()
	fh.files = append(fh.files, filepath)
	return filepath, nil
}

// CreateTempDir 创建临时目录
func (fh *FileHelper) CreateTempDir() (string, error) {
	fh.mutex.Lock()
	defer fh.mutex.Unlock()

	dir, err := os.MkdirTemp(fh.tempDir, "test_dir_*")
	if err != nil {
		return "", err
	}

	fh.files = append(fh.files, dir)
	return dir, nil
}

// WriteFile 写入文件
func (fh *FileHelper) WriteFile(filename, content string) error {
	path := filepath.Join(fh.tempDir, filename)

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	fh.mutex.Lock()
	fh.files = append(fh.files, path)
	fh.mutex.Unlock()

	return os.WriteFile(path, []byte(content), 0644)
}

// ReadFile 读取文件
func (fh *FileHelper) ReadFile(filename string) (string, error) {
	path := filepath.Join(fh.tempDir, filename)
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// GetTempDir 获取临时目录
func (fh *FileHelper) GetTempDir() string {
	return fh.tempDir
}

// Cleanup 清理临时文件
func (fh *FileHelper) Cleanup() {
	fh.mutex.Lock()
	defer fh.mutex.Unlock()

	for _, file := range fh.files {
		os.RemoveAll(file)
	}

	if fh.tempDir != "" {
		os.RemoveAll(fh.tempDir)
	}
}

// CaptureOutput 捕获输出
func CaptureOutput(fn func()) (string, error) {
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}

	os.Stdout = w

	outC := make(chan string)
	go func() {
		var buf strings.Builder
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	fn()

	w.Close()
	os.Stdout = old
	output := <-outC

	return output, nil
}

// GetCurrentTestName 获取当前测试名称
func GetCurrentTestName() string {
	pc, _, _, ok := runtime.Caller(2)
	if !ok {
		return "unknown"
	}

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "unknown"
	}

	name := fn.Name()
	if idx := strings.LastIndex(name, "."); idx >= 0 {
		name = name[idx+1:]
	}

	return name
}

// SkipCI 在CI环境中跳过测试
func SkipCI(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping test in CI environment")
	}
}

// SkipShort 在短测试模式下跳过
func SkipShort(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
}

// ============= 全局测试实例 =============

var (
	globalTestRunner *TestRunner
	testRunnerOnce   sync.Once
)

// GetGlobalTestRunner 获取全局测试运行器
func GetGlobalTestRunner() *TestRunner {
	testRunnerOnce.Do(func() {
		globalTestRunner = NewTestRunner(DefaultTestConfig())
	})
	return globalTestRunner
}

// RunTest 使用全局运行器运行测试
func RunTest(t *testing.T, testFunc func(*testing.T)) {
	GetGlobalTestRunner().RunTest(t, testFunc)
}

// AddGlobalSuite 添加全局测试套件
func AddGlobalSuite(suite TestSuite) {
	GetGlobalTestRunner().AddSuite(suite)
}

// AddGlobalCleanup 添加全局清理函数
func AddGlobalCleanup(fn func()) {
	GetGlobalTestRunner().AddCleanup(fn)
}
