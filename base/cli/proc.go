package cli

import (
	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"

	"os"
	"strings"
)

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{},
		Commands: map[string]*ice.Command{
			"proc": {Name: "proc name=ice.bin PID auto", Help: "进程管理", Action: map[string]*ice.Action{
				"kill": {Name: "kill", Help: "结束", Hand: func(m *ice.Message, arg ...string) {
					if p, e := os.FindProcess(kit.Int(m.Option("PID"))); m.Assert(e) {
						m.Assert(p.Kill())
					}
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				msg := m.Spawn()
				msg.Split(m.Cmdx(SYSTEM, "ps", "ux"), "", " ", "\n")
				msg.Table(func(index int, value map[string]string, head []string) {
					if m.Appendv(ice.MSG_APPEND, "action", head); len(arg) > 1 && value["PID"] != arg[1] {
						return
					}
					if len(arg) > 0 && !strings.Contains(value["COMMAND"], arg[0]) {
						return
					}
					m.PushButton("结束")
					m.Push("", value)
				})
			}},
		},
	}, nil)
}
