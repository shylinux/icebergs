package tcp

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/toolkits"

	"net"
	"strings"
)

var Index = &ice.Context{Name: "tcp", Help: "通信模块",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"getport": &ice.Config{Name: "getport", Help: "getport", Value: kit.Data(
			"begin", 10000, "end", 20000,
		)},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Load()
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Save("getport")
		}},

		"ip": {Name: "ifconfig [name]", Help: "网络配置", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if addr, e := net.InterfaceAddrs(); m.Assert(e) {
				for _, v := range addr {
					m.Info("%v", v)
				}
			}
		}},
		"getport": {Name: "getport", Help: "分配端口", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			begin := kit.Int(m.Conf("getport", "meta.begin"))
			end := kit.Int(m.Conf("getport", "meta.end"))
			if begin >= end {
				begin = 10000
			}
			for i := begin; i < end; i++ {
				if m.Cmd(ice.CLI_SYSTEM, "lsof", "-i", kit.Format(":%d", i)).Append("code") != "0" {
					m.Conf("getport", "meta.begin", i+1)
					m.Echo("%d", i)
					break
				}
			}
		}},
		"netstat": {Name: "netstat [name]", Help: "网络配置", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmdy(ice.CLI_SYSTEM, "netstat", "-lanp")
		}},
		"ifconfig": {Name: "ifconfig [name]", Help: "网络配置", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if ifs, e := net.Interfaces(); m.Assert(e) {
				for _, v := range ifs {
					if len(arg) > 0 && !strings.Contains(v.Name, arg[0]) {
						continue
					}
					if ips, e := v.Addrs(); m.Assert(e) {
						for _, x := range ips {
							ip := strings.Split(x.String(), "/")
							if strings.Contains(ip[0], ":") || len(ip) == 0 {
								continue
							}
							if len(v.HardwareAddr.String()) == 0 {
								continue
							}

							m.Push("index", v.Index)
							m.Push("name", v.Name)
							m.Push("ip", ip[0])
							m.Push("mask", ip[1])
							m.Push("hard", v.HardwareAddr.String())
						}
					}
				}
			}
		}},
	},
}

func init() {
	ice.Index.Register(Index, nil)
}
