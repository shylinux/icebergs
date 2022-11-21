package cli

import (
	"os"
	"path"

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
			START: {Hand: func(m *ice.Message, arg ...string) {
				env := []string{PATH, BinPath(), HOME, kit.Select(kit.Path(""), os.Getenv(HOME))}
				for _, k := range ENV_LIST {
					if kit.Env(k) != "" {
						env = append(env, k, kit.Env(k))
					}
				}
				for _, v := range os.Environ() {
					if ls := kit.Split(v, ice.EQ, ice.EQ); kit.IndexOf(env, ls[0]) == -1 && len(ls) > 1 {
						env = append(env, ls[0], ls[1])
					}
				}
				m.Optionv(CMD_ENV, env)
				m.Optionv(CMD_INPUT, os.Stdin)
				m.Optionv(CMD_OUTPUT, os.Stdout)
				m.Optionv(CMD_ERRPUT, os.Stderr)
				if p := kit.Env(CTX_LOG); p != "" {
					m.Optionv(CMD_ERRPUT, p)
				}
				m.Cmd(FOREVER, STOP)
				if bin := kit.Select(os.Args[0], ice.BIN_ICE_BIN, nfs.ExistsFile(m, ice.BIN_ICE_BIN)); len(arg) > 0 && arg[0] == ice.SPACE {
					m.Cmdy(FOREVER, bin, ice.SPACE, "dial", ice.DEV, ice.OPS, arg[2:])
				} else {
					m.Cmdy(FOREVER, bin, ice.SERVE, START, ice.DEV, "", aaa.USERNAME, aaa.ROOT, aaa.PASSWORD, aaa.ROOT, arg)
				}
			}},
			RESTART: {Hand: func(m *ice.Message, arg ...string) { m.Cmd(gdb.SIGNAL, gdb.RESTART) }},
			STOP:    {Hand: func(m *ice.Message, arg ...string) { m.Cmd(gdb.SIGNAL, gdb.STOP) }},
		}, Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				m.Cmdy(RUNTIME, BOOTINFO)
				return
			}
			for {
				if logs.Println("run %s", kit.Join(arg, ice.SP)); IsSuccess(m.Cmd(SYSTEM, arg)) {
					logs.Println(ice.EXIT)
					break
				}
				if logs.Println(); m.Config("log.save") == ice.TRUE {
					back := kit.Format("var/log.%s", logs.Now().Format("20060102_150405"))
					m.Cmd(SYSTEM, "cp", "-r", "var/log", back, ice.Maps{CMD_OUTPUT: ""})
					m.Cmd(SYSTEM, "cp", "bin/boot.log", path.Join(back, "boot.log"), ice.Maps{CMD_OUTPUT: ""})
				}
			}
		}},
	})
}
