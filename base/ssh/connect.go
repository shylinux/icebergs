package ssh

import (
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
	_ssh_dial(m, func(c net.Conn) {
		// 保存界面
		fd := int(os.Stdin.Fd())
		if oldState, err := terminal.MakeRaw(fd); err == nil {
			defer terminal.Restore(fd, oldState)
		}

		// 设置宽高
		w, h, _ := terminal.GetSize(fd)
		c.Write([]byte(fmt.Sprintf("#height:%d,width:%d\n", h, w)))

		// 初始命令
		for _, item := range kit.Simple(m.Optionv(kit.MDB_LIST)) {
			m.Sleep("500ms")
			c.Write([]byte(item + "\n"))
		}

		m.Go(func() { io.Copy(c, os.Stdin) })
		io.Copy(os.Stdout, c)
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

	_ssh_conn(m, func(client *ssh.Client) {
		if l, e := net.Listen("unix", p); m.Assert(e) {
			defer func() { os.Remove(p) }()
			defer l.Close()

			m.Go(func() {
				for {
					c, e := l.Accept()
					if e != nil {
						break
					}

					func(c net.Conn) {
						w, h, _ := terminal.GetSize(int(os.Stdin.Fd()))
						buf := make([]byte, ice.MOD_BUFS)
						if n, e := c.Read(buf); m.Assert(e) {
							fmt.Sscanf(string(buf[:n]), "#height:%d,width:%d", &h, &w)
						}

						m.Go(func() {
							defer c.Close()

							session, e := client.NewSession()
							if e != nil {
								return
							}

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

		if c, e := net.Dial("unix", p); e == nil {
			cb(c) // 会话连接
		}
	}, arg...)
}
func _ssh_conn(m *ice.Message, cb func(*ssh.Client), arg ...string) {
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
		key, err := ssh.ParsePrivateKey([]byte(m.Cmdx(nfs.CAT, path.Join(os.Getenv(cli.HOME), m.Option(PRIVATE)))))
		return []ssh.Signer{key}, err
	}))
	methods = append(methods, ssh.PasswordCallback(func() (string, error) {
		return m.Option(aaa.PASSWORD), nil
	}))

	m.Option(kit.Keycb(tcp.DIAL), func(c net.Conn) {
		conn, chans, reqs, err := ssh.NewClientConn(c, m.Option(tcp.HOST)+":"+m.Option(tcp.PORT), &ssh.ClientConfig{
			User: m.Option(aaa.USERNAME), Auth: methods, BannerCallback: func(message string) error { return nil },
			HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil },
		})

		m.Assert(err)
		cb(ssh.NewClient(conn, chans, reqs))
	})
	m.Cmdy(tcp.CLIENT, tcp.DIAL, kit.MDB_TYPE, SSH, kit.MDB_NAME, m.Option(tcp.HOST),
		tcp.PORT, m.Option(tcp.PORT), tcp.HOST, m.Option(tcp.HOST), arg)
}

const CONNECT = "connect"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			CONNECT: {Name: CONNECT, Help: "连接", Value: kit.Data()},
		},
		Commands: map[string]*ice.Command{
			ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Richs(CONNECT, "", kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
					if value = kit.GetMeta(value); kit.Value(value, kit.MDB_STATUS) == tcp.OPEN {
						m.Cmd(CONNECT, tcp.DIAL, aaa.USERNAME, value[aaa.USERNAME], kit.MDB_HASH, key, value)
					}
				})
			}},
			CONNECT: {Name: "connect hash auto dial prunes", Help: "连接", Action: map[string]*ice.Action{
				tcp.OPEN: {Name: "open authfile= username=shy password= verfiy= host=shylinux.com port=22 private=.ssh/id_rsa", Help: "终端", Hand: func(m *ice.Message, arg ...string) {
					_ssh_open(m.OptionLoad(m.Option("authfile")), arg...)
					m.Echo("exit %v:%v\n", m.Option(tcp.HOST), m.Option(tcp.PORT))
				}},
				tcp.DIAL: {Name: "dial username=shy host=shylinux.com port=22 private=.ssh/id_rsa", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Go(func() {
						_ssh_conn(m, func(client *ssh.Client) {
							h := m.Option(kit.MDB_HASH)
							if h == "" {
								h = m.Rich(CONNECT, "", kit.Dict(
									aaa.USERNAME, m.Option(aaa.USERNAME),
									tcp.HOST, m.Option(tcp.HOST), tcp.PORT, m.Option(tcp.PORT),
									kit.MDB_STATUS, tcp.OPEN, CONNECT, client,
								))
							} else {
								m.Conf(CONNECT, kit.Keys(kit.MDB_HASH, h, CONNECT), client)
							}
							m.Cmd(CONNECT, SESSION, kit.MDB_HASH, h)
						}, arg...)
					})
					m.ProcessRefresh("300ms")
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, CONNECT, "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH))
				}},
				mdb.PRUNES: {Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
					m.Option(mdb.FIELDS, "time,hash,status,username,host,port")
					m.Cmdy(mdb.PRUNES, CONNECT, "", mdb.HASH, kit.MDB_STATUS, tcp.ERROR)
					m.Cmdy(mdb.PRUNES, CONNECT, "", mdb.HASH, kit.MDB_STATUS, tcp.CLOSE)
				}},

				SESSION: {Name: "session hash", Help: "会话", Hand: func(m *ice.Message, arg ...string) {
					var client *ssh.Client
					m.Richs(CONNECT, "", m.Option(kit.MDB_HASH), func(key string, value map[string]interface{}) {
						client, _ = value[CONNECT].(*ssh.Client)
					})

					h := m.Rich(SESSION, "", kit.Data(kit.MDB_STATUS, tcp.OPEN, CONNECT, m.Option(kit.MDB_HASH)))
					if session, e := _ssh_session(m, h, client); m.Assert(e) {
						session.Shell()
						session.Wait()
					}
					m.Echo(h)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Fields(len(arg) == 0, "time,hash,status,username,host,port")
				if m.Cmdy(mdb.SELECT, CONNECT, "", mdb.HASH, kit.MDB_HASH, arg); len(arg) == 0 {
					m.Table(func(index int, value map[string]string, head []string) {
						m.PushButton(kit.Select("", SESSION, value[kit.MDB_STATUS] == tcp.OPEN), mdb.REMOVE)
					})
				}
			}},
		},
	})
}
