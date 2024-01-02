package tcp

import (
	"fmt"
	"net"
	"os"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _host_domain(m *ice.Message) string {
	return kit.GetValid(
		func() string { return m.Option(ice.TCP_DOMAIN) },
		func() string { return mdb.Config(m, DOMAIN) },
		func() string { return os.Getenv(ice.TCP_DOMAIN) },
		func() string {
			if !kit.IsIn(m.ActionKey(), "", ice.LIST) {
				return m.Cmdv(HOST, mdb.Config(m, ice.MAIN), aaa.IP)
			}
			return ""
		},
	)
}
func _host_list(m *ice.Message, name string) {
	if ifs, e := net.Interfaces(); m.Assert(e) {
		for _, v := range ifs {
			if !strings.Contains(v.Name, name) || len(v.HardwareAddr.String()) == 0 {
				continue
			}
			if ips, e := v.Addrs(); m.Assert(e) {
				for _, x := range ips {
					ip := strings.Split(x.String(), nfs.PS)
					if strings.Contains(ip[0], nfs.DF) || len(ip) == 0 {
						continue
					}
					m.Push(mdb.INDEX, v.Index).Push(mdb.NAME, v.Name).Push(aaa.IP, ip[0]).Push("mask", ip[1]).Push("hard", v.HardwareAddr.String())
				}
			}
		}
	}
	if len(m.Appendv(aaa.IP)) == 0 {
		m.Push(mdb.INDEX, -1).Push(mdb.NAME, LOCALHOST).Push(aaa.IP, "127.0.0.1").Push("mask", "255.0.0.0").Push("hard", "")
	}
	m.SortInt(mdb.INDEX)
	m.StatusTimeCount(DOMAIN, _host_domain(m))
}

const (
	LOCALHOST = "localhost"

	ISLOCAL = "islocal"
	PUBLISH = "publish"
	GATEWAY = "gateway"
	DOMAIN  = "domain"
)
const HOST = "host"

func init() {
	Index.MergeCommands(ice.Commands{
		HOST: {Name: "host name auto domain", Help: "主机", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd("", func(value ice.Maps) { m.Cmd("", aaa.WHITE, LOCALHOST, value[aaa.IP]) })
			}},
			mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
				if mdb.IsSearchPreview(m, arg) {
					ip := m.Cmdv(HOST, GATEWAY, aaa.IP)
					m.PushSearch(mdb.TYPE, GATEWAY, mdb.NAME, ip, mdb.TEXT, "http://"+ip)
				}
			}},
			aaa.WHITE: {Name: "white name text", Help: "白名单", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashCreate(m, mdb.TYPE, m.ActionKey(), m.OptionSimple(mdb.NAME, mdb.TEXT))
			}},
			aaa.BLACK: {Name: "black name text", Help: "黑名单", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashCreate(m, mdb.TYPE, m.ActionKey(), m.OptionSimple(mdb.NAME, mdb.TEXT))
			}},
			ISLOCAL: {Hand: func(m *ice.Message, arg ...string) {
				if arg[0] = strings.Split(strings.TrimPrefix(arg[0], "["), "]")[0]; arg[0] == "::1" || strings.HasPrefix(arg[0], "127.") || arg[0] == LOCALHOST {
					m.Echo(ice.OK)
				} else if mdb.HashSelectField(m, strings.Split(arg[0], nfs.DF)[0], mdb.TYPE) == aaa.WHITE {
					m.Echo(ice.OK)
				}
			}},
			PUBLISH: {Hand: func(m *ice.Message, arg ...string) {
				if strings.Contains(arg[0], LOCALHOST) {
					arg[0] = strings.Replace(arg[0], LOCALHOST, _host_domain(m), 1)
				} else if strings.Contains(arg[0], "127.0.0.1") {
					arg[0] = strings.Replace(arg[0], "127.0.0.1", _host_domain(m), 1)
				}
				m.Echo(arg[0])
			}},
			GATEWAY: {Hand: func(m *ice.Message, arg ...string) {
				m.Push(aaa.IP, kit.Keys(kit.Slice(strings.Split(m.Cmdv(HOST, aaa.IP), nfs.PT), 0, 3), "1"))
			}},
			DOMAIN: {Name: "domain ip", Hand: func(m *ice.Message, arg ...string) {
				kit.If(m.Option(aaa.IP), func(p string) { mdb.Config(m, DOMAIN, p) })
				m.Echo(mdb.Config(m, DOMAIN))
			}},
		}, mdb.HashAction(mdb.SHORT, mdb.TEXT)), Hand: func(m *ice.Message, arg ...string) {
			_host_list(m, kit.Select("", arg, 0))
			m.PushAction(DOMAIN)
		}},
	})
}

func Address(host, port string) string                   { return fmt.Sprintf("%s:%s", host, port) }
func IsLocalHost(m *ice.Message, ip string) bool         { return m.Cmdx(HOST, ISLOCAL, ip) == ice.OK }
func PublishLocalhost(m *ice.Message, url string) string { return m.Cmdx(HOST, PUBLISH, url) }
