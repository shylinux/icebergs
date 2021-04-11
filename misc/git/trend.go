package git

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/ctx"
	kit "github.com/shylinux/toolkits"
)

const TREND = "trend"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		TREND: {Name: "trend name@key begin_time@date auto", Help: "趋势图", Meta: kit.Dict(
			kit.MDB_DISPLAY, "/plugin/story/trend.js",
		), Action: map[string]*ice.Action{
			ctx.COMMAND: {Name: "ctx.command"},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 { // 仓库列表
				m.Cmdy(REPOS)
				return
			}

			m.Cmdy(TOTAL, arg)
		}},
	}})
}
