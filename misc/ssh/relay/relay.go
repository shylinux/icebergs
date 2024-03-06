package relay

import (
	"net/http"
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
	SHELL      = "shell"
	KERNEL     = "kernel"
	ARCH       = "arch"
	VCPU       = "vcpu"
	NCPU       = "ncpu"
	MHZ        = "mhz"
	MEM        = "mem"
	MEM_TOTAL  = "mem_total"
	MEM_FREE   = "mem_free"
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
	field       string `data:"time,icons,machine,username,host,port,portal,dream,module,version,commitTime,compileTime,bootTime,go,git,package,shell,kernel,arch,vcpu,ncpu,mhz,mem,disk,network,listen,socket,proc,vendor"`
	create      string `name:"create host* port=22 username machine icons"`
	stats       string `name:"stats machine" help:"采集" icon:"bi bi-card-list"`
	publish     string `name:"publish" help:"发布" icon:"bi bi-send-check"`
	pubkey      string `name:"pubkey" help:"公钥" icon:"bi bi-person-vcard"`
	forEach     string `name:"forEach machine cmd*:textarea=pwd"`
	forFlow     string `name:"forFlow machine cmd*:textarea=pwd"`
	statsTables string `name:"statsTables" event:"stats.tables"`
	list        string `name:"list machine auto" help:"机器" icon:"relay.png"`
	install     string `name:"install dream param='forever start' dev portal=9020 nodename" help:"安装"`
	pushbin     string `name:"pushbin dream param='forever start' dev portal=9020 nodename" help:"部署" icon:"bi bi-box-arrow-in-up"`
	adminCmd    string `name:"adminCmd cmd" help:"命令" icon:"bi bi-terminal-plus"`
}

func (s relay) Init(m *ice.Message, arg ...string) {
	s.Hash.Init(m).TransInput(MACHINE, "机器",
		PACKAGE, "软件包", SHELL, "命令行", KERNEL, "内核", ARCH, "架构", VCPU, "虚拟核", NCPU, "处理器", MHZ, "频率",
		MEM, "内存", DISK, "磁盘", NETWORK, "流量", LISTEN, "服务", SOCKET, "连接", PROC, "进程",
		ice.DEV, "上位机", tcp.NODENAME, "节点名",
	)
	msg := m.Spawn()
	m.GoSleep3s(func() { s.Hash.List(msg).Table(func(value ice.Maps) { s.xterm(msg.Spawn(value)) }) })
}
func (s relay) Inputs(m *ice.Message, arg ...string) {
	switch s.Hash.Inputs(m, arg...); arg[0] {
	case MACHINE:
		if m.Option(ctx.ACTION) == mdb.CREATE {
			m.Message.Copy(m.Cmd(web.SPIDE).CutTo(web.CLIENT_NAME, arg[0]))
		}
	case aaa.USERNAME:
		m.Message.Copy(m.Cmd(aaa.USER).Cut(aaa.USERNAME).Push(arg[0], aaa.ROOT).Push(arg[0], ice.SHY))
	case tcp.HOST:
		m.Message.Copy(m.Options(ice.MSG_FIELDS, web.CLIENT_HOSTNAME).Cmd(web.SPIDE).CutTo(web.CLIENT_HOSTNAME, arg[0]))
	case tcp.PORT:
		m.Push(arg[0], tcp.PORT_22, tcp.PORT_9022)
	case cli.PARAM:
		m.Push(arg[0], `forever start`)
	case ice.DEV:
		s.Hash.List(m).CutTo(web.LINK, arg[0]).Push(arg[0], "http://localhost:9020")
	case web.PORTAL:
		kit.If(m.Option(tcp.LISTEN), func(p string) { m.Push(arg[0], kit.Split(p)) })
		m.Push(arg[0], tcp.PORT_443, tcp.PORT_80, tcp.PORT_9020, "9030", "9040", "9050")
	}
}
func (s relay) Create(m *ice.Message, arg ...string) {
	s.Hash.Create(m, kit.Simple(arg,
		tcp.PORT, m.OptionDefault(tcp.PORT, tcp.PORT_22),
		aaa.USERNAME, m.OptionDefault(aaa.USERNAME, m.Option(ice.MSG_USERNAME)),
		tcp.MACHINE, m.OptionDefault(tcp.MACHINE, kit.Split(m.Option(tcp.HOST), nfs.PT)[0]),
		mdb.ICONS, m.OptionDefault(mdb.ICONS, html.ICONS_SSH),
	)...)
}
func (s relay) Stats(m *ice.Message) {
	cmds := []string{
		cli.GO, `go version`, cli.GIT, `git version`,
		PACKAGE, `if yum -h &>/dev/null; then echo yum; elif apk version &>/dev/null; then echo apk; elif opkg -v &>/dev/null; then echo opkg; elif apt -h &>/dev/null; then echo apt; fi`,
		SHELL, `echo $SHELL`, KERNEL, `uname -s`, ARCH, `uname -m`,
		VCPU, `cat /proc/cpuinfo | grep "processor" | sort | uniq | wc -l`,
		NCPU, `cat /proc/cpuinfo | grep "physical id" | sort | uniq | wc -l`,
		MHZ, `cat /proc/cpuinfo | grep "\(cpu MHz\|BogoMIPS\)" | grep -o "[0-9.]\+" | sort -nr | head -n1`,
		MEM, `cat /proc/meminfo | head -n3 | grep -o "[0-9.]*"`, DISK, `df | grep "^/dev/"`,
		NETWORK, `cat /proc/net/dev | grep "eth0\|eth1" | sort -r | head -n1`,
		LISTEN, `netstat -ntl 2>/dev/null | grep "^tcp" | grep -o ":[0-9]\+" | grep -o "[0-9]\+" | sort -n | uniq`,
		SOCKET, `netstat -nt 2>/dev/null | grep "^tcp" | wc -l`, PROC, `ps aux | wc -l 2>/dev/null`,
	}
	trans := map[string]func([]string) string{
		cli.GO:  func(ls []string) string { return kit.TrimPrefix(kit.Select("", ls, 2), cli.GO) },
		cli.GIT: func(ls []string) string { return kit.Select("", ls, 2) },
		MEM: func(ls []string) string {
			return kit.FmtSize(kit.Int(kit.Select("", ls, 2))*1024, kit.Int(kit.Select("", ls, 0))*1024)
		},
		DISK: func(ls []string) string {
			return kit.FmtSize(kit.Int(kit.Select("", ls, 2))*1024, kit.Int(kit.Select("", ls, 1))*1024)
		},
		NETWORK: func(ls []string) string {
			return kit.FmtSize(kit.Int(kit.Select("", ls, 1)), kit.Int(kit.Select("", ls, 9)))
		},
		PROC: func(ls []string) string { return kit.Format(kit.Int(ls[0])) },
	}
	machine := m.Option(MACHINE)
	m.GoToast(func(toast func(string, int, int)) {
		kit.For(cmds, func(key, value string, index int) {
			toast(key, index/2, len(cmds)/2)
			s.foreachModify(m, machine, key, value, trans[key])
		})
	}).ProcessInner()
	s.foreach(m.Spawn(ice.Maps{MACHINE: machine}), func(msg *ice.Message, cmd []string) {
		ssh.CombinedOutput(msg.Message, s.admins(msg, cli.RUNTIME), func(res string) {
			if !strings.HasPrefix(res, "warn: ") {
				s.Modify(m, kit.Simple(MACHINE, msg.Option(MACHINE), kit.Dict(cli.ParseMake(res)))...)
			}
		})
	})
}
func (s relay) Publish(m *ice.Message, arg ...string) {
	if m.Option(MACHINE) == "" {
		s.Hash.ForEach(m, "", func(msg *ice.Message) { s.Publish(msg) })
		m.Cmdy(nfs.DIR, ice.USR_PUBLISH).Set(ctx.ACTION)
		return
	}
	kit.If(!nfs.Exists(m.Message, path.Join(ice.USR_PUBLISH, RELAY)), func() { s.Compile(m) })
	os.Symlink(RELAY, ice.USR_PUBLISH+m.Option(MACHINE))
	m.Cmd(nfs.SAVE, kit.HomePath(".ssh/"+m.Option(MACHINE)+".json"), kit.Formats(kit.Dict(m.OptionSimple("username,host,port")))+ice.NL)
}
func (s relay) Compile(m *ice.Message) {
	m.Cmdy(code.COMPILE, SRC_RELAY_GO, path.Join(ice.USR_PUBLISH, RELAY)).ProcessInner()
}
func (s relay) Pubkey(m *ice.Message, arg ...string) {
	m.EchoScript(m.Cmdx(nfs.CAT, kit.HomePath(ssh.ID_RSA_PUB))).ProcessInner()
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
			web.PushNoticeGrow(m.Options(ctx.DISPLAY, html.PLUGIN_XTERM).Message, res)
		})
	})
}
func (s relay) StatsTables(m *ice.Message, arg ...string) {
	web.PushStats(m.Message, "", mdb.HashSelects(m.Spawn().Message).Length(), "", "服务器数量")
}
func (s relay) List(m *ice.Message, arg ...string) *ice.Message {
	if s.Hash.List(m, arg...); len(arg) == 0 {
		if m.Length() == 0 {
			m.Action(s.Create)
		} else {
			m.Action(s.Create, s.Upgrade, s.Version, s.Stats, s.Publish)
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
		m.PushButton(s.Portal, s.Desktop, s.Dream, s.Admin, s.Vimer, s.Login, s.AdminCmd, s.Upgrade, s.Pushbin, s.Xterm, s.Remove)
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
func (s relay) Install(m *ice.Message, arg ...string) {
	m.Options(web.DOMAIN, m.SpideOrigin(ice.SHY), ice.MSG_USERPOD, m.Option(web.DREAM))
	m.Options(nfs.SOURCE, kit.Value(kit.UnMarshal(m.AdminCmd(cli.RUNTIME).Result()), "make.remote"))
	m.Spawn().DreamList().Table(func(value ice.Maps) {
		kit.If(value[mdb.NAME] == m.Option(web.DREAM), func() { m.Option(nfs.SOURCE, value[nfs.REPOS]) })
	})
	s.shell(m, m.PublishScript(nfs.SOURCE)+lex.SP+kit.JoinCmds(ice.DEV, m.Option(ice.DEV), tcp.PORT, m.Option(web.PORTAL), tcp.NODENAME, m.OptionDefault(tcp.NODENAME, m.Option(MACHINE))), arg...)
	s.Modify(m, kit.Simple(m.OptionSimple(MACHINE, web.DREAM, web.PORTAL))...)
}
func (s relay) Upgrade(m *ice.Message, arg ...string) { s.foreachScript(m, UPGRADE_SH, arg...) }
func (s relay) Version(m *ice.Message, arg ...string) { s.foreachScript(m, VERSION_SH, arg...) }
func (s relay) Pushbin(m *ice.Message, arg ...string) {
	bin := "ice"
	switch strings.ToLower(m.Option(KERNEL)) {
	case cli.LINUX:
		bin = kit.Keys(bin, cli.LINUX)
	default:
		bin = kit.Keys(bin, cli.LINUX)
	}
	switch m.Option(ARCH) {
	case cli.X86, "i686":
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
	s.shell(m, m.Template(PUSHBIN_SH)+lex.SP+kit.JoinCmds(ice.DEV, m.Option(ice.DEV), tcp.PORT, m.Option(web.PORTAL), tcp.NODENAME, m.OptionDefault(tcp.NODENAME, m.Option(MACHINE))), arg...)
	s.Modify(m, kit.Simple(m.OptionSimple(MACHINE, web.DREAM, web.PORTAL))...)
}

func (s relay) AdminCmd(m *ice.Message, arg ...string) {
	s.shell(m, "cd "+kit.Select(ice.CONTEXTS, m.Option(web.DREAM))+"; "+s.admin(m, m.Option(ice.CMD)), arg...)
}
func (s relay) Xterm(m *ice.Message, arg ...string) {
	m.ProcessXterm(kit.JoinWord(m.Option(MACHINE), ice.INIT, kit.Format("%q", "cd "+kit.Select(ice.CONTEXTS, m.Option(web.DREAM)))), arg...)
}
func (s relay) Login(m *ice.Message, arg ...string) {
	if m.Options(s.Hash.List(m.Spawn(), m.Option(MACHINE)).AppendSimple()); m.Option(ice.BACK) == "" {
		defer m.ToastProcess()()
		ssh.CombinedOutput(m.Message, s.admins(m, kit.JoinCmds(web.HEADER, mdb.CREATE,
			"--", mdb.TYPE, "oauth", mdb.NAME, m.CommandKey(), mdb.ICONS, html.ICONS_SSH, mdb.ORDER, "100",
			web.LINK, m.MergePodCmd("", "", ctx.ACTION, m.ActionKey(), MACHINE, m.Option(MACHINE)),
		)), func(res string) { m.ProcessHold() })
		m.ProcessOpen(kit.MergeURL2(m.Option(mdb.LINK), web.C(web.HEADER)))
	} else if m.Option(ice.MSG_METHOD) == http.MethodGet {
		m.EchoInfoButton("")
	} else {
		defer m.ToastProcess()()
		ssh.CombinedOutput(m.Message, s.admins(m, kit.JoinCmds(web.SHARE, mdb.CREATE, mdb.TYPE, aaa.LOGIN, "--", mdb.TEXT, m.Option(ice.BACK))), func(res string) {
			m.ProcessReplace(kit.MergeURL2(m.Option(ice.BACK), "/share/"+strings.TrimSpace(res)))
		})
	}
}
func (s relay) Repos(m *ice.Message, arg ...string)   { s.iframe(m, web.CODE_GIT_STATUS, arg...) }
func (s relay) Vimer(m *ice.Message, arg ...string)   { s.iframe(m, "", arg...) }
func (s relay) Admin(m *ice.Message, arg ...string)   { s.iframe(m, "", arg...) }
func (s relay) Dream(m *ice.Message, arg ...string)   { s.iframe(m, "", arg...) }
func (s relay) Desktop(m *ice.Message, arg ...string) { s.iframe(m, "", arg...) }
func (s relay) Portal(m *ice.Message, arg ...string)  { s.iframe(m, "", arg...) }

func init() { ice.Cmd(SSH_RELAY, relay{}) }

func (s relay) iframe(m *ice.Message, cmd string, arg ...string) {
	p := kit.MergeURL2(m.Option(web.LINK), web.C(kit.Select(m.ActionKey(), cmd)))
	if strings.HasPrefix(m.Option(ice.MSG_USERWEB), ice.HTTPS) {
		m.ProcessIframe(m.Option(MACHINE), p, arg...)
	} else {
		m.ProcessOpen(p)
	}
}
func (s relay) shell(m *ice.Message, init string, arg ...string) {
	m.ProcessXterm(kit.JoinWord(m.Option(MACHINE), ice.INIT, kit.Format("%q", strings.ReplaceAll(init, lex.NL, "; "))), arg...)
}
func (s relay) foreachScript(m *ice.Message, script string, arg ...string) {
	m.Option(ice.MSG_TITLE, kit.Keys(m.Option(ice.MSG_USERPOD), m.CommandKey(), m.ActionKey()))
	if len(arg) == 0 && (m.Option(MACHINE) == "" || strings.Contains(m.Option(MACHINE), ",")) {
		s.foreach(m, func(msg *ice.Message, cmd []string) {
			if msg.Option(cli.GO) == "" {
				return
			}
			ssh.PushShell(msg.Message, strings.Split(msg.Template(script), lex.NL), func(res string) {
				web.PushNoticeGrow(m.Options(ctx.DISPLAY, html.PLUGIN_XTERM, ice.MSG_COUNT, "0", ice.MSG_DEBUG, ice.FALSE, ice.LOG_DISABLE, ice.TRUE).Message, res)
			})
		})
	} else {
		s.shell(m, m.Template(script), arg...)
	}
}
func (s relay) foreachModify(m *ice.Message, machine, key, cmd string, cb func([]string) string) {
	kit.If(cb == nil, func() { cb = func(ls []string) string { return kit.Join(ls) } })
	s.ForEach(m.Spawn(ice.Maps{MACHINE: machine, ice.CMD: cmd, ice.MSG_DAEMON: ""})).Table(func(value ice.Maps) {
		s.Modify(m, MACHINE, value[MACHINE], key, cb(kit.Split(value[ice.RES])), mdb.TIME, time.Now().Format(ice.MOD_TIME))
		m.Push(mdb.TIME, time.Now().Format(ice.MOD_TIME)).Push(MACHINE, value[MACHINE]).Push(tcp.HOST, value[tcp.HOST])
		m.Push(ice.CMD, cmd).Push(ice.RES, value[ice.RES]).PushButton(s.ForFlow, s.ForEach)
	})
}
func (s relay) foreach(m *ice.Message, cb func(*ice.Message, []string)) {
	cmd := kit.Filters(strings.Split(m.Option(ice.CMD), lex.NL), "")
	s.Hash.ForEach(m, MACHINE, func(msg *ice.Message) { cb(msg, cmd) })
}
func (s relay) admins(m *ice.Message, arg ...string) string {
	return kit.Select(ice.CONTEXTS, m.Option(web.DREAM)) + nfs.PS + s.admin(m, arg...)
}
func (s relay) admin(m *ice.Message, arg ...string) string {
	return kit.JoinWord(kit.Simple(ice.BIN_ICE_BIN, web.ADMIN, arg)...)
}
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
