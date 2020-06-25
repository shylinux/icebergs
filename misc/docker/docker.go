package docker

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/gdb"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/code"
	kit "github.com/shylinux/toolkits"

	"strings"
)

const DOCKER = "docker"
const (
	IMAGE     = "image"
	CONTAINER = "container"
)

var Index = &ice.Context{Name: "docker", Help: "虚拟机",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		DOCKER: {Name: "docker", Help: "虚拟机", Value: kit.Data(
			kit.MDB_SHORT, "name", "build", []interface{}{}),
		},
	},
	Commands: map[string]*ice.Command{
		IMAGE: {Name: "image", Help: "镜像管理", Meta: kit.Dict("detail", []string{"运行", "清理", "删除"}), List: ListLook("IMAGE_ID"), Action: map[string]*ice.Action{
			"run": {Name: "run", Help: "运行", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(cli.SYSTEM, DOCKER, "run", "-dt", m.Option("REPOSITORY")+":"+m.Option("TAG"))
			}},
			"prune": {Name: "prune", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(cli.SYSTEM, DOCKER, "prune", "-f")
			}},
			mdb.DELETE: {Name: "delete", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(cli.SYSTEM, DOCKER, "rm", m.Option("IMAGE_ID"))
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			prefix := []string{cli.SYSTEM, DOCKER, IMAGE}
			if len(arg) > 0 {
				// 容器详情
				res := m.Cmdx(prefix, "inspect", arg[0])
				m.Push("detail", kit.KeyValue(map[string]interface{}{}, "", kit.Parse(nil, "", kit.Split(res)...)))
				return
			}

			// 镜像列表
			m.Split(strings.Replace(m.Cmdx(prefix, "ls"), "IMAGE ID", "IMAGE_ID", 1), "index", " ", "\n")
			m.Sort("REPOSITORY")
		}},
		CONTAINER: {Name: "container", Help: "容器管理", List: ListLook("CONTAINER_ID"), Meta: kit.Dict("detail", []string{"进入", "启动", "停止", "重启", "清理", "编辑", "删除"}), Action: map[string]*ice.Action{
			"prune": {Name: "prune", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(cli.SYSTEM, DOCKER, "prune", "-f")
				m.Cmdy(prefix, "prune", "-f")
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			prefix := []string{cli.SYSTEM, DOCKER, CONTAINER}
			if len(arg) > 1 && arg[0] == "action" {
				switch arg[1] {
				case "进入":
					m.Cmdy(cli.SYSTEM, "tmux", "new-window", "-t", m.Option("NAMES"), "-n", m.Option("NAMES"),
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

		"init": {Name: "init", Help: "初始化", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Watch(gdb.DREAM_START, m.Prefix("auto"))

			if m.Richs(web.FAVOR, nil, "alpine.auto", nil) == nil {
				m.Cmd(web.FAVOR, "alpine.auto", web.TYPE_SHELL, "镜像源", `sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories`)
				m.Cmd(web.FAVOR, "alpine.auto", web.TYPE_SHELL, "软件包", `apk add bash`)
				m.Cmd(web.FAVOR, "alpine.auto", web.TYPE_SHELL, "软件包", `apk add curl`)
			}
		}},
		"auto": {Name: "auto", Help: "自动化", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			prefix := []string{cli.SYSTEM, "docker"}

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
					args = append(args, "--mount", kit.Format("type=bind,source=%s,target=/root", kit.Path(m.Conf(web.DREAM, "meta.path"), arg[0])))
				}
			})

			// 创建容器
			pid := m.Cmdx(prefix, "run", "-dt", args, "--name", arg[0], "alpine")
			m.Log(ice.LOG_CREATE, "%s: %s", arg[0], pid)

			m.Cmd(web.FAVOR, kit.Select("alpine.auto", arg, 1)).Table(func(index int, value map[string]string, head []string) {
				if value["type"] == web.TYPE_SHELL {
					// 执行命令
					m.Cmd(prefix, "exec", arg[0], kit.Split(value["text"]))
				}
			})
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
			prefix := []string{cli.SYSTEM, "docker", "container"}
			m.Cmdy(prefix, "exec", arg[0], arg[1:]).Set("append")
		}},
	},
}

func init() { code.Index.Register(Index, nil) }

func ListLook(name ...string) []interface{} {
	list := []interface{}{}
	for _, k := range name {
		list = append(list, kit.MDB_INPUT, "text", "name", k, "action", "auto")
	}
	return kit.List(append(list,
		kit.MDB_INPUT, "button", "name", "查看", "action", "auto",
		kit.MDB_INPUT, "button", "name", "返回", "cb", "Last",
	)...)
}
