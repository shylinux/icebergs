package vim

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/core/code"
)

const FAVOR = "favor"

func init() {
	Index.MergeCommands(ice.Commands{
		"/favor": {Name: "/favor", Help: "收藏", Actions: ice.Actions{
			mdb.SELECT: {Name: "select", Help: "主题", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(FAVOR).Tables(func(value ice.Maps) {
					m.Echo(value[mdb.ZONE]).Echo(ice.NL)
				})
			}},
			mdb.INSERT: {Name: "insert", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(FAVOR, mdb.INSERT)
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			m.Cmd(FAVOR, m.Option(mdb.ZONE)).Tables(func(value ice.Maps) {
				m.Echo("%v\n", m.Option(mdb.ZONE)).Echo("%v:%v:%v:(%v): %v\n",
					value[nfs.FILE], value[nfs.LINE], "1", value[mdb.NAME], value[mdb.TEXT])
			})
		}},
		FAVOR: {Name: "favor zone id auto", Help: "收藏夹", Actions: ice.MergeActions(ice.Actions{
			mdb.INSERT: {Name: "insert zone=数据结构 type name=hi text=hello file line", Help: "添加"},
			code.INNER: {Name: "inner", Help: "源码", Hand: func(m *ice.Message, arg ...string) {
				p := path.Join(m.Option(cli.PWD), m.Option(nfs.FILE))
				ctx.ProcessCommand(m, code.INNER, []string{path.Dir(p) + ice.PS, path.Base(p), m.Option(nfs.LINE)}, arg...)
			}},
		}, mdb.ZoneAction(
			mdb.SHORT, mdb.ZONE, mdb.FIELD, "time,id,type,name,text,file,line,pwd",
		)), Hand: func(m *ice.Message, arg ...string) {
			if mdb.ZoneSelect(m, arg...); len(arg) == 0 {
				m.Action(mdb.CREATE, mdb.EXPORT, mdb.IMPORT)
			} else {
				m.PushAction(code.INNER)
				m.StatusTimeCount()
			}
		}},
	})
}
