package code

import (
	"path"
	"runtime"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _compile_target(m *ice.Message, arg ...string) (string, string, string, string) {
	main, file, goos, arch := ice.SRC_MAIN_GO, "", runtime.GOOS, runtime.GOARCH
	for _, k := range arg {
		switch k {
		case cli.AMD64, cli.X86, cli.ARM, cli.ARM64, cli.MIPSLE:
			arch = k
		case cli.LINUX, cli.DARWIN, cli.WINDOWS:
			goos = k
		default:
			kit.If(kit.Ext(k) == GO, func() { main = k }, func() { file = k })
		}
	}
	if file == "" {
		file = path.Join(ice.USR_PUBLISH, kit.Keys(kit.Select(ice.ICE, kit.TrimExt(main, GO), main != ice.SRC_MAIN_GO), goos, arch))
	}
	return main, file, goos, arch
}

func _compile_get(m *ice.Message, main string) {
	block, list := false, []string{}
	m.Cmd(lex.SPLIT, main, func(ls []string) {
		switch ls[0] {
		case IMPORT:
			if ls[1] == "(" {
				block = true
			} else {
				list = append(list, ls[1])
			}
		case ")":
			block = false
		default:
			if block {
				list = append(list, kit.Select("", ls, -1))
			}
		}
	})
	kit.For(list, func(p string) {
		_list := _compile_mod(m)
		if _, ok := _list[p]; ok {
			return
		} else if ls := kit.Slice(strings.Split(p, nfs.PS), 0, 3); _list[path.Join(ls...)] {
			return
		}
		GoGet(m, p)
	})
}
func _compile_mod(m *ice.Message) map[string]bool {
	block, list := false, map[string]bool{}
	m.Cmd(lex.SPLIT, ice.GO_MOD, func(ls []string) {
		switch ls[0] {
		case MODULE:
			list[ls[1]] = true
		case REQUIRE:
			if ls[1] == "(" {
				block = true
			} else {
				list[ls[1]] = true
			}
		case ")":
			block = false
		default:
			kit.If(block, func() { list[kit.Select("", ls, 0)] = true })
		}
	})
	return list
}

const COMPILE = "compile"

func init() {
	Index.MergeCommands(ice.Commands{
		COMPILE: {Name: "compile arch=amd64,386,arm,arm64,mipsle os=linux,darwin,windows src=src/main.go@key run binpack webpack devpack install", Icon: "go.png", Help: "编译", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) { cli.IsAlpine(m, GO, "go git") }},
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case SERVICE:
					m.Push(arg[0], kit.MergeURL2(m.Cmd(web.SPIDE, ice.DEV).Append(web.CLIENT_ORIGIN), "/publish/"))
				case VERSION:
					m.Push(arg[0], "1.13.5", "1.15.5", "1.17.3", "1.20.3")
				default:
					m.Cmdy(nfs.DIR, ice.SRC, nfs.DIR_CLI_FIELDS, kit.Dict(nfs.DIR_REG, kit.ExtReg(GO)))
				}
			}},
			BINPACK: {Help: "发布", Hand: func(m *ice.Message, arg ...string) { m.Cmdy(AUTOGEN, BINPACK) }},
			WEBPACK: {Help: "打包", Hand: func(m *ice.Message, arg ...string) { m.Cmdy(AUTOGEN, WEBPACK) }},
			DEVPACK: {Help: "开发", Hand: func(m *ice.Message, arg ...string) { m.Cmdy(AUTOGEN, DEVPACK) }},
			INSTALL: {Name: "install service*='https://golang.google.cn/dl/' version*=1.13.5", Help: "安装", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(INSTALL, web.DOWNLOAD, kit.Format("%s/go%s.%s-%s.%s", m.Option(SERVICE), m.Option(VERSION), runtime.GOOS, runtime.GOARCH, kit.Select("tar.gz", "zip", runtime.GOOS == cli.WINDOWS)), ice.USR_LOCAL)
			}},
			web.DREAM_TABLES: {Hand: func(m *ice.Message, arg ...string) {
				kit.If(m.Option(mdb.TYPE) == web.WORKER, func() {
					kit.If(nfs.Exists(m, path.Join(ice.USR_LOCAL_WORK, m.Option(mdb.NAME), ice.SRC_MAIN_GO)), func() {
						kit.If(cli.SystemFind(m, GO), func() { m.PushButton(kit.Dict(m.CommandKey(), "构建")) })
					})
				})
			}},
			web.DREAM_ACTION: {Hand: func(m *ice.Message, arg ...string) { web.DreamProcess(m, []string{}, arg...) }},
		}, ctx.ConfAction(cli.ENV, kit.Dict(GOPRIVATE, "shylinux.com,github.com", GOPROXY, "https://goproxy.cn,direct", CGO_ENABLED, "0"))), Hand: func(m *ice.Message, arg ...string) {
			defer web.ToastProcess(m)()
			main, file, goos, arch := _compile_target(m, arg...)
			env := kit.Simple(cli.PATH, cli.BinPath(), cli.HOME, kit.Select(kit.Path(""), kit.Env(cli.HOME)), mdb.Configv(m, cli.ENV), m.Optionv(cli.ENV), cli.GOOS, goos, cli.GOARCH, arch)
			kit.If(runtime.GOOS == cli.WINDOWS, func() { env = append(env, GOPATH, kit.HomePath(GO), GOCACHE, kit.HomePath("go/go-build")) })
			m.Options(cli.CMD_ENV, env).Cmd(AUTOGEN, VERSION)
			_compile_get(m, main)
			defer m.StatusTime(VERSION, strings.TrimPrefix(GoVersion(m), "go version"))
			if msg := GoBuild(m.Spawn(), file, main, ice.SRC_VERSION_GO, ice.SRC_BINPACK_GO); !cli.IsSuccess(msg) {
				m.Copy(msg)
			} else {
				m.Logs(nfs.SAVE, nfs.TARGET, file, nfs.SOURCE, main)
				m.Cmdy(nfs.DIR, file, "time,path,size,hash,link")
				kit.If(!m.IsCliUA() && strings.Contains(file, ice.ICE), func() {
					defer m.EchoScript(m.Cmdv(web.SPIDE, ice.DEMO, web.CLIENT_URL))
					defer m.EchoScript("docker run -p 20000:9020 -w /root -it alpine")
					m.Cmdy(PUBLISH, ice.CONTEXTS, ice.APP)
				})
			}
		}},
	})
}
