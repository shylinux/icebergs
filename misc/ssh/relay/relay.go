package relay

import (
	"os"
	"path"
	"strings"
	"time"

	"shylinux.com/x/ice"
	icebergs "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/base/web/html"
	"shylinux.com/x/icebergs/core/code"
	"shylinux.com/x/icebergs/misc/ssh"
	"shylinux.com/x/icebergs/misc/xterm"
	kit "shylinux.com/x/toolkits"
)

const (
	RELAY        = "relay"
	SSH_RELAY    = "ssh.relay"
	SRC_RELAY_GO = "src/relay.go"
	CONTEXTS     = "contexts/"

	INSTALL_SH = "install.sh"
	UPGRADE_SH = "upgrade.sh"
	VERSION_SH = "version.sh"
	PUSHBIN_SH = "pushbin.sh"
)
const (
	MACHINE    = "machine"
	PACKAGE    = "package"
	KERNEL     = "kernel"
	SHELL      = "shell"
	ARCH       = "arch"
	NCPU       = "ncpu"
	VCPU       = "vcpu"
	MHZ        = "mhz"
	MEM        = "mem"
	MEM_FREE   = "mem_free"
	MEM_TOTAL  = "mem_total"
	DISK       = "disk"
	DISK_USED  = "disk_used"
	DISK_TOTAL = "disk_total"
	NETWORK    = "network"
	LISTEN     = "listen"
	SOCKET     = "socket"
	PROC       = "proc"
)

type relay struct {
	ice.Hash
	ice.Code
	checkbox    string `data:"true"`
	short       string `data:"machine"`
	field       string `data:"time,machine,username,host,port,portal,dream,module,version,commitTime,compileTime,bootTime,package,shell,kernel,arch,ncpu,vcpu,mhz,mem,disk,network,listen,socket,proc,vendor"`
	statsTables string `name:"statsTables" event:"stats.tables"`
	create      string `name:"create machine* username* host* port*=22"`
	stats       string `name:"stats machine" help:"采集" icon:"bi bi-pc-display"`
	dream       string `name:"dream" help:"空间" icon:"bi bi-grid-3x3-gap"`
	forEach     string `name:"forEach machine cmd*:textarea=pwd" help:"遍历" icon:"bi bi-card-list"`
	forFlow     string `name:"forFlow machine cmd*:textarea=pwd" help:"流程" icon:"bi bi-terminal"`
	pubkey      string `name:"pubkey" help:"公钥" icon:"bi bi-person-vcard"`
	publish     string `name:"publish" help:"发布" icon:"bi bi-send-check"`
	list        string `name:"list machine auto" help:"机器" icon:"relay.png"`
	install     string `name:"install dream param" help:"安装"`
	pushbin     string `name:"pushbin dream param" help:"部署"`
	adminCmd    string `name:"adminCmd cmd" help:"命令" icon:"bi bi-terminal-plus"`
}

func (s relay) StatsTables(m *ice.Message, arg ...string) {
	web.PushStats(m.Message, "", mdb.HashSelects(m.Spawn().Message).Length(), "", "服务器数量")
}
func (s relay) Init(m *ice.Message, arg ...string) {
	s.Hash.Init(m).TransInput(MACHINE, "机器",
		PACKAGE, "软件包", SHELL, "命令行", KERNEL, "内核", ARCH, "架构",
		NCPU, "处理器", VCPU, "虚拟核", MHZ, "频率",
		MEM, "内存", DISK, "磁盘", NETWORK, "流量",
		LISTEN, "服务", SOCKET, "连接", PROC, "进程",
	)
	msg := m.Spawn(ice.Maps{ice.MSG_FIELDS: ""})
	m.GoSleep3s(func() { s.Hash.List(msg).Table(func(value ice.Maps) { s.xterm(m.Spawn(value)) }) })
}
func (s relay) Inputs(m *ice.Message, arg ...string) {
	switch s.Hash.Inputs(m, arg...); arg[0] {
	case MACHINE:
		if m.Option(ctx.ACTION) == mdb.CREATE {
			m.Copy(m.Cmd(web.SPIDE).CutTo(web.CLIENT_NAME, arg[0]))
		}
	case aaa.USERNAME:
		m.Copy(m.Cmd(aaa.USER).Cut(aaa.USERNAME).Push(arg[0], aaa.ROOT).Push(arg[0], ice.SHY))
	case tcp.HOST:
		m.Copy(m.Options(ice.MSG_FIELDS, web.CLIENT_HOSTNAME).Cmd(web.SPIDE).CutTo(web.CLIENT_HOSTNAME, arg[0]))
	case tcp.PORT:
		m.Push(arg[0], tcp.PORT_22, tcp.PORT_9022)
	case web.PORTAL:
		kit.If(m.Option(tcp.LISTEN), func(p string) { m.Push(arg[0], kit.Split(p)) })
		m.Push(arg[0], tcp.PORT_443, tcp.PORT_80, tcp.PORT_9020)
	case cli.PARAM:
		m.Push(arg[0], `forever start dev "" port 9020`)
	}
}
func (s relay) Stats(m *ice.Message) {
	cmds := []string{
		PACKAGE, `if yum -h &>/dev/null; then echo yum; elif apk version &>/dev/null; then echo apk; elif apt -h &>/dev/null; then echo apt; fi`,
		SHELL, `echo $SHELL`, KERNEL, `uname -s`, ARCH, `uname -m`,
		NCPU, `cat /proc/cpuinfo | grep "physical id" | sort | uniq | wc -l`,
		VCPU, `cat /proc/cpuinfo | grep "processor" | sort | uniq | wc -l`,
		MHZ, `cat /proc/cpuinfo | grep "cpu MHz" | grep -o "[0-9.]\+" | sort -nr | head -n1`,
		MEM, `cat /proc/meminfo | head -n2 | grep -o "[0-9.]*"`, DISK, `df | grep "^/dev/"`,
		NETWORK, `cat /proc/net/dev | grep "eth0\|eth1" | sort -r | head -n1`,
		LISTEN, `netstat -ntl 2>/dev/null | grep "^tcp" | grep -o ":[0-9]\+" | grep -o "[0-9]\+" | sort -n | uniq`,
		SOCKET, `netstat -nt 2>/dev/null | grep "^tcp" | wc -l`, PROC, `ps aux | wc -l`,
	}
	trans := map[string]func([]string) string{
		MEM: func(ls []string) string {
			return kit.FmtSize(kit.Int(kit.Select("", ls, 1))*1024, kit.Int(kit.Select("", ls, 0))*1024)
		},
		DISK: func(ls []string) string {
			return kit.FmtSize(kit.Int(kit.Select("", ls, 2))*1024, kit.Int(kit.Select("", ls, 1))*1024)
		},
		NETWORK: func(ls []string) string {
			return kit.FmtSize(kit.Int(kit.Select("", ls, 1)), kit.Int(kit.Select("", ls, 9)))
		},
	}
	machine := m.Option(MACHINE)
	web.GoToast(m.Message, "", func(toast func(string, int, int)) []string {
		kit.For(cmds, func(key, value string, index int) {
			toast(key, index/2, len(cmds)/2)
			s.foreachModify(m, machine, key, value, trans[key])
		})
		return nil
	}).ProcessInner()
	s.foreach(m.Spawn(ice.Maps{MACHINE: machine}), func(msg *ice.Message, cmd []string) {
		ssh.CombinedOutput(msg.Message, s.admins(msg, cli.RUNTIME), func(res string) {
			if !strings.HasPrefix(res, "warn: ") {
				s.Modify(m, kit.Simple(MACHINE, msg.Option(MACHINE), kit.Dict(cli.ParseMake(res)))...)
			}
		})
	})
}
func (s relay) Dream(m *ice.Message) {
	if m.Option(web.PORTAL) != "" {
		m.ProcessOpen(web.HostPort(m.Message, m.Option(tcp.HOST), m.Option(web.PORTAL), "", web.DREAM))
		return
	}
	fields := "time,machine,host,space,type,status,module,version,commitTime,compileTime,bootTime,link"
	s.foreach(m, func(msg *ice.Message, cmd []string) {
		m.Push("", kit.Dict(msg.OptionSimple(fields), mdb.TYPE, web.SERVER, mdb.STATUS, web.ONLINE, web.SPACE, ice.CONTEXTS, web.LINK, web.HostPort(m.Message, msg.Option(tcp.HOST), msg.Option(web.PORTAL))), kit.Split(fields))
		ssh.CombinedOutput(msg.Message, s.admins(msg, web.ROUTE), func(res string) {
			_msg := m.Spawn().SplitIndex(res)
			m.Copy(_msg.Table(func(value ice.Maps) {
				switch _msg.Push(MACHINE, msg.Option(MACHINE)).Push(tcp.HOST, msg.Option(tcp.HOST)); msg.Option(web.PORTAL) {
				case "":
					_msg.Push(web.LINK, "")
				default:
					_msg.Push(web.LINK, web.HostPort(m.Message, msg.Option(tcp.HOST), msg.Option(web.PORTAL), value[web.SPACE]))
				}
			}).Cut(fields))
		})
	})
	m.Options(ice.MSG_PROCESS, "")
	if m.Action(s.Dream, "filter:text"); tcp.IsLocalHost(m.Message, m.Option(ice.MSG_USERIP)) {
		if _msg := m.Cmd(cli.SYSTEM, ice.BIN_ICE_BIN, web.ADMIN, cli.RUNTIME); len(_msg.Result()) > 0 {
			m.Push(MACHINE, tcp.LOCALHOST).Push(tcp.HOST, tcp.PublishLocalhost(m.Message, tcp.LOCALHOST))
			m.Push("", kit.Dict(cli.ParseMake(_msg.Result())), kit.Split("time,space,module,version,commitTime,compileTime,bootTime"))
			m.Push(mdb.TYPE, web.SERVER).Push(mdb.STATUS, web.ONLINE).Push(web.LINK, web.UserHost(m.Message))
		}
		if _msg := m.Spawn().SplitIndex(m.Cmdx(cli.SYSTEM, kit.Split(s.admin(m, web.ROUTE)))); _msg.Length() > 0 {
			m.Copy(_msg.Table(func(value ice.Maps) {
				_msg.Push(MACHINE, tcp.LOCALHOST).Push(tcp.HOST, tcp.PublishLocalhost(m.Message, tcp.LOCALHOST))
				_msg.Push(web.LINK, web.UserHost(m.Message)+web.S(value[web.SPACE]))
			}).Cut(fields))
		}
	}
}
func (s relay) ForEach(m *ice.Message, arg ...string) *ice.Message {
	s.foreach(m, func(msg *ice.Message, cmd []string) {
		kit.For(cmd, func(cmd string) {
			begin := time.Now()
			ssh.CombinedOutput(msg.Message, cmd, func(res string) {
				m.Push(mdb.TIME, begin.Format(ice.MOD_TIME))
				m.Push(MACHINE, msg.Option(MACHINE)).Push(tcp.HOST, msg.Option(tcp.HOST))
				m.Push(cli.COST, kit.FmtDuration(time.Now().Sub(begin)))
				m.Push(ice.CMD, cmd).Push(ice.RES, res).PushButton(s.ForEach)
			})
		})
	})
	return m
}
func (s relay) ForFlow(m *ice.Message) {
	s.foreach(m, func(msg *ice.Message, cmd []string) {
		ssh.PushShell(msg.Message, cmd, func(res string) {
			web.PushNoticeGrow(m.Options(ctx.DISPLAY, web.PLUGIN_XTERM).Message, res)
		})
	})
}
func (s relay) Compile(m *ice.Message) {
	m.Cmdy(code.COMPILE, SRC_RELAY_GO, path.Join(ice.USR_PUBLISH, RELAY)).ProcessInner()
}
func (s relay) Publish(m *ice.Message, arg ...string) {
	if m.Option(MACHINE) == "" {
		s.Hash.ForEach(m, "", func(msg *ice.Message) { s.Publish(msg) })
		m.Cmdy(nfs.DIR, ice.USR_PUBLISH).Set(ctx.ACTION)
		return
	}
	kit.If(!nfs.Exists(m, path.Join(ice.USR_PUBLISH, RELAY)), func() { s.Compile(m) })
	os.Symlink(RELAY, ice.USR_PUBLISH+m.Option(MACHINE))
	m.Cmd(nfs.SAVE, kit.HomePath(".ssh/"+m.Option(MACHINE)+".json"), kit.Formats(kit.Dict(m.OptionSimple("username,host,port")))+ice.NL)
}
func (s relay) Pubkey(m *ice.Message, arg ...string) {
	m.EchoScript(m.Cmdx(nfs.CAT, kit.HomePath(ssh.ID_RSA_PUB))).ProcessInner()
}
func (s relay) List(m *ice.Message, arg ...string) *ice.Message {
	if s.Hash.List(m, arg...); len(arg) == 0 {
		if m.Length() == 0 {
			m.Action(s.Create)
		} else {
			m.Action(s.Create, s.Upgrade, s.Version, s.Stats, s.Dream)
		}
	}
	stats := map[string]int{}
	m.Table(func(value ice.Maps) {
		stats[VCPU] += kit.Int(value[VCPU])
		stats[PROC] += kit.Int(value[PROC])
		stats[SOCKET] += kit.Int(value[SOCKET])
		if ls := kit.Split(value[MEM], " /"); len(ls) > 1 {
			stats[MEM_FREE] += kit.Int(ls[0])
			stats[MEM_TOTAL] += kit.Int(ls[1])
		}
		if ls := kit.Split(value[DISK], " /"); len(ls) > 1 {
			stats[DISK_USED] += kit.Int(ls[0])
			stats[DISK_TOTAL] += kit.Int(ls[1])
		}
		if value[web.PORTAL] == "" {
			m.Push(web.LINK, "").PushButton(s.Install, s.Pushbin, s.Xterm, s.Remove)
			return
		}
		m.Push(web.LINK, web.HostPort(m.Message, value[tcp.HOST], value[web.PORTAL]))
		m.PushButton(s.Admin, s.Dream, s.Vimer, s.AdminCmd, s.Upgrade, s.Pushbin, s.Xterm, s.Remove)
		kit.If(len(arg) > 0, func() { m.PushQRCode(cli.QRCODE, m.Append(web.LINK)) })
	})
	_stats := kit.Dict(MEM, kit.FmtSize(stats[MEM_FREE], stats[MEM_TOTAL]), DISK, kit.FmtSize(stats[DISK_USED], stats[DISK_TOTAL]))
	m.StatusTimeCount(m.Spawn().Options(stats, _stats).OptionSimple(VCPU, MEM, DISK, SOCKET, PROC))
	m.RewriteAppend(func(value, key string, index int) string {
		if key == MEM {
			if ls := kit.Split(value, " /"); len(ls) > 0 && kit.Int(ls[0]) < 256*1024*1024 {
				return html.FormatDanger(value)
			}
		}
		return value
	})
	return m
}
func (s relay) Pushbin(m *ice.Message, arg ...string) {
	bin := "ice"
	switch m.Option(KERNEL) {
	case "Linux", cli.LINUX:
		bin = kit.Keys(bin, cli.LINUX)
	default:
		bin = kit.Keys(bin, cli.LINUX)
	}
	switch m.Option(ARCH) {
	case "i686", cli.X86:
		bin = kit.Keys(bin, cli.X86)
	default:
		bin = kit.Keys(bin, cli.AMD64)
	}
	kit.If(m.Option(web.DREAM), func() {
		m.Options(nfs.FROM, path.Join(kit.Select(ice.USR_LOCAL_WORK, "..", ice.Info.NodeType == web.WORKER), m.Option(web.DREAM), ice.USR_PUBLISH+bin))
		m.Options(nfs.PATH, m.Option(web.DREAM), nfs.FILE, ice.BIN_ICE_BIN)
	}, func() {
		m.Options(nfs.FROM, ice.USR_PUBLISH+bin, nfs.PATH, path.Base(kit.Path("")), nfs.FILE, ice.BIN_ICE_BIN)
	})
	m.Cmd(SSH_TRANS, tcp.SEND)
	s.shell(m, m.Template(PUSHBIN_SH), arg...)
	s.Modify(m, kit.Simple(m.OptionSimple(MACHINE, web.DREAM))...)
}
func (s relay) Install(m *ice.Message, arg ...string) {
	m.Options(web.DOMAIN, "https://shylinux.com", ice.MSG_USERPOD, m.Option(web.DREAM))
	m.Options(nfs.SOURCE, kit.Value(kit.UnMarshal(web.AdminCmd(m.Message, cli.RUNTIME)), "make.remote"))
	web.DreamList(m.Spawn().Message).Table(func(value ice.Maps) {
		kit.If(value[mdb.NAME] == m.Option(web.DREAM), func() { m.Option(nfs.SOURCE, value[nfs.REPOS]) })
	})
	s.shell(m, m.Cmdx(web.CODE_PUBLISH, ice.CONTEXTS, nfs.SOURCE, kit.Dict("format", "raw")), arg...)
	s.Modify(m, kit.Simple(m.OptionSimple(MACHINE, web.DREAM))...)
}
func (s relay) Upgrade(m *ice.Message, arg ...string) {
	if len(arg) == 0 && (m.Option(MACHINE) == "" || strings.Contains(m.Option(MACHINE), ",")) {
		s.foreach(m, func(msg *ice.Message, cmd []string) {
			ssh.PushShell(msg.Message, strings.Split(msg.Template(UPGRADE_SH), lex.NL), func(res string) {
				web.PushNoticeGrow(m.Options(ctx.DISPLAY, web.PLUGIN_XTERM).Message, res)
			})
		})
	} else {
		s.shell(m, m.Template(UPGRADE_SH), arg...)
	}
}
func (s relay) Version(m *ice.Message, arg ...string) {
	s.foreach(m, func(msg *ice.Message, cmd []string) {
		ssh.PushShell(msg.Message, strings.Split(msg.Template(VERSION_SH), lex.NL), func(res string) {
			web.PushNoticeGrow(m.Options(ctx.DISPLAY, web.PLUGIN_XTERM).Message, res)
		})
	})
}
func (s relay) AdminCmd(m *ice.Message, arg ...string) {
	s.shell(m, s.admins(m, m.Option(ice.CMD)), arg...)
}
func (s relay) Xterm(m *ice.Message, arg ...string) { s.Code.Xterm(m, m.Option(MACHINE), arg...) }
func (s relay) Repos(m *ice.Message, arg ...string) { s.iframeCmd(m, web.CODE_GIT_STATUS, arg...) }
func (s relay) Vimer(m *ice.Message, arg ...string) { s.iframeCmd(m, web.CODE_VIMER, arg...) }
func (s relay) Admin(m *ice.Message, arg ...string) { s.iframeCmd(m, web.ADMIN, arg...) }

func init() { ice.Cmd(SSH_RELAY, relay{}) }

func (s relay) xterm(m *ice.Message) {
	xterm.AddCommand(m.Option(MACHINE), func(msg *icebergs.Message, arg ...string) (x xterm.XTerm, e error) {
		m.GoWait(func(done func()) {
			e = ssh.Shell(m.Message, func(_x xterm.XTerm) {
				defer done()
				kit.If(len(arg) > 1 && arg[0] == ice.INIT, func() { m.Sleep300ms(); _x.Write([]byte(strings.TrimSpace(arg[1]) + lex.NL)) })
				x = _x
			})
			if e != nil {
				defer done()
				x, e = xterm.Command(m.Message, "", m.OptionDefault(SHELL, code.SH))
				m.GoSleep300ms(func() { x.Write([]byte(kit.Format("ssh-copy-id %s@%s\n", m.Option(aaa.USERNAME), m.Option(tcp.HOST)))) })
			}
		})
		return
	})
}
func (s relay) shell(m *ice.Message, init string, arg ...string) {
	s.Code.Xterm(m, kit.JoinWord(m.Option(MACHINE), ice.INIT, kit.Format("%q", strings.ReplaceAll(init, lex.NL, "; "))), arg...)
}
func (s relay) foreach(m *ice.Message, cb func(*ice.Message, []string)) {
	cmd := kit.Filters(strings.Split(m.Option(ice.CMD), lex.NL), "")
	s.Hash.ForEach(m, MACHINE, func(msg *ice.Message) { cb(msg, cmd) })
}
func (s relay) foreachModify(m *ice.Message, machine, key, cmd string, cb func([]string) string) {
	kit.If(cb == nil, func() { cb = func(ls []string) string { return kit.Join(ls) } })
	s.ForEach(m.Spawn(ice.Maps{MACHINE: machine, ice.CMD: cmd, ice.MSG_DAEMON: ""})).Table(func(value ice.Maps) {
		s.Modify(m, MACHINE, value[MACHINE], key, cb(kit.Split(value[ice.RES])), mdb.TIME, time.Now().Format(ice.MOD_TIME))
		m.Push(mdb.TIME, time.Now().Format(ice.MOD_TIME)).Push(MACHINE, value[MACHINE]).Push(tcp.HOST, value[tcp.HOST])
		m.Push(ice.CMD, cmd).Push(ice.RES, value[ice.RES]).PushButton(s.ForFlow, s.ForEach)
	})
}
func (s relay) iframeCmd(m *ice.Message, cmd string, arg ...string) {
	if p := kit.MergeURL2(m.Option(web.LINK), web.C(cmd)); kit.HasPrefix(m.Option(ice.MSG_USERWEB), "https://") {
		s.Code.Iframe(m, m.Option(MACHINE), p, arg...)
	} else {
		m.ProcessOpen(p)
	}
}
func (s relay) admin(m *ice.Message, arg ...string) string {
	return kit.JoinWord(kit.Simple(ice.BIN_ICE_BIN, web.ADMIN, arg)...)
}
func (s relay) admins(m *ice.Message, arg ...string) string {
	return kit.Select(ice.CONTEXTS, m.Option(web.DREAM)) + nfs.PS + s.admin(m, arg...)
}
