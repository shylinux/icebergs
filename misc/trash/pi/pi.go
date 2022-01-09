package pi

import (
	"os"
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/core/chat"
	kit "shylinux.com/x/toolkits"
)

var Index = &ice.Context{Name: "pi", Help: "开发板",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"pi": {Name: "pi", Help: "pi", Value: kit.Data(mdb.SHORT, "name")},
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
