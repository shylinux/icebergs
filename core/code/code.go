package code

import (
	"github.com/shylinux/icebergs"
	_ "github.com/shylinux/icebergs/base"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/toolkits"

	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

var Dockfile = `
FROM {{options . "base"}}

WORKDIR /home/{{options . "user"}}/context
Env ctx_dev {{options . "host"}}

RUN wget -q -O - $ctx_dev/publish/boot.sh | sh -s install

CMD sh bin/boot.sh

`

var Index = &ice.Context{Name: "code", Help: "编程中心",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"login": {Name: "login", Help: "登录", Value: kit.Data()},

		"compile": {Name: "compile", Help: "编译", Value: kit.Data("path", "usr/publish")},
		"publish": {Name: "publish", Help: "发布", Value: kit.Data("path", "usr/publish")},
		"upgrade": {Name: "upgrade", Help: "升级", Value: kit.Dict(kit.MDB_HASH, kit.Dict(
			"system", kit.Dict(kit.MDB_LIST, kit.List(
				kit.MDB_INPUT, "bin", "file", "ice.sh", "path", "bin/ice.sh",
				kit.MDB_INPUT, "bin", "file", "ice.bin", "path", "bin/ice.bin",
			)),
		))},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Load()

			m.Watch(ice.SYSTEM_INIT, "compile", "linux")
			m.Watch(ice.SYSTEM_INIT, "publish", "bin/ice.sh")

			if m.Richs(ice.WEB_FAVOR, nil, "auto.init", nil) == nil {
				m.Cmd(ice.WEB_FAVOR, "auto.init", ice.TYPE_SHELL, "下载脚本", `curl -s "$ctx_dev/publish/auto.sh" -o auto.sh`)
				m.Cmd(ice.WEB_FAVOR, "auto.init", ice.TYPE_SHELL, "加载脚本", `source auto.sh`)
			}
			if m.Richs(ice.WEB_FAVOR, nil, "ice.init", nil) == nil {
				m.Cmd(ice.WEB_FAVOR, "ice.init", ice.TYPE_SHELL, "一键启动", `curl -s "$ctx_dev/publish/ice.sh" |sh`)
			}
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Save("login")
		}},

		"compile": {Name: "compile", Help: "编译", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
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
			m.Add("option", "cmd_env", "GOCACHE", os.Getenv("GOCACHE"))
			m.Add("option", "cmd_env", "GOARCH", arch, "GOOS", goos)
			m.Add("option", "cmd_env", "HOME", os.Getenv("HOME"))
			m.Add("option", "cmd_env", "CGO_ENABLED", "0")
			m.Cmd("cli.system", "go", "build", "-o", file, main)

			// 编译记录
			m.Cmdy(ice.WEB_STORY, "catch", "bin", file)
			m.Log(ice.LOG_EXPORT, "%s: %s", main, file)
		}},
		"publish": {Name: "publish", Help: "发布", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				// 目录列表
				m.Cmdy("nfs.dir", "", m.Conf("publish", "meta.path"), "time size path")
				return
			}

			p := arg[0]
			if s, e := os.Stat(arg[0]); m.Assert(e) && s.IsDir() {
				// 发布目录
				p = path.Base(arg[0]) + ".tar.gz"
				m.Cmd("cli.system", "tar", "-zcf", p, arg[0])
				defer func() { os.Remove(p) }()
				arg[0] = p
			}

			// 发布文件
			target := path.Join(m.Conf("publish", "meta.path"), path.Base(arg[0]))
			os.Remove(target)
			os.Link(arg[0], target)

			// 发布记录
			m.Cmdy(ice.WEB_STORY, "catch", "bin", p)
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

				os.Rename(kit.Format(value["path"]), kit.Keys(value["path"], "bak"))
				os.Link(m.Cmd(ice.WEB_STORY, "index", h).Append("file"), kit.Format(value["path"]))
				os.Chmod(kit.Format(value["path"]), 777)
				m.Log(ice.LOG_EXPORT, "%s: %s", h, value["path"])
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

		"_tmux": {Name: "tmux [session [window [pane cmd]]]", Help: "窗口", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			prefix := kit.Simple(m.Confv("prefix", "tmux"))
			if len(arg) > 1 {
				switch arg[1] {
				case "cmd":

				case "favor":
					env := m.Cmdx(prefix, "show-environment", "-g") + m.Cmdx(prefix, "show-environment", "-t", arg[0])
					for _, l := range strings.Split(env, "\n") {
						if strings.HasPrefix(l, "ctx_") {
							v := strings.SplitN(l, "=", 2)
							m.Option(v[0], v[1])
						}
					}
					m.Option("ctx_dev", m.Option("ctx_self"))

					m.Confm("tmux", "favor."+kit.Select("index", arg, 4), func(index int, value string) {
						if index == 0 {
							keys := strings.Split(value, " ")
							value = "export"
							for _, k := range keys {
								value += " " + k + "=" + m.Option(k)
							}

						}
						m.Cmdy(prefix, "send-keys", "-t", arg[0], value, "Enter")
						time.Sleep(100 * time.Millisecond)
					})
					m.Echo(strings.TrimSpace(m.Cmdx(prefix, "capture-pane", "-pt", arg[0])))
					return
				}
			}
			return
		}},
		"_docker": {Name: "docker image|volume|network|container", Help: "容器", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			prefix := kit.Simple(m.Confv("prefix", "docker"))
			switch arg[0] {
			case "image":
				if prefix = append(prefix, "image"); len(arg) < 3 {
					m.Cmdy(prefix, "ls", "cmd_parse", "cut", "cmd_headers", "IMAGE ID", "IMAGE_ID")
					break
				}

				switch arg[2] {
				case "运行":
					m.Cmdy(prefix[:2], "run", "-dt", m.Option("REPOSITORY")+":"+m.Option("TAG"))
				case "清理":
					m.Cmdy(prefix, "prune", "-f")
				case "delete":
					m.Cmdy(prefix, "rm", m.Option("IMAGE_ID"))
				case "创建":
					m.Option("base", m.Option("REPOSITORY")+":"+m.Option("TAG"))
					app := m.Conf("runtime", "boot.ctx_app")
					m.Option("name", app+":"+m.Time("20060102"))
					m.Option("file", m.Conf("docker", "output"))
					m.Option("user", m.Conf("runtime", "boot.username"))
					m.Option("host", "http://"+m.Conf("runtime", "boot.hostname")+".local"+m.Conf("runtime", "boot.web_port"))

					if f, _, e := kit.Create(m.Option("file")); m.Assert(e) {
						defer f.Close()
						// if m.Assert(ctx.ExecuteStr(m, f, m.Conf("docker", "template."+app))) {
						// 	m.Cmdy(prefix, "build", "-f", m.Option("file"), "-t", m.Option("name"), ".")
						// }
					}

				default:
					if len(arg) == 3 {
						m.Cmdy(prefix, "pull", arg[1]+":"+arg[2])
						break
					}
				}

			case "volume":
				if prefix = append(prefix, "volume"); len(arg) == 1 {
					m.Cmdy(prefix, "ls", "cmd_parse", "cut", "cmd_headers", "VOLUME NAME", "VOLUME_NAME")
					break
				}

			case "network":
				if prefix = append(prefix, "network"); len(arg) == 1 {
					m.Cmdy(prefix, "ls", "cmd_parse", "cut", "cmd_headers", "NETWORK ID", "NETWORK_ID")
					break
				}

				kit.Fetch(kit.Value(kit.UnMarshal(m.Cmdx(prefix, "inspect", arg[1])), "0.Containers"), func(key string, value map[string]interface{}) {
					m.Push("CONTAINER_ID", key[:12])
					m.Push("name", value["Name"])
					m.Push("IPv4", value["IPv4Address"])
					m.Push("IPv6", value["IPV4Address"])
					m.Push("Mac", value["MacAddress"])
				})
				m.Table()

			case "container":
				if prefix = append(prefix, "container"); len(arg) > 1 {
					switch arg[2] {
					case "进入":
						m.Cmdy(m.Confv("prefix", "tmux"), "new-window", "-t", "", "-n", m.Option("CONTAINER_NAME"),
							"-PF", "#{session_name}:#{window_name}.1", "docker exec -it "+arg[1]+" sh")
						return

					case "停止":
						m.Cmd(prefix, "stop", arg[1])

					case "启动":
						m.Cmd(prefix, "start", arg[1])

					case "重启":
						m.Cmd(prefix, "restart", arg[1])

					case "清理":
						m.Cmd(prefix, "prune", "-f")

					case "modify":
						switch arg[3] {
						case "NAMES":
							m.Cmd(prefix, "rename", arg[1], arg[4:])
						}

					case "delete":
						m.Cmdy(prefix, "rm", arg[1])

					default:
						if len(arg) == 2 {
							m.Cmdy(prefix, "inspect", arg[1])
							return
						}
						m.Cmdy(prefix, "exec", arg[1], arg[2:])
						return
					}
				}
				m.Cmdy(prefix, "ls", "-a", "cmd_parse", "cut", "cmd_headers", "CONTAINER ID", "CONTAINER_ID")

			case "command":
				switch arg[3] {
				case "base":
					m.Echo("\n0[%s]$ %s %s\n", time.Now().Format("15:04:05"), arg[2], m.Conf("package", arg[2]+".update"))
					m.Cmdy(prefix, "exec", arg[1], arg[2], strings.Split(m.Conf("package", arg[2]+".update"), " "))
					m.Confm("package", []string{arg[2], arg[3]}, func(index int, value string) {
						m.Echo("\n%d[%s]$ %s %s %s\n", index+1, time.Now().Format("15:04:05"), arg[2], m.Conf("package", arg[2]+".install"), value)
						m.Cmdy(prefix, "exec", arg[1], arg[2], strings.Split(m.Conf("package", arg[2]+".install"), " "), value)
					})
				}

			default:
				m.Cmdy(prefix, arg)
			}
			return
		}},
		"_git": {Name: "git init|diff|status|commit|branch|remote|pull|push|sum", Help: "版本", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			prefix, arg := append(kit.Simple(m.Confv("prefix", "git")), "cmd_dir", kit.Select(".", arg[0])), arg[1:]

			switch arg[0] {
			case "init":
				if _, e := os.Stat(path.Join(prefix[len(prefix)-1], ".git")); e != nil {
					m.Cmdy(prefix, "init")
				}
				if len(arg) > 1 {
					m.Cmdy(prefix, "remote", "add", "-f", kit.Select("origin", arg, 2), arg[1])
					m.Cmdy(prefix, "pull", kit.Select("origin", arg, 2), kit.Select("master", arg, 3))
				}

				m.Confm("git", "alias", func(key string, value string) {
					m.Cmdy(prefix, "config", "alias."+key, value)
				})

			case "diff":
				m.Cmdy(prefix, "diff")
			case "status":
				m.Cmdy(prefix, "status", "-sb", "cmd_parse", "cut", " ", "2", "tags file")
			case "commit":
				if len(arg) > 1 && m.Cmdy(prefix, "commit", "-am", arg[1]).Result() == "" {
					break
				}
				m.Cmdy(prefix, "log", "--stat", "-n", "3")
			case "branch":
				if len(arg) > 1 {
					m.Cmd(prefix, "branch", arg[1])
					m.Cmd(prefix, "checkout", arg[1])
				}
				for _, v := range strings.Split(m.Cmdx(prefix, "branch", "-v"), "\n") {
					if len(v) > 0 {
						m.Push("tags", v[:2])
						vs := strings.SplitN(strings.TrimSpace(v[2:]), " ", 2)
						m.Push("branch", vs[0])
						vs = strings.SplitN(strings.TrimSpace(vs[1]), " ", 2)
						m.Push("hash", vs[0])
						m.Push("note", strings.TrimSpace(vs[1]))
					}
				}
				m.Table()
			case "remote":
				m.Cmdy(prefix, "remote", "-v", "cmd_parse", "cut", " ", "3", "remote url tag")

			case "push":
				m.Cmdy(prefix, "push")
			case "sum":
				total := false
				if len(arg) > 1 && arg[1] == "total" {
					total, arg = true, arg[1:]
				}

				args := []string{"log", "--shortstat", "--pretty=commit: %ad %n%s", "--date=iso", "--reverse"}
				if len(arg) > 1 {
					args = append(args, kit.Select("-n", "--since", strings.Contains(arg[1], "-")))
					if strings.Contains(arg[1], "-") && !strings.Contains(arg[1], ":") {
						arg[1] = arg[1] + " 00:00:00"
					}
					args = append(args, arg[1:]...)
				} else {
					args = append(args, "-n", "30")
				}

				var total_day time.Duration
				count, count_add, count_del := 0, 0, 0
				if out, e := exec.Command("git", args...).CombinedOutput(); e == nil {
					for i, v := range strings.Split(string(out), "commit: ") {
						if i > 0 {
							l := strings.Split(v, "\n")
							hs := strings.Split(l[0], " ")

							add, del := "0", "0"
							if len(l) > 3 {
								fs := strings.Split(strings.TrimSpace(l[3]), ", ")
								if adds := strings.Split(fs[1], " "); len(fs) > 2 {
									dels := strings.Split(fs[2], " ")
									add = adds[0]
									del = dels[0]
								} else if adds[1] == "insertions(+)" {
									add = adds[0]
								} else {
									del = adds[0]
								}
							}

							if total {
								if count++; i == 1 {
									if t, e := time.Parse(ice.ICE_DATE, hs[0]); e == nil {
										total_day = time.Now().Sub(t)
										m.Append("from", hs[0])
									}
								}
								count_add += kit.Int(add)
								count_del += kit.Int(del)
								continue
							}

							m.Push("date", hs[0])
							m.Push("adds", add)
							m.Push("dels", del)
							m.Push("rest", kit.Int(add)-kit.Int(del))
							m.Push("note", l[1])
							m.Push("hour", strings.Split(hs[1], ":")[0])
							m.Push("time", hs[1])
						}
					}
					if total {
						m.Append("days", int(total_day.Hours())/24)
						m.Append("commit", count)
						m.Append("adds", count_add)
						m.Append("dels", count_del)
						m.Append("rest", count_add-count_del)
					}
					m.Table()
				} else {
					m.Log("warn", "%v", string(out))
				}

			default:
				m.Cmdy(prefix, arg)
			}
			return
		}},
	},
}

func init() { web.Index.Register(Index, &web.Frame{}) }
