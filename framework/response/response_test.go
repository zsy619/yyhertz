package response

import (
	"errors"
	"testing"

	"github.com/zsy619/yyhertz/framework/constant"
)

func TestJSONResponse(t *testing.T) {
	t.Run("NewJSONResponse", func(t *testing.T) {
		r := NewJSONResponse(constant.CodeSuccess, "success")
		if r.Code != constant.CodeSuccess || r.Message != "success" {
			t.Errorf("NewJSONResponse failed, got: %v, want: %v", r, constant.CodeSuccess)
		}
	})

	t.Run("NewJSONDataResponse", func(t *testing.T) {
		r := NewJSONDataResponse(constant.CodeSuccess, "success", "data")
		if r.Code != constant.CodeSuccess || r.Message != "success" || r.Data != "data" {
			t.Errorf("NewJSONDataResponse failed, got: %v, want data: %v", r, "data")
		}
	})

	t.Run("SetResult", func(t *testing.T) {
		r := &JSONResponse{}
		r.SetResult(constant.CodeError, "error")
		if r.Code != constant.CodeError || r.Message != "error" {
			t.Errorf("SetResult failed, got: %v, want: %v", r, constant.CodeError)
		}
	})
}

func TestJSONResponsePage(t *testing.T) {
	t.Run("NewJSONPageResponse", func(t *testing.T) {
		r := NewJSONPageResponse(constant.CodeSuccess, "success", "data", 10)
		if r.Code != constant.CodeSuccess || r.Message != "success" || r.Data != "data" || r.Count != 10 {
			t.Errorf("NewJSONPageResponse failed, got: %v, want count: %v", r, 10)
		}
	})
}

func TestErrorResponse(t *testing.T) {
	t.Run("NewErrorResponse", func(t *testing.T) {
		r := NewErrorResponse(constant.CodeParamError, "invalid input", "validation error", "/path", nil, 1234567890)
		if r.Code != constant.CodeParamError || r.Message != "invalid input" || r.Error != "validation error" || r.Path != "/path" {
			t.Errorf("NewErrorResponse failed, got: %v, want error: %v", r, "validation error")
		}
	})
}

func TestValidationResponse(t *testing.T) {
	t.Run("NewValidationResponse", func(t *testing.T) {
		errors := []ValidationError{{
			Field:   "username",
			Message: "required",
		}}
		r := NewValidationResponse("validation failed", errors)
		if r.Code != constant.CodeParamError || r.Message != "validation failed" || len(r.Errors) != 1 {
			t.Errorf("NewValidationResponse failed, got: %v, want errors: %v", r, errors)
		}
	})
}

func TestTableResponse(t *testing.T) {
	t.Run("NewTableResponse", func(t *testing.T) {
		r := NewTableResponse(0, "", "data", 10)
		if r.Code != 0 || r.Msg != "" || r.Data != "data" || r.Count != 10 {
			t.Errorf("NewTableResponse failed, got: %v, want count: %v", r, 10)
		}
	})
}

func TestTreeResponse(t *testing.T) {
	t.Run("NewTreeResponse", func(t *testing.T) {
		r := NewTreeResponse(0, "", "data")
		if r.Code != 0 || r.Msg != "" || r.Data != "data" {
			t.Errorf("NewTreeResponse failed, got: %v, want data: %v", r, "data")
		}
	})
}

func TestSelectResponse(t *testing.T) {
	t.Run("NewSelectResponse", func(t *testing.T) {
		items := []SelectItem{{
			Value: "1",
			Label: "Option 1",
		}}
		r := NewSelectResponse(0, "", items)
		if r.Code != 0 || r.Msg != "" || len(r.Data) != 1 {
			t.Errorf("NewSelectResponse failed, got: %v, want items: %v", r, items)
		}
	})
}

func TestDataResult(t *testing.T) {
	t.Run("NewDataResult", func(t *testing.T) {
		r := NewDataResult(0, "", "data", 10)
		if r.Code != 0 || r.Msg != "" || r.Data != "data" || r.Count != 10 {
			t.Errorf("NewDataResult failed, got: %v, want count: %v", r, 10)
		}
	})
}

func TestJSONUploadFile(t *testing.T) {
	t.Run("NewJSONUploadFile", func(t *testing.T) {
		r := NewJSONUploadFile(1024, ".txt", "/path", "/abs/path", "file.txt")
		if r.FileSize != 1024 || r.FileExt != ".txt" || r.FileUrl1 != "/path" || r.FileUrl2 != "/abs/path" || r.FileName != "file.txt" {
			t.Errorf("NewJSONUploadFile failed, got: %v, want filename: %v", r, "file.txt")
		}
	})
}

func TestResultOption(t *testing.T) {
	t.Run("NewResultWithOptions", func(t *testing.T) {
		r := NewResultWithOptions(
			WithResultID(1),
			WithResultCode(0),
			WithResultMsg("success"),
		)
		if r.Id != 1 || r.Code != 0 || r.Msg != "success" {
			t.Errorf("NewResultWithOptions failed, got: %v, want id: %v", r, 1)
		}
	})
}

func TestBuildErrorResp(t *testing.T) {
	t.Run("BuildErrorResp with error", func(t *testing.T) {
		err := errors.New("test error")
		r := BuildErrorResp(err)
		if r.Code != constant.CodeError || r.Message != "test error" {
			t.Errorf("BuildErrorResp failed, got: %v, want error: %v", r, "test error")
		}
	})

	t.Run("BuildErrorResp with nil", func(t *testing.T) {
		r := BuildErrorResp(nil)
		if r.Code != constant.CodeSuccess || r.Message != "success" {
			t.Errorf("BuildErrorResp failed, got: %v, want success", r)
		}
	})
}

func TestBuildPageResp(t *testing.T) {
	t.Run("BuildPageResp", func(t *testing.T) {
		r := BuildPageResp("data", 10, 1, 10)
		if r.Code != constant.CodeSuccess || r.Message != "success" || r.Data != "data" || r.Count != 10 || r.Page != 1 {
			t.Errorf("BuildPageResp failed, got: %v, want page: %v", r, 1)
		}
	})
}

func TestBuildUploadResp(t *testing.T) {
	t.Run("BuildUploadResp", func(t *testing.T) {
		r := BuildUploadResp("/path", "file.txt", 1024)
		if r.FileSize != 1024 || r.FileUrl2 != "/path" || r.FileName != "file.txt" {
			t.Errorf("BuildUploadResp failed, got: %v, want filename: %v", r, "file.txt")
		}
	})
}
