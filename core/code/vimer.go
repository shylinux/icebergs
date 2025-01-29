package code

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/ssh"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _vimer_make(m *ice.Message, dir string, msg *ice.Message) {
	for _, line := range strings.Split(msg.Append(cli.CMD_ERR), lex.NL) {
		if !strings.Contains(line, nfs.DF) {
			continue
		} else if ls := strings.SplitN(line, nfs.DF, 4); len(ls) > 3 {
			for i, p := range kit.Split(dir) {
				if strings.HasPrefix(ls[0], p) {
					m.Push(nfs.PATH, p).Push(nfs.FILE, strings.TrimPrefix(ls[0], p)).Push(nfs.LINE, ls[1]).Push(mdb.TEXT, ls[3])
					break
				} else if i == strings.Count(dir, mdb.FS) {
					ps := nfs.SplitPath(m, ls[0])
					m.Push(nfs.PATH, ps[0]).Push(nfs.FILE, ps[1]).Push(nfs.LINE, ls[1]).Push(mdb.TEXT, ls[3])
				}
			}
		}
	}
	m.Echo(msg.Append(cli.CMD_OUT)).Echo(msg.Append(cli.CMD_ERR))
}

const (
	MAIN_GO   = "main.go"
	MAIN_JS   = "main.js"
	DEMO_C    = "demo.c"
	DEMO_SH   = "demo.sh"
	DEMO_SHY  = "demo.shy"
	DEMO_PY   = "demo.py"
	DEMO_GO   = "demo.go"
	DEMO_JS   = "demo.js"
	DEMO_CSS  = "demo.css"
	DEMO_HTML = "demo.html"

	VIMER_SAVE = "vimer.save"
)
const VIMER = "vimer"

func init() {
	Index.MergeCommands(ice.Commands{
		VIMER: {Name: "vimer path=src/ file=main.go line=1 list", Help: "编辑器", Icon: "vimer.png", Role: aaa.VOID, Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				if ice.Info.CodeMain != "" {
					ls := nfs.SplitPath(m, ice.Info.CodeMain)
					kit.Value(m.Command().List, "0.value", ls[0])
					kit.Value(m.Command().List, "1.value", ls[1])
				}
			}},
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch m.Option(ctx.ACTION) {
				case nfs.MODULE:
					m.Cmdy(AUTOGEN, mdb.INPUTS, arg)
				case nfs.SCRIPT, mdb.CREATE, mdb.RENAME:
					if strings.HasSuffix(m.Option(nfs.FILE), nfs.PS) {
						m.Option(nfs.FILE, path.Join(m.Option(nfs.FILE), path.Base(strings.TrimSuffix(m.Option(nfs.FILE), nfs.PS)+".go")))
					}
					kit.For([]string{JS, CSS, SHY, "json"}, func(ext string) {
						m.Push(nfs.PATH, kit.ExtChange(m.Option(nfs.FILE), ext))
					})
					m.Push(nfs.PATH, path.Join(path.Dir(m.Option(nfs.FILE)), "trans.json"))
					m.Option(nfs.DIR_REG, kit.ExtReg(SH, SHY, PY, JS, CSS, HTML))
					nfs.DirDeepAll(m, nfs.SRC, nfs.PWD, nil, nfs.PATH)
				default:
					switch arg[0] {
					case ice.CMD:
						m.OptionFields(ctx.INDEX)
						m.Cmd(ctx.COMMAND, mdb.SEARCH, ctx.COMMAND).Table(func(value ice.Maps) {
							kit.If(strings.HasPrefix(value[ctx.INDEX], kit.Select("", arg, 1)), func() { m.Push(arg[0], strings.TrimPrefix(value[ctx.INDEX], arg[1]+".")) })
						})
					case ctx.INDEX:
						m.Cmd(ctx.COMMAND).Table(func(value ice.Maps) {
							kit.If(strings.HasPrefix(value[ctx.INDEX], kit.Select("", arg, 1)), func() { m.Push(arg[0], value[ctx.INDEX]) })
						})
					case nfs.PATH:
						p := kit.Select(nfs.PWD, arg, 1)
						m.Cmdy(nfs.DIR, p, nfs.DIR_CLI_FIELDS)
						var bind = []string{"usr/icebergs/core/", "usr/volcanos/plugin/local/"}
						kit.If(strings.HasPrefix(p, bind[0]), func() { m.Cmdy(nfs.DIR, strings.Replace(p, bind[0], bind[1], 1), nfs.DIR_CLI_FIELDS) })
						kit.If(strings.HasPrefix(p, bind[1]), func() { m.Cmdy(nfs.DIR, strings.Replace(p, bind[1], bind[0], 1), nfs.DIR_CLI_FIELDS) })
					case nfs.FILE:
						list := ice.Map{}
						push := func(k, p string) {
							kit.IfNoKey(list, kit.Select(k, k+nfs.DF, k != "")+p, func(p string) { m.Push(nfs.PATH, p) })
						}
						m.Cmd(mdb.SEARCH, mdb.FOREACH, "", "").Sort("").Table(func(value ice.Maps) {
							switch value[mdb.TYPE] {
							case nfs.FILE:
								push("", value[mdb.TEXT])
							case web.LINK:
								push(web.SPACE, value[mdb.TEXT])
							case web.WORKER:
								push(web.SPACE, value[mdb.NAME])
							case web.SERVER:
								push(web.SPACE, value[mdb.TEXT])
							case ctx.INDEX:
								push(ctx.INDEX, value[mdb.TEXT])
							case ssh.SHELL:
								push(ssh.SHELL, value[mdb.TEXT])
							}
						})
						for _, p := range kit.Split(kit.Select(m.Option(nfs.PATH), m.Option("paths"))) {
							nfs.DirDeepAll(m.Spawn(), nfs.PWD, p, func(value ice.Maps) { push("", value[nfs.PATH]) }, nfs.PATH)
						}
						m.Cmd(ctx.COMMAND).Table(func(value ice.Maps) { push(ctx.INDEX, value[ctx.INDEX]) })
					}
				}
			}},
			nfs.MODULE: {Name: "module name*=hi help type*=Hash,Zone,Code main*=main.go zone*=hi top", Help: "创建模块", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(AUTOGEN, nfs.MODULE, arg)
			}},
			nfs.SCRIPT: {Name: "script file*", Help: "脚本", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(nfs.DEFS, path.Join(m.Option(nfs.PATH), m.Option(nfs.FILE)), m.Cmdx("", TEMPLATE))
			}},
			mdb.CREATE: {Name: "create file*", Help: "添加文件", Icon: "bi bi-file-earmark-text", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(nfs.DEFS, path.Join(m.Option(nfs.PATH), m.Option(nfs.FILE)), m.Cmdx("", TEMPLATE))
			}},
			mdb.RENAME: {Name: "rename to*", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(nfs.MOVE, path.Join(m.Option(nfs.PATH), m.Option(nfs.TO)), path.Join(m.Option(nfs.PATH), m.Option(nfs.FILE)))
			}},
			nfs.TRASH: {Hand: func(m *ice.Message, arg ...string) {
				nfs.Trash(m, arg[0])
			}},
			nfs.SAVE: {Hand: func(m *ice.Message, arg ...string) {
				kit.If(m.Option(nfs.CONTENT) == "", func() { m.Option(nfs.CONTENT, m.Cmdx("", TEMPLATE)) })
				m.Cmd(nfs.SAVE, path.Join(m.Option(nfs.PATH), m.Option(nfs.FILE)))
				gdb.Event(m, VIMER_SAVE)
			}},
			TEMPLATE: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(TEMPLATE, kit.Ext(m.Option(mdb.FILE)), m.Option(nfs.FILE), m.Option(nfs.PATH))
			}},
			COMPLETE: {Hand: func(m *ice.Message, arg ...string) {
				return
				m.Cmdy(COMPLETE, kit.Ext(m.Option(mdb.FILE)), m.Option(nfs.FILE), m.Option(nfs.PATH))
			}},
			COMPILE: {Hand: func(m *ice.Message, arg ...string) {
				if m.Option(nfs.PATH) == ice.USR_PROGRAM {
					m.Cmdy("web.chat.wx.ide", "make")
					return
				}
				if msg := m.Cmd(COMPILE, ice.SRC_MAIN_GO, ice.BIN_ICE_BIN); cli.IsSuccess(msg) {
					m.GoSleep30ms(func() { m.Cmd(UPGRADE, cli.RESTART) })
				} else {
					_vimer_make(m, nfs.PWD, msg)
				}
			}},
			REPOS: {Role: aaa.VOID, Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(REPOS, ice.OptionFields(nfs.PATH)).Sort(nfs.PATH)
			}},
			ice.APP: {Help: "本机", Hand: func(m *ice.Message, arg ...string) {
				kit.If(len(arg) == 0, func() { arg = append(arg, m.Option(nfs.PATH), m.Option(nfs.FILE), m.Option(nfs.LINE)) })
				cli.OpenCmds(m, "cd "+kit.Path(""), "vim "+path.Join(arg[0], arg[1])+" +"+arg[2]).ProcessHold()
			}},
			web.DREAM_TABLES: {Hand: func(m *ice.Message, arg ...string) {
				kit.If(m.IsDebug() && aaa.IsTechOrRoot(m), func() {
					m.PushButton(kit.Dict(m.CommandKey(), "编程"))
				})
			}},
		}, web.DreamTablesAction("编程"), ctx.ConfAction(ctx.TOOLS, "xterm,runtime,compile", web.ONLINE, ice.TRUE)), Hand: func(m *ice.Message, arg ...string) {
			if m.Cmdy(INNER, arg); arg[0] == ctx.ACTION {
				return
			} else if len(arg) == 1 {
				m.PushAction(mdb.CREATE, mdb.RENAME, nfs.TRASH).Action(nfs.MODULE)
				return
			}
			if m.IsMobileUA() {
				m.Action(nfs.SAVE, COMPILE)
			} else if web.IsLocalHost(m) {
				m.Action(nfs.SAVE, COMPILE, mdb.SHOW, ice.APP)
			} else {
				m.Action(nfs.SAVE, COMPILE, mdb.SHOW)
			}
			ctx.DisplayLocal(m, "")
			ctx.Toolkit(m)
		}},
	})
}

const TEMPLATE = "template"
const COMPLETE = "complete"
const NAVIGATE = "navigate"

func init() {
	Index.MergeCommands(ice.Commands{
		TEMPLATE: {Name: "template type name text auto", Help: "模板", Actions: mdb.RenderAction()},
		COMPLETE: {Name: "complete type name text auto", Help: "补全", Actions: mdb.RenderAction()},
		NAVIGATE: {Name: "navigate type name text auto", Help: "跳转", Actions: mdb.RenderAction()},
	})
	ice.AddMergeAction(func(c *ice.Context, key string, cmd *ice.Command, sub string, action *ice.Action) ice.Handler {
		switch sub {
		case TEMPLATE, COMPLETE, NAVIGATE:
			return func(m *ice.Message, arg ...string) { m.Cmd(sub, mdb.CREATE, key, m.ShortKey()) }
		default:
			return nil
		}
	})
}
func Complete(m *ice.Message, text string, data ice.Map) {
	if strings.HasSuffix(text, nfs.PT) {
		m.Push(mdb.TEXT, kit.Simple(data[kit.Slice(kit.Split(text, " ."), -1)[0]]))
	} else {
		m.Push(mdb.TEXT, data[""])
		kit.For(data, func(k string, v ice.Any) { m.Push(mdb.TEXT, k) })
	}
}
