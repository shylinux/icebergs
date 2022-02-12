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
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _runtime_init(m *ice.Message) {
	// 版本信息 make
	kit.Fetch(kit.UnMarshal(kit.Format(ice.Info.Make)), func(key string, value interface{}) {
		m.Conf(RUNTIME, kit.Keys(MAKE, strings.ToLower(key)), value)
	})

	// 环境变量 conf
	for _, k := range []string{CTX_SHY, CTX_DEV, CTX_OPS, CTX_ARG, CTX_PID, CTX_USER, CTX_SHARE, CTX_RIVER} {
		m.Conf(RUNTIME, kit.Keys(CONF, k), os.Getenv(k))
	}

	// 主机信息 host
	m.Conf(RUNTIME, kit.Keys(HOST, GOARCH), runtime.GOARCH)
	m.Conf(RUNTIME, kit.Keys(HOST, GOOS), runtime.GOOS)
	m.Conf(RUNTIME, kit.Keys(HOST, "pid"), os.Getpid())
	m.Conf(RUNTIME, kit.Keys(HOST, HOME), os.Getenv(HOME))
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
		m.Conf(RUNTIME, kit.Keys(BOOT, HOSTNAME), kit.Select(name, os.Getenv("HOSTNAME")))
	}
	if name, e := os.Getwd(); e == nil {
		name = path.Base(kit.Select(name, os.Getenv("PWD")))
		ls := strings.Split(name, ice.PS)
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

	// 启动次数 boot
	count := kit.Int(m.Conf(RUNTIME, kit.Keys(BOOT, mdb.COUNT))) + 1
	m.Conf(RUNTIME, kit.Keys(BOOT, mdb.COUNT), count)
	m.Conf(RUNTIME, kit.Keys(BOOT, ice.BIN), m.Cmdx(SYSTEM, "which", os.Args[0]))

	// 节点信息 node
	m.Conf(RUNTIME, kit.Keys(NODE, mdb.TIME), m.Time())
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
	m.Push("uptime", kit.Split(m.Cmdx(SYSTEM, "uptime"), ice.FS)[0])
}

func NodeInfo(m *ice.Message, kind, name string) {
	name = strings.ReplaceAll(name, ice.PT, "_")
	m.Conf(RUNTIME, kit.Keys(NODE, mdb.TYPE), kind)
	m.Conf(RUNTIME, kit.Keys(NODE, mdb.NAME), name)
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
	GOARCH  = "GOARCH"
	GOOS    = "GOOS"
	X386    = "386"
	AMD64   = "amd64"
	ARM64   = "arm64"
	ARM     = "arm"
	LINUX   = "linux"
	DARWIN  = "darwin"
	WINDOWS = "windows"

	OSID   = "OSID"
	CENTOS = "centos"
	UBUNTU = "ubuntu"
	ALPINE = "alpine"
)
const (
	USER = "USER"
	HOME = "HOME"
	PATH = "PATH"
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
		RUNTIME: {Name: "runtime info=ifconfig,hostinfo,hostname,userinfo,procinfo,bootinfo,diskinfo auto", Help: "运行环境", Action: map[string]*ice.Action{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				_runtime_init(m)
			}},
			IFCONFIG: {Name: "ifconfig", Help: "网卡配置", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy("tcp.host")
			}},
			HOSTINFO: {Name: "hostinfo", Help: "主机信息", Hand: func(m *ice.Message, arg ...string) {
				_runtime_hostinfo(m)
			}},
			HOSTNAME: {Name: "hostname", Help: "主机域名", Hand: func(m *ice.Message, arg ...string) {
				if len(arg) > 0 {
					m.Conf(RUNTIME, kit.Keys(NODE, mdb.NAME), arg[0])
					m.Conf(RUNTIME, kit.Keys(BOOT, HOSTNAME), arg[0])
					ice.Info.HostName = arg[0]
				}
				m.Echo(ice.Info.HostName)
			}},
			USERINFO: {Name: "userinfo", Help: "用户信息", Hand: func(m *ice.Message, arg ...string) {
				m.Split(m.Cmdx(SYSTEM, "who"), "user term time")
			}},
			PROCINFO: {Name: "procinfo", Help: "进程信息", Hand: func(m *ice.Message, arg ...string) {
				m.Split(m.Cmdx(SYSTEM, "ps", "u"))
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
				m.Display("/plugin/story/pie.js?field=Size")
				m.RenameAppend("%iused", "piused")
				m.RenameAppend("Use%", "Usep")
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 0 && arg[0] == BOOTINFO {
				arg = arg[1:]
			}
			m.Cmdy(ctx.CONFIG, RUNTIME, arg)
			m.Display("/plugin/story/json.js")
		}},
	}})
}
