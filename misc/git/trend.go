package git

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/ctx"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

const TREND = "trend"

func init() {
	Index.Merge(&ice.Context{
		Commands: map[string]*ice.Command{
			TREND: {Name: "trend name=icebergs@key begin_time@date auto", Help: "趋势图", Meta: kit.Dict(
				"display", "/plugin/story/trend.js",
			), Action: map[string]*ice.Action{
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(REPOS).Appendv(ice.MSG_APPEND, kit.Split("name,branch,commit"))
				}},
				ctx.COMMAND: {Name: "command", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
					for _, k := range arg {
						m.Cmdy(ctx.COMMAND, k)
					}
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					m.Option(ice.MSG_DISPLAY, "table")
				}
				m.Cmdy(TOTAL, arg)
			}},
		},
	})
}
