package web

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
)

const (
	STATS_TABLES = "stats.tables"
)
const STATS = "stats"

func init() {
	Index.MergeCommands(ice.Commands{
		STATS: {Name: "stats name auto", Help: "汇总量", Meta: kit.Dict(
			ice.CTX_TRANS, kit.Dict(html.INPUT, kit.Dict(
				"repos.total", "代码库总数",
				"dream.total", "空间总数",
				"dream.start", "已启动空间",
				"asset.amount", "资产总额",
				"asset.count", "资产数量",
				"goods.amount", "商品总额",
				"goods.count", "商品数量",
				"user.total", "用户总数",
				"sess.total", "会话总数",
				"task.total", "任务总数",
				"disk.total", "磁盘总量",
				"disk.used", "磁盘用量",
				"mem.total", "内存总量",
				"mem.used", "内存用量",
			)),
		), Hand: func(m *ice.Message, arg ...string) {
			m.Push(mdb.NAME, kit.Keys(aaa.SESS, mdb.TOTAL)).Push(mdb.VALUE, m.Cmd(aaa.SESS).Length())
			m.Push(mdb.NAME, kit.Keys(aaa.USER, mdb.TOTAL)).Push(mdb.VALUE, m.Cmd(aaa.USER).Length())
			m.Push("units", "")
			m.Push("units", "")
			ctx.DisplayStory(m, "stats.js")
			gdb.Event(m, STATS_TABLES)
			PushPodCmd(m, "", arg...)
		}},
	})
}
func StatsAction() ice.Actions {
	return ice.MergeActions(ice.Actions{
		STATS_TABLES: {Hand: func(m *ice.Message, arg ...string) {
			if msg := mdb.HashSelects(m.Spawn()); msg.Length() > 0 {
				m.Push(mdb.NAME, kit.Keys(m.CommandKey(), mdb.TOTAL)).Push(mdb.VALUE, msg.Length())
				m.Push("units", "")
			}
		}},
	}, gdb.EventsAction(STATS_TABLES))
}
