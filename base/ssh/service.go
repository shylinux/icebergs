package ssh

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/base/tcp"
	kit "github.com/shylinux/toolkits"
	"golang.org/x/crypto/ssh"

	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/kr/pty"
	"io"
	"net"
	"os"
	"path"
	"strings"
)

func _ssh_config(m *ice.Message, h string) *ssh.ServerConfig {
	config := &ssh.ServerConfig{
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			meta, err := map[string]string{aaa.USERNAME: conn.User()}, errors.New(ice.ErrNotAuth)
			if tcp.IsLocalHost(m, strings.Split(conn.RemoteAddr().String(), ":")[0]) {
				m.Log_AUTH(aaa.HOSTPORT, conn.RemoteAddr(), aaa.USERNAME, conn.User())
				err = nil
			} else {
				m.Cmd(mdb.SELECT, SERVICE, kit.Keys(kit.MDB_HASH, h), mdb.LIST).Table(func(index int, value map[string]string, head []string) {
					if !strings.HasPrefix(value[kit.MDB_NAME], conn.User()+"@") {
						return
					}
					if s, e := base64.StdEncoding.DecodeString(value[kit.MDB_TEXT]); !m.Warn(e != nil, e) {
						if pub, e := ssh.ParsePublicKey([]byte(s)); !m.Warn(e != nil, e) {

							if bytes.Compare(pub.Marshal(), key.Marshal()) == 0 {
								m.Log_AUTH(aaa.HOSTPORT, conn.RemoteAddr(), aaa.USERNAME, conn.User(), aaa.HOSTNAME, value[kit.MDB_NAME])
								meta[aaa.USERNAME] = value[kit.MDB_NAME]
								err = nil
							}
						}
					}
				})
			}
			return &ssh.Permissions{Extensions: meta}, err
		},
		PasswordCallback: func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
			meta, err := map[string]string{aaa.USERNAME: conn.User()}, errors.New(ice.ErrNotAuth)
			m.Richs(aaa.USER, "", conn.User(), func(k string, value map[string]interface{}) {
				if string(password) == kit.Format(value[aaa.PASSWORD]) {
					m.Log_AUTH(aaa.HOSTPORT, conn.RemoteAddr(), aaa.USERNAME, conn.User(), aaa.PASSWORD, strings.Repeat("*", len(kit.Format(value[aaa.PASSWORD]))))
					err = nil
				}
			})
			return &ssh.Permissions{Extensions: meta}, err
		},

		BannerCallback: func(conn ssh.ConnMetadata) string {
			m.Log_IMPORT(aaa.HOSTPORT, conn.RemoteAddr(), aaa.USERNAME, conn.User())
			return m.Conf(SERVICE, "meta.welcome")
		},
	}

	if key, err := ssh.ParsePrivateKey([]byte(m.Cmdx(nfs.CAT, path.Join(os.Getenv("HOME"), m.Option("private"))))); m.Assert(err) {
		config.AddHostKey(key)
	}
	return config
}
func _ssh_accept(m *ice.Message, h string, c net.Conn) {
	sc, chans, reqs, err := ssh.NewServerConn(c, _ssh_config(m, h))
	if m.Warn(err != nil, err) {
		return
	}

	m.Go(func() { ssh.DiscardRequests(reqs) })

	for ch := range chans {
		channel, requests, err := ch.Accept()
		if m.Warn(err != nil, err) {
			continue
		}

		func(channel ssh.Channel, requests <-chan *ssh.Request) {
			m.Go(func() { _ssh_handle(m, sc.Permissions.Extensions, c, channel, requests) })
		}(channel, requests)
	}
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

	h := m.Rich(CHANNEL, "", kit.Data(kit.MDB_STATUS, tcp.OPEN, TTY, tty.Name(), aaa.HOSTPORT, c.RemoteAddr().String()))
	meta[CHANNEL] = h

	for request := range requests {
		m.Logs("request", aaa.HOSTPORT, c.RemoteAddr(), kit.MDB_TYPE, request.Type)

		switch request.Type {
		case "pty-req":
			termLen := request.Payload[3]
			termEnv := string(request.Payload[4 : termLen+4])
			_ssh_size(pty.Fd(), request.Payload[termLen+4:])
			list = append(list, "TERM="+termEnv)

		case "window-change":
			_ssh_size(pty.Fd(), request.Payload)

		case "env":
			var env struct{ Name, Value string }
			if err := ssh.Unmarshal(request.Payload, &env); err != nil {
				continue
			}
			list = append(list, env.Name+"="+env.Value)

		case "exec":
			_ssh_exec(m, shell, []string{"-c", string(request.Payload[4 : request.Payload[3]+4])}, list, channel, func() {
				channel.Close()
			})
		case "shell":
			m.Go(func() { io.Copy(channel, pty) })

			_ssh_exec(m, shell, nil, list, tty, func() {
				defer m.Cmd(mdb.MODIFY, CHANNEL, "", mdb.HASH, kit.MDB_HASH, h, kit.MDB_STATUS, tcp.CLOSE)
				_ssh_close(m, c, channel)
			})

			_ssh_watch(m, meta, h, channel, pty, channel)
		}
		request.Reply(true, nil)
	}
}

const SERVICE = "service"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			SERVICE: {Name: SERVICE, Help: "服务", Value: kit.Data(
				"welcome", "\r\nwelcome to context world\r\n",
				"goodbye", "\r\ngoodbye of context world\r\n",
			)},
		},
		Commands: map[string]*ice.Command{
			SERVICE: {Name: "service hash id auto 监听 清理", Help: "服务", Action: map[string]*ice.Action{
				tcp.LISTEN: {Name: "listen port=9030 private=.ssh/id_rsa", Help: "监听", Hand: func(m *ice.Message, arg ...string) {
					h := m.Cmdx(mdb.INSERT, SERVICE, "", mdb.HASH, kit.MDB_STATUS, tcp.OPEN, arg)

					m.Option(tcp.LISTEN_CB, func(c net.Conn) { m.Go(func() { _ssh_accept(m, h, c) }) })
					m.Go(func() { m.Cmdy(tcp.SERVER, tcp.LISTEN, kit.MDB_NAME, SSH, tcp.PORT, m.Option(tcp.PORT)) })
				}},

				mdb.EXPORT: {Name: "export file=.ssh/authorized_keys", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
					list := []string{}
					m.Cmd(mdb.SELECT, SERVICE, kit.Keys(kit.MDB_HASH, m.Option(kit.MDB_HASH)), mdb.LIST).Table(func(index int, value map[string]string, head []string) {
						list = append(list, fmt.Sprintf("%s %s %s", value[kit.MDB_TYPE], value[kit.MDB_TEXT], value[kit.MDB_NAME]))
					})

					if len(list) > 0 {
						m.Cmdy(nfs.SAVE, path.Join(os.Getenv("HOME"), m.Option(kit.MDB_FILE)), strings.Join(list, "\n")+"\n")
					}
				}},
				mdb.IMPORT: {Name: "import file=.ssh/authorized_keys", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
					p := path.Join(os.Getenv("HOME"), m.Option(kit.MDB_FILE))
					for _, pub := range strings.Split(m.Cmdx(nfs.CAT, p), "\n") {
						if ls := kit.Split(pub); len(pub) > 10 {
							m.Cmd(mdb.INSERT, SERVICE, kit.Keys(kit.MDB_HASH, m.Option(kit.MDB_HASH)), mdb.LIST,
								kit.MDB_TYPE, ls[0], kit.MDB_NAME, ls[len(ls)-1], kit.MDB_TEXT, strings.Join(ls[1:len(ls)-1], "+"))
						}
					}
					m.Echo(p)
				}},
				mdb.PRUNES: {Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.PRUNES, SERVICE, "", mdb.HASH, kit.MDB_STATUS, tcp.CLOSE)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					m.Option(mdb.FIELDS, "time,hash,status,port,private,count")
					m.Cmdy(mdb.SELECT, SERVICE, "", mdb.HASH, kit.MDB_HASH, arg)
					m.PushAction("导入")
					return
				}

				m.Option(mdb.FIELDS, kit.Select("time,id,type,name,text", mdb.DETAIL, len(arg) > 1))
				m.Cmdy(mdb.SELECT, SERVICE, kit.Keys(kit.MDB_HASH, arg[0]), mdb.LIST, kit.MDB_ID, arg[1:])
			}},
		},
	}, nil)
}
