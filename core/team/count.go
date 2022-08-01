package team

import (
	"strings"
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const COUNT = "count"

func init() {
	Index.MergeCommands(ice.Commands{
		COUNT: {Name: "count begin_time@date end_time@date auto insert", Help: "倒计时", Meta: kit.Dict(
			ice.Display(""),
		), Actions: ice.MergeAction(ice.Actions{
			mdb.INSERT: {Name: "insert zone type=once,step,week name text begin_time@date close_time@date", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(TASK, mdb.INSERT, arg)
			}},
		}, TASK), Hand: func(m *ice.Message, arg ...string) {
			begin_time, end_time := _plan_scope(m, 8, append([]string{LONG}, arg...)...)
			msg := _plan_list(m.Spawn(), begin_time, end_time)
			msg.SortTime(BEGIN_TIME)

			tz := int64(8)
			msg.Tables(func(value ice.Maps) {
				if value[mdb.STATUS] == CANCEL {
					return
				}

				show := []string{}
				for _, k := range []string{mdb.NAME, mdb.TEXT} {
					show = append(show, kit.Format(`<div class="%v">%v</div>`, k, value[k]))
				}

				t := (kit.Time(value[BEGIN_TIME])+int64(time.Hour)*tz)/int64(time.Second)/3600/24 - (time.Now().Unix()+3600*tz)/3600/24
				m.Echo(`<div class="item %s" title="%s">距离 %v%v%v<span class="day"> %v </span>天</div>`,
					kit.Select("gone", "come", t > 0), value[mdb.TEXT],
					strings.Split(value[BEGIN_TIME], " ")[0],
					strings.Join(show, ""),
					kit.Select("已经", "还有", t > 0), t,
				)
			})
		}},
	})
}
