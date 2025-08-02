package mvc

// 重新导出BaseController，保持向后兼容
import "github.com/zsy619/yyhertz/framework/mvc/core"

// 类型别名，保持向后兼容
type BaseController = core.BaseController

// 重新导出构造函数
var (
	NewBaseController            = core.NewBaseController
	NewBaseControllerWithContext = core.NewBaseControllerWithContext
)