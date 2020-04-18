package vim

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/code"
	kit "github.com/shylinux/toolkits"

	"io/ioutil"
	"os"
	"path"
	"strings"
)

var Index = &ice.Context{Name: "vim", Help: "编辑器",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"vim": {Name: "vim", Help: "编辑器", Value: kit.Data(
			"source", "ftp://ftp.vim.org/pub/vim/unix/vim-8.1.tar.bz2",
			"target", "usr/local", "version", "vim81", "config", []interface{}{
				"--enable-pythoninterp=yes",
				"--enable-luainterp=yes",
				"--enable-cscope=yes",
			}, "history", "vim.history",
		)},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Conf(ice.WEB_FAVOR, "meta.render.vimrc", m.AddCmd(&ice.Command{Name: "render favor id", Help: "渲染引擎", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				value := m.Optionv("value").(map[string]interface{})
				switch value["name"] {
				case "read":
					p := path.Join(kit.Format(kit.Value(value, "extra.pwd")), kit.Format(kit.Value(value, "extra.buf")))
					f, e := os.Open(p)
					m.Assert(e)
					defer f.Close()
					b, e := ioutil.ReadAll(f)
					m.Assert(e)
					m.Echo(string(b))
				case "write":
				case "exec":
				}
			}}))
		}},
		ice.CODE_INSTALL: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			p := path.Join(m.Conf("install", "meta.path"), m.Conf("vim", "meta.version"))
			if _, e := os.Stat(p); e != nil {
				// 下载源码
				m.Option("cmd_dir", m.Conf("install", "meta.path"))
				m.Cmd(ice.CLI_SYSTEM, "wget", "-O", "vim.tar.gz", m.Conf("vim", "meta.source"))
				m.Cmd(ice.CLI_SYSTEM, "tar", "xvf", "vim.tar.gz")
			}

			// 配置选项
			m.Option("cmd_dir", p)
			m.Cmdy(ice.CLI_SYSTEM, "./configure", "--prefix="+kit.Path(m.Conf("vim", "meta.target")),
				"--enable-multibyte=yes", m.Confv("vim", "meta.config"))

			// 编译安装
			m.Cmdy(ice.CLI_SYSTEM, "make", "-j4")
			m.Cmdy(ice.CLI_SYSTEM, "make", "install")
		}},
		ice.CODE_PREPARE: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
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
		ice.CODE_PROJECT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		}},

		ice.WEB_LOGIN: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if f, _, e := m.R.FormFile("sub"); e == nil {
				defer f.Close()
				// 文件参数
				if b, e := ioutil.ReadAll(f); e == nil {
					m.Option("sub", string(b))
				}
			}

			m.Richs("login", nil, m.Option("sid"), func(key string, value map[string]interface{}) {
				// 查找空间
				m.Option("you", value["you"])
			})

			m.Logs(ice.LOG_AUTH, "you", m.Option("you"), "url", m.Option(ice.MSG_USERURL), "cmd", m.Optionv("cmds"), "sub", m.Optionv("sub"))
			m.Option(ice.MSG_OUTPUT, ice.RENDER_RESULT)
		}},
		"/help": {Name: "/help", Help: "帮助", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmdy("help")
		}},
		"/login": {Name: "/login", Help: "登录", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmdy("login", "init", c.Name)
		}},
		"/logout": {Name: "/logout", Help: "登出", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmdy("login", "exit")
		}},

		"/sync": {Name: "/sync", Help: "同步", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch arg[0] {
			case "read", "write", "exec", "insert":
				m.Cmd(ice.WEB_FAVOR, m.Conf("vim", "meta.history"), ice.TYPE_VIMRC, arg[0], kit.Select(m.Option("arg"), m.Option("sub")),
					"sid", m.Option("sid"), "pwd", m.Option("pwd"), "buf", m.Option("buf"), "row", m.Option("row"), "col", m.Option("col"))

			default:
				m.Richs("login", nil, m.Option("sid"), func(key string, value map[string]interface{}) {
					kit.Value(value, kit.Keys("sync", arg[0]), kit.Dict(
						"time", m.Time(), "text", m.Option("sub"),
						"pwd", m.Option("pwd"), "buf", m.Option("buf"),
					))
				})
			}
		}},
		"/input": {Name: "/input", Help: "补全", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if strings.HasPrefix(strings.TrimSpace(arg[0]), "ice ") {
				list := kit.Split(strings.TrimSpace(arg[0]))
				switch list[1] {
				case "add":
					// 添加词汇
					m.Cmd("web.code.input.push", list[2:])
					arg[0] = list[4]
				default:
					// 执行命令
					if m.Cmdy(list[1:]); m.Result() == "" {
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
			if m.Options("arg") {
				// 添加收藏
				m.Cmdy(ice.WEB_FAVOR, kit.Select(m.Conf("vim", "meta.history"), m.Option("tab")),
					ice.TYPE_VIMRC, m.Option("note"), m.Option("arg"),
					"pwd", m.Option("pwd"), "buf", m.Option("buf"), "row", m.Option("row"), "col", m.Option("col"))
				return
			}

			// 查看收藏
			m.Cmd(ice.WEB_PROXY, m.Option("you"), ice.WEB_FAVOR, m.Option("tab"), "extra", "extra.pwd", "extra.buf", "extra.row", "extra.col").Table(func(index int, value map[string]string, head []string) {
				switch value["type"] {
				case ice.TYPE_VIMRC:
					m.Echo("%v\n", m.Option("tab")).Echo("%v:%v:%v:(%v): %v\n",
						value["extra.buf"], value["extra.row"], value["extra.col"], value["name"], value["text"])
				}
			})
		}},
	},
}

func init() { code.Index.Register(Index, &web.Frame{}) }
