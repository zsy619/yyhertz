package types

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrNo(t *testing.T) {
	t.Run("创建错误", func(t *testing.T) {
		err := NewErrNo(1001, "test error")
		
		assert.Equal(t, int64(1001), err.ErrCode)
		assert.Equal(t, "test error", err.ErrMsg)
		assert.Equal(t, "err_code=1001, err_msg=test error", err.Error())
	})
	
	t.Run("修改错误消息", func(t *testing.T) {
		err := UserNotExist.WithMessage("用户ID不存在")
		
		assert.Equal(t, UserNotExist.ErrCode, err.ErrCode)
		assert.Equal(t, "用户ID不存在", err.ErrMsg)
		assert.NotEqual(t, UserNotExist.ErrMsg, err.ErrMsg)
	})
}

func TestPredefinedErrors(t *testing.T) {
	t.Run("预定义错误测试", func(t *testing.T) {
		testCases := []struct {
			name     string
			err      ErrNo
			code     int64
			minRange int64
			maxRange int64
		}{
			{"SuccessErrNo", SuccessErrNo, 0, 0, 999},
			{"ServiceError", ServiceError, 10001, 10000, 10999},
			{"ParamError", ParamError, 10002, 10000, 10999},
			{"UserNotExist", UserNotExist, 20001, 20000, 20999},
			{"UserAlreadyExist", UserAlreadyExist, 20002, 20000, 20999},
			{"PermissionDenied", PermissionDenied, 30001, 30000, 30999},
			{"TokenExpired", TokenExpired, 30002, 30000, 30999},
			{"DataNotFound", DataNotFound, 40001, 40000, 40999},
			{"DatabaseError", DatabaseError, 40006, 40000, 40999},
			{"FileNotFound", FileNotFound, 50001, 50000, 50999},
			{"NetworkError", NetworkError, 60001, 60000, 60999},
			{"RPCError", RPCError, 70001, 70000, 70999},
			{"CacheError", CacheError, 80001, 80000, 80999},
			{"ConfigError", ConfigError, 90001, 90000, 90999},
		}
		
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				assert.Equal(t, tc.code, tc.err.ErrCode)
				assert.NotEmpty(t, tc.err.ErrMsg)
				assert.GreaterOrEqual(t, tc.err.ErrCode, tc.minRange)
				assert.LessOrEqual(t, tc.err.ErrCode, tc.maxRange)
			})
		}
	})
}

func TestBuildBaseResp(t *testing.T) {
	t.Run("构建成功响应", func(t *testing.T) {
		resp := BuildBaseResp(nil)
		
		assert.Equal(t, CodeSuccess, resp.Code)
		assert.Equal(t, "success", resp.Message)
	})
	
	t.Run("构建ErrNo错误响应", func(t *testing.T) {
		err := UserNotExist
		resp := BuildBaseResp(err)
		
		assert.Equal(t, CodeResult(err.ErrCode), resp.Code)
		assert.Equal(t, err.ErrMsg, resp.Message)
	})
	
	t.Run("构建普通错误响应", func(t *testing.T) {
		err := errors.New("standard error")
		resp := BuildBaseResp(err)
		
		assert.Equal(t, CodeResult(ParamError.ErrCode), resp.Code)
		assert.Equal(t, ParamError.ErrMsg, resp.Message)
	})
}

func TestParseBaseResp(t *testing.T) {
	t.Run("解析成功响应", func(t *testing.T) {
		resp := &JSONResponse{
			Code:    CodeSuccess,
			Message: "success",
		}
		
		err := ParseBaseResp(resp)
		assert.Nil(t, err)
	})
	
	t.Run("解析错误响应", func(t *testing.T) {
		resp := &JSONResponse{
			Code:    CodeError,
			Message: "error occurred",
		}
		
		err := ParseBaseResp(resp)
		assert.NotNil(t, err)
		
		errNo, ok := err.(ErrNo)
		assert.True(t, ok)
		assert.Equal(t, int64(CodeError), errNo.ErrCode)
		assert.Equal(t, "error occurred", errNo.ErrMsg)
	})
}

func TestErrorCategorization(t *testing.T) {
	t.Run("错误分类测试", func(t *testing.T) {
		// 系统错误 (10000-10999)
		systemErrors := []ErrNo{ServiceError, ParamError, AuthorizeFail, TooManyRequests, ForbiddenError}
		for _, err := range systemErrors {
			assert.GreaterOrEqual(t, err.ErrCode, int64(10000))
			assert.Less(t, err.ErrCode, int64(11000))
		}
		
		// 用户错误 (20000-20999)
		userErrors := []ErrNo{UserNotExist, UserAlreadyExist, UserCreateError, UserLoginError}
		for _, err := range userErrors {
			assert.GreaterOrEqual(t, err.ErrCode, int64(20000))
			assert.Less(t, err.ErrCode, int64(21000))
		}
		
		// 权限错误 (30000-30999)
		authErrors := []ErrNo{PermissionDenied, TokenExpired, TokenInvalid, TokenMissing}
		for _, err := range authErrors {
			assert.GreaterOrEqual(t, err.ErrCode, int64(30000))
			assert.Less(t, err.ErrCode, int64(31000))
		}
		
		// 数据错误 (40000-40999)
		dataErrors := []ErrNo{DataNotFound, DataCreateError, DataUpdateError, DatabaseError}
		for _, err := range dataErrors {
			assert.GreaterOrEqual(t, err.ErrCode, int64(40000))
			assert.Less(t, err.ErrCode, int64(41000))
		}
	})
}

func TestErrorInterface(t *testing.T) {
	t.Run("Error接口实现", func(t *testing.T) {
		err := UserNotExist
		
		// 测试可以作为error使用
		var e error = err
		assert.NotNil(t, e)
		assert.Contains(t, e.Error(), "err_code=20001")
		assert.Contains(t, e.Error(), "User does not exists")
	})
}