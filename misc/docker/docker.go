package docker

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/gdb"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/code"
	kit "github.com/shylinux/toolkits"

	"os"
	"path"
	"strings"
)

const DOCKER = "docker"
const (
	IMAGE     = "image"
	CONTAINER = "container"
	COMMAND   = "command"
)

var _docker = []string{cli.SYSTEM, DOCKER}
var _image = []string{cli.SYSTEM, DOCKER, IMAGE}
var _container = []string{cli.SYSTEM, DOCKER, CONTAINER}

var Index = &ice.Context{Name: "docker", Help: "虚拟机",
	Configs: map[string]*ice.Config{
		DOCKER: {Name: "docker", Help: "虚拟机", Value: kit.Data(
			"repos", "centos", "build", []interface{}{"home", "mount"},
		)},
	},
	Commands: map[string]*ice.Command{
		IMAGE: {Name: "image IMAGE_ID=auto auto", Help: "镜像管理", Meta: kit.Dict(
			"detail", []string{"运行", "清理", "删除"},
		), Action: map[string]*ice.Action{
			"run": {Name: "run", Help: "运行", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(_docker, "run", "-dt", m.Option("REPOSITORY")+":"+m.Option("TAG"))
			}},
			gdb.PRUNE: {Name: "prune", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(_image, "prune", "-f")
			}},
			gdb.CLEAR: {Name: "clear", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(_image, "rm", m.Option("IMAGE_ID"))
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 0 {
				// 容器详情
				res := m.Cmdx(_image, "inspect", arg[0])
				m.Push("detail", kit.KeyValue(map[string]interface{}{}, "", kit.Parse(nil, "", kit.Split(res)...)))
				return
			}

			// 镜像列表
			m.Split(strings.Replace(m.Cmdx(_image, "ls"), "IMAGE ID", "IMAGE_ID", 1), "index", " ", "\n")
			m.Sort("REPOSITORY")
		}},
		CONTAINER: {Name: "container CONTAINER_ID=auto auto", Help: "容器管理", Meta: kit.Dict(
			"detail", []string{"进入", "启动", "停止", "重启", "清理", "编辑", "删除"},
		), Action: map[string]*ice.Action{
			gdb.OPEN: {Name: "open", Help: "进入", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(cli.SYSTEM, "tmux", "new-window", "-t", m.Option("NAMES"), "-n", m.Option("NAMES"),
					"-PF", "#{session_name}:#{window_name}.1", "docker exec -it "+m.Option("NAMES")+" bash")
			}},
			gdb.START: {Name: "start", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(_container, "start", m.Option("CONTAINER_ID"))
			}},
			gdb.STOP: {Name: "stop", Help: "停止", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(_container, "stop", m.Option("CONTAINER_ID"))
			}},
			gdb.RESTART: {Name: "restart", Help: "重启", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(_container, "restart", m.Option("CONTAINER_ID"))
			}},
			gdb.CHANGE: {Name: "change", Help: "更改", Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case "NAMES":
					m.Cmdy(_container, "rename", m.Option("CONTAINER_ID"), arg[1])
				}
			}},
			gdb.PRUNE: {Name: "prune", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(_container, "prune", "-f")
			}},
			gdb.CLEAR: {Name: "clear", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(_container, "rm", m.Option("CONTAINER_ID"))
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 0 {
				// 容器详情
				res := m.Cmdx(_container, "inspect", arg[0])
				m.Push("detail", kit.KeyValue(map[string]interface{}{}, "", kit.Parse(nil, "", kit.Split(res)...)))
				return
			}

			// 容器列表
			m.Split(strings.Replace(m.Cmdx(_container, "ls", "-a"), "CONTAINER ID", "CONTAINER_ID", 1), "index", " ", "\n")
			m.Sort("NAMES")

			m.Table(func(index int, value map[string]string, head []string) {
				for _, k := range []string{"open", "start", "stop", "restart", "clear"} {
					m.Push(k, m.Cmdx(mdb.RENDER, web.RENDER.Button, k))
				}
			})
		}},
		COMMAND: {Name: "command NAMES=auto cmd... auto", Help: "命令管理", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) < 2 {
				m.Cmdy(CONTAINER)
				return
			}
			m.Echo(m.Cmdx(_container, "exec", arg[0], kit.Split(kit.Select("pwd", arg, 1), " ", " ")))
		}},

		gdb.INIT: {Name: "init", Help: "初始化", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Watch(web.DREAM_START)
			return

			if m.Richs(web.FAVOR, nil, "alpine.auto", nil) == nil {
				m.Cmd(web.FAVOR, "alpine.auto", web.TYPE_SHELL, "镜像源", `sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories`)
				m.Cmd(web.FAVOR, "alpine.auto", web.TYPE_SHELL, "软件包", `apk add bash`)
				m.Cmd(web.FAVOR, "alpine.auto", web.TYPE_SHELL, "软件包", `apk add curl`)
			}
		}},
		gdb.AUTO: {Name: "auto", Help: "自动化", Action: map[string]*ice.Action{
			web.DREAM_START: {Name: "dream.start", Hand: func(m *ice.Message, arg ...string) {
				if m.Cmd(_container, "start", arg[0]).Append(cli.CMD_CODE) == "0" {
					// 启动容器
					return
				}

				args := []string{"--name", arg[0]}
				kit.Fetch(m.Confv(DOCKER, "meta.build"), func(index int, value string) {
					switch value {
					case "home":
						args = append(args, "-w", "/root")
					case "mount":
						p := kit.Path(m.Conf(web.DREAM, "meta.path"), arg[0])
						args = append(args, "--mount", kit.Format("type=bind,source=%s,target=/root", p))

						p = path.Join(p, ".bashrc")
						if _, e := os.Stat(p); e != nil {
							m.Cmd(nfs.SAVE, p, kit.Format("export ctx_dev=%s\nexport ctx_pod=%s ctx_user=%s\n",
								m.Conf(cli.RUNTIME, "conf.ctx_dev"), arg[0], cli.UserName))
						}

					}
				})

				// 创建容器
				repos := m.Conf(DOCKER, "meta.repos")
				pid := m.Cmdx(_docker, "run", "-dt", args, repos)
				m.Log_CREATE(repos, arg[0], "pid", pid)
				return

				m.Cmd(web.FAVOR, kit.Select(repos+".auto", arg, 1)).Table(func(index int, value map[string]string, head []string) {
					if value[kit.MDB_TYPE] == web.TYPE_SHELL {
						m.Cmd(_container, "exec", arg[0], kit.Split(value[kit.MDB_TEXT]))
						// 执行命令
					}
				})
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},
	},
}

func init() { code.Index.Register(Index, nil) }
