package vim

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

func _vim_pkg(m *ice.Message, url string) string {
	p := path.Join(m.Conf(code.INSTALL, kit.META_PATH), path.Base(url))
	return kit.Path(m.Conf(code.INSTALL, kit.META_PATH), strings.Split(m.Cmdx(cli.SYSTEM, "sh", "-c", kit.Format("tar tf %s| head -n1", p)), "/")[0])
}

const VIMRC = "vimrc"
const VIM = "vim"

var Index = &ice.Context{Name: VIM, Help: "编辑器", Commands: map[string]*ice.Command{
	VIM: {Name: "vim path auto order build download", Help: "编辑器", Action: ice.MergeAction(map[string]*ice.Action{
		cli.BUILD: {Name: "build", Help: "构建", Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(code.INSTALL, cli.BUILD, _vim_pkg(m, m.Config(cli.SOURCE)), "--enable-multibyte=yes",
				"--enable-pythoninterp=yes", "--enable-luainterp=yes", "--enable-cscope=yes")
		}},

		mdb.PLUGIN: {Hand: func(m *ice.Message, arg ...string) {
			m.Echo(m.Config(code.PLUG))
		}},
		mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(nfs.CAT, path.Join(arg[2], arg[1]))
		}},
	}, code.InstallAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		m.Cmdy(code.INSTALL, cli.SOURCE, _vim_pkg(m, m.Config(cli.SOURCE)), arg)
	}},

	ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		m.Save()
	}},
	ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		m.Load()
		m.Cmd(mdb.PLUGIN, mdb.CREATE, VIM, m.Prefix(VIM))
		m.Cmd(mdb.RENDER, mdb.CREATE, VIM, m.Prefix(VIM))
		m.Cmd(mdb.PLUGIN, mdb.CREATE, VIMRC, m.Prefix(VIM))
		m.Cmd(mdb.RENDER, mdb.CREATE, VIMRC, m.Prefix(VIM))
	}},
}, Configs: map[string]*ice.Config{
	VIM: {Name: VIM, Help: "编辑器", Value: kit.Data(
		cli.SOURCE, "http://mirrors.tencent.com/macports/distfiles/vim/vim-8.2.2681.tar.gz",

		code.PLUG, kit.Dict(
			code.SPLIT, kit.Dict(
				"space", " \t",
				"operator", "{[(&.,;!|<>)]}",
			),
			code.PREFIX, kit.Dict(
				"\"", "comment",
			),
			code.PREPARE, kit.Dict(
				code.KEYWORD, kit.Simple(
					"source", "finish",
					"set", "let", "end",
					"if", "else", "elseif", "endif",
					"for", "in", "continue", "break", "endfor",
					"try", "catch", "finally", "endtry",
					"call", "function", "return", "endfunction",

					"autocmd", "command", "execute",
					"nnoremap", "cnoremap", "inoremap",
					"colorscheme", "highlight", "syntax",
				),
				code.FUNCTION, kit.Simple(
					"has", "type", "empty",
					"exists", "executable",
				),
			),
		),
	)},
}}

func init() { code.Index.Register(Index, &web.Frame{}) }
