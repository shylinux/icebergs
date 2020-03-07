package docker

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/core/code"
	"github.com/shylinux/toolkits"
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

var Index = &ice.Context{Name: "docker", Help: "虚拟机",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"docker": {Name: "docker", Help: "虚拟机", Value: kit.Data(kit.MDB_SHORT, "name")},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			// if m.Richs(ice.WEB_FAVOR, nil, "alpine.init", nil) == nil {
			// 	m.Cmd(ice.WEB_FAVOR, "alpine.init", ice.TYPE_SHELL, "镜像源", `sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories`)
			// 	m.Cmd(ice.WEB_FAVOR, "alpine.init", ice.TYPE_SHELL, "软件包", `apk add bash`)
			// 	m.Cmd(ice.WEB_FAVOR, "alpine.init", ice.TYPE_SHELL, "软件包", `apk add curl`)
			// }
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},

		"image": {Name: "image", Help: "镜像管理", Meta: kit.Dict("detail", []string{"运行", "清理", "删除"}), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			prefix := []string{ice.CLI_SYSTEM, "docker", "image"}
			if len(arg) > 2 {
				switch arg[1] {
				case "运行":
					m.Cmdy(prefix[:2], "run", "-dt", m.Option("REPOSITORY")+":"+m.Option("TAG")).Set("append")
					return
				case "清理":
					m.Cmdy(prefix, "prune", "-f").Set("append")
					return
				case "delete":
					m.Cmdy(prefix, "rm", m.Option("IMAGE_ID")).Set("append")
					return
				}
			}

			if len(arg) > 0 {
				// 下载镜像
				m.Cmdy(prefix, "pull", arg[0]+":"+kit.Select("latest", arg, 1)).Set("append")
				return
			}
			// 镜像列表
			m.Split(strings.Replace(m.Cmdx(prefix, "ls"), "IMAGE ID", "IMAGE_ID", 1), "index", " ", "\n")
			m.Sort("REPOSITORY")
		}},
		"container": {Name: "container", Help: "容器管理", Meta: kit.Dict(
			"exports", []string{"CONTAINER_ID", "CONTAINER_ID"},
			"detail", []string{"进入", "启动", "停止", "重启", "清理", "编辑", "删除"}), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			prefix := []string{ice.CLI_SYSTEM, "docker", "container"}
			if len(arg) > 2 {
				switch arg[1] {
				case "进入":
					m.Cmd("cli.tmux.session").Table(func(index int, value map[string]string, head []string) {
						if value["tag"] == "1" {
							m.Cmdy(ice.CLI_SYSTEM, "tmux", "new-window", "-t", value["session"], "-n", m.Option("NAMES"),
								"-PF", "#{session_name}:#{window_name}.1", "docker exec -it "+m.Option("NAMES")+" sh").Set("append")
						}
					})
					return
				case "停止":
					m.Cmd(prefix, "stop", m.Option("CONTAINER_ID"))
				case "启动":
					m.Cmd(prefix, "start", m.Option("CONTAINER_ID"))
				case "重启":
					m.Cmd(prefix, "restart", m.Option("CONTAINER_ID"))
				case "清理":
					m.Cmd(prefix, "prune", "-f")
				case "modify":
					switch arg[2] {
					case "NAMES":
						m.Cmd(prefix, "rename", arg[4], arg[3])
					}
				case "delete":
					m.Cmdy(prefix, "rm", m.Option("CONTAINER_ID")).Set("append")
					return
				}
			}

			// 容器列表
			m.Split(strings.Replace(m.Cmdx(prefix, "ls", "-a"), "CONTAINER ID", "CONTAINER_ID", 1), "index", " ", "\n")
			m.Sort("NAMES")
		}},
		"command": {Name: "command", Help: "命令", List: kit.List(
			kit.MDB_INPUT, "text", "name", "CONTAINER_ID", "imports", "CONTAINER_ID",
			kit.MDB_INPUT, "text", "name", "command",
			kit.MDB_INPUT, "button", "name", "执行",
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			prefix := []string{ice.CLI_SYSTEM, "docker", "container"}
			m.Cmdy(prefix, "exec", arg[0], arg[1:]).Set("append")
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

func init() { code.Index.Register(Index, nil) }
