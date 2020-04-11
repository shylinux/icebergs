package cli

import (
	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"

	"bytes"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path"
	"runtime"
	"strings"
)

var Index = &ice.Context{Name: "cli", Help: "命令模块",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		ice.CLI_RUNTIME: {Name: "runtime", Help: "运行环境", Value: kit.Dict()},
		ice.CLI_SYSTEM:  {Name: "system", Help: "系统命令", Value: kit.Data()},
		ice.CLI_DAEMON:  {Name: "daemon", Help: "守护进程", Value: kit.Data(kit.MDB_SHORT, "name")},
		"python":        {Name: "python", Help: "系统命令", Value: kit.Data("python", "python", "pip", "pip")},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Load()

			// 启动配置
			m.Conf(ice.CLI_RUNTIME, "conf.ctx_self", os.Getenv("ctx_self"))
			m.Conf(ice.CLI_RUNTIME, "conf.ctx_dev", os.Getenv("ctx_dev"))
			m.Conf(ice.CLI_RUNTIME, "conf.ctx_shy", os.Getenv("ctx_shy"))
			m.Conf(ice.CLI_RUNTIME, "conf.ctx_pid", os.Getenv("ctx_pid"))

			// 主机信息
			m.Conf(ice.CLI_RUNTIME, "host.GOARCH", runtime.GOARCH)
			m.Conf(ice.CLI_RUNTIME, "host.GOOS", runtime.GOOS)
			m.Conf(ice.CLI_RUNTIME, "host.pid", os.Getpid())

			// 启动信息
			if name, e := os.Hostname(); e == nil {
				m.Conf(ice.CLI_RUNTIME, "boot.hostname", kit.Select(name, os.Getenv("HOSTNAME")))
			}
			if user, e := user.Current(); e == nil {
				m.Conf(ice.CLI_RUNTIME, "boot.username", path.Base(kit.Select(user.Name, os.Getenv("USER"))))
				m.Cmd(ice.AAA_ROLE, "root", m.Conf(ice.CLI_RUNTIME, "boot.username"))
			}
			if name, e := os.Getwd(); e == nil {
				name = path.Base(kit.Select(name, os.Getenv("PWD")))
				ls := strings.Split(name, "/")
				name = ls[len(ls)-1]
				ls = strings.Split(name, "\\")
				name = ls[len(ls)-1]
				m.Conf(ice.CLI_RUNTIME, "boot.pathname", name)
			}

			// 启动记录
			count := m.Confi(ice.CLI_RUNTIME, "boot.count") + 1
			m.Conf(ice.CLI_RUNTIME, "boot.count", count)

			// 节点信息
			m.Conf(ice.CLI_RUNTIME, "node.time", m.Time())
			m.Conf(ice.CLI_RUNTIME, "node.type", ice.WEB_WORKER)
			m.Conf(ice.CLI_RUNTIME, "node.name", m.Conf(ice.CLI_RUNTIME, "boot.pathname"))
			m.Log("info", "runtime %v", kit.Formats(m.Confv(ice.CLI_RUNTIME)))
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Save(ice.CLI_RUNTIME, ice.CLI_SYSTEM)
		}},

		ice.CLI_RUNTIME: {Name: "runtime", Help: "运行环境", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		}},
		ice.CLI_SYSTEM: {Name: "system", Help: "系统命令", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			cmd := exec.Command(arg[0], arg[1:]...)

			// 运行目录
			cmd.Dir = m.Option("cmd_dir")
			if len(cmd.Dir) > 0 {
				m.Info("dir: %s", cmd.Dir)
				if _, e := os.Stat(cmd.Dir); e != nil && os.IsNotExist(e) {
					os.MkdirAll(cmd.Dir, 0777)
				}
			}

			// 环境变量
			env := kit.Simple(m.Optionv("cmd_env"))
			for i := 0; i < len(env)-1; i += 2 {
				cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", env[i], env[i+1]))
			}
			if len(cmd.Env) > 0 {
				m.Info("env: %s", cmd.Env)
			}
			if m.Option("cmd_stdout") != "" {
				if f, p, e := kit.Create(m.Option("cmd_stdout")); m.Assert(e) {
					m.Info("stdout: %s", p)
					cmd.Stdout = f
					cmd.Stderr = f
				}
			}
			if m.Option("cmd_stderr") != "" {
				if f, p, e := kit.Create(m.Option("cmd_stderr")); m.Assert(e) {
					m.Info("stderr: %s", p)
					cmd.Stderr = f
				}
			}

			switch m.Option("cmd_type") {
			case "daemon":
				// 守护进程
				cmd.Env = append(cmd.Env, fmt.Sprintf("PATH=%s", os.Getenv("PATH")))
				if e := cmd.Start(); e != nil {
					m.Warn(e != nil, "%v start: %s", arg, e)
					return
				}

				m.Rich("daemon", nil, kit.Dict(kit.MDB_NAME, cmd.Process.Pid, "status", "running"))
				m.Echo("%d", cmd.Process.Pid)

				m.Gos(m, func(m *ice.Message) {
					if e := cmd.Wait(); e != nil {
						m.Warn(e != nil, "%v wait: %s", arg, e)
					} else {
						m.Cost("%v exit: %v", arg, cmd.ProcessState.ExitCode())
						m.Rich("daemon", nil, func(key string, value map[string]interface{}) {
							value["status"] = "exited"
						})
					}
				})

			default:
				// 系统命令
				out := bytes.NewBuffer(make([]byte, 0, 1024))
				err := bytes.NewBuffer(make([]byte, 0, 1024))
				cmd.Stdout = out
				cmd.Stderr = err
				if e := cmd.Run(); e != nil {
					m.Warn(e != nil, "%v run: %s", arg, kit.Select(e.Error(), err.String()))
				} else {
					m.Cost("%v exit: %v", arg, cmd.ProcessState.ExitCode())
				}
				m.Push("code", int(cmd.ProcessState.ExitCode()))
				m.Echo(out.String())
			}
		}},
		ice.CLI_DAEMON: {Name: "daemon", Help: "守护进程", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Option("cmd_type", "daemon")
			m.Cmdy(ice.CLI_SYSTEM, arg)
		}},
		"python": {Name: "python", Help: "运行环境", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			prefix := []string{ice.CLI_SYSTEM, m.Conf("python", "meta.python")}
			switch arg[0] {
			case "qrcode":
				m.Cmdy(prefix, "-c", fmt.Sprintf(`import pyqrcode; print(pyqrcode.create('%s').terminal(module_color='%s', quiet_zone=1))`, kit.Select("hello world", arg, 1), kit.Select("blue", arg, 2)))
			case "install":
				m.Cmdy(prefix[:1], m.Conf("python", "meta.pip"), "install", arg[1:])
			default:
				m.Cmdy(prefix, arg)
			}
		}},
	},
}

func init() { ice.Index.Register(Index, nil) }
