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

func _ssh_conn(m *ice.Message, conn net.Conn, username, hostport string) (*ssh.Client, error) {
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
						client, e := _ssh_conn(m, c, kit.Select("shy", m.Option(aaa.USERNAME)),
							kit.Select("shylinux.com", m.Option(tcp.HOST))+":"+kit.Select("22", m.Option(tcp.PORT)),
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
						m.Echo(h)
					})

					m.Cmds(tcp.CLIENT, tcp.DIAL, arg)
					m.Sleep("100ms")
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

				"open": {Name: "dial username=shy host=shylinux.com port=22 private=.ssh/id_rsa", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Option(tcp.DIAL_CB, func(c net.Conn) {
						client, e := _ssh_conn(m, c, kit.Select("shy", m.Option(aaa.USERNAME)),
							kit.Select("shylinux.com", m.Option(tcp.HOST))+":"+kit.Select("22", m.Option(tcp.PORT)),
						)
						m.Assert(e)

						m.Debug("what")
						m.Debug("some")
						session, e := client.NewSession()
						m.Assert(e)

						session.Stdin = os.Stdin
						session.Stdout = os.Stdout
						session.Stderr = os.Stderr
						session.Start("/bin/bash")
						m.Debug("what")
						m.Debug("some")
					})

					m.Cmd(tcp.CLIENT, tcp.DIAL, tcp.PORT, m.Option(tcp.PORT), tcp.HOST, m.Option(tcp.HOST), arg)
				}},
				mdb.DELETE: {Name: "delete", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, CONNECT, "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH))
				}},
				mdb.PRUNES: {Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.PRUNES, CONNECT, "", mdb.HASH, kit.MDB_STATUS, tcp.CLOSE)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(mdb.FIELDS, kit.Select("time,hash,status,username,host,port", mdb.DETAIL, len(arg) > 0))
				if m.Cmdy(mdb.SELECT, CONNECT, "", mdb.HASH, kit.MDB_HASH, arg); len(arg) == 0 {
					m.Table(func(index int, value map[string]string, head []string) {
						m.PushButton(kit.Select("", "删除", value[kit.MDB_STATUS] == tcp.CLOSE))
					})
				}
			}},
		},
	}, nil)
}
