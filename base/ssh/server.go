package ssh

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	kit "github.com/shylinux/toolkits"

	"bytes"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"github.com/kr/pty"
	"golang.org/x/crypto/ssh"
	"io"
	"net"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
	"unsafe"
)

type Winsize struct {
	Height uint16
	Width  uint16
	x      uint16
	y      uint16
}

func _ssh_size(fd uintptr, b []byte) {
	w := binary.BigEndian.Uint32(b)
	h := binary.BigEndian.Uint32(b[4:])

	ws := &Winsize{Width: uint16(w), Height: uint16(h)}
	syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TIOCSWINSZ), uintptr(unsafe.Pointer(ws)))
}
func _ssh_exec(m *ice.Message, cmd string, arg []string, env []string, tty io.ReadWriter, done func()) {
	m.Log_IMPORT("cmd", cmd, "arg", arg, "env", env)
	c := exec.Command(cmd, arg...)
	c.Env = env

	c.Stdout = tty
	c.Stdin = tty
	c.Stderr = tty

	err := c.Start()
	m.Assert(err)

	go func() {
		defer done()
		_, err := c.Process.Wait()
		m.Assert(err)
	}()
}

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			"dial": {Name: "dial", Help: "远程连接", Value: kit.Data()},
		},
		Commands: map[string]*ice.Command{
			"dial": {Name: "dial hash cmd auto 创建", Help: "守护进程", Meta: kit.Dict(), Action: map[string]*ice.Action{
				"create": {Name: "create", Help: "创建", List: kit.List(
					kit.MDB_INPUT, "text", "name", "hostport", "value", "shylinux.com:22",
					kit.MDB_INPUT, "text", "name", "username", "value", "shy",
					kit.MDB_INPUT, "password", "name", "password", "value", "",
				), Hand: func(m *ice.Message, arg ...string) {
					for i := 0; i < len(arg); i += 2 {
						m.Option(arg[i], arg[i+1])
					}

					connect, e := ssh.Dial("tcp", m.Option("hostport"), &ssh.ClientConfig{
						User: m.Option("username"), Auth: []ssh.AuthMethod{ssh.Password(m.Option("password"))},
						HostKeyCallback: ssh.InsecureIgnoreHostKey(),
					})
					m.Assert(e)

					h := m.Rich("dial", "", kit.Dict(
						"hostport", m.Option("hostport"),
						"username", m.Option("username"),
						"password", m.Option("password"),
						"connect", connect,
					))
					m.Echo(h)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					m.Option(mdb.FIELDS, "time,hash,hostport,username")
					m.Cmdy(mdb.SELECT, m.Prefix("dial"), "", mdb.HASH)
					return
				}

				m.Richs("dial", "", arg[0], func(key string, value map[string]interface{}) {
					connect, ok := value["connect"].(*ssh.Client)
					if !ok {
						connect, e := ssh.Dial("tcp", kit.Format(value["hostport"]), &ssh.ClientConfig{
							User: kit.Format(value["username"]), Auth: []ssh.AuthMethod{ssh.Password(kit.Format(value["password"]))},
							HostKeyCallback: ssh.InsecureIgnoreHostKey(),
						})
						m.Assert(e)
						value["connect"] = connect
					}

					session, e := connect.NewSession()
					m.Assert(e)
					defer session.Close()

					var b bytes.Buffer
					session.Stdout = &b

					err := session.Run(arg[1])
					m.Assert(err)

					m.Echo(b.String())
				})
			}},
			"listen": {Name: "listen host:port", Help: "守护进程", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Richs(aaa.USER, "", cli.UserName, func(key string, value map[string]interface{}) {
					value["publicKey"] = m.Cmdx(nfs.CAT, path.Join(os.Getenv("HOME"), ".ssh/id_rsa.pub"))
				})

				config := &ssh.ServerConfig{
					BannerCallback: func(conn ssh.ConnMetadata) string {
						m.Info("%s %s->%s", conn.User(), conn.RemoteAddr(), conn.LocalAddr())
						return "hello context world\n"
					},
					PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
						res := errors.New(ice.ErrNotAuth)
						m.TryCatch(m, true, func(m *ice.Message) {
							m.Richs(aaa.USER, "", conn.User(), func(k string, value map[string]interface{}) {
								if kit.Format(value["publicKey"]) == "" {
									return
								}
								if s, e := base64.StdEncoding.DecodeString(strings.Split(kit.Format(value["publicKey"]), " ")[1]); !m.Warn(e != nil, e) {
									if pub, e := ssh.ParsePublicKey([]byte(s)); !m.Warn(e != nil) {
										if bytes.Compare(pub.Marshal(), key.Marshal()) == 0 {
											res = nil
										}
									}
								}
							})
						})
						return &ssh.Permissions{Extensions: map[string]string{"method": "publickey"}}, res
					},
					// KeyboardInteractiveCallback: func(conn ssh.ConnMetadata, client ssh.KeyboardInteractiveChallenge) (*ssh.Permissions, error) {
					// 	m.Debug("what")
					// 	return &ssh.Permissions{Extensions: map[string]string{"key-id": "2"}}, nil
					// },
					PasswordCallback: func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
						res := errors.New(ice.ErrNotAuth)
						m.TryCatch(m, true, func(m *ice.Message) {
							m.Richs(aaa.USER, "", conn.User(), func(k string, value map[string]interface{}) {
								if string(password) == kit.Format(value["password"]) {
									res = nil
								}
							})
						})
						return &ssh.Permissions{Extensions: map[string]string{"method": "password"}}, res
					},
					AuthLogCallback: func(conn ssh.ConnMetadata, method string, err error) {
						m.Info("%s %s->%s %v", conn.User(), conn.RemoteAddr(), conn.LocalAddr(), err)
					},
				}

				if key, err := ssh.ParsePrivateKey([]byte(m.Cmdx(nfs.CAT, path.Join(os.Getenv("HOME"), ".ssh/id_rsa")))); m.Assert(err) {
					config.AddHostKey(key)
				}

				l, e := net.Listen("tcp", arg[0])
				m.Assert(e)
				m.Logs("listen", "address", l.Addr())

				for {
					c, e := l.Accept()
					if m.Warn(e != nil, e) {
						continue
					}

					go func(c net.Conn) {
						defer c.Close()
						defer m.Logs("disconn", "remote", c.RemoteAddr(), "local", c.LocalAddr())

						_, sessions, req, err := ssh.NewServerConn(c, config)
						if m.Warn(err != nil, err) {
							return
						}
						go ssh.DiscardRequests(req)

						for session := range sessions {
							channel, requests, err := session.Accept()
							if m.Warn(err != nil, err) {
								continue
							}

							go func(channel ssh.Channel, requests <-chan *ssh.Request) {
								defer m.Logs("dischan", "remote", c.RemoteAddr(), "local", c.LocalAddr())
								m.Logs("channel", "remote", c.RemoteAddr(), "local", c.LocalAddr())
								shell := kit.Select("bash", os.Getenv("SHELL"))
								list := []string{"PATH=" + os.Getenv("PATH")}

								tty, f, err := pty.Open()
								if m.Warn(err != nil, err) {
									return
								}
								defer f.Close()

								for request := range requests {
									m.Logs("request", "type", request.Type)

									switch request.Type {
									case "pty-req":
										termLen := request.Payload[3]
										termEnv := string(request.Payload[4 : termLen+4])
										_ssh_size(tty.Fd(), request.Payload[termLen+4:])
										list = append(list, "TERM="+termEnv)

									case "window-change":
										_ssh_size(tty.Fd(), request.Payload)

									case "env":
										var env struct {
											Name  string
											Value string
										}
										if err := ssh.Unmarshal(request.Payload, &env); err != nil {
											continue
										}
										list = append(list, env.Name+"="+env.Value)

									case "exec":
										_ssh_exec(m, shell, []string{"-c", string(request.Payload[4 : request.Payload[3]+4])}, list,
											channel, func() { channel.Close() })
									case "shell":
										_ssh_exec(m, shell, nil, list, f, func() { channel.Close() })
										go func() { io.Copy(channel, tty) }()
										go func() { io.Copy(tty, channel) }()
									}
									request.Reply(true, nil)
								}
							}(channel, requests)
						}
					}(c)
				}
			}},
		},
	}, nil)
}
