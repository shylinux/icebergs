package ssh

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"strings"

	"golang.org/x/crypto/ssh"
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	psh "shylinux.com/x/icebergs/base/ssh"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _ssh_meta(conn ssh.ConnMetadata) ice.Maps {
	return ice.Maps{aaa.USERNAME: conn.User(), tcp.HOSTPORT: conn.RemoteAddr().String()}
}
func _ssh_config(m *ice.Message, h string) *ssh.ServerConfig {
	config := &ssh.ServerConfig{
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			meta, err := _ssh_meta(conn), errors.New(ice.ErrNotRight)
			if tcp.IsLocalHost(m, strings.Split(conn.RemoteAddr().String(), ice.DF)[0]) {
				m.Logs(ice.LOG_AUTH, tcp.HOSTPORT, conn.RemoteAddr(), aaa.USERNAME, conn.User())
				err = nil // 本机用户
			} else {
				mdb.ZoneSelectCB(m, h, func(value ice.Maps) {
					if !strings.HasPrefix(value[mdb.NAME], conn.User()+"@") {
						return
					}
					if s, e := base64.StdEncoding.DecodeString(value[mdb.TEXT]); !m.Warn(e) {
						if pub, e := ssh.ParsePublicKey([]byte(s)); !m.Warn(e) {

							if bytes.Compare(pub.Marshal(), key.Marshal()) == 0 {
								m.Logs(ice.LOG_AUTH, tcp.HOSTPORT, conn.RemoteAddr(), aaa.USERNAME, conn.User(), tcp.HOSTNAME, value[mdb.NAME])
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
				m.Logs(ice.LOG_AUTH, tcp.HOSTPORT, conn.RemoteAddr(), aaa.USERNAME, conn.User(), aaa.PASSWORD, strings.Repeat("*", len(string(password))))
				err = nil // 密码登录
			}
			return &ssh.Permissions{Extensions: meta}, err
		},
		BannerCallback: func(conn ssh.ConnMetadata) string {
			m.Logs(ice.LOG_AUTH, tcp.HOSTPORT, conn.RemoteAddr(), aaa.USERNAME, conn.User())
			return m.Config(WELCOME)
		},
	}

	if key, err := ssh.ParsePrivateKey([]byte(m.Cmdx(nfs.CAT, kit.HomePath(m.Option(PRIVATE))))); m.Assert(err) {
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
	AUTHKEY = "authkey"
	REQUEST = "request"
)
const SERVICE = "service"

func init() {
	psh.Index.Merge(&ice.Context{Configs: ice.Configs{
		SERVICE: {Name: SERVICE, Help: "服务", Value: kit.Data(
			mdb.SHORT, tcp.PORT, mdb.FIELD, "time,port,status,private,authkey,count",
			WELCOME, "\r\nwelcome to contexts world\r\n", GOODBYE, "\r\ngoodbye of contexts world\r\n",
		)},
	}, Commands: ice.Commands{
		SERVICE: {Name: "service port id auto listen prunes", Help: "服务", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				mdb.HashSelect(m).Tables(func(value ice.Maps) {
					if value[mdb.STATUS] == tcp.OPEN {
						m.Cmd(SERVICE, tcp.LISTEN, tcp.PORT, value[tcp.PORT], value)
					}
				})
			}},
			tcp.LISTEN: {Name: "listen port=9030 private=.ssh/id_rsa authkey=.ssh/authorized_keys", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				if mdb.HashSelect(m, m.Option(tcp.PORT)).Length() > 0 {
					mdb.HashModify(m, m.Option(tcp.PORT), mdb.STATUS, tcp.OPEN)
				} else {
					mdb.HashCreate(m, mdb.STATUS, tcp.OPEN, arg)
					m.Cmd("", nfs.LOAD, m.OptionSimple(AUTHKEY))
				}

				m.Go(func() {
					m.Cmdy(tcp.SERVER, tcp.LISTEN, mdb.TYPE, SSH, mdb.NAME, tcp.PORT, m.OptionSimple(tcp.PORT), func(c net.Conn) {
						m.Go(func() { _ssh_accept(m, kit.Hashs(m.Option(tcp.PORT)), c) })
					})
				})
			}},

			mdb.INSERT: {Name: "insert text:textarea", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				if ls := kit.Split(m.Option(mdb.TEXT)); len(ls) > 2 {
					m.Cmdy(mdb.INSERT, SERVICE, kit.Keys(mdb.HASH, kit.Hashs(m.Option(tcp.PORT))), mdb.LIST,
						mdb.TYPE, ls[0], mdb.NAME, ls[len(ls)-1], mdb.TEXT, strings.Join(ls[1:len(ls)-1], "+"))
				}
			}},
			nfs.LOAD: {Name: "load authkey=.ssh/authorized_keys", Help: "加载", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(nfs.CAT, kit.HomePath(m.Option(AUTHKEY)), func(pub string) {
					m.Cmd(SERVICE, mdb.INSERT, mdb.TEXT, pub)
				})
			}},
			nfs.SAVE: {Name: "save authkey=.ssh/authorized_keys", Help: "保存", Hand: func(m *ice.Message, arg ...string) {
				list := []string{}
				mdb.ZoneSelectCB(m, m.Option(tcp.PORT), func(value ice.Maps) {
					list = append(list, fmt.Sprintf("%s %s %s", value[mdb.TYPE], value[mdb.TEXT], value[mdb.NAME]))
				})
				if len(list) > 0 {
					m.Cmdy(nfs.SAVE, kit.HomePath(m.Option(AUTHKEY)), strings.Join(list, ice.NL)+ice.NL)
				}
			}},
			aaa.INVITE: {Name: "invite", Help: "邀请", Hand: func(m *ice.Message, arg ...string) {
				m.Option(cli.HOSTNAME, web.OptionUserWeb(m).Hostname())
				if buf, err := kit.Render(`ssh -p {{.Option "port"}} {{.Option "user.name"}}@{{.Option "hostname"}}`, m); err == nil {
					m.EchoScript(string(buf))
				}
			}},
		}, mdb.HashStatusAction()), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 { // 服务列表
				mdb.HashSelect(m, arg...).PushAction(aaa.INVITE, mdb.INSERT, nfs.LOAD, nfs.SAVE)
				return
			}

			// 公钥列表
			m.Fields(len(arg[1:]), "time,id,type,name,text")
			mdb.ZoneSelect(m, arg...)
		}},
	}})
}
