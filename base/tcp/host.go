package tcp

import (
	"net"
	"strings"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	kit "github.com/shylinux/toolkits"
)

func _host_list(m *ice.Message, name string) {
	if ifs, e := net.Interfaces(); m.Assert(e) {
		for _, v := range ifs {
			if name != "" && !strings.Contains(v.Name, name) {
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

					m.Push(kit.MDB_INDEX, v.Index)
					m.Push(kit.MDB_NAME, v.Name)
					m.Push(IP, ip[0])
					m.Push(MASK, ip[1])
					m.Push(HARD, v.HardwareAddr.String())
				}
			}
		}
	}

	if len(m.Appendv(IP)) == 0 {
		m.Push(kit.MDB_INDEX, -1)
		m.Push(kit.MDB_NAME, LOCALHOST)
		m.Push(IP, "127.0.0.1")
		m.Push(MASK, "255.0.0.0")
		m.Push(HARD, "")
	}
}

func _islocalhost(m *ice.Message, ip string) (ok bool) {
	if ip == "::1" || strings.HasPrefix(ip, "127.") {
		return true
	}
	if m.Richs(HOST, kit.Keym(aaa.BLACK), ip, nil) != nil {
		return false
	}
	if m.Richs(HOST, kit.Keym(aaa.WHITE), ip, nil) != nil {
		m.Log_AUTH(aaa.WHITE, ip)
		return true
	}
	return false
}
func IsLocalHost(m *ice.Message, ip string) bool { return _islocalhost(m, ip) }

const (
	HOSTPORT = "hostport"
	HOSTNAME = "hostname"
	PROTOCOL = "protocol"

	LOCALHOST = "localhost"

	HARD = "hard"
	MASK = "mask"
	IP   = "ip"
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
					m.Rich(HOST, kit.Keym(aaa.BLACK), kit.Dict(kit.MDB_TEXT, arg[0]))
				}},
				aaa.WHITE: {Name: "white", Help: "白名单", Hand: func(m *ice.Message, arg ...string) {
					m.Rich(HOST, kit.Keym(aaa.WHITE), kit.Dict(kit.MDB_TEXT, arg[0]))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				_host_list(m, kit.Select("", arg, 0))
			}},
		},
	})
}
