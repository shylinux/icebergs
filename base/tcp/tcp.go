package tcp

import (
	"github.com/shylinux/icebergs"

	"net"
	"strings"
)

var Index = &ice.Context{Name: "tcp", Help: "网络模块",
	Caches:  map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
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
