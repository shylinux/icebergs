package code

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const (
	RELAY = "relay"
)
const COMPILE = "compile"

func init() {
	const GIT = "git"
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		COMPILE: {Name: COMPILE, Help: "编译", Value: kit.Data(nfs.PATH, ice.USR_PUBLISH,
			cli.ENV, kit.Dict("GOPROXY", "https://goproxy.cn,direct", "GOPRIVATE", "shylinux.com,github.com", "CGO_ENABLED", "0"),
		)},
	}, Commands: map[string]*ice.Command{
		COMPILE: {Name: "compile arch=amd64,386,arm,arm64 os=linux,darwin,windows src=src/main.go@key run binpack relay install", Help: "编译", Action: map[string]*ice.Action{
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(nfs.DIR, ice.SRC, nfs.DIR_CLI_FIELDS, kit.Dict(nfs.DIR_REG, `.*\.go$`)).Sort(nfs.PATH)
			}},
			INSTALL: {Name: "compile", Help: "安装", Hand: func(m *ice.Message, arg ...string) {
				if strings.Contains(m.Cmdx(cli.RUNTIME, kit.Keys(tcp.HOST, cli.OSID)), cli.ALPINE) {
					web.PushStream(m)
					m.Cmd(cli.SYSTEM, "apk", "add", GIT, GO)
					m.Cmd(cli.SYSTEM, GO, "get", "shylinux.com/x/ice")
					return
				}
				if m.Cmdx(cli.SYSTEM, nfs.FIND, GIT) == "" {
					m.Toast("please install git")
					m.Echo(ice.FALSE)
					return
				}
				m.Cmd(INSTALL, web.DOWNLOAD, "https://golang.google.cn/dl/go1.15.5.linux-amd64.tar.gz", ice.USR_LOCAL)
			}},
			BINPACK: {Name: "binpack", Help: "打包", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(AUTOGEN, BINPACK)
			}},
			RELAY: {Name: "relay", Help: "跳板", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(COMPILE, ice.SRC_RELAY_GO, path.Join(ice.USR_PUBLISH, RELAY))
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if m.Cmdx(cli.SYSTEM, nfs.FIND, GO) == "" && m.Cmdx(COMPILE, INSTALL) == ice.FALSE {
				return
			}
			m.Cmd(cli.SYSTEM, GO, "get", "shylinux.com/x/ice")

			// 交叉编译
			main, file := ice.SRC_MAIN_GO, ""
			goos := m.Conf(cli.RUNTIME, kit.Keys(tcp.HOST, cli.GOOS))
			arch := m.Conf(cli.RUNTIME, kit.Keys(tcp.HOST, cli.GOARCH))
			for _, k := range arg {
				switch k {
				case cli.X386, cli.AMD64, cli.ARM64, cli.ARM:
					arch = k
				case cli.WINDOWS, cli.DARWIN, cli.LINUX:
					goos = k
				default:
					if kit.Ext(k) == GO {
						main = k
					} else {
						file = k
					}
				}
			}
			if file == "" {
				file = path.Join(m.Config(nfs.PATH), kit.Keys(kit.Select(ice.ICE, kit.TrimExt(main), main != ice.SRC_MAIN_GO), goos, arch))
			}

			// 执行编译
			_autogen_version(m.Spawn())
			m.Optionv(cli.CMD_ENV, kit.Simple(m.Configv(cli.ENV), cli.HOME, kit.Env(cli.HOME), cli.PATH, kit.Env(cli.PATH), cli.GOOS, goos, cli.GOARCH, arch))
			if msg := m.Cmd(cli.SYSTEM, GO, cli.BUILD, "-o", file, main, ice.SRC_VERSION_GO, ice.SRC_BINPACK_GO); !cli.IsSuccess(msg) {
				m.Copy(msg)
				return
			}

			// 编译成功
			m.Log_EXPORT(nfs.SOURCE, main, nfs.TARGET, file)
			m.Cmdy(nfs.DIR, file, nfs.DIR_WEB_FIELDS)
			m.Cmdy(PUBLISH, ice.CONTEXTS)
			m.StatusTimeCount()
		}},
	}})
}
