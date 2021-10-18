package cli

import (
	"os"
	"os/user"
	"path"
	"runtime"
	"strings"

	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

func _cli_init(m *ice.Message) {
	// 版本信息
	kit.Fetch(kit.UnMarshal(kit.Format(ice.Info.Make)), func(key string, value interface{}) {
		m.Conf(RUNTIME, kit.Keys(MAKE, strings.ToLower(key)), value)
	})

	// 环境变量
	for _, k := range []string{CTX_SHY, CTX_DEV, CTX_OPS, CTX_PID, CTX_USER, CTX_SHARE, CTX_RIVER} {
		m.Conf(RUNTIME, kit.Keys(CONF, k), os.Getenv(k))
	}

	// 主机信息
	m.Conf(RUNTIME, kit.Keys(HOST, "GOARCH"), runtime.GOARCH)
	m.Conf(RUNTIME, kit.Keys(HOST, "GOOS"), runtime.GOOS)
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
func NodeInfo(m *ice.Message, kind, name string) {
	name = strings.ReplaceAll(name, ".", "_")
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
	HOSTNAME = "hostname"
	PATHNAME = "pathname"
	USERNAME = "username"
)
const (
	CTX_SHY   = "ctx_shy"
	CTX_DEV   = "ctx_dev"
	CTX_OPS   = "ctx_ops"
	CTX_PID   = "ctx_pid"
	CTX_LOG   = "ctx_log"
	CTX_USER  = "ctx_user"
	CTX_SHARE = "ctx_share"
	CTX_RIVER = "ctx_river"
)
const CLI = "cli"

var Index = &ice.Context{Name: CLI, Help: "命令模块",
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Load()
			_cli_init(m)
		}},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Save()
		}},
	},
}

func init() { ice.Index.Register(Index, nil, RUNTIME, SYSTEM, DAEMON, QRCODE) }
