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
		TREND: {Name: "trend name begin_time@date auto", Help: "趋势图", Meta: kit.Dict(
			ice.Display("/plugin/story/trend.js"),
		), Action: ice.MergeAction(map[string]*ice.Action{
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(REPOS, ice.OptionFields("name,time"))
			}}, code.INNER: {Name: "web.code.inner"},
		}, ctx.CmdAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 { // 仓库列表
				m.Cmdy(REPOS)
				return
			}
			arg[0] = kit.Replace(arg[0], "src", "contexts")
			m.Cmdy(TOTAL, arg)
		}},
	}})
}
