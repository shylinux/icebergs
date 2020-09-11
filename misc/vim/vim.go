package vim

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/code"
	kit "github.com/shylinux/toolkits"

	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

const VIM = "vim"
const VIMRC = "vimrc"

var Index = &ice.Context{Name: VIM, Help: "编辑器",
	Commands: map[string]*ice.Command{
		VIM: {Name: "vim port=auto path=auto auto 启动 构建 下载", Help: "编辑器", Action: map[string]*ice.Action{
			"download": {Name: "download", Help: "下载", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(code.INSTALL, "download", m.Conf(VIM, kit.META_SOURCE))
			}},
			"build": {Name: "build", Help: "构建", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(code.INSTALL, "build", strings.Replace(strings.Replace(kit.TrimExt(m.Conf(VIM, kit.META_SOURCE)), ".", "", -1), "-", "", -1))
			}},
			"start": {Name: "start", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
				m.Optionv("prepare", func(p string) []string {
					return []string{}
					list := kit.Simple(m.Confv(VIM, "meta.start"))
					for i := 0; i < len(list); i += 2 {
						m.Cmd(web.SPIDE, "dev", web.SPIDE_SAVE, path.Join(os.Getenv("HOME"), list[i]),
							web.SPIDE_GET, m.Conf(VIM, "meta.remote")+list[i+1])
					}
					return []string{"-T", "screen", "-c", "PlugInstall", "-c", "exit", "-c", "exit"}
				})
				m.Cmdy(code.INSTALL, "start", strings.Replace(strings.Replace(kit.TrimExt(m.Conf(VIM, kit.META_SOURCE)), ".", "", -1), "-", "", -1), "bin/vim")

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
					m.Cmdy(web.SPIDE, "dev", "raw", "GET", arg[2]+arg[1])
					return
				}
				m.Cmdy(nfs.CAT, path.Join(arg[2], arg[1]))
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmdy(code.INSTALL, path.Base(m.Conf(VIM, kit.META_SOURCE)), arg)
		}},

		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Conf(web.FAVOR, "meta.render.vimrc", m.AddCmd(&ice.Command{Name: "render favor id", Help: "渲染引擎", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				value := m.Optionv("value").(map[string]interface{})
				switch value["name"] {
				case "read", "write", "exec":
					p := path.Join(kit.Format(kit.Value(value, "extra.pwd")), kit.Format(kit.Value(value, "extra.buf")))
					if strings.HasPrefix(kit.Format(kit.Value(value, "extra.buf")), "/") {
						p = path.Join(kit.Format(kit.Value(value, "extra.buf")))
					}

					f, e := os.Open(p)
					m.Assert(e)
					defer f.Close()
					b, e := ioutil.ReadAll(f)
					m.Assert(e)
					m.Echo(string(b))
				default:
					m.Cmdy(cli.SYSTEM, "sed", "-n", fmt.Sprintf("/%s/,/^}$/p", value["text"]), kit.Value(value, "extra.buf"))
				}
			}}))

			m.Cmd(mdb.PLUGIN, mdb.CREATE, VIMRC, VIM, c.Cap(ice.CTX_FOLLOW))
			m.Cmd(mdb.RENDER, mdb.CREATE, VIMRC, VIM, c.Cap(ice.CTX_FOLLOW))
			m.Cmd(mdb.PLUGIN, mdb.CREATE, VIM, VIM, c.Cap(ice.CTX_FOLLOW))
			m.Cmd(mdb.RENDER, mdb.CREATE, VIM, VIM, c.Cap(ice.CTX_FOLLOW))

			m.Cmd("web.spide_rewrite", "create", "from", "ftp://ftp.vim.org/pub/vim/unix/vim-8.1.tar.bz2", "to", "http://localhost:9020/publish/vim-8.1.tar.bz2")
			m.Cmd("web.spide_rewrite", "create", "from", "https://raw.githubusercontent.com/shylinux/contexts/master/etc/conf/vimrc", "to", "http://localhost:9020/publish/vimrc")
			m.Cmd("web.spide_rewrite", "create", "from", "https://raw.githubusercontent.com/shylinux/contexts/master/etc/conf/plug.vim", "to", "http://localhost:9020/publish/plug.vim")
			m.Cmd("nfs.file_rewrite", "create", "from", "etc/conf/plug.vim", "to", "https://raw.githubusercontent.com/shylinux/contexts/master/etc/conf/plug.vim")
		}},

		"/sync": {Name: "/sync", Help: "同步", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Render(ice.RENDER_RESULT)
			switch arg[0] {
			case "read", "write", "exec", "insert":
				m.Cmd(web.FAVOR, m.Conf(VIM, "meta.history"), web.TYPE_VIMRC, arg[0], kit.Select(m.Option("arg"), m.Option("sub")),
					"pwd", m.Option("pwd"), "buf", m.Option("buf"), "row", m.Option("row"), "col", m.Option("col"))
			case "trans":
				if m.Cmdy(kit.Split(m.Option("arg"))); m.Result() == "" {
					m.Table()
				}
			}
		}},
		"/input": {Name: "/input", Help: "补全", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Render(ice.RENDER_RESULT)
			arg[0] = strings.TrimSpace(arg[0])

			if strings.HasPrefix(arg[0], "ice ") {
				list := kit.Split(strings.TrimSpace(arg[0]))
				switch list[1] {
				case "add":
					// ice add person 想你 shwq
					m.Cmd("web.code.input.push", list[2:])
					arg[0] = list[4]
				default:
					// ice command
					if m.Cmdy(list[1:]); m.Result() == "" {
						m.Echo("%s\n", arg[0])
						m.Table()
					}
					return
				}
			}

			// 词汇列表
			m.Cmd("web.code.input.find", arg[0]).Table(func(index int, value map[string]string, head []string) {
				m.Echo("%s\n", value["text"])
			})
		}},
		"/favor": {Name: "/favor", Help: "收藏", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Render(ice.RENDER_RESULT)
			if m.Options("arg") {
				// 添加收藏
				m.Cmdy(web.FAVOR, kit.Select(m.Conf("vim", "meta.history"), m.Option("tab")),
					web.TYPE_VIMRC, m.Option("note"), m.Option("arg"),
					"pwd", m.Option("pwd"), "buf", m.Option("buf"), "row", m.Option("row"), "col", m.Option("col"))
				return
			}

			// 查看收藏
			m.Richs(web.FAVOR, nil, m.Option("tab"), func(key string, val map[string]interface{}) {
				m.Grows(web.FAVOR, kit.Keys(kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
					extra := value["extra"].(map[string]interface{})
					switch value[kit.MDB_TYPE] {
					case web.TYPE_VIMRC:
						m.Echo("%v\n", m.Option("tab")).Echo("%v:%v:%v:(%v): %v\n",
							extra["buf"], extra["row"], extra["col"], value["name"], value["text"])
					}
				})
			})
		}},
	},
	Configs: map[string]*ice.Config{
		VIM: {Name: "vim", Help: "编辑器", Value: kit.Data(
			"source", "ftp://ftp.vim.org/pub/vim/unix/vim-8.1.tar.bz2",
			"remote", "https://raw.githubusercontent.com/shylinux/contexts/master/etc/conf/",
			"build", []interface{}{
				"--enable-multibyte=yes",
				"--enable-pythoninterp=yes",
				"--enable-luainterp=yes",
				"--enable-cscope=yes",
			},
			"start", []interface{}{
				".vimrc", "vimrc",
				".vim/autoload/plug.vim", "plug.vim",
				".vim/syntax/javascript.vim", "javascript.vim",
				".vim/syntax/shy.vim", "shy.vim",
				".vim/syntax/go.vim", "go.vim",
			},

			"history", "vim.history",

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
