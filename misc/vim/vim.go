package vim

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

func _vim_pkg(m *ice.Message) string {
	return kit.Replace(kit.TrimExt(m.Conf(VIM, kit.Keym(cli.BUILD))), "_.", "")
}

const VIMRC = "vimrc"
const VIM = "vim"

var Index = &ice.Context{Name: VIM, Help: "编辑器",
	Commands: map[string]*ice.Command{
		VIM: {Name: "vim port path auto start build download", Help: "编辑器", Action: map[string]*ice.Action{
			web.DOWNLOAD: {Name: "download", Help: "下载", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(code.INSTALL, web.DOWNLOAD, m.Conf(VIM, kit.Keym(cli.SOURCE)))
			}},
			cli.BUILD: {Name: "build", Help: "构建", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(code.INSTALL, cli.BUILD, _vim_pkg(m), m.Confv(VIM, kit.Keym(cli.BUILD)))
			}},
			cli.START: {Name: "start", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(code.INSTALL, cli.START, _vim_pkg(m), "bin/vim")
			}},

			mdb.PLUGIN: {Hand: func(m *ice.Message, arg ...string) {
				m.Echo(m.Conf(VIM, kit.Keym(code.PLUG)))
			}},
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(nfs.CAT, path.Join(arg[2], arg[1]))
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmdy(code.INSTALL, path.Base(m.Conf(VIM, kit.Keym(cli.SOURCE))), arg)
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
	},
	Configs: map[string]*ice.Config{
		VIM: {Name: VIM, Help: "编辑器", Value: kit.Data(
			cli.SOURCE, "http://ftp.vim.org/pub/vim/unix/vim-8.1.tar.bz2",
			cli.BUILD, []interface{}{
				"--enable-multibyte=yes",
				"--enable-pythoninterp=yes",
				"--enable-luainterp=yes",
				"--enable-cscope=yes",
			},

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
	},
}

func init() { code.Index.Register(Index, &web.Frame{}) }
