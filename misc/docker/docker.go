package docker

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/toolkits"
	"strings"
)

var Index = &ice.Context{Name: "docker", Help: "docker",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"docker": {Name: "docker", Help: "docker", Value: kit.Data(kit.MDB_SHORT, "name")},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},

		"docker": {Name: "docker", Help: "docker", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Echo("hello world")
		}},
		"image": {Name: "image", Help: "镜像管理", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Split(strings.Replace(m.Cmdx(ice.CLI_SYSTEM, "docker", "image", "ls"), "IMAGE ID", "IMAGE_ID", 1), "index", " ", "\n")
		}},
		"container": {Name: "container", Help: "容器管理", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			prefix := []string{ice.CLI_SYSTEM, "docker", "container"}
			if len(arg) > 2 {
				switch arg[1] {
				case "modify":
					switch arg[2] {
					case "NAMES":
						m.Cmd(prefix, "rename", arg[4], arg[3])
					}
				}
			}
			m.Split(strings.Replace(m.Cmdx(prefix, "ls"), "CONTAINER ID", "CONTAINER_ID", 1), "index", " ", "\n")
		}},
	},
}

func init() { cli.Index.Register(Index, nil) }
