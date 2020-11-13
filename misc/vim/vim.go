package vim

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/gdb"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/code"
	kit "github.com/shylinux/toolkits"

	"path"
	"strings"
)

func _vim_pkg(m *ice.Message) string {
	return strings.Replace(strings.Replace(kit.TrimExt(m.Conf(VIM, kit.META_SOURCE)), ".", "", -1), "-", "", -1)
}

const VIMRC = "vimrc"
const VIM = "vim"

var Index = &ice.Context{Name: VIM, Help: "编辑器",
	Commands: map[string]*ice.Command{
		VIM: {Name: "vim port path auto start build download", Help: "编辑器", Action: map[string]*ice.Action{
			web.DOWNLOAD: {Name: "download", Help: "下载", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(code.INSTALL, web.DOWNLOAD, m.Conf(VIM, kit.META_SOURCE))
			}},
			gdb.BUILD: {Name: "build", Help: "构建", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(code.INSTALL, gdb.BUILD, _vim_pkg(m), m.Confv(VIM, "meta.build"))
			}},
			gdb.START: {Name: "start", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
				m.Optionv("prepare", func(p string) []string {
					list := kit.Simple(m.Confv(VIM, "meta.start"))
					for i := 0; i < len(list); i += 2 {
						m.Cmd(web.SPIDE, web.SPIDE_DEV, web.SPIDE_SAVE, path.Join(p, list[i]),
							web.SPIDE_GET, "/share/local/usr/intshell/misc/vim/"+list[i+1])
					}
					return []string{}
					return []string{"-T", "screen", "-c", "PlugInstall", "-c", "exit", "-c", "exit"}
				})
				m.Cmdy(code.INSTALL, gdb.START, _vim_pkg(m), "bin/vim")

				// 安装插件
				m.Echo("\n")
				m.Echo("vim -c PlugInstall\n")
				m.Echo("vim -c GoInstallBinaries\n")
			}},

			mdb.PLUGIN: {Hand: func(m *ice.Message, arg ...string) {
				m.Echo(m.Conf(VIM, "meta.plug"))
			}},
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
				if strings.HasPrefix(arg[2], "http") {
					m.Cmdy(web.SPIDE, web.SPIDE_DEV, web.SPIDE_RAW, web.SPIDE_GET, arg[2]+arg[1])
					return
				}
				m.Cmdy(nfs.CAT, path.Join(arg[2], arg[1]))
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmdy(code.INSTALL, path.Base(m.Conf(VIM, kit.META_SOURCE)), arg)
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
		VIM: {Name: "vim", Help: "编辑器", Value: kit.Data(
			"source", "ftp://ftp.vim.org/pub/vim/unix/vim-8.1.tar.bz2",
			"build", []interface{}{
				"--enable-multibyte=yes",
				"--enable-pythoninterp=yes",
				"--enable-luainterp=yes",
				"--enable-cscope=yes",
			},
			"start", []interface{}{
				".vimrc", "vimrc",
				".vim/autoload/plug.vim", "plug.vim",
				".vim/autoload/auto.vim", "auto.vim",
				".vim/syntax/javascript.vim", "javascript.vim",
				".vim/syntax/shy.vim", "shy.vim",
				".vim/syntax/shy.vim", "sh.vim",
				".vim/syntax/go.vim", "go.vim",
			},

			"plug", kit.Dict(
				"split", kit.Dict(
					"space", " \t",
					"operator", "{[(&.,;!|<>)]}",
				),
				"prefix", kit.Dict(
					"\"", "comment",
				),
				"keyword", kit.Dict(
					"colorscheme", "keyword",
					"highlight", "keyword",
					"syntax", "keyword",

					"nnoremap", "keyword",
					"cnoremap", "keyword",
					"inoremap", "keyword",

					"autocmd", "keyword",
					"command", "keyword",
					"execute", "keyword",

					"set", "keyword",
					"let", "keyword",
					"if", "keyword",
					"else", "keyword",
					"elseif", "keyword",
					"endif", "keyword",
					"end", "keyword",
					"for", "keyword",
					"in", "keyword",
					"continue", "keyword",
					"break", "keyword",
					"endfor", "keyword",
					"try", "keyword",
					"catch", "keyword",
					"finally", "keyword",
					"endtry", "keyword",

					"call", "keyword",
					"return", "keyword",
					"source", "keyword",
					"finish", "keyword",
					"function", "keyword",
					"endfunction", "keyword",

					"has", "function",
					"type", "function",
					"empty", "function",
					"exists", "function",
					"executable", "function",
				),
			),
		)},
	},
}

func init() { code.Index.Register(Index, &web.Frame{}) }
