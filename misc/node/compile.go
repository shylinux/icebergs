package node

import (
	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/cli"
)

const (
	NPM  = "npm"
	NODE = "node"
)

type compile struct {
	ice.Code
	regexp string `data:".*.js"`
	linux  string `data:"https://mirrors.tencent.com/nodejs-release/v16.15.1/node-v16.15.1-linux-x64.tar.xz"`
	list   string `name:"list path auto xterm listScript order install" help:"编译器"`
}

func (s compile) Init(m *ice.Message) {
	m.Go(func() {
		m.Sleep300ms() // after runtime init
		cli.IsAlpine(m.Message, NPM)
		cli.IsAlpine(m.Message, NODE, "nodejs")
	})
}
func (s compile) List(m *ice.Message, arg ...string) {
	s.Code.Source(m, "", arg...)
}
func (s compile) Xterm(m *ice.Message, arg ...string) {
	s.Code.Xterm(m, NODE, arg...)
}

func init() { ice.CodeCtxCmd(compile{}) }
