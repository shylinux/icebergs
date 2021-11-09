package chrome

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const FIELD = "field"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		FIELD: {Name: "field", Help: "工具", Value: kit.Data(
			kit.MDB_SHORT, kit.MDB_ZONE, kit.MDB_FIELD, "time,id,index,args,style,left,top,right,bottom,selection",
		)},
	}, Commands: map[string]*ice.Command{
		FIELD: {Name: "field zone id auto insert", Help: "工具", Action: ice.MergeAction(map[string]*ice.Action{
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case kit.MDB_ZONE:
					m.Cmdy(CHROME, mdb.INPUTS)
				case kit.MDB_INDEX:
					m.Cmdy(ctx.COMMAND)
				}
			}},
			mdb.INSERT: {Name: "insert zone=golang.google.cn index=cli.system args=pwd", Help: "添加"},
			ctx.COMMAND: {Name: "command", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(FIELD, m.Option(tcp.HOST), arg).Table(func(index int, value map[string]string, head []string) {
					if len(arg) == 0 { // 命令列表
						m.Option(ice.MSG_OPTS, head)
						for k, v := range value {
							m.Option(k, v)
						}
						m.Cmdy(web.SPACE, CHROME, CHROME, "1", m.Option("tid"), FIELD, value[kit.MDB_ID], value[kit.MDB_ARGS])
					} else { // 命令详情
						m.Cmdy(ctx.COMMAND, value[kit.MDB_INDEX])
					}
				})
			}},
			ice.RUN: {Name: "run", Help: "执行", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(FIELD, m.Option(tcp.HOST), arg[0]).Table(func(index int, value map[string]string, head []string) {
					m.Cmdy(value[kit.MDB_INDEX], arg[1:])
				})
			}},
		}, mdb.ZoneAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			mdb.ZoneSelect(m, arg...)
		}},
	}})
}
