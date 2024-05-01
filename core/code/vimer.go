package code

import (
	"os"
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
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/base/web/html"
	"shylinux.com/x/icebergs/core/chat"
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
	DEMO_C    = "demo.c"
	DEMO_SH   = "demo.sh"
	DEMO_SHY  = "demo.shy"
	DEMO_PY   = "demo.py"
	DEMO_GO   = "demo.go"
	DEMO_JS   = "demo.js"
	DEMO_CSS  = "demo.css"
	DEMO_HTML = "demo.html"
	MAIN_GO   = "main.go"
	MAIN_JS   = "main.js"

	VIMER_SAVE = "vimer.save"
)
const VIMER = "vimer"

func init() {
	web.Index.MergeCommands(ice.Commands{ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) { aaa.White(m, nfs.REQUIRE) }},
		nfs.REQUIRE_SRC: {Hand: func(m *ice.Message, arg ...string) { web.ShareLocalFile(m, ice.SRC, path.Join(arg...)) }},
		nfs.REQUIRE_USR: {Hand: func(m *ice.Message, arg ...string) { web.ShareLocalFile(m, ice.USR, path.Join(arg...)) }},
		nfs.REQUIRE_MODULES: {Hand: func(m *ice.Message, arg ...string) {
			p := path.Join(nfs.USR_MODULES, path.Join(arg...))
			kit.If(!nfs.Exists(m, p), func() {
				if kit.IsIn(m.Option(ice.MSG_USERROLE), aaa.TECH, aaa.ROOT) {
					kit.If(!nfs.Exists(m, nfs.USR_PACKAGE), func() {
						m.Cmd(nfs.SAVE, nfs.USR_PACKAGE, kit.Formats(kit.Dict(mdb.NAME, "usr", nfs.VERSION, "0.0.1")))
					})
					m.Cmd(cli.SYSTEM, "npm", INSTALL, arg[0], kit.Dict(cli.CMD_DIR, ice.USR))
				}
			})
			m.RenderDownload(p)
		}},
	})
	Index.MergeCommands(ice.Commands{
		VIMER: {Name: "vimer path=src/ file=main.go line=1 list", Help: "编辑器", Icon: "vimer.png", Role: aaa.VOID, Meta: kit.Dict(
			ctx.STYLE, INNER, ice.CTX_TRANS, kit.Dict(html.INPUT, kit.Dict(
				cli.MAIN, "程序", "top", "顶域",
			)),
		), Actions: ice.MergeActions(ice.Actions{
			mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
				if mdb.IsSearchPreview(m, arg) {
					kit.For([]string{ice.SRC_MAIN_SHY, ice.SRC_MAIN_SH, ice.SRC_MAIN_GO, ice.SRC_MAIN_JS}, func(p string) {
						m.PushSearch(mdb.TYPE, nfs.FILE, mdb.NAME, path.Base(p), mdb.TEXT, p)
					})
				}
			}},
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch m.Option(ctx.ACTION) {
				case nfs.MODULE:
					m.Cmdy(AUTOGEN, mdb.INPUTS, arg)
				case nfs.SCRIPT, mdb.CREATE:
					if strings.HasSuffix(m.Option(nfs.FILE), nfs.PS) {
						m.Option(nfs.FILE, path.Join(m.Option(nfs.FILE), path.Base(strings.TrimSuffix(m.Option(nfs.FILE), nfs.PS)+".go")))
					}
					kit.For([]string{JS, CSS, SHY, "json"}, func(ext string) {
						m.Push(nfs.PATH, kit.ExtChange(m.Option(nfs.FILE), ext))
					})
					m.Push(nfs.PATH, path.Join(path.Dir(m.Option(nfs.FILE)), "trans.json"))
					m.Option(nfs.DIR_REG, kit.ExtReg(SH, SHY, PY, JS, CSS, HTML))
					nfs.DirDeepAll(m, nfs.SRC, nfs.PWD, nil, nfs.PATH)
				case web.DREAM, XTERM, AUTOGEN:
					m.Cmdy(m.Option(ctx.ACTION), mdb.INPUTS, arg)
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
							case tcp.GATEWAY:
								push(web.SPACE, value[mdb.TEXT])
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
							case cli.OPENS:
								push(cli.OPENS, value[mdb.TEXT])
							}
						})
						for _, p := range kit.Split(kit.Select(m.Option(nfs.PATH), m.Option("paths"))) {
							nfs.DirDeepAll(m.Spawn(), nfs.PWD, p, func(value ice.Maps) { push("", value[nfs.PATH]) }, nfs.PATH)
						}
						m.Cmd(ctx.COMMAND).Table(func(value ice.Maps) { push(ctx.INDEX, value[ctx.INDEX]) })
					default:
						m.Cmdy(INNER, mdb.INPUTS, arg)
					}
				}
			}},
			nfs.SAVE: {Hand: func(m *ice.Message, arg ...string) {
				kit.If(m.Option(nfs.CONTENT) == "", func() { m.Option(nfs.CONTENT, m.Cmdx("", TEMPLATE)) })
				m.Cmd(nfs.SAVE, path.Join(m.Option(nfs.PATH), m.Option(nfs.FILE)))
				gdb.Event(m, VIMER_SAVE)
			}},
			nfs.TRASH: {Hand: func(m *ice.Message, arg ...string) { nfs.Trash(m, arg[0]) }},
			nfs.MODULE: {Name: "module name*=hi help=示例 type*=Hash,Zone,Data,Code main*=main.go zone=hi top=web.chat", Help: "模块", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(AUTOGEN, nfs.MODULE, arg)
			}},
			nfs.SCRIPT: {Name: "script file*", Help: "脚本", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(nfs.DEFS, path.Join(m.Option(nfs.PATH), m.Option(nfs.FILE)), m.Cmdx("", TEMPLATE))
			}},
			cli.OPENS: {Hand: func(m *ice.Message, arg ...string) { cli.Opens(m, arg...) }},
			cli.MAKE: {Hand: func(m *ice.Message, arg ...string) {
				defer web.ToastProcess(m)()
				web.PushStream(m)
				m.Cmd(cli.SYSTEM, "echo")
				m.Cmd(cli.SYSTEM, "date")
				m.Cmd(cli.SYSTEM, cli.MAKE, m.Option(nfs.TARGET), kit.Dict(cli.CMD_DIR, m.Option(nfs.PATH)))
			}},
			ice.APP: {Help: "本机", Hand: func(m *ice.Message, arg ...string) {
				cli.OpenCmds(m, "cd "+kit.Path(""), "vim "+path.Join(arg[0], arg[1])+" +"+arg[2]).ProcessHold()
			}},
			COMPILE: {Help: "编译", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(nfs.PATH) != "" && nfs.ExistsFile(m, path.Join(m.Option(nfs.PATH), ice.MAKEFILE)) {
					web.PushStream(m).Cmdy(cli.SYSTEM, cli.MAKE, kit.Dict(cli.CMD_DIR, m.Option(nfs.PATH)))
					return
				} else if m.Option(nfs.PATH) == ice.USR_VOLCANOS && strings.HasPrefix(m.Option(nfs.FILE), "publish/client/mp/") {
					web.PushStream(m).Cmdy(cli.SYSTEM, cli.MAKE, kit.Dict(cli.CMD_DIR, path.Join(m.Option(nfs.PATH), "publish/client/mp/")))
					return
				}
				const app, _app = "usr/publish/Contexts.app", "Contents/MacOS/Contexts"
				isWebview := func() bool { return strings.HasSuffix(os.Args[0], _app) }
				cmds := []string{COMPILE, ice.SRC_MAIN_GO, ice.BIN_ICE_BIN}
				if isWebview() {
					m.Option(cli.ENV, "CGO_ENABLED", "1", cli.HOME, kit.Env(cli.HOME), cli.PATH, kit.Path(ice.USR_LOCAL_GO_BIN)+nfs.DF+kit.Env(cli.PATH))
					cmds = []string{COMPILE, ice.SRC_WEBVIEW_GO, path.Join(app, _app)}
				}
				if msg := m.Cmd(cmds); cli.IsSuccess(msg) {
					if m.GoSleep30ms(func() { m.Cmd(UPGRADE, cli.RESTART) }); isWebview() {
						m.Cmd(cli.DAEMON, ice.BIN_ICE_BIN, cli.FOREVER, cli.DELAY, cli.TIME_300ms, cli.SYSTEM, cli.OPEN, app)
					}
				} else {
					_vimer_make(m, nfs.PWD, msg)
				}
			}},
			TEMPLATE: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(TEMPLATE, kit.Ext(m.Option(mdb.FILE)), m.Option(nfs.FILE), m.Option(nfs.PATH))
			}},
			COMPLETE: {Hand: func(m *ice.Message, arg ...string) {
				return
				m.Cmdy(COMPLETE, kit.Ext(m.Option(mdb.FILE)), m.Option(nfs.FILE), m.Option(nfs.PATH))
			}},
			chat.FAVOR_INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case mdb.TYPE:
					m.Push(arg[0], nfs.FILE)
				case mdb.TEXT:
					kit.If(m.Option(mdb.TYPE) == nfs.FILE, func() { m.Push(arg[0], ice.SRC_MAIN_SHY, ice.SRC_MAIN_GO, ice.SRC_MAIN_JS) })
				}
			}},
			chat.FAVOR_TABLES: {Hand: func(m *ice.Message, arg ...string) {
				kit.If(m.Option(mdb.TYPE) == nfs.FILE, func() { m.PushButton(kit.Dict(m.CommandKey(), "源码")) })
			}},
			chat.FAVOR_ACTION: {Hand: func(m *ice.Message, arg ...string) {
				kit.If(m.Option(mdb.TYPE) == nfs.FILE, func() { ctx.ProcessField(m, m.ShortKey(), nfs.SplitPath(m, m.Option(mdb.TEXT))) })
			}},
			web.DREAM_TABLES: {Hand: func(m *ice.Message, _ ...string) {
				kit.If(m.IsDebug(), func() { m.PushButton(kit.Dict(m.CommandKey(), "编程")) })
			}},
		}, chat.FavorAction(), web.DreamTablesAction("编程"), ctx.ConfAction(ctx.TOOLS, "xterm,compile,runtime", web.ONLINE, ice.TRUE)), Hand: func(m *ice.Message, arg ...string) {
			if m.Cmdy(INNER, arg); arg[0] != ctx.ACTION {
				if web.IsLocalHost(m) {
					m.Action(nfs.SAVE, COMPILE, mdb.SHOW, ice.APP)
				} else if m.IsMobileUA() {
					m.Action(nfs.SAVE, COMPILE)
				} else {
					m.Action(nfs.SAVE, COMPILE, mdb.SHOW)
				}
				ctx.DisplayLocal(m, "")
				ctx.Toolkit(m)
			}
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
