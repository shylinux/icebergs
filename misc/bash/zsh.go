package bash

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/core/code"
)

const ZSH = "zsh"

func init() {
	Index.MergeCommands(ice.Commands{
		ZSH: {Name: "zsh path auto order build download", Help: "命令行", Actions: ice.MergeAction(ice.Actions{
			cli.ORDER: {Name: "order", Help: "加载", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(code.INSTALL, cli.ORDER, m.Config(nfs.SOURCE), "_install/bin")
			}},
		}, code.InstallAction(nfs.SOURCE, "https://nchc.dl.sourceforge.net/project/zsh/zsh/5.8/zsh-5.8.tar.xz")), Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(code.INSTALL, nfs.SOURCE, m.Config(nfs.SOURCE), arg)
		}},
	})
}
