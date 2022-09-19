package code

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _py_main_script(m *ice.Message, arg ...string) {
	const PYTHON2 = "python2"

	if _, e := nfs.DiskFile.StatFile(path.Join(arg[2], arg[1])); e == nil {
		m.Cmdy(cli.SYSTEM, PYTHON2, path.Join(arg[2], arg[1]))

	} else if b, e := nfs.ReadFile(m, path.Join(arg[2], arg[1])); e == nil {
		m.Cmdy(cli.SYSTEM, PYTHON2, "-c", string(b))
	}
	if m.StatusTime(); cli.IsSuccess(m) {
		m.SetAppend()
	}
	m.Echo(ice.NL)
}

const PY = nfs.PY

func init() {
	Index.Merge(&ice.Context{Commands: ice.Commands{
		PY: {Name: "py path auto", Help: "脚本", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Go(func() {
					m.Sleep300ms() // after runtime init
					cli.IsAlpine(m, "python", "python2")
					cli.IsAlpine(m, "python2")
					cli.IsAlpine(m, "python3")
				})
				m.Cmd(mdb.ENGINE, mdb.CREATE, PY, m.PrefixKey())
				m.Cmd(mdb.RENDER, mdb.CREATE, PY, m.PrefixKey())
				m.Cmd(TEMPLATE, mdb.CREATE, m.CommandKey(), m.PrefixKey())
			}},
			TEMPLATE: {Hand: func(m *ice.Message, arg ...string) {
				if kit.Ext(m.Option(mdb.FILE)) != m.CommandKey() {
					return
				}
				m.Echo(`
print "hello world"
`)
			}},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) {
				_py_main_script(m, arg...)
			}},
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
				_py_main_script(m, arg...)
			}},
		}, PlugAction()), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) > 0 && kit.Ext(arg[0]) == PY {
				_py_main_script(m, PY, arg[0], ice.SRC)
				return
			}
			m.Option(nfs.DIR_DEEP, ice.TRUE)
			m.Option(nfs.DIR_ROOT, ice.SRC)
			m.Option(nfs.DIR_REG, ".*.(py)$")
			m.Cmdy(nfs.DIR, arg)
		}},
	}, Configs: ice.Configs{
		PY: {Name: PY, Help: "脚本", Value: kit.Data(PLUG, kit.Dict(
			SPLIT, kit.Dict(SPACE, " ", OPERATE, "{[(.,;!|<>)]}"),
			PREFIX, kit.Dict("#!", COMMENT, "# ", COMMENT), SUFFIX, kit.Dict(" {", COMMENT),
			PREPARE, kit.Dict(
				KEYWORD, kit.Simple(
					"import", "from", "return",
				),
				FUNCTION, kit.Simple(
					"print",
				),
			), KEYWORD, kit.Dict(),
		))},
	}})
}
