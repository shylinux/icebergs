package vim

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

const FAVOR = "favor"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		FAVOR: {Name: FAVOR, Help: "收藏夹", Value: kit.Data(
			kit.MDB_SHORT, kit.MDB_ZONE, kit.MDB_FIELD, "time,id,type,name,text,file,line,pwd",
		)},
	}, Commands: map[string]*ice.Command{
		"/favor": {Name: "/favor", Help: "收藏", Action: map[string]*ice.Action{
			mdb.SELECT: {Name: "select", Help: "主题", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(FAVOR).Table(func(index int, value map[string]string, head []string) {
					m.Echo(value[kit.MDB_ZONE]).Echo(ice.NL)
				})
			}},
			mdb.INSERT: {Name: "insert", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(FAVOR, mdb.INSERT)
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd(FAVOR, m.Option(kit.MDB_ZONE)).Table(func(index int, value map[string]string, head []string) {
				m.Echo("%v\n", m.Option(kit.MDB_ZONE)).Echo("%v:%v:%v:(%v): %v\n",
					value[kit.MDB_FILE], value[kit.MDB_LINE], "1", value[kit.MDB_NAME], value[kit.MDB_TEXT])
			})
		}},
		FAVOR: {Name: "favor zone id auto", Help: "收藏夹", Action: ice.MergeAction(map[string]*ice.Action{
			mdb.INSERT: {Name: "insert zone=数据结构 type name=hi text=hello file line", Help: "添加"},
			code.INNER: {Name: "inner", Help: "源码", Hand: func(m *ice.Message, arg ...string) {
				p := path.Join(m.Option(cli.PWD), m.Option(kit.MDB_FILE))
				m.ProcessCommand(code.INNER, []string{path.Dir(p) + ice.PS, path.Base(p), m.Option(kit.MDB_LINE)}, arg...)
			}},
		}, mdb.ZoneAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if mdb.ZoneSelect(m, arg...); len(arg) == 0 {
				m.Action(mdb.CREATE, mdb.EXPORT, mdb.IMPORT)
				m.PushAction(mdb.REMOVE)
			} else {
				m.PushAction(code.INNER)
				m.StatusTimeCount()
			}
		}},
	}})
}
