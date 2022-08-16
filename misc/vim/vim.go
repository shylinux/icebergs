package vim

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/code"
)

const VIM = "vim"

var Index = &ice.Context{Name: VIM, Help: "编辑器", Commands: ice.Commands{
	VIM: {Name: "vim path auto order build download", Help: "编辑器", Actions: ice.MergeActions(ice.Actions{
		cli.BUILD: {Name: "build", Help: "构建", Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(code.INSTALL, cli.BUILD, m.Config(nfs.SOURCE), "--enable-multibyte=yes",
				"--enable-pythoninterp=yes", "--enable-luainterp=yes", "--enable-cscope=yes")
		}},
		cli.ORDER: {Name: "order", Help: "加载", Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(code.INSTALL, cli.ORDER, m.Config(nfs.SOURCE), "_install/bin")
		}},
	}, code.InstallAction(nfs.SOURCE, "http://mirrors.tencent.com/macports/distfiles/vim/vim-8.2.2681.tar.gz")), Hand: func(m *ice.Message, arg ...string) {
		m.Cmdy(code.INSTALL, nfs.SOURCE, m.Config(nfs.SOURCE), arg)
	}},
}}

func init() { code.Index.Register(Index, &web.Frame{}) }
