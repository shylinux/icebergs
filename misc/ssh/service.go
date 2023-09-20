package ssh

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	psh "shylinux.com/x/icebergs/base/ssh"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/misc/xterm"
	kit "shylinux.com/x/toolkits"
)

func _ssh_meta(conn ssh.ConnMetadata) ice.Maps {
	return ice.Maps{aaa.USERNAME: conn.User(), tcp.HOSTPORT: conn.RemoteAddr().String()}
}
func _ssh_config(m *ice.Message, h string) *ssh.ServerConfig {
	config := &ssh.ServerConfig{
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			meta, err := _ssh_meta(conn), errors.New(ice.ErrNotRight)
			if tcp.IsLocalHost(m, meta[tcp.HOSTPORT]) {
				m.Auth(kit.SimpleKV(kit.Fields(aaa.USERNAME, tcp.HOSTPORT), meta))
				err = nil
			} else {
				mdb.ZoneSelectCB(m, h, func(value ice.Maps) {
					if !strings.HasPrefix(value[mdb.NAME], meta[aaa.USERNAME]+"@") {
						return
					}
					if s, e := base64.StdEncoding.DecodeString(value[mdb.TEXT]); !m.Warn(e, ice.ErrNotValid, value[mdb.TEXT]) {
						if pub, e := ssh.ParsePublicKey([]byte(s)); !m.Warn(e, ice.ErrNotValid, value[mdb.TEXT]) {
							if bytes.Compare(pub.Marshal(), key.Marshal()) == 0 {
								meta[tcp.HOSTNAME] = kit.Select("", kit.Split(value[mdb.NAME], "@"), 1)
								m.Auth(kit.SimpleKV(kit.Fields(aaa.USERNAME, tcp.HOSTPORT, tcp.HOSTNAME), meta))
								err = nil
							}
						}
					}
				})
			}
			return &ssh.Permissions{Extensions: meta}, err
		},
		PasswordCallback: func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
			meta, err := _ssh_meta(conn), errors.New(ice.ErrNotRight)
			// if aaa.UserLogin(m, meta[aaa.USERNAME], string(password)) {
			// 	m.Auth(kit.SimpleKV(kit.Fields(aaa.USERNAME, tcp.HOSTPORT, tcp.HOSTNAME), meta))
			// 	err = nil
			// }
			return &ssh.Permissions{Extensions: meta}, err
		},
	}
	if key, err := ssh.ParsePrivateKey([]byte(m.Cmdx(nfs.CAT, kit.HomePath(m.Option(PRIVATE))))); !m.Warn(err, ice.ErrNotValid, m.Option(PRIVATE)) {
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
		m.Go(func() {
			m.Logs(CHANNEL, tcp.HOSTPORT, c.RemoteAddr().String()+" -> "+c.LocalAddr().String())
			defer m.Logs("dischan", tcp.HOSTPORT, c.RemoteAddr().String()+" -> "+c.LocalAddr().String())
			_ssh_prepare(m.Spawn(conn.Permissions.Extensions), channel, requests)
		})
	}
}
func _ssh_prepare(m *ice.Message, channel ssh.Channel, requests <-chan *ssh.Request) {
	pty, tty, err := xterm.Open()
	if m.Warn(err) {
		return
	}
	defer tty.Close()
	list := kit.EnvSimple(cli.PATH)
	for request := range requests {
		m.Logs(REQUEST, m.OptionSimple(tcp.HOSTPORT), mdb.TYPE, request.Type)
		switch request.Type {
		case "pty-req":
			termLen := request.Payload[3]
			termEnv := string(request.Payload[4 : termLen+4])
			_ssh_size(pty.Fd(), request.Payload[termLen+4:])
			list = append(list, cli.TERM, termEnv)
		case "window-change":
			_ssh_size(pty.Fd(), request.Payload)
		case "env":
			var env struct{ Name, Value string }
			if err := ssh.Unmarshal(request.Payload, &env); err != nil {
				continue
			}
			list = append(list, env.Name, env.Value)
		case "shell":
			_ssh_handle(m, channel, pty, tty, list)
		case "exec":
			defer channel.Close()
			m.Options(cli.CMD_OUTPUT, channel, cli.CMD_ENV, list)
			m.Cmd(cli.SYSTEM, kit.Select("sh", kit.Env(cli.SHELL)), "-c", string(request.Payload[4:request.Payload[3]+4]))
			return
		}
		request.Reply(true, nil)
	}
}
func _ssh_handle(m *ice.Message, channel ssh.Channel, pty, tty *os.File, list []string) {
	h := m.Cmdx(mdb.INSERT, m.Prefix(CHANNEL), "", mdb.HASH, mdb.STATUS, tcp.OPEN, TTY, tty.Name(), m.OptionSimple(aaa.USERNAME, tcp.HOSTPORT), kit.Dict(mdb.TARGET, pty))
	p := _ssh_watch(m, h, pty, channel)
	m.Go(func() { io.Copy(channel, pty) })
	channel.Write([]byte(mdb.Config(m, WELCOME)))
	m.Options(cli.CMD_INPUT, tty, cli.CMD_OUTPUT, tty)
	m.Cmd(cli.DAEMON, kit.Select("sh", kit.Env(cli.SHELL)), func() {
		defer m.Cmd(mdb.MODIFY, m.Prefix(CHANNEL), "", mdb.HASH, mdb.HASH, h, mdb.STATUS, tcp.CLOSE)
		channel.Write([]byte(mdb.Config(m, GOODBYE)))
		channel.Close()
		p.Close()
	})
}

const (
	WELCOME = "welcome"
	GOODBYE = "goodbye"
	AUTHKEY = "authkey"
	REQUEST = "request"
)
const SERVICE = "service"

func init() {
	psh.Index.MergeCommands(ice.Commands{
		SERVICE: {Name: "service port id auto listen prunes", Icon: "ssh.png", Help: "服务", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				mdb.HashSelect(m).Table(func(value ice.Maps) {
					if value[mdb.STATUS] == tcp.OPEN {
						m.Cmd(SERVICE, tcp.LISTEN, tcp.PORT, value[tcp.PORT], value)
					}
				})
			}},
			tcp.LISTEN: {Name: "listen port=9022 private=.ssh/id_rsa authkey=.ssh/authorized_keys", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
				if mdb.HashSelect(m, m.Option(tcp.PORT)).Length() > 0 {
					mdb.HashModify(m, m.Option(tcp.PORT), mdb.STATUS, tcp.OPEN)
				} else {
					mdb.HashCreate(m.Spawn(), m.OptionSimple(mdb.Config(m, mdb.FIELDS)), mdb.STATUS, tcp.OPEN)
					m.Cmd("", ctx.LOAD, m.OptionSimple(AUTHKEY))
				}
				m.Go(func() {
					m.Cmd(web.BROAD, "send", mdb.TYPE, "sshd", mdb.NAME, ice.Info.Hostname, tcp.HOST, m.Cmd(tcp.HOST).Append(aaa.IP), tcp.PORT, m.Option(tcp.PORT))
					m.Cmd(tcp.SERVER, tcp.LISTEN, mdb.TYPE, SSH, mdb.NAME, m.Option(tcp.PORT), m.OptionSimple(tcp.PORT), func(c net.Conn) {
						if _c := tcp.NewPeekConn(c); _c.IsHTTP() {
							_c.Redirect(http.StatusTemporaryRedirect, m.Cmdx(web.SPACE, web.DOMAIN))
						} else {
							m.Go(func() { _ssh_accept(m, kit.Hashs(m.Option(tcp.PORT)), _c) })
						}
					})
				})
			}},
			mdb.INSERT: {Name: "insert text:textarea", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				if ls := kit.Split(m.Option(mdb.TEXT)); len(ls) > 2 {
					mdb.ZoneInsert(m, m.OptionSimple(tcp.PORT), mdb.TYPE, ls[0], mdb.NAME, ls[len(ls)-1], mdb.TEXT, strings.Join(ls[1:len(ls)-1], "+"))
				}
			}},
			ctx.LOAD: {Name: "load authkey=.ssh/authorized_keys", Help: "加载", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(nfs.CAT, kit.HomePath(m.Option(AUTHKEY)), func(pub string) { m.Cmd(SERVICE, mdb.INSERT, mdb.TEXT, pub) })
			}},
			ctx.SAVE: {Name: "save authkey=.ssh/authorized_keys", Help: "保存", Hand: func(m *ice.Message, arg ...string) {
				list := []string{}
				mdb.ZoneSelectCB(m, m.Option(tcp.PORT), func(value ice.Maps) {
					list = append(list, fmt.Sprintf("%s %s %s", value[mdb.TYPE], value[mdb.TEXT], value[mdb.NAME]))
				})
				if len(list) > 0 {
					m.Cmdy(nfs.SAVE, kit.HomePath(m.Option(AUTHKEY)), strings.Join(list, lex.NL)+lex.NL)
				}
			}},
			aaa.INVITE: {Help: "邀请", Hand: func(m *ice.Message, arg ...string) {
				m.Option(cli.HOSTNAME, tcp.PublishLocalhost(m, web.UserWeb(m).Hostname()))
				if buf, err := kit.Render(`ssh -p {{.Option "port"}} {{.Option "user.name"}}@{{.Option "hostname"}}`, m); err == nil {
					m.EchoScript(string(buf))
				}
			}},
		}, mdb.StatusHashAction(
			mdb.SHORT, tcp.PORT, mdb.FIELDS, "time,port,status,private,authkey,count", mdb.FIELD, "time,id,type,name,text",
			WELCOME, "welcome to contexts world\r\n", GOODBYE, "goodbye of contexts world\r\n",
		)), Hand: func(m *ice.Message, arg ...string) {
			if mdb.ZoneSelect(m, arg...); len(arg) == 0 {
				m.PushAction(aaa.INVITE, mdb.INSERT, ctx.LOAD, ctx.SAVE, mdb.REMOVE)
			}
		}},
	})
}
