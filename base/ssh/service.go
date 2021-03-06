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

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/base/tcp"
	kit "github.com/shylinux/toolkits"
	"golang.org/x/crypto/ssh"
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
				m.Cmd(mdb.SELECT, SERVICE, kit.Keys(kit.MDB_HASH, h), mdb.LIST).Table(func(index int, value map[string]string, head []string) {
					if !strings.HasPrefix(value[kit.MDB_NAME], conn.User()+"@") {
						return
					}
					if s, e := base64.StdEncoding.DecodeString(value[kit.MDB_TEXT]); !m.Warn(e != nil, e) {
						if pub, e := ssh.ParsePublicKey([]byte(s)); !m.Warn(e != nil, e) {

							if bytes.Compare(pub.Marshal(), key.Marshal()) == 0 {
								m.Log_AUTH(tcp.HOSTPORT, conn.RemoteAddr(), aaa.USERNAME, conn.User(), tcp.HOSTNAME, value[kit.MDB_NAME])
								meta[tcp.HOSTNAME] = kit.Select("", kit.Split(value[kit.MDB_NAME], "@"), 1)
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
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			SERVICE: {Name: SERVICE, Help: "服务", Value: kit.Data(
				WELCOME, "\r\nwelcome to context world\r\n",
				GOODBYE, "\r\ngoodbye of context world\r\n",
				kit.MDB_SHORT, tcp.PORT,
			)},
		},
		Commands: map[string]*ice.Command{
			ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Richs(SERVICE, "", kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
					if value = kit.GetMeta(value); kit.Value(value, kit.MDB_STATUS) == tcp.OPEN {
						m.Cmd(SERVICE, tcp.LISTEN, tcp.PORT, value[tcp.PORT], value)
					}
				})
			}},
			SERVICE: {Name: "service port id auto listen prunes", Help: "服务", Action: map[string]*ice.Action{
				tcp.LISTEN: {Name: "listen port=9030 private=.ssh/id_rsa authkey=.ssh/authorized_keys", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					if m.Richs(SERVICE, "", m.Option(tcp.PORT), func(key string, value map[string]interface{}) {
						kit.Value(value, kit.Keym(kit.MDB_STATUS), tcp.OPEN)
					}) == nil {
						m.Cmd(mdb.INSERT, SERVICE, "", mdb.HASH, tcp.PORT, m.Option(tcp.PORT),
							PRIVATE, m.Option(PRIVATE), AUTHKEY, m.Option(AUTHKEY), kit.MDB_STATUS, tcp.OPEN, arg)
						m.Cmd(SERVICE, mdb.IMPORT, AUTHKEY, m.Option(AUTHKEY))
					}

					m.Option(kit.Keycb(tcp.LISTEN), func(c net.Conn) { m.Go(func() { _ssh_accept(m, kit.Hashs(m.Option(tcp.PORT)), c) }) })
					m.Go(func() {
						m.Cmdy(tcp.SERVER, tcp.LISTEN, kit.MDB_TYPE, SSH, kit.MDB_NAME, tcp.PORT, tcp.PORT, m.Option(tcp.PORT))
					})
				}},

				mdb.INSERT: {Name: "insert text:textarea", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					if ls := kit.Split(m.Option(kit.MDB_TEXT)); len(ls) > 2 {
						m.Cmdy(mdb.INSERT, SERVICE, kit.Keys(kit.MDB_HASH, kit.Hashs(m.Option(tcp.PORT))), mdb.LIST,
							kit.MDB_TYPE, ls[0], kit.MDB_NAME, ls[len(ls)-1], kit.MDB_TEXT, strings.Join(ls[1:len(ls)-1], "+"))
					}
				}},
				mdb.EXPORT: {Name: "export authkey=.ssh/authorized_keys", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
					list := []string{}
					m.Cmd(mdb.SELECT, SERVICE, kit.Keys(kit.MDB_HASH, kit.Hashs(m.Option(tcp.PORT))), mdb.LIST).Table(func(index int, value map[string]string, head []string) {
						list = append(list, fmt.Sprintf("%s %s %s", value[kit.MDB_TYPE], value[kit.MDB_TEXT], value[kit.MDB_NAME]))
					})

					if len(list) > 0 {
						m.Cmdy(nfs.SAVE, path.Join(os.Getenv(cli.HOME), m.Option(AUTHKEY)), strings.Join(list, "\n")+"\n")
					}
				}},
				mdb.IMPORT: {Name: "import authkey=.ssh/authorized_keys", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
					p := path.Join(os.Getenv(cli.HOME), m.Option(AUTHKEY))
					for _, pub := range strings.Split(strings.TrimSpace(m.Cmdx(nfs.CAT, p)), "\n") {
						m.Cmd(SERVICE, mdb.INSERT, kit.MDB_TEXT, pub)
					}
					m.Echo(p)
				}},
				mdb.PRUNES: {Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
					m.Option(mdb.FIELDS, "time,port,status,private,authkey,count")
					m.Cmdy(mdb.PRUNES, SERVICE, "", mdb.HASH, kit.MDB_STATUS, tcp.ERROR)
					m.Cmdy(mdb.PRUNES, SERVICE, "", mdb.HASH, kit.MDB_STATUS, tcp.CLOSE)
				}},
				aaa.INVITE: {Name: "invite", Help: "邀请", Hand: func(m *ice.Message, arg ...string) {
					u := kit.ParseURL(m.Option(ice.MSG_USERWEB))
					m.Option(cli.HOSTNAME, strings.Split(u.Host, ":")[0])
					m.ProcessInner()

					if buf, err := kit.Render(`ssh -p {{.Option "port"}} {{.Option "user.name"}}@{{.Option "hostname"}}`, m); err == nil {
						m.EchoScript(string(buf))
					}
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 { // 服务列表
					m.Fields(len(arg) == 0, "time,port,status,private,authkey,count")
					m.Cmdy(mdb.SELECT, SERVICE, "", mdb.HASH)
					m.PushAction(mdb.IMPORT, mdb.INSERT, mdb.EXPORT, aaa.INVITE)
					return
				}

				// 公钥列表
				m.Fields(len(arg) == 1, "time,id,type,name,text")
				m.Cmdy(mdb.SELECT, SERVICE, kit.Keys(kit.MDB_HASH, kit.Hashs(arg[0])), mdb.LIST, kit.MDB_ID, arg[1:])
			}},
		},
	})
}
