package cli

import (
	"bytes"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
)

func _runtime_init(m *ice.Message) {
	count := kit.Int(m.Conf(RUNTIME, kit.Keys(BOOT, mdb.COUNT)))
	kit.For(kit.UnMarshal(kit.Format(ice.Info.Make)), func(k string, v ice.Any) { m.Conf(RUNTIME, kit.Keys(MAKE, strings.ToLower(k)), v) })
	m.Conf(RUNTIME, kit.Keys(HOST, GOARCH), runtime.GOARCH)
	m.Conf(RUNTIME, kit.Keys(HOST, GOOS), runtime.GOOS)
	m.Conf(RUNTIME, kit.Keys(HOST, OSID), release(m))
	m.Conf(RUNTIME, kit.Keys(HOST, PID), os.Getpid())
	m.Conf(RUNTIME, kit.Keys(HOST, PWD), kit.Path(""))
	m.Conf(RUNTIME, kit.Keys(HOST, HOME), kit.HomePath(""))
	m.Conf(RUNTIME, kit.Keys(HOST, MAXPROCS), runtime.GOMAXPROCS(0))
	kit.For(ENV_LIST, func(k string) {
		m.Conf(RUNTIME, kit.Keys(CONF, k), kit.Env(k))
		kit.If(k == CTX_PID, func() { ice.Info.PidPath = kit.Env(k) })
	})
	m.Conf(RUNTIME, kit.Keys(BOOT, HOSTNAME), kit.Env("HOSTNAME"))
	if name, e := os.Hostname(); e == nil && name != "" {
		m.Conf(RUNTIME, kit.Keys(BOOT, HOSTNAME), name)
	}
	m.Conf(RUNTIME, kit.Keys(BOOT, PATHNAME), path.Base(kit.Path("")))
	m.Conf(RUNTIME, kit.Keys(BOOT, USERNAME), kit.UserName())
	ice.Info.Hostname = m.Conf(RUNTIME, kit.Keys(BOOT, HOSTNAME))
	ice.Info.Pathname = m.Conf(RUNTIME, kit.Keys(BOOT, PATHNAME))
	ice.Info.Username = m.Conf(RUNTIME, kit.Keys(BOOT, USERNAME))
	ice.Info.System = m.Conf(RUNTIME, kit.Keys(HOST, OSID))
	aaa.UserRoot(ice.Pulse, "", ice.Info.Username, aaa.ROOT, ice.OPS)
	aaa.UserRoot(ice.Pulse, "", ice.Info.Make.Username, aaa.TECH, ice.DEV)
	if runtime.GOARCH != MIPSLE {
		msg := m.Cmd(nfs.DIR, _system_find(m, os.Args[0]), "time,path,size,hash")
		m.Conf(RUNTIME, kit.Keys(BOOT, mdb.TIME), msg.Append(mdb.TIME))
		m.Conf(RUNTIME, kit.Keys(BOOT, mdb.HASH), msg.Append(mdb.HASH))
		m.Conf(RUNTIME, kit.Keys(BOOT, nfs.SIZE), msg.Append(nfs.SIZE))
		m.Conf(RUNTIME, kit.Keys(BOOT, ice.BIN), msg.Append(nfs.PATH))
		ice.Info.Hash = msg.Append(mdb.HASH)
		ice.Info.Size = msg.Append(nfs.SIZE)
	}
	m.Conf(RUNTIME, kit.Keys(BOOT, mdb.COUNT), count+1)
	m.Conf(RUNTIME, mdb.META, "")
	m.Conf(RUNTIME, mdb.HASH, "")
}
func _runtime_hostinfo(m *ice.Message) {
	m.Push("time", ice.Info.Make.Time)
	m.Push("nCPU", runtime.NumCPU())
	m.Push("GOMAXPROCS", runtime.GOMAXPROCS(0))
	m.Push("NumGoroutine", runtime.NumGoroutine())
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	m.Push("Sys", kit.FmtSize(int64(stats.Sys)))
	m.Push("Alloc", kit.FmtSize(int64(stats.Alloc)))
	m.Push("Objects", stats.HeapObjects)
	m.Push("NumGC", stats.NumGC)
	m.Push("LastGC", time.Unix(int64(stats.LastGC)/int64(time.Second), int64(stats.LastGC)%int64(time.Second)))
	m.Push("uptime", kit.Split(m.Cmdx(SYSTEM, "uptime"), mdb.FS)[0])
	if runtime.GOOS == LINUX {
		for i, ls := range strings.Split(m.Cmdx(nfs.CAT, "/proc/meminfo"), lex.NL) {
			if vs := kit.Split(ls, ": "); len(vs) > 1 {
				if m.Push(strings.TrimSpace(vs[0]), kit.FmtSize(kit.Int64(strings.TrimSpace(vs[1]))*1024)); i > 1 {
					break
				}
			}
		}
	} else {
		m.Push("MemAvailable", "")
		m.Push("MemTotal", "")
		m.Push("MemFree", "")
	}
}
func _runtime_diskinfo(m *ice.Message) {
	m.Spawn().Split(kit.Replace(m.Cmdx(SYSTEM, "df", "-h"), "Mounted on", "Mountedon"), "", lex.SP, lex.NL).Table(func(index int, value ice.Maps, head []string) {
		kit.If(strings.HasPrefix(value["Filesystem"], "/dev"), func() { m.Push("", value, head) })
	})
	m.RenameAppend("%iused", "piused", "Use%", "Usep")
	ctx.DisplayStory(m, "pie.js?field=Size")
}

const (
	MAKE = "make"
	TEST = "test"
	HOST = "host"
	CONF = "conf"
	BOOT = "boot"
	NODE = "node"
)
const (
	GOARCH  = "GOARCH"
	AMD64   = "amd64"
	X86     = "386"
	ARM     = "arm"
	ARM64   = "arm64"
	MIPSLE  = "mipsle"
	GOOS    = "GOOS"
	LINUX   = "linux"
	MACOS   = "macos"
	DARWIN  = "darwin"
	WINDOWS = "windows"
)
const (
	PATH  = "PATH"
	HOME  = "HOME"
	USER  = "USER"
	TERM  = "TERM"
	SHELL = "SHELL"
)
const (
	CTX_SHY = "ctx_shy"
	CTX_COM = "ctx_com"
	CTX_DEV = "ctx_dev"
	CTX_OPS = "ctx_ops"
	CTX_ARG = "ctx_arg"
	CTX_PID = "ctx_pid"
	CTX_LOG = "ctx_log"
	CTX_POD = "ctx_pod"
	CTX_ENV = "ctx_env"

	CTX_DAEMON = "ctx_daemon"
)

var ENV_LIST = []string{TERM, SHELL, CTX_SHY, CTX_COM, CTX_DEV, CTX_OPS, CTX_ARG, CTX_PID, CTX_DAEMON}

const (
	HOSTNAME = "hostname"
	PATHNAME = "pathname"
	USERNAME = "username"
)
const (
	IFCONFIG = "ifconfig"
	DISKINFO = "diskinfo"
	HOSTINFO = "hostinfo"
	USERINFO = "userinfo"
	PROCSTAT = "procstat"
	PROCINFO = "procinfo"
	PROCKILL = "prockill"
	BOOTINFO = "bootinfo"
	MAXPROCS = "maxprocs"
)
const RUNTIME = "runtime"

func init() {
	Index.MergeCommands(ice.Commands{
		RUNTIME: {Name: "runtime info=bootinfo,ifconfig,diskinfo,hostinfo,userinfo,procstat,procinfo,bootinfo,role,api,cli,cmd,mod,env,path,chain,routine auto upgrade restart", Help: "运行环境", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				aaa.White(m, ice.ETC_PATH)
				aaa.White(m, ice.LICENSE)
				_runtime_init(m)
			}},
			IFCONFIG: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(tcp.HOST) }},
			DISKINFO: {Hand: func(m *ice.Message, arg ...string) { _runtime_diskinfo(m) }},
			HOSTINFO: {Hand: func(m *ice.Message, arg ...string) { _runtime_hostinfo(m) }},
			HOSTNAME: {Hand: func(m *ice.Message, arg ...string) {
				if len(arg) > 0 {
					ice.Info.Hostname = mdb.Conf(m, RUNTIME, kit.Keys(NODE, mdb.NAME), mdb.Conf(m, RUNTIME, kit.Keys(BOOT, HOSTNAME), arg[0]))
				}
				m.Echo(ice.Info.Hostname)
			}},
			USERINFO: {Hand: func(m *ice.Message, arg ...string) { m.Split(m.Cmdx(SYSTEM, "who"), "user term time") }},
			PROCSTAT: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(PROCSTAT) }},
			PROCINFO: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(PROCINFO) }},
			PROCKILL: {Help: "结束进程", Hand: func(m *ice.Message, arg ...string) { m.Cmdy(gdb.SIGNAL, gdb.STOP, m.Option("PID")).ProcessRefresh() }},
			MAXPROCS: {Hand: func(m *ice.Message, arg ...string) {
				kit.If(len(arg) > 0, func() { runtime.GOMAXPROCS(kit.Int(mdb.Conf(m, RUNTIME, kit.Keys(HOST, MAXPROCS), arg[0]))) })
				m.Echo("%d", runtime.GOMAXPROCS(0))
			}},
			API: {Hand: func(m *ice.Message, arg ...string) {
				if len(arg) > 1 {
					m.Cmdy(ctx.COMMAND, "web.code.inner").Push(ctx.ARGS, kit.Format(nfs.SplitPath(m, m.Option(nfs.FILE))))
					return
				}
				ctx.DisplayStorySpide(m.Options(nfs.DIR_ROOT, nfs.PS), lex.PREFIX, kit.Fields(ctx.ACTION, m.ActionKey()))
				kit.For(ice.Info.Route, func(k, v string) { m.Push(nfs.PATH, k).Push(nfs.FILE, v) })
				m.Sort(nfs.PATH).StatusTimeCount()
			}},
			CLI: {Hand: func(m *ice.Message, arg ...string) {
				if len(arg) > 1 {
					m.Cmdy(ctx.COMMAND, "web.code.inner").Push(ctx.ARGS, kit.Format(nfs.SplitPath(m, m.Option(nfs.FILE))))
					return
				}
				ctx.DisplayStorySpide(m.Options(nfs.DIR_ROOT, "ice."), lex.PREFIX, kit.Fields(ctx.ACTION, m.ActionKey()), mdb.FIELD, mdb.NAME, lex.SPLIT, nfs.PT)
				kit.For(ice.Info.File, func(k, v string) { m.Push(nfs.FILE, k).Push(mdb.NAME, v) })
				m.Sort(mdb.NAME).StatusTimeCount()
			}},
			CMD: {Hand: func(m *ice.Message, arg ...string) {
				m.OptionFields(ctx.INDEX, mdb.NAME, mdb.HELP, nfs.FILE)
				m.Cmdy(ctx.COMMAND, mdb.SEARCH, ctx.COMMAND).StatusTimeCount()
			}},
			"mod": {Hand: func(m *ice.Message, arg ...string) {
				kit.For(ice.Info.Gomod, func(k string, v string) { m.Push(nfs.MODULE, k).Push(nfs.VERSION, v) })
				m.StatusTimeCount()
			}},
			"routine": {Hand: func(m *ice.Message, arg ...string) {
				status := map[string]int{}
				buf := make([]byte, 4096*4096)
				runtime.Stack(buf, true)
			outer:
				for _, v := range bytes.Split(buf, []byte(lex.NL+lex.NL)) {
					ls := bytes.Split(v, []byte(lex.NL))
					if ls := strings.SplitN(string(ls[0]), " ", 3); len(ls) > 0 {
						m.Push(mdb.ID, ls[1]).Push("status", ls[2])
						status[kit.Split(string(ls[2]), " []:")[0]]++
					}
					for i := 1; i < len(ls); i += 2 {
						if bytes.HasPrefix(ls[i], []byte("shylinux.com/x/")) {
							m.Push(nfs.PATH, kit.TrimPath(string(ls[i+1]))).Push("func", string(ls[i]))
							continue outer
						}
					}
					m.Push(nfs.PATH, kit.TrimPath(string(ls[2]))).Push("func", string(ls[1]))
				}
				var stats runtime.MemStats
				runtime.ReadMemStats(&stats)
				m.StatusTimeCount(status, "GOMAXPROCS", runtime.GOMAXPROCS(0), "NumGC", stats.NumGC, "Alloc", kit.FmtSize(int64(stats.Alloc)), "Sys", kit.FmtSize(int64(stats.Sys)))
				m.Echo("%v", string(buf))
			}},
			ENV: {Hand: func(m *ice.Message, arg ...string) {
				kit.For(os.Environ(), func(v string) {
					ls := strings.SplitN(v, mdb.EQ, 2)
					m.Push(mdb.NAME, ls[0]).Push(mdb.VALUE, ls[1])
				})
				m.StatusTimeCount().Sort(mdb.NAME)
			}},
			nfs.PATH: {Hand: func(m *ice.Message, arg ...string) {
				kit.For(_path_split(os.Getenv(PATH)), func(p string) { m.Push(nfs.PATH, p) })
			}},
			"chain":   {Hand: func(m *ice.Message, arg ...string) { m.Echo(m.FormatChain()) }},
			"upgrade": {Hand: func(m *ice.Message, arg ...string) { m.Cmdy("web.code.upgrade") }},
			RESTART:   {Hand: func(m *ice.Message, arg ...string) { m.Cmd(ice.EXIT, 1) }},
			aaa.ROLE: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(aaa.ROLE, func(value ice.Maps) { m.Push(mdb.KEY, kit.Keys(value[aaa.ROLE], value[mdb.ZONE], value[mdb.KEY])) })
				ctx.DisplayStorySpide(m.Options(nfs.DIR_ROOT, "ice."), mdb.FIELD, mdb.KEY, lex.SPLIT, nfs.PT)
			}},
		}, ctx.CmdAction(), ctx.ConfAction("")), Hand: func(m *ice.Message, arg ...string) {
			kit.If(len(arg) > 0 && arg[0] == BOOTINFO, func() { arg = arg[1:] })
			m.Cmdy(ctx.CONFIG, RUNTIME, arg)
			ctx.DisplayStoryJSON(m)
			m.Status(mdb.TIME, ice.Info.Make.Time,
				mdb.HASH, kit.Cut(ice.Info.Hash, 6), nfs.SIZE, ice.Info.Size,
				ice.BIN, _system_find(m, os.Args[0]), mdb.NAME, ice.Info.NodeName,
				nfs.REMOTE, ice.Info.Make.Remote, nfs.VERSION, ice.Info.Make.Versions(),
			)
		}},
	})
}
func NodeInfo(m *ice.Message, arg ...string) {
	mdb.Conf(m, RUNTIME, kit.Keys(NODE, mdb.TIME), m.Time())
	ice.Info.NodeName = mdb.Conf(m, RUNTIME, kit.Keys(NODE, mdb.NAME), kit.Select(ice.Info.NodeName, arg, 0))
	ice.Info.NodeType = mdb.Conf(m, RUNTIME, kit.Keys(NODE, mdb.TYPE), kit.Select(ice.Info.NodeType, arg, 1))
}
func IsWindows() bool { return runtime.GOOS == WINDOWS }
