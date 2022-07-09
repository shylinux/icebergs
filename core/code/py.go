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
	if kit.FileExists(kit.Path(arg[2], arg[1])) {
		m.Cmdy(cli.SYSTEM, PYTHON2, kit.Path(arg[2], arg[1]))
	} else if b, ok := ice.Info.Pack[path.Join(arg[2], arg[1])]; ok && len(b) > 0 {
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
		PY: {Name: "py path auto", Help: "脚本", Actions: ice.MergeAction(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(mdb.ENGINE, mdb.CREATE, PY, m.PrefixKey())
				m.Cmd(mdb.RENDER, mdb.CREATE, PY, m.PrefixKey())
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
