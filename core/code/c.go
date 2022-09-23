package code

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _c_show(m *ice.Message, arg ...string) {
	TagsList(m, "ctags", "--excmd=number", "--sort=no", "-f", "-", path.Join(m.Option(nfs.PATH), m.Option(nfs.FILE)))
}
func _c_exec(m *ice.Message, arg ...string) {
	name := strings.TrimSuffix(arg[1], path.Ext(arg[1])) + ".bin"
	if msg := m.Cmd(cli.SYSTEM, "gcc", arg[1], "-o", name, kit.Dict(cli.CMD_DIR, arg[2])); !cli.IsSuccess(msg) {
		_vimer_make(m, arg[2], msg)
		return
	}
	if m.Cmdy(cli.SYSTEM, path.Join(arg[2], name)); m.Append(cli.CMD_ERR) == "" {
		m.Result(m.Append(cli.CMD_OUT))
		m.SetAppend()
	}
	m.StatusTime()
}
func _c_tags(m *ice.Message, man string, cmd ...string) {
	if !nfs.ExistsFile(m, path.Join(m.Option(nfs.PATH), nfs.TAGS)) {
		m.Cmd(cli.SYSTEM, cmd, kit.Dict(cli.CMD_DIR, m.Option(nfs.PATH)))
	}
	if _inner_tags(m, m.Option(nfs.PATH), m.Option(mdb.NAME)); !cli.IsSuccess(m) || m.Length() == 0 {
		m.Push(nfs.FILE, kit.Keys(m.Option(mdb.NAME), man))
		m.Push(nfs.LINE, "1")
	}
}

const H = "h"
const C = "c"
const MAN = "man"

func init() {
	Index.MergeCommands(ice.Commands{
		C: {Name: "c path auto", Help: "系统", Actions: ice.MergeActions(ice.Actions{
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) { _c_show(m, arg...) }},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) { _c_exec(m, arg...) }},
			NAVIGATE:   {Hand: func(m *ice.Message, arg ...string) { _c_tags(m, MAN, "ctags", "-a", "-R", nfs.PWD) }},
		}, PlugAction(), LangAction())},
		H: {Name: "c path auto", Help: "系统", Actions: ice.MergeActions(ice.Actions{
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) { _c_show(m, arg...) }},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) { _c_exec(m, arg...) }},
			NAVIGATE:   {Hand: func(m *ice.Message, arg ...string) { _c_tags(m, MAN, "ctags", "-a", "-R", nfs.PWD) }},
		}, PlugAction(), LangAction())},
		MAN: {Name: MAN, Help: "手册", Actions: ice.MergeActions(ice.Actions{
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
				if len(arg) == 1 {
					arg = append(arg, "")
				}
				key := kit.TrimExt(arg[1], arg[0])
				m.Option(cli.CMD_ENV, "COLUMNS", kit.Int(kit.Select("1920", m.Option("width")))/12)
				m.Cmdy(cli.SYSTEM, "sh", "-c", kit.Format("man %s %s|col -b", "", key))
			}},
		}, PlugAction())},
	})
}
