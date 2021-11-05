package bash

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

const ZSH = "zsh"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		ZSH: {Name: ZSH, Help: "命令行", Value: kit.Data(
			cli.SOURCE, "https://nchc.dl.sourceforge.net/project/zsh/zsh/5.8/zsh-5.8.tar.xz",
		)},
	}, Commands: map[string]*ice.Command{
		ZSH: {Name: "zsh path auto order build download", Help: "命令行", Action: ice.MergeAction(map[string]*ice.Action{
			cli.ORDER: {Name: "order", Help: "加载", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(code.INSTALL, cli.ORDER, m.Config(cli.SOURCE), "_install/bin")
			}},
		}, code.InstallAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmdy(code.INSTALL, cli.SOURCE, m.Config(cli.SOURCE), arg)
		}},
	}})
}
