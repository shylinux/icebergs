package code

import (
	"path"
	"runtime"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"

	// "shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _compile_target(m *ice.Message, arg ...string) (string, string, string, string) {
	main, file, goos, arch := ice.SRC_MAIN_GO, "", runtime.GOOS, runtime.GOARCH
	for _, k := range arg {
		switch k {
		case cli.AMD64, cli.X86, cli.MIPSLE, cli.ARM, cli.ARM64:
			arch = k
		case cli.LINUX, cli.DARWIN, cli.WINDOWS:
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
		file = path.Join(ice.USR_PUBLISH, kit.Keys(kit.Select(ice.ICE, kit.TrimExt(main), main != ice.SRC_MAIN_GO), goos, arch))
	}
	return main, file, goos, arch
}

const (
	RELAY = "relay"
)
const COMPILE = "compile"

func init() {
	Index.Merge(&ice.Context{Configs: ice.Configs{
		COMPILE: {Value: kit.Data(cli.ENV, kit.Dict("GOPRIVATE", "shylinux.com,github.com", "GOPROXY", "https://goproxy.cn,direct", "CGO_ENABLED", "0"))},
	}, Commands: ice.Commands{
		COMPILE: {Name: "compile arch=amd64,386,mipsle,arm,arm64 os=linux,darwin,windows src=src/main.go@key run binpack relay", Help: "编译", Actions: ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				cli.IsAlpine(m, "curl")
				cli.IsAlpine(m, "make")
				cli.IsAlpine(m, "gcc")
				cli.IsAlpine(m, "vim")
				cli.IsAlpine(m, "tmux")
				if cli.IsAlpine(m, "git"); !cli.IsAlpine(m, "go", "go git") {
					m.Cmd(mdb.INSERT, cli.CLI, "", mdb.ZONE, cli.CLI, "go", cli.CMD, kit.Format("install download https://golang.google.cn/dl/go1.15.5.%s-%s.tar.gz usr/local", runtime.GOOS, runtime.GOARCH))
				}
			}},
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(nfs.DIR, ice.SRC, nfs.DIR_CLI_FIELDS, kit.Dict(nfs.DIR_REG, `.*\.go$`)).Sort(nfs.PATH)
			}},
			BINPACK: {Help: "打包", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(AUTOGEN, BINPACK)
			}},
			RELAY: {Help: "跳板", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(COMPILE, ice.SRC_RELAY_GO, path.Join(ice.USR_PUBLISH, RELAY))
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			_autogen_version(m.Spawn())

			// 执行编译
			// web.PushStream(m)
			main, file, goos, arch := _compile_target(m, arg...)
			m.Optionv(cli.CMD_ENV, kit.Simple(cli.HOME, kit.Env(cli.HOME), cli.PATH, kit.Env(cli.PATH), m.Configv(cli.ENV), m.Optionv(cli.ENV), cli.GOOS, goos, cli.GOARCH, arch))
			// m.Cmd(cli.SYSTEM, GO, "get", "shylinux.com/x/ice")
			if msg := m.Cmd(cli.SYSTEM, GO, cli.BUILD, "-o", file, main, ice.SRC_VERSION_GO, ice.SRC_BINPACK_GO); !cli.IsSuccess(msg) {
				m.Copy(msg)
				return
			}
			m.Option(cli.CMD_OUTPUT, "")

			// 编译成功
			m.Logs(mdb.EXPORT, nfs.SOURCE, main, nfs.TARGET, file)
			m.Cmdy(nfs.DIR, file, nfs.DIR_WEB_FIELDS)
			m.Cmdy(PUBLISH, ice.CONTEXTS)
			m.StatusTimeCount()
			m.Process("")
		}},
	}})
}
