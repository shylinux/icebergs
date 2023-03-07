package code

import (
	"path"
	"runtime"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
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
		file = path.Join(ice.USR_PUBLISH, kit.Keys(kit.Select(ice.ICE, kit.TrimExt(main, GO), main != ice.SRC_MAIN_GO), goos, arch))
	}
	return main, file, goos, arch
}

const (
	VERSION = "version"
)
const COMPILE = "compile"

func init() {
	Index.MergeCommands(ice.Commands{
		COMPILE: {Name: "compile arch=amd64,386,arm,arm64,mipsle os=linux,darwin,windows src=src/main.go@key run binpack webpack devpack upgrade", Help: "编译", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) { cli.IsAlpine(m, GO, "go git") }},
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case VERSION:
					m.Push(arg[0], "1.15.15")
				default:
					m.Cmdy(nfs.DIR, ice.SRC, nfs.DIR_CLI_FIELDS, kit.Dict(nfs.DIR_REG, kit.ExtReg(GO)))
				}
			}},
			BINPACK: {Help: "版本", Hand: func(m *ice.Message, arg ...string) { m.Cmdy(AUTOGEN, BINPACK) }},
			WEBPACK: {Help: "打包", Hand: func(m *ice.Message, arg ...string) { m.Cmdy(AUTOGEN, WEBPACK) }},
			DEVPACK: {Help: "开发", Hand: func(m *ice.Message, arg ...string) { m.Cmdy(AUTOGEN, DEVPACK) }},
			UPGRADE: {Help: "升级", Hand: func(m *ice.Message, arg ...string) { m.Cmdy(UPGRADE, nfs.TARGET) }},
			INSTALL: {Name: "install version=1.15.15", Help: "安装", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(INSTALL, web.DOWNLOAD, kit.Format("https://golang.google.cn/dl/go%s.%s-%s.%s", m.Option(VERSION), runtime.GOOS, runtime.GOARCH, kit.Select("tar.gz", "zip", runtime.GOOS == cli.WINDOWS)), ice.USR_LOCAL)
			}},
		}, ctx.ConfAction(cli.ENV, kit.Dict("GOPRIVATE", "shylinux.com,github.com", "GOPROXY", "https://goproxy.cn,direct", "CGO_ENABLED", "0"))), Hand: func(m *ice.Message, arg ...string) {
			main, file, goos, arch := _compile_target(m, arg...)
			env := kit.Simple(cli.PATH, kit.Env(cli.PATH), cli.HOME, kit.Select(kit.Path(""), kit.Env(cli.HOME)), m.Configv(cli.ENV), m.Optionv(cli.ENV), cli.GOOS, goos, cli.GOARCH, arch)
			if runtime.GOOS == cli.WINDOWS {
				env = append(env, "GOPATH", kit.HomePath(GO), "GOCACHE", kit.HomePath("go/go-build"))
			}
			m.Optionv(cli.CMD_ENV, env)
			if !strings.Contains(m.Cmdx(nfs.CAT, ice.GO_MOD), "shylinux.com/x/ice") {
				m.Cmd(cli.SYSTEM, GO, "get", "shylinux.com/x/ice")
			}
			m.Cmd(AUTOGEN, VERSION)
			defer m.StatusTime(VERSION, strings.TrimPrefix(m.Cmdx(cli.SYSTEM, GO, VERSION), "go version "))
			if msg := m.Cmd(cli.SYSTEM, GO, cli.BUILD, "-o", file, main, ice.SRC_VERSION_GO, ice.SRC_BINPACK_GO); !cli.IsSuccess(msg) {
				m.Copy(msg)
				return
			}
			m.Logs(COMPILE, nfs.TARGET, file, nfs.SOURCE, main)
			if m.Cmdy(nfs.DIR, file, "time,path,size,hash,link"); strings.Contains(file, ice.ICE) {
				m.Cmdy(PUBLISH, ice.CONTEXTS)
			}
		}},
	})
}
