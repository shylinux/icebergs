package git

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

const TREND = "trend"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		TREND: {Name: "trend name@key begin_time@date auto", Help: "趋势图", Meta: kit.Dict(
			ice.Display("/plugin/story/trend.js"),
		), Action: map[string]*ice.Action{
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(REPOS, ice.OptionFields("name,time"))
			}}, ctx.COMMAND: {Name: "ctx.command"}, code.INNER: {Name: "web.code.inner"},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 { // 仓库列表
				m.Cmdy(REPOS)
				return
			}

			m.Cmdy(TOTAL, arg)
		}},
	}})
}
