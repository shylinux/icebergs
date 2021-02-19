package team

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"

	"strings"
	"time"
)

const COUNT = "count"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		COUNT: {Name: "count zone id auto insert", Help: "倒计时", Meta: kit.Dict(kit.MDB_DISPLAY, COUNT), Action: map[string]*ice.Action{
			mdb.INSERT: {Name: "insert zone type=once,step,week name text begin_time@date close_time@date", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(TASK, mdb.INSERT, arg)
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 0 && arg[0] == kit.MDB_ACTION {
				m.Cmdy(TASK, arg)
				return
			}

			if _task_list(m, kit.Select("", arg, 0), kit.Select("", arg, 1)); len(arg) == 0 {
				return
			}

			m.SortTime(TaskField.BEGIN_TIME)
			m.Table(func(index int, value map[string]string, head []string) {
				if value[kit.MDB_STATUS] == TaskStatus.CANCEL {
					return
				}

				show := []string{}
				for _, k := range []string{kit.MDB_NAME, kit.MDB_TEXT} {
					show = append(show, kit.Format(`<div class="%v">%v</div>`, k, value[k]))
				}

				t := kit.Time(value[TaskField.BEGIN_TIME])/int64(time.Second)/3600/24 - time.Now().Unix()/3600/24
				m.Echo(`<div class="item %s" title="%s">距离 %v%v%v<span class="day"> %v </span>天</div>`,
					kit.Select("gone", "come", t > 0), value[kit.MDB_TEXT],
					strings.Split(value[TaskField.BEGIN_TIME], " ")[0],
					strings.Join(show, ""),
					kit.Select("已经", "还有", t > 0), t,
				)
			})
		}},
	}})
}
