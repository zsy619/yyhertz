package response

import (
	"github.com/zsy619/yyhertz/framework/constant"
	"github.com/zsy619/yyhertz/framework/errors"
)

// BuildBaseResp 构建基础响应(来自FreeCar项目)
func BuildBaseResp(err error) *JSONResponse {
	if err == nil {
		return &JSONResponse{
			Code:    constant.CodeSuccess,
			Message: "success",
		}
	}

	errNo := errors.ParamError
	if e, ok := err.(errors.ErrNo); ok {
		errNo = e
	}

	return &JSONResponse{
		Code:    constant.CodeResult(errNo.ErrCode),
		Message: errNo.ErrMsg,
	}
}

// ParseBaseResp 解析基础响应
func ParseBaseResp(resp *JSONResponse) error {
	if resp.Code == constant.CodeSuccess {
		return nil
	}

	return errors.NewErrNo(int64(resp.Code), resp.Message)
}
