package docker

import (
	"path"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/gdb"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/tcp"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/code"
	kit "github.com/shylinux/toolkits"

	"strings"
)

func MOD(str string) string { return str }
func CMD(str string) string { return str }
func ARG(str string) string { return str }

var DOCKER = MOD("docker")
var (
	IMAGE     = CMD("image")
	CONTAINER = CMD("container")
	COMMAND   = CMD("command")
)
var (
	ALPINE = ARG("alpine")
	CENTOS = ARG("centos")
)
var (
	_REPOSITORY = ARG("REPOSITORY")
	_TAG        = ARG("TAG")

	_CONTAINER_ID = ARG("CONTAINER_ID")
	_IMAGE_ID     = ARG("IMAGE_ID")
)

var _docker = []string{cli.SYSTEM, DOCKER}
var _image = []string{cli.SYSTEM, DOCKER, IMAGE}
var _container = []string{cli.SYSTEM, DOCKER, CONTAINER}

var Index = &ice.Context{Name: DOCKER, Help: "虚拟机",
	Configs: map[string]*ice.Config{
		DOCKER: {Name: DOCKER, Help: "虚拟机", Value: kit.Data(
			"path", "usr/docker",
			"repos", CENTOS, "build", []interface{}{
				"home",
				// "mount",
			},
			ALPINE, []interface{}{
				`sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories`,
				`apk add curl`,
			},
			CENTOS, []interface{}{
				`curl -o /etc/yum.repos.d/CentOS-Base.repo http://mirrors.aliyun.com/repo/Centos-8.repo`,
				`yum makecache`,
			},
		)},
		IMAGE: {Name: IMAGE, Help: "镜像", Value: kit.Data(
			"action", []interface{}{"build", "push", "pull", "start", "clear"},
		)},
		CONTAINER: {Name: CONTAINER, Help: "容器", Value: kit.Data(
			"action", []interface{}{"open", "start", "stop", "restart", "clear"},
		)},
	},
	Commands: map[string]*ice.Command{
		IMAGE: {Name: "image IMAGE_ID=auto auto 清理:button", Help: "镜像管理", Meta: kit.Dict(
			"detail", []string{"运行", "清理", "删除"},
		), Action: map[string]*ice.Action{
			web.PULL: {Name: "pull", Help: "更新", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(_image, "pull", m.Option(_REPOSITORY))
			}},
			web.PUSH: {Name: "push", Help: "上传", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(_image, "push", m.Option(_REPOSITORY))
			}},
			gdb.BUILD: {Name: "build", Help: "生成", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(cli.SYSTEM, "rm", "-r", "usr/docker/meta/demo/")
				m.Cmd(cli.SYSTEM, "rm", "-r", "usr/docker/meta/volcanos/")
				m.Cmd(cli.SYSTEM, "cp", "-r", "usr/demo/", "usr/docker/meta/demo/")
				m.Cmd(cli.SYSTEM, "cp", "-r", "usr/volcanos/", "usr/docker/meta/volcanos/")
				m.Cmdy(_docker, "build", m.Conf(DOCKER, "meta.path"), "-t", m.Option(_REPOSITORY),
					"-f", path.Join(m.Conf(DOCKER, "meta.path"), m.Option(_REPOSITORY)),
				)
			}},
			gdb.START: {Name: "start", Help: "运行", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(_docker, "run", "-dt", "-p", m.Cmdx(tcp.PORT, "get")+":9020", m.Option(_REPOSITORY)+":"+m.Option(_TAG))
			}},
			gdb.PRUNE: {Name: "prune", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(_image, "prune", "-f")
			}},
			gdb.CLEAR: {Name: "clear", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(_image, "rm", m.Option(_IMAGE_ID))
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 0 {
				// 容器详情
				res := m.Cmdx(_image, "inspect", arg[0])
				m.PushDetail(kit.KeyValue(nil, "", kit.Parse(nil, "", kit.Split(res)...)))
				return
			}

			// 镜像列表
			m.Split(strings.Replace(m.Cmdx(_docker, "images"), "IMAGE ID", _IMAGE_ID, 1), "index", " ", "\n")
			m.Sort(_REPOSITORY)

			// 镜像操作
			m.PushAction(m.Confv(IMAGE, "meta.action"))
		}},
		CONTAINER: {Name: "container CONTAINER_ID=auto auto 清理:button", Help: "容器管理", Meta: kit.Dict(
			"detail", []string{"进入", "启动", "停止", "重启", "清理", "编辑", "删除"},
		), Action: map[string]*ice.Action{
			gdb.OPEN: {Name: "open", Help: "进入", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd("web.code.tmux", m.Option("NAMES"))
				m.Cmdy(cli.SYSTEM, "tmux", "new-window", "-t", m.Option("NAMES"), "-n", m.Option("NAMES"),
					"-PF", "#{session_name}:#{window_name}.1", "docker exec -it "+m.Option("NAMES")+" sh")
			}},
			gdb.START: {Name: "start", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(_container, "start", m.Option(_CONTAINER_ID))
			}},
			gdb.STOP: {Name: "stop", Help: "停止", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(_container, "stop", m.Option(_CONTAINER_ID))
			}},
			gdb.RESTART: {Name: "restart", Help: "重启", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(_container, "restart", m.Option(_CONTAINER_ID))
			}},
			gdb.CHANGE: {Name: "change", Help: "更改", Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case "NAMES":
					m.Cmdy(_container, "rename", m.Option(_CONTAINER_ID), arg[1])
				}
			}},
			gdb.PRUNE: {Name: "prune", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(_container, "prune", "-f")
			}},
			gdb.CLEAR: {Name: "clear", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(_container, "rm", m.Option(_CONTAINER_ID))
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 0 {
				// 容器详情
				res := m.Cmdx(_container, "inspect", arg[0])
				m.PushDetail(kit.KeyValue(nil, "", kit.Parse(nil, "", kit.Split(res)...)))
				return
			}

			// 容器列表
			m.Split(strings.Replace(m.Cmdx(_docker, "ps", "-a"), "CONTAINER ID", _CONTAINER_ID, 1), "index", " ", "\n")
			m.Sort("NAMES")
			m.Table(func(index int, value map[string]string, head []string) {
				if strings.TrimSpace(value["PORTS"]) == "" {
					return
				}
				ls := strings.Split(value["PORTS"], "->")
				ls = strings.Split(ls[0], ":")
				u := kit.ParseURL(m.Option(ice.MSG_USERWEB))
				p := kit.Format("http://%s:%s", u.Hostname(), ls[1])
				m.Echo("%s\n", m.Cmdx(mdb.RENDER, web.RENDER.A, p, p))
			})

			// 容器操作
			m.PushAction(m.Confv(CONTAINER, "meta.action"))
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
		}},
		gdb.AUTO: {Name: "auto", Help: "自动化", Action: map[string]*ice.Action{
			web.DREAM_START: {Hand: func(m *ice.Message, arg ...string) {
				if m.Cmd(_container, "start", arg[0]).Append(cli.CMD_CODE) == "0" {
					// 重启容器
					return
				}

				args := []string{"--name", arg[0],
					"-e", "ctx_user=" + cli.UserName,
					"-e", "ctx_dev=" + m.Conf(cli.RUNTIME, "conf.ctx_dev"),
				}
				kit.Fetch(m.Confv(DOCKER, "meta.build"), func(index int, value string) {
					switch value {
					case "home":
						args = append(args, "-w", "/root")
					case "mount":
						p := kit.Path(m.Conf(web.DREAM, "meta.path"), arg[0])
						args = append(args, "--mount", kit.Format("type=bind,source=%s,target=/root", p))
					}
				})

				// 创建容器
				repos := m.Conf(DOCKER, "meta.repos")
				pid := m.Cmdx(_docker, "run", "-dt", args, repos)
				m.Log_CREATE(repos, arg[0], "pid", pid)

				kit.Fetch(m.Confv(DOCKER, kit.Keys("meta", repos)), func(index int, value string) {
					m.Logs("cmd", value, "res", m.Cmdx(_container, "exec", arg[0], kit.Split(value)))
				})
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},
	},
}

func init() { code.Index.Register(Index, nil) }
