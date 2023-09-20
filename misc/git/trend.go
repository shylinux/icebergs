package git

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

const TREND = "trend"

func init() {
	Index.MergeCommands(ice.Commands{
		TREND: {Name: "trend repos@key begin_time@date auto", Help: "趋势图", Actions: ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(REPOS, ice.OptionFields("repos,time")) }},
			mdb.DETAIL: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy("", code.INNER, m.Option(REPOS), MASTER, m.Option(mdb.HASH), m.Cmdv(REPOS, m.Option(REPOS), MASTER, m.Option(mdb.HASH), nfs.FILE))
			}},
			code.INNER: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(REPOS, code.INNER, arg)
				ctx.DisplayLocal(m, "code/inner.js", "style", "float")
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				m.Cmdy(REPOS)
			} else {
				ctx.DisplayStory(m.Cmdy(TOTAL, kit.Slice(arg, 0, 2)), "")
			}
		}},
	})
}
