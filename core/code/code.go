package code

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"

	"os"
	"path"
	"runtime"
	"strings"

	"net/http"
	_ "net/http/pprof"
)

var Index = &ice.Context{Name: "code", Help: "编程中心",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"install": {Name: "install", Help: "安装", Value: kit.Data("path", "usr/install",
			"linux", "https://dl.google.com/go/go1.14.2.linux-amd64.tar.gz",
			"darwin", "https://dl.google.com/go/go1.14.2.darwin-amd64.pkg",
			"windows", "https://dl.google.com/go/go1.14.2.windows-amd64.msi",
			"source", "https://dl.google.com/go/go1.14.2.src.tar.gz",
			"target", "usr/local", "script", ".ish/pluged/golang/init.sh", "export", kit.Dict(
				"GOPROXY", "https://goproxy.cn,direct",
				"GOPRIVATE", "https://github.com",
			),
		)},
		"prepare": {Name: "prepare", Help: "配置", Value: kit.Data("path", "usr/prepare")},
		"project": {Name: "project", Help: "项目", Value: kit.Data("path", "usr/project")},

		"compile": {Name: "compile", Help: "编译", Value: kit.Data("path", "usr/publish")},
		"publish": {Name: "publish", Help: "发布", Value: kit.Data("path", "usr/publish")},
		"upgrade": {Name: "upgrade", Help: "升级", Value: kit.Dict(kit.MDB_HASH, kit.Dict(
			"system", kit.Dict(kit.MDB_LIST, kit.List(
				kit.MDB_INPUT, "bin", "file", "ice.bin", "path", "bin/ice.bin",
				kit.MDB_INPUT, "bin", "file", "ice.sh", "path", "bin/ice.sh",
			)),
		))},

		"login": {Name: "login", Help: "终端接入", Value: kit.Data()},
		"pprof": {Name: "pprof", Help: "性能分析", Value: kit.Data(kit.MDB_SHORT, kit.MDB_NAME,
			"stop", "ps aux|grep pprof|grep -v grep|cut -d' ' -f2|xargs -n1 kill",
		)},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Load()
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Save("login")
		}},

		ice.CODE_INSTALL: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			p := path.Join(m.Conf("install", "meta.path"), path.Base(m.Conf("install", kit.Keys("meta", runtime.GOOS))))
			// 下载
			if _, e := os.Stat(p); e != nil {
				m.Option("cmd_dir", m.Conf("install", "meta.path"))
				m.Cmd(ice.CLI_SYSTEM, "wget", m.Conf("install", kit.Keys("meta", runtime.GOOS)))
			}

			// 安装
			m.Option("cmd_dir", "")
			os.MkdirAll(m.Conf("install", kit.Keys("meta.target")), 0777)
			m.Cmdy(ice.CLI_SYSTEM, "tar", "xvf", p, "-C", m.Conf("install", kit.Keys("meta.target")))
		}},
		ice.CODE_PREPARE: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			export := []string{}
			kit.Fetch(m.Confv("install", "meta.export"), func(key string, val string) {
				export = append(export, key+"="+val)
			})

			m.Cmd("nfs.save", m.Conf("install", "meta.script"), kit.Format(`
export GOROOT=%s GOPATH=%s:$GOPATH GOBIN=%s
export PATH=$GOBIN:$GOROOT/bin:$PATH
export %s
`, kit.Path(m.Conf("install", kit.Keys("meta.target")), "go"), kit.Path("src"), kit.Path("bin"), strings.Join(export, " ")))
		}},
		ice.CODE_PROJECT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		}},

		"install": {Name: "install", Help: "安装", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		}},
		"prepare": {Name: "prepare", Help: "配置", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		}},
		"project": {Name: "project", Help: "项目", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		}},

		"compile": {Name: "compile [os [arch [main]]]", Help: "编译", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				// 目录列表
				m.Cmdy("nfs.dir", m.Conf(cmd, "meta.path"), "time size path")
				return
			}

			// 编译目标
			main := kit.Select("src/main.go", arg, 2)
			arch := kit.Select(m.Conf(ice.CLI_RUNTIME, "host.GOARCH"), arg, 1)
			goos := kit.Select(m.Conf(ice.CLI_RUNTIME, "host.GOOS"), arg, 0)
			file := path.Join(m.Conf(cmd, "meta.path"), kit.Keys("ice", goos, arch))

			// 编译参数
			m.Optionv("cmd_env", "GOCACHE", os.Getenv("GOCACHE"), "HOME", os.Getenv("HOME"),
				"GOARCH", arch, "GOOS", goos, "CGO_ENABLED", "0")
			m.Cmd(ice.CLI_SYSTEM, "go", "build", "-o", file, main)

			// 编译记录
			m.Cmdy(ice.WEB_STORY, ice.STORY_CATCH, "bin", file)
			m.Logs(ice.LOG_EXPORT, "source", main, "target", file)
		}},
		"publish": {Name: "publish [source]", Help: "发布", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				// 目录列表
				m.Cmdy("nfs.dir", m.Conf(cmd, "meta.path"), "time size path")
				return
			}

			if s, e := os.Stat(arg[0]); m.Assert(e) && s.IsDir() {
				// 发布目录
				p := path.Base(arg[0]) + ".tar.gz"
				m.Cmd(ice.CLI_SYSTEM, "tar", "-zcf", p, arg[0])
				defer func() { os.Remove(p) }()
				arg[0] = p
			}

			// 发布文件
			target := path.Join(m.Conf(cmd, "meta.path"), path.Base(arg[0]))
			os.Remove(target)
			os.MkdirAll(path.Dir(target), 0777)
			os.Link(arg[0], target)

			// 发布记录
			m.Cmdy(ice.WEB_STORY, ice.STORY_CATCH, "bin", target)
			m.Logs(ice.LOG_EXPORT, "source", arg[0], "target", target)
		}},
		"upgrade": {Name: "upgrade which", Help: "升级", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			exit := false
			m.Grows(cmd, kit.Keys(kit.MDB_HASH, kit.Select("system", arg, 0)), "", "", func(index int, value map[string]interface{}) {
				if value["file"] == "ice.bin" {
					// 程序文件
					value["file"] = kit.Keys("ice", m.Conf(ice.CLI_RUNTIME, "host.GOOS"), m.Conf(ice.CLI_RUNTIME, "host.GOARCH"))
					exit = true
				}

				// 下载文件
				h := m.Cmdx(ice.WEB_SPIDE, "dev", "cache", "GET", "/publish/"+kit.Format(value["file"]))
				if h == "" {
					exit = false
					return
				}

				// 升级记录
				m.Cmd(ice.WEB_STORY, "add", "bin", value["path"], h)
				m.Cmd(ice.WEB_STORY, ice.STORY_WATCH, h, value["path"])
				os.Chmod(kit.Format(value["path"]), 0777)
			})
			if exit {
				m.Sleep("1s").Gos(m, func(m *ice.Message) { m.Cmd("exit") })
			}
		}},

		"login": {Name: "login key", Help: "登录", Meta: kit.Dict(
			"detail", []string{"编辑", "删除", "清理", "清空"},
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 0 && arg[0] == "action" {
				switch arg[1] {
				case "modify", "编辑":
					m.Richs(cmd, nil, m.Option("key"), func(key string, value map[string]interface{}) {
						m.Logs(ice.LOG_MODIFY, cmd, key, "field", arg[2], "value", kit.Value(value, arg[2]), "->", arg[3])
						kit.Value(value, arg[2], arg[3])
					})

				case "delete", "删除":
					m.Logs(ice.LOG_DELETE, cmd, m.Option("key"), "value", m.Conf(cmd, kit.Keys(kit.MDB_HASH, m.Option("key"))))
					m.Conf(cmd, kit.Keys(kit.MDB_HASH, m.Option("key")), "")

				case "prune", "清理":
					m.Cmdy(cmd, "prune")

				case "clear", "清空":
					m.Cmdy(cmd, "prune", "all")
				}
				return
			}

			switch kit.Select("list", arg, 0) {
			case "init":
				if m.Option("sid") != "" && m.Conf(cmd, []string{kit.MDB_HASH, m.Option("sid"), "status"}) != "" {
					// 复用会话
					m.Conf(cmd, []string{kit.MDB_HASH, m.Option("sid"), "status"}, "login")
					m.Logs(ice.LOG_AUTH, "sid", m.Option("sid"))
					m.Echo(m.Option("sid"))
					return
				}

				you := m.Conf(ice.WEB_SHARE, kit.Keys(kit.MDB_HASH, m.Option("share"), "name"))
				// 添加会话
				h := m.Rich(cmd, nil, kit.Dict(
					"type", kit.Select("zsh", arg, 1),
					"status", "login",
					"you", you,
					"pwd", m.Option("pwd"),
					"pid", m.Option("pid"),
					"pane", m.Option("pane"),
					"hostname", m.Option("hostname"),
					"username", m.Option("username"),
				))
				m.Logs(ice.LOG_AUTH, "sid", h, "you", you)
				m.Echo(h)

			case "exit":
				// 退出会话
				m.Richs(cmd, nil, m.Option("sid"), func(key string, value map[string]interface{}) {
					m.Logs(ice.LOG_AUTH, "sid", m.Option("sid"))
					value["status"] = "logout"
					m.Echo(key)
				})

			case "prune":
				list := []string{}
				m.Richs(cmd, nil, "*", func(key string, value map[string]interface{}) {
					if len(arg) > 1 && arg[1] == "all" || value["status"] == "logout" {
						list = append(list, key)
					}
				})

				// 清理会话
				kit.Fetch(list, func(index int, value string) {
					m.Logs(ice.LOG_DELETE, "login", value, "value", m.Conf(cmd, kit.Keys(kit.MDB_HASH, value)))
					m.Conf(cmd, kit.Keys(kit.MDB_HASH, value), "")
				})
				m.Echo("%d", len(list))

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

			default:
				// 会话详情
				m.Richs(cmd, nil, arg[0], func(key string, value map[string]interface{}) {
					m.Push("detail", value)
				})
			}
		}},
		"pprof": {Name: "pprof run name time", Help: "性能分析", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if m.Show(cmd, arg...) {
				return
			}

			switch arg[0] {
			case "run":
				m.Richs(cmd, nil, arg[1], func(key string, value map[string]interface{}) {
					m.Gos(m.Spawn(), func(msg *ice.Message) {
						m.Sleep("1s").Grows(cmd, kit.Keys(kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
							// 压测命令
							m.Cmd(ice.WEB_FAVOR, "pprof", "shell", value[kit.MDB_TEXT], m.Cmdx(kit.Split(kit.Format(value[kit.MDB_TEXT]))))
						})
					})

					// 启动监控
					name := arg[1] + ".pd.gz"
					value = value["meta"].(map[string]interface{})
					msg := m.Cmd(ice.WEB_SPIDE, "self", "cache", "GET", kit.Select("/code/pprof/profile", value["remote"]), "seconds", kit.Select("5", arg, 2))
					m.Cmd(ice.WEB_FAVOR, "pprof", "shell", "text", m.Cmdx(ice.CLI_SYSTEM, "go", "tool", "pprof", "-text", msg.Append("text")))
					m.Cmd(ice.WEB_FAVOR, "pprof", "pprof", name, msg.Append("data"))

					arg = kit.Simple("web", value[kit.MDB_TEXT], msg.Append("text"))
				})

				fallthrough
			case "web":
				// 展示结果
				p := kit.Format("%s:%s", m.Conf(ice.WEB_SHARE, "meta.host"), m.Cmdx("tcp.getport"))
				m.Cmd(ice.CLI_DAEMON, "go", "tool", "pprof", "-http="+p, arg[1:])
				m.Cmd(ice.WEB_FAVOR, "pprof", "bin", arg[1], m.Cmd(ice.WEB_CACHE, "catch", "bin", arg[1]).Append("data"))
				m.Cmd(ice.WEB_FAVOR, "pprof", "spide", arg[2], "http://"+p)
				m.Echo(p)

			case "stop":
				m.Cmd(ice.CLI_SYSTEM, "sh", "-c", m.Conf(cmd, "meta.stop"))

			case "add":
				key := m.Rich(cmd, nil, kit.Data(
					kit.MDB_NAME, arg[1], kit.MDB_TEXT, arg[2], "remote", arg[3],
				))

				for i := 4; i < len(arg)-1; i += 2 {
					m.Grow(cmd, kit.Keys(kit.MDB_HASH, key), kit.Dict(
						kit.MDB_NAME, arg[i], kit.MDB_TEXT, arg[i+1],
					))
				}
			}
		}},
		"/pprof/": {Name: "/pprof/", Help: "性能分析", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.R.URL.Path = strings.Replace("/code"+m.R.URL.Path, "code", "debug", 1)
			http.DefaultServeMux.ServeHTTP(m.W, m.R)
			m.Render(ice.RENDER_VOID)
		}},
	},
}

func init() { web.Index.Register(Index, &web.Frame{}) }
