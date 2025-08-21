package mvc

import "github.com/zsy619/yyhertz/framework/mvc/view"

func init() {
	LogDebug("开始加载系统模板函数。。。")

	for name, fn := range view.GetBuiltinTemplateFunctions() {
		LogDebugf("添加模板函数 %s 成功 %v", name, fn)
	}

	LogDebug("结束加载系统模板函数。。。")
}
