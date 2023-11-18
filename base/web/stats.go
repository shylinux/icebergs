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
		STATS: {Help: "汇总量", Meta: kit.Dict(
			ice.CTX_TRANS, kit.Dict(html.INPUT, kit.Dict(
				"goods.amount", "商品总额",
				"goods.count", "商品数量",
				"asset.amount", "资产总额",
				"asset.count", "资产数量",
				"task.total", "任务总数",
				"dream.total", "空间总数",
				"dream.start", "已启动空间",
				"repos.total", "代码库总数",
				"command.total", "命令总数",
				"share.total", "共享总数",
				"token.total", "令牌总数",
				"user.total", "用户总数",
				"sess.total", "会话总数",
				"cpu.total", "处理器核数",
				"cpu.used", "处理器使用率",
				"mem.used", "内存用量",
				"mem.total", "内存总量",
				"disk.used", "磁盘用量",
				"disk.total", "磁盘总量",
			)),
		), Hand: func(m *ice.Message, arg ...string) {
			defer ctx.DisplayStory(m, "")
			if m.Option(ice.MSG_USERPOD) == "" {
				PushStats(m, kit.Keys(aaa.SESS, mdb.TOTAL), m.Cmd(aaa.SESS).Length(), "")
				if ice.Info.Username == ice.Info.Make.Username {
					PushStats(m, kit.Keys(aaa.USER, mdb.TOTAL), m.Cmd(aaa.USER).Length()-1, "")
				} else {
					PushStats(m, kit.Keys(aaa.USER, mdb.TOTAL), m.Cmd(aaa.USER).Length()-2, "")
				}
				PushStats(m, kit.Keys(ctx.COMMAND, mdb.TOTAL), m.Cmd(ctx.COMMAND).Length(), "")
			}
			gdb.Event(m, STATS_TABLES)
			PushPodCmd(m, "", arg...)
		}},
	})
}
func StatsAction() ice.Actions {
	return ice.MergeActions(ice.Actions{
		STATS_TABLES: {Hand: func(m *ice.Message, arg ...string) {
			if msg := mdb.HashSelects(m.Spawn()); msg.Length() > 0 {
				PushStats(m, kit.Keys(m.CommandKey(), mdb.TOTAL), msg.Length(), "")
			}
		}},
	}, gdb.EventsAction(STATS_TABLES))
}
func PushStats(m *ice.Message, name string, value ice.Any, units string) {
	kit.If(value != 0, func() { m.Push(mdb.NAME, name).Push(mdb.VALUE, value).Push(mdb.UNITS, units) })
}
