package tcp

import (
	"net"
	"path"
	"strconv"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _port_right(m *ice.Message, arg ...string) string {
	current, end := kit.Int(kit.Select(mdb.Config(m, CURRENT), arg, 0)), kit.Int(mdb.Config(m, END))
	kit.If(current >= end, func() { current = kit.Int(mdb.Config(m, BEGIN)) })
	for i := current; i < end; i++ {
		if p := path.Join(ice.USR_LOCAL_DAEMON, kit.Format(i)); nfs.Exists(m, p) {

		} else if c, e := net.Dial(TCP, kit.Format(":%d", i)); e == nil {
			m.Info("port exists %v", i)
			c.Close()
		} else {
			nfs.MkdirAll(m, p)
			m.Logs(mdb.SELECT, PORT, i)
			return mdb.Config(m, CURRENT, i)
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
		PORT: {Name: "port port path auto socket", Help: "端口", Actions: ice.MergeActions(ice.Actions{
			CURRENT:   {Hand: func(m *ice.Message, arg ...string) { m.Echo(mdb.Config(m, CURRENT)) }},
			aaa.RIGHT: {Hand: func(m *ice.Message, arg ...string) { m.Echo(_port_right(m, arg...)) }},
			nfs.TRASH: {Hand: func(m *ice.Message, arg ...string) {
				m.Assert(m.Option(PORT) != "")
				nfs.Trash(m, path.Join(ice.USR_LOCAL_DAEMON, m.Option(PORT)))
			}},
			"socket": {Hand: func(m *ice.Message, arg ...string) {
				parse := func(str string) int64 {
					port, _ := strconv.ParseInt(str, 16, 32)
					return port
				}
				trans := func(str string) string {
					switch str {
					case "0A":
						return "LISTEN"
					case "01":
						return "ESTABLISHED"
					case "06":
						return "TIME_WAIT"
					default:
						return str
					}
				}
				stats := map[string]int{}
				m.Spawn().Split(m.Cmdx(nfs.CAT, "/proc/net/tcp")).Table(func(value ice.Maps) {
					stats[trans(value["st"])]++
					m.Push("status", trans(value["st"]))
					ls := kit.Split(value["local_address"], ":")
					m.Push("local", kit.Format("%d.%d.%d.%d:%d", parse(ls[0][6:8]), parse(ls[0][4:6]), parse(ls[0][2:4]), parse(ls[0][:2]), parse(ls[1])))
					ls = kit.Split(value["rem_address"], ":")
					m.Push("remote", kit.Format("%d.%d.%d.%d:%d", parse(ls[0][6:8]), parse(ls[0][4:6]), parse(ls[0][2:4]), parse(ls[0][:2]), parse(ls[1])))
				})
				m.Spawn().Split(m.Cmdx(nfs.CAT, "/proc/net/tcp6")).Table(func(value ice.Maps) {
					stats[trans(value["st"])]++
					m.Push("status", trans(value["st"]))
					ls := kit.Split(value["local_address"], ":")
					m.Push("local", kit.Format("%d.%d.%d.%d:%d", parse(ls[0][30:32]), parse(ls[0][28:30]), parse(ls[0][26:28]), parse(ls[0][24:26]), parse(ls[1])))
					ls = kit.Split(value["remote_address"], ":")
					m.Push("remote", kit.Format("%d.%d.%d.%d:%d", parse(ls[0][30:32]), parse(ls[0][28:30]), parse(ls[0][26:28]), parse(ls[0][24:26]), parse(ls[1])))
				})
				m.Sort("status,local").StatusTimeCount(stats)
			}},
		}, mdb.HashAction(BEGIN, 10000, CURRENT, 10000, END, 20000)), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) > 0 {
				m.Cmdy(nfs.DIR, arg[1:], kit.Dict(nfs.DIR_ROOT, path.Join(ice.USR_LOCAL_DAEMON, arg[0])))
				return
			}
			current := kit.Int(mdb.Config(m, BEGIN))
			m.Options(nfs.DIR_ROOT, ice.USR_LOCAL_DAEMON).Cmd(nfs.DIR, nfs.PWD, func(value ice.Maps) {
				bin := m.Cmdv(nfs.DIR, path.Join(value[nfs.PATH], ice.BIN), nfs.PATH)
				kit.If(bin == "", func() { bin = m.Cmdv(nfs.DIR, path.Join(value[nfs.PATH], "sbin"), nfs.PATH) })
				port := kit.Int(path.Base(value[nfs.PATH]))
				m.Push(mdb.TIME, value[mdb.TIME]).Push(PORT, port).Push(nfs.SIZE, value[nfs.SIZE]).Push(ice.BIN, strings.TrimPrefix(bin, value[nfs.PATH]))
				current = kit.Max(current, port)
			})
			m.PushAction(nfs.TRASH).StatusTimeCount(ctx.ConfigSimple(m, BEGIN, CURRENT, END)).SortInt(PORT)
			mdb.Config(m, CURRENT, current)
		}},
	})
}
