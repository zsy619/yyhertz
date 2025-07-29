package util

import (
	"hertz-controller/framework/types"
)

// BuildSuccessResp 构建成功响应
func BuildSuccessResp(data any) *types.JSONResponse {
	return &types.JSONResponse{
		Code:    types.CodeSuccess,
		Message: "success",
		Data:    data,
	}
}

// BuildErrorResp 构建错误响应
func BuildErrorResp(err error) *types.JSONResponse {
	if err == nil {
		return BuildSuccessResp(nil)
	}

	if errNo, ok := err.(types.ErrNo); ok {
		return &types.JSONResponse{
			Code:    types.CodeResult(errNo.ErrCode),
			Message: errNo.ErrMsg,
		}
	}

	return &types.JSONResponse{
		Code:    types.CodeError,
		Message: err.Error(),
	}
}

// BuildBaseResp 构建基础响应(兼容FreeCar风格)
func BuildBaseResp(err error) *types.JSONResponse {
	return types.BuildBaseResp(err)
}

// ParseBaseResp 解析基础响应
func ParseBaseResp(resp *types.JSONResponse) error {
	return types.ParseBaseResp(resp)
}

// BuildPageResp 构建分页响应
func BuildPageResp(data any, total int64, page, pageSize int) *types.JSONResponsePage {
	return &types.JSONResponsePage{
		Code:    types.CodeSuccess,
		Message: "success",
		Data:    data,
		Count:   total,
		Page:    page,
	}
}

// BuildUploadResp 构建上传文件响应
func BuildUploadResp(url, filename string, size int64) *types.JSONUploadFile {
	return &types.JSONUploadFile{
		FileSize: size,
		FileUrl2: url,
		FileName: filename,
	}
}
