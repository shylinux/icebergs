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
	Index.MergeCommands(ice.Commands{
		TREND: {Name: "trend repos@key begin_time@date auto", Help: "趋势图", Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(REPOS, ice.OptionFields("name,time"))
			}}, code.INNER: {Name: "web.code.inner"},
		}, ctx.CmdAction()), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 { // 仓库列表
				m.Cmdy(REPOS)
				return
			}
			arg[0] = kit.Replace(arg[0], ice.SRC, ice.CONTEXTS)
			m.Cmdy(TOTAL, kit.Slice(arg, 0, 2))
			ctx.DisplayStory(m, "trend.js")
		}},
	})
}
