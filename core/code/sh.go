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

func _sh_exec(m *ice.Message, arg ...string) {
	m.Cmdy(cli.SYSTEM, SH, "-c", kit.Format(_sh_template, m.Option(ice.MSG_USERHOST), m.Option(ice.MSG_USERPOD), path.Join(arg[2], arg[1])))
	m.StatusTime("script", kit.Renders(kit.Format(`export ctx_dev={{.Option "user.host"}}; temp=$(mktemp); wget -O $temp -q $ctx_dev; source $temp %s`, path.Join(arg[2], arg[1])), m))
}

const SH = nfs.SH

func init() {
	Index.MergeCommands(ice.Commands{
		SH: {Name: "sh path auto", Help: "命令", Actions: ice.MergeActions(ice.Actions{
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
				cmds, text := SH, kit.Format(_sh_template, m.Option(ice.MSG_USERHOST), m.Option(ice.MSG_USERPOD), path.Join(arg[2], arg[1]))
				if strings.HasPrefix(text, "#!") {
					// cmds = strings.TrimSpace(strings.SplitN(text, ice.NL, 2)[0][2:])
				}
				_xterm_show(m, cmds, text)
			}},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) {
				_sh_exec(m, arg...)
			}},
			NAVIGATE: {Hand: func(m *ice.Message, arg ...string) { _c_tags(m, MAN, "ctags", "-a", "-R", nfs.PWD) }},
		}, PlugAction())},
	})
}

var _sh_template = `#!/bin/sh
export ctx_dev=%s ctx_pod=%s ctx_mod=%s
temp=$(mktemp); if curl -V &>/dev/null; then curl -o $temp -fsSL $ctx_dev; else wget -O $temp -q $ctx_dev; fi && source $temp $ctx_mod
`
