package cli

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func init() {
	Index.MergeCommands(ice.Commands{
		PROCINFO: {Name: "procinfo PID auto filter:text monitor docker", Help: "进程列表", Actions: ice.MergeActions(ice.Actions{
			PROCKILL:   {Help: "结束进程", Hand: func(m *ice.Message, arg ...string) { m.Cmdy(gdb.SIGNAL, gdb.STOP, m.Option("PID")).ProcessRefresh() }},
			"terminal": {Help: "终端", Hand: func(m *ice.Message, arg ...string) { Opens(m, "Terminal.app") }},
			"monitor":  {Help: "监控", Hand: func(m *ice.Message, arg ...string) { Opens(m, "Activity Monitor.app") }},
			"docker":   {Help: "容器", Hand: func(m *ice.Message, arg ...string) { Opens(m, "Docker Desktop.app") }},
		}), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 || arg[0] == "" {
				m.Split(m.Cmdx(SYSTEM, "ps", "u")).PushAction(PROCKILL).Sort("COMMAND").StatusTimeCount(m.Cmd(RUNTIME, HOSTINFO).AppendSimple("nCPU,MemTotal,MemFree"))
				return
			}
			m.OptionFields(mdb.DETAIL)
			for _, line := range kit.SplitLine(m.Cmdx(nfs.CAT, kit.Format("/proc/%s/status", arg[0]))) {
				ls := strings.SplitN(line, nfs.DF, 2)
				m.Push(ls[0], ls[1])
			}
		}},
	})
}
