package tcp

import (
	"net"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
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
					ip := strings.Split(x.String(), ice.PS)
					if strings.Contains(ip[0], ice.DF) || len(ip) == 0 {
						continue
					}

					m.Push(mdb.INDEX, v.Index)
					m.Push(mdb.NAME, v.Name)
					m.Push(aaa.IP, ip[0])
					m.Push("mask", ip[1])
					m.Push("hard", v.HardwareAddr.String())
				}
			}
		}
	}

	if len(m.Appendv(aaa.IP)) == 0 {
		m.Push(mdb.INDEX, -1)
		m.Push(mdb.NAME, LOCALHOST)
		m.Push(aaa.IP, "127.0.0.1")
		m.Push("mask", "255.0.0.0")
		m.Push("hard", "")
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
func ReplaceLocalhost(m *ice.Message, url string) string {
	if strings.Contains(url, "://"+LOCALHOST) {
		url = strings.Replace(url, "://"+LOCALHOST, "://"+m.Cmd(HOST).Append(aaa.IP), 1)
	}
	return url
}

const (
	LOCALHOST = "localhost"
)
const HOST = "host"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		HOST: {Name: HOST, Help: "主机", Value: kit.Data(
			aaa.BLACK, kit.Data(mdb.SHORT, mdb.TEXT), aaa.WHITE, kit.Data(mdb.SHORT, mdb.TEXT),
		)},
	}, Commands: map[string]*ice.Command{
		HOST: {Name: "host name auto", Help: "主机", Action: map[string]*ice.Action{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(HOST).Table(func(index int, value map[string]string, head []string) {
					m.Cmd(HOST, aaa.WHITE, value[aaa.IP])
				})
			}},
			aaa.BLACK: {Name: "black", Help: "黑名单", Hand: func(m *ice.Message, arg ...string) {
				m.Log_CREATE(aaa.BLACK, arg[0])
				m.Rich(HOST, kit.Keym(aaa.BLACK), kit.Dict(mdb.TEXT, arg[0]))
			}},
			aaa.WHITE: {Name: "white", Help: "白名单", Hand: func(m *ice.Message, arg ...string) {
				m.Log_CREATE(aaa.WHITE, arg[0])
				m.Rich(HOST, kit.Keym(aaa.WHITE), kit.Dict(mdb.TEXT, arg[0]))
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			_host_list(m, kit.Select("", arg, 0))
		}},
	}})
}
