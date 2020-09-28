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

	pty, tty, err := pty.Open()
	if m.Warn(err != nil, err) {
		return
	}
	defer tty.Close()

	h := m.Cmdx(mdb.INSERT, m.Prefix(SESSION), "", mdb.HASH, aaa.HOSTPORT, c.RemoteAddr().String(), kit.MDB_STATUS, "open", "pty", pty.Name())
	m.Richs(SESSION, "", h, func(key string, value map[string]interface{}) { value["channel"] = channel })
	meta[SESSION] = h

	for request := range requests {
		m.Logs(REQUEST, aaa.HOSTPORT, c.RemoteAddr(), "type", request.Type)

		switch request.Type {
		case "pty-req":
			termLen := request.Payload[3]
			termEnv := string(request.Payload[4 : termLen+4])
			_ssh_size(pty.Fd(), request.Payload[termLen+4:])
			list = append(list, "TERM="+termEnv)

		case "window-change":
			_ssh_size(pty.Fd(), request.Payload)

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
				m.I, m.O = tty, tty
				m.Render(ice.RENDER_VOID)
				m.Gos(m, func(m *ice.Message) {
					m.Cmdy(SOURCE, pty.Name())
					_ssh_close(m, c, channel)
				})
			} else {
				_ssh_exec(m, shell, nil, list, tty, func() {
					defer m.Cmd(mdb.MODIFY, m.Prefix(SESSION), "", mdb.HASH, kit.MDB_HASH, h, kit.MDB_STATUS, "close")
					_ssh_close(m, c, channel)
				})
			}

			m.Gos(m, func(m *ice.Message) { io.Copy(channel, pty) })
			_ssh_watch(m, meta, channel, pty, channel)
		}
		request.Reply(true, nil)
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
			if tcp.IsLocalHost(m, strings.Split(conn.RemoteAddr().String(), ":")[0]) {
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

const (
	ADDRESS = "address"
	CHANNEL = "channel"
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
			COMMAND: {Name: COMMAND, Help: "命令", Value: kit.Data(
				kit.MDB_FIELD, "time,id,username,hostname,cmd",
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
				mdb.CREATE: {Name: "create name=tcp port=9030", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
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
				m.Cmdy(mdb.SELECT, m.Prefix(LISTEN), "", mdb.HASH, kit.MDB_HASH, arg)
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
		},
	}, nil)
}
