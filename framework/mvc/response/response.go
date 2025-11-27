package response

import (
	"fmt"

	"github.com/zsy619/yyhertz/framework/constant"
	"github.com/zsy619/yyhertz/framework/errors"
)

// =============== 响应类型定义 ===============

// JSONResponse 标准JSON响应结构
type JSONResponse struct {
	Code    constant.CodeResult `json:"code"`
	Message string              `json:"message"`
	Data    any                 `json:"data"`
}

// SetResult 设置响应结果
func (r *JSONResponse) SetResult(code constant.CodeResult, message string) {
	r.Code = code
	r.Message = message
}

// NewJSONResponse 创建标准JSON响应
func NewJSONResponse(code constant.CodeResult, message string) *JSONResponse {
	return &JSONResponse{
		Code:    code,
		Message: message,
	}
}

// NewJSONDataResponse 创建带数据的JSON响应
func NewJSONDataResponse(code constant.CodeResult, message string, data any) *JSONResponse {
	return &JSONResponse{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

// JSONResponsePage 分页响应结构
type JSONResponsePage struct {
	Code    constant.CodeResult `json:"code"`
	Message string              `json:"message"`
	Data    any                 `json:"data"`
	Count   int64               `json:"count"`
	Page    int                 `json:"page,omitempty"`
	Size    int                 `json:"size,omitempty"`
	Total   int64               `json:"total,omitempty"`
}

// NewJSONPageResponse 创建分页响应
func NewJSONPageResponse(code constant.CodeResult, message string, data any, count int64) *JSONResponsePage {
	return &JSONResponsePage{
		Code:    code,
		Message: message,
		Data:    data,
		Count:   count,
	}
}

// JSONResponseAPI API响应结构
type JSONResponseAPI struct {
	Code    constant.CodeResult `json:"code"`
	Message string              `json:"msg"`
	Data    any                 `json:"data"`
}

// Result 简单结果结构
type Result struct {
	Id   int64  `json:"id,string,omitempty"`
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// NewResult 创建简单结果
func NewResult(code int, msg string) *Result {
	return &Result{
		Code: code,
		Msg:  msg,
	}
}

// DataResult 数据结果结构
type DataResult struct {
	Code  int    `json:"code"`
	Msg   string `json:"msg"`
	Data  any    `json:"data"`
	Count int64  `json:"count"`
}

// NewDataResult 创建数据结果
func NewDataResult(code int, msg string, data any, count int64) *DataResult {
	return &DataResult{
		Code:  code,
		Msg:   msg,
		Data:  data,
		Count: count,
	}
}

// JSONUploadFile 文件上传响应结构
type JSONUploadFile struct {
	FileSize int64  `json:"fileSize"` // 文件大小
	FileExt  string `json:"fileExt"`  // 文件后缀
	FileUrl1 string `json:"fileUrl1"` // 文件相对路径
	FileUrl2 string `json:"fileUrl2"` // 文件绝对路径
	FileName string `json:"fileName"` // 保存文件名称
}

// NewJSONUploadFile 创建文件上传响应
func NewJSONUploadFile(size int64, ext, relPath, absPath, fileName string) *JSONUploadFile {
	return &JSONUploadFile{
		FileSize: size,
		FileExt:  ext,
		FileUrl1: relPath,
		FileUrl2: absPath,
		FileName: fileName,
	}
}

// TableResponse 表格响应结构(LayUI格式)
type TableResponse struct {
	Code  int    `json:"code"`
	Msg   string `json:"msg"`
	Data  any    `json:"data"`
	Count int64  `json:"count"`
}

// NewTableResponse 创建表格响应
func NewTableResponse(code int, msg string, data any, count int64) *TableResponse {
	return &TableResponse{
		Code:  code,
		Msg:   msg,
		Data:  data,
		Count: count,
	}
}

// TreeResponse 树形结构响应
type TreeResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

// NewTreeResponse 创建树形响应
func NewTreeResponse(code int, msg string, data any) *TreeResponse {
	return &TreeResponse{
		Code: code,
		Msg:  msg,
		Data: data,
	}
}

// SelectResponse 下拉选择响应结构
type SelectResponse struct {
	Code int          `json:"code"`
	Msg  string       `json:"msg"`
	Data []SelectItem `json:"data"`
}

// SelectItem 下拉选择项
type SelectItem struct {
	Value    string `json:"value"`
	Label    string `json:"label"`
	Selected bool   `json:"selected,omitempty"`
	Disabled bool   `json:"disabled,omitempty"`
}

// NewSelectResponse 创建下拉选择响应
func NewSelectResponse(code int, msg string, data []SelectItem) *SelectResponse {
	return &SelectResponse{
		Code: code,
		Msg:  msg,
		Data: data,
	}
}

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Code      constant.CodeResult `json:"code"`
	Message   string              `json:"message"`
	Error     string              `json:"error,omitempty"`
	Details   any                 `json:"details,omitempty"`
	Timestamp int64               `json:"timestamp"`
	Path      string              `json:"path,omitempty"`
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(code constant.CodeResult, message, error, path string, details any, timestamp int64) *ErrorResponse {
	return &ErrorResponse{
		Code:      code,
		Message:   message,
		Error:     error,
		Details:   details,
		Timestamp: timestamp,
		Path:      path,
	}
}

// ValidationError 验证错误结构
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   any    `json:"value,omitempty"`
}

// Error 实现error接口
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error in field '%s': %s", e.Field, e.Message)
}

// ValidationResponse 验证错误响应
type ValidationResponse struct {
	Code    constant.CodeResult `json:"code"`
	Message string              `json:"message"`
	Errors  []ValidationError   `json:"errors"`
}

// NewValidationResponse 创建验证错误响应
func NewValidationResponse(message string, errors []ValidationError) *ValidationResponse {
	return &ValidationResponse{
		Code:    constant.CodeParamError,
		Message: message,
		Errors:  errors,
	}
}

// =============== 便捷响应方法 ===============

// Success 成功响应
func Success(message string, data any) *JSONResponse {
	return NewJSONDataResponse(constant.CodeSuccess, message, data)
}

// Error 错误响应
func Error(message string) *JSONResponse {
	return NewJSONResponse(constant.CodeError, message)
}

// SuccessPage 成功分页响应
func SuccessPage(message string, data any, count int64) *JSONResponsePage {
	return NewJSONPageResponse(constant.CodeSuccess, message, data, count)
}

// ErrorPage 错误分页响应
func ErrorPage(message string, data any, count int64) *JSONResponsePage {
	return NewJSONPageResponse(constant.CodeError, message, data, count)
}

// SuccessTable 成功表格响应
func SuccessTable(data any, count int64) *TableResponse {
	return NewTableResponse(0, "", data, count)
}

// ErrorTable 错误表格响应
func ErrorTable(msg string) *TableResponse {
	return NewTableResponse(1, msg, nil, 0)
}

// BuildSuccessResp 构建成功响应
func BuildSuccessResp(data any) *JSONResponse {
	return &JSONResponse{
		Code:    constant.CodeSuccess,
		Message: "success",
		Data:    data,
	}
}

// BuildErrorResp 构建错误响应
func BuildErrorResp(err error) *JSONResponse {
	if err == nil {
		return BuildSuccessResp(nil)
	}

	if errNo, ok := err.(errors.ErrNo); ok {
		return &JSONResponse{
			Code:    constant.CodeResult(errNo.ErrCode),
			Message: errNo.ErrMsg,
		}
	}

	return &JSONResponse{
		Code:    constant.CodeError,
		Message: err.Error(),
	}
}

// BuildPageResp 构建分页响应
func BuildPageResp(data any, total int64, page, pageSize int) *JSONResponsePage {
	return &JSONResponsePage{
		Code:    constant.CodeSuccess,
		Message: "success",
		Data:    data,
		Count:   total,
		Page:    page,
	}
}

// BuildUploadResp 构建上传文件响应
func BuildUploadResp(url, filename string, size int64) *JSONUploadFile {
	return &JSONUploadFile{
		FileSize: size,
		FileUrl2: url,
		FileName: filename,
	}
}

// =============== 选项模式支持 ===============

// ResultOption Result选项函数
type ResultOption func(*Result)

// WithResultID 设置结果ID
func WithResultID(id int64) ResultOption {
	return func(r *Result) {
		r.Id = id
	}
}

// WithResultCode 设置结果代码
func WithResultCode(code int) ResultOption {
	return func(r *Result) {
		r.Code = code
	}
}

// WithResultMsg 设置结果消息
func WithResultMsg(msg string) ResultOption {
	return func(r *Result) {
		r.Msg = msg
	}
}

// NewResultWithOptions 使用选项模式创建结果
func NewResultWithOptions(opts ...ResultOption) *Result {
	result := &Result{}
	for _, opt := range opts {
		opt(result)
	}
	return result
}

// DataResultOption DataResult选项函数
type DataResultOption func(*DataResult)

// WithDataResultCode 设置数据结果代码
func WithDataResultCode(code int) DataResultOption {
	return func(dr *DataResult) {
		dr.Code = code
	}
}

// WithDataResultMsg 设置数据结果消息
func WithDataResultMsg(msg string) DataResultOption {
	return func(dr *DataResult) {
		dr.Msg = msg
	}
}

// WithDataResultData 设置数据结果数据
func WithDataResultData(data any) DataResultOption {
	return func(dr *DataResult) {
		dr.Data = data
	}
}

// WithDataResultCount 设置数据结果计数
func WithDataResultCount(count int64) DataResultOption {
	return func(dr *DataResult) {
		dr.Count = count
	}
}
