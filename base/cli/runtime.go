package cli

import (
	"os"
	"path"
	"runtime"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _runtime_init(m *ice.Message) {
	kit.Fetch(kit.UnMarshal(kit.Format(ice.Info.Make)), func(key string, value ice.Any) {
		m.Conf(RUNTIME, kit.Keys(MAKE, strings.ToLower(key)), value)
	})
	m.Conf(RUNTIME, kit.Keys(HOST, GOARCH), runtime.GOARCH)
	m.Conf(RUNTIME, kit.Keys(HOST, GOOS), runtime.GOOS)
	m.Conf(RUNTIME, kit.Keys(HOST, OSID), release(m))
	m.Conf(RUNTIME, kit.Keys(HOST, PID), os.Getpid())
	m.Conf(RUNTIME, kit.Keys(HOST, PWD), kit.Path(""))
	m.Conf(RUNTIME, kit.Keys(HOST, HOME), kit.Env(HOME))
	m.Conf(RUNTIME, kit.Keys(HOST, MAXPROCS), runtime.GOMAXPROCS(0))
	for _, k := range ENV_LIST {
		switch m.Conf(RUNTIME, kit.Keys(CONF, k), kit.Env(k)); k {
		case CTX_PID:
			ice.Info.PidPath = kit.Select(path.Join(ice.VAR_RUN, "ice.pid"), kit.Env(k))
		case CTX_SHARE:
			ice.Info.CtxShare = kit.Env(k)
		case CTX_RIVER:
			ice.Info.CtxRiver = kit.Env(k)
		}
	}
	m.Conf(RUNTIME, kit.Keys(BOOT, HOSTNAME), kit.Env("HOSTNAME"))
	if name, e := os.Hostname(); e == nil && name != "" {
		m.Conf(RUNTIME, kit.Keys(BOOT, HOSTNAME), name)
	}
	m.Conf(RUNTIME, kit.Keys(BOOT, PATHNAME), path.Base(kit.Env("PWD")))
	if name, e := os.Getwd(); e == nil && name != "" {
		m.Conf(RUNTIME, kit.Keys(BOOT, PATHNAME), path.Base(name))
	}
	m.Conf(RUNTIME, kit.Keys(BOOT, USERNAME), kit.Select(kit.UserName(), kit.Env(CTX_USER)))
	ice.Info.HostName = m.Conf(RUNTIME, kit.Keys(BOOT, HOSTNAME))
	ice.Info.PathName = m.Conf(RUNTIME, kit.Keys(BOOT, PATHNAME))
	ice.Info.UserName = m.Conf(RUNTIME, kit.Keys(BOOT, USERNAME))
	aaa.UserRoot(ice.Pulse, ice.Info.UserName)
	bin := _system_find(m, os.Args[0])
	m.Conf(RUNTIME, kit.Keys(BOOT, ice.BIN), bin)
	m.Conf(RUNTIME, kit.Keys(BOOT, mdb.TIME), m.Time())
	if s, e := nfs.StatFile(m, bin); e == nil {
		m.Conf(RUNTIME, kit.Keys(BOOT, nfs.SIZE), kit.FmtSize(s.Size()))
		if f, e := nfs.OpenFile(m, bin); e == nil {
			defer f.Close()
			m.Conf(RUNTIME, kit.Keys(BOOT, mdb.HASH), kit.Hashs(f))
		}
	}
	m.Conf(RUNTIME, mdb.META, "")
	m.Conf(RUNTIME, mdb.HASH, "")
	m.Conf(RUNTIME, kit.Keys(BOOT, mdb.COUNT), kit.Int(m.Conf(RUNTIME, kit.Keys(BOOT, mdb.COUNT)))+1)
}
func _runtime_hostinfo(m *ice.Message) {
	m.Push("nCPU", strings.Count(m.Cmdx(nfs.CAT, "/proc/cpuinfo"), "processor"))
	for i, ls := range strings.Split(m.Cmdx(nfs.CAT, "/proc/meminfo"), ice.NL) {
		vs := kit.Split(ls, ": ")
		if m.Push(strings.TrimSpace(vs[0]), kit.FmtSize(kit.Int64(strings.TrimSpace(vs[1]))*1024)); i > 1 {
			break
		}
	}
	m.Push("uptime", kit.Split(m.Cmdx(SYSTEM, "uptime"), ice.FS)[0])
}
func _runtime_diskinfo(m *ice.Message) {
	m.Spawn().Split(kit.Replace(m.Cmdx(SYSTEM, "df", "-h"), "Mounted on", "Mountedon"), "", ice.SP, ice.NL).Table(func(index int, value ice.Maps, head []string) {
		if strings.HasPrefix(value["Filesystem"], "/dev") {
			m.Push("", value, head)
		}
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
	GOARCH = "GOARCH"
	AMD64  = "amd64"
	X86    = "386"
	ARM    = "arm"
	ARM64  = "arm64"
	MIPSLE = "mipsle"

	GOOS    = "GOOS"
	LINUX   = "linux"
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
	CTX_POD = "ctx_pod"
	CTX_ARG = "ctx_arg"
	CTX_ENV = "ctx_env"
	CTX_PID = "ctx_pid"
	CTX_LOG = "ctx_log"

	CTX_USER   = "ctx_user"
	CTX_SHARE  = "ctx_share"
	CTX_RIVER  = "ctx_river"
	CTX_DAEMON = "ctx_daemon"

	MAKE_DOMAIN = "make.domain"
)

var ENV_LIST = []string{
	TERM, SHELL, CTX_SHY, CTX_COM, CTX_DEV, CTX_OPS, CTX_ARG, CTX_PID, CTX_USER, CTX_SHARE, CTX_RIVER, CTX_DAEMON,
}

const (
	HOSTNAME = "hostname"
	PATHNAME = "pathname"
	USERNAME = "username"
)
const (
	IFCONFIG = "ifconfig"
	HOSTINFO = "hostinfo"
	USERINFO = "userinfo"
	PROCINFO = "procinfo"
	PROCKILL = "prockill"
	DISKINFO = "diskinfo"
	BOOTINFO = "bootinfo"
	MAXPROCS = "maxprocs"
)
const RUNTIME = "runtime"

func init() {
	Index.MergeCommands(ice.Commands{
		RUNTIME: {Name: "runtime info=ifconfig,hostinfo,hostname,userinfo,procinfo,diskinfo,bootinfo,api,cli,cmd,env auto", Help: "运行环境", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) { _runtime_init(m) }},
			IFCONFIG:     {Hand: func(m *ice.Message, arg ...string) { m.Cmdy("tcp.host") }},
			HOSTINFO:     {Hand: func(m *ice.Message, arg ...string) { _runtime_hostinfo(m) }},
			HOSTNAME: {Hand: func(m *ice.Message, arg ...string) {
				if len(arg) > 0 {
					ice.Info.HostName = m.Conf(RUNTIME, kit.Keys(BOOT, HOSTNAME), m.Conf(RUNTIME, kit.Keys(NODE, mdb.NAME), arg[0]))
				}
				m.Echo(ice.Info.HostName)
			}},
			USERINFO: {Hand: func(m *ice.Message, arg ...string) { m.Split(m.Cmdx(SYSTEM, "who"), "user term time") }},
			PROCINFO: {Hand: func(m *ice.Message, arg ...string) {
				msg := m.Cmd("", HOSTINFO)
				m.Split(m.Cmdx(SYSTEM, "ps", "u")).PushAction(PROCKILL).SortIntR("RSS")
				m.StatusTimeCount("nCPU", msg.Append("nCPU"), "MemTotal", msg.Append("MemTotal"), "MemFree", msg.Append("MemFree"))
			}},
			PROCKILL: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(gdb.SIGNAL, gdb.STOP, m.Option("PID")).ProcessRefresh() }},
			MAXPROCS: {Hand: func(m *ice.Message, arg ...string) {
				if len(arg) > 0 {
					runtime.GOMAXPROCS(kit.Int(m.Conf(RUNTIME, kit.Keys(HOST, MAXPROCS), arg[0])))
				}
				m.Echo("%d", runtime.GOMAXPROCS(0))
			}},
			DISKINFO: {Hand: func(m *ice.Message, arg ...string) { _runtime_diskinfo(m) }},
			API: {Hand: func(m *ice.Message, arg ...string) {
				for k, v := range ice.Info.Route {
					m.Push(nfs.PATH, k).Push(nfs.FILE, v)
				}
				m.Sort(nfs.PATH).StatusTimeCount()
			}},
			CLI: {Hand: func(m *ice.Message, arg ...string) {
				for k, v := range ice.Info.File {
					m.Push(nfs.FILE, k).Push(mdb.NAME, v)
				}
				m.Sort(nfs.FILE).StatusTimeCount()
			}},
			CMD: {Hand: func(m *ice.Message, arg ...string) {
				m.OptionFields(ctx.INDEX, mdb.NAME, mdb.HELP, nfs.FILE)
				m.Cmdy(ctx.COMMAND, mdb.SEARCH, ctx.COMMAND).StatusTimeCount()
			}},
			ENV: {Hand: func(m *ice.Message, arg ...string) {
				for _, v := range os.Environ() {
					ls := strings.SplitN(v, ice.EQ, 2)
					m.Push(mdb.NAME, ls[0]).Push(mdb.VALUE, ls[1])
				}
				m.StatusTimeCount()
			}},
			MAKE_DOMAIN: {Hand: func(m *ice.Message, arg ...string) {
				if os.Getenv(CTX_DEV) == "" || os.Getenv(CTX_POD) == "" {
					m.Echo(m.Conf(RUNTIME, MAKE_DOMAIN))
				} else {
					m.Echo(kit.MergePOD(os.Getenv(CTX_DEV), os.Getenv(CTX_POD)))
				}
			}},
		}, ctx.ConfAction("hi")), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) > 0 && arg[0] == BOOTINFO {
				arg = arg[1:]
			}
			m.Cmdy(ctx.CONFIG, RUNTIME, arg)
			ctx.DisplayStoryJSON(m)
		}},
	})
}
func NodeInfo(m *ice.Message, kind, name string) {
	m.Conf(RUNTIME, kit.Keys(NODE, mdb.TIME), m.Time())
	ice.Info.NodeType = m.Conf(RUNTIME, kit.Keys(NODE, mdb.TYPE), kind)
	ice.Info.NodeName = m.Conf(RUNTIME, kit.Keys(NODE, mdb.NAME), strings.ReplaceAll(name, ice.PT, "_"))
}
