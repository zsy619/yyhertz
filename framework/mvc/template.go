package mvc

func init() {
	LogDebug("开始加载系统模板函数。。。")

	AddFuncMap("dateformat", DateFormat)
	AddFuncMap("date", Date)
	AddFuncMap("compare", Compare)
	AddFuncMap("compare_not", CompareNot)
	AddFuncMap("not_nil", NotNil)
	AddFuncMap("not_null", NotNil)
	AddFuncMap("substr", Substr)
	AddFuncMap("html2str", HTML2str)
	AddFuncMap("str2html", Str2html)
	AddFuncMap("htmlquote", Htmlquote)
	AddFuncMap("htmlunquote", Htmlunquote)
	AddFuncMap("renderform", RenderForm)
	AddFuncMap("assets_js", AssetsJs)
	AddFuncMap("assets_css", AssetsCSS)
	AddFuncMap("config", GetConfig)
	AddFuncMap("map_get", MapGet)

	// Comparisons
	AddFuncMap("eq", eq) // ==
	AddFuncMap("ge", ge) // >=
	AddFuncMap("gt", gt) // >
	AddFuncMap("le", le) // <=
	AddFuncMap("lt", lt) // <
	AddFuncMap("ne", ne) // !=

	AddFuncMap("urlfor", URLFor)

	LogDebug("结束加载系统模板函数。。。")
}
