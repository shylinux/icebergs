package vim

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/code"
)

const FAVOR = "favor"

func init() {
	Index.MergeCommands(ice.Commands{
		FAVOR: {Name: "favor zone id auto insert", Help: "收藏夹", Actions: ice.MergeActions(ice.Actions{
			mdb.INSERT: {Name: "insert zone=数据结构 type name=hi text=hello file line"},
			code.INNER: {Help: "源码", Hand: func(m *ice.Message, arg ...string) {
				if len(arg) > 0 && arg[0] == ice.RUN {
					ctx.ProcessField(m, "", nil, arg...)
				} else {
					p := path.Join(m.Option(cli.PWD), m.Option(nfs.FILE))
					ctx.ProcessField(m, "", []string{path.Dir(p) + ice.PS, path.Base(p), m.Option(nfs.LINE)})
				}
			}},
		}, mdb.ZoneAction(mdb.FIELD, "time,id,type,name,text,file,line,pwd")), Hand: func(m *ice.Message, arg ...string) {
			if mdb.ZoneSelect(m, arg...).PushAction(code.INNER); len(arg) == 0 {
				m.Action(mdb.EXPORT, mdb.IMPORT)
			}
		}},
		web.PP(FAVOR): {Actions: ice.Actions{
			mdb.INSERT: {Hand: func(m *ice.Message, arg ...string) { m.Cmd(FAVOR, mdb.INSERT) }},
			mdb.SELECT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(FAVOR, func(value ice.Maps) { m.Echo(value[mdb.ZONE]).Echo(ice.NL) })
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			m.Cmd(FAVOR, m.Option(mdb.ZONE), func(value ice.Maps) {
				m.Echo("%v\n", m.Option(mdb.ZONE)).Echo("%v:%v:%v:(%v): %v\n", value[nfs.FILE], value[nfs.LINE], "1", value[mdb.NAME], value[mdb.TEXT])
			})
		}},
	})
}
