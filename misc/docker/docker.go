package docker

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/core/code"
	"github.com/shylinux/toolkits"

	"strings"
)

var Index = &ice.Context{Name: "docker", Help: "虚拟机",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"docker": {Name: "docker", Help: "虚拟机", Value: kit.Data(kit.MDB_SHORT, "name", "build", []interface{}{})},
	},
	Commands: map[string]*ice.Command{
		"init": {Name: "init", Help: "初始化", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Watch(ice.DREAM_START, m.Prefix("auto"))

			if m.Richs(ice.WEB_FAVOR, nil, "alpine.auto", nil) == nil {
				m.Cmd(ice.WEB_FAVOR, "alpine.auto", ice.TYPE_SHELL, "镜像源", `sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories`)
				m.Cmd(ice.WEB_FAVOR, "alpine.auto", ice.TYPE_SHELL, "软件包", `apk add bash`)
				m.Cmd(ice.WEB_FAVOR, "alpine.auto", ice.TYPE_SHELL, "软件包", `apk add curl`)
			}
		}},
		"auto": {Name: "auto", Help: "自动化", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			prefix := []string{ice.CLI_SYSTEM, "docker"}

			if m.Cmd(prefix, "container", "start", arg[0]).Append("code") != "1" {
				// 启动容器
				return
			}

			args := []string{}
			kit.Fetch(m.Confv("docker", "meta.build"), func(index int, value string) {
				switch value {
				case "home":
					args = append(args, "-w", "/root")
				case "mount":
					args = append(args, "--mount", kit.Format("type=bind,source=%s,target=/root", kit.Path(m.Conf(ice.WEB_DREAM, "meta.path"), arg[0])))
				}
			})

			// 创建容器
			pid := m.Cmdx(prefix, "run", "-dt", args, "--name", arg[0], "alpine")
			m.Log(ice.LOG_CREATE, "%s: %s", arg[0], pid)

			m.Cmd(ice.WEB_FAVOR, kit.Select("alpine.auto", arg, 1)).Table(func(index int, value map[string]string, head []string) {
				if value["type"] == ice.TYPE_SHELL {
					// 执行命令
					m.Cmd(prefix, "exec", arg[0], kit.Split(value["text"]))
				}
			})
		}},

		"image": {Name: "image", Help: "镜像管理", Meta: kit.Dict("detail", []string{"运行", "清理", "删除"}), List: ice.ListLook("IMAGE_ID"), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			prefix := []string{ice.CLI_SYSTEM, "docker", "image"}
			if len(arg) > 1 && arg[0] == "action" {
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
				// 容器详情
				res := m.Cmdx(prefix, "inspect", arg[0])
				m.Push("detail", kit.KeyValue(map[string]interface{}{}, "", kit.Parse(nil, "", kit.Split(res)...)))
				return
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
		"container": {Name: "container", Help: "容器管理", List: ice.ListLook("CONTAINER_ID"), Meta: kit.Dict("detail", []string{"进入", "启动", "停止", "重启", "清理", "编辑", "删除"}), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			prefix := []string{ice.CLI_SYSTEM, "docker", "container"}
			if len(arg) > 1 && arg[0] == "action" {
				switch arg[1] {
				case "进入":
					m.Cmdy(ice.CLI_SYSTEM, "tmux", "new-window", "-t", m.Option("NAMES"), "-n", m.Option("NAMES"),
						"-PF", "#{session_name}:#{window_name}.1", "docker exec -it "+m.Option("NAMES")+" bash").Set("append")
					return
				case "停止":
					m.Cmdy(prefix, "stop", m.Option("CONTAINER_ID"))
				case "启动":
					m.Cmdy(prefix, "start", m.Option("CONTAINER_ID"))
				case "重启":
					m.Cmdy(prefix, "restart", m.Option("CONTAINER_ID"))
				case "清理":
					m.Cmdy(prefix, "prune", "-f")
				case "modify":
					switch arg[2] {
					case "NAMES":
						m.Cmdy(prefix, "rename", arg[4], arg[3])
					}
				case "delete":
					m.Cmdy(prefix, "rm", m.Option("CONTAINER_ID")).Set("append")
				}
				return
			}

			if len(arg) > 0 {
				// 容器详情
				res := m.Cmdx(prefix, "inspect", arg[0])
				m.Push("detail", kit.KeyValue(map[string]interface{}{}, "", kit.Parse(nil, "", kit.Split(res)...)))
				return
			}

			// 容器列表
			m.Split(strings.Replace(m.Cmdx(prefix, "ls", "-a"), "CONTAINER ID", "CONTAINER_ID", 1), "index", " ", "\n")
			m.Sort("NAMES")
		}},
		"command": {Name: "command", Help: "命令", List: kit.List(
			kit.MDB_INPUT, "text", "name", "CONTAINER_ID",
			kit.MDB_INPUT, "text", "name", "cmd", "className", "args cmd",
			kit.MDB_INPUT, "button", "value", "执行",
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) < 2 {
				m.Cmdy("container")
				return
			}
			prefix := []string{ice.CLI_SYSTEM, "docker", "container"}
			m.Cmdy(prefix, "exec", arg[0], arg[1:]).Set("append")
		}},
	},
}

func init() { code.Index.Register(Index, nil) }
