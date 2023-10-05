package code

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/ssh"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _sh_cmds(m *ice.Message, p string) (string, string) {
	cmds, text := kit.Select(SH, mdb.Config(m, ssh.SHELL)), kit.Format(strings.TrimSpace(nfs.Template(m, "cmd.sh")), web.UserHost(m), m.Option(ice.MSG_USERPOD), p)
	if head := kit.Select("", strings.Split(m.Cmdx(nfs.CAT, p), lex.NL), 0); strings.HasPrefix(head, "#!") {
		cmds = strings.TrimSpace(strings.TrimPrefix(head, "#!"))
	}
	return cmds, text
}

const (
	BASH = "bash"
	CONF = "conf"
	VIM  = "vim"
	ISH  = "ish"
)
const SH = nfs.SH

func init() {
	Index.MergeCommands(ice.Commands{
		SH: {Actions: ice.MergeActions(ice.Actions{
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
				cmds, text := _sh_cmds(m, path.Join(arg[2], arg[1]))
				ProcessXterm(m, cmds, text, path.Join(arg[2], arg[1]))
			}},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) {
				cmds, text := _sh_cmds(m, path.Join(arg[2], arg[1]))
				m.Cmdy(cli.SYSTEM, cmds, "-c", text).Status(ssh.SHELL, strings.ReplaceAll(text, lex.NL, "; "))
			}},
			TEMPLATE: {Hand: func(m *ice.Message, arg ...string) { m.Echo(nfs.Template(m, DEMO_SH)) }},
			NAVIGATE: {Hand: func(m *ice.Message, arg ...string) { _c_tags(m, CTAGS, "-a", "-R", nfs.PWD) }},
		}, PlugAction())},
	})
}
