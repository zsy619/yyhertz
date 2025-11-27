package core

import (
	"testing"
)

func TestApp_generateRouteVariantsMap(t *testing.T) {
	// app := &App{}
	outList := GeneratePathVariants("admin/wechat/MpConfig/AccountEdit")
	for k, v := range outList {
		t.Log(k, " => ", v)
	}
}
