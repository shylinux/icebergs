package chrome

import (
	"shylinux.com/x/ice"
)

type chrome struct {
	ice.Code

	source string `data:"https://mirrors.tencent.com/tinycorelinux/4.x/x86/tcz/src/chromium-browser/chromium-22.0.1229.79.tar.xz"`
	list   string `name:"list path auto order build download" help:"源码"`
}

func (c chrome) List(m *ice.Message, arg ...string) {
	c.Code.Source(m, "", arg...)
}

func init() { ice.CodeCtxCmd(chrome{}) }
