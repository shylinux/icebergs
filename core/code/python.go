package code

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const PYTHON = "python"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		PYTHON: {Name: PYTHON, Help: "脚本命令", Value: kit.Data(
			nfs.SOURCE, "http://mirrors.sohu.com/python/3.5.2/Python-3.5.2.tar.xz",
			PYTHON, "python", "pip", "pip",
		)},
	}, Commands: map[string]*ice.Command{
		PYTHON: {Name: "python path auto order build download", Help: "脚本命令", Action: ice.MergeAction(map[string]*ice.Action{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(mdb.ENGINE, mdb.CREATE, mdb.TYPE, "py", mdb.NAME, m.PrefixKey())
			}},
			ice.RUN: {Name: "run", Help: "运行", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(cli.SYSTEM, m.Config(PYTHON), arg)
			}},
			"pip": {Name: "pip", Help: "安装", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(cli.SYSTEM, m.Config("pip"), "install", arg)
			}},
		}, InstallAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmdy(INSTALL, nfs.SOURCE, m.Config(nfs.SOURCE), arg)
		}},
	}})
}
