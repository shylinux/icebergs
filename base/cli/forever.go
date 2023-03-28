package cli

import (
	"os"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/logs"
)

func _path_sep() string { return kit.Select(ice.DF, ";", strings.Contains(os.Getenv(PATH), ";")) }
func BinPath(arg ...string) string {
	list := []string{}
	push := func(p string) {
		kit.If(kit.IndexOf(list, p) == -1, func() { list = append(list, kit.ReplaceAll(p, "\\", ice.PS)) })
	}
	kit.For(arg, func(p string) {
		list = append(list, kit.Path(p, ice.BIN), kit.Path(p, ice.USR_PUBLISH), kit.Path(p, ice.USR_LOCAL_BIN), kit.Path(p, ice.USR_LOCAL_GO_BIN))
		kit.For(kit.Reverse(strings.Split(ice.Pulse.Cmdx(nfs.CAT, kit.Path(p, ice.ETC_PATH)), ice.NL)), func(l string) {
			kit.If(strings.TrimSpace(l) != "" && !strings.HasPrefix(strings.TrimSpace(l), "#"), func() { push(kit.Path(p, l)) })
		})
	})
	kit.For(strings.Split(kit.Env(PATH), _path_sep()), func(p string) { push(p) })
	return kit.Join(list, _path_sep())
}

const FOREVER = "forever"

func init() {
	Index.MergeCommands(ice.Commands{
		FOREVER: {Name: "forever auto", Help: "启动", Actions: ice.Actions{
			START: {Hand: func(m *ice.Message, arg ...string) {
				env := []string{PATH, BinPath(""), HOME, kit.Select(kit.Path(""), os.Getenv(HOME))}
				kit.For(ENV_LIST, func(k string) { kit.If(kit.Env(k) != "", func() { env = append(env, k, kit.Env(k)) }) })
				kit.For(os.Environ(), func(v string) {
					if ls := kit.Split(v, ice.EQ, ice.EQ); kit.IndexOf(env, ls[0]) == -1 && len(ls) > 1 {
						env = append(env, ls[0], ls[1])
					}
				})
				m.Options(CMD_ENV, env, CMD_INPUT, os.Stdin, CMD_OUTPUT, os.Stdout, CMD_ERRPUT, os.Stderr)
				kit.If(kit.Env(CTX_LOG), func(p string) { m.Optionv(CMD_ERRPUT, p) })
				m.Cmd(FOREVER, STOP)
				if bin := kit.Select(os.Args[0], ice.BIN_ICE_BIN, nfs.Exists(m, ice.BIN_ICE_BIN)); len(arg) > 0 && arg[0] == ice.SPACE {
					m.Cmdy(FOREVER, bin, ice.SPACE, START, ice.DEV, ice.OPS, arg[1:])
				} else {
					kit.If(len(arg) == 0 || arg[0] != ice.DEV, func() { arg = append([]string{ice.DEV, ""}, arg...) })
					m.Cmdy(FOREVER, bin, ice.SERVE, START, arg)
				}
			}},
			RESTART: {Hand: func(m *ice.Message, arg ...string) { m.Cmd(gdb.SIGNAL, gdb.RESTART) }},
			STOP:    {Hand: func(m *ice.Message, arg ...string) { m.Cmd(gdb.SIGNAL, gdb.STOP) }},
			DELAY:   {Hand: func(m *ice.Message, arg ...string) { m.Sleep(arg[0]).Cmdy(arg[1:]) }},
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
				logs.Println()
			}
		}},
	})
}
