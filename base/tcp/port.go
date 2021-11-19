package tcp

import (
	"net"
	"os"
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _port_right(m *ice.Message, arg ...string) string {
	current := kit.Int(kit.Select(m.Config(CURRENT), arg, 0))
	end := kit.Int(m.Config(END))
	if current >= end {
		current = kit.Int(m.Config(BEGIN))
	}

	for i := current; i < end; i++ {
		if c, e := net.Dial(TCP, kit.Format(":%d", i)); e == nil {
			m.Info("port exists %v", i)
			c.Close()
			continue
		}
		p := path.Join(m.Conf(cli.DAEMON, kit.META_PATH), kit.Format(i))
		if _, e := os.Stat(p); e == nil {
			continue
		}
		os.MkdirAll(p, ice.MOD_DIR)

		m.Log_SELECT(PORT, i)
		return m.Config(CURRENT, i)
	}
	return ""
}

const (
	RANDOM  = "random"
	CURRENT = "current"
	BEGIN   = "begin"
	END     = "end"
)
const PORT = "port"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		PORT: {Name: PORT, Help: "端口", Value: kit.Data(BEGIN, 10000, CURRENT, 10000, END, 20000)},
	}, Commands: map[string]*ice.Command{
		PORT: {Name: "port port path auto", Help: "端口", Action: map[string]*ice.Action{
			aaa.RIGHT: {Name: "right", Help: "分配", Hand: func(m *ice.Message, arg ...string) {
				m.Echo(_port_right(m, arg...))
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				m.Option(nfs.DIR_ROOT, m.Conf(cli.DAEMON, kit.META_PATH))
				m.Cmd(nfs.DIR, ice.PWD, "time,path,size").Table(func(index int, value map[string]string, head []string) {
					m.Push(kit.MDB_TIME, value[kit.MDB_TIME])
					m.Push(PORT, path.Base(value[kit.MDB_PATH]))
					m.Push(kit.MDB_SIZE, value[kit.MDB_SIZE])
				})
				m.SortInt(PORT)
				return
			}
			m.Option(nfs.DIR_ROOT, path.Join(m.Conf(cli.DAEMON, kit.META_PATH), arg[0]))
			m.Cmdy(nfs.DIR, arg[1:])
		}},
	}})
}
