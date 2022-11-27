package tcp

import (
	"net"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _port_right(m *ice.Message, arg ...string) string {
	current, end := kit.Int(kit.Select(m.Config(CURRENT), arg, 0)), kit.Int(m.Config(END))
	if current >= end {
		current = kit.Int(m.Config(BEGIN))
	}
	for i := current; i < end; i++ {
		if p := path.Join(ice.USR_LOCAL_DAEMON, kit.Format(i)); nfs.ExistsFile(m, p) {

		} else if c, e := net.Dial(TCP, kit.Format(":%d", i)); e == nil {
			m.Info("port exists %v", i)
			c.Close()
		} else {
			nfs.MkdirAll(m, p)
			m.Logs(mdb.SELECT, PORT, i)
			return m.Config(CURRENT, i)
		}
	}
	return ""
}

const (
	BEGIN   = "begin"
	CURRENT = "current"
	RANDOM  = "random"
	END     = "end"
)
const PORT = "port"

func init() {
	Index.MergeCommands(ice.Commands{
		PORT: {Name: "port port path auto", Help: "端口", Actions: ice.MergeActions(ice.Actions{
			CURRENT:   {Hand: func(m *ice.Message, arg ...string) { m.Echo(m.Config(CURRENT)) }},
			aaa.RIGHT: {Hand: func(m *ice.Message, arg ...string) { m.Echo(_port_right(m, arg...)) }},
			nfs.TRASH: {Hand: func(m *ice.Message, arg ...string) {
				m.Assert(m.Option(PORT) != "")
				nfs.Trash(m, path.Join(ice.USR_LOCAL_DAEMON, m.Option(PORT)))
			}},
		}, mdb.HashAction(BEGIN, 10000, CURRENT, 10000, END, 20000)), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) > 0 {
				m.Cmdy(nfs.DIR, arg[1:], kit.Dict(nfs.DIR_ROOT, path.Join(ice.USR_LOCAL_DAEMON, arg[0])))
				return
			}
			current := kit.Int(m.Config(BEGIN))
			m.Cmd(nfs.DIR, nfs.PWD, kit.Dict(nfs.DIR_ROOT, ice.USR_LOCAL_DAEMON), func(value ice.Maps) {
				bin := m.CmdAppend(nfs.DIR, path.Join(value[nfs.PATH], ice.BIN), nfs.PATH)
				if bin == "" {
					bin = m.CmdAppend(nfs.DIR, path.Join(value[nfs.PATH], "sbin"), nfs.PATH)
				}
				port := kit.Int(path.Base(value[nfs.PATH]))
				m.Push(mdb.TIME, value[mdb.TIME])
				m.Push(PORT, port)
				m.Push(nfs.SIZE, value[nfs.SIZE])
				m.Push(ice.BIN, strings.TrimPrefix(bin, value[nfs.PATH]))
				current = kit.Max(current, port)
			})
			m.Config(CURRENT, current)
			m.PushAction(nfs.TRASH).StatusTimeCount(m.ConfigSimple(BEGIN, CURRENT, END)).SortInt(PORT)
		}},
	})
}
