package code

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/toolkits"

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
		"login": {Name: "login", Help: "登录", List: ice.ListLook("key"), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 0 && arg[0] == "action" {
				switch arg[1] {
				case "modify":
					m.Richs("login", nil, m.Option("key"), func(key string, value map[string]interface{}) {
						m.Log(ice.LOG_MODIFY, "%s %s %v->%s", key, arg[2], value[arg[2]], arg[3])
						value[arg[2]] = arg[3]
					})
				case "delete":
					m.Log(ice.LOG_DELETE, "%s %s", m.Option("key"), m.Conf("login", kit.Keys("hash", m.Option("key"))))
					m.Conf("login", kit.Keys("hash", m.Option("key")), "")
				}
				return
			}

			switch kit.Select("list", arg, 0) {
			case "open":
			case "init":
				if m.Option("sid") != "" && m.Confs("login", []string{"hash", m.Option("sid"), "status"}) {
					// 复用会话
					m.Conf("login", []string{"hash", m.Option("sid"), "status"}, "login")
					m.Log(ice.LOG_LOGIN, "sid: %s", m.Option("sid"))
					m.Echo(m.Option("sid"))
					return
				}

				you := m.Conf(ice.WEB_SHARE, kit.Keys("hash", m.Option("share"), "name"))
				// 添加会话
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
				m.Log(ice.LOG_LOGIN, "sid: %s you: %s", h, you)
				m.Echo(h)

			case "list":
				// 会话列表
				m.Richs("login", nil, "*", func(key string, value map[string]interface{}) {
					m.Push(key, value, []string{"time", "key", "type", "status", "you"})
					pwd := strings.Split(kit.Format(value["pwd"]), "/")
					if len(pwd) > 3 {
						m.Push("pwd", strings.Join(pwd[len(pwd)-3:len(pwd)], "/"))
					} else {
						m.Push("pwd", value["pwd"])
					}
					m.Push(key, value, []string{"pid", "pane", "hostname", "username"})
				})

			case "prune":
				list := []string{}
				m.Richs("login", nil, "*", func(key string, value map[string]interface{}) {
					if value["status"] == "logout" {
						list = append(list, key)
					}
				})

				kit.Fetch(list, func(index int, value string) {
					m.Log(ice.LOG_DELETE, "%s: %s", value, m.Conf("login", kit.Keys("hash", value)))
					m.Conf("login", kit.Keys("hash", value), "")
				})

			case "exit":
				// 退出会话
				m.Richs("login", nil, m.Option("sid"), func(key string, value map[string]interface{}) {
					m.Log(ice.LOG_LOGOUT, "sid: %s", m.Option("sid"))
					value["status"] = "logout"
				})
			default:
				// 会话详情
				m.Richs("login", nil, arg[0], func(key string, value map[string]interface{}) {
					m.Push("detail", value)
				})
			}
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
			os.MkdirAll(path.Dir(target), 0777)
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
	},
}

func init() { web.Index.Register(Index, &web.Frame{}) }
