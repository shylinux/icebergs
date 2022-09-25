package node

import (
	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
)

const (
	NPM  = "npm"
	NODE = "node"
)

type compile struct {
	ice.Code
	regexp string `data:".*.js"`
	linux  string `data:"https://mirrors.tencent.com/nodejs-release/v16.15.1/node-v16.15.1-linux-x64.tar.xz"`
	darwin string `data:"https://mirrors.tencent.com/nodejs-release/v16.15.1/node-v16.15.1-darwin-x64.tar.gz"`
	list   string `name:"list path auto xterm listScript order install" help:"编译器"`
}

func (s compile) Init(m *ice.Message) {
	cli.IsAlpine(m.Message, NPM)
	cli.IsAlpine(m.Message, NODE, "nodejs")
}
func (s compile) List(m *ice.Message, arg ...string) {
	s.Code.Source(m, "", arg...)
}
func (s compile) Xterm(m *ice.Message, arg ...string) {
	s.Code.Xterm(m, []string{mdb.TYPE, NODE}, arg...)
}

func init() { ice.CodeCtxCmd(compile{}) }
