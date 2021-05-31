package ssh

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"strings"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/gdb"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/base/tcp"
	kit "github.com/shylinux/toolkits"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

func _ssh_open(m *ice.Message, arg ...string) {
	// 加载配置
	if f, e := os.Open(m.Option("authfile")); e == nil {
		defer f.Close()

		var data interface{}
		json.NewDecoder(f).Decode(&data)

		kit.Fetch(data, func(key string, value interface{}) { m.Option(key, kit.Simple(value)) })
	}

	_ssh_dial(m, func(c net.Conn) {
		// 保存界面
		fd := int(os.Stdin.Fd())
		if oldState, err := terminal.MakeRaw(fd); err == nil {
			defer terminal.Restore(fd, oldState)
		}

		// 设置宽高
		w, h, _ := terminal.GetSize(fd)
		c.Write([]byte(fmt.Sprintf("height:%d,width:%d\n", h, w)))

		// 初始命令
		for _, item := range kit.Simple(m.Optionv("list")) {
			m.Sleep("10ms")
			c.Write([]byte(item + "\n"))
		}

		m.Go(func() { io.Copy(os.Stdout, c) })
		io.Copy(c, os.Stdin)
	}, arg...)
}
func _ssh_dial(m *ice.Message, cb func(net.Conn), arg ...string) {
	p := path.Join(os.Getenv(cli.HOME), ".ssh/", fmt.Sprintf("%s@%s:%s", m.Option(aaa.USERNAME), m.Option(tcp.HOST), m.Option(tcp.PORT)))
	if _, e := os.Stat(p); e == nil {
		if c, e := net.Dial("unix", p); e == nil {
			cb(c) // 会话连接
			return
		}
		os.Remove(p)
	}

	var client *ssh.Client
	if l, e := net.Listen("unix", p); m.Assert(e) {
		defer func() { os.Remove(p) }()
		defer l.Close()

		m.Go(func() {
			for {
				c, e := l.Accept()
				m.Assert(e)

				func(c net.Conn) {
					w, h, _ := terminal.GetSize(int(os.Stdin.Fd()))
					buf := make([]byte, ice.MOD_BUFS)
					if n, e := c.Read(buf); m.Assert(e) {
						fmt.Sscanf(string(buf[:n]), "height:%d,width:%d", &h, &w)
					}

					m.Go(func() {
						defer c.Close()

						session, e := client.NewSession()
						m.Assert(e)

						session.Stdin = c
						session.Stdout = c
						session.Stderr = c

						session.RequestPty(os.Getenv("TERM"), h, w, ssh.TerminalModes{
							ssh.ECHO:          1,
							ssh.TTY_OP_ISPEED: 14400,
							ssh.TTY_OP_OSPEED: 14400,
						})

						gdb.SignalNotify(m, 28, func() {
							w, h, _ := terminal.GetSize(int(os.Stdin.Fd()))
							session.WindowChange(h, w)
						})

						session.Shell()
						session.Wait()
					})
				}(c)
			}
		})
	}

	m.Option(kit.Keycb(tcp.DIAL), func(c net.Conn) {
		client = _ssh_conn(m, c, m.Option(aaa.USERNAME), m.Option(tcp.HOST)+":"+m.Option(tcp.PORT))

		if c, e := net.Dial("unix", p); e == nil {
			cb(c) // 会话连接
		}
	})
	m.Cmdy(tcp.CLIENT, tcp.DIAL, kit.MDB_TYPE, SSH, kit.MDB_NAME, m.Option(tcp.HOST),
		tcp.PORT, m.Option(tcp.PORT), tcp.HOST, m.Option(tcp.HOST), arg)
}
func _ssh_conn(m *ice.Message, conn net.Conn, username, hostport string) *ssh.Client {
	methods := []ssh.AuthMethod{}
	methods = append(methods, ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) (res []string, err error) {
		for _, q := range questions {
			p := strings.TrimSpace(strings.ToLower(q))
			switch {
			case strings.HasSuffix(p, "verification code:"):
				if verify := m.Option("verify"); verify == "" {
					fmt.Printf(q)
					fmt.Scanf("%s\n", &verify)

					res = append(res, verify)
				} else {
					res = append(res, aaa.TOTP_GET(verify, 6, 30))
				}
			case strings.HasSuffix(p, "password:"):
				res = append(res, m.Option(aaa.PASSWORD))
			default:
			}
		}
		return
	}))

	methods = append(methods, ssh.PublicKeysCallback(func() ([]ssh.Signer, error) {
		key, err := ssh.ParsePrivateKey([]byte(m.Cmdx(nfs.CAT, path.Join(os.Getenv(cli.HOME), m.Option("private")))))
		return []ssh.Signer{key}, err
	}))
	methods = append(methods, ssh.PasswordCallback(func() (string, error) {
		return m.Option(aaa.PASSWORD), nil
	}))

	c, chans, reqs, err := ssh.NewClientConn(conn, hostport, &ssh.ClientConfig{
		User: username, Auth: methods, BannerCallback: func(message string) error { return nil },
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil },
	})

	m.Assert(err)
	return ssh.NewClient(c, chans, reqs)
}

const CONNECT = "connect"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			CONNECT: {Name: CONNECT, Help: "连接", Value: kit.Data()},
		},
		Commands: map[string]*ice.Command{
			CONNECT: {Name: "connect hash auto dial prunes", Help: "连接", Action: map[string]*ice.Action{
				tcp.OPEN: {Name: "open authfile= username=shy password= verfiy= host=shylinux.com port=22 private=.ssh/id_rsa", Help: "终端", Hand: func(m *ice.Message, arg ...string) {
					_ssh_open(m, arg...)
				}},
				tcp.DIAL: {Name: "dial username=shy host=shylinux.com port=22 private=.ssh/id_rsa", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Option(kit.Keycb(tcp.DIAL), func(c net.Conn) {
						client := _ssh_conn(m, c, kit.Select("shy", m.Option(aaa.USERNAME)),
							kit.Select("shylinux.com", m.Option(tcp.HOST))+":"+kit.Select("22", m.Option(tcp.PORT)),
						)

						h := m.Rich(CONNECT, "", kit.Dict(
							aaa.USERNAME, m.Option(aaa.USERNAME),
							tcp.HOST, m.Option(tcp.HOST), tcp.PORT, m.Option(tcp.PORT),
							kit.MDB_STATUS, tcp.OPEN, CONNECT, client,
						))
						m.Cmd(CONNECT, SESSION, kit.MDB_HASH, h)
						m.Echo(h)
					})

					m.Cmds(tcp.CLIENT, tcp.DIAL, kit.MDB_TYPE, SSH, kit.MDB_NAME, m.Option(aaa.USERNAME),
						tcp.PORT, m.Option(tcp.PORT), tcp.HOST, m.Option(tcp.HOST))
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, CONNECT, "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH))
				}},
				mdb.PRUNES: {Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.PRUNES, CONNECT, "", mdb.HASH, kit.MDB_STATUS, tcp.ERROR)
					m.Cmdy(mdb.PRUNES, CONNECT, "", mdb.HASH, kit.MDB_STATUS, tcp.CLOSE)
				}},

				SESSION: {Name: "session hash", Help: "会话", Hand: func(m *ice.Message, arg ...string) {
					m.Richs(CONNECT, "", m.Option(kit.MDB_HASH), func(key string, value map[string]interface{}) {
						client, ok := value[CONNECT].(*ssh.Client)
						m.Assert(ok)

						h := m.Rich(SESSION, "", kit.Data(kit.MDB_STATUS, tcp.OPEN, CONNECT, key))

						if session, e := _ssh_session(m, h, client); m.Assert(e) {
							session.Shell()
							session.Wait()
						}
					})
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Fields(len(arg) == 0, "time,hash,status,username,host,port")
				if m.Cmdy(mdb.SELECT, CONNECT, "", mdb.HASH, kit.MDB_HASH, arg); len(arg) == 0 {
					m.Table(func(index int, value map[string]string, head []string) {
						m.PushButton(kit.Select("", mdb.REMOVE, value[kit.MDB_STATUS] == tcp.CLOSE))
					})
				}
			}},
		},
	})
}
