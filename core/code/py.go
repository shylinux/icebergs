package code

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
)

func _py_exec(m *ice.Message, arg ...string) {
	const PYTHON2 = "python2"

	if _, e := nfs.DiskFile.StatFile(path.Join(arg[2], arg[1])); e == nil {
		m.Cmdy(cli.SYSTEM, PYTHON2, path.Join(arg[2], arg[1]))

	} else if b, e := nfs.ReadFile(m, path.Join(arg[2], arg[1])); e == nil {
		m.Cmdy(cli.SYSTEM, PYTHON2, "-c", string(b))
	}

	if m.StatusTime(); cli.IsSuccess(m) {
		m.SetAppend()
	}
}

const PY = nfs.PY

func init() {
	Index.MergeCommands(ice.Commands{
		PY: {Name: "py path auto", Help: "脚本", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Go(func() {
					m.Sleep300ms()
					cli.IsAlpine(m, "python", "python2")
					cli.IsAlpine(m, "python2")
					cli.IsAlpine(m, "python3")
				})
			}},
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
				// _py_exec(m, arg...)
			}},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) {
				_py_exec(m, arg...)
			}},
			TEMPLATE: {Hand: func(m *ice.Message, arg ...string) {
				m.Echo(`print "hello world"`)
			}},
		}, PlugAction(), LangAction())},
	})
}
