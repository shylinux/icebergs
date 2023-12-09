package web

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const (
	STATS_TABLES = "stats.tables"
)
const STATS = "stats"

func init() {
	Index.MergeCommands(ice.Commands{
		STATS: {Help: "汇总量", Hand: func(m *ice.Message, arg ...string) {
			defer ctx.DisplayStory(m, "")
			if m.Option(ice.MSG_USERPOD) == "" {
				PushStats(m, kit.Keys(aaa.SESS, mdb.TOTAL), m.Cmd(aaa.SESS).Length(), "", "会话总数")
				if ice.Info.Username == ice.Info.Make.Username {
					PushStats(m, kit.Keys(aaa.USER, mdb.TOTAL), m.Cmd(aaa.USER).Length()-1, "", "用户总数")
				} else {
					PushStats(m, kit.Keys(aaa.USER, mdb.TOTAL), m.Cmd(aaa.USER).Length()-2, "", "用户总数")
				}
				PushStats(m, kit.Keys(ctx.COMMAND, mdb.TOTAL), m.Cmd(ctx.COMMAND).Length(), "", "命令总数")
			}
			gdb.Event(m, STATS_TABLES)
			PushPodCmd(m, "", arg...)
		}},
	})
}
func StatsAction(arg ...string) ice.Actions {
	return ice.MergeActions(ice.Actions{
		STATS_TABLES: {Hand: func(m *ice.Message, _ ...string) {
			if msg := mdb.HashSelects(m.Spawn()); msg.Length() > 0 {
				PushStats(m, kit.Keys(m.CommandKey(), mdb.TOTAL), msg.Length(), arg...)
			}
		}},
	}, gdb.EventsAction(STATS_TABLES))
}
func PushStats(m *ice.Message, name string, value ice.Any, arg ...string) {
	kit.If(name == "", func() { name = kit.Keys(m.CommandKey(), mdb.TOTAL) })
	kit.If(value != 0, func() {
		m.Push(mdb.NAME, name).Push(mdb.VALUE, value).Push(mdb.UNITS, kit.Select("", arg, 0)).Push(ctx.TRANS, kit.Select("", arg, 1))
		m.Push(ctx.INDEX, m.PrefixKey())
	})
}
