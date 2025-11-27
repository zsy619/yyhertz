package response

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zsy619/yyhertz/framework/constant"
	"github.com/zsy619/yyhertz/framework/errors"
)

func TestBuildSuccessResp(t *testing.T) {
	t.Run("构建成功响应", func(t *testing.T) {
		data := map[string]any{
			"id":   1,
			"name": "test",
		}

		resp := BuildSuccessResp(data)

		assert.Equal(t, constant.CodeSuccess, resp.Code)
		assert.Equal(t, "success", resp.Message)
		assert.Equal(t, data, resp.Data)
	})

	t.Run("构建成功响应-空数据", func(t *testing.T) {
		resp := BuildSuccessResp(nil)

		assert.Equal(t, constant.CodeSuccess, resp.Code)
		assert.Equal(t, "success", resp.Message)
		assert.Nil(t, resp.Data)
	})
}

func TestBuildErrorResp(t *testing.T) {
	t.Run("构建错误响应-nil错误", func(t *testing.T) {
		resp := BuildErrorResp(nil)

		assert.Equal(t, constant.CodeSuccess, resp.Code)
		assert.Equal(t, "success", resp.Message)
		assert.Nil(t, resp.Data)
	})

	t.Run("构建错误响应-ErrNo类型", func(t *testing.T) {
		err := errors.UserNotExist
		resp := BuildErrorResp(err)

		assert.Equal(t, constant.CodeResult(err.ErrCode), resp.Code)
		assert.Equal(t, err.ErrMsg, resp.Message)
	})

	t.Run("构建错误响应-普通错误", func(t *testing.T) {
		err := errors.NewSystemError("test error")
		resp := BuildErrorResp(err)

		assert.Equal(t, constant.CodeError, resp.Code)
		assert.Equal(t, "test error", resp.Message)
	})
}

func TestBuildPageResp(t *testing.T) {
	t.Run("构建分页响应", func(t *testing.T) {
		data := []map[string]any{
			{"id": 1, "name": "item1"},
			{"id": 2, "name": "item2"},
		}

		resp := BuildPageResp(data, 100, 1, 20)

		assert.Equal(t, constant.CodeSuccess, resp.Code)
		assert.Equal(t, "success", resp.Message)
		assert.Equal(t, data, resp.Data)
		assert.Equal(t, int64(100), resp.Count)
		assert.Equal(t, 1, resp.Page)
		assert.Equal(t, 20, resp.Size)
	})
}

func TestBuildUploadResp(t *testing.T) {
	t.Run("构建上传响应", func(t *testing.T) {
		url := "http://example.com/file.jpg"
		filename := "test.jpg"
		size := int64(1024)

		resp := BuildUploadResp(url, filename, size)

		assert.Equal(t, url, resp.FileUrl2)
		assert.Equal(t, filename, resp.FileName)
		assert.Equal(t, size, resp.FileSize)
		// JSONUploadFile没有Code和Message字段
	})
}

func TestParseBaseResp(t *testing.T) {
	t.Run("解析成功响应", func(t *testing.T) {
		resp := &JSONResponse{
			Code:    constant.CodeSuccess,
			Message: "success",
		}

		err := ParseBaseResp(resp)
		assert.Nil(t, err)
	})

	t.Run("解析错误响应", func(t *testing.T) {
		resp := &JSONResponse{
			Code:    constant.CodeError,
			Message: "error message",
		}

		err := ParseBaseResp(resp)
		assert.NotNil(t, err)

		errNo, ok := err.(errors.ErrNo)
		assert.True(t, ok)
		assert.Equal(t, int64(constant.CodeError), errNo.ErrCode)
		assert.Equal(t, "error message", errNo.ErrMsg)
	})
}
