package code

import (
	"github.com/shylinux/icebergs"
	_ "github.com/shylinux/icebergs/base"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/toolkits"

	"io/ioutil"
	"os"
	"path"
	"strings"
)

var Index = &ice.Context{Name: "code", Help: "编程中心",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"compile": {Name: "compile", Help: "编译", Value: kit.Data("path", "usr/publish")},
		"publish": {Name: "publish", Help: "发布", Value: kit.Data("path", "usr/publish")},
		"upgrade": {Name: "upgrade", Help: "升级", Value: kit.Dict(kit.MDB_HASH, kit.Dict(
			"system", kit.Dict(kit.MDB_LIST, kit.List(
				kit.MDB_INPUT, "bin", "file", "ice.bin", "path", "bin/ice.bin",
				kit.MDB_INPUT, "bin", "file", "ice.sh", "path", "bin/ice.sh",
			)),
		))},

		"login": {Name: "login", Help: "登录", Value: kit.Data()},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Load()

			// m.Watch(ice.SYSTEM_INIT, "compile", "linux")
			// m.Watch(ice.SYSTEM_INIT, "publish", "bin/ice.sh")
			//
			// if m.Richs(ice.WEB_FAVOR, nil, "auto.init", nil) == nil {
			// 	m.Cmd(ice.WEB_FAVOR, "auto.init", ice.TYPE_SHELL, "下载脚本", `curl -s "$ctx_dev/publish/auto.sh" -o auto.sh`)
			// 	m.Cmd(ice.WEB_FAVOR, "auto.init", ice.TYPE_SHELL, "加载脚本", `source auto.sh`)
			// }
			// if m.Richs(ice.WEB_FAVOR, nil, "ice.init", nil) == nil {
			// 	m.Cmd(ice.WEB_FAVOR, "ice.init", ice.TYPE_SHELL, "一键启动", `curl -s "$ctx_dev/publish/ice.sh" |sh`)
			// }
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Save("login")
		}},

		"compile": {Name: "compile [os [arch]]", Help: "编译", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				// 目录列表
				m.Cmdy("nfs.dir", m.Conf("publish", "meta.path"), "time size path")
				return
			}

			// 编译目标
			main := kit.Select("src/main.go", arg, 2)
			arch := kit.Select(m.Conf(ice.CLI_RUNTIME, "host.GOARCH"), arg, 1)
			goos := kit.Select(m.Conf(ice.CLI_RUNTIME, "host.GOOS"), arg, 0)
			file := path.Join(m.Conf("compile", "meta.path"), kit.Keys("ice", goos, arch))

			// 编译参数
			m.Optionv("cmd_env", "GOCACHE", os.Getenv("GOCACHE"), "HOME", os.Getenv("HOME"),
				"GOARCH", arch, "GOOS", goos, "CGO_ENABLED", "0")
			m.Cmd(ice.CLI_SYSTEM, "go", "build", "-o", file, main)

			// 编译记录
			m.Cmdy(ice.WEB_STORY, ice.STORY_CATCH, "bin", file)
			m.Log(ice.LOG_EXPORT, "%s: %s", main, file)
		}},
		"publish": {Name: "publish", Help: "发布", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				// 目录列表
				m.Cmdy("nfs.dir", m.Conf("publish", "meta.path"), "time size path")
				return
			}

			p := arg[0]
			if s, e := os.Stat(arg[0]); m.Assert(e) && s.IsDir() {
				// 发布目录
				p = path.Base(arg[0]) + ".tar.gz"
				m.Cmd(ice.CLI_SYSTEM, "tar", "-zcf", p, arg[0])
				defer func() { os.Remove(p) }()
				arg[0] = p
			}

			// 发布文件
			target := path.Join(m.Conf("publish", "meta.path"), path.Base(arg[0]))
			os.Remove(target)
			os.Link(arg[0], target)

			// 发布记录
			m.Cmdy(ice.WEB_STORY, ice.STORY_CATCH, "bin", p)
			m.Log(ice.LOG_EXPORT, "%s: %s", arg[0], target)
		}},
		"upgrade": {Name: "upgrade", Help: "升级", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			exit := true
			m.Grows("upgrade", "hash.system", "", "", func(index int, value map[string]interface{}) {
				if value["file"] == "ice.bin" {
					value["file"] = kit.Keys("ice", m.Conf(ice.CLI_RUNTIME, "host.GOOS"), m.Conf(ice.CLI_RUNTIME, "host.GOARCH"))
				}

				h := m.Cmdx(ice.WEB_SPIDE, "dev", "cache", "GET", "/publish/"+kit.Format(value["file"]))
				if h == "" {
					exit = false
					return
				}

				m.Cmd(ice.WEB_STORY, "add", "bin", value["path"], h)
				m.Cmd(ice.WEB_STORY, ice.STORY_WATCH, h, value["path"])
				os.Chmod(kit.Format(value["path"]), 777)
			})

			if exit {
				m.Cmd("exit")
			}
		}},

		"login": {Name: "login", Help: "登录", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch kit.Select("list", arg, 0) {
			case "open":
			case "init":
				if m.Option("sid") != "" {
					if m.Confs("login", []string{"hash", m.Option("sid"), "status"}) {
						m.Conf("login", []string{"hash", m.Option("sid"), "status"}, "login")
						m.Echo(m.Option("sid"))
						return
					}
				}

				you := m.Conf(ice.WEB_SHARE, kit.Keys("hash", m.Option("share"), "name"))
				// 添加终端
				h := m.Rich("login", nil, kit.Dict(
					"status", "login",
					"type", kit.Select("zsh", arg, 1),
					"you", you,
					"pwd", m.Option("pwd"),
					"pid", m.Option("pid"),
					"pane", m.Option("pane"),
					"hostname", m.Option("hostname"),
					"username", m.Option("username"),
				))
				m.Info("%s: %s", you, h)
				m.Echo(h)

			case "list":
				m.Richs("login", nil, "*", func(key string, value map[string]interface{}) {
					m.Push(key, value, []string{"time", "type", "status", "you"})
					pwd := strings.Split(kit.Format(value["pwd"]), "/")
					if len(pwd) > 3 {
						m.Push("pwd", strings.Join(pwd[len(pwd)-3:len(pwd)], "/"))
					} else {
						m.Push("pwd", value["pwd"])
					}

					m.Push(key, value, []string{"pid", "pane", "hostname", "username"})
				})

			case "exit":
				m.Richs("login", nil, m.Option("sid"), func(key string, value map[string]interface{}) {
					m.Info("logout: %s", m.Option("sid"))
					value["status"] = "logout"
				})
			}
		}},
		"/zsh": {Name: "/zsh", Help: "命令行", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if f, _, e := m.R.FormFile("sub"); e == nil {
				defer f.Close()
				if b, e := ioutil.ReadAll(f); e == nil {
					m.Option("sub", string(b))
				}
			}

			m.Option("you", "")
			m.Richs("login", nil, m.Option("sid"), func(key string, value map[string]interface{}) {
				m.Option("you", value["you"])
			})
			m.Info("%s%s %s arg: %v sub: %v", m.Option("you"), cmd, m.Option("cmd"), m.Optionv("arg"), m.Optionv("sub"))

			m.Push("_output", "result")
			switch m.Option("cmd") {
			case "login":
				m.Cmdy("login", "init", cmd)
			case "logout":
				m.Cmdy("login", "exit")
			case "upload":
				// 缓存文件
				you := m.Option("you")
				m.Option("you", "")
				msg := m.Cmd(ice.WEB_STORY, "upload")
				m.Echo("data: %s\n", msg.Append("data"))
				m.Echo("time: %s\n", msg.Append("time"))
				m.Echo("type: %s\n", msg.Append("type"))
				m.Echo("name: %s\n", msg.Append("name"))
				m.Echo("size: %s\n", msg.Append("size"))
				m.Push("_output", "result")

				// 下发文件
				m.Option("you", you)
				m.Cmd(ice.WEB_SPACE, msg.Option("you"), ice.WEB_STORY, ice.STORY_PULL, "dev", msg.Append("name"))

			case "download":
				// 下载文件
				m.Option("you", "")
				if m.Cmdy(ice.WEB_STORY, "index", m.Option("arg")).Append("text") == "" {
					m.Cmdy(ice.WEB_SPACE, m.Option("pod"), ice.WEB_STORY, "index", m.Optionv("arg"))
				}
				m.Append("_output", kit.Select("file", "result", m.Append("file") == ""))

			case "history":
				vs := strings.SplitN(strings.TrimSpace(m.Option("arg")), " ", 2)
				m.Cmd(ice.WEB_SPACE, m.Option("you"), ice.WEB_FAVOR, "zsh.history", "shell", m.Option("sid"), kit.Select("", vs, 1),
					"sid", m.Option("sid"), "num", vs[0], "pwd", m.Option("pwd"))
				m.Push("_output", "void")

			case "favor":
				if m.Options("arg") {
					m.Cmdy(ice.WEB_SPACE, m.Option("you"), ice.WEB_FAVOR,
						m.Option("tab"), ice.TYPE_SHELL, m.Option("note"), m.Option("arg"))
					break
				}
				m.Echo("#/bin/sh\n\n")
				m.Cmd(ice.WEB_SPACE, m.Option("you"), ice.WEB_FAVOR, m.Option("tab")).Table(func(index int, value map[string]string, head []string) {
					switch value["type"] {
					case ice.TYPE_SHELL:
						m.Echo("# %v:%v\n%v\n\n", value["type"], value["name"], value["text"])
					}
				})
				m.Push("_output", "result")
			}
		}},
		"/vim": {Name: "/vim", Help: "编辑器", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if f, _, e := m.R.FormFile("sub"); e == nil {
				defer f.Close()
				if b, e := ioutil.ReadAll(f); e == nil {
					m.Option("sub", string(b))
				}
			}

			m.Option("you", "")
			m.Richs("login", nil, m.Option("sid"), func(key string, value map[string]interface{}) {
				m.Option("you", value["you"])
			})
			m.Info("%s%s %s arg: %v sub: %v", m.Option("you"), cmd, m.Option("cmd"), m.Optionv("arg"), m.Optionv("sub"))

			m.Push("_output", "result")
			switch m.Option("cmd") {
			case "login":
				m.Cmdy("login", "init", cmd)
			case "logout":
				m.Cmdy("login", "exit")

			case "read", "write", "exec":
				m.Cmd(ice.WEB_FAVOR, "vim.history", "vimrc", m.Option("cmd"), m.Option("arg"),
					"sid", m.Option("sid"), "pwd", m.Option("pwd"), "buf", m.Option("buf"))

			case "tasklet":
				m.Cmd(ice.APP_MISS, m.Option("arg"), m.Option("sub"))

			case "trans":
				if strings.HasPrefix(strings.TrimSpace(m.Option("arg")), "ice ") {
					arg := kit.Split(strings.TrimPrefix(strings.TrimSpace(m.Option("arg")), "ice "))
					switch arg[0] {
					case "add":
						// 添加词汇
						m.Cmd("input.push", arg[1:])
						m.Option("arg", arg[2])
					default:
						// 执行命令
						m.Set("append")
						if m.Cmdy(arg).Table(); strings.TrimSpace(m.Result()) == "" {
							m.Cmdy(ice.CLI_SYSTEM, arg)
						}
						m.Push("_output", "result")
						return
					}
				}
				// 词汇列表
				m.Cmd("input.find", m.Option("arg")).Table(func(index int, value map[string]string, head []string) {
					m.Echo("%s\n", value["text"])
					m.Push("_output", "result")
				})

			case "favor":
				if m.Options("arg") {
					m.Cmd(ice.WEB_FAVOR, m.Option("tab"), "vimrc", m.Option("note"), m.Option("arg"),
						"buf", m.Option("buf"), "line", m.Option("line"), "col", m.Option("col"),
					)
					break
				}
				m.Cmd(ice.WEB_FAVOR, m.Option("tab"), "extra", "buf line col").Table(func(index int, value map[string]string, head []string) {
					switch value["type"] {
					case ice.TYPE_VIMRC:
						m.Echo("%v\n", m.Option("tab")).Echo("%v:%v:%v:(%v): %v\n",
							value["buf"], value["line"], value["col"], value["name"], value["text"])
					}
				})
			}
		}},
	},
}

func init() { web.Index.Register(Index, &web.Frame{}) }
