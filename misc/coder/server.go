package coder

import (
	"path"

	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
)

type server struct {
	ice.Code
	linux    string `data:"https://github.com/coder/code-server/releases/download/v4.4.0/code-server-4.4.0-linux-amd64.tar.gz"`
	darwin   string `data:"https://github.com/coder/code-server/releases/download/v4.4.0/code-server-4.4.0-macos-amd64.tar.gz"`
	source   string `data:"https://github.com/coder/code-server/archive/refs/tags/v4.4.0.tar.gz"`
	password string `data:"1234"`
	start    string `name:"start port host password" help:"启动"`
	list     string `name:"list port path auto start install" help:"编辑器"`
}

func (s server) Search(m *ice.Message, arg ...string) {
	if mdb.IsSearchForEach(m.Message, arg, nil) {
		s.Code.List(m.Spawn(kit.Dict(ice.MSG_FIELDS, "time,port,status,pid,cmd,dir")), "")
		m.Table(func(value ice.Maps) {
			m.PushSearch(mdb.TYPE, value[cli.STATUS], mdb.NAME, value[nfs.PATH], mdb.TEXT, value[mdb.LINK])
		})
	}
}
func (s server) Start(m *ice.Message, arg ...string) {
	s.Code.Start(m, "", "bin/code-server", func(p string) []string {
		m.Cmd(nfs.SAVE, kit.Path(p, "config"), kit.Format(`
user-data-dir: %s
bind-addr: %s:%s
password: %s
`, "./data", kit.Select("0.0.0.0", m.Option(tcp.HOST)), path.Base(p), kit.Select(mdb.Config(m, aaa.PASSWORD), m.Option(aaa.PASSWORD))))
		return []string{"--config=config", kit.Path(nfs.PWD)}
	})
}
func (s server) List(m *ice.Message, arg ...string) {
	if s.Code.List(m, "", arg...); len(arg) == 0 {
		m.Table(func(value ice.Maps) {
			switch value[cli.STATUS] {
			case cli.START:
				m.PushButton(s.Open, s.Stop)
			default:
				m.PushButton("")
			}
		})
	}
}
func init() { ice.CodeCtxCmd(server{}) }
