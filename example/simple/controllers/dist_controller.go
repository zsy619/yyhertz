package controllers

import (
	"github.com/zsy619/yyhertz/framework/mvc"
)

type DistController struct {
	mvc.BaseController
}

func (ctrl *DistController) GetSearchBill() {
	respWriter := ctrl.Ctx.HTTPResponseWriter()
	respWriter.Header().Set("Content-Type", "application/json; charset=utf-8")
	respWriter.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	respWriter.Header().Set("Pragma", "no-cache")
	respWriter.Header().Set("Expires", "0")
	out := struct {
		Message string `json:"message"`
		UserID  string `json:"user_id,omitempty"`
	}{
		Message: "SearchBill",
		UserID:  "12345",
	}
	ctrl.Data["json"] = out
	ctrl.ServeJSON()
	ctrl.StopRun()
}
