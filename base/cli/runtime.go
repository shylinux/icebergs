package cli

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/ctx"
	kit "github.com/shylinux/toolkits"
)

const (
	DISKINFO = "diskinfo"
	IFCONFIG = "ifconfig"
	HOSTINFO = "hostinfo"
	USERINFO = "userinfo"
	PROCINFO = "procinfo"
	BOOTINFO = "bootinfo"
)
const RUNTIME = "runtime"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			RUNTIME: {Name: RUNTIME, Help: "运行环境", Value: kit.Dict()},
		},
		Commands: map[string]*ice.Command{
			RUNTIME: {Name: "runtime info=diskinfo,ifconfig,hostname,hostinfo,userinfo,procinfo,bootinfo auto", Help: "运行环境", Action: map[string]*ice.Action{
				DISKINFO: {Name: "diskinfo", Help: "磁盘信息", Hand: func(m *ice.Message, arg ...string) {
					m.Spawn().Split(m.Cmdx(SYSTEM, "df", "-h"), "", " ", "\n").Table(func(index int, value map[string]string, head []string) {
						if strings.HasPrefix(value["Filesystem"], "/dev") {
							m.Push("", value, head)
						}
					})
				}},
				IFCONFIG: {Name: "ifconfig", Help: "网卡配置", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy("tcp.host")
				}},
				HOSTNAME: {Name: "hostname", Help: "主机域名", Hand: func(m *ice.Message, arg ...string) {
					if len(arg) > 0 {
						m.Conf(RUNTIME, kit.Keys(NODE, kit.MDB_NAME), arg[0])
						m.Conf(RUNTIME, kit.Keys(BOOT, HOSTNAME), arg[0])
						ice.Info.HostName = arg[0]
					}
					m.Echo(ice.Info.HostName)
				}},
				HOSTINFO: {Name: "hostinfo", Help: "主机信息", Hand: func(m *ice.Message, arg ...string) {
					if f, e := os.Open("/proc/cpuinfo"); e == nil {
						defer f.Close()
						if b, e := ioutil.ReadAll(f); e == nil {
							m.Push("nCPU", bytes.Count(b, []byte("processor")))
						}
					}
					if f, e := os.Open("/proc/meminfo"); e == nil {
						defer f.Close()
						if b, e := ioutil.ReadAll(f); e == nil {
							for i, ls := range strings.Split(string(b), "\n") {
								vs := kit.Split(ls, ": ")
								m.Push(strings.TrimSpace(vs[0]), kit.FmtSize(kit.Int64(strings.TrimSpace(vs[1]))*1024))
								if i > 1 {
									break
								}
							}
						}
					}
					m.Push("uptime", kit.Split(m.Cmdx(SYSTEM, "uptime"), ",")[0])
				}},
				USERINFO: {Name: "userinfo", Help: "用户信息", Hand: func(m *ice.Message, arg ...string) {
					m.Split(m.Cmdx(SYSTEM, "who"), "user term time", " ", "\n")
				}},
				PROCINFO: {Name: "procinfo", Help: "进程信息", Hand: func(m *ice.Message, arg ...string) {
					m.Split(m.Cmdx(SYSTEM, "ps", "u"), "", " ", "\n")
					m.PushAction("kill")
				}},
				"kill": {Name: "kill", Help: "结束进程", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(SYSTEM, "kill", m.Option("PID"))
					m.ProcessRefresh("10ms")
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) > 0 && arg[0] == BOOTINFO {
					arg = arg[1:]
				}
				m.Cmdy(ctx.CONFIG, RUNTIME, arg)
			}},
		},
	})
}
