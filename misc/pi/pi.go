package pi

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/core/chat"
	"github.com/shylinux/toolkits"
	"os"
	"path"
)

var Index = &ice.Context{Name: "pi", Help: "pi",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"pi": {Name: "pi", Help: "pi", Value: kit.Data(kit.MDB_SHORT, "name")},
	},
	Commands: map[string]*ice.Command{
		"GPIO": {Name: "GPIO", Help: "GPIO", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			p := kit.Format("/sys/class/gpio/gpio%s", arg[0])
			if _, e := os.Stat(p); e != nil {
				if m.Warn(!os.IsNotExist(e), "%s", e) {
					return
				}
				m.Cmd("nfs.echo", "/sys/class/gpio/export", arg[0])
			}

			if len(arg) > 1 {
				m.Cmd("nfs.echo", path.Join(p, "direction"), "out")
				m.Cmd("nfs.echo", path.Join(p, "value"), arg[1])
			}
			m.Cmdy("nfs.cat", path.Join(p, "value"))
		}},
	},
}

func init() { chat.Index.Register(Index, nil) }
