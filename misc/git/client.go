package git

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/core/code"
)

const CLIENT = "client"

func init() {
	Index.MergeCommands(ice.Commands{
		CLIENT: {Help: "代码库", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				cli.IsRedhat(m, GIT, "git")
			}},
			cli.ORDER: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(code.INSTALL, cli.ORDER, mdb.Config(m, nfs.SOURCE), "_install/bin")
				m.Cmd(code.INSTALL, cli.ORDER, mdb.Config(m, nfs.SOURCE), "_install/libexec/git-core")
			}},
		}, code.InstallAction(nfs.SOURCE, "http://mirrors.tencent.com/macports/distfiles/git-cinnabar/git-2.31.1.tar.gz"))},
	})
}
