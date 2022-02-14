package cli

import (
	"os"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const FOREVER = "forever"

func init() {
	const SERVE = "serve"
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		FOREVER: {Name: "forever", Help: "启动", Action: map[string]*ice.Action{
			SERVE: {Name: "serve", Help: "服务", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(FOREVER, os.Args[0], SERVE, START, ice.DEV, "", aaa.USERNAME, aaa.ROOT, aaa.PASSWORD, aaa.ROOT, arg)
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			for {
				if arg[1] == SERVE {
					if _, e := os.Stat(ice.BIN_ICE_BIN); e == nil {
						arg[0] = nfs.PWD + ice.BIN_ICE_BIN
					}
				}
				println(kit.Format("%s run %s", kit.Now(), kit.Join(arg, ice.SP)))
				m.Option(CMD_OUTPUT, os.Stdout)
				m.Option(CMD_ERRPUT, os.Stderr)
				switch msg := m.Cmd(SYSTEM, arg); msg.Append(CODE) {
				case "0":
					println(kit.Format("%s exit", kit.Now()))
					return
				default:
					m.Sleep("1s")
				}
			}
		}}},
	})
}
