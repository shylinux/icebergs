package vim

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/code"
	"github.com/shylinux/toolkits"

	"io/ioutil"
	"strings"
)

var Index = &ice.Context{Name: "vim", Help: "编辑器",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"vim": {Name: "vim", Help: "编辑器", Value: kit.Data(kit.MDB_SHORT, "name", "history", "vim.history")},
	},
	Commands: map[string]*ice.Command{
		ice.WEB_LOGIN: {Name: "_login", Help: "_login", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if f, _, e := m.R.FormFile("sub"); e == nil {
				defer f.Close()
				if b, e := ioutil.ReadAll(f); e == nil {
					// 加载参数
					m.Option("sub", string(b))
				}
			}

			m.Option("you", "")
			m.Richs("login", nil, m.Option("sid"), func(key string, value map[string]interface{}) {
				// 查找空间
				m.Option("you", value["you"])
			})

			m.Info("%s %s cmd: %v sub: %v", m.Option("you"), m.Option(ice.MSG_USERURL), m.Optionv("cmds"), m.Optionv("sub"))
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
				cmds := []string{ice.WEB_FAVOR, m.Conf("vim", "meta.history"), "vimrc", arg[0], kit.Select(m.Option("arg"), m.Option("sub")),
					"sid", m.Option("sid"), "pwd", m.Option("pwd"), "buf", m.Option("buf"), "row", m.Option("row"), "col", m.Option("col")}
				if m.Cmd(cmds); m.Option("you") != "" {
					m.Cmd(ice.WEB_SPACE, m.Option("you"), cmds)
				}
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
					m.Set("append")
					if m.Cmdy(arg); m.Result() == "" {
						m.Table()
					}
					return
				}
			}

			// 词汇列表
			m.Cmd("web.code.input.find", arg[0]).Table(func(index int, value map[string]string, head []string) {
				m.Echo("%s\n", value["text"])
			})
			m.Info("input %v", m.Result())
		}},
		"/favor": {Name: "/favor", Help: "收藏", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if m.Options("arg") {
				// 添加收藏
				m.Cmdy(ice.WEB_FAVOR, m.Option("tab"), "vimrc", m.Option("note"), m.Option("arg"),
					"pwd", m.Option("pwd"), "buf", m.Option("buf"), "row", m.Option("row"), "col", m.Option("col"))
				return
			}

			// 查看收藏
			m.Cmd(ice.WEB_SPACE, m.Option("you"), ice.WEB_FAVOR, m.Option("tab"), "extra", "extra.pwd", "extra.buf", "extra.row", "extra.col").Table(func(index int, value map[string]string, head []string) {
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
