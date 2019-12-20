package code

import (
	"github.com/shylinux/toolkits"

	"github.com/shylinux/icebergs"
	_ "github.com/shylinux/icebergs/base"
	"github.com/shylinux/icebergs/base/web"

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

var Index = &ice.Context{Name: "code", Help: "编程模块",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"prefix": {Name: "prefix", Help: "外部命令", Value: map[string]interface{}{
			"zsh":    []interface{}{"cli.system", "zsh"},
			"tmux":   []interface{}{"cli.system", "tmux"},
			"docker": []interface{}{"cli.system", "docker"},
			"git":    []interface{}{"cli.system", "git"},
			"vim":    []interface{}{"cli.system", "vim"},
		}},
		"docker": {Name: "docker", Help: "容器", Value: map[string]interface{}{
			"template": map[string]interface{}{"shy": Dockfile},
			"output":   "etc/Dockerfile",
		}},
		"tmux": {Name: "tmux", Help: "终端", Value: map[string]interface{}{
			"favor": map[string]interface{}{
				"index": []interface{}{
					"ctx_dev ctx_share",
					"curl -s $ctx_dev/publish/auto.sh >auto.sh",
					"source auto.sh",
					"ShyLogin",
				},
			},
		}},
		"git": {Name: "git", Help: "记录", Value: map[string]interface{}{
			"alias": map[string]interface{}{"s": "status", "b": "branch"},
		}},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd(ice.GDB_EVENT, "listen", "miss", "start", "web.code.docker", "image")
		}},
		"tmux": {Name: "tmux [session [window [pane cmd]]]", Help: "窗口", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			prefix := kit.Simple(m.Confv("prefix", "tmux"))
			if len(arg) > 1 {
				switch arg[1] {
				case "cmd":

				case "pane":
					prefix = append(prefix, "list-panes")
					if arg[0] == "" {
						prefix = append(prefix, "-a")
					} else {
						prefix = append(prefix, "-s", "-t", arg[0])
					}
					m.Cmd(prefix, "cmd_parse", "cut", " ", "8", "pane_name size some lines bytes haha pane_id tag").Table(func(index int, value map[string]string) {
						m.Push("pane_id", strings.TrimPrefix(value["pane_id"], "%"))
						m.Push("pane_name", strings.TrimSuffix(value["pane_name"], ":"))
						m.Push("size", value["size"])
						m.Push("lines", strings.TrimSuffix(value["lines"], ","))
						m.Push("bytes", kit.FmtSize(kit.Int64(value["bytes"])))
						m.Push("tag", value["tag"])
					})

					m.Sort("pane_name")
					return

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

				case "buffer":
					// 写缓存
					if len(arg) > 5 {
						switch arg[3] {
						case "modify":
							switch arg[4] {
							case "text":
								m.Cmdy(prefix, "set-buffer", "-b", arg[2], arg[5])
							}
						}
					} else if len(arg) > 3 {
						m.Cmd(prefix, "set-buffer", "-b", arg[2], arg[3])
					}

					// 读缓存
					if len(arg) > 2 {
						m.Cmdy(prefix, "show-buffer", "-b", arg[2])
						return
					}

					m.Cmd(prefix, "list-buffers", "cmd_parse", "cut", ": ", "3", "buffer size text").Table(func(index int, value map[string]string, head []string) {
						m.Push("buffer", value["buffer"])
						m.Push("size", value["size"])
						if index < 3 {
							m.Push("text", m.Cmdx(prefix, "show-buffer", "-b", value["buffer"]))
						} else {
							m.Push("text", value["text"][2:len(value["text"])-1])
						}
					})
					return

				case "select":
					// 切换会话
					if m.Options("session") {
						m.Cmd(prefix, "switch-client", "-t", arg[0])
						arg = arg[:0]
						break
					}
					m.Cmd(prefix, "switch-client", "-t", m.Option("session"))

					// 切换窗口
					if !m.Options("window") {
						m.Cmd(prefix, "select-window", "-t", m.Option("session")+":"+arg[0])
						arg = []string{m.Option("session")}
						break
					}
					m.Cmd(prefix, "select-window", "-t", m.Option("session")+":"+m.Option("window"))

					// 切换终端
					m.Cmd(prefix, "select-pane", "-t", m.Option("session")+":"+m.Option("window")+"."+arg[0])
					arg = []string{m.Option("session"), m.Option("window")}

				case "modify":
					switch arg[2] {
					case "session":
						// 重命名会话
						m.Cmdy(prefix, "rename-session", "-t", arg[0], arg[3])
						arg = arg[:0]

					case "window":
						// 重命名窗口
						m.Cmdy(prefix, "rename-window", "-t", m.Option("session")+":"+arg[0], arg[3])
						arg = []string{m.Option("session")}

					default:
						return
					}
				case "delete":
					// 删除会话
					if !m.Options("session") {
						m.Cmdy(prefix, "kill-session", "-t", arg[0])
						arg = arg[:0]
						break
					}

					// 删除窗口
					if !m.Options("window") {
						m.Cmdy(prefix, "kill-window", "-t", m.Option("session")+":"+arg[0])
						arg = []string{m.Option("session")}
						break
					}

					// 删除终端
					m.Cmd(prefix, "kill-pane", "-t", m.Option("session")+":"+m.Option("window")+"."+arg[3])
					arg = []string{m.Option("session"), m.Option("window")}
				}
			}

			// 查看会话
			if x := m.Cmdx(prefix, "list-session", "-F", "#{session_id},#{session_attached},#{session_name},#{session_windows},#{session_height},#{session_width}"); len(arg) == 0 {
				for _, l := range kit.Split(x, "\n") {
					ls := kit.Split(l, ",")
					m.Push("id", ls[0])
					m.Push("tag", ls[1])
					m.Push("session", ls[2])
					m.Push("windows", ls[3])
					m.Push("height", ls[4])
					m.Push("width", ls[5])
				}
				return
			}

			// 创建会话
			if arg[0] != "" && kit.IndexOf(m.Appendv("session"), arg[0]) == -1 {
				m.Cmdy(prefix, "new-session", "-ds", arg[0])
			}
			m.Set(ice.MSG_APPEND).Set(ice.MSG_RESULT)

			// 查看窗口
			if m.Cmdy(prefix, "list-windows", "-t", arg[0], "-F", "#{window_id},#{window_active},#{window_name},#{window_panes},#{window_height},#{window_width}",
				"cmd_parse", "cut", ",", "6", "id tag window panes height width"); len(arg) == 1 {
				return
			}

			// 创建窗口
			if arg[1] != "" && kit.IndexOf(m.Appendv("window"), arg[1]) == -1 {
				m.Cmdy(prefix, "new-window", "-dt", arg[0], "-n", arg[1])
			}
			m.Set(ice.MSG_APPEND).Set(ice.MSG_RESULT)

			// 查看面板
			if len(arg) == 2 {
				m.Cmdy(prefix, "list-panes", "-t", arg[0]+":"+arg[1], "-F", "#{pane_id},#{pane_active},#{pane_index},#{pane_tty},#{pane_height},#{pane_width}",
					"cmd_parse", "cut", ",", "6", "id tag pane tty height width")
				return
			}

			// 执行命令
			target := arg[0] + ":" + arg[1] + "." + arg[2]
			if len(arg) > 3 {
				m.Cmdy(prefix, "send-keys", "-t", target, strings.Join(arg[3:], " "), "Enter")
				time.Sleep(1 * time.Second)
			}

			// 查看终端
			m.Echo(strings.TrimSpace(m.Cmdx(prefix, "capture-pane", "-pt", target)))
			return
		}},
		"docker": {Name: "docker image|volume|network|container", Help: "容器", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
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
		"git": {Name: "git init|diff|status|commit|branch|remote|pull|push|sum", Help: "版本", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
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
