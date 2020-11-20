package cli

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/ctx"
	kit "github.com/shylinux/toolkits"

	"bytes"
	"io/ioutil"
	"os"
	"strings"
)

const RUNTIME = "runtime"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			RUNTIME: {Name: RUNTIME, Help: "运行环境", Value: kit.Dict()},
		},
		Commands: map[string]*ice.Command{
			RUNTIME: {Name: "runtime info=procinfo,hostinfo,diskinfo,ifconfig,hostname,userinfo,bootinfo auto", Help: "运行环境", Action: map[string]*ice.Action{
				"procinfo": {Name: "procinfo", Help: "procinfo", Hand: func(m *ice.Message, arg ...string) {
					m.Split(m.Cmdx(SYSTEM, "ps", "u"), "", " ", "\n")
				}},
				"hostinfo": {Name: "hostinfo", Help: "hostinfo", Hand: func(m *ice.Message, arg ...string) {
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
				"diskinfo": {Name: "diskinfo", Help: "diskinfo", Hand: func(m *ice.Message, arg ...string) {
					m.Spawn().Split(m.Cmdx(SYSTEM, "df", "-h"), "", " ", "\n").Table(func(index int, value map[string]string, head []string) {
						if strings.HasPrefix(value["Filesystem"], "/dev") {
							m.Push("", value, head)
						}
					})
				}},
				"ifconfig": {Name: "ifconfig", Help: "ifconfig", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy("tcp.host")
				}},
				"hostname": {Name: "hostname", Help: "hostname", Hand: func(m *ice.Message, arg ...string) {
					m.Conf(RUNTIME, "boot.hostname", arg[0])
					m.Conf(RUNTIME, "node.name", arg[0])
					ice.Info.HostName = arg[0]
					m.Echo(ice.Info.HostName)
				}},
				"userinfo": {Name: "userinfo", Help: "userinfo", Hand: func(m *ice.Message, arg ...string) {
					m.Split(m.Cmdx(SYSTEM, "who"), "user term time", " ", "\n")
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) > 0 && arg[0] == "bootinfo" {
					arg = arg[1:]
				}
				m.Cmdy(ctx.CONFIG, RUNTIME, arg)
			}},
		},
	})
}
