package code

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _c_show(m *ice.Message, arg ...string) { TagsList(m) }
func _c_exec(m *ice.Message, arg ...string) {
	target := path.Join(ice.BIN, kit.TrimExt(arg[1], arg[0]))
	if msg := m.Cmd(cli.SYSTEM, "gcc", "-o", target, path.Join(arg[2], arg[1])); cli.IsSuccess(msg) {
		m.Cmdy(cli.SYSTEM, target).StatusTime(nfs.PATH, target)
	} else {
		_vimer_make(m, arg[2], msg)
	}
}
func _c_tags(m *ice.Message, man string, cmd ...string) {
	if !nfs.ExistsFile(m, path.Join(m.Option(nfs.PATH), nfs.TAGS)) {
		m.Cmd(cli.SYSTEM, cmd, kit.Dict(cli.CMD_DIR, m.Option(nfs.PATH)))
	}
	if _inner_tags(m, m.Option(nfs.PATH), m.Option(mdb.NAME)); !cli.IsSuccess(m) || m.Length() == 0 {
		m.Push(nfs.FILE, kit.Keys(m.Option(mdb.NAME), man)).Push(nfs.LINE, "1")
	}
}

const MAN = "man"
const H = "h"
const C = "c"

func init() {
	Index.MergeCommands(ice.Commands{
		MAN: {Name: MAN, Help: "系统手册", Actions: ice.MergeActions(ice.Actions{
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) { m.Cmd(cli.SYSTEM, cli.MAN, kit.TrimExt(arg[1], arg[0])) }},
		}, PlugAction())},
		H: {Name: "h path auto", Help: "系统编程", Actions: ice.MergeActions(ice.Actions{
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) { _c_show(m, arg...) }},
			NAVIGATE:   {Hand: func(m *ice.Message, arg ...string) { _c_tags(m, MAN, "ctags", "-a", "-R", nfs.PWD) }},
		}, PlugAction(), LangAction())},
		C: {Name: "c path auto", Help: "系统编程", Actions: ice.MergeActions(ice.Actions{
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) { _c_show(m, arg...) }},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) { _c_exec(m, arg...) }},
			NAVIGATE:   {Hand: func(m *ice.Message, arg ...string) { _c_tags(m, MAN, "ctags", "-a", "-R", nfs.PWD) }},
			TEMPLATE:   {Hand: func(m *ice.Message, arg ...string) { kit.If(arg[0] == C, func() { m.Echo(_c_template) }) }},
		}, PlugAction(), LangAction())},
	})
}

var _c_template = `#include <stdio.h>

int main(int argc, char *argv[]) {
	printf("hello world\n");
}
`
