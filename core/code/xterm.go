package code

import (
	"encoding/base64"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/ssh"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/base/web/html"
	"shylinux.com/x/icebergs/core/chat"
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
		kit.If(value[nfs.PATH] != "" && !strings.HasSuffix(value[nfs.PATH], nfs.PS), func() { value[nfs.PATH] = path.Dir(value[nfs.PATH]) })
		term, e := xterm.Command(m, value[nfs.PATH], kit.Select(ls[0], cli.SystemFind(m, ls[0], value[nfs.PATH])), ls[1:]...)
		if m.WarnNotValid(e) {
			return nil
		}
		m.Go(func() {
			m.Log(cli.START, strings.Join(ls, lex.SP))
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
	// m.Options(ice.LOG_DISABLE, ice.TRUE)
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
	shell := kit.Env("SHELL", "/bin/sh")
	Index.MergeCommands(ice.Commands{
		XTERM: {Name: "xterm hash auto", Help: "终端", Icon: "Terminal.png", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				kit.If(m.Cmd("").Length() == 0, func() {
					m.Cmd("", mdb.CREATE, mdb.TYPE, shell)
					m.Cmd("", mdb.CREATE, mdb.TYPE, "/bin/ish")
				})
			}},
			mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
				if mdb.IsSearchPreview(m, arg) || kit.HasPrefixList(arg, SHELL) {
					kit.For([]string{shell, "/bin/ish"}, func(p string) {
						m.PushSearch(mdb.TYPE, SHELL, mdb.NAME, path.Base(p), mdb.TEXT, p)
					})
				}
			}},
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch mdb.HashInputs(m, arg); arg[0] {
				case mdb.HASH:
					fallthrough
				case mdb.TYPE:
					m.Cmd(mdb.SEARCH, SHELL, "", "", func(value ice.Maps) {
						kit.If(value[mdb.TYPE] == SHELL, func() { m.Push(arg[0], value[mdb.TEXT]) })
					})
				case mdb.NAME:
					m.Push(arg[0], path.Base(m.Option(mdb.TYPE)), ice.Info.Hostname)
				case nfs.PATH:
					m.Cmdy(nfs.DIR, ice.USR_LOCAL_WORK, nfs.PATH)
					m.Cmdy(nfs.DIR, ice.USR_LOCAL_REPOS, nfs.PATH)
					m.Cmdy(nfs.DIR, ice.USR_LOCAL_DAEMON, nfs.PATH)
				case nfs.FILE:
					push := func(arg ...string) { m.Push(nfs.FILE, strings.Join(arg, nfs.DF)) }
					m.Cmd("", func(value ice.Maps) {
						kit.If(value[mdb.TYPE] == html.LAYOUT, func() { push(html.LAYOUT, value[mdb.HASH], value[mdb.NAME]) })
					})
					m.Cmd("", mdb.INPUTS, mdb.TYPE, func(value ice.Maps) { push(ssh.SHELL, value[mdb.TYPE]) })
					m.Cmd(nfs.CAT, kit.HomePath(".bash_history"), func(text string) { push(text) })
					m.Cmd(nfs.CAT, kit.HomePath(".zsh_history"), func(text string) { push(text) })
					m.Cmd(ctx.COMMAND, mdb.SEARCH, ctx.COMMAND, "", "", ice.OptionFields(ctx.INDEX), func(value ice.Maps) { push(ctx.INDEX, value[ctx.INDEX]) })
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
			html.OUTPUT: {Help: "全屏", Hand: func(m *ice.Message, arg ...string) {
				m.ProcessOpen(m.MergePodCmd("", "", mdb.HASH, kit.Select(m.Option(mdb.HASH), arg, 0), ctx.STYLE, html.OUTPUT))
			}},
			INSTALL: {Help: "安装", Hand: func(m *ice.Message, arg ...string) {
				_xterm_get(m, kit.Select("", arg, 0)).Write([]byte(m.Cmdx(PUBLISH, ice.CONTEXTS, ice.APP, kit.Dict("format", "raw")) + ice.NL))
				m.ProcessHold()
			}},
			ice.APP: {Help: "本机", Icon: "bi bi-terminal", Hand: func(m *ice.Message, arg ...string) {
				if h := kit.Select(m.Option(mdb.HASH), arg, 0); h == "" {
					cli.OpenCmds(m, "cd "+kit.Path(""))
				} else {
					cli.OpenCmds(m, "cd "+kit.Path("")+"; "+m.Cmdv("", h, mdb.TYPE))
				}
				m.ProcessHold()
			}},
			chat.FAVOR_INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case mdb.TYPE:
					m.Push(arg[0], SHELL)
				case mdb.TEXT:
					if m.Option(mdb.TYPE) == SHELL {
						m.Cmd(mdb.SEARCH, mdb.FOREACH, "", "", func(value ice.Maps) {
							kit.If(value[mdb.TYPE] == ssh.SHELL, func() { m.Push(arg[0], value[mdb.TEXT]) })
						})
					}
				}
			}},
			chat.FAVOR_TABLES: {Hand: func(m *ice.Message, arg ...string) {
				kit.If(m.Option(mdb.TYPE) == SHELL, func() { m.PushButton(kit.Dict(m.CommandKey(), "终端")) })
			}},
			chat.FAVOR_ACTION: {Hand: func(m *ice.Message, arg ...string) {
				kit.If(m.Option(mdb.TYPE) == SHELL, func() {
					ctx.ProcessField(m, m.ShortKey(), m.Cmdx("", mdb.CREATE, mdb.TYPE, m.Option(mdb.TEXT), mdb.NAME, m.Option(mdb.NAME), mdb.TEXT, ""))
				})
			}},
			web.DREAM_TABLES: {Hand: func(m *ice.Message, arg ...string) {
				kit.If(aaa.IsTechOrRoot(m), func() { m.PushButton(kit.Dict(m.CommandKey(), m.Commands("").Help)) })
			}},
			web.DREAM_ACTION: {Hand: func(m *ice.Message, arg ...string) { web.DreamProcess(m, "", cli.Shell(m), arg...) }},
		}, web.DreamTablesAction(), chat.FavorAction(), mdb.HashAction(mdb.FIELD, "time,hash,type,name,text,path")), Hand: func(m *ice.Message, arg ...string) {
			if mdb.HashSelect(m, arg...); len(arg) == 0 {
				if web.IsLocalHost(m) {
					m.Action(mdb.CREATE, mdb.PRUNES, ice.APP)
				} else {
					m.Action(mdb.CREATE, mdb.PRUNES)
				}
			} else {
				kit.If(m.Length() == 0, func() {
					if arg[0] == SH {
						arg[0] = cli.Shell(m)
					}
					arg[0] = m.Cmdx("", mdb.CREATE, arg)
					mdb.HashSelect(m, arg[0])
				})
				m.Push(mdb.HASH, arg[0]).Action(ice.APP)
				ctx.DisplayLocal(m, "")
			}
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
