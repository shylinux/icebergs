package code

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _c_show(m *ice.Message, arg ...string) {
	target := path.Join(ice.BIN, kit.TrimExt(arg[1], arg[0]))
	if msg := m.Cmd(cli.SYSTEM, "gcc", "-o", target, path.Join(arg[2], arg[1])); cli.IsSuccess(msg) {
		ProcessXterm(m, nfs.SH, target, path.Join(arg[2], arg[1]))
	} else {
		_vimer_make(m, arg[2], msg)
	}
}
func _c_exec(m *ice.Message, arg ...string) {
	target := path.Join(ice.BIN, kit.TrimExt(arg[1], arg[0]))
	if msg := m.Cmd(cli.SYSTEM, "gcc", "-o", target, path.Join(arg[2], arg[1])); cli.IsSuccess(msg) {
		m.Cmdy(cli.SYSTEM, target).StatusTime(nfs.PATH, target)
	} else {
		_vimer_make(m, arg[2], msg)
	}
}
func _c_tags(m *ice.Message, cmd ...string) {
	if !nfs.Exists(m, path.Join(m.Option(nfs.PATH), nfs.TAGS)) {
		m.Cmd(cli.SYSTEM, cmd, kit.Dict(cli.CMD_DIR, m.Option(nfs.PATH)))
	}
	_inner_tags(m, m.Option(nfs.PATH), m.Option(mdb.NAME))
}

const H = "h"
const C = "c"

func init() {
	Index.MergeCommands(ice.Commands{
		C: {Name: "c path auto", Help: "编程", Actions: ice.MergeActions(ice.Actions{
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) { _c_show(m, arg...) }},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) { _c_exec(m, arg...) }},
			TEMPLATE:   {Hand: func(m *ice.Message, arg ...string) { m.Echo(nfs.Template(m, "demo.c")) }},
			NAVIGATE:   {Hand: func(m *ice.Message, arg ...string) { _c_tags(m, "ctags", "-a", "-R", nfs.PWD) }},
		}, PlugAction())},
		H: {Name: "h path auto", Help: "编程", Actions: ice.MergeActions(ice.Actions{
			NAVIGATE: {Hand: func(m *ice.Message, arg ...string) { _c_tags(m, "ctags", "-a", "-R", nfs.PWD) }},
		}, PlugAction())},
	})
}
