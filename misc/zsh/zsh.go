package zsh

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/code"
	"github.com/shylinux/toolkits"

	"io/ioutil"
	"strings"
	"unicode"
)

var Index = &ice.Context{Name: "zsh", Help: "命令行",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"zsh": {Name: "zsh", Help: "命令行", Value: kit.Data(kit.MDB_SHORT, "name")},
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
			m.Append("_output", "result")
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

		"/download": {Name: "/download", Help: "下载", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			you := m.Option("you")
			m.Option("you", "")

			if len(arg) == 0 || arg[0] == "" {
				// 文件列表
				m.Cmdy(ice.WEB_SPACE, you, ice.WEB_STORY)
				m.Table()
				return
			}

			// 查找文件
			if m.Cmdy(ice.WEB_STORY, "index", arg[0]).Append("text") == "" && you != "" {
				// 上发文件
				m.Cmdy(ice.WEB_SPACE, you, ice.WEB_STORY, "index", arg[0])
			}

			// 下载文件
			m.Append("_output", kit.Select("file", "result", m.Append("file") == ""))
		}},
		"/upload": {Name: "/upload", Help: "上传", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			you := m.Option("you")
			m.Option("you", "")

			// 缓存文件
			msg := m.Cmd(ice.WEB_STORY, "upload")
			m.Echo("data: %s\n", msg.Append("data"))
			m.Echo("time: %s\n", msg.Append("time"))
			m.Echo("type: %s\n", msg.Append("type"))
			m.Echo("name: %s\n", msg.Append("name"))
			m.Echo("size: %s\n", msg.Append("size"))

			if you != "" {
				// 下发文件
				m.Cmd(ice.WEB_SPACE, you, ice.WEB_STORY, ice.STORY_PULL, "dev", msg.Append("name"))
			}
		}},
		"/sync": {Name: "/sync", Help: "同步", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch arg[0] {
			default:
				m.Richs("login", nil, m.Option("sid"), func(key string, value map[string]interface{}) {
					kit.Value(value, kit.Keys("sync", arg[0]), kit.Dict(
						"time", m.Time(), "text", m.Option("sub"),
						"pwd", m.Option("pwd"), "cmd", arg[1:],
					))
				})
			}
		}},
		"/input": {Name: "/input", Help: "补全", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			list := kit.Split(m.Option("line"), m.Option("break"))
			word := list[kit.Int(m.Option("index"))]
			switch arg[0] {
			case "shy":
				m.Cmd("web.code.input.find", word).Table(func(index int, value map[string]string, head []string) {
					m.Echo(value["text"]).Echo(" ")
				})

			case "line":
				if strings.HasPrefix(m.Option("line"), "ice ") {
					list := kit.Split(m.Option("line"))
					switch list[1] {
					case "add":
						m.Cmd("web.code.input.push", list[2:])
						m.Option("line", list[4])
						m.Option("point", 0)
					default:
						m.Cmdy(list[1:])
						break
					}
				}

				line := []rune(m.Option("line"))
				if begin := kit.Int(m.Option("point")); begin < len(line) {
					m.Richs("login", nil, m.Option("sid"), func(key string, value map[string]interface{}) {
						m.Echo(string(line[:begin]))
						for i := begin; i < len(line); i++ {
							if i-begin < 3 && i < len(line)-1 {
								continue
							}
							// 编码转换
							for j := 0; j < 4; j++ {
								code := string(line[begin : i+1-j])
								list := append(m.Cmd("web.code.input.find", code).Appendv("text"), code)
								if len(list) > 1 {
									m.Echo(kit.Select(code, list[0]))
									m.Info("input %s->%s", code, list[0])
									i = i - j
									break
								}
							}
							// 输出编码
							begin = i + 1
						}
					})
					break
				}
				fallthrough
			case "end":
				m.Richs("login", nil, m.Option("sid"), func(key string, value map[string]interface{}) {
					last_text := kit.Format(kit.Value(value, "last.text"))
					last_list := kit.Simple(kit.Value(value, "last.list"))
					last_index := kit.Int(kit.Value(value, "last.index"))

					if last_text != "" && strings.HasSuffix(m.Option("line"), last_text) {
						// 补全记录
						index := last_index + 1
						text := last_list[index%len(last_list)]
						kit.Value(value, "last.index", index)
						kit.Value(value, "last.text", text)
						m.Echo(strings.TrimSuffix(m.Option("line"), last_text) + text)
						m.Info("%d %v", index, last_list)
						return
					}

					line := []rune(m.Option("line"))
					for i := len(line); i >= 0; i-- {
						if i > 0 && len(line)-i < 4 && unicode.IsLower(line[i-1]) {
							continue
						}

						// 编码转换
						code := string(line[i:])
						list := append(m.Cmd("web.code.input.find", code).Appendv("text"), code)
						value["last"] = kit.Dict("code", code, "text", list[0], "list", list, "index", 0)

						// 输出编码
						m.Echo(string(line[:i]))
						m.Echo(kit.Select(code, list[0]))
						m.Info("input %s->%s", code, list[0])
						break
					}
				})
			}
			m.Info("trans: %v", m.Result())
		}},
		"/favor": {Name: "/favor", Help: "收藏", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			you := m.Option("you")
			m.Option("you", "")

			if len(arg) > 0 && arg[0] != "sh" {
				// 添加收藏
				if m.Cmdy(ice.WEB_FAVOR, m.Option("tab"), ice.TYPE_SHELL, m.Option("note"), arg[0]); you != "" {
					m.Cmdy(ice.WEB_SPACE, you, ice.WEB_FAVOR, m.Option("tab"), ice.TYPE_SHELL, m.Option("note"), arg[0])
				}
				return
			}

			m.Echo("#/bin/sh\n\n")
			m.Cmd(ice.WEB_SPACE, you, ice.WEB_FAVOR, m.Option("tab")).Table(func(index int, value map[string]string, head []string) {
				switch value["type"] {
				case ice.TYPE_SHELL:
					// 查看收藏
					if m.Option("note") == "" || m.Option("note") == value["name"] {
						m.Echo("# %v\n%v\n\n", value["name"], value["text"])
					}
				}
			})
		}},
		"/history": {Name: "/history", Help: "历史", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			vs := strings.SplitN(strings.TrimSpace(arg[0]), " ", 2)
			m.Cmd(ice.WEB_SPACE, m.Option("you"), ice.WEB_FAVOR, "zsh.history", ice.TYPE_SHELL, vs[0], kit.Select("", vs, 1),
				"sid", m.Option("sid"), "pwd", m.Option("pwd"))
		}},
	},
}

func init() { code.Index.Register(Index, &web.Frame{}) }
