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
	field       string `data:"time,machine,username,password,host,port,portal,module,version,commit,compile,boot,package,shell,kernel,arch,ncpu,vcpu,mhz,mem,disk,network,listen,socket,proc,vendor"`
	statsTables string `name:"statsTables" event:"stats.tables"`
	create      string `name:"create machine* username* password host* port*=22 portal vendor"`
	pubkey      string `name:"pubkey" help:"公钥"`
	version     string `name:"version" help:"版本"`
	stats       string `name:"stats machine" help:"采集"`
	forEach     string `name:"forEach machine cmd*:textarea=pwd" help:"遍历"`
	forFlow     string `name:"forFlow machine cmd*:textarea=pwd" help:"流程"`
	list        string `name:"list machine auto" help:"代理"`
	pushbin     string `name:"pushbin" help:"部署"`
	adminCmd    string `name:"adminCmd cmd" help:"命令"`
}

func (s relay) Init(m *ice.Message, arg ...string) {
	s.Hash.Init(m).TransInput(MACHINE, "机器",
		PACKAGE, "软件包", SHELL, "命令行", KERNEL, "内核", ARCH, "架构",
		NCPU, "处理器", VCPU, "虚拟核", MHZ, "频率",
		MEM, "内存", DISK, "磁盘", NETWORK, "流量",
		LISTEN, "服务", SOCKET, "连接", PROC, "进程",
		nfs.COMMIT, "发布时间", code.COMPILE, "编译时间",
		"boot", "启动时间",
	)
	msg := m.Spawn(ice.Maps{ice.MSG_FIELDS: ""})
	m.GoSleep3s(func() { s.Hash.List(msg).Table(func(value ice.Maps) { s.xterm(m.Spawn(value)) }) })
}
func (s relay) StatsTables(m *ice.Message, arg ...string) {
	web.PushStats(m.Message, "", mdb.HashSelects(m.Spawn().Message).Length(), "", "服务器数量")
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
		m.Push(arg[0], tcp.PORT_443, tcp.PORT_80, tcp.PORT_9020)
	}
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
	m.Cmd(nfs.SAVE, kit.HomePath(".ssh/"+m.Option(MACHINE)+".json"), kit.Formats(kit.Dict(m.OptionSimple("username,password,host,port")))+ice.NL)
}
func (s relay) Pubkey(m *ice.Message, arg ...string) {
	m.EchoScript(m.Cmdx(nfs.CAT, kit.HomePath(ssh.ID_RSA_PUB))).ProcessInner()
}
func (s relay) Stats(m *ice.Message) {
	cmds := []string{
		PACKAGE, `if yum -h &>/dev/null; then echo yum; elif apk version &>/dev/null; then echo apk; elif apt -h &>/dev/null; then echo apt; fi`,
		KERNEL, `uname -s`, ARCH, `uname -m`,
		NCPU, `cat /proc/cpuinfo | grep "physical id" | sort | uniq | wc -l`,
		VCPU, `cat /proc/cpuinfo | grep "processor" | sort | uniq | wc -l`,
		MHZ, `cat /proc/cpuinfo | grep "cpu MHz" | grep -o "[0-9.]\+" | sort -nr | head -n1`,
		MEM, `cat /proc/meminfo | head -n2 | grep -o "[0-9.]*"`,
		DISK, `df | grep "^/dev/"`,
		NETWORK, `cat /proc/net/dev | grep "eth0\|eth1" | sort -r | head -n1`,
		LISTEN, `netstat -ntl 2>/dev/null | grep "^tcp" | grep -o ":[0-9]\+" | grep -o "[0-9]\+" | sort -n | uniq`,
		SOCKET, `netstat -nt 2>/dev/null | grep "^tcp" | wc -l`,
		PROC, `ps aux | wc -l`,
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
			s.foreachModify(m, key, value, trans[key])
		})
		return nil
	}).ProcessInner()
	s.ForEach(m.Spawn(ice.Maps{MACHINE: machine, ice.CMD: kit.JoinWord("contexts/"+ice.BIN_ICE_BIN, web.ADMIN, cli.RUNTIME)})).Table(func(value ice.Maps) {
		res := kit.UnMarshal(value[ice.RES])
		data := kit.Value(res, cli.MAKE)
		s.Modify(m, kit.Simple(MACHINE, value[MACHINE], kit.Dict(
			nfs.MODULE, kit.Value(data, nfs.MODULE), nfs.VERSION, kit.Join(kit.TrimArg(kit.Simple(
				kit.Value(data, nfs.VERSION), kit.Value(data, "forword"), kit.Cut(kit.Format(kit.Value(data, mdb.HASH)), 6),
			)...), "-"),
			nfs.COMMIT, kit.Value(data, "when"), code.COMPILE, kit.Value(data, mdb.TIME), "boot", kit.Value(res, "boot.time"),
			SHELL, kit.Value(res, "conf.SHELL"), KERNEL, kit.Value(res, "host.GOOS"), ARCH, kit.Value(res, "host.GOARCH"),
		))...)
	})
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
		ssh.PushShell(msg.Message, cmd, func(res string) { web.PushNoticeGrow(m.Options(ctx.DISPLAY, web.PLUGIN_XTERM), res) })
	})
}
func (s relay) List(m *ice.Message, arg ...string) *ice.Message {
	if s.Hash.List(m, arg...); len(arg) == 0 {
		if m.Length() == 0 {
			m.Action(s.Create, s.Compile, s.Publish, s.Pubkey)
		} else {
			m.Action(s.Create, s.Upgrade, s.Version, s.Stats, s.ForEach, s.ForFlow, s.Compile, s.Publish, s.Pubkey)
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
			m.Push(web.LINK, "").PushButton(s.Xterm, s.AdminCmd, s.Pushbin, s.Install, s.Remove)
			return
		}
		m.PushButton(s.Admin, s.Vimer, s.Repos, s.Xterm, s.AdminCmd, s.Pushbin, s.Upgrade, s.Remove)
		switch value[web.PORTAL] {
		case tcp.PORT_443:
			m.Push(web.LINK, kit.Format("https://%s", value[tcp.HOST]))
		case tcp.PORT_80:
			m.Push(web.LINK, kit.Format("http://%s", value[tcp.HOST]))
		default:
			m.Push(web.LINK, kit.Format("http://%s:%s", value[tcp.HOST], value[web.PORTAL]))
		}
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
	s.shell(m, m.Template(INSTALL_SH), arg...)
}
func (s relay) Upgrade(m *ice.Message, arg ...string) {
	if len(arg) == 0 && (m.Option(MACHINE) == "" || strings.Contains(m.Option(MACHINE), ",")) {
		m.Options(ice.CMD, m.Template(UPGRADE_SH), cli.DELAY, "0", cli.INTERVAL, "3s")
		s.ForFlow(m)
	} else {
		s.shell(m, m.Template(UPGRADE_SH), arg...)
	}
}
func (s relay) Version(m *ice.Message, arg ...string) {
	m.Options(ice.CMD, m.Template(VERSION_SH), cli.DELAY, "0")
	s.ForFlow(m)
}
func (s relay) Pushbin(m *ice.Message, arg ...string) {
	if len(arg) == 0 {
		p := "ice.linux.amd64"
		switch m.Option(ARCH) {
		case "i686", "386":
			p = "ice.linux.386"
		}
		m.Options(nfs.FROM, ice.USR_PUBLISH+p, nfs.PATH, "contexts/", nfs.FILE, ice.BIN_ICE_BIN)
		m.Cmd(SSH_TRANS, tcp.SEND)
	}
	s.shell(m, m.Template(PUSHBIN_SH), arg...)
}
func (s relay) AdminCmd(m *ice.Message, arg ...string) {
	s.shell(m, kit.JoinWord("contexts/"+ice.BIN_ICE_BIN, web.ADMIN, m.Option(ice.CMD)), arg...)
}

func (s relay) Xterm(m *ice.Message, arg ...string) { s.Code.Xterm(m, m.Option(MACHINE), arg...) }
func (s relay) Repos(m *ice.Message, arg ...string) { s.iframeCmd(m, web.CODE_GIT_STATUS, arg...) }
func (s relay) Vimer(m *ice.Message, arg ...string) { s.iframeCmd(m, web.CODE_VIMER, arg...) }
func (s relay) Admin(m *ice.Message, arg ...string) { s.iframeCmd(m, web.CHAT_PORTAL, arg...) }

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
func (s relay) foreachModify(m *ice.Message, key, cmd string, cb func([]string) string) {
	kit.If(cb == nil, func() { cb = func(ls []string) string { return kit.Join(ls) } })
	s.ForEach(m.Spawn(ice.Maps{MACHINE: "", ice.CMD: cmd, ice.MSG_DAEMON: ""})).Table(func(value ice.Maps) {
		s.Modify(m, MACHINE, value[MACHINE], key, cb(kit.Split(value[ice.RES])), mdb.TIME, time.Now().Format(ice.MOD_TIME))
		m.Push(mdb.TIME, time.Now().Format(ice.MOD_TIME)).Push(MACHINE, value[MACHINE]).Push(tcp.HOST, value[tcp.HOST])
		m.Push(ice.CMD, cmd).Push(ice.RES, value[ice.RES]).PushButton(s.ForFlow, s.ForEach)
	})
}
func (s relay) iframeCmd(m *ice.Message, cmd string, arg ...string) {
	if p := kit.MergeURL2(m.Option(web.LINK), web.CHAT_CMD+cmd); kit.HasPrefix(m.Option(ice.MSG_USERWEB), "https://") {
		s.Code.Iframe(m, m.Option(MACHINE), p, arg...)
	} else {
		m.ProcessOpen(p)
	}
}
