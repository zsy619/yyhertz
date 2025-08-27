// Package mybatis 懒加载和代理对象实现
//
// 实现MyBatis风格的关联查询懒加载机制
package mybatis

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// LazyLoadConfiguration 懒加载配置
type LazyLoadConfiguration struct {
	LazyLoadingEnabled      bool     `json:"lazyLoadingEnabled"`      // 是否启用懒加载
	AggressiveLazyLoading   bool     `json:"aggressiveLazyLoading"`   // 积极懒加载模式
	LazyLoadTriggerMethods  []string `json:"lazyLoadTriggerMethods"`  // 触发懒加载的方法
	ProxyFactory            string   `json:"proxyFactory"`            // 代理工厂类型
}

// DefaultLazyLoadConfiguration 默认懒加载配置
func DefaultLazyLoadConfiguration() *LazyLoadConfiguration {
	return &LazyLoadConfiguration{
		LazyLoadingEnabled:     true,
		AggressiveLazyLoading:  false,
		LazyLoadTriggerMethods: []string{"equals", "clone", "hashCode", "toString"},
		ProxyFactory:           "REFLECTION",
	}
}

// LazyLoader 懒加载器接口
type LazyLoader interface {
	// LoadProperty 加载指定属性
	LoadProperty(ctx context.Context, property string) error
	// IsLoaded 检查属性是否已加载
	IsLoaded(property string) bool
	// LoadAll 加载所有未加载的属性
	LoadAll(ctx context.Context) error
}

// PropertyLoader 属性加载器
type PropertyLoader func(ctx context.Context) (any, error)

// LazyLoadProxy 懒加载代理
type LazyLoadProxy struct {
	target         any                        // 目标对象
	session        any                        // 会话对象
	loadedProps    map[string]bool            // 已加载属性
	loaders        map[string]PropertyLoader  // 属性加载器
	mutex          sync.RWMutex               // 读写锁
	config         *LazyLoadConfiguration    // 懒加载配置
}

// NewLazyLoadProxy 创建懒加载代理
func NewLazyLoadProxy(target any, session any, config *LazyLoadConfiguration) *LazyLoadProxy {
	if config == nil {
		config = DefaultLazyLoadConfiguration()
	}
	
	return &LazyLoadProxy{
		target:      target,
		session:     session,
		loadedProps: make(map[string]bool),
		loaders:     make(map[string]PropertyLoader),
		config:      config,
	}
}

// RegisterLoader 注册属性加载器
func (p *LazyLoadProxy) RegisterLoader(property string, loader PropertyLoader) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.loaders[property] = loader
}

// LoadProperty 加载指定属性
func (p *LazyLoadProxy) LoadProperty(ctx context.Context, property string) error {
	if !p.config.LazyLoadingEnabled {
		return nil
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.loadedProps[property] {
		return nil // 已经加载过
	}

	loader, exists := p.loaders[property]
	if !exists {
		return fmt.Errorf("no loader found for property: %s", property)
	}

	value, err := loader(ctx)
	if err != nil {
		return fmt.Errorf("failed to load property %s: %w", property, err)
	}

	// 设置属性值到目标对象
	err = p.setPropertyValue(property, value)
	if err != nil {
		return fmt.Errorf("failed to set property %s: %w", property, err)
	}

	p.loadedProps[property] = true
	return nil
}

// IsLoaded 检查属性是否已加载
func (p *LazyLoadProxy) IsLoaded(property string) bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.loadedProps[property]
}

// LoadAll 加载所有未加载的属性
func (p *LazyLoadProxy) LoadAll(ctx context.Context) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for property, loader := range p.loaders {
		if !p.loadedProps[property] {
			value, err := loader(ctx)
			if err != nil {
				return fmt.Errorf("failed to load property %s: %w", property, err)
			}

			err = p.setPropertyValue(property, value)
			if err != nil {
				return fmt.Errorf("failed to set property %s: %w", property, err)
			}

			p.loadedProps[property] = true
		}
	}

	return nil
}

// setPropertyValue 设置属性值
func (p *LazyLoadProxy) setPropertyValue(property string, value any) error {
	targetValue := reflect.ValueOf(p.target)
	if targetValue.Kind() == reflect.Ptr {
		targetValue = targetValue.Elem()
	}

	if targetValue.Kind() != reflect.Struct {
		return fmt.Errorf("target must be a struct, got %T", p.target)
	}

	fieldValue := targetValue.FieldByName(strings.Title(property))
	if !fieldValue.IsValid() {
		return fmt.Errorf("field %s not found", property)
	}

	if !fieldValue.CanSet() {
		return fmt.Errorf("field %s cannot be set", property)
	}

	valueReflect := reflect.ValueOf(value)
	if !valueReflect.Type().AssignableTo(fieldValue.Type()) {
		return fmt.Errorf("value type %T is not assignable to field type %T", value, fieldValue.Interface())
	}

	fieldValue.Set(valueReflect)
	return nil
}

// ProxyObject 代理对象接口
type ProxyObject interface {
	GetLazyLoader() LazyLoader
	TriggerLoad(ctx context.Context, method string) error
}

// AssociationType 关联类型
type AssociationType int

const (
	AssociationTypeOne AssociationType = iota // 一对一
	AssociationTypeMany                       // 一对多
)

// AssociationMapping 关联映射
type AssociationMapping struct {
	Property     string            `json:"property"`     // 属性名
	Column       string            `json:"column"`       // 关联列
	Select       string            `json:"select"`       // 查询语句ID
	ForeignKey   string            `json:"foreignKey"`   // 外键字段
	Type         AssociationType   `json:"type"`         // 关联类型
	ResultType   reflect.Type      `json:"resultType"`   // 结果类型
	LazyLoad     bool              `json:"lazyLoad"`     // 是否懒加载
}

// LazyLoadManager 懒加载管理器
type LazyLoadManager struct {
	associations map[string][]*AssociationMapping // 类型 -> 关联映射
	proxies      map[any]*LazyLoadProxy           // 对象 -> 代理
	mutex        sync.RWMutex
	config       *LazyLoadConfiguration
}

// NewLazyLoadManager 创建懒加载管理器
func NewLazyLoadManager(config *LazyLoadConfiguration) *LazyLoadManager {
	if config == nil {
		config = DefaultLazyLoadConfiguration()
	}
	
	return &LazyLoadManager{
		associations: make(map[string][]*AssociationMapping),
		proxies:      make(map[any]*LazyLoadProxy),
		config:       config,
	}
}

// RegisterAssociation 注册关联映射
func (m *LazyLoadManager) RegisterAssociation(typeName string, mapping *AssociationMapping) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.associations[typeName] = append(m.associations[typeName], mapping)
}

// CreateProxy 为对象创建懒加载代理
func (m *LazyLoadManager) CreateProxy(obj any, session any) (any, error) {
	if !m.config.LazyLoadingEnabled {
		return obj, nil
	}

	objType := reflect.TypeOf(obj)
	if objType.Kind() == reflect.Ptr {
		objType = objType.Elem()
	}

	typeName := objType.Name()
	associations, exists := m.associations[typeName]
	if !exists || len(associations) == 0 {
		return obj, nil // 没有关联映射，返回原对象
	}

	proxy := NewLazyLoadProxy(obj, session, m.config)

	// 注册属性加载器
	for _, assoc := range associations {
		if assoc.LazyLoad {
			loader := m.createPropertyLoader(assoc, obj, session)
			proxy.RegisterLoader(assoc.Property, loader)
		}
	}

	m.mutex.Lock()
	m.proxies[obj] = proxy
	m.mutex.Unlock()

	return &ProxyWrapper{
		target: obj,
		proxy:  proxy,
	}, nil
}

// createPropertyLoader 创建属性加载器
func (m *LazyLoadManager) createPropertyLoader(mapping *AssociationMapping, parent any, session any) PropertyLoader {
	return func(ctx context.Context) (any, error) {
		// 获取外键值
		foreignKeyValue, err := m.getForeignKeyValue(parent, mapping.ForeignKey)
		if err != nil {
			return nil, fmt.Errorf("failed to get foreign key value: %w", err)
		}

		// 执行关联查询
		// 使用类型断言来处理不同的会话类型
		if simpleSession, ok := session.(SimpleSession); ok {
			switch mapping.Type {
			case AssociationTypeOne:
				return simpleSession.SelectOne(ctx, mapping.Select, foreignKeyValue)
			case AssociationTypeMany:
				return simpleSession.SelectList(ctx, mapping.Select, foreignKeyValue)
			default:
				return nil, fmt.Errorf("unsupported association type: %v", mapping.Type)
			}
		}
		
		if xmlSession, ok := session.(XMLSession); ok {
			switch mapping.Type {
			case AssociationTypeOne:
				return xmlSession.SelectOne(ctx, mapping.Select, foreignKeyValue)
			case AssociationTypeMany:
				return xmlSession.SelectList(ctx, mapping.Select, foreignKeyValue)
			default:
				return nil, fmt.Errorf("unsupported association type: %v", mapping.Type)
			}
		}
		
		return nil, fmt.Errorf("unsupported session type: %T", session)
	}
}

// getForeignKeyValue 获取外键值
func (m *LazyLoadManager) getForeignKeyValue(obj any, foreignKey string) (any, error) {
	objValue := reflect.ValueOf(obj)
	if objValue.Kind() == reflect.Ptr {
		objValue = objValue.Elem()
	}

	fieldValue := objValue.FieldByName(strings.Title(foreignKey))
	if !fieldValue.IsValid() {
		return nil, fmt.Errorf("foreign key field %s not found", foreignKey)
	}

	return fieldValue.Interface(), nil
}

// ProxyWrapper 代理包装器
type ProxyWrapper struct {
	target any
	proxy  *LazyLoadProxy
}

// GetLazyLoader 获取懒加载器
func (w *ProxyWrapper) GetLazyLoader() LazyLoader {
	return w.proxy
}

// TriggerLoad 触发加载
func (w *ProxyWrapper) TriggerLoad(ctx context.Context, method string) error {
	// 检查是否为触发方法
	for _, triggerMethod := range w.proxy.config.LazyLoadTriggerMethods {
		if strings.EqualFold(method, triggerMethod) {
			return w.proxy.LoadAll(ctx)
		}
	}
	return nil
}

// GetTarget 获取目标对象
func (w *ProxyWrapper) GetTarget() any {
	return w.target
}

// LazyLoadingExecutor 懒加载执行器
type LazyLoadingExecutor struct {
	session any
	manager *LazyLoadManager
}

// NewLazyLoadingExecutor 创建懒加载执行器
func NewLazyLoadingExecutor(session any, manager *LazyLoadManager) *LazyLoadingExecutor {
	return &LazyLoadingExecutor{
		session: session,
		manager: manager,
	}
}

// ProcessResult 处理查询结果，创建懒加载代理
func (e *LazyLoadingExecutor) ProcessResult(ctx context.Context, result any) (any, error) {
	resultValue := reflect.ValueOf(result)

	switch resultValue.Kind() {
	case reflect.Slice:
		// 处理列表结果
		return e.processSliceResult(ctx, result)
	case reflect.Ptr, reflect.Struct:
		// 处理单个对象
		return e.manager.CreateProxy(result, e.session)
	default:
		return result, nil
	}
}

// processSliceResult 处理切片结果
func (e *LazyLoadingExecutor) processSliceResult(ctx context.Context, result any) (any, error) {
	resultValue := reflect.ValueOf(result)
	if resultValue.Kind() != reflect.Slice {
		return result, nil
	}

	length := resultValue.Len()
	if length == 0 {
		return result, nil
	}

	// 创建新的切片来存储代理对象
	newSlice := reflect.MakeSlice(resultValue.Type(), length, length)

	for i := 0; i < length; i++ {
		item := resultValue.Index(i).Interface()
		proxy, err := e.manager.CreateProxy(item, e.session)
		if err != nil {
			return nil, fmt.Errorf("failed to create proxy for item %d: %w", i, err)
		}
		newSlice.Index(i).Set(reflect.ValueOf(proxy))
	}

	return newSlice.Interface(), nil
}

// Association 关联查询辅助结构
type Association struct {
	Property   string                 `json:"property"`
	Column     string                 `json:"column"`
	Select     string                 `json:"select"`
	ResultType string                 `json:"resultType"`
	LazyLoad   bool                   `json:"lazyLoad"`
	JoinColumn string                 `json:"joinColumn"`
	FetchType  string                 `json:"fetchType"` // LAZY 或 EAGER
}

// Collection 集合关联查询辅助结构
type Collection struct {
	Property   string `json:"property"`
	OfType     string `json:"ofType"`
	Column     string `json:"column"`
	Select     string `json:"select"`
	LazyLoad   bool   `json:"lazyLoad"`
	FetchType  string `json:"fetchType"` // LAZY 或 EAGER
}