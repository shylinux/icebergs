package cli

import (
	"os"
	"runtime"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/logs"
)

func _path_sep() string {
	return kit.Select(":", ";", strings.Contains(os.Getenv(PATH), ";"))
}
func BinPath(arg ...string) string {
	return kit.Join(kit.Simple(arg, kit.Path(""), kit.Path(ice.BIN), kit.Path(ice.USR_PUBLISH), kit.Path(ice.USR_LOCAL_BIN), kit.Path(ice.USR_LOCAL_GO_BIN), kit.Env(PATH)), _path_sep())
}

const FOREVER = "forever"

func init() {
	Index.MergeCommands(ice.Commands{
		FOREVER: {Name: "forever auto", Help: "启动", Actions: ice.Actions{
			START: {Hand: func(m *ice.Message, arg ...string) {
				if runtime.GOOS == WINDOWS {
					m.Cmdy("serve", "start", arg)
					return
				}
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
				m.Options(CMD_ENV, env, CMD_INPUT, os.Stdin, CMD_OUTPUT, os.Stdout, CMD_ERRPUT, os.Stderr)
				if p := kit.Env(CTX_LOG); p != "" {
					m.Optionv(CMD_ERRPUT, p)
				}
				m.Cmd(FOREVER, STOP)
				if bin := kit.Select(os.Args[0], ice.BIN_ICE_BIN, nfs.ExistsFile(m, ice.BIN_ICE_BIN)); len(arg) > 0 && arg[0] == ice.SPACE {
					m.Cmdy(FOREVER, bin, ice.SPACE, "dial", ice.DEV, ice.OPS, arg[1:])
				} else {
					if len(arg) == 0 || arg[0] != ice.DEV {
						arg = append([]string{ice.DEV, ""}, arg...)
					}
					m.Cmdy(FOREVER, bin, ice.SERVE, START, arg)
				}
			}},
			RESTART: {Hand: func(m *ice.Message, arg ...string) {
				if runtime.GOOS == WINDOWS {
					return
				}
				m.Cmd(gdb.SIGNAL, gdb.RESTART)
			}},
			STOP:  {Hand: func(m *ice.Message, arg ...string) { m.Cmd(gdb.SIGNAL, gdb.STOP) }},
			DELAY: {Hand: func(m *ice.Message, arg ...string) { m.Sleep(arg[0]).Cmdy(arg[1:]) }},
		}, Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				m.Cmdy(RUNTIME, BOOTINFO)
				return
			}
			for {
				if logs.Println("run %s", kit.Join(arg, ice.SP)); IsSuccess(m.Cmd(SYSTEM, arg)) {
					logs.Println("what %v", 123)
					logs.Println(ice.EXIT)
					break
				}
				logs.Println("what %v", 123)
				m.Debug("what %v", arg)
				m.Sleep("1s")
				if logs.Println(); m.Config("log.save") == ice.TRUE {
					back := kit.Format("var/log.%s", logs.Now().Format("20060102_150405"))
					m.Cmd(SYSTEM, "cp", "-r", "var/log", back, ice.Maps{CMD_OUTPUT: ""})
				}
			}
		}},
	})
}
