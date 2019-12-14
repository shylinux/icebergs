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
			m.Log("info", "runtime %v", kit.Formats(m.Confv("runtime")))
		}},
		"runtime": {Name: "runtime", Help: "hello", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		}},
		"system": {Name: "system", Help: "hello", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			out := bytes.NewBuffer(make([]byte, 0, 1024))
			err := bytes.NewBuffer(make([]byte, 0, 1024))

			sys := exec.Command(arg[0], arg[1:]...)
			sys.Stdout = out
			sys.Stderr = err

			if e := sys.Run(); e != nil {
				m.Echo("error: ").Echo(kit.Select(e.Error(), err.String()))
				return
			}
			m.Echo(out.String())
		}},
		"timer": {Name: "timer", Help: "hello", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		}},
		"hi": {Name: "hi", Help: "hello", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Echo("hello %s world", c.Name)
		}},
	},
}

func init() { ice.Index.Register(Index, nil) }
