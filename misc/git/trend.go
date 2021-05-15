package git

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/ctx"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/core/code"
	kit "github.com/shylinux/toolkits"
)

const TREND = "trend"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		TREND: {Name: "trend name@key begin_time@date auto", Help: "趋势图", Meta: kit.Dict(
			kit.MDB_DISPLAY, "/plugin/story/trend.js",
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
