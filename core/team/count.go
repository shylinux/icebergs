package team

import (
	"strings"
	"time"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

const COUNT = "count"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		COUNT: {Name: "count begin_time@date end_time@date auto insert", Help: "倒计时", Meta: kit.Dict(kit.MDB_DISPLAY, COUNT), Action: map[string]*ice.Action{
			mdb.INSERT: {Name: "insert zone type=once,step,week name text begin_time@date close_time@date", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(TASK, mdb.INSERT, arg)
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			begin_time, end_time := _task_scope(m, 8, append([]string{LONG}, arg...)...)
			msg := _plan_list(m.Spawn(), begin_time, end_time)
			// m.PushPodCmd(COUNT, arg...)
			msg.SortTime(BEGIN_TIME)

			tz := int64(8)
			msg.Table(func(index int, value map[string]string, head []string) {
				if value[kit.MDB_STATUS] == CANCEL {
					return
				}

				show := []string{}
				for _, k := range []string{kit.MDB_NAME, kit.MDB_TEXT} {
					show = append(show, kit.Format(`<div class="%v">%v</div>`, k, value[k]))
				}

				t := (kit.Time(value[BEGIN_TIME])+int64(time.Hour)*tz)/int64(time.Second)/3600/24 - (time.Now().Unix()+3600*tz)/3600/24
				m.Echo(`<div class="item %s" title="%s">距离 %v%v%v<span class="day"> %v </span>天</div>`,
					kit.Select("gone", "come", t > 0), value[kit.MDB_TEXT],
					strings.Split(value[BEGIN_TIME], " ")[0],
					strings.Join(show, ""),
					kit.Select("已经", "还有", t > 0), t,
				)
			})
		}},
	}})
}
