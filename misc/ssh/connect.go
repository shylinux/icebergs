package ssh

import (
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	psh "shylinux.com/x/icebergs/base/ssh"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/code"
	"shylinux.com/x/icebergs/misc/xterm"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/task"
)

func _ssh_open(m *ice.Message, arg ...string) {
	_ssh_dial(m, func(c net.Conn) {
		fd := int(os.Stdin.Fd())
		if oldState, err := terminal.MakeRaw(fd); err == nil {
			defer terminal.Restore(fd, oldState)
		}
		w, h, _ := terminal.GetSize(fd)
		c.Write([]byte(fmt.Sprintf("#height:%d,width:%d"+lex.NL, h, w)))
		if m.Option(ice.INIT) != "" {
			kit.For(kit.Simple(m.Optionv(ice.INIT)), func(cmd string) {
				defer c.Write([]byte(cmd + lex.NL))
				m.Sleep300ms()
			})
		}
		m.Go(func() { io.Copy(c, os.Stdin) })
		io.Copy(os.Stdout, c)
	}, arg...)
}
func _ssh_dial(m *ice.Message, cb func(net.Conn), arg ...string) {
	os.Mkdir(kit.HomePath(".ssh/sess"), 0755)
	p := kit.HomePath(".ssh/sess", fmt.Sprintf("%s@%s:%s", m.Option(aaa.USERNAME), m.Option(tcp.HOST), m.Option(tcp.PORT)))
	if nfs.Exists(m, p) {
		if c, e := net.Dial(tcp.UNIX, p); e == nil {
			cb(c)
			return
		}
		nfs.Remove(m, p)
	}
	_ssh_conn(m, func(client *ssh.Client) {
		if l, e := net.Listen(tcp.UNIX, p); !m.WarnNotValid(e) {
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
							s, e := client.NewSession()
							if e != nil {
								return
							}
							defer c.Close()
							s.Stdin, s.Stdout, s.Stderr = c, c, c
							s.RequestPty(kit.Env(cli.TERM), h, w, ssh.TerminalModes{ssh.ECHO: 1, ssh.TTY_OP_ISPEED: 14400, ssh.TTY_OP_OSPEED: 14400})
							gdb.SignalNotify(m, 28, func() {
								w, h, _ := terminal.GetSize(int(os.Stdin.Fd()))
								s.WindowChange(h, w)
							})
							defer s.Wait()
							s.Shell()
						})
					}(c)
				}
			})
		}
		if c, e := net.Dial(tcp.UNIX, p); !m.WarnNotValid(e) {
			cb(c)
		}
	}, arg...)
}
func _ssh_conn(m *ice.Message, cb func(*ssh.Client), arg ...string) (err error) {
	methods := []ssh.AuthMethod{}
	methods = append(methods, ssh.PublicKeysCallback(func() ([]ssh.Signer, error) {
		key, err := ssh.ParsePrivateKey([]byte(m.Cmdx(nfs.CAT, kit.HomePath(m.OptionDefault(PRIVATE, ID_RSA_KEY)))))
		return []ssh.Signer{key}, err
	}))
	methods = append(methods, ssh.PasswordCallback(func() (string, error) { return m.Option(aaa.PASSWORD), nil }))
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
					res = append(res, aaa.TOTP_GET(verify, 30, 6))
				}
			case strings.HasSuffix(p, "password:"):
				res = append(res, m.Option(aaa.PASSWORD))
			}
		}
		return
	}))
	m.OptionDefault(tcp.PORT, tcp.PORT_22, aaa.USERNAME, aaa.ROOT)
	m.Cmd("tcp.client", tcp.DIAL, mdb.TYPE, SSH, mdb.NAME, m.Option(tcp.HOST), m.OptionSimple(tcp.HOST, tcp.PORT, aaa.USERNAME), arg, func(c net.Conn) {
		conn, chans, reqs, _err := ssh.NewClientConn(c, m.Option(tcp.HOST)+nfs.DF+m.Option(tcp.PORT), &ssh.ClientConfig{
			User: m.Option(aaa.USERNAME), Auth: methods, BannerCallback: func(message string) error { return nil },
			HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil },
		})
		kit.If(!m.WarnNotValid(_err), func() { cb(ssh.NewClient(conn, chans, reqs)) })
		err = _err
	})
	return
}
func _ssh_hold(m *ice.Message, c *ssh.Client) {
	if s, e := _ssh_session(m, c); !m.WarnNotValid(e) {
		defer s.Wait()
		s.Shell()
	}
}
func _ssh_target(m *ice.Message, name string) *ssh.Client {
	return mdb.HashSelectTarget(m, name, func(value ice.Maps) (res ice.Any) {
		m.GoWait(func(done func()) {
			_ssh_conn(m.Spawn(value), func(c *ssh.Client) {
				defer _ssh_hold(m, c)
				defer done()
				res = c
			})
		})
		return
	}).(*ssh.Client)
}

const SSH = "ssh"
const (
	DIRECT   = "direct"
	AUTHFILE = "authfile"

	ID_RSA_KEY = ".ssh/id_rsa"
	ID_RSA_PUB = ".ssh/id_rsa.pub"
)
const CONNECT = "connect"

func init() {
	psh.Index.MergeCommands(ice.Commands{
		CONNECT: {Help: "连接", Actions: ice.MergeActions(ice.Actions{
			tcp.OPEN: {Name: "open authfile username host=shylinux.com port=22 cmds init private=.ssh/id_rsa password verfiy", Help: "终端", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(ctx.CMDS) == "" {
					defer nfs.OptionLoad(m, m.Option(AUTHFILE)).Echo("exit %s@%s:%s\n", m.Option(aaa.USERNAME), m.Option(tcp.HOST), m.Option(tcp.PORT))
					_ssh_open(m.SetResult(), arg...)
				} else {
					_ssh_conn(m, func(c *ssh.Client) {
						s, _ := c.NewSession()
						defer s.Close()
						if b, e := s.CombinedOutput(m.Option(ctx.CMDS)); !m.WarnNotValid(e) {
							m.Echo(string(b))
						}
					})
				}
			}},
			tcp.DIAL: {Name: "dial name=shylinux host=shylinux.com port=22 username=shy private=.ssh/id_rsa", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				m.Go(func() {
					msg := m.Spawn()
					_ssh_conn(m, func(c *ssh.Client) {
						defer _ssh_hold(m, c)
						mdb.HashCreate(msg, m.OptionSimple(mdb.NAME, tcp.HOST, tcp.PORT, aaa.USERNAME, PRIVATE), kit.Dict(mdb.TARGET, c))
					}, arg...)
				}).Sleep3s()
			}},
			SESSION: {Help: "会话", Hand: func(m *ice.Message, arg ...string) {
				_ssh_hold(m, _ssh_target(m, m.Option(mdb.NAME)))
			}},
			DIRECT: {Name: "direct cmd=pwd", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(mdb.NAME) == "" {
					web.GoToastTable(m.Cmds(""), mdb.NAME, func(value ice.Maps) {
						msg := m.Cmds("", m.ActionKey(), value)
						kit.If(len(msg.Resultv()) == 0, func() { msg.TableEcho() })
						m.Push(mdb.TIME, msg.Time())
						m.Push(mdb.NAME, value[mdb.NAME])
						m.Push(cli.COST, m.FormatCost())
						m.Push(RES, msg.Result())
					}).ProcessInner()
				} else if s, e := _ssh_target(m, m.Option(mdb.NAME)).NewSession(); !m.WarnNotValid(e) {
					defer s.Close()
					if b, e := s.CombinedOutput(m.Option(ice.CMD)); !m.WarnNotValid(e) {
						m.Echo(string(b)).ProcessInner()
					}
				} else {
					mdb.HashSelectUpdate(m, m.Option(mdb.NAME), func(value ice.Map) { delete(value, mdb.TARGET) })
				}
			}},
			code.XTERM: {Hand: func(m *ice.Message, arg ...string) {
				ctx.ProcessField(m, code.XTERM, kit.JoinWord(SSH, m.Option(mdb.NAME)), arg...)
			}},
		}, mdb.StatusHashAction(mdb.SHORT, mdb.NAME, mdb.FIELD, "time,name,username,private,host,port"), mdb.ImportantHashAction()), Hand: func(m *ice.Message, arg ...string) {
			if mdb.HashSelect(m, arg...).PushAction(code.XTERM, DIRECT, SESSION, mdb.REMOVE); len(arg) == 0 {
				m.Sort(mdb.NAME).Action(tcp.DIAL, DIRECT)
			}
		}},
	})
}

type session struct {
	name string
	sess *ssh.Session
	pty  *os.File
}

func NewSession(m *ice.Message, arg ...string) (xterm.XTerm, error) {
	sess := &session{name: arg[0]}
	m.GoWait(func(done func()) {
		m.Cmd("ssh.connect", SESSION, kit.Dict(mdb.NAME, arg[0]), func(s *ssh.Session) {
			defer done()
			pty, tty, _ := xterm.Open()
			sess.sess, sess.pty = s, pty
			s.Stdin, s.Stdout, s.Stderr = tty, tty, tty
			s.RequestPty(kit.Env(cli.TERM), 24, 80, ssh.TerminalModes{
				ssh.ECHO: 0, ssh.TTY_OP_ISPEED: 14400, ssh.TTY_OP_OSPEED: 14400,
			})
		})
	})
	return sess, nil
}
func (s session) Setsize(h, w string) error     { return s.sess.WindowChange(kit.Int(h), kit.Int(w)) }
func (s session) Write(buf []byte) (int, error) { return s.pty.Write(buf) }
func (s session) Read(buf []byte) (int, error)  { return s.pty.Read(buf) }
func (s session) Close() error                  { return s.sess.Close() }

func init() { xterm.AddCommand(SSH, NewSession) }

func CombinedOutput(m *ice.Message, cmd string, cb func(string)) {
	_ssh_conn(m, func(c *ssh.Client) {
		if s, e := c.NewSession(); !m.WarnNotValid(e) {
			defer s.Close()
			m.Debug("cmd %v", cmd)
			if b, e := s.CombinedOutput(cmd); !m.WarnNotValid(e) {
				cb(string(b))
			}
		}
	})
}
func PushOutput(m *ice.Message, cmd string, cb func(string)) {
	_ssh_conn(m, func(c *ssh.Client) {
		if s, e := c.NewSession(); !m.WarnNotValid(e) {
			defer s.Close()
			r, _ := s.StdoutPipe()
			m.Debug("res %v", cmd)
			s.Run(cmd)
			kit.For(r, func(res []byte) {
				m.Debug("res %v", string(res))
				cb(string(res))
			})
		}
	})
}
func PushShell(m *ice.Message, cmds []string, cb func(string)) {
	_ssh_conn(m, func(c *ssh.Client) {
		if s, e := c.NewSession(); !m.WarnNotValid(e) {
			defer s.Close()
			w, _ := s.StdinPipe()
			r, _ := s.StdoutPipe()
			width, height, _ := terminal.GetSize(int(os.Stdin.Fd()))
			s.RequestPty(kit.Env(cli.TERM), height, width, ssh.TerminalModes{ssh.ECHO: 1, ssh.TTY_OP_ISPEED: 14400, ssh.TTY_OP_OSPEED: 14400})
			defer s.Wait()
			s.Shell()

			list := [][]string{}
			cmd := kit.Format("%s@%s[%s]%s$ ssh %s@%s\r\n",
				m.Option(aaa.USERNAME), ice.Info.Hostname, kit.Split(time.Now().Format(ice.MOD_TIME))[1], path.Base(kit.Path("")),
				m.Option(aaa.USERNAME), m.Option(tcp.HOST))
			list = append(list, []string{cmd})
			lock := task.Lock{}
			m.Debug("cmd %v", cmd)
			cb(cmd)
			defer cb("\r\n\r\n")
			m.Go(func() {
				kit.For(cmds, func(cmd string) {
					for {
						m.Sleep300ms()
						if func() bool { defer lock.Lock()(); return len(list[len(list)-1]) > 1 }() {
							break
						}
					}
					m.Debug("cmd %v", cmd)
					fmt.Fprintln(w, cmd)
					defer lock.Lock()()
					list = append(list, []string{cmd})
				})
				defer fmt.Fprintln(w, cli.EXIT)
				m.Sleep(m.OptionDefault("interval", "3s"))
			})
			kit.For(r, func(res []byte) {
				m.Debug("res %v", string(res))
				m.Debug("res %v", res)
				cb(string(res))
				defer lock.Lock()()
				list[len(list)-1] = append(list[len(list)-1], string(res))
			})
		}
	})
}
func RunConnect(arg ...string) string {
	defer func() { recover() }()
	kit.If(len(arg) == 0, func() { arg = append(arg, os.Args...) })
	return ice.Run(kit.Simple("ssh.connect", "open", AUTHFILE, kit.HomePath(".ssh/host/", path.Base(arg[0])+".json"), arg[1:])...)
}
