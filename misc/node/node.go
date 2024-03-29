package node

import (
	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
)

const NODE = "node"

type node struct {
	ice.Code
	regexp string `data:"js"`
	darwin string `data:"https://mirrors.tencent.com/nodejs-release/v16.15.1/node-v16.15.1-darwin-x64.tar.gz"`
	linux  string `data:"https://mirrors.tencent.com/nodejs-release/v16.15.1/node-v16.15.1-linux-x64.tar.xz"`
	list   string `name:"list path auto xterm order install" help:"运行时"`
}

func (s node) Init(m *ice.Message) {
	cli.IsAlpine(m.Message, NODE, "nodejs")
	cli.IsRedhat(m.Message, NODE, "nodejs")
}
func (s node) List(m *ice.Message, arg ...string) {
	s.Code.Source(m, "", arg...)
}
func (s node) Xterm(m *ice.Message, arg ...string) {
	s.Code.Xterm(m, "", []string{mdb.TYPE, NODE}, arg...)
}
func init() { ice.CodeCtxCmd(node{}) }
