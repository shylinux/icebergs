package ssh

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/base/tcp"
	kit "github.com/shylinux/toolkits"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"

	"encoding/json"
	"net"
	"os"
	"path"
	"strings"
)

func _ssh_conn(m *ice.Message, conn net.Conn, username, hostport string) (*ssh.Client, error) {
	methods := []ssh.AuthMethod{}
	methods = append(methods, ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) (res []string, err error) {
		for _, k := range questions {
			switch strings.TrimSpace(strings.ToLower(k)) {
			case "verification code:":
				res = append(res, aaa.TOTP_GET(m.Option("verify"), 6, 30))
			case "password:":
				res = append(res, m.Option(aaa.PASSWORD))
			default:
			}
		}
		m.Debug("question: %v res: %d", questions, len(res))
		return
	}))

	methods = append(methods, ssh.PublicKeysCallback(func() ([]ssh.Signer, error) {
		key, err := ssh.ParsePrivateKey([]byte(m.Cmdx(nfs.CAT, path.Join(os.Getenv("HOME"), m.Option("private")))))
		m.Debug("publickeys")
		return []ssh.Signer{key}, err
	}))
	methods = append(methods, ssh.PasswordCallback(func() (string, error) {
		m.Debug("password")
		return m.Option(aaa.PASSWORD), nil
	}))

	c, chans, reqs, err := ssh.NewClientConn(conn, hostport, &ssh.ClientConfig{
		User: username, Auth: methods, HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			m.Logs(CONNECT, tcp.HOSTNAME, hostname, tcp.HOSTPORT, remote.String())
			return nil
		},
	})

	return ssh.NewClient(c, chans, reqs), err
}

const CONNECT = "connect"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			CONNECT: {Name: CONNECT, Help: "连接", Value: kit.Data()},
		},
		Commands: map[string]*ice.Command{
			CONNECT: {Name: "connect hash auto dial prunes", Help: "连接", Action: map[string]*ice.Action{
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

					m.Cmds(tcp.CLIENT, tcp.DIAL, kit.MDB_TYPE, SSH, kit.MDB_NAME, m.Option(aaa.USERNAME), tcp.PORT, m.Option(tcp.PORT), tcp.HOST, m.Option(tcp.HOST))
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

				"open": {Name: "open authfile= username=shy password= verfiy= host=shylinux.com port=22 private=.ssh/id_rsa", Help: "终端", Hand: func(m *ice.Message, arg ...string) {
					if f, e := os.Open(m.Option("authfile")); e == nil {
						defer f.Close()

						var data interface{}
						json.NewDecoder(f).Decode(&data)

						kit.Fetch(data, func(key string, value string) { m.Option(key, value) })
					}

					m.Option(tcp.DIAL_CB, func(c net.Conn) {
						client, e := _ssh_conn(m, c, m.Option(aaa.USERNAME), m.Option(tcp.HOST)+":"+m.Option(tcp.PORT))
						m.Assert(e)

						session, e := client.NewSession()
						m.Assert(e)

						fd := int(os.Stdin.Fd())
						oldState, err := terminal.MakeRaw(fd)
						if err != nil {
							panic(err)
						}
						defer terminal.Restore(fd, oldState)

						w, h, e := terminal.GetSize(fd)
						m.Assert(e)

						fd1 := int(os.Stdout.Fd())
						oldState1, err := terminal.MakeRaw(fd1)
						if err != nil {
							panic(err)
						}
						defer terminal.Restore(fd1, oldState1)

						session.Stdin = os.Stdin
						session.Stdout = os.Stdout
						session.Stderr = os.Stderr

						modes := ssh.TerminalModes{
							ssh.ECHO:          1,
							ssh.TTY_OP_ISPEED: 14400,
							ssh.TTY_OP_OSPEED: 14400,
						}

						session.RequestPty(os.Getenv("TERM"), h, w, modes)
						session.Shell()
						session.Wait()
					})

					m.Cmdy(tcp.CLIENT, tcp.DIAL, kit.MDB_TYPE, "ssh", kit.MDB_NAME, m.Option(tcp.HOST),
						tcp.PORT, m.Option(tcp.PORT), tcp.HOST, m.Option(tcp.HOST), arg)

					m.Echo("exit %s\n", m.Option(tcp.HOST))
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, CONNECT, "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH))
				}},
				mdb.PRUNES: {Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.PRUNES, CONNECT, "", mdb.HASH, kit.MDB_STATUS, tcp.CLOSE)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(mdb.FIELDS, kit.Select("time,hash,status,username,host,port", mdb.DETAIL, len(arg) > 0))
				if m.Cmdy(mdb.SELECT, CONNECT, "", mdb.HASH, kit.MDB_HASH, arg); len(arg) == 0 {
					m.Table(func(index int, value map[string]string, head []string) {
						m.PushButton(kit.Select("", mdb.REMOVE, value[kit.MDB_STATUS] == tcp.CLOSE))
					})
				}
			}},
		},
	}, nil)
}
