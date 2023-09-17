package node

import (
	"path"

	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const NODE = "node"

type node struct {
	ice.Code
	regexp string `data:"js"`
	darwin string `data:"https://mirrors.tencent.com/nodejs-release/v16.15.1/node-v16.15.1-darwin-x64.tar.gz"`
	linux  string `data:"https://mirrors.tencent.com/nodejs-release/v16.15.1/node-v16.15.1-linux-x64.tar.xz"`
	list   string `name:"list path auto listScript xterm order install" help:"运行时"`
}

func (s node) Init(m *ice.Message) {
	cli.IsAlpine(m.Message, NODE, "nodejs")
	cli.IsCentos(m.Message, NODE, "nodejs")
}
func (s node) Install(m *ice.Message, arg ...string) {
	s.Code.Install(m, arg...)
	s.Code.System(m, ice.USR_INSTALL, nfs.TAR, "xf", path.Base(s.Code.Link(m)))
}
func (s node) List(m *ice.Message, arg ...string) {
	s.Code.Source(m, path.Join(ice.USR_INSTALL, kit.TrimExt(s.Code.Link(m), "tar.xz")), arg...)
}
func (s node) Xterm(m *ice.Message, arg ...string) {
	s.Code.Xterm(m, []string{mdb.TYPE, NODE}, arg...)
}
func init() { ice.CodeCtxCmd(node{}) }
