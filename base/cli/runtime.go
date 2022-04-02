package cli

import (
	"os"
	"path"
	"runtime"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _runtime_init(m *ice.Message) {
	// 版本信息 make
	kit.Fetch(kit.UnMarshal(kit.Format(ice.Info.Make)), func(key string, value interface{}) {
		m.Conf(RUNTIME, kit.Keys(MAKE, strings.ToLower(key)), value)
	})

	// 主机信息 host
	m.Conf(RUNTIME, kit.Keys(HOST, GOARCH), runtime.GOARCH)
	m.Conf(RUNTIME, kit.Keys(HOST, GOOS), runtime.GOOS)
	m.Conf(RUNTIME, kit.Keys(HOST, PID), os.Getpid())
	m.Conf(RUNTIME, kit.Keys(HOST, HOME), kit.Env(HOME))
	osid := ""
	m.Cmd(nfs.CAT, "/etc/os-release", func(text string) {
		if ls := kit.Split(text, "="); len(ls) > 1 {
			switch ls[0] {
			case "ID", "ID_LIKE":
				osid = strings.TrimSpace(ls[1] + ice.SP + osid)
			}
		}
	})
	m.Conf(RUNTIME, kit.Keys(HOST, OSID), osid)

	// 启动信息 boot
	if name, e := os.Hostname(); e == nil {
		m.Conf(RUNTIME, kit.Keys(BOOT, HOSTNAME), kit.Select(name, kit.Env("HOSTNAME")))
	}
	if name, e := os.Getwd(); e == nil {
		name = path.Base(kit.Select(name, kit.Env("PWD")))
		name = kit.Slice(strings.Split(name, ice.PS), -1)[0]
		name = kit.Slice(strings.Split(name, "\\"), -1)[0]
		m.Conf(RUNTIME, kit.Keys(BOOT, PATHNAME), name)
	}

	m.Conf(RUNTIME, kit.Keys(BOOT, USERNAME), kit.Select(kit.UserName(), kit.Env(CTX_USER)))
	ice.Info.HostName = m.Conf(RUNTIME, kit.Keys(BOOT, HOSTNAME))
	ice.Info.PathName = m.Conf(RUNTIME, kit.Keys(BOOT, PATHNAME))
	ice.Info.UserName = m.Conf(RUNTIME, kit.Keys(BOOT, USERNAME))

	// 启动次数 boot
	m.Conf(RUNTIME, kit.Keys(BOOT, mdb.COUNT), kit.Int(m.Conf(RUNTIME, kit.Keys(BOOT, mdb.COUNT)))+1)
	m.Conf(RUNTIME, kit.Keys(BOOT, ice.BIN), _system_find(m, os.Args[0]))

	// 环境变量 conf
	for _, k := range []string{CTX_SHY, CTX_DEV, CTX_OPS, CTX_ARG, CTX_PID, CTX_USER, CTX_SHARE, CTX_RIVER} {
		m.Conf(RUNTIME, kit.Keys(CONF, k), kit.Env(k))
	}
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
	m.Spawn().Split(m.Cmdx(SYSTEM, "df", "-h"), "", ice.SP, ice.NL).Table(func(index int, value map[string]string, head []string) {
		if strings.HasPrefix(value["Filesystem"], "/dev") {
			m.Push("", value, head)
		}
	})
	m.RenameAppend("%iused", "piused", "Use%", "Usep")
	m.DisplayStory("pie.js?field=Size")
}

func NodeInfo(m *ice.Message, kind, name string) {
	m.Conf(RUNTIME, kit.Keys(NODE, mdb.TIME), m.Time())
	ice.Info.NodeType = m.Conf(RUNTIME, kit.Keys(NODE, mdb.TYPE), kind)
	ice.Info.NodeName = m.Conf(RUNTIME, kit.Keys(NODE, mdb.NAME), strings.ReplaceAll(name, ice.PT, "_"))
}

const (
	MAKE = "make"
	TEST = "test"
	HOST = "host"
	BOOT = "boot"
	CONF = "conf"
	NODE = "node"
)
const (
	GOARCH = "GOARCH"
	AMD64  = "amd64"
	X386   = "386"
	ARM    = "arm"
	ARM64  = "arm64"

	GOOS    = "GOOS"
	LINUX   = "linux"
	DARWIN  = "darwin"
	WINDOWS = "windows"
)
const (
	SHELL = "SHELL"
	TERM  = "TERM"
	USER  = "USER"
	HOME  = "HOME"
	PATH  = "PATH"
)
const (
	CTX_SHY = "ctx_shy"
	CTX_DEV = "ctx_dev"
	CTX_OPS = "ctx_ops"
	CTX_ARG = "ctx_arg"

	CTX_PID   = "ctx_pid"
	CTX_LOG   = "ctx_log"
	CTX_USER  = "ctx_user"
	CTX_SHARE = "ctx_share"
	CTX_RIVER = "ctx_river"
)
const (
	HOSTNAME = "hostname"
	PATHNAME = "pathname"
	USERNAME = "username"
)
const (
	MAXPROCS = "maxprocs"
	IFCONFIG = "ifconfig"
	HOSTINFO = "hostinfo"
	USERINFO = "userinfo"
	PROCINFO = "procinfo"
	PROCKILL = "prockill"
	BOOTINFO = "bootinfo"
	DISKINFO = "diskinfo"
)
const RUNTIME = "runtime"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		RUNTIME: {Name: RUNTIME, Help: "运行环境", Value: kit.Dict()},
	}, Commands: map[string]*ice.Command{
		RUNTIME: {Name: "runtime info=ifconfig,hostinfo,hostname,userinfo,procinfo,bootinfo,diskinfo,env,file,route auto", Help: "运行环境", Action: map[string]*ice.Action{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				_runtime_init(m)
				m.Cmd(RUNTIME, MAXPROCS, "1")
			}},
			MAXPROCS: {Name: "maxprocs", Help: "最大并发", Hand: func(m *ice.Message, arg ...string) {
				if len(arg) > 0 {
					runtime.GOMAXPROCS(kit.Int(m.Conf(RUNTIME, kit.Keys(HOST, "GOMAXPROCS"), kit.Select("1", arg, 0))))
				}
				m.Echo("%d", runtime.GOMAXPROCS(0))
			}},
			IFCONFIG: {Name: "ifconfig", Help: "网卡配置", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy("tcp.host")
			}},
			HOSTINFO: {Name: "hostinfo", Help: "主机信息", Hand: func(m *ice.Message, arg ...string) {
				_runtime_hostinfo(m)
			}},
			HOSTNAME: {Name: "hostname", Help: "主机域名", Hand: func(m *ice.Message, arg ...string) {
				if len(arg) > 0 {
					ice.Info.HostName = m.Conf(RUNTIME, kit.Keys(BOOT, HOSTNAME), m.Conf(RUNTIME, kit.Keys(NODE, mdb.NAME), arg[0]))
				}
				m.Echo(ice.Info.HostName)
			}},
			USERINFO: {Name: "userinfo", Help: "用户信息", Hand: func(m *ice.Message, arg ...string) {
				m.Split(m.Cmdx(SYSTEM, "who"), "user term time")
			}},
			PROCINFO: {Name: "procinfo", Help: "进程信息", Hand: func(m *ice.Message, arg ...string) {
				m.Split(m.Cmdx(SYSTEM, "ps", "u")).PushAction(PROCKILL)
				m.StatusTimeCount()
			}},
			PROCKILL: {Name: "prockill", Help: "结束进程", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(SYSTEM, KILL, m.Option("PID"))
				m.ProcessRefresh30ms()
			}},
			DISKINFO: {Name: "diskinfo", Help: "磁盘信息", Hand: func(m *ice.Message, arg ...string) {
				_runtime_diskinfo(m)
			}},
			"env": {Name: "env", Help: "环境变量", Hand: func(m *ice.Message, arg ...string) {
				for _, v := range os.Environ() {
					ls := strings.SplitN(v, "=", 2)
					m.Push(mdb.NAME, ls[0])
					m.Push(mdb.VALUE, ls[1])
				}
			}},
			"file": {Name: "file", Help: "模块文件", Hand: func(m *ice.Message, arg ...string) {
				for k, v := range ice.Info.File {
					m.Push(nfs.FILE, k)
					m.Push(mdb.NAME, v)
				}
				m.Sort(nfs.FILE)
			}},
			"route": {Name: "route", Help: "接口命令", Hand: func(m *ice.Message, arg ...string) {
				for k, v := range ice.Info.Route {
					m.Push(nfs.PATH, k)
					m.Push(nfs.FILE, v)
				}
				m.Sort(nfs.PATH)
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 0 && arg[0] == BOOTINFO {
				arg = arg[1:]
			}
			m.Cmdy(ctx.CONFIG, RUNTIME, arg)
			m.DisplayStoryJSON()
		}},
	}})
}
