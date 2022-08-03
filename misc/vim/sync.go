package vim

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

const SYNC = "sync"

func init() {
	Index.MergeCommands(ice.Commands{
		"/sync": {Name: "/sync", Help: "同步", Hand: func(m *ice.Message, arg ...string) {
			switch m.Option(ARG) {
			case "wq", "q", "qa":
				m.Cmd("/sess", aaa.LOGOUT)
			}

			m.Cmd(SYNC, mdb.INSERT, mdb.TYPE, VIMRC, mdb.NAME, arg[0], mdb.TEXT, kit.Select(m.Option(ARG), m.Option(SUB)),
				m.OptionSimple(cli.PWD, BUF, ROW, COL))
		}},
		SYNC: {Name: "sync id auto page export import", Help: "同步流", Actions: ice.MergeAction(ice.Actions{
			code.INNER: {Name: "inner", Help: "源码", Hand: func(m *ice.Message, arg ...string) {
				p := path.Join(m.Option(cli.PWD), m.Option(BUF))
				ctx.ProcessCommand(m, code.INNER, []string{path.Dir(p) + ice.PS, path.Base(p), m.Option(ROW)}, arg...)
			}},
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(FAVOR, mdb.INPUTS, arg)
			}},
			FAVOR: {Name: "favor zone=some@key type name text buf row pwd", Help: "收藏", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(FAVOR, mdb.INSERT, m.OptionSimple(mdb.ZONE, "type,name,text,pwd"),
					nfs.FILE, m.Option(BUF), nfs.LINE, m.Option(ROW))
			}},
		}, mdb.ListAction(mdb.FIELD, "time,id,type,name,text,pwd,buf,row,col")), Hand: func(m *ice.Message, arg ...string) {
			mdb.OptionPage(m, kit.Slice(arg, 1)...)
			mdb.ListSelect(m, kit.Slice(arg, 0, 1)...)
			m.PushAction(code.INNER, FAVOR)
			m.StatusTimeCountTotal(m.Config(mdb.COUNT))
		}},
	})
}
