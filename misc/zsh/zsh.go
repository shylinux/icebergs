package zsh

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/code"
	kit "github.com/shylinux/toolkits"

	"io/ioutil"
	"net/url"
	"os"
	"path"
	"strings"
	"unicode"
)

var Index = &ice.Context{Name: "zsh", Help: "命令行",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"zsh": {Name: "zsh", Help: "命令行", Value: kit.Data(
			"proxy", "tmux", "history", "zsh.history", "script", []interface{}{
				".vim/syntax/sh.vim", "etc/conf/sh.vim",
				".bashrc", "etc/conf/bashrc",
				".zshrc", "etc/conf/zshrc",
			},
		)},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Conf(ice.WEB_FAVOR, "meta.render.shell", m.AddCmd(&ice.Command{Name: "render type name text", Help: "渲染引擎", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				value, _ := m.Optionv(kit.MDB_VALUE).(map[string]interface{})
				m.Option("cmd_dir", kit.Value(value, "extra.pwd"))
				m.Cmdy(ice.CLI_SYSTEM, kit.Select(kit.Format(value["text"]), arg, 2))
			}}))
			m.Conf(ice.WEB_FAVOR, "meta.render.cmd", m.AddCmd(&ice.Command{Name: "render type name text", Help: "渲染引擎", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				value, _ := m.Optionv(kit.MDB_VALUE).(map[string]interface{})
				m.Cmdy(kit.Split(kit.Format(kit.Select(kit.Format(value["text"], arg, 2)))))
			}}))
			m.Conf(ice.WEB_FAVOR, "meta.render.bin", m.AddCmd(&ice.Command{Name: "render type name text", Help: "渲染引擎", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmdy(ice.CLI_SYSTEM, "file", arg[2])
			}}))
		}},
		ice.CODE_PREPARE: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd("web.code.git.repos", "shell", "usr/shell")
			p := path.Join(os.Getenv("HOME"), ".ish")
			m.Cmd(ice.CLI_SYSTEM, "rm", p)
			m.Cmd(ice.CLI_SYSTEM, "ln", "-s", kit.Path("usr/shell"), p)

			list := kit.Simple(m.Confv("zsh", "meta.script"))
			for i := 0; i < len(list); i += 2 {
				m.Cmd("nfs.link", path.Join(os.Getenv("HOME"), list[i]), list[i+1])
			}
		}},
		ice.WEB_LOGIN: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if f, _, e := m.R.FormFile("sub"); e == nil {
				defer f.Close()
				// 文件参数
				if b, e := ioutil.ReadAll(f); e == nil {
					m.Option("sub", string(b))
				}
			}

			m.Option("you", m.Conf("zsh", "meta.proxy"))
			m.Richs("login", nil, m.Option("sid"), func(key string, value map[string]interface{}) {
				// 查找空间
				m.Option("you", kit.Select(m.Conf("zsh", "meta.proxy"), value["you"]))
			})

			m.Option("arg", strings.ReplaceAll(m.Option("arg"), "%20", " "))
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
			case "history":
				vs := strings.SplitN(strings.TrimSpace(m.Option("arg")), " ", 4)
				if strings.Contains(m.Option("SHELL"), "zsh") {
					vs = []string{vs[0], m.Time("2006-01-02"), m.Time("15:04:05"), strings.Join(vs[1:], " ")}
				}
				m.Cmd(ice.WEB_FAVOR, m.Conf("zsh", "meta.history"), ice.TYPE_SHELL, vs[0], kit.Select("", vs, 3),
					"sid", m.Option("sid"), "pwd", m.Option("pwd"), "time", vs[1]+" "+vs[2])

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
			if len(arg) > 0 && arg[0] != "sh" {
				// 添加收藏
				m.Cmdy(ice.WEB_FAVOR, kit.Select(m.Conf("zsh", "meta.history"), m.Option("tab")),
					kit.Select(ice.TYPE_SHELL, m.Option("type")), m.Option("note"), arg[0])
				return
			}

			if m.Option("tab") == "" {
				// 收藏列表
				m.Cmdy(ice.WEB_FAVOR).Table()
				return
			}

			m.Echo("#/bin/sh\n\n")
			m.Cmd(ice.WEB_PROXY, m.Option("you"), ice.WEB_FAVOR, m.Option("tab")).Table(func(index int, value map[string]string, head []string) {
				switch value["type"] {
				case ice.TYPE_SHELL:
					// 查看收藏
					if m.Option("note") == "" || m.Option("note") == value["name"] {
						m.Echo("# %v\n%v\n\n", value["name"], value["text"])
					}
				}
			})
		}},
		"/ish": {Name: "/ish", Help: "命令", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if sub, e := url.QueryUnescape(m.Option("sub")); m.Assert(e) {
				m.Cmdy(kit.Split(sub))
				if len(m.Resultv()) == 0 {
					m.Table()
				}
			}
		}},

		"/download": {Name: "/download", Help: "下载", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 || arg[0] == "" {
				// 文件列表
				m.Cmdy(ice.WEB_PROXY, m.Option("you"), ice.WEB_STORY).Table()
				return
			}

			// 查找文件
			if m.Cmdy(ice.WEB_STORY, "index", arg[0]).Append("text") == "" && m.Option("you") != "" {
				// 上发文件
				m.Cmd(ice.WEB_PROXY, m.Option("you"), ice.WEB_STORY, "push", arg[0], "dev", arg[0])
				m.Cmdy(ice.WEB_STORY, "index", arg[0])
			}

			// 下载文件
			m.Render(kit.Select(ice.RENDER_DOWNLOAD, ice.RENDER_RESULT, m.Append("file") == ""), m.Append("text"))
		}},
		"/upload": {Name: "/upload", Help: "上传", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			// 缓存文件
			msg := m.Cmd(ice.WEB_STORY, "upload")
			m.Echo("data: %s\n", msg.Append("data"))
			m.Echo("time: %s\n", msg.Append("time"))
			m.Echo("type: %s\n", msg.Append("type"))
			m.Echo("name: %s\n", msg.Append("name"))
			m.Echo("size: %s\n", msg.Append("size"))
		}},
	},
}

func init() { code.Index.Register(Index, &web.Frame{}) }
