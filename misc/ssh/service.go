package ssh

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"os"
	"path"
	"strings"

	"golang.org/x/crypto/ssh"
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	psh "shylinux.com/x/icebergs/base/ssh"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
)

func _ssh_meta(conn ssh.ConnMetadata) map[string]string {
	return map[string]string{aaa.USERNAME: conn.User(), tcp.HOSTPORT: conn.RemoteAddr().String()}
}
func _ssh_config(m *ice.Message, h string) *ssh.ServerConfig {
	config := &ssh.ServerConfig{
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			meta, err := _ssh_meta(conn), errors.New(ice.ErrNotRight)
			if tcp.IsLocalHost(m, strings.Split(conn.RemoteAddr().String(), ":")[0]) {
				m.Log_AUTH(tcp.HOSTPORT, conn.RemoteAddr(), aaa.USERNAME, conn.User())
				err = nil // 本机用户
			} else {
				m.Cmd(mdb.SELECT, SERVICE, kit.Keys(mdb.HASH, h), mdb.LIST).Table(func(index int, value map[string]string, head []string) {
					if !strings.HasPrefix(value[mdb.NAME], conn.User()+"@") {
						return
					}
					if s, e := base64.StdEncoding.DecodeString(value[mdb.TEXT]); !m.Warn(e) {
						if pub, e := ssh.ParsePublicKey([]byte(s)); !m.Warn(e) {

							if bytes.Compare(pub.Marshal(), key.Marshal()) == 0 {
								m.Log_AUTH(tcp.HOSTPORT, conn.RemoteAddr(), aaa.USERNAME, conn.User(), tcp.HOSTNAME, value[mdb.NAME])
								meta[tcp.HOSTNAME] = kit.Select("", kit.Split(value[mdb.NAME], "@"), 1)
								err = nil // 认证成功
							}
						}
					}
				})
			}
			return &ssh.Permissions{Extensions: meta}, err
		},
		PasswordCallback: func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
			meta, err := _ssh_meta(conn), errors.New(ice.ErrNotRight)
			if aaa.UserLogin(m, conn.User(), string(password)) {
				m.Log_AUTH(tcp.HOSTPORT, conn.RemoteAddr(), aaa.USERNAME, conn.User(), aaa.PASSWORD, strings.Repeat("*", len(string(password))))
				err = nil // 密码登录
			}
			return &ssh.Permissions{Extensions: meta}, err
		},

		BannerCallback: func(conn ssh.ConnMetadata) string {
			m.Log_IMPORT(tcp.HOSTPORT, conn.RemoteAddr(), aaa.USERNAME, conn.User())
			return m.Conf(SERVICE, kit.Keym(WELCOME))
		},
	}

	if key, err := ssh.ParsePrivateKey([]byte(m.Cmdx(nfs.CAT, path.Join(os.Getenv(cli.HOME), m.Option(PRIVATE))))); m.Assert(err) {
		config.AddHostKey(key)
	}
	return config
}
func _ssh_accept(m *ice.Message, h string, c net.Conn) {
	conn, chans, reqs, err := ssh.NewServerConn(c, _ssh_config(m, h))
	if m.Warn(err) {
		return
	}

	m.Go(func() { ssh.DiscardRequests(reqs) })

	for ch := range chans {
		channel, requests, err := ch.Accept()
		if m.Warn(err) {
			continue
		}

		func(channel ssh.Channel, requests <-chan *ssh.Request) {
			m.Go(func() { _ssh_handle(m, conn.Permissions.Extensions, c, channel, requests) })
		}(channel, requests)
	}
}

const (
	WELCOME = "welcome"
	GOODBYE = "goodbye"
	PRIVATE = "private"
	AUTHKEY = "authkey"
	REQUEST = "request"
)
const SERVICE = "service"

func init() {
	psh.Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		SERVICE: {Name: SERVICE, Help: "服务", Value: kit.Data(
			WELCOME, "\r\nwelcome to context world\r\n", GOODBYE, "\r\ngoodbye of context world\r\n",
			mdb.SHORT, tcp.PORT, mdb.FIELD, "time,port,status,private,authkey,count",
		)},
	}, Commands: map[string]*ice.Command{
		SERVICE: {Name: "service port id auto listen prunes", Help: "服务", Action: ice.MergeAction(map[string]*ice.Action{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Richs(SERVICE, "", mdb.FOREACH, func(key string, value map[string]interface{}) {
					if value = kit.GetMeta(value); kit.Value(value, mdb.STATUS) == tcp.OPEN {
						m.Cmd(SERVICE, tcp.LISTEN, tcp.PORT, value[tcp.PORT], value)
					}
				})
			}},
			tcp.LISTEN: {Name: "listen port=9030 private=.ssh/id_rsa authkey=.ssh/authorized_keys", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				if m.Richs(SERVICE, "", m.Option(tcp.PORT), func(key string, value map[string]interface{}) {
					kit.Value(value, kit.Keym(mdb.STATUS), tcp.OPEN)
				}) == nil {
					m.Cmd(mdb.INSERT, SERVICE, "", mdb.HASH, tcp.PORT, m.Option(tcp.PORT),
						PRIVATE, m.Option(PRIVATE), AUTHKEY, m.Option(AUTHKEY), mdb.STATUS, tcp.OPEN, arg)
					m.Cmd(SERVICE, mdb.IMPORT, AUTHKEY, m.Option(AUTHKEY))
				}

				m.OptionCB(tcp.SERVER, func(c net.Conn) { m.Go(func() { _ssh_accept(m, kit.Hashs(m.Option(tcp.PORT)), c) }) })
				m.Go(func() {
					m.Cmdy(tcp.SERVER, tcp.LISTEN, mdb.TYPE, SSH, mdb.NAME, tcp.PORT, tcp.PORT, m.Option(tcp.PORT))
				})
			}},

			mdb.INSERT: {Name: "insert text:textarea", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				if ls := kit.Split(m.Option(mdb.TEXT)); len(ls) > 2 {
					m.Cmdy(mdb.INSERT, SERVICE, kit.Keys(mdb.HASH, kit.Hashs(m.Option(tcp.PORT))), mdb.LIST,
						mdb.TYPE, ls[0], mdb.NAME, ls[len(ls)-1], mdb.TEXT, strings.Join(ls[1:len(ls)-1], "+"))
				}
			}},
			mdb.EXPORT: {Name: "export authkey=.ssh/authorized_keys", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
				list := []string{}
				m.Cmd(mdb.SELECT, SERVICE, kit.Keys(mdb.HASH, kit.Hashs(m.Option(tcp.PORT))), mdb.LIST).Table(func(index int, value map[string]string, head []string) {
					list = append(list, fmt.Sprintf("%s %s %s", value[mdb.TYPE], value[mdb.TEXT], value[mdb.NAME]))
				})

				if len(list) > 0 {
					m.Cmdy(nfs.SAVE, path.Join(os.Getenv(cli.HOME), m.Option(AUTHKEY)), strings.Join(list, ice.NL)+ice.NL)
				}
			}},
			mdb.IMPORT: {Name: "import authkey=.ssh/authorized_keys", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
				p := path.Join(os.Getenv(cli.HOME), m.Option(AUTHKEY))
				for _, pub := range strings.Split(strings.TrimSpace(m.Cmdx(nfs.CAT, p)), ice.NL) {
					m.Cmd(SERVICE, mdb.INSERT, mdb.TEXT, pub)
				}
				m.Echo(p)
			}},
			aaa.INVITE: {Name: "invite", Help: "邀请", Hand: func(m *ice.Message, arg ...string) {
				u := kit.ParseURL(m.Option(ice.MSG_USERWEB))
				m.Option(cli.HOSTNAME, strings.Split(u.Host, ":")[0])
				m.ProcessInner()

				if buf, err := kit.Render(`ssh -p {{.Option "port"}} {{.Option "user.name"}}@{{.Option "hostname"}}`, m); err == nil {
					m.EchoScript(string(buf))
				}
			}},
		}, mdb.HashActionStatus()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 { // 服务列表
				mdb.HashSelect(m, arg...)
				m.PushAction(mdb.IMPORT, mdb.INSERT, mdb.EXPORT, aaa.INVITE)
				return
			}

			// 公钥列表
			m.Fields(len(arg[1:]), "time,id,type,name,text")
			mdb.ZoneSelect(m, arg...)
		}},
	}})
}
