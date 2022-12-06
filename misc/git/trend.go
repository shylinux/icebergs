package git

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const TREND = "trend"

func init() {
	Index.MergeCommands(ice.Commands{
		TREND: {Name: "trend repos@key begin_time@date auto", Help: "趋势图", Actions: ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(REPOS, ice.OptionFields("repos,time")) }},
		}, Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				m.Cmdy(REPOS)
			} else {
				m.Cmdy(TOTAL, kit.Slice(arg, 0, 2))
				ctx.DisplayStory(m, "")
			}
		}},
	})
}
