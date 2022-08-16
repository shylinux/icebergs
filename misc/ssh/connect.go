package ssh

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	psh "shylinux.com/x/icebergs/base/ssh"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
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
		for _, item := range kit.Simple(m.Optionv(mdb.LIST)) {
			m.Sleep300ms()
			c.Write([]byte(item + ice.NL))
		}

		m.Go(func() { io.Copy(c, os.Stdin) })
		io.Copy(os.Stdout, c)
	}, arg...)
}
func _ssh_dial(m *ice.Message, cb func(net.Conn), arg ...string) {
	p := kit.HomePath(".ssh", fmt.Sprintf("%s@%s:%s", m.Option(aaa.USERNAME), m.Option(tcp.HOST), m.Option(tcp.PORT)))
	if nfs.ExistsFile(m, p) {
		if c, e := net.Dial("unix", p); e == nil {
			cb(c) // 会话复用
			return
		}
		nfs.Remove(m, p)
	}

	_ssh_conn(m, func(client *ssh.Client) {
		if l, e := net.Listen("unix", p); m.Assert(e) {
			defer func() { nfs.Remove(m, p) }()
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

							s, e := client.NewSession()
							if e != nil {
								return
							}

							s.Stdin, s.Stdout, s.Stderr = c, c, c
							s.RequestPty(kit.Env(cli.TERM), h, w, ssh.TerminalModes{ssh.ECHO: 1, ssh.TTY_OP_ISPEED: 14400, ssh.TTY_OP_OSPEED: 14400})
							defer s.Wait()

							gdb.SignalNotify(m, 28, func() {
								w, h, _ := terminal.GetSize(int(os.Stdin.Fd()))
								s.WindowChange(h, w)
							})

							s.Shell()
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
		key, err := ssh.ParsePrivateKey([]byte(m.Cmdx(nfs.CAT, kit.HomePath(m.Option(PRIVATE)))))
		return []ssh.Signer{key}, err
	}))
	methods = append(methods, ssh.PasswordCallback(func() (string, error) {
		return m.Option(aaa.PASSWORD), nil
	}))

	m.Cmdy(tcp.CLIENT, tcp.DIAL, mdb.TYPE, SSH, mdb.NAME, m.Option(tcp.HOST), m.OptionSimple(tcp.HOST, tcp.PORT), arg, func(c net.Conn) {
		conn, chans, reqs, err := ssh.NewClientConn(c, m.Option(tcp.HOST)+":"+m.Option(tcp.PORT), &ssh.ClientConfig{
			User: m.Option(aaa.USERNAME), Auth: methods, BannerCallback: func(message string) error { return nil },
			HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil },
		})
		m.Assert(err)
		cb(ssh.NewClient(conn, chans, reqs))
	})
}

const SSH = "ssh"
const CONNECT = "connect"

func init() {
	psh.Index.MergeCommands(ice.Commands{
		CONNECT: {Name: "connect name auto", Help: "连接", Actions: ice.MergeActions(ice.Actions{
			tcp.OPEN: {Name: "open authfile username=shy password verfiy host=shylinux.com port=22 private=.ssh/id_rsa", Help: "终端", Hand: func(m *ice.Message, arg ...string) {
				aaa.UserRoot(m)
				_ssh_open(nfs.OptionLoad(m, m.Option("authfile")), arg...)
				m.Echo("exit %s@%s:%s\n", m.Option(aaa.USERNAME), m.Option(tcp.HOST), m.Option(tcp.PORT))
			}},
			tcp.DIAL: {Name: "dial name=shylinux host=shylinux.com port=22 username=shy private=.ssh/id_rsa", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				m.Go(func() {
					_ssh_conn(m, func(client *ssh.Client) {
						mdb.HashCreate(m.Spawn(), m.OptionSimple(mdb.NAME, tcp.HOST, tcp.PORT, aaa.USERNAME), mdb.STATUS, tcp.OPEN, kit.Dict(mdb.TARGET, client))
						m.Cmd("", SESSION, m.OptionSimple(mdb.NAME))
					}, arg...)
				})
				m.Sleep300ms()
			}},
			SESSION: {Name: "session", Help: "会话", Hand: func(m *ice.Message, arg ...string) {
				if c, e := _ssh_session(m, mdb.HashTarget(m, m.Option(mdb.NAME), nil).(*ssh.Client)); m.Assert(e) {
					defer c.Wait()
					c.Shell()
				}
			}},
			ctx.COMMAND: {Name: "command cmd=pwd", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
				client := mdb.HashTarget(m, m.Option(mdb.NAME), nil).(*ssh.Client)
				if s, e := client.NewSession(); m.Assert(e) {
					defer s.Close()
					if b, e := s.CombinedOutput(m.Option(ice.CMD)); m.Assert(e) {
						m.Echo(string(b))
					}
				}
			}},
		}, mdb.HashStatusAction(mdb.SHORT, "name", mdb.FIELD, "time,name,status,username,host,port")), Hand: func(m *ice.Message, arg ...string) {
			if mdb.HashSelect(m, arg...).Tables(func(value ice.Maps) {
				m.PushButton(kit.Select("", "command,session", value[mdb.STATUS] == tcp.OPEN), mdb.REMOVE)
			}); len(arg) == 0 {
				m.Action(tcp.DIAL)
			}
		}},
	})
}
