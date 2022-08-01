package cli

import (
	"os"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/logs"
)

func BinPath(arg ...string) string {
	return kit.Join(kit.Simple(arg, kit.Path(ice.BIN), kit.Path(ice.USR_LOCAL_BIN), kit.Path(ice.USR_LOCAL_GO_BIN), kit.Env(PATH)), ice.DF)
}

const FOREVER = "forever"

func init() {
	Index.MergeCommands(ice.Commands{
		FOREVER: {Name: "forever auto", Help: "启动", Actions: ice.Actions{
			gdb.START: {Name: "start", Help: "服务", Hand: func(m *ice.Message, arg ...string) {
				env := []string{PATH, BinPath(), HOME, kit.Select(kit.Path(""), os.Getenv(HOME))}
				for _, k := range ENV_LIST {
					if kit.Env(k) != "" {
						env = append(env, k, kit.Env(k))
					}
				}
				m.Option(CMD_ENV, env)

				m.Optionv(CMD_INPUT, os.Stdin)
				m.Optionv(CMD_OUTPUT, os.Stdout)
				m.Optionv(CMD_ERRPUT, os.Stderr)
				if p := kit.Env(CTX_LOG); p != "" {
					m.Optionv(CMD_ERRPUT, p)
				}

				m.Cmd(FOREVER, STOP)
				m.Cmdy(FOREVER, kit.Select(os.Args[0], nfs.PWD+ice.BIN_ICE_BIN, kit.FileExists(ice.BIN_ICE_BIN)),
					"serve", START, ice.DEV, "", aaa.USERNAME, aaa.ROOT, aaa.PASSWORD, aaa.ROOT, arg)
			}},
			gdb.RESTART: {Name: "restart", Help: "重启", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(gdb.SIGNAL, gdb.RESTART)
			}},
			gdb.STOP: {Name: "stop", Help: "停止", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(gdb.SIGNAL, gdb.STOP)
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				m.Cmdy(RUNTIME, BOOTINFO)
				return
			}

			for {
				logs.Println("run %s", kit.Join(arg, ice.SP))
				if m.Sleep("100ms"); IsSuccess(m.Cmd(SYSTEM, arg)) {
					logs.Println("exit")
					break
				}
			}
		}},
	})
}
