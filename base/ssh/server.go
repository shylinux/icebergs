package ssh

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/base/tcp"
	kit "github.com/shylinux/toolkits"

	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/kr/pty"
	"golang.org/x/crypto/ssh"
)

type Winsize struct {
	Height uint16
	Width  uint16
	x      uint16
	y      uint16
}

func _ssh_exec(m *ice.Message, cmd string, arg []string, env []string, tty io.ReadWriter, done func()) {
	m.Log_IMPORT("cmd", cmd, "arg", arg, "env", env)
	c := exec.Command(cmd, arg...)
	// c.Env = env

	c.Stdin = tty
	c.Stdout = tty
	c.Stderr = tty

	err := c.Start()
	m.Assert(err)

	m.Gos(m, func(m *ice.Message) {
		defer done()
		c.Process.Wait()
	})
}
func _ssh_close(m *ice.Message, c net.Conn, channel ssh.Channel) {
	defer channel.Close()
	channel.Write([]byte(m.Conf(PUBLIC, "meta.goodbye")))
}
func _ssh_trace(m *ice.Message, meta map[string]string, input io.Reader, output io.Writer, display io.Writer) {
	m.Gos(m, func(m *ice.Message) {
		i, buf := 0, make([]byte, 1024)
		for {
			n, e := input.Read(buf[i:])
			if e != nil {
				break
			}
			switch buf[i] {
			case '\r', '\n':
				cmd := strings.TrimSpace(string(buf[:i]))
				m.Log_IMPORT("hostname", meta["hostname"], "username", meta["username"], "buf", buf[:i+n])
				m.Conf(CONNECT, kit.Keys(kit.MDB_HASH, meta[CONNECT], "duration"), m.Format("cost"))
				m.Conf(SESSION, kit.Keys(kit.MDB_HASH, meta[SESSION], "cmd"), cmd)

				msg := m.Cmd(cmd).Table()
				res := strings.TrimSpace(strings.ReplaceAll(msg.Result(), "\n", "\r\n"))
				if len(res) > 0 {

					fmt.Fprintf(display, "\r\n")
					fmt.Fprintf(display, res)
					fmt.Fprintf(display, "\r\n")
					output.Write([]byte{21, 10})
				} else {
					output.Write(buf[i : i+n])
				}
				i = 0
			default:
				output.Write(buf[i : i+n])

				if i += n; i >= 1024 {
					i = 0
				}
			}
		}
	})
}
func _ssh_watch(m *ice.Message, meta map[string]string, input io.Reader, output io.Writer, display io.Writer) {
	m.Gos(m, func(m *ice.Message) {
		r, w := io.Pipe()
		bio := io.TeeReader(input, w)
		m.Gos(m, func(m *ice.Message) {
			i, buf := 0, make([]byte, 1024)
			for {
				n, e := bio.Read(buf[i:])
				if e != nil {
					break
				}
				switch buf[i] {
				case '\r', '\n':
					cmd := strings.TrimSpace(string(buf[:i]))
					m.Log_IMPORT(aaa.HOSTNAME, meta[aaa.HOSTNAME], aaa.USERNAME, meta[aaa.USERNAME], "cmd", cmd, "buf", buf[:i+n])
					m.Conf(CONNECT, kit.Keys(kit.MDB_HASH, meta[CONNECT], "duration"), m.Format("cost"))
					m.Conf(SESSION, kit.Keys(kit.MDB_HASH, meta[SESSION], "cmd"), cmd)

					m.Cmdy(mdb.INSERT, m.Prefix(COMMAND), "", mdb.LIST, aaa.HOSTNAME, meta[aaa.HOSTNAME], aaa.USERNAME, meta[aaa.USERNAME], "cmd", cmd)
					i = 0
				default:
					if i += n; i >= 1024 {
						i = 0
					}
				}
			}
		})
		io.Copy(output, r)
	})
}
func _ssh_handle(m *ice.Message, meta map[string]string, c net.Conn, channel ssh.Channel, requests <-chan *ssh.Request) {
	m.Logs(CHANNEL, aaa.HOSTPORT, c.RemoteAddr(), "->", c.LocalAddr())
	defer m.Logs("dischan", aaa.HOSTPORT, c.RemoteAddr(), "->", c.LocalAddr())

	shell := kit.Select("bash", os.Getenv("SHELL"))
	list := []string{"PATH=" + os.Getenv("PATH")}

	tty, f, err := pty.Open()
	if m.Warn(err != nil, err) {
		return
	}
	defer f.Close()

	h := m.Cmdx(mdb.INSERT, m.Prefix(SESSION), "", mdb.HASH, aaa.HOSTPORT, c.RemoteAddr().String(), kit.MDB_STATUS, "open", "tty", tty.Name())
	m.Richs(SESSION, "", h, func(key string, value map[string]interface{}) { value["channel"] = channel })
	meta[SESSION] = h

	for request := range requests {
		m.Logs(REQUEST, aaa.HOSTPORT, c.RemoteAddr(), "type", request.Type)

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
			if meta["username"] == "ssh" {
				m.I, m.O = f, f
				m.Render(ice.RENDER_VOID)
				m.Gos(m, func(m *ice.Message) {
					m.Cmdy(SOURCE, tty.Name())
					_ssh_close(m, c, channel)
				})
			} else {
				_ssh_exec(m, shell, nil, list, f, func() {
					defer m.Cmd(mdb.MODIFY, m.Prefix(SESSION), "", mdb.HASH, kit.MDB_HASH, h, kit.MDB_STATUS, "close")
					_ssh_close(m, c, channel)
				})
			}

			m.Gos(m, func(m *ice.Message) { io.Copy(channel, tty) })
			_ssh_watch(m, meta, channel, tty, channel)
			// _ssh_trace(m, meta, channel, tty, channel)
		}
		request.Reply(true, nil)
	}
}
func _ssh_listen(m *ice.Message, hostport string) {
	h := m.Cmdx(mdb.INSERT, m.Prefix(LISTEN), "", mdb.HASH, aaa.HOSTPORT, hostport, kit.MDB_STATUS, "listen")
	defer m.Cmd(mdb.MODIFY, m.Prefix(LISTEN), "", mdb.HASH, kit.MDB_HASH, h, kit.MDB_STATUS, "close")

	config := _ssh_config(m)

	l, e := net.Listen("tcp", hostport)
	m.Assert(e)
	defer l.Close()
	m.Logs(LISTEN, ADDRESS, l.Addr())

	for {
		c, e := l.Accept()
		if m.Warn(e != nil, e) {
			continue
		}

		func(c net.Conn) {
			m.Gos(m.Spawn(), func(msg *ice.Message) {
				defer c.Close()

				m.Logs(CONNECT, aaa.HOSTPORT, c.RemoteAddr(), "->", c.LocalAddr())
				defer m.Logs("disconn", aaa.HOSTPORT, c.RemoteAddr(), "->", c.LocalAddr())

				sc, sessions, req, err := ssh.NewServerConn(c, config)
				if m.Warn(err != nil, err) {
					return
				}

				hostname := sc.Permissions.Extensions["hostname"]
				username := sc.Permissions.Extensions["username"]
				begin := time.Now()
				h := m.Cmdx(mdb.INSERT, m.Prefix(CONNECT), "", mdb.HASH, aaa.HOSTPORT, c.RemoteAddr().String(), kit.MDB_STATUS, "connect", "hostname", hostname, "username", username)
				defer m.Cmd(mdb.MODIFY, m.Prefix(CONNECT), "", mdb.HASH, kit.MDB_HASH, h, kit.MDB_STATUS, "close", "close_time", time.Now().Format(ice.MOD_TIME), "duration", time.Now().Sub(begin).String())
				sc.Permissions.Extensions[CONNECT] = h

				m.Gos(m, func(m *ice.Message) {
					ssh.DiscardRequests(req)
				})

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
			})
		}(c)
	}
}
func _ssh_config(m *ice.Message) *ssh.ServerConfig {
	config := &ssh.ServerConfig{
		BannerCallback: func(conn ssh.ConnMetadata) string {
			m.Log_IMPORT(aaa.HOSTPORT, conn.RemoteAddr(), aaa.USERNAME, conn.User())
			return m.Conf(PUBLIC, "meta.welcome")
		},
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			meta, res := map[string]string{"username": conn.User()}, errors.New(ice.ErrNotAuth)
			if tcp.IPIsLocal(m, strings.Split(conn.RemoteAddr().String(), ":")[0]) {
				m.Log_AUTH(aaa.HOSTPORT, conn.RemoteAddr(), aaa.USERNAME, conn.User())
				res = nil
			} else {
				m.Richs(PUBLIC, "", kit.MDB_FOREACH, func(k string, value map[string]interface{}) {
					if !strings.HasPrefix(kit.Format(value[kit.MDB_NAME]), conn.User()+"@") {
						return
					}
					if s, e := base64.StdEncoding.DecodeString(kit.Format(value[kit.MDB_TEXT])); !m.Warn(e != nil, e) {
						if pub, e := ssh.ParsePublicKey([]byte(s)); !m.Warn(e != nil) {
							if bytes.Compare(pub.Marshal(), key.Marshal()) == 0 {
								m.Log_AUTH(aaa.HOSTPORT, conn.RemoteAddr(), aaa.USERNAME, conn.User(), "publickey", value[kit.MDB_NAME])
								meta["hostname"] = kit.Format(value[kit.MDB_NAME])
								res = nil
							}
						}
					}
				})
			}
			return &ssh.Permissions{Extensions: meta}, res
		},
		PasswordCallback: func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
			meta, res := map[string]string{"username": conn.User()}, errors.New(ice.ErrNotAuth)
			m.Richs(aaa.USER, "", conn.User(), func(k string, value map[string]interface{}) {
				if string(password) == kit.Format(value[aaa.PASSWORD]) {
					m.Log_AUTH(aaa.HOSTPORT, conn.RemoteAddr(), aaa.USERNAME, conn.User(), aaa.PASSWORD, strings.Repeat("*", len(kit.Format(value[aaa.PASSWORD]))))
					res = nil
				}
			})
			return &ssh.Permissions{Extensions: meta}, res
		},
	}

	if key, err := ssh.ParsePrivateKey([]byte(m.Cmdx(nfs.CAT, path.Join(os.Getenv("HOME"), m.Conf(PUBLIC, "meta.private"))))); m.Assert(err) {
		config.AddHostKey(key)
	}
	return config
}
func _ssh_dial(m *ice.Message, username, hostport string) (*ssh.Client, error) {
	methods := []ssh.AuthMethod{}
	if key, e := ssh.ParsePrivateKey([]byte(m.Cmdx(nfs.CAT, path.Join(os.Getenv("HOME"), m.Conf(PUBLIC, "meta.private"))))); !m.Warn(e != nil) {
		methods = append(methods, ssh.PublicKeys(key))
	} else {
		return nil, e
	}

	connect, e := ssh.Dial("tcp", hostport, &ssh.ClientConfig{User: username, Auth: methods,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			m.Logs(CONNECT, "hostname", hostname, aaa.HOSTPORT, remote.String())
			return nil
		},
	})
	return connect, e
}

const (
	ADDRESS = "address"
	CONNECT = "connect"
	CHANNEL = "channel"
	SESSION = "session"
	REQUEST = "request"
	COMMAND = "command"
)
const (
	METHOD = "method"
	PUBLIC = "public"
	LISTEN = "listen"
	DIAL   = "dial"
)

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			PUBLIC: {Name: PUBLIC, Help: "公钥", Value: kit.Data(
				"private", ".ssh/id_rsa", "public", ".ssh/id_rsa.pub",
				"welcome", "\r\nwelcome to context world\r\n",
				"goodbye", "\r\ngoodbye of context world\r\n",
				kit.MDB_SHORT, kit.MDB_TEXT,
			)},

			LISTEN: {Name: LISTEN, Help: "服务", Value: kit.Data(kit.MDB_SHORT, aaa.HOSTPORT,
				kit.MDB_FIELD, "time,hash,hostport,status",
			)},
			CONNECT: {Name: CONNECT, Help: "连接", Value: kit.Data(
				kit.MDB_FIELD, "time,hash,hostport,status,duration,close_time,hostname,username",
			)},
			SESSION: {Name: SESSION, Help: "会话", Value: kit.Data(
				kit.MDB_FIELD, "time,hash,hostport,status,tty,cmd",
			)},
			COMMAND: {Name: COMMAND, Help: "命令", Value: kit.Data(
				kit.MDB_FIELD, "time,id,username,hostname,cmd",
			)},

			DIAL: {Name: DIAL, Help: "连接", Value: kit.Data(
				kit.MDB_FIELD, "time,hash,hostport,username",
			)},
		},
		Commands: map[string]*ice.Command{
			PUBLIC: {Name: "public hash=auto auto 添加 导出 导入", Help: "公钥", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create publickey:textarea", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					ls := kit.Split(m.Option("publickey"))
					m.Cmdy(mdb.INSERT, m.Prefix(PUBLIC), "", mdb.HASH, kit.MDB_TYPE, ls[0],
						kit.MDB_NAME, ls[len(ls)-1], kit.MDB_TEXT, strings.Join(ls[1:len(ls)-1], "+"))
				}},
				mdb.DELETE: {Name: "delete", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, m.Prefix(PUBLIC), "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH))
				}},
				mdb.EXPORT: {Name: "export file=.ssh/authorized_keys", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
					list := []string{}
					if m.Cmd(mdb.SELECT, m.Prefix(PUBLIC), "", mdb.HASH).Table(func(index int, value map[string]string, head []string) {
						list = append(list, fmt.Sprintf("%s %s %s", value[kit.MDB_TYPE], value[kit.MDB_TEXT], value[kit.MDB_NAME]))
					}); len(list) > 0 {
						m.Cmdy(nfs.SAVE, path.Join(os.Getenv("HOME"), m.Option(kit.MDB_FILE)), strings.Join(list, "\n")+"\n")
					}
				}},
				mdb.IMPORT: {Name: "import file=.ssh/authorized_keys", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
					p := path.Join(os.Getenv("HOME"), m.Option(kit.MDB_FILE))
					for _, pub := range strings.Split(m.Cmdx(nfs.CAT, p), "\n") {
						if len(pub) > 10 {
							m.Cmd(PUBLIC, mdb.CREATE, "publickey", pub)
						}
					}
					m.Echo(p)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) > 0 {
					m.Option(mdb.FIELDS, mdb.DETAIL)
				} else {
					defer m.PushAction("删除")
				}
				m.Cmdy(mdb.SELECT, m.Prefix(PUBLIC), "", mdb.HASH, kit.MDB_HASH, arg)
			}},

			LISTEN: {Name: "listen hash=auto auto", Help: "服务", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create port=9030", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
					m.Gos(m, func(m *ice.Message) { _ssh_listen(m, ":"+m.Option("port")) })
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if m.Option(mdb.FIELDS, m.Conf(LISTEN, kit.META_FIELD)); len(arg) > 0 {
					m.Option(mdb.FIELDS, mdb.DETAIL)
				}
				m.Cmdy(mdb.SELECT, m.Prefix(LISTEN), "", mdb.HASH, kit.MDB_HASH, arg)
			}},
			CONNECT: {Name: "connect hash=auto auto 清理", Help: "连接", Action: map[string]*ice.Action{
				mdb.PRUNES: {Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.PRUNES, m.Prefix(CONNECT), "", mdb.HASH, kit.MDB_STATUS, "close")
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if m.Option(mdb.FIELDS, m.Conf(CONNECT, kit.META_FIELD)); len(arg) > 0 {
					m.Option(mdb.FIELDS, mdb.DETAIL)
				}
				m.Cmdy(mdb.SELECT, m.Prefix(CONNECT), "", mdb.HASH, kit.MDB_HASH, arg)
			}},
			SESSION: {Name: "session hash auto 清理", Help: "会话", Action: map[string]*ice.Action{
				mdb.PRUNES: {Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.PRUNES, m.Prefix(SESSION), "", mdb.HASH, kit.MDB_STATUS, "close")
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if m.Option(mdb.FIELDS, m.Conf(SESSION, kit.META_FIELD)); len(arg) > 0 {
					m.Option(mdb.FIELDS, mdb.DETAIL)
				}
				m.Cmdy(mdb.SELECT, m.Prefix(SESSION), "", mdb.HASH, kit.MDB_HASH, arg)
			}},
			COMMAND: {Name: "command id=auto auto", Help: "命令", Action: map[string]*ice.Action{
				mdb.PRUNES: {Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.PRUNES, m.Prefix(CONNECT), "", mdb.HASH, kit.MDB_STATUS, "close")
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if m.Option(mdb.FIELDS, m.Conf(COMMAND, kit.META_FIELD)); len(arg) > 0 {
					m.Option(mdb.FIELDS, mdb.DETAIL)
				}
				m.Cmdy(mdb.SELECT, m.Prefix(COMMAND), "", mdb.LIST, kit.MDB_ID, arg)
			}},

			DIAL: {Name: "dial hash=auto auto 登录 cmd:textarea=pwd", Help: "连接", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create username=shy hostname=shylinux.com port=22", Help: "登录", Hand: func(m *ice.Message, arg ...string) {
					if connect, e := _ssh_dial(m, m.Option(aaa.USERNAME),
						m.Option(aaa.HOSTPORT, m.Option("hostname")+":"+m.Option("port"))); m.Assert(e) {
						h := m.Rich(DIAL, "", kit.Dict(
							aaa.USERNAME, m.Option(aaa.USERNAME), aaa.HOSTPORT, m.Option(aaa.HOSTPORT), CONNECT, connect,
						))
						m.Echo(h)
					}
				}},
				mdb.DELETE: {Name: "delete", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, m.Prefix(DIAL), "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 || arg[0] == "" {
					m.Option(mdb.FIELDS, m.Conf(DIAL, kit.META_FIELD))
					m.Cmdy(mdb.SELECT, m.Prefix(DIAL), "", mdb.HASH)
					m.PushAction("删除")
					return
				}

				m.Richs(DIAL, "", arg[0], func(key string, value map[string]interface{}) {
					connect, ok := value[CONNECT].(*ssh.Client)
					if !ok {
						if c, e := _ssh_dial(m, kit.Format(value[aaa.USERNAME]), kit.Format(value[aaa.HOSTPORT])); m.Assert(e) {
							connect, value[CONNECT] = c, c
						}
					}

					session, e := connect.NewSession()
					m.Assert(e)
					defer session.Close()

					var b bytes.Buffer
					session.Stdout = &b

					m.Assert(session.Run(arg[1]))
					m.Echo(b.String())
				})
			}},
		},
	}, nil)
}
