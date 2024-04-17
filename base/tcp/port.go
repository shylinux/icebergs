package tcp

import (
	"net"
	"path"
	"runtime"
	"strconv"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _port_right(m *ice.Message, current, begin, end int) string {
	kit.If(current >= end, func() { current = begin })
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
	PORT_22   = "22"
	PORT_80   = "80"
	PORT_443  = "443"
	PORT_9020 = "9020"
	PORT_9022 = "9022"

	SOCKET  = "socket"
	BEGIN   = "begin"
	CURRENT = "current"
	RANDOM  = "random"
	END     = "end"
	PID     = "pid"
	SPACE   = "space"
)
const PORT = "port"

func init() {
	Index.MergeCommands(ice.Commands{
		PORT: {Name: "port port path auto socket", Help: "端口", Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case HOST, SERVER:
					m.Cmd(PORT, SOCKET, func(value ice.Maps) {
						switch value[mdb.STATUS] {
						case "LISTEN":
							m.Push(arg[0], strings.Replace(value["local"], "0.0.0.0", "127.0.0.1", 1))
						}
					})
				case PORT:
					if runtime.GOOS == "darwin" {
						ls := kit.SplitLine(m.Cmd("system", "sh", "-c", `lsof -nP -i4TCP | grep LISTEN | awk '{print $1 " " $9 }'`).Result())
						kit.For(ls, func(p string) {
							ls := kit.SplitWord(p)
							m.Push(arg[0], kit.Split(ls[1], ":")[1]).Push(SERVER, ls[0])
						})
						m.Sort(arg[0], ice.INT)
						return
					}
					m.Cmd(PORT, SOCKET, func(value ice.Maps) {
						switch value[mdb.STATUS] {
						case "LISTEN":
							m.Push(arg[0], strings.TrimPrefix(value["local"], "0.0.0.0:"))
							m.Push(mdb.NAME, "listen")
						}
					})
				}
			}},
			SOCKET: {Help: "端口", Hand: func(m *ice.Message, arg ...string) {
				parse := func(str string) int64 { port, _ := strconv.ParseInt(str, 16, 32); return port }
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
					m.Push(mdb.STATUS, trans(value["st"]))
					ls := kit.Split(value["local_address"], ":")
					m.Push("local", kit.Format("%d.%d.%d.%d:%d", parse(ls[0][6:8]), parse(ls[0][4:6]), parse(ls[0][2:4]), parse(ls[0][:2]), parse(ls[1])))
					ls = kit.Split(value["rem_address"], ":")
					m.Push("remote", kit.Format("%d.%d.%d.%d:%d", parse(ls[0][6:8]), parse(ls[0][4:6]), parse(ls[0][2:4]), parse(ls[0][:2]), parse(ls[1])))
				})
				m.Spawn().Split(m.Cmdx(nfs.CAT, "/proc/net/tcp6")).Table(func(value ice.Maps) {
					stats[trans(value["st"])]++
					m.Push(mdb.STATUS, trans(value["st"]))
					ls := kit.Split(value["local_address"], ":")
					m.Push("local", kit.Format("%d.%d.%d.%d:%d", parse(ls[0][30:32]), parse(ls[0][28:30]), parse(ls[0][26:28]), parse(ls[0][24:26]), parse(ls[1])))
					ls = kit.Split(value["remote_address"], ":")
					m.Push("remote", kit.Format("%d.%d.%d.%d:%d", parse(ls[0][30:32]), parse(ls[0][28:30]), parse(ls[0][26:28]), parse(ls[0][24:26]), parse(ls[1])))
				})
				m.Sort("status,local").StatusTimeCount(stats)
			}},
			nfs.TRASH: {Hand: func(m *ice.Message, arg ...string) {
				m.Assert(m.Option(PORT) != "")
				nfs.Trash(m, path.Join(ice.USR_LOCAL_DAEMON, m.Option(PORT)))
				mdb.HashRemove(m)
			}},
			aaa.RIGHT: {Hand: func(m *ice.Message, arg ...string) { m.Echo(PortRight(m, arg...)) }},
			CURRENT:   {Hand: func(m *ice.Message, arg ...string) { m.Echo(mdb.Config(m, CURRENT)) }},
			STOP:      {Hand: func(m *ice.Message, arg ...string) { PortCmds(m, arg...); mdb.HashModify(m, PID, "") }},
			START:     {Hand: func(m *ice.Message, arg ...string) { PortCmds(m, arg...); mdb.HashModify(m, PID, m.Append(PID)) }},
		}, mdb.HashAction(BEGIN, 10000, END, 20000,
			mdb.SHORT, PORT, mdb.FIELD, "time,port,pid,cmd,name,text,icon,space,index",
		)), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) > 0 {
				m.Cmdy(nfs.DIR, arg[1:], kit.Dict(nfs.DIR_ROOT, path.Join(ice.USR_LOCAL_DAEMON, arg[0])))
				return
			}
			current := kit.Int(mdb.Config(m, BEGIN))
			mdb.HashSelect(m, arg...).Table(func(value ice.Maps) {
				current = kit.Max(current, kit.Int(value[PORT]))
				if value[PID] == "" {
					m.PushButton(START, nfs.TRASH)
				} else {
					m.PushButton(STOP)
				}
			})
			mdb.Config(m, CURRENT, current)
			m.StatusTimeCount(mdb.ConfigSimple(m, BEGIN, CURRENT, END)).SortInt(PORT)
		}},
	})
	ice.Info.Inputs = append(ice.Info.Inputs, func(m *ice.Message, arg ...string) {
		switch arg[0] {
		case PORT:
			m.SetAppend().Cmdy(PORT, mdb.INPUTS, arg)
		}
	})
}
func PortRight(m *ice.Message, arg ...string) string {
	current, begin, end := kit.Select("20000", mdb.Config(m, CURRENT)), kit.Select("20000", mdb.Config(m, BEGIN)), kit.Select("30000", mdb.Config(m, END))
	return _port_right(m, kit.Int(kit.Select(kit.Select(begin, current), arg, 0)), kit.Int(kit.Select(begin, arg, 1)), kit.Int(kit.Select(end, arg, 2)))
}
func PortCmds(m *ice.Message, arg ...string) {
	m.Cmdy(SPACE, m.Option(SPACE), m.Option(ctx.INDEX), m.ActionKey())
}
