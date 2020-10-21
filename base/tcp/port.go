package tcp

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/nfs"
	kit "github.com/shylinux/toolkits"

	"net"
	"os"
	"path"
)

func _port_list(m *ice.Message, port string, dir string) {
	if m.Option(nfs.DIR_ROOT, path.Join("var/daemon", port)); port != "" {
		m.Cmdy(nfs.DIR, dir)
		return
	}

	m.Cmd(nfs.DIR, "./").Table(func(index int, value map[string]string, head []string) {
		m.Push(kit.MDB_TIME, value[kit.MDB_TIME])
		m.Push(kit.MDB_SIZE, value[kit.MDB_SIZE])
		m.Push(PORT, path.Base(value[kit.MDB_PATH]))
	})
}
func _port_right(m *ice.Message, begin string) string {
	current := kit.Int(kit.Select(m.Conf(PORT, "meta.current"), begin))
	end := kit.Int(m.Conf(PORT, "meta.end"))
	if current >= end {
		current = kit.Int(m.Conf(PORT, "meta.begin"))
	}

	for i := current; i < end; i++ {
		if c, e := net.Dial(TCP, kit.Format(":%d", i)); e == nil {
			m.Info("port exists %v", i)
			defer c.Close()
			continue
		}

		m.Log_SELECT(PORT, i)
		m.Conf(PORT, "meta.current", i)
		return kit.Format("%d", i)
	}
	return ""
}

const PORT = "port"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			PORT: {Name: PORT, Help: "端口", Value: kit.Data(
				"begin", 10000, "current", 10000, "end", 20000,
			)},
		},
		Commands: map[string]*ice.Command{
			PORT: {Name: "port port path auto", Help: "端口", Action: map[string]*ice.Action{
				aaa.RIGHT: {Name: "right [begin]", Help: "分配", Hand: func(m *ice.Message, arg ...string) {
					port, p := kit.Select("", arg, 0), ""
					for i := 0; i < 10; i++ {
						port = _port_right(m, port)
						p = path.Join(m.Conf(cli.DAEMON, kit.META_PATH), port)
						if _, e := os.Stat(p); e != nil && os.IsNotExist(e) {
							os.MkdirAll(p, ice.MOD_DIR)
							break
						}
						port = kit.Format(kit.Int(port) + 1)
					}
					m.Echo(port)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				_port_list(m, kit.Select("", arg, 0), kit.Select("", arg, 1))
			}},
		},
	}, nil)
}
