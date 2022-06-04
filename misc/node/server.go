package node

import (
	"shylinux.com/x/ice"
)

type server struct {
	ice.Code
	command string `data:"node"`
	regexp  string `data:".*.js"`
	linux   string `data:"https://mirrors.tencent.com/nodejs-release/v16.15.1/node-v16.15.1-linux-x64.tar.xz"`
	list    string `name:"list path auto listScript order install" help:"编译执行"`
}

func (s server) Order(m *ice.Message) {
	s.Code.Order(m, "", ice.BIN)
}
func (s server) List(m *ice.Message, arg ...string) {
	s.Code.Source(m, "", arg...)
}

func init() { ice.CodeCtxCmd(server{}) }
