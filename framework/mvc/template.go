package mvc

import "github.com/zsy619/yyhertz/framework/mvc/view"

func init() {
	LogDebug("开始加载系统模板函数。。。")

	for name, fn := range view.TemplateFuncs {
		if err := AddFuncMap(name, fn); err != nil {
			LogErrorf("添加模板函数 %s 失败: %v", name, err)
		}
	}

	LogDebug("结束加载系统模板函数。。。")
}
