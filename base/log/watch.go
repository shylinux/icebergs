package log

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const WATCH = "watch"

func init() {
	Index.MergeCommands(ice.Commands{
		WATCH: {Help: "记录", Hand: func(m *ice.Message, arg ...string) {
			stats := map[string]int{}
			m.Cmd(nfs.CAT, path.Join(ice.VAR_LOG, "watch.log"), func(text string) {
				ls := kit.Split(text)
				m.Push(mdb.TIME, ls[0]+lex.SP+ls[1]).Push(mdb.ID, ls[2]).Push(nfs.SOURCE, kit.Slice(ls, -1)[0])
				m.Push(ctx.SHIP, ls[3]).Push("operate", ls[4]).Push(nfs.CONTENT, kit.Join(kit.Slice(ls, 5, -1), lex.SP))
				stats[ls[4]]++
			})
			m.Action("filter:text").StatusTimeCount(stats)
		}},
	})
}
