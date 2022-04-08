package web

import (
	"net"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
)

func _udp_addr(m *ice.Message, host, port string) *net.UDPAddr {
	addr, err := net.ResolveUDPAddr("udp4", kit.Format("%s:%s", host, port))
	m.Assert(err)
	return addr
}
func _udp_broad(m *ice.Message, host, port string, remote_host, remote_port string) {
	if s, e := net.DialUDP("udp", nil, _udp_addr(m, remote_host, remote_port)); m.Assert(e) {
		defer s.Close()
		msg := m.Spawn(kit.Dict(tcp.HOST, host, tcp.PORT, port))
		m.Debug("broad %v to %v:%v", msg.FormatMeta(), remote_host, remote_port)
		s.Write([]byte(msg.FormatMeta()))
	}
}
func _serve_udp(m *ice.Message, host, port string) {
	m.Cmd(BROAD, mdb.CREATE, tcp.HOST, host, tcp.PORT, port)
	_udp_broad(m, host, port, "255.255.255.255", "9020")

	if s, e := net.ListenUDP("udp", _udp_addr(m, "0.0.0.0", port)); m.Assert(e) {
		defer s.Close()
		m.Debug("listen %v %v", host, port)

		buf := make([]byte, 1024)
		for {
			n, addr, err := s.ReadFromUDP(buf[:])
			if err != nil {
				m.Debug("what %v", err)
				continue
			}
			m.Debug("recv %v %v", string(buf[:n]), addr)

			msg := m.Spawn(buf[:n])
			if m.Cmd(BROAD, kit.Format("%s,%s", msg.Option(tcp.HOST), msg.Option(tcp.PORT))).Length() > 0 {
				continue
			}

			if remote, err := net.ResolveUDPAddr("udp4", kit.Format("%s:%s", msg.Option(tcp.HOST), msg.Option(tcp.PORT))); err == nil {
				m.Cmd(BROAD).Table(func(index int, value map[string]string, head []string) {
					m.Debug("broad %v to %v", kit.Format(value), kit.Format(remote))
					s.WriteToUDP([]byte(m.Spawn(value).FormatMeta()), remote)
				})
				m.Cmd(BROAD, mdb.CREATE, msg.OptionSimple(tcp.HOST, tcp.PORT))
			} else {
				m.Debug("what %v", err)
			}
		}
	}
}
func _broad_search(m *ice.Message, kind, name, text string, arg ...string) {
	m.Richs(BROAD, nil, mdb.FOREACH, func(key string, value map[string]interface{}) {
		value = kit.GetMeta(value)
		m.PushSearch(mdb.TYPE, "friend", mdb.TEXT, kit.Format("http://%s:%s", value[tcp.HOST], value[tcp.PORT]), value)
	})
}

const BROAD = "broad"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		BROAD: {Name: "broad hash auto", Help: "广播", Action: ice.MergeAction(map[string]*ice.Action{
			SERVE: {Name: "broad port=9020", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
				_serve_udp(m, m.Cmd(tcp.HOST).Append("ip"), m.Option(tcp.PORT))
			}},
			SPACE: {Name: "space dev", Help: "连接", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(SPIDE, mdb.CREATE, mdb.NAME, m.Option(ice.DEV), ADDRESS,
					kit.Format("http://%s:%s", m.Option(tcp.HOST), m.Option(tcp.PORT)))
				m.Cmd(SPACE, tcp.DIAL, m.OptionSimple(ice.DEV))
			}},
		}, mdb.HashAction(
			mdb.SHORT, "host,port", mdb.FIELD, "time,hash,host,port",
		)), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			mdb.HashSelect(m, arg...)
			m.PushAction(SPACE, mdb.REMOVE)
		}},
	}})
}
