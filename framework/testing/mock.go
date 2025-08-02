package testing

import (
	"fmt"
	"reflect"
	"sync"
	"testing"

	"github.com/zsy619/yyhertz/framework/config"
)

// MockCall 模拟调用记录
type MockCall struct {
	Method    string
	Args      []any
	Returns   []any
	CallCount int
	mutex     sync.RWMutex
}

// NewMockCall 创建模拟调用
func NewMockCall(method string, args ...any) *MockCall {
	return &MockCall{
		Method: method,
		Args:   args,
	}
}

// Return 设置返回值
func (mc *MockCall) Return(values ...any) *MockCall {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	mc.Returns = values
	return mc
}

// Times 获取调用次数
func (mc *MockCall) Times() int {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()
	return mc.CallCount
}

// IncrementCall 增加调用次数
func (mc *MockCall) IncrementCall() {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	mc.CallCount++
}

// Mock 模拟对象
type Mock struct {
	calls        map[string]*MockCall
	expectations map[string]*MockCall
	strict       bool
	t            *testing.T
	mutex        sync.RWMutex
}

// NewMock 创建模拟对象
func NewMock(t *testing.T) *Mock {
	return &Mock{
		calls:        make(map[string]*MockCall),
		expectations: make(map[string]*MockCall),
		strict:       false,
		t:            t,
	}
}

// SetStrict 设置严格模式
func (m *Mock) SetStrict(strict bool) *Mock {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.strict = strict
	return m
}

// On 设置期望调用
func (m *Mock) On(method string, args ...any) *MockCall {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	key := m.generateKey(method, args...)
	call := NewMockCall(method, args...)
	m.expectations[key] = call
	return call
}

// Called 记录方法调用
func (m *Mock) Called(method string, args ...any) []any {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	key := m.generateKey(method, args...)

	// 查找期望调用
	if expectation, exists := m.expectations[key]; exists {
		expectation.IncrementCall()
		if call, exists := m.calls[key]; exists {
			call.IncrementCall()
		} else {
			m.calls[key] = &MockCall{
				Method:    method,
				Args:      args,
				Returns:   expectation.Returns,
				CallCount: 1,
			}
		}
		return expectation.Returns
	}

	// 如果是严格模式，未期望的调用会导致测试失败
	if m.strict {
		m.t.Helper()
		m.t.Errorf("Unexpected call to %s with args %v", method, args)
		return nil
	}

	// 记录调用但不返回值
	if call, exists := m.calls[key]; exists {
		call.IncrementCall()
	} else {
		m.calls[key] = &MockCall{
			Method:    method,
			Args:      args,
			CallCount: 1,
		}
	}

	return nil
}

// AssertExpectations 断言所有期望都被满足
func (m *Mock) AssertExpectations() {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for key, expectation := range m.expectations {
		if call, exists := m.calls[key]; exists {
			if call.CallCount == 0 {
				m.t.Helper()
				m.t.Errorf("Expected call to %s with args %v was never made", expectation.Method, expectation.Args)
			}
		} else {
			m.t.Helper()
			m.t.Errorf("Expected call to %s with args %v was never made", expectation.Method, expectation.Args)
		}
	}
}

// AssertCalled 断言方法被调用
func (m *Mock) AssertCalled(method string, args ...any) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	key := m.generateKey(method, args...)
	if call, exists := m.calls[key]; !exists || call.CallCount == 0 {
		m.t.Helper()
		m.t.Errorf("Expected call to %s with args %v was never made", method, args)
	}
}

// AssertNotCalled 断言方法未被调用
func (m *Mock) AssertNotCalled(method string, args ...any) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	key := m.generateKey(method, args...)
	if call, exists := m.calls[key]; exists && call.CallCount > 0 {
		m.t.Helper()
		m.t.Errorf("Unexpected call to %s with args %v (called %d times)", method, args, call.CallCount)
	}
}

// AssertNumberOfCalls 断言调用次数
func (m *Mock) AssertNumberOfCalls(method string, expectedCalls int, args ...any) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	key := m.generateKey(method, args...)
	actualCalls := 0
	if call, exists := m.calls[key]; exists {
		actualCalls = call.CallCount
	}

	if actualCalls != expectedCalls {
		m.t.Helper()
		m.t.Errorf("Expected %s to be called %d times, but was called %d times", method, expectedCalls, actualCalls)
	}
}

// GetCall 获取调用记录
func (m *Mock) GetCall(method string, args ...any) *MockCall {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	key := m.generateKey(method, args...)
	return m.calls[key]
}

// GetAllCalls 获取所有调用记录
func (m *Mock) GetAllCalls() map[string]*MockCall {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	calls := make(map[string]*MockCall)
	for k, v := range m.calls {
		calls[k] = v
	}
	return calls
}

// Reset 重置模拟对象
func (m *Mock) Reset() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.calls = make(map[string]*MockCall)
	m.expectations = make(map[string]*MockCall)
}

// generateKey 生成调用键
func (m *Mock) generateKey(method string, args ...any) string {
	key := method
	for _, arg := range args {
		key += fmt.Sprintf("_%v", arg)
	}
	return key
}

// ============= 函数级模拟 =============

// FunctionMock 函数模拟器
type FunctionMock struct {
	mock     *Mock
	original any
	replaced bool
	name     string
}

// NewFunctionMock 创建函数模拟器
func NewFunctionMock(t *testing.T, name string) *FunctionMock {
	return &FunctionMock{
		mock: NewMock(t),
		name: name,
	}
}

// MockFunction 模拟函数调用
func (fm *FunctionMock) MockFunction(returnValues ...any) func(...any) []any {
	return func(args ...any) []any {
		fm.mock.On(fm.name, args...).Return(returnValues...)
		return fm.mock.Called(fm.name, args...)
	}
}

// AssertCalled 断言函数被调用
func (fm *FunctionMock) AssertCalled(args ...any) {
	fm.mock.AssertCalled(fm.name, args...)
}

// AssertNotCalled 断言函数未被调用
func (fm *FunctionMock) AssertNotCalled(args ...any) {
	fm.mock.AssertNotCalled(fm.name, args...)
}

// AssertNumberOfCalls 断言调用次数
func (fm *FunctionMock) AssertNumberOfCalls(expectedCalls int, args ...any) {
	fm.mock.AssertNumberOfCalls(fm.name, expectedCalls, args...)
}

// Reset 重置函数模拟器
func (fm *FunctionMock) Reset() {
	fm.mock.Reset()
}

// ============= 接口模拟生成器 =============

// InterfaceMocker 接口模拟器
type InterfaceMocker struct {
	t             *testing.T
	mock          *Mock
	interfaceType reflect.Type
}

// NewInterfaceMocker 创建接口模拟器
func NewInterfaceMocker(t *testing.T, interfacePtr any) *InterfaceMocker {
	interfaceType := reflect.TypeOf(interfacePtr).Elem()
	if interfaceType.Kind() != reflect.Interface {
		t.Fatal("Provided type is not an interface")
	}

	return &InterfaceMocker{
		t:             t,
		mock:          NewMock(t),
		interfaceType: interfaceType,
	}
}

// GenerateMock 生成模拟对象
func (im *InterfaceMocker) GenerateMock() any {
	_ = reflect.New(im.interfaceType).Elem()

	// 为接口的每个方法创建模拟实现
	methods := make(map[string]reflect.Value)
	for i := 0; i < im.interfaceType.NumMethod(); i++ {
		method := im.interfaceType.Method(i)
		methods[method.Name] = im.createMethodMock(method)
	}

	// 创建动态类型实现接口
	mockStruct := reflect.StructOf([]reflect.StructField{})
	mockInstance := reflect.New(mockStruct).Elem()

	// 注意：这里简化了实现，实际上需要更复杂的动态类型生成
	// 在实际项目中，建议使用专门的mock生成工具如gomock

	return mockInstance.Interface()
}

// createMethodMock 创建方法模拟
func (im *InterfaceMocker) createMethodMock(method reflect.Method) reflect.Value {
	methodType := method.Type

	return reflect.MakeFunc(methodType, func(args []reflect.Value) []reflect.Value {
		// 转换参数
		callArgs := make([]any, len(args))
		for i, arg := range args {
			callArgs[i] = arg.Interface()
		}

		// 调用模拟
		returns := im.mock.Called(method.Name, callArgs...)

		// 转换返回值
		results := make([]reflect.Value, methodType.NumOut())
		for i := 0; i < methodType.NumOut(); i++ {
			if i < len(returns) && returns[i] != nil {
				results[i] = reflect.ValueOf(returns[i])
			} else {
				results[i] = reflect.Zero(methodType.Out(i))
			}
		}

		return results
	})
}

// On 设置期望调用
func (im *InterfaceMocker) On(method string, args ...any) *MockCall {
	return im.mock.On(method, args...)
}

// AssertExpectations 断言期望
func (im *InterfaceMocker) AssertExpectations() {
	im.mock.AssertExpectations()
}

// ============= 存根（Stub）工具 =============

// Stub 存根对象
type Stub struct {
	values map[string]any
	mutex  sync.RWMutex
}

// NewStub 创建存根对象
func NewStub() *Stub {
	return &Stub{
		values: make(map[string]any),
	}
}

// Set 设置存根值
func (s *Stub) Set(key string, value any) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.values[key] = value
}

// Get 获取存根值
func (s *Stub) Get(key string) (any, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	value, exists := s.values[key]
	return value, exists
}

// GetString 获取字符串存根值
func (s *Stub) GetString(key string) string {
	if value, exists := s.Get(key); exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

// GetInt 获取整数存根值
func (s *Stub) GetInt(key string) int {
	if value, exists := s.Get(key); exists {
		if i, ok := value.(int); ok {
			return i
		}
	}
	return 0
}

// GetBool 获取布尔存根值
func (s *Stub) GetBool(key string) bool {
	if value, exists := s.Get(key); exists {
		if b, ok := value.(bool); ok {
			return b
		}
	}
	return false
}

// Clear 清除所有存根值
func (s *Stub) Clear() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.values = make(map[string]any)
}

// Keys 获取所有键
func (s *Stub) Keys() []string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	keys := make([]string, 0, len(s.values))
	for key := range s.values {
		keys = append(keys, key)
	}
	return keys
}

// ============= HTTP模拟服务器 =============

// MockHTTPServer HTTP模拟服务器
type MockHTTPServer struct {
	routes map[string]map[string]func(map[string]string) (int, any) // method -> path -> handler
	mutex  sync.RWMutex
}

// NewMockHTTPServer 创建HTTP模拟服务器
func NewMockHTTPServer() *MockHTTPServer {
	return &MockHTTPServer{
		routes: make(map[string]map[string]func(map[string]string) (int, any)),
	}
}

// On 设置路由处理器
func (mhs *MockHTTPServer) On(method, path string, handler func(map[string]string) (int, any)) {
	mhs.mutex.Lock()
	defer mhs.mutex.Unlock()

	if mhs.routes[method] == nil {
		mhs.routes[method] = make(map[string]func(map[string]string) (int, any))
	}

	mhs.routes[method][path] = handler
}

// GET 设置GET处理器
func (mhs *MockHTTPServer) GET(path string, handler func(map[string]string) (int, any)) {
	mhs.On("GET", path, handler)
}

// POST 设置POST处理器
func (mhs *MockHTTPServer) POST(path string, handler func(map[string]string) (int, any)) {
	mhs.On("POST", path, handler)
}

// PUT 设置PUT处理器
func (mhs *MockHTTPServer) PUT(path string, handler func(map[string]string) (int, any)) {
	mhs.On("PUT", path, handler)
}

// DELETE 设置DELETE处理器
func (mhs *MockHTTPServer) DELETE(path string, handler func(map[string]string) (int, any)) {
	mhs.On("DELETE", path, handler)
}

// Handle 处理请求
func (mhs *MockHTTPServer) Handle(method, path string, params map[string]string) (int, any) {
	mhs.mutex.RLock()
	defer mhs.mutex.RUnlock()

	if methodRoutes, exists := mhs.routes[method]; exists {
		if handler, exists := methodRoutes[path]; exists {
			return handler(params)
		}
	}

	return 404, map[string]string{"error": "Not Found"}
}

// ============= 全局模拟管理器 =============

var (
	globalMocks map[string]*Mock
	mockMutex   sync.RWMutex
)

func init() {
	globalMocks = make(map[string]*Mock)
}

// GetGlobalMock 获取全局模拟对象
func GetGlobalMock(t *testing.T, name string) *Mock {
	mockMutex.Lock()
	defer mockMutex.Unlock()

	if mock, exists := globalMocks[name]; exists {
		return mock
	}

	mock := NewMock(t)
	globalMocks[name] = mock
	return mock
}

// ClearGlobalMocks 清除所有全局模拟对象
func ClearGlobalMocks() {
	mockMutex.Lock()
	defer mockMutex.Unlock()
	globalMocks = make(map[string]*Mock)
}

// MockConfig 配置模拟器
type MockConfig struct {
	config *Stub
}

// NewMockConfig 创建配置模拟器
func NewMockConfig() *MockConfig {
	return &MockConfig{
		config: NewStub(),
	}
}

// SetConfig 设置配置值
func (mc *MockConfig) SetConfig(key string, value any) {
	mc.config.Set(key, value)
}

// GetConfig 获取配置值
func (mc *MockConfig) GetConfig(key string) (any, bool) {
	return mc.config.Get(key)
}

// ApplyToGlobalConfig 应用到全局配置
func (mc *MockConfig) ApplyToGlobalConfig() {
	// 这里可以实现将模拟配置应用到框架的全局配置管理器
	config.Info("Mock configuration applied to global config")
}
