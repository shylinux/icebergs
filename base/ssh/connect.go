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
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"strings"
)

func _ssh_tick(m *ice.Message, pw io.Writer) {
	if m.Option("tick") == "" {
		return
	}
	m.Go(func() {
		for {
			m.Sleep(m.Option("tick"))
			pw.Write([]byte("\n"))
		}
	})
}
func _ssh_password(m *ice.Message, file string) {
	if f, e := os.Open(file); e == nil {
		defer f.Close()

		var data interface{}
		json.NewDecoder(f).Decode(&data)

		kit.Fetch(data, func(key string, value string) { m.Option(key, value) })
	}
}
func _ssh_stream(m *ice.Message, stdin *os.File) (io.Reader, io.Writer) {
	pr, pw := io.Pipe()
	m.Go(func() {
		buf := make([]byte, 1024)
		for {
			if n, e := stdin.Read(buf); m.Assert(e) {
				pw.Write(buf[:n])
			}
		}
	})
	return pr, pw
}
func _ssh_store(stdio *os.File) func() {
	fd := int(stdio.Fd())
	oldState, err := terminal.MakeRaw(fd)
	if err != nil {
		panic(err)
	}
	return func() { terminal.Restore(fd, oldState) }
}

func _ssh_session(m *ice.Message, client *ssh.Client, w, h int, stdin io.Reader, stdout, stderr io.Writer) *ssh.Session {
	session, e := client.NewSession()
	m.Assert(e)

	session.Stdin = stdin
	session.Stdout = stdout
	session.Stderr = stderr

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	session.RequestPty(os.Getenv("TERM"), h, w, modes)
	session.Shell()
	return session
}
func _ssh_init(m *ice.Message, pw io.Writer) {
	for _, k := range []string{"one", "two"} {
		if m.Sleep("100ms"); m.Option(k) != "" {
			pw.Write([]byte(m.Option(k) + "\n"))
		}
	}
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
		key, err := ssh.ParsePrivateKey([]byte(m.Cmdx(nfs.CAT, path.Join(os.Getenv("HOME"), m.Option("private")))))
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
				tcp.DIAL: {Name: "dial username=shy host=shylinux.com port=22 private=.ssh/id_rsa", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Option(tcp.DIAL_CB, func(c net.Conn) {
						client := _ssh_conn(m, c, kit.Select("shy", m.Option(aaa.USERNAME)),
							kit.Select("shylinux.com", m.Option(tcp.HOST))+":"+kit.Select("22", m.Option(tcp.PORT)),
						)

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

				"open": {Name: "open authfile= username=shy password= verfiy= host=shylinux.com port=22 private=.ssh/id_rsa tick=", Help: "终端", Hand: func(m *ice.Message, arg ...string) {
					var client *ssh.Client
					w, h, _ := terminal.GetSize(int(os.Stdin.Fd()))

					_ssh_password(m, m.Option("authfile"))
					p := path.Join(os.Getenv("HOME"), ".ssh/", fmt.Sprintf("%s@%s", m.Option("username"), m.Option("host")))
					if _, e := os.Stat(p); e == nil {
						if c, e := net.Dial("unix", p); e == nil {

							pr, pw := _ssh_stream(m, os.Stdin)
							defer _ssh_store(os.Stdout)()
							defer _ssh_store(os.Stdin)()

							c.Write([]byte(fmt.Sprintf("height:%d,width:%d\n", h, w)))

							m.Go(func() { io.Copy(c, pr) })
							_ssh_init(m, pw)
							m.Echo("logout\n")
							io.Copy(os.Stdout, c)
							return
						} else {
							os.Remove(p)
						}
					}

					if l, e := net.Listen("unix", p); m.Assert(e) {
						defer func() { os.Remove(p) }()
						defer l.Close()

						m.Go(func() {
							for {
								if c, e := l.Accept(); e == nil {
									buf := make([]byte, 1024)
									if n, e := c.Read(buf); m.Assert(e) {
										fmt.Sscanf(string(buf[:n]), "height:%d,width:%d", &h, &w)
									}

									session := _ssh_session(m, client, w, h, c, c, c)
									func(session *ssh.Session) {
										m.Go(func() {
											defer c.Close()
											session.Wait()
										})
									}(session)
								} else {
									break
								}
							}
						})
					}

					m.Option(tcp.DIAL_CB, func(c net.Conn) {
						client = _ssh_conn(m, c, m.Option(aaa.USERNAME), m.Option(tcp.HOST)+":"+m.Option(tcp.PORT))

						pr, pw := _ssh_stream(m, os.Stdin)
						defer _ssh_store(os.Stdout)()
						defer _ssh_store(os.Stdin)()

						session := _ssh_session(m, client, w, h, pr, os.Stdout, os.Stderr)
						_ssh_init(m, pw)
						_ssh_tick(m, pw)
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
	})
}
