package code

import (
	"os"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _vimer_make(m *ice.Message, dir string, msg *ice.Message) {
	defer m.StatusTime()
	for _, line := range strings.Split(msg.Append(cli.CMD_ERR), ice.NL) {
		if !strings.Contains(line, ice.DF) {
			continue
		}
		if ls := strings.SplitN(line, ice.DF, 4); len(ls) > 3 {
			for i, p := range kit.Split(dir) {
				if strings.HasPrefix(ls[0], p) {
					m.Push(nfs.PATH, p)
					m.Push(nfs.FILE, strings.TrimPrefix(ls[0], p))
					m.Push(nfs.LINE, ls[1])
					m.Push(mdb.TEXT, ls[3])
					break
				} else if n := 2; i == strings.Count(dir, ice.FS) {
					if strings.HasPrefix(ls[0], "src/") {
						n = 1
					}
					m.Push(nfs.PATH, kit.Join(kit.Slice(kit.Split(ls[0], ice.PS, ice.PS), 0, n), ice.PS)+ice.PS)
					m.Push(nfs.FILE, kit.Join(kit.Slice(kit.Split(ls[0], ice.PS, ice.PS), n), ice.PS))
					m.Push(nfs.LINE, ls[1])
					m.Push(mdb.TEXT, ls[3])
				}
			}
		}
	}
	if m.Length() == 0 {
		m.Echo(msg.Append(cli.CMD_OUT)).Echo(msg.Append(cli.CMD_ERR))
	}
}

const VIMER = "vimer"

func init() {
	Index.MergeCommands(ice.Commands{
		VIMER: {Name: "vimer path=src/ file=main.go line=1 list", Help: "编辑器", Actions: ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch m.Option(ctx.ACTION) {
				case AUTOGEN, web.DREAM, XTERM:
					m.Cmdy(m.Option(ctx.ACTION), mdb.INPUTS, arg)
				case nfs.SCRIPT, web.WEBSITE:
					m.Cmdy(COMPLETE, mdb.FOREACH, kit.Select("", arg, 1), m.Option(ctx.ACTION))
				case "extension":
					nfs.DirDeepAll(m, "usr/volcanos/plugin/local/code/", "inner/", nil, nfs.PATH)
				default:
					switch arg[0] {
					case ctx.INDEX:
						m.Cmdy(ctx.COMMAND, mdb.SEARCH, ctx.COMMAND, ice.OptionFields(ctx.INDEX))
					default:
						p := m.Option(nfs.PATH)
						list := ice.Map{}
						m.Cmd(FAVOR, "_recent_file").Tables(func(value ice.Maps) {
							if p := value[nfs.PATH]+value[nfs.FILE]; list[p] == nil {
								m.Push(nfs.PATH, p)
								list[p] = value
							}
						})
						m.Cmd(mdb.RENDER, kit.Ext(m.Option(nfs.FILE)), m.Option(nfs.FILE), p).Tables(func(value ice.Maps) {
							m.Push(nfs.PATH, kit.Format("line:%s:%s:%s", value[nfs.LINE], value["kind"], value[mdb.NAME]))
						})
						for _, p := range kit.Split(kit.Select(m.Option(nfs.PATH), m.Option("paths"))) {
							nfs.DirDeepAll(m, nfs.PWD, p, nil, nfs.PATH)
						}
						m.Cmd(ctx.COMMAND, mdb.SEARCH, ctx.COMMAND, ice.OptionFields(ctx.INDEX)).Tables(func(value ice.Maps) {
							m.Push(nfs.PATH, "index:"+value[ctx.INDEX])
						})
						m.Cmd(FAVOR, "_system_app").Tables(func(value ice.Maps) {
							m.Push(nfs.PATH, "_open:"+strings.ToLower(value[mdb.NAME]))
						})
					}
				}
			}},
			nfs.SAVE: {Name: "save", Help: "保存", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(nfs.CONTENT) == "" {
					m.Option(nfs.CONTENT, m.Cmdx("", TEMPLATE))
				}
				m.Cmdy(nfs.SAVE, path.Join(m.Option(nfs.PATH), m.Option(nfs.FILE)))
			}},
			nfs.TRASH: {Name: "trash path", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(nfs.TRASH, arg[0])
			}},
			nfs.SCRIPT: {Name: "script file=hi/hi.js", Help: "脚本", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(nfs.DEFS, path.Join(m.Option(nfs.PATH), m.Option(nfs.FILE)), m.Cmdx("", TEMPLATE))
			}},
			web.WEBSITE: {Name: "website file=hi.zml", Help: "网页", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(nfs.DEFS, path.Join(m.Option(nfs.PATH), m.Option(nfs.FILE)), m.Cmdx("", TEMPLATE))
			}},
			web.DREAM: {Name: "dream name=hi repos", Help: "空间", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(web.DREAM, cli.START, arg)
			}},
			XTERM: {Name: "xterm type=sh name text", Help: "终端", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(XTERM, mdb.CREATE, arg)
			}},
			FAVOR: {Name: "favor", Help: "收藏"},

			"keyboard": {Name: "keyboard", Help: "远程控制", Hand: func(m *ice.Message, arg ...string) {
				hash := m.Cmdx("web.chat.keyboard", mdb.CREATE, web.SPACE, m.Option(ice.MSG_DAEMON), ctx.INDEX, m.Option(ctx.INDEX), "input", "")
				link := tcp.ReplaceLocalhost(m, web.MergePodCmd(m, "", "web.chat.keyboard", mdb.HASH, hash))
				m.PushQRCode(mdb.TEXT, link)
				m.Push(mdb.NAME, link)
			}},
			"_open": {Name: "_open", Help: "打开", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(cli.DAEMON, cli.OPEN, "-a", kit.Split(arg[0], ice.PT, ice.PT)[0])
			}},
			cli.MAKE: {Name: "make", Help: "构建", Hand: func(m *ice.Message, arg ...string) {
				_vimer_make(m, m.Option(nfs.PATH), m.Cmd(cli.SYSTEM, cli.MAKE, arg))
			}},

			TEMPLATE: {Name: "template", Help: "模板", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(TEMPLATE, kit.Ext(m.Option(mdb.FILE)), m.Option(nfs.FILE), m.Option(nfs.PATH))
			}},
			COMPLETE: {Name: "complete", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(COMPLETE, kit.Ext(m.Option(mdb.FILE)), m.Option(nfs.FILE), m.Option(nfs.PATH))
			}},

			DEVPACK: {Name: "devpack", Help: "开发模式", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(nfs.LINK, ice.GO_SUM, path.Join(ice.SRC_DEBUG, ice.GO_SUM))
				m.Cmd(nfs.LINK, ice.GO_MOD, path.Join(ice.SRC_DEBUG, ice.GO_MOD))
				m.Cmdy(nfs.CAT, ice.GO_MOD)
				m.Cmdy(WEBPACK, mdb.REMOVE)
				web.ToastSuccess(m)
				m.ProcessInner()
			}},
			BINPACK: {Name: "binpack", Help: "打包模式", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(WEBPACK, mdb.CREATE)
				m.Cmdy(AUTOGEN, BINPACK)
				web.ToastSuccess(m)
				m.ProcessInner()
			}},
			AUTOGEN: {Name: "create name=h2 help=示例 type=Zone,Hash,Data,Code main=main.go zone key", Help: "模块", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(AUTOGEN, mdb.CREATE, arg)
			}},
			COMPILE: {Name: "compile", Help: "编译", Hand: func(m *ice.Message, arg ...string) {
				cmds := []string{COMPILE, ice.SRC_MAIN_GO, ice.BIN_ICE_BIN}
				if strings.HasSuffix(os.Args[0], "contexts.app/Contents/MacOS/contexts") {
					m.Option(cli.ENV, "CGO_ENABLED", "1", cli.HOME, kit.Env(cli.HOME), cli.PATH, kit.Path("usr/local/go/bin")+ice.DF+kit.Env(cli.PATH))
					cmds = []string{COMPILE, "src/webview.go", "usr/publish/contexts.app/Contents/MacOS/contexts"}
					// cmds = []string{cli.SYSTEM, cli.MAKE, "app"}
				}
				if msg := m.Cmd(cmds); cli.IsSuccess(msg) {
					m.Cmd(UPGRADE, cli.RESTART)
				} else {
					_vimer_make(m, nfs.PWD, msg)
				}
			}},
			PUBLISH: {Name: "publish", Help: "发布", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(PUBLISH, ice.CONTEXTS)
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			if m.Cmdy(INNER, arg); arg[0] != ctx.ACTION {
				m.Action(nfs.SAVE, COMPILE, AUTOGEN, nfs.SCRIPT, web.WEBSITE, web.DREAM, XTERM)
				m.Option("tabs", m.Config("show.tabs"))
				m.Option("plug", m.Config("show.plug"))
				m.Option("exts", m.Config("show.exts"))
				ctx.DisplayLocal(m, "")
			}
		}},
	})
}

const TEMPLATE = "template"
const COMPLETE = "complete"
const NAVIGATE = "navigate"

func init() {
	Index.MergeCommands(ice.Commands{TEMPLATE: {Name: "template type name text auto", Help: "模板", Actions: mdb.RenderAction()}})
	Index.MergeCommands(ice.Commands{COMPLETE: {Name: "complete type name text auto", Help: "补全", Actions: mdb.RenderAction()}})
	Index.MergeCommands(ice.Commands{NAVIGATE: {Name: "navigate type name text auto", Help: "跳转", Actions: mdb.RenderAction()}})
}
func LangAction() ice.Actions {
	return ice.Actions{
		ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
			for _, cmd := range []string{TEMPLATE, COMPLETE, NAVIGATE} {
				m.Cmd(cmd, mdb.CREATE, m.CommandKey(), m.PrefixKey())
			}
		}},
		TEMPLATE: {Hand: func(m *ice.Message, arg ...string) {}},
		COMPLETE: {Hand: func(m *ice.Message, arg ...string) {}},
		NAVIGATE: {Hand: func(m *ice.Message, arg ...string) {}},
	}
}
func Complete(m *ice.Message, text string, data ice.Map) {
	if strings.HasSuffix(text, ice.PT) {
		key := kit.Slice(kit.Split(text, " ."), -1)[0]
		m.Push(mdb.TEXT, kit.Simple(data[key]))
	} else {
		m.Push(mdb.TEXT, data[""])
		for k := range data {
			m.Push(mdb.TEXT, k)
		}
	}
}
