package core

import (
	"reflect"
	"strings"
	"sync"
)

// ControllerFactory Controller工厂接口
type ControllerFactory interface {
	CreateController(controllerType reflect.Type) IController
	InitController(controller IController) error
}

// DefaultControllerFactory 默认Controller工厂实现
type DefaultControllerFactory struct{}

// CreateController 创建Controller实例（类似Beego的反射实例化）
func (f *DefaultControllerFactory) CreateController(controllerType reflect.Type) IController {
	// 通过反射创建Controller实例
	if controllerType.Kind() == reflect.Ptr {
		controllerType = controllerType.Elem()
	}

	// 创建实例
	controllerValue := reflect.New(controllerType)
	controller := controllerValue.Interface().(IController)

	// 自动初始化
	f.InitController(controller)

	return controller
}

// InitController 初始化Controller（类似Beego的Init方法）
func (f *DefaultControllerFactory) InitController(controller IController) error {
	// 通过反射获取Controller类型信息
	controllerType := reflect.TypeOf(controller)
	if controllerType.Kind() == reflect.Ptr {
		controllerType = controllerType.Elem()
	}

	// 提取Controller名称
	controllerName := extractControllerNameFromType(controllerType)

	// 获取BaseController并设置信息
	if baseController := getBaseControllerFromInterface(controller); baseController != nil {
		baseController.initFromFactory(controllerName, controller)
		return nil
	}

	return nil
}

// ============= Controller管理器 =============

// ControllerManager Controller管理器（类似Beego的ControllerRegister）
type ControllerManager struct {
	factory       ControllerFactory
	controllerMap map[string]reflect.Type // 类型注册表
}

// NewControllerManager 创建Controller管理器
func NewControllerManager() *ControllerManager {
	return &ControllerManager{
		factory:       &DefaultControllerFactory{},
		controllerMap: make(map[string]reflect.Type),
	}
}

// RegisterController 注册Controller类型（类似Beego的路由注册）
func (cm *ControllerManager) RegisterController(name string, controller IController) {
	controllerType := reflect.TypeOf(controller)
	if controllerType.Kind() == reflect.Ptr {
		controllerType = controllerType.Elem()
	}

	cm.controllerMap[name] = controllerType
}

// NewController 创建指定名称的Controller（Beego风格）
func (cm *ControllerManager) NewController(name string) IController {
	if controllerType, exists := cm.controllerMap[name]; exists {
		return cm.factory.CreateController(controllerType)
	}
	return nil
}

// NewControllerByType 根据类型创建Controller
func (cm *ControllerManager) NewControllerByType(controllerType reflect.Type) IController {
	return cm.factory.CreateController(controllerType)
}

// ============= BaseController扩展方法 =============

// initFromFactory 由工厂调用的初始化方法（类似Beego的Init）
func (c *BaseController) initFromFactory(controllerName string, appController IController) {
	c.ControllerName = controllerName
	c.AppController = appController
	c.initialized = true

	// 创建默认方法映射
	if len(c.MethodMapping) == 0 {
		c.MethodMapping = CreateDefaultMethodMapping(appController)
	}

	// 初始化数据映射
	if c.Data == nil {
		c.Data = make(map[string]any)
	}
}

// ============= 便捷的构造函数生成器 =============

// ControllerConstructor Controller构造函数类型
type ControllerConstructor func() IController

// CreateConstructor 为Controller类型创建构造函数
func CreateConstructor(controllerType reflect.Type) ControllerConstructor {
	factory := &DefaultControllerFactory{}

	return func() IController {
		return factory.CreateController(controllerType)
	}
}

// ============= 全局Controller管理器实例 =============

var (
	defaultControllerManager *ControllerManager
	managerOnce              sync.Once
)

// GetControllerManager 获取默认Controller管理器
func GetControllerManager() *ControllerManager {
	managerOnce.Do(func() {
		defaultControllerManager = NewControllerManager()
	})
	return defaultControllerManager
}

// ============= 用户友好的API函数 =============

// RegisterControllerByName 注册Controller类型（包级别函数）
func RegisterControllerByName(name string, controller IController) {
	GetControllerManager().RegisterController(name, controller)
}

// RegisterController 注册控制器（便捷函数，从路径提取名称）
func RegisterController(pattern string, controller IController) {
	GetControllerManager().RegisterController(pattern, controller)
}

// NewControllerInstance 创建Controller实例（包级别函数）
func NewControllerInstance(name string) IController {
	return GetControllerManager().NewController(name)
}

// CreateController 创建控制器（便捷函数）
func CreateController(pattern string) IController {
	return GetControllerManager().NewController(pattern)
}

// CreateControllerByType 根据类型创建Controller（包级别函数）
func CreateControllerByType(controllerType reflect.Type) IController {
	return GetControllerManager().NewControllerByType(controllerType)
}

// ============= 智能构造函数生成器 =============

// GenerateConstructors 为给定的Controller类型生成构造函数
func GenerateConstructors(controllers map[string]IController) map[string]ControllerConstructor {
	constructors := make(map[string]ControllerConstructor)

	for name, controller := range controllers {
		controllerType := reflect.TypeOf(controller)
		constructors[name] = CreateConstructor(controllerType)

		// 同时注册到全局管理器
		RegisterControllerByName(name, controller)
	}

	return constructors
}

// ============= 辅助函数 =============

// extractControllerNameFromType 从类型中提取Controller名称
func extractControllerNameFromType(controllerType reflect.Type) string {
	typeName := controllerType.Name()

	// 移除Controller后缀
	if strings.HasSuffix(typeName, "Controller") {
		return typeName[:len(typeName)-10]
	}

	return typeName
}

// getBaseControllerFromInterface 从Controller接口获取BaseController
func getBaseControllerFromInterface(controller IController) *BaseController {
	// 通过反射查找嵌入的BaseController
	controllerValue := reflect.ValueOf(controller)
	if controllerValue.Kind() == reflect.Ptr {
		controllerValue = controllerValue.Elem()
	}

	// 查找BaseController字段
	for i := 0; i < controllerValue.NumField(); i++ {
		field := controllerValue.Field(i)
		fieldType := field.Type()

		// 检查是否是BaseController或其指针
		if fieldType == reflect.TypeOf(BaseController{}) {
			return field.Addr().Interface().(*BaseController)
		} else if fieldType == reflect.TypeOf(&BaseController{}) {
			return field.Interface().(*BaseController)
		}
	}

	return nil
}
