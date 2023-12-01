package code

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
)

func _py_exec(m *ice.Message, p string) {
	if _, e := nfs.DiskFile.StatFile(p); e == nil {
		m.Cmdy(cli.SYSTEM, mdb.Config(m, cli.CLI), p)
	} else if b, e := nfs.ReadFile(m, p); e == nil {
		m.Cmdy(cli.SYSTEM, mdb.Config(m, cli.CLI), "-c", string(b))
	}
}

const (
	PYTHON  = "python"
	PYTHON2 = "python2"
	PYTHON3 = "python3"
)
const PY = nfs.PY

func init() {
	Index.MergeCommands(ice.Commands{
		PY: {Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				mdb.Config(m, cli.CLI, PYTHON)
				cli.IsAlpine(m, PYTHON, "python3")
				if cli.IsRedhat(m, PYTHON2, "python2") {
					mdb.Config(m, cli.CLI, PYTHON2)
				}
			}},
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
				ProcessXterm(m, mdb.Config(m, cli.CLI)+" -i "+path.Join(arg[2], arg[1]), "")
			}},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) { _py_exec(m, path.Join(arg[2], arg[1])) }},
			TEMPLATE:   {Hand: func(m *ice.Message, arg ...string) { m.Echo(nfs.Template(m, DEMO_PY)) }},
		}, PlugAction(), ctx.ConfAction(cli.CLI))},
	})
}
