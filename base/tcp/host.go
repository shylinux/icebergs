package tcp

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	kit "github.com/shylinux/toolkits"

	"net"
	"strings"
)

func _host_list(m *ice.Message, ifname string) {
	if ifs, e := net.Interfaces(); m.Assert(e) {
		for _, v := range ifs {
			if ifname != "" && !strings.Contains(v.Name, ifname) {
				continue
			}
			if len(v.HardwareAddr.String()) == 0 {
				continue
			}

			if ips, e := v.Addrs(); m.Assert(e) {
				for _, x := range ips {
					ip := strings.Split(x.String(), "/")
					if strings.Contains(ip[0], ":") || len(ip) == 0 {
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

	if len(m.Appendv("ip")) == 0 {
		m.Push("index", -1)
		m.Push("name", "local")
		m.Push("ip", "127.0.0.1")
		m.Push("mask", "255.0.0.0")
		m.Push("hard", "")
	}
}

func _islocalhost(m *ice.Message, ip string) (ok bool) {
	if ip == "::1" || strings.HasPrefix(ip, "127.") {
		return true
	}
	if m.Richs(HOST, kit.Keys("meta.black"), ip, nil) != nil {
		return false
	}
	if m.Richs(HOST, kit.Keys("meta.white"), ip, nil) != nil {
		m.Log_AUTH(aaa.WHITE, ip)
		return true
	}
	return false
}
func IsLocalHost(m *ice.Message, ip string) bool { return _islocalhost(m, ip) }

const (
	HOSTPORT = "hostport"
	HOSTNAME = "hostname"
)
const HOST = "host"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			HOST: {Name: HOST, Help: "主机", Value: kit.Data(
				aaa.BLACK, kit.Data(kit.MDB_SHORT, kit.MDB_TEXT),
				aaa.WHITE, kit.Data(kit.MDB_SHORT, kit.MDB_TEXT),
			)},
		},
		Commands: map[string]*ice.Command{
			HOST: {Name: "host name auto", Help: "主机", Action: map[string]*ice.Action{
				aaa.BLACK: {Name: "black", Help: "黑名单", Hand: func(m *ice.Message, arg ...string) {
					m.Rich(HOST, kit.Keys("meta.black"), kit.Dict(kit.MDB_TEXT, arg[0]))
				}},
				aaa.WHITE: {Name: "white", Help: "白名单", Hand: func(m *ice.Message, arg ...string) {
					m.Rich(HOST, kit.Keys("meta.white"), kit.Dict(kit.MDB_TEXT, arg[0]))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				_host_list(m, kit.Select("", arg, 0))
			}},
		},
	}, nil)
}
