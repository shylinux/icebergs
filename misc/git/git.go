package git

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/code"
)

func _git_cmd(m *ice.Message, arg ...string) *ice.Message { return m.Cmd(cli.SYSTEM, GIT, arg) }
func _git_cmds(m *ice.Message, arg ...string) string { return _git_cmd(m, arg...).Result() }

const GIT = "git"

var Index = &ice.Context{Name: GIT, Help: "代码库", Commands: ice.Commands{
	GIT: {Name: "git path auto order build download", Help: "代码库", Actions: ice.MergeActions(ice.Actions{
		cli.ORDER: {Name: "order", Help: "加载", Hand: func(m *ice.Message, arg ...string) {
			m.Cmd(code.INSTALL, cli.ORDER, m.Config(nfs.SOURCE), "_install/libexec/git-core")
			m.Cmdy(code.INSTALL, cli.ORDER, m.Config(nfs.SOURCE), "_install/bin")
		}},
	}, code.InstallAction(nfs.SOURCE, "http://mirrors.tencent.com/macports/distfiles/git-cinnabar/git-2.31.1.tar.gz")), Hand: func(m *ice.Message, arg ...string) {
		m.Cmdy(code.INSTALL, nfs.SOURCE, m.Config(nfs.SOURCE), arg)
	}},
}}

func init() { code.Index.Register(Index, &web.Frame{}, REPOS) }
