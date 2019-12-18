package cli

import (
	"bytes"
	"github.com/shylinux/icebergs"
	"github.com/shylinux/toolkits"
	"os"
	"os/exec"
	"os/user"
	"path"
	"runtime"
)

var Index = &ice.Context{Name: "cli", Help: "命令模块",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"runtime": {Name: "runtime", Value: map[string]interface{}{
			"host": map[string]interface{}{},
			"boot": map[string]interface{}{},
			"node": map[string]interface{}{},
			"user": map[string]interface{}{},
			"work": map[string]interface{}{},
		}},
	},
	Commands: map[string]*ice.Command{
		"_init": {Name: "_init", Help: "hello", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd("ctx.config", "load", "var/conf/cli.json")

			m.Conf("runtime", "host.GOARCH", runtime.GOARCH)
			m.Conf("runtime", "host.GOOS", runtime.GOOS)
			m.Conf("runtime", "host.pid", os.Getpid())

			if name, e := os.Hostname(); e == nil {
				m.Conf("runtime", "boot.hostname", kit.Select(name, os.Getenv("HOSTNAME")))
			}
			if user, e := user.Current(); e == nil {
				m.Conf("runtime", "boot.username", path.Base(kit.Select(user.Name, os.Getenv("USER"))))
			}
			if name, e := os.Getwd(); e == nil {
				m.Conf("runtime", "boot.pathname", path.Base(kit.Select(name, os.Getenv("PWD"))))
			}

			m.Conf("runtime", "node.type", "worker")
			m.Conf("runtime", "node.name", m.Conf("runtime", "boot.pathname"))
			m.Log("info", "runtime %v", kit.Formats(m.Confv("runtime")))
		}},
		"_exit": {Name: "_exit", Help: "hello", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd("ctx.config", "save", "var/conf/cli.json", "cli.runtime")
		}},
		"runtime": {Name: "runtime", Help: "hello", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		}},
		"system": {Name: "system", Help: "hello", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			cmd := exec.Command(arg[0], arg[1:]...)
			cmd.Dir = m.Option("cmd_dir")

			m.Log("info", "dir: %s", cmd.Dir)
			if m.Option("cmd_type") == "daemon" {
				m.Gos(m, func(m *ice.Message) {
					if e := cmd.Start(); e != nil {
						m.Log("warn", "%v start %s", arg, e)
					} else if e := cmd.Wait(); e != nil {
						m.Log("warn", "%v wait %s", arg, e)
					} else {
						m.Log("info", "%v exit", arg)
					}
				})
				return
			}

			out := bytes.NewBuffer(make([]byte, 0, 1024))
			err := bytes.NewBuffer(make([]byte, 0, 1024))

			cmd.Stdout = out
			cmd.Stderr = err

			if e := cmd.Run(); e != nil {
				m.Echo("error: ").Echo(kit.Select(e.Error(), err.String()))
				return
			}
			m.Echo(out.String())
		}},
		"timer": {Name: "timer", Help: "hello", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		}},
	},
}

func init() { ice.Index.Register(Index, nil) }
