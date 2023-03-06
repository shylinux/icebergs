package code

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/ssh"
	kit "shylinux.com/x/toolkits"
)

func _sh_cmds(m *ice.Message, p string) (string, string) {
	cmds, text := kit.Select(SH, m.Config(ssh.SHELL)), kit.Format(strings.TrimSpace(_sh_cmd_template), m.Option(ice.MSG_USERHOST), m.Option(ice.MSG_USERPOD), p)
	if head := kit.Select("", strings.Split(m.Cmdx(nfs.CAT, p), ice.NL), 0); strings.HasPrefix(head, "#!") {
		cmds = strings.TrimSpace(strings.TrimPrefix(head, "#!"))
	}
	return cmds, text
}

const SH = nfs.SH

func init() {
	Index.MergeCommands(ice.Commands{
		SH: {Name: "sh path auto", Help: "命令", Actions: ice.MergeActions(ice.Actions{
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
				cmds, text := _sh_cmds(m, path.Join(arg[2], arg[1]))
				_xterm_show(m, cmds, text, path.Join(arg[2], arg[1]))
			}},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) {
				cmds, text := _sh_cmds(m, path.Join(arg[2], arg[1]))
				m.Cmdy(cli.SYSTEM, cmds, "-c", text).Status(ssh.SHELL, strings.ReplaceAll(text, ice.NL, "; "))
			}},
			TEMPLATE: {Hand: func(m *ice.Message, arg ...string) { m.Echo(_sh_template) }},
			NAVIGATE: {Hand: func(m *ice.Message, arg ...string) { _c_tags(m, MAN, "ctags", "-a", "-R", nfs.PWD) }},
		}, PlugAction())},
	})
}

var _sh_template = `#!/bin/sh
		
demo() {
	echo "hello world"
}
`
var _sh_cmd_template = `export ctx_dev=%s ctx_pod=%s ctx_mod=%s
temp=$(mktemp); if curl -V &>/dev/null; then curl -o $temp -fsSL $ctx_dev; else wget -O $temp -q $ctx_dev; fi && source $temp $ctx_mod
`
