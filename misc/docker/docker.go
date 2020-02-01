package docker

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/core/code"
	"github.com/shylinux/toolkits"
	"strings"
)

var Index = &ice.Context{Name: "docker", Help: "容器管理",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"docker": {Name: "docker", Help: "docker", Value: kit.Data(kit.MDB_SHORT, "name")},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if m.Richs(ice.WEB_FAVOR, nil, "alpine.init", nil) == nil {
				m.Cmd(ice.WEB_FAVOR, "alpine.init", ice.TYPE_SHELL, "镜像源", `sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories`)
				m.Cmd(ice.WEB_FAVOR, "alpine.init", ice.TYPE_SHELL, "软件包", `apk add bash`)
				m.Cmd(ice.WEB_FAVOR, "alpine.init", ice.TYPE_SHELL, "软件包", `apk add curl`)
			}
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
	},
}

func init() { code.Index.Register(Index, nil) }
