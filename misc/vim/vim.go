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

var Index = &ice.Context{Name: "vim", Help: "编辑器",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		VIM: {Name: "vim", Help: "编辑器", Value: kit.Data(
			"source", "ftp://ftp.vim.org/pub/vim/unix/vim-8.1.tar.bz2",
			"target", "usr/local", "version", "vim81", "config", []interface{}{
				"--enable-pythoninterp=yes",
				"--enable-luainterp=yes",
				"--enable-cscope=yes",
			}, "history", "vim.history",
			"plug", kit.Dict(
				"prefix", kit.Dict(
					"\"", "comment",
				),
				"keyword", kit.Dict(
					"highlight", "keyword",
					"syntax", "keyword",
				),
			),
		)},
	},
	Commands: map[string]*ice.Command{
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
		}},
		VIM: {Name: VIM, Help: "vim", Action: map[string]*ice.Action{
			mdb.PLUGIN: {Hand: func(m *ice.Message, arg ...string) {
				m.Echo(m.Conf(VIM, "meta.plug"))
			}},
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(nfs.CAT, path.Join(arg[2], arg[1]))
			}},
		}},
		code.INSTALL: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			p := path.Join(m.Conf("install", "meta.path"), m.Conf("vim", "meta.version"))
			if _, e := os.Stat(p); e != nil {
				// 下载源码
				m.Option("cmd_dir", m.Conf("install", "meta.path"))
				m.Cmd(cli.SYSTEM, "wget", "-O", "vim.tar.gz", m.Conf("vim", "meta.source"))
				m.Cmd(cli.SYSTEM, "tar", "xvf", "vim.tar.gz")
			}

			// 配置选项
			m.Option("cmd_dir", p)
			m.Cmdy(cli.SYSTEM, "./configure", "--prefix="+kit.Path(m.Conf("vim", "meta.target")),
				"--enable-multibyte=yes", m.Confv("vim", "meta.config"))

			// 编译安装
			m.Cmdy(cli.SYSTEM, "make", "-j4")
			m.Cmdy(cli.SYSTEM, "make", "install")
		}},
		code.PREPARE: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			// 语法脚本
			for _, s := range []string{"go.vim", "shy.vim", "javascript.vim"} {
				m.Cmd("nfs.link", path.Join(os.Getenv("HOME"), ".vim/syntax/"+s), "etc/conf/"+s)
			}

			// 启动脚本
			m.Cmd("nfs.link", path.Join(os.Getenv("HOME"), ".vim/autoload/plug.vim"), "etc/conf/plug.vim")
			m.Cmd("nfs.link", path.Join(os.Getenv("HOME"), ".vimrc"), "etc/conf/vimrc")

			// 安装插件
			m.Echo("vim -c PlugInstall\n")
			m.Echo("vim -c GoInstallBinaries\n")
		}},
		code.PROJECT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
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
					// 添加词汇
					m.Cmd("web.code.input.push", list[2:])
					arg[0] = list[4]
				default:
					// ice add person 想你 shwq
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
}

func init() { code.Index.Register(Index, &web.Frame{}) }
