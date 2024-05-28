package code

import (
	"encoding/base64"
	"path"
	"runtime"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/base/web/html"
	"shylinux.com/x/icebergs/misc/xterm"
	kit "shylinux.com/x/toolkits"
)

func _xterm_get(m *ice.Message, h string) xterm.XTerm {
	h = kit.Select(m.Option(mdb.HASH), h)
	m.Assert(h != "")
	mdb.HashModify(m, mdb.TIME, m.Time(), cli.DAEMON, m.Option(ice.MSG_DAEMON))
	return mdb.HashSelectTarget(m, h, func(value ice.Maps) ice.Any {
		text := strings.Split(value[mdb.TEXT], lex.NL)
		ls := kit.Split(strings.Split(kit.Select(ISH, value[mdb.TYPE]), " # ")[0])
		kit.If(ls[0] == cli.SH, func() { ls[0] = cli.Shell(m) })
		kit.If(value[nfs.PATH] != "" && !strings.HasSuffix(value[nfs.PATH], nfs.PS), func() { value[nfs.PATH] = path.Dir(value[nfs.PATH]) })
		term, e := xterm.Command(m, value[nfs.PATH], kit.Select(ls[0], cli.SystemFind(m, ls[0], value[nfs.PATH])), ls[1:]...)
		if m.WarnNotValid(e) {
			return nil
		}
		m.Go(func() {
			m.Log(cli.START, kit.JoinCmdArgs(ls...))
			defer mdb.HashRemove(m, mdb.HASH, h)
			buf := make([]byte, ice.MOD_BUFS)
			for {
				if n, e := term.Read(buf); !m.WarnNotValid(e) && e == nil {
					if _xterm_echo(m, h, string(buf[:n])); len(text) > 0 {
						kit.If(text[0], func(cmd string) { m.GoSleep300ms(func() { term.Write([]byte(cmd + lex.NL)) }) })
						text = text[1:]
					}
				} else {
					_xterm_echo(m, h, "~~~end~~~")
					break
				}
			}
		})
		return term
	}).(xterm.XTerm)
}
func _xterm_echo(m *ice.Message, h string, str string) {
	m.Options(ice.MSG_DAEMON, mdb.HashSelectField(m, h, cli.DAEMON), ice.MSG_COUNT, "0")
	m.Options(ice.LOG_DISABLE, ice.TRUE)
	web.PushNoticeGrow(m, h, str)
}
func _xterm_cmds(m *ice.Message, h string, cmd string, arg ...ice.Any) {
	kit.If(cmd != "", func() { _xterm_get(m, h).Write([]byte(kit.Format(cmd, arg...) + lex.NL)) })
	m.ProcessHold()
}

const (
	SHELL = "shell"
)
const XTERM = "xterm"

func init() {
	Index.MergeCommands(ice.Commands{
		XTERM: {Name: "xterm refresh", Help: "终端", Icon: "Terminal.png", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				web.AddPortalProduct(m, "命令行", `
一款网页版的命令行，打开网页即可随时随地的敲命令，
无论这些命令是运行在本机，还是远程，还是任何虚拟的空间，无论是内存还是磁盘。
`, 500.0)
				kit.For([]string{
					"xterm/lib/xterm.js",
					"xterm/css/xterm.css",
					"xterm-addon-fit/lib/xterm-addon-fit.js",
					"xterm-addon-web-links/lib/xterm-addon-web-links.js",
				}, func(p string) {
					m.Cmd(WEBPACK, mdb.INSERT, p)
					m.Cmd(BINPACK, mdb.INSERT, nfs.USR_MODULES+p)
				})
			}},
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch mdb.HashInputs(m, arg); arg[0] {
				case mdb.HASH:
					fallthrough
				case mdb.TYPE:
					m.Push(arg[0], cli.Shell(m))
					m.Cmd(mdb.SEARCH, SHELL, "", "", func(value ice.Maps) {
						kit.If(value[mdb.TYPE] == SHELL, func() { m.Push(arg[0], value[mdb.TEXT]) })
					})
				case mdb.NAME:
					m.Push(arg[0], path.Base(m.Option(mdb.TYPE)), ice.Info.Hostname)
				case nfs.PATH:
					m.Cmdy(nfs.DIR, ice.USR_LOCAL_WORK, nfs.PATH)
					m.Cmdy(nfs.DIR, ice.USR_LOCAL_REPOS, nfs.PATH)
					m.Cmdy(nfs.DIR, ice.USR_LOCAL_DAEMON, nfs.PATH)
				}
			}},
			mdb.CREATE: {Hand: func(m *ice.Message, arg ...string) {
				m.ProcessRewrite(mdb.HASH, mdb.HashCreate(m))
			}},
			html.RESIZE: {Hand: func(m *ice.Message, arg ...string) {
				_xterm_get(m, "").Setsize(m.OptionDefault("rows", "24"), m.OptionDefault("cols", "80"))
			}},
			html.INPUT: {Hand: func(m *ice.Message, arg ...string) {
				if b, e := base64.StdEncoding.DecodeString(strings.Join(arg, "")); !m.WarnNotValid(e) {
					_xterm_get(m, "").Write(b)
				}
			}},
			ice.APP: {Help: "本机", Icon: "bi bi-terminal", Hand: func(m *ice.Message, arg ...string) {
				if h := kit.Select(m.Option(mdb.HASH), arg, 0); h == "" {
					cli.OpenCmds(m, "cd "+kit.Path(""))
				} else {
					cli.OpenCmds(m, "cd "+kit.Path("")+"; "+m.Cmdv("", h, mdb.TYPE))
				}
				m.ProcessHold()
			}},
			web.DREAM_TABLES: {Hand: func(m *ice.Message, arg ...string) {
				kit.If(m.IsDebug() && aaa.IsTechOrRoot(m), func() {
					// m.PushButton(cli.RUNTIME)
					m.PushButton(cli.RUNTIME, kit.Dict(m.CommandKey(), m.Commands("").Help))
				})
				// kit.If(m.IsDebug(), func() { list = append(list, cli.RUNTIME) })
			}},
			web.DREAM_ACTION: {Hand: func(m *ice.Message, arg ...string) { web.DreamProcess(m, "", cli.SH, arg...) }},
		}, web.DreamTablesAction(), mdb.HashAction(mdb.FIELD, "time,hash,type,name,text,path,daemon")), Hand: func(m *ice.Message, arg ...string) {
			if mdb.HashSelect(m, arg...); len(arg) == 0 {
				list := m.CmdMap(web.SPACE, mdb.NAME)
				m.Table(func(value ice.Maps) {
					if list[value[cli.DAEMON]] == nil {
						m.Push(mdb.STATUS, web.OFFLINE)
					} else {
						m.Push(mdb.STATUS, web.ONLINE)
					}
				})
			} else {
				kit.If(m.Length() == 0, func() {
					kit.If(arg[0] == cli.SH, func() {
						if runtime.GOOS == cli.WINDOWS {
							arg[0] = "ish"
						} else {
							arg[0] = cli.Shell(m)
						}
					})
					arg[0] = m.Cmdx("", mdb.CREATE, arg)
					mdb.HashSelect(m, arg[0])
				})
				m.Push(mdb.HASH, arg[0])
				kit.If(m.IsLocalhost() && m.Append(mdb.TYPE) != "ish", func() { m.Action(ice.APP) })
			}
			ctx.DisplayLocal(m, "")
		}},
	})
}

func ProcessXterm(m *ice.Message, cmds, text string, arg ...string) {
	ctx.ProcessField(m, XTERM, func() []string {
		if ls := kit.Simple(kit.UnMarshal(m.Option(ctx.ARGS))); len(ls) > 0 {
			return ls
		} else {
			return []string{cmds, kit.Select("", arg, 0), text}
		}
	}, arg...)
}
