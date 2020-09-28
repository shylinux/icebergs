package ssh

import (
	"net"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/tcp"
	kit "github.com/shylinux/toolkits"
	"golang.org/x/crypto/ssh"
)

func _ssh_accept(m *ice.Message, c net.Conn) {
	sc, sessions, req, err := ssh.NewServerConn(c, _ssh_config(m))
	if m.Warn(err != nil, err) {
		return
	}

	m.Gos(m, func(m *ice.Message) { ssh.DiscardRequests(req) })

	for session := range sessions {
		channel, requests, err := session.Accept()
		if m.Warn(err != nil, err) {
			continue
		}

		func(channel ssh.Channel, requests <-chan *ssh.Request) {
			m.Gos(m, func(m *ice.Message) {
				_ssh_handle(m, sc.Permissions.Extensions, c, channel, requests)
			})
		}(channel, requests)
	}
}

const SERVICE = "service"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			SERVICE: {Name: SERVICE, Help: "服务", Value: kit.Data()},
		},
		Commands: map[string]*ice.Command{
			SERVICE: {Name: "service", Help: "服务", Action: map[string]*ice.Action{
				tcp.LISTEN: {Name: "listen name=tcp port=9030", Help: "监听", Hand: func(m *ice.Message, arg ...string) {
					m.Option(tcp.LISTEN_CB, func(c net.Conn) {
						m.Gos(m.Spawn(), func(msg *ice.Message) { _ssh_accept(msg, c) })
					})
					m.Gos(m, func(m *ice.Message) {
						m.Cmdy(tcp.SERVER, tcp.LISTEN, kit.MDB_NAME, "ssh", tcp.PORT, m.Option(tcp.PORT))
					})
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if m.Option(mdb.FIELDS, m.Conf(LISTEN, kit.META_FIELD)); len(arg) > 0 {
					m.Option(mdb.FIELDS, mdb.DETAIL)
				}
				m.Option(mdb.FIELDS, "time,hash,status,host,port")
				m.Cmdy(mdb.SELECT, m.Prefix(LISTEN), "", mdb.HASH, kit.MDB_HASH, arg)

			}},
		},
	}, nil)
}
