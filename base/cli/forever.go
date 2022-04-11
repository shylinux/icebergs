package cli

import (
	"os"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func BinPath(arg ...string) string {
	return kit.Join(kit.Simple(arg, kit.Path(ice.BIN), kit.Path(ice.USR_LOCAL_BIN), kit.Path(ice.USR_LOCAL_GO_BIN), kit.Env(PATH)), ice.DF)
}

const FOREVER = "forever"

func init() {
	const SERVE = "serve"
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		FOREVER: {Name: "forever", Help: "启动", Action: map[string]*ice.Action{
			SERVE: {Name: "serve", Help: "服务", Hand: func(m *ice.Message, arg ...string) {
				env := []string{PATH, BinPath(), HOME, kit.Select(kit.Path(""), os.Getenv(HOME))}
				for _, k := range []string{TERM, SHELL, CTX_SHY, CTX_DEV, CTX_OPS, CTX_ARG, CTX_PID, CTX_USER, CTX_SHARE, CTX_RIVER} {
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

				if p := m.Cmdx(nfs.CAT, m.Conf("gdb.signal", kit.Keym(nfs.PATH))); p != "" {
					m.Cmd(SYSTEM, "kill", "-s", "QUIT", p)
				}

				m.Cmdy(FOREVER, kit.Select(os.Args[0], nfs.PWD+ice.BIN_ICE_BIN, kit.FileExists(ice.BIN_ICE_BIN)),
					SERVE, START, ice.DEV, "", aaa.USERNAME, aaa.ROOT, aaa.PASSWORD, aaa.ROOT, arg)
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			for {
				println(kit.Format("%s run %s", kit.Now(), kit.Join(arg, ice.SP)))
				if m.Sleep("1s"); IsSuccess(m.Cmd(SYSTEM, arg)) {
					println(kit.Format("%s exit", kit.Now()))
					return
				}
			}
		}}},
	})
}
