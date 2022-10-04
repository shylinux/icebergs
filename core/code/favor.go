package code

import (
	"strings"
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const FAVOR = "favor"

func init() {
	Index.MergeCommands(ice.Commands{
		FAVOR: {Name: "favor zone id auto insert page", Help: "收藏夹", Actions: ice.MergeActions(ice.Actions{
			mdb.INSERT: {Name: "insert zone=数据结构 type=go name=hi text=hello path file line", Help: "添加"},
			XTERM: {Name: "xterm", Help: "终端", Hand: func(m *ice.Message, arg ...string) {
				ctx.Process(m, m.ActionKey(), m.OptionSimple(mdb.TYPE, mdb.NAME, mdb.TEXT), arg...)
			}},
			INNER: {Name: "inner", Help: "源码", Hand: func(m *ice.Message, arg ...string) {
				ctx.Process(m, m.ActionKey(), m.OptionSplit(nfs.PATH, nfs.FILE, nfs.LINE), arg...)
			}},
			"click": {Name: "click", Help: "源码", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(cli.DAEMON, m.Option(mdb.TYPE))
			}},
		}, mdb.ZoneAction(mdb.SHORT, mdb.ZONE, mdb.FIELD, "time,id,type,name,text,path,file,line")), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 || arg[0] == "" {
				m.Push(mdb.TIME, m.Time())
				m.Push(mdb.ZONE, "_history")
				m.Push(mdb.COUNT, "100")
				m.PushButton("")
			} else if arg[0] == "_history" {
				last := ""
				list := map[string]string{}
				m.Cmd(nfs.CAT, kit.HomePath(".bash_history"), func(line string) {
					if strings.HasPrefix(line, "#") {
						last = time.Unix(kit.Int64(line[1:]), 0).Format(ice.MOD_TIME)
					} else if last != "" {
						list[line] = last
					}
				})
				for k, v := range list {
					m.Push(mdb.TIME, v)
					m.Push(mdb.TYPE, k)
				}
				m.SortTimeR(mdb.TIME)
				m.StatusTimeCount()
				return
			}
			m.Option(mdb.CACHE_LIMIT, "30")
			m.Option(mdb.LIMIT, "30")
			if mdb.ZoneSelectPage(m, arg...); len(arg) > 0 && arg[0] != "" {
				m.Option(ctx.STYLE, arg[0])
				m.Tables(func(value ice.Maps) {
					m.PushButton(kit.Select(INNER, XTERM, value[mdb.TEXT] == "" || value[nfs.FILE] == ""))
				})
			}
		}},
	})
}
