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
	RELAY          = "relay"
	SSH_RELAY      = "ssh.relay"
	SRC_RELAY_GO   = "src/relay.go"
	SSH_CONNECT    = "ssh.connect"
	SSH_AUTHORIZED = ".ssh/authorized_keys"

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
	checkbox      string `data:"true"`
	export        string `data:"true"`
	short         string `data:"machine"`
	tools         string `data:"ssh.trans,ssh.auths,aaa.rsa"`
	field         string `data:"time,icons,machine,username,host,port,portal,dream,module,version,commitTime,compileTime,bootTime,go,git,package,shell,kernel,arch,vcpu,ncpu,mhz,mem,disk,network,listen,socket,proc,vendor"`
	create        string `name:"create host* port=22 username=root machine icons"`
	upgrade       string `name:"upgrade machine"`
	stats         string `name:"stats machine" icon:"bi bi-card-list"`
	publish       string `name:"publish" icon:"bi bi-send-check"`
	forEach       string `name:"forEach machine cmd*:textarea=pwd"`
	forFlow       string `name:"forFlow machine cmd*:textarea=pwd"`
	list          string `name:"list machine auto" help:"机器" icon:"relay.png"`
	opsOriginOpen string `name:"opsOriginOpen" event:"ops.origin.open"`
	opsServerOpen string `name:"opsServerOpen" event:"ops.server.open"`
	opsDreamSpawn string `name:"opsDreamSpawn" event:"ops.dream.spawn"`
	install       string `name:"install dream portal nodename dev"`
	pushbin       string `name:"pushbin dream portal nodename dev" icon:"bi bi-box-arrow-in-up"`
	adminCmd      string `name:"adminCmd cmd" icon:"bi bi-terminal-plus"`
	pushkey       string `name:"pushkey server" icon:"bi bi-person-fill-up"`
}

func (s relay) Init(m *ice.Message, arg ...string) {
	s.Hash.Init(m, arg...)
	xterm.AddCommand(RELAY, func(m *icebergs.Message, arg ...string) (xterm.XTerm, error) {
		m.Cmd(SSH_CONNECT, tcp.DIAL, mdb.NAME, m.Option(mdb.NAME, arg[1]), arg)
		return ssh.NewSession(m, arg[1])
	})
}
func (s relay) Inputs(m *ice.Message, arg ...string) {
	switch s.Hash.Inputs(m, arg...); arg[0] {
	case MACHINE:
		if m.Option(ctx.ACTION) == mdb.CREATE {
			m.CmdInputs(web.SPIDE, web.CLIENT_NAME, arg[0])
		}
	case aaa.USERNAME:
		m.CmdInputs(aaa.USER, aaa.USERNAME, arg[0]).Push(arg[0], aaa.ROOT).Push(arg[0], ice.SHY)
	case tcp.HOST:
		m.CmdInputs(web.SPIDE, web.CLIENT_HOSTNAME, arg[0])
	case tcp.PORT:
		m.Push(arg[0], tcp.PORT_22, tcp.PORT_9022)
	case cli.PARAM:
		m.Push(arg[0], `forever start`)
	case ice.DEV:
		m.Cmdy("").CutTo(web.LINK, arg[0])
		m.CmdInputs(web.SPIDE, web.CLIENT_ORIGIN, arg[0])
		m.Push(arg[0], m.UserHost())
	case web.PORTAL:
		kit.If(m.Option(tcp.LISTEN), func(p string) { m.Push(arg[0], kit.Split(p)) })
		m.Push(arg[0], tcp.PORT_443, tcp.PORT_80, tcp.PORT_9020, "9030", "9040", "9050")
	case tcp.NODENAME:
		m.Cmdy("").CutTo(MACHINE, arg[0])
	case web.SERVER:
		m.Copy(m.AdminCmd(web.DREAM, web.SERVER))
		ctx.DisplayInputKey(m.Message, "style", "_nameicon")
	}
}
func (s relay) Create(m *ice.Message, arg ...string) {
	s.Hash.Create(m, kit.Simple(arg, tcp.PORT, m.OptionDefault(tcp.PORT, tcp.PORT_22),
		tcp.MACHINE, m.OptionDefault(tcp.MACHINE, kit.Split(m.Option(tcp.HOST), nfs.PT)[0]),
		mdb.ICONS, m.OptionDefault(mdb.ICONS, html.ICONS_SSH),
	)...)
}
func (s relay) Stats(m *ice.Message) {
	cmds := []string{cli.GO, `go version`, cli.GIT, `git version`,
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
		m.Cmdy(nfs.DIR, ice.USR_PUBLISH).PushAction()
		return
	}
	if m.Option(ice.INIT, ""); m.Option(web.PORTAL) != "" {
		m.Option(ice.INIT, kit.Format("cd %s", path.Base(m.DreamPath(m.Option(web.DREAM)))))
	}
	m.Cmd(nfs.SAVE, kit.HomePath(".ssh/"+m.Option(MACHINE)+".json"), kit.Formats(kit.Dict(m.OptionSimple("username,host,port,init")))+ice.NL)
	kit.If(!m.Exists(path.Join(ice.USR_PUBLISH, RELAY)), func() { s.Compile(m) })
	os.Symlink(RELAY, ice.USR_PUBLISH+m.Option(MACHINE))
}
func (s relay) Compile(m *ice.Message) {
	m.Cmdy(code.COMPILE, SRC_RELAY_GO, path.Join(ice.USR_PUBLISH, RELAY)).ProcessInner()
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
func (s relay) List(m *ice.Message, arg ...string) *ice.Message {
	if s.Hash.List(m, arg...); len(arg) == 0 {
		if m.Length() == 0 {
			m.Action(s.Create)
		} else {
			m.Action(s.Create, s.Upgrade, s.Version, s.Stats)
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
			m.Push(web.LINK, "").PushButton(s.Xterm, s.Pushbin, s.Install, s.Remove)
		} else {
			m.Push(web.LINK, m.HostPort(value[tcp.HOST], value[web.PORTAL]))
			m.PushButton(s.Portal, s.Desktop, s.Dream, s.Admin, s.Open, s.Status, s.Vimer, s.Login, s.Pushkey, s.Spide, s.AdminCmd, s.Xterm, s.Pushbin, s.Upgrade, s.Remove)
			kit.If(len(arg) > 0, func() { m.PushQRCode(cli.QRCODE, m.Append(web.LINK)) })
		}
	})
	m.RewriteAppend(func(value, key string, index int) string {
		switch key {
		case MEM:
			if ls := kit.Split(value, " /"); len(ls) > 0 && kit.Int(ls[0]) < 256*1024*1024 {
				value = html.FormatDanger(value)
			}
		case DISK:
			if ls := kit.Split(value, " /"); len(ls) > 1 && kit.Int(ls[0])*100/kit.Int(ls[1]) > 80 {
				value = html.FormatDanger(value)
			}
		}
		return value
	})
	_stats := kit.Dict(MEM, kit.FmtSize(stats[MEM_FREE], stats[MEM_TOTAL]), DISK, kit.FmtSize(stats[DISK_USED], stats[DISK_TOTAL]))
	m.StatusTimeCount(m.Spawn().Options(stats, _stats).OptionSimple(VCPU, MEM, DISK, SOCKET, PROC))
	return m
}
func (s relay) OpsOriginOpen(m *ice.Message, arg ...string) {
	kit.If(m.Option(nfs.MODULE) == ice.Info.Make.Module, func() { s.sendData(m, m.Option(mdb.NAME)) })
}
func (s relay) OpsServerOpen(m *ice.Message, arg ...string) {
	kit.If(m.Option(nfs.MODULE) == ice.Info.Make.Module, func() { s.sendData(m, m.Option(mdb.NAME)) })
}
func (s relay) OpsDreamSpawn(m *ice.Message, arg ...string) {
	kit.If(m.Option(mdb.NAME) == ice.Info.NodeName, func() { s.sendData(m, kit.Keys(m.Option(web.DOMAIN), m.Option(mdb.NAME))) })
}
func (s relay) sendData(m *ice.Message, space string) {
	if m.IsTech() {
		m.Cmd("").Table(func(value ice.Maps) {
			m.AdminCmd(web.SPACE, space, m.PrefixKey(), mdb.CREATE, tcp.HOST, "", kit.Simple(value))
			m.Cmd(SSH_AUTHS, mdb.INSERT, value[MACHINE], space, kit.Dict(ice.SPACE_NOECHO, ice.FALSE))
		})
	}
}
func (s relay) Install(m *ice.Message, arg ...string) {
	m.Options(web.DOMAIN, m.SpideOrigin(ice.DEV), ice.MSG_USERPOD, m.Option(web.DREAM), nfs.SOURCE, m.DreamRepos(m.Option(web.DREAM)))
	s.Modify(m, m.OptionSimple(MACHINE, web.DREAM, web.PORTAL)...)
	s.shell(m, m.PublishScript(nfs.SOURCE)+lex.SP+s.param(m), arg...)
}
func (s relay) Upgrade(m *ice.Message, arg ...string) { s.foreachScript(m, UPGRADE_SH, arg...) }
func (s relay) Version(m *ice.Message, arg ...string) { s.foreachScript(m, VERSION_SH, arg...) }
func (s relay) Pushbin(m *ice.Message, arg ...string) {
	if kit.HasPrefixList(arg, ctx.RUN) {
		m.ProcessXterm("", nil, arg...)
		return
	}
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
	dream := m.DreamPath(m.Option(web.DREAM))
	m.Options(nfs.FROM, path.Join(dream, ice.USR_PUBLISH+bin), nfs.PATH, path.Base(dream), nfs.FILE, ice.BIN_ICE_BIN)
	if m.Cmd(SSH_TRANS, tcp.SEND); m.OptionDefault(web.PORTAL, tcp.PORT_9020) == tcp.PORT_443 {
		msg := m.Cmd(aaa.CERT, mdb.CREATE, m.Option(tcp.HOST))
		m.Cmd(SSH_TRANS, tcp.SEND, nfs.FROM, msg.Append(ssh.PEM), nfs.FILE, nfs.ETC_CERT_PEM)
		m.Cmd(SSH_TRANS, tcp.SEND, nfs.FROM, msg.Append(ssh.KEY), nfs.FILE, nfs.ETC_CERT_KEY)
	}
	s.Modify(m, kit.Simple(m.OptionSimple(MACHINE, web.DREAM, web.PORTAL))...)
	s.shell(m, "export ctx_dev="+m.SpideOrigin(ice.DEV)+"; "+m.Template(PUSHBIN_SH)+lex.SP+s.param(m), arg...)
}
func (s relay) Xterm(m *ice.Message, arg ...string) {
	init := ""
	kit.If(m.Option(web.PORTAL), func() { init = "cd " + path.Base(m.DreamPath(m.Option(web.DREAM))) })
	s.shell(m, init, arg...)
}
func (s relay) AdminCmd(m *ice.Message, arg ...string) {
	s.shell(m, s.admins(m, m.Option(ice.CMD)), arg...)
}
func (s relay) Spide(m *ice.Message, arg ...string) {
	ssh.CombinedOutput(m.Message, s.admins(m, kit.JoinCmds(web.TOKEN, mdb.CREATE, "--", mdb.TYPE, web.SERVER, mdb.NAME, aaa.ROOT, mdb.TEXT, ice.Info.Hostname)), func(res string) {
		m.AdminCmd(web.SPIDE, mdb.CREATE, m.Option(web.LINK), m.Option(MACHINE), "", nfs.REPOS, strings.TrimSpace(res))
		m.AdminCmd(web.SPACE, tcp.DIAL, m.Option(MACHINE))
	})
}
func (s relay) Pushkey(m *ice.Message, arg ...string) {
	p := kit.Format("/home/%s/"+SSH_AUTHORIZED, m.Option(aaa.USERNAME))
	kit.If(m.Option(aaa.USERNAME) == aaa.ROOT, func() { p = kit.Format("/root/" + SSH_AUTHORIZED) })
	list := kit.Split(m.AdminCmdx(web.SPACE, m.Option(MACHINE), nfs.CAT, p), lex.NL, lex.NL)
	if key := ssh.PublicKey(m.Message, m.Option(web.SERVER)); !kit.IsIn(key, list...) {
		m.AdminCmd(web.SPACE, m.Option(MACHINE), nfs.PUSH, p, key+lex.NL)
		m.Echo(m.AdminCmdx(web.SPACE, m.Option(MACHINE), nfs.CAT, p)).ProcessInner()
	} else {
		m.Echo(strings.Join(list, lex.NL)).ProcessInner()
	}
	m.ProcessInner()
}
func (s relay) Login(m *ice.Message, arg ...string) {
	if m.Options(m.Cmd("", m.Option(MACHINE)).AppendSimple()); m.Option(ice.BACK) == "" {
		defer m.ToastProcess()()
		ssh.CombinedOutput(m.Message, s.admins(m, kit.JoinCmds(web.HEADER, mdb.CREATE, "--",
			mdb.TYPE, "oauth", mdb.NAME, m.CommandKey(), mdb.ICONS, html.ICONS_SSH, mdb.ORDER, "100",
			web.LINK, m.MergePodCmd("", "", ctx.ACTION, m.ActionKey(), MACHINE, m.Option(MACHINE)),
		)), func(res string) { m.Echo(res) })
		m.ProcessOpen(m.Option(mdb.LINK))
	} else if m.Option(ice.MSG_METHOD) == http.MethodGet {
		m.EchoInfoButton("")
	} else {
		defer m.ToastProcess()()
		ssh.CombinedOutput(m.Message, s.admins(m, kit.JoinCmds(web.SHARE, mdb.CREATE, mdb.TYPE, aaa.LOGIN, "--", mdb.TEXT, m.Option(ice.BACK))), func(res string) {
			m.ProcessReplace(kit.MergeURL2(m.Option(ice.BACK), "/share/"+strings.TrimSpace(res)))
		})
	}
}
func (s relay) Vimer(m *ice.Message, arg ...string)   { s.iframe(m, "", arg...) }
func (s relay) Status(m *ice.Message, arg ...string)  { s.iframe(m, "", arg...) }
func (s relay) Open(m *ice.Message, arg ...string)    { m.ProcessOpen(m.Option(web.LINK)) }
func (s relay) Admin(m *ice.Message, arg ...string)   { s.iframe(m, "", arg...) }
func (s relay) Dream(m *ice.Message, arg ...string)   { s.iframe(m, "", arg...) }
func (s relay) Desktop(m *ice.Message, arg ...string) { s.iframe(m, "", arg...) }
func (s relay) Portal(m *ice.Message, arg ...string)  { s.iframe(m, "", arg...) }

func init() { ice.Cmd(SSH_RELAY, relay{}) }

func (s relay) iframe(m *ice.Message, cmd string, arg ...string) {
	p := kit.MergeURL2(m.Option(web.LINK), web.C(kit.Select(m.ActionKey(), cmd)))
	if m.IsChromeUA() && m.UserWeb().Scheme == ice.HTTP && strings.HasPrefix(p, ice.HTTPS) {
		m.ProcessOpen(p)
	} else {
		m.ProcessIframe(m.Option(MACHINE), p, arg...)
	}
}
func (s relay) shell(m *ice.Message, init string, arg ...string) {
	m.ProcessXterm(kit.Keys(m.Option(MACHINE), m.ActionKey()), s.CmdArgs(m, init), arg...)
}
func (s relay) foreachScript(m *ice.Message, script string, arg ...string) {
	m.Option(ice.MSG_TITLE, kit.Keys(m.Option(ice.MSG_USERPOD), m.CommandKey(), m.ActionKey()))
	if !kit.HasPrefixList(arg, ctx.RUN) && (m.Option(MACHINE) == "" || strings.Contains(m.Option(MACHINE), ",")) {
		s.foreach(m, func(msg *ice.Message, cmd []string) {
			if msg.Option(cli.GO) == "" {
				return
			}
			msg.Option(web.DREAM, path.Base(m.DreamPath(msg.Option(web.DREAM))))
			msg.Option(web.LINK, m.HostPort(msg.Option(tcp.HOST), msg.Option(web.PORTAL)))
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
	return path.Base(m.DreamPath(m.Option(web.DREAM))) + nfs.PS + s.admin(m, arg...)
}
func (s relay) admin(m *ice.Message, arg ...string) string {
	return kit.JoinWord(kit.Simple(ice.BIN_ICE_BIN, web.ADMIN, arg)...)
}
func (s relay) param(m *ice.Message, arg ...string) string {
	return kit.JoinCmdArgs(ice.DEV, m.Option(ice.DEV), tcp.PORT, m.Option(web.PORTAL), tcp.NODENAME, m.OptionDefault(tcp.NODENAME, m.Option(MACHINE)), ice.TCP_DOMAIN, m.Option(tcp.HOST))
}
func (s relay) CmdArgs(m *ice.Message, init string, arg ...string) string {
	kit.If(m.Option(web.PORTAL) != "" && init == "", func() { init = "cd " + path.Base(m.DreamPath(m.Option(web.DREAM))) })
	return strings.TrimPrefix(os.Args[0], kit.Path("")+nfs.PS) + " " + kit.JoinCmds(kit.Simple(
		SSH_CONNECT, tcp.OPEN, ssh.AUTHFILE, "", m.OptionSimple(aaa.USERNAME, tcp.HOST, tcp.PORT), ice.INIT, init, arg)...)
}

type Relay struct {
	relay
	pushbin string `name:"pushbin dream param='forever start' dev portal nodename" icon:"bi bi-box-arrow-in-up"`
}

func (s Relay) Cmds(m *ice.Message, host string, cmds string) *ice.Message {
	return m.Cmd(cli.SYSTEM, os.Args[0], SSH_CONNECT, tcp.OPEN, ssh.AUTHFILE, "", aaa.USERNAME, aaa.ROOT, tcp.HOST, host, tcp.PORT, tcp.PORT_22, ctx.CMDS, cmds)
}
func (s Relay) CmdsWait(m *ice.Message, host string, cmds string, res string) bool {
	for i := 0; i < 10; i++ {
		if strings.TrimSpace(s.Cmds(m.Sleep(cli.TIME_3s), host, cmds).Result()) == res {
			return true
		}
	}
	return false
}
func init() { ice.Cmd(SSH_RELAY, Relay{}) }
