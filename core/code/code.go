package code

import (
	"net/http"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"

	"net/url"
	"os"
	"path"
	"runtime"
	"strings"
)

const (
	VEDIO  = "vedio"
	QRCODE = "qrcode"
)

const ( // CODE
	// INSTALL = "_install"
	PREPARE = "_prepare"
	PROJECT = "_project"
)

var Index = &ice.Context{Name: "code", Help: "编程中心",
	Configs: map[string]*ice.Config{
		"_install": {Name: "install", Help: "安装", Value: kit.Data(
			"path", "usr/install", "target", "usr/local",
			"linux", "https://dl.google.com/go/go1.14.2.linux-amd64.tar.gz",
			"darwin", "https://dl.google.com/go/go1.14.6.darwin-amd64.tar.gz",
			"windows", "https://golang.google.cn/dl/go1.14.6.windows-amd64.zip",
		)},
		PREPARE: {Name: "prepare", Help: "配置", Value: kit.Data("path", "usr/prepare",
			"script", ".ish/pluged/golang/init.sh", "export", kit.Dict(
				"GOPROXY", "https://goproxy.cn,direct",
				"GOPRIVATE", "https://github.com",
			),
		)},
		PROJECT: {Name: "project", Help: "项目", Value: kit.Data("path", "usr/project")},

		"login": {Name: "login", Help: "终端接入", Value: kit.Data()},
	},
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Load()
			m.Cmd(mdb.ENGINE, mdb.CREATE, BENCH)
		}},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Save(INSTALL)
		}},

		"_install": {Name: "install url 安装:button", Help: "安装", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			target := m.Conf("_install", kit.Keys("meta", runtime.GOOS))

			p := path.Join(m.Conf("_install", "meta.path"), path.Base(target))
			if _, e := os.Stat(p); e != nil {
				// 下载
				msg := m.Cmd(web.SPIDE, "dev", web.CACHE, http.MethodGet, target)
				m.Cmd(web.CACHE, web.WATCH, msg.Append(web.DATA), p)
			}

			os.MkdirAll(m.Conf("_install", kit.Keys("meta.target")), ice.MOD_DIR)
			m.Cmdy(cli.SYSTEM, "tar", "xvf", p, "-C", m.Conf("_install", kit.Keys("meta.target")))
		}},
		PREPARE: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			export := []string{}
			kit.Fetch(m.Confv(PREPARE, "meta.export"), func(key string, val string) {
				export = append(export, key+"="+val)
			})

			m.Cmd(nfs.SAVE, m.Conf(PREPARE, "meta.script"), kit.Format(`
export GOROOT=%s GOPATH=%s:$GOPATH GOBIN=%s
export PATH=$GOBIN:$GOROOT/bin:$PATH
export %s
`, kit.Path(m.Conf("_install", kit.Keys("meta.target")), "go"), kit.Path("src"), kit.Path("bin"), strings.Join(export, " ")))
		}},
		PROJECT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		}},

		"login": {Name: "login key", Help: "登录", Meta: kit.Dict(
			"detail", []string{"编辑", "删除", "清理", "清空"},
		), Action: map[string]*ice.Action{}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
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

				you := m.Conf(web.SHARE, kit.Keys(kit.MDB_HASH, m.Option("share"), "name"))
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

		"/miss/": {Name: "/miss/", Help: "任务", Action: map[string]*ice.Action{
			"pwd": {Name: "pwd", Help: "pwd", Hand: func(m *ice.Message, arg ...string) {
				m.Render(ice.RENDER_RESULT)
				m.Echo("hello world\n")
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			u, e := url.QueryUnescape(m.Option("arg"))
			m.Assert(e)
			args := kit.Split(u)
			if len(arg) == 0 || arg[0] == "" {
				return
			}

			m.Render(ice.RENDER_RESULT)
			if m.Cmdy(arg, args); len(m.Resultv()) == 0 {
				m.Table()
			}
		}},
	},
}

func init() {
	web.Index.Register(Index, &web.Frame{},
		INSTALL,
		COMPILE,
		UPGRADE,
		PUBLISH,
		BENCH,
		PPROF,
	)
}
