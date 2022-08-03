package git

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

const GIT = "git"

var Index = &ice.Context{Name: GIT, Help: "代码库", Configs: ice.Configs{
	GIT: {Name: GIT, Help: "代码库", Value: kit.Data(
		nfs.SOURCE, "http://mirrors.tencent.com/macports/distfiles/git-cinnabar/git-2.31.1.tar.gz",
	)},
}, Commands: ice.Commands{
	GIT: {Name: "git path auto install order build download", Help: "代码库", Actions: ice.MergeAction(ice.Actions{
		code.INSTALL: {Name: "install", Help: "安装", Hand: func(m *ice.Message, arg ...string) {
			web.PushStream(m)
			defer m.ProcessInner()

			m.Cmdy(cli.SYSTEM, "yum", "install", "-y", "git")
		}},
		cli.ORDER: {Name: "order", Help: "加载", Hand: func(m *ice.Message, arg ...string) {
			m.Cmd(code.INSTALL, cli.ORDER, m.Config(nfs.SOURCE), "_install/bin")
			m.Cmdy(code.INSTALL, cli.ORDER, m.Config(nfs.SOURCE), "_install/libexec/git-core")
		}},
	}, code.InstallAction()), Hand: func(m *ice.Message, arg ...string) {
		m.Cmdy(code.INSTALL, nfs.SOURCE, m.Config(nfs.SOURCE), arg)
		m.Echo("hello world %v", arg)
	}},
}}

func init() { code.Index.Register(Index, &web.Frame{}, REPOS) }
