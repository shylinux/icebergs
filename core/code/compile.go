package code

import (
	"os"
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
)

const COMPILE = "compile"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		COMPILE: {Name: COMPILE, Help: "编译", Value: kit.Data(
			nfs.PATH, ice.USR_PUBLISH, cli.ENV, kit.Dict(
				"GOPRIVATE", "shylinux.com,github.com", "GOPROXY", "https://goproxy.cn,direct",
				cli.HOME, os.Getenv(cli.HOME), cli.PATH, os.Getenv(cli.PATH),
				"GOCACHE", os.Getenv("GOCACHE"), "CGO_ENABLED", "0",
			), GO, kit.List(GO, cli.BUILD),
		)},
	}, Commands: map[string]*ice.Command{
		COMPILE: {Name: "compile arch=amd64,386,arm os=linux,darwin,windows src=src/main.go@key run:button", Help: "编译", Action: map[string]*ice.Action{
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(nfs.DIR, ice.SRC, "path,size,time", ice.Option{nfs.DIR_REG, `.*\.go$`})
				m.Sort(kit.MDB_PATH)
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				m.Cmdy(nfs.DIR, m.Config(nfs.PATH))
				return
			}

			// 交叉编译
			main, file := ice.SRC_MAIN_GO, ""
			goos := m.Conf(cli.RUNTIME, kit.Keys(tcp.HOST, cli.GOOS))
			arch := m.Conf(cli.RUNTIME, kit.Keys(tcp.HOST, cli.GOARCH))
			for _, k := range arg {
				switch k {
				case cli.LINUX, cli.DARWIN, cli.WINDOWS:
					goos = k
				case cli.AMD64, cli.I386, cli.ARM:
					arch = k
				default:
					if kit.Ext(k) == GO {
						main = k
					} else {
						file = k
					}
				}
			}
			if file == "" {
				file = path.Join(kit.Select("", m.Config(nfs.PATH), m.Option(cli.CMD_DIR) == ""),
					kit.Keys(kit.Select(ice.ICE, kit.TrimExt(main), main != ice.SRC_MAIN_GO), goos, arch))
			}

			// 执行编译
			_autogen_version(m.Spawn())
			m.Optionv(cli.CMD_ENV, kit.Simple(m.Configv(cli.ENV), cli.GOOS, goos, cli.GOARCH, arch))
			if msg := m.Cmd(cli.SYSTEM, m.Configv(GO), "-o", file, main, ice.SRC_VERSION_GO, ice.SRC_BINPACK_GO); !cli.IsSuccess(msg) {
				m.Copy(msg)
				return
			}

			// 编译成功
			m.Log_EXPORT(cli.SOURCE, main, cli.TARGET, file)
			m.Cmdy(nfs.DIR, file, "time,path,size,link,action")
			m.Cmdy(PUBLISH, ice.CONTEXTS, ice.BASE)
			m.StatusTimeCount()
		}},
	}})
}
