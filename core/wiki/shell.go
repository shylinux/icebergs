package wiki

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _shell_show(m *ice.Message, name, text string, arg ...string) {
	m.Option(OUTPUT, cli.SystemCmds(m, m.Option(INPUT, text)))
	_wiki_template(m, SHELL, name, text, arg...)
}

const (
	INPUT  = "input"
	OUTPUT = "output"
)
const SHELL = "shell"

func init() {
	Index.Merge(&ice.Context{Commands: ice.Commands{
		SHELL: {Name: "shell [name] cmd", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 1 {
				m.Cmdy(SPARK, SHELL, arg)
			} else {
				m.Cmdy(SPARK, arg)
			}
			return
			arg = _name(m, arg)
			_shell_show(m, arg[0], kit.Select(arg[0], arg[1]), arg[2:]...)
		}},
	}, Configs: ice.Configs{
		SHELL: {Name: SHELL, Help: "命令", Value: kit.Data(
			nfs.TEMPLATE, `<code {{.OptionTemplate}}>$ {{.Option "input"}} # {{.Option "name"}}
{{.Option "output"}}</code>`,
		)},
	}})
}
