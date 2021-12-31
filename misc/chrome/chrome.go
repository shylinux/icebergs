package chrome

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

const CHROME = "chrome"

var Index = &ice.Context{Name: CHROME, Help: "浏览器", Configs: map[string]*ice.Config{
	CHROME: {Name: CHROME, Help: "浏览器", Value: kit.Data()},
}, Commands: map[string]*ice.Command{
	CHROME: {Name: "chrome wid tid url auto start build download", Help: "浏览器", Action: ice.MergeAction(map[string]*ice.Action{
		mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
			m.Cmd(CHROME).Table(func(index int, value map[string]string, head []string) {
				m.Cmd(CHROME, value["wid"]).Table(func(index int, value map[string]string, head []string) {
					m.Push(kit.MDB_ZONE, kit.ParseURL(value["url"]).Host)
				})
			})
		}},
		cli.ORDER: {Name: "order", Help: "加载", Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(code.INSTALL, cli.ORDER, m.Config(nfs.SOURCE), "_install/bin")
		}},
	}, code.InstallAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		m.Cmdy(web.SPACE, CHROME, CHROME, arg)
	}},
}}

func init() { code.Index.Register(Index, &web.Frame{}) }
