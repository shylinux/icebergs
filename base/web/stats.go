package web

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
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
				PushStats(m, "", "", "", "注册总数", aaa.APPLY)
				PushStats(m, "", "", "", "邀请总数", aaa.OFFER)
				if ice.Info.Username == ice.Info.Make.Username {
					PushStats(m, "", m.Cmd(aaa.USER).Length()-1, "", "用户总数", aaa.USER)
				} else {
					PushStats(m, "", m.Cmd(aaa.USER).Length()-2, "", "用户总数", aaa.USER)
				}
				PushStats(m, "", "", "", "会话总数", aaa.SESS)
				PushStats(m, "", "", "", "令牌总数", TOKEN)
				PushStats(m, "", "", "", "共享总数", SHARE)
				PushStats(m, "", "", "", "命令总数", ctx.COMMAND)
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
	index := kit.Select(m.PrefixKey(), arg, 2)
	kit.If(name == "", func() { name = kit.Keys(kit.Select("", kit.Split(index, nfs.PT), -1), mdb.TOTAL) })
	kit.If(value == "", func() { value = m.Cmd(index).Length() })
	kit.If(value != 0, func() {
		m.Push(mdb.NAME, name).Push(mdb.VALUE, value).Push(mdb.UNITS, kit.Select("", arg, 0)).Push(ctx.TRANS, kit.Select("", arg, 1)).Push(ctx.INDEX, index)
	})
}
