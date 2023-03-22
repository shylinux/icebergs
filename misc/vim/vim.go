package vim

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/code"
)

const VIM = "vim"

var Index = &ice.Context{Name: VIM, Help: "编辑器", Commands: ice.Commands{
	VIM: {Name: "vim path auto order build download", Help: "编辑器", Actions: ice.MergeActions(ice.Actions{
		cli.BUILD: {Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(code.INSTALL, cli.BUILD, mdb.Config(m, nfs.SOURCE), "--enable-multibyte=yes", "--enable-cscope=yes", "--enable-luainterp=yes", "--enable-pythoninterp=yes")
		}},
	}, code.InstallAction(nfs.SOURCE, "http://mirrors.tencent.com/macports/distfiles/vim/vim-8.2.2681.tar.gz")), Hand: func(m *ice.Message, arg ...string) {
		m.Cmdy(code.INSTALL, ctx.ConfigSimple(m, nfs.SOURCE), arg)
	}},
}}

func init() { code.Index.Register(Index, &web.Frame{}) }
