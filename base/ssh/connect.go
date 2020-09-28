package ssh

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/base/tcp"
	kit "github.com/shylinux/toolkits"

	"net"
	"os"
	"path"

	"golang.org/x/crypto/ssh"
)

func _ssh_conn(m *ice.Message, conn net.Conn, hostport, username string) (*ssh.Client, error) {
	key, e := ssh.ParsePrivateKey([]byte(m.Cmdx(nfs.CAT, path.Join(os.Getenv("HOME"), m.Option("private")))))
	m.Assert(e)

	methods := []ssh.AuthMethod{}
	methods = append(methods, ssh.PublicKeys(key))

	c, chans, reqs, err := ssh.NewClientConn(conn, hostport, &ssh.ClientConfig{
		User: username, Auth: methods, HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			m.Logs(CONNECT, aaa.HOSTNAME, hostname, aaa.HOSTPORT, remote.String())
			return nil
		},
	})

	if err != nil {
		return nil, err
	}
	return ssh.NewClient(c, chans, reqs), nil
}

const CONNECT = "connect"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			CONNECT: {Name: CONNECT, Help: "连接", Value: kit.Data()},
		},
		Commands: map[string]*ice.Command{
			CONNECT: {Name: "connect hash auto 添加 清理", Help: "连接", Action: map[string]*ice.Action{
				tcp.DIAL: {Name: "dial username=shy host=shylinux.com port=22 private=.ssh/id_rsa", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Option(tcp.DIAL_CB, func(c net.Conn) {
						client, e := _ssh_conn(m, c,
							kit.Select(m.Option(tcp.HOST), "shylinux.com")+":"+kit.Select("22", m.Option(tcp.PORT)),
							kit.Select("shy", m.Option(aaa.USERNAME)),
						)
						m.Assert(e)

						h := m.Rich(CONNECT, "", kit.Dict(
							aaa.USERNAME, m.Option(aaa.USERNAME),
							tcp.HOST, m.Option(tcp.HOST),
							tcp.PORT, m.Option(tcp.PORT),
							kit.MDB_STATUS, tcp.OPEN,
							CONNECT, client,
						))
						m.Cmd(CONNECT, SESSION, kit.MDB_HASH, h)
					})

					m.Cmds(tcp.CLIENT, tcp.DIAL, arg)
				}},
				SESSION: {Name: "session hash", Help: "会话", Hand: func(m *ice.Message, arg ...string) {
					m.Richs(CONNECT, "", m.Option(kit.MDB_HASH), func(key string, value map[string]interface{}) {
						client, ok := value[CONNECT].(*ssh.Client)
						m.Assert(ok)

						h := m.Rich(SESSION, "", kit.Data(kit.MDB_STATUS, tcp.OPEN, CONNECT, key))

						if session, e := _ssh_sess(m, h, client); m.Assert(e) {
							session.Shell()
							session.Wait()
						}
					})
				}},

				mdb.PRUNES: {Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.PRUNES, CONNECT, "", mdb.HASH, kit.MDB_STATUS, tcp.CLOSE)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(mdb.FIELDS, kit.Select("time,hash,status,username,host,port", mdb.DETAIL, len(arg) > 0))
				m.Cmdy(mdb.SELECT, CONNECT, "", mdb.HASH, kit.MDB_HASH, arg)
				m.PushAction("会话")
			}},
		},
	}, nil)
}
