package cli

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"runtime"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	kit "shylinux.com/x/toolkits"
)

func _runtime_init(m *ice.Message) {
	// 版本信息
	kit.Fetch(kit.UnMarshal(kit.Format(ice.Info.Make)), func(key string, value interface{}) {
		m.Conf(RUNTIME, kit.Keys(MAKE, strings.ToLower(key)), value)
	})

	// 环境变量
	for _, k := range []string{CTX_SHY, CTX_DEV, CTX_OPS, CTX_ARG, CTX_PID, CTX_USER, CTX_SHARE, CTX_RIVER} {
		m.Conf(RUNTIME, kit.Keys(CONF, k), os.Getenv(k))
	}

	// 主机信息
	m.Conf(RUNTIME, kit.Keys(HOST, GOARCH), runtime.GOARCH)
	m.Conf(RUNTIME, kit.Keys(HOST, GOOS), runtime.GOOS)
	m.Conf(RUNTIME, kit.Keys(HOST, "pid"), os.Getpid())

	// 启动信息
	if name, e := os.Hostname(); e == nil {
		m.Conf(RUNTIME, kit.Keys(BOOT, HOSTNAME), kit.Select(name, os.Getenv("HOSTNAME")))
	}
	if name, e := os.Getwd(); e == nil {
		name = path.Base(kit.Select(name, os.Getenv("PWD")))
		ls := strings.Split(name, "/")
		name = ls[len(ls)-1]
		ls = strings.Split(name, "\\")
		name = ls[len(ls)-1]
		m.Conf(RUNTIME, kit.Keys(BOOT, PATHNAME), name)
	}
	if m.Conf(RUNTIME, kit.Keys(BOOT, USERNAME), kit.Select(os.Getenv(USER), os.Getenv(CTX_USER))) == "" {
		if user, e := user.Current(); e == nil && user.Name != "" {
			m.Conf(RUNTIME, kit.Keys(BOOT, USERNAME), kit.Select(user.Name, os.Getenv(CTX_USER)))
		}
	}
	ice.Info.HostName = m.Conf(RUNTIME, kit.Keys(BOOT, HOSTNAME))
	ice.Info.PathName = m.Conf(RUNTIME, kit.Keys(BOOT, PATHNAME))
	ice.Info.UserName = m.Conf(RUNTIME, kit.Keys(BOOT, USERNAME))

	// 启动次数
	count := kit.Int(m.Conf(RUNTIME, kit.Keys(BOOT, kit.MDB_COUNT))) + 1
	m.Conf(RUNTIME, kit.Keys(BOOT, kit.MDB_COUNT), count)

	// 节点信息
	m.Conf(RUNTIME, kit.Keys(NODE, kit.MDB_TIME), m.Time())
	NodeInfo(m, "worker", m.Conf(RUNTIME, kit.Keys(BOOT, PATHNAME)))

	runtime.GOMAXPROCS(kit.Int(kit.Select("1", m.Conf(RUNTIME, kit.Keys(HOST, "GOMAXPROCS")))))
}
func _runtime_hostinfo(m *ice.Message) {
	if f, e := os.Open("/proc/cpuinfo"); e == nil {
		defer f.Close()
		if b, e := ioutil.ReadAll(f); e == nil {
			m.Push("nCPU", bytes.Count(b, []byte("processor")))
		}
	}
	if f, e := os.Open("/proc/meminfo"); e == nil {
		defer f.Close()
		if b, e := ioutil.ReadAll(f); e == nil {
			for i, ls := range strings.Split(string(b), ice.NL) {
				vs := kit.Split(ls, ": ")
				m.Push(strings.TrimSpace(vs[0]), kit.FmtSize(kit.Int64(strings.TrimSpace(vs[1]))*1024))
				if i > 1 {
					break
				}
			}
		}
	}
	m.Push("uptime", kit.Split(m.Cmdx(SYSTEM, "uptime"), ",")[0])
}

func NodeInfo(m *ice.Message, kind, name string) {
	name = strings.ReplaceAll(name, ice.PT, "_")
	m.Conf(RUNTIME, kit.Keys(NODE, kit.MDB_TYPE), kind)
	m.Conf(RUNTIME, kit.Keys(NODE, kit.MDB_NAME), name)
	ice.Info.NodeName = name
	ice.Info.NodeType = kind
}

const (
	MAKE = "make"
	TEST = "test"
	CONF = "conf"
	HOST = "host"
	BOOT = "boot"
	NODE = "node"
)
const (
	SOURCE = "source"
	TARGET = "target"
)
const (
	GOARCH  = "GOARCH"
	GOOS    = "GOOS"
	X386    = "386"
	AMD64   = "amd64"
	ARM64   = "arm64"
	ARM     = "arm"
	LINUX   = "linux"
	DARWIN  = "darwin"
	WINDOWS = "windows"
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
	IFCONFIG = "ifconfig"
	HOSTINFO = "hostinfo"
	USERINFO = "userinfo"
	PROCINFO = "procinfo"
	BOOTINFO = "bootinfo"
	DISKINFO = "diskinfo"
)
const RUNTIME = "runtime"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		RUNTIME: {Name: RUNTIME, Help: "运行环境", Value: kit.Dict()},
	}, Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			_runtime_init(m)
		}},
		RUNTIME: {Name: "runtime info=ifconfig,hostinfo,hostname,userinfo,procinfo,bootinfo,diskinfo auto", Help: "运行环境", Action: map[string]*ice.Action{
			IFCONFIG: {Name: "ifconfig", Help: "网卡配置", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy("tcp.host")
			}},
			HOSTINFO: {Name: "hostinfo", Help: "主机信息", Hand: func(m *ice.Message, arg ...string) {
				_runtime_hostinfo(m)
			}},
			HOSTNAME: {Name: "hostname", Help: "主机域名", Hand: func(m *ice.Message, arg ...string) {
				if len(arg) > 0 {
					m.Conf(RUNTIME, kit.Keys(NODE, kit.MDB_NAME), arg[0])
					m.Conf(RUNTIME, kit.Keys(BOOT, HOSTNAME), arg[0])
					ice.Info.HostName = arg[0]
				}
				m.Echo(ice.Info.HostName)
			}},
			USERINFO: {Name: "userinfo", Help: "用户信息", Hand: func(m *ice.Message, arg ...string) {
				m.Split(m.Cmdx(SYSTEM, "who"), "user term time", ice.SP, ice.NL)
			}},
			PROCINFO: {Name: "procinfo", Help: "进程信息", Hand: func(m *ice.Message, arg ...string) {
				m.Split(m.Cmdx(SYSTEM, "ps", "u"), "", ice.SP, ice.NL)
				m.PushAction("prockill")
				m.StatusTimeCount()
			}},
			"prockill": {Name: "prockill", Help: "结束进程", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(SYSTEM, "prockill", m.Option("PID"))
				m.ProcessRefresh30ms()
			}},
			DISKINFO: {Name: "diskinfo", Help: "磁盘信息", Hand: func(m *ice.Message, arg ...string) {
				m.Spawn().Split(m.Cmdx(SYSTEM, "df", "-h"), "", ice.SP, ice.NL).Table(func(index int, value map[string]string, head []string) {
					if strings.HasPrefix(value["Filesystem"], "/dev") {
						m.Push("", value, head)
					}
				})
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 0 && arg[0] == BOOTINFO {
				arg = arg[1:]
			}
			m.Cmdy(ctx.CONFIG, RUNTIME, arg)
		}},
	}})
}
