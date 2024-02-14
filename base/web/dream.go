package web

import (
	"net/http"
	"os"
	"path"
	"regexp"
	"runtime"
	"strings"
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
)

func _dream_list(m *ice.Message, simple bool) *ice.Message {
	if ice.Info.NodeType == WORKER {
		return m
	}
	list := m.CmdMap(SPACE, mdb.NAME)
	mdb.HashSelect(m).Table(func(value ice.Maps) {
		if space, ok := list[value[mdb.NAME]]; ok {
			if m.IsCliUA() || simple {
				m.Push(mdb.TYPE, space[mdb.TYPE]).Push(cli.STATUS, cli.START)
				m.Push(nfs.MODULE, space[nfs.MODULE]).Push(nfs.VERSION, space[nfs.VERSION]).Push(mdb.TEXT, "")
				m.PushButton(cli.STOP)
			} else {
				msg := gdb.Event(m.Spawn(value, space), DREAM_TABLES).Copy(m.Spawn().PushButton(cli.STOP))
				m.Push(mdb.TYPE, space[mdb.TYPE]).Push(cli.STATUS, cli.START)
				m.Push(nfs.MODULE, space[nfs.MODULE]).Push(nfs.VERSION, space[nfs.VERSION]).Push(mdb.TEXT, msg.Append(mdb.TEXT))
				m.PushButton(strings.Join(msg.Appendv(ctx.ACTION), ""))
			}
		} else {
			if m.Push(mdb.TYPE, WORKER); nfs.Exists(m, path.Join(ice.USR_LOCAL_WORK, value[mdb.NAME])) {
				m.Push(cli.STATUS, cli.STOP)
				m.Push(nfs.MODULE, "").Push(nfs.VERSION, "").Push(mdb.TEXT, "")
				m.PushButton(cli.START, nfs.TRASH)
			} else {
				m.Push(cli.STATUS, cli.BEGIN)
				m.Push(nfs.MODULE, "").Push(nfs.VERSION, "").Push(mdb.TEXT, "")
				m.PushButton(cli.START, mdb.REMOVE)
			}
		}
	})
	m.RewriteAppend(func(value, key string, index int) string {
		if key == mdb.TIME {
			if space, ok := list[m.Appendv(mdb.NAME)[index]]; ok {
				value = space[mdb.TIME]
			}
		}
		return value
	})
	return m
}
func _dream_list_icon(m *ice.Message) {
	m.RewriteAppend(func(value, key string, index int) string {
		if key == mdb.ICONS {
			if kit.HasPrefix(value, HTTP, nfs.PS) {
				return value
			} else if nfs.ExistsFile(m, path.Join(ice.USR_LOCAL_WORK, m.Appendv(mdb.NAME)[index], value)) {
				return m.Spawn(kit.Dict(ice.MSG_USERPOD, m.Appendv(mdb.NAME)[index])).FileURI(value)
			} else if nfs.ExistsFile(m, value) {
				return m.FileURI(value)
			} else {
				return m.FileURI(nfs.USR_ICONS_ICEBERGS)
			}
		}
		return value
	})
}
func _dream_list_more(m *ice.Message, simple bool) *ice.Message {
	if m.IsCliUA() {
		return m
	}
	list := m.Spawn(ice.Maps{ice.MSG_FIELDS: ""}).CmdMap(SPIDE, CLIENT_NAME)
	m.Cmds(SPACE).Table(func(value ice.Maps) {
		value[mdb.ICONS] = nfs.USR_ICONS_VOLCANOS
		value[nfs.REPOS] = "https://" + value[nfs.MODULE]
		value[mdb.STATUS] = cli.START
		switch value[mdb.TYPE] {
		case SERVER:
			value[mdb.ICONS] = nfs.USR_ICONS_ICEBERGS
			value[mdb.TEXT] = kit.JoinLine(value[nfs.MODULE], value[mdb.TEXT])

			if simple {
				defer m.PushButton("")
			} else {
				msg := gdb.Event(m.Spawn(value), DREAM_TABLES)
				defer m.PushButton(strings.Join(msg.Appendv(ctx.ACTION), ""))
			}
		case MASTER:
			if spide, ok := list[value[mdb.NAME]]; ok {
				value[mdb.ICONS] = kit.Select(value[mdb.ICONS], spide[mdb.ICONS])
			}
			value[mdb.TEXT] = kit.JoinLine(value[nfs.MODULE], value[mdb.TEXT])
			if simple {
				defer m.PushButton("")
			} else {
				msg := gdb.Event(m.Spawn(value), DREAM_TABLES)
				defer m.PushButton(strings.Join(msg.Appendv(ctx.ACTION), ""))
			}
		case aaa.LOGIN:
			value[mdb.ICONS] = kit.Select(value[mdb.ICONS], agentIcons[value[AGENT]])
			value[mdb.TEXT] = kit.JoinWord(value[AGENT], value[cli.SYSTEM], value[aaa.IP])
			defer m.PushButton(GRANT)
		default:
			return
		}
		m.Push("", value, kit.Split(mdb.Config(m, mdb.FIELD)+",type,status,module,version,text"))
	})
	return m
}
func _dream_start(m *ice.Message, name string) {
	if m.WarnNotValid(name == "", mdb.NAME) {
		return
	}
	defer ToastProcess(m)()
	defer m.ProcessOpen(m.MergePod(name))
	defer mdb.Lock(m, m.PrefixKey(), cli.START, name)()
	p := path.Join(ice.USR_LOCAL_WORK, name)
	if p := path.Join(p, ice.Info.PidPath); nfs.Exists(m, p) {
		if pid := m.Cmdx(nfs.CAT, p, kit.Dict(ice.MSG_USERROLE, aaa.TECH)); pid != "" && nfs.Exists(m, "/proc/"+pid) {
			m.Info("already exists %v", pid)
			return
		}
		for i := 0; i < 3; i++ {
			if m.Cmd(SPACE, name).Length() > 0 {
				m.Info("already exists %v", name)
				return
			}
			m.Sleep300ms()
		}
	}
	defer m.Options(cli.CMD_DIR, "", cli.CMD_ENV, "", cli.CMD_OUTPUT, "")
	m.Options(cli.CMD_DIR, kit.Path(p), cli.CMD_ENV, kit.EnvList(kit.Simple(m.OptionSimple(ice.TCP_DOMAIN),
		cli.CTX_OPS, HostPort(m, tcp.LOCALHOST, m.Cmdv(SERVE, tcp.PORT)), cli.CTX_LOG, ice.VAR_LOG_BOOT_LOG, cli.CTX_PID, ice.VAR_LOG_ICE_PID,
		cli.CTX_ROOT, kit.Path(""), cli.PATH, cli.BinPath(p, ""), cli.USER, ice.Info.Username,
	)...), cli.CMD_OUTPUT, path.Join(p, ice.VAR_LOG_BOOT_LOG), mdb.CACHE_CLEAR_ONEXIT, ice.TRUE)
	gdb.Event(m, DREAM_CREATE, m.OptionSimple(mdb.NAME, mdb.TYPE, cli.CMD_DIR))
	kit.If(m.Option(nfs.BINARY) == "" && cli.SystemFind(m, "go") == "", func(p string) { m.Option(nfs.BINARY, SpideOrigin(m, ice.DEV_IP)+S(name)) })
	kit.If(m.Option(nfs.BINARY), func(p string) { _dream_binary(m, p) })
	kit.If(m.Option(nfs.TEMPLATE), func(p string) { _dream_template(m, p) })
	bin := kit.Select(kit.Path(os.Args[0]), cli.SystemFind(m, ice.ICE_BIN, nfs.PWD+path.Join(p, ice.BIN), nfs.PWD+ice.BIN))
	m.Cmd(cli.DAEMON, bin, SPACE, tcp.DIAL, ice.DEV, ice.OPS, mdb.TYPE, WORKER, m.OptionSimple(mdb.NAME), cli.DAEMON, ice.OPS)
	gdb.WaitEvent(m, DREAM_OPEN, func(m *ice.Message, arg ...string) bool { return m.Option(mdb.NAME) == name })
	m.Sleep300ms()
}
func _dream_binary(m *ice.Message, p string) {
	if bin := path.Join(m.Option(cli.CMD_DIR), ice.BIN_ICE_BIN); nfs.Exists(m, bin) {
		return
	} else if kit.IsUrl(p) {
		GoToast(m, DOWNLOAD, func(toast func(string, int, int)) (list []string) {
			begin := time.Now()
			SpideSave(m, bin, kit.MergeURL(p, cli.GOOS, runtime.GOOS, cli.GOARCH, runtime.GOARCH), func(count, total, value int) {
				toast(m.Option(mdb.NAME)+" "+kit.FormatShow(cli.COST, kit.FmtDuration(time.Now().Sub(begin))), count, total)
			})
			return nil
		})
		os.Chmod(bin, ice.MOD_DIR)
	} else {
		m.Cmd(nfs.LINK, bin, kit.Path(p))
	}
}
func _dream_template(m *ice.Message, p string) {
	kit.For([]string{
		ice.README_MD, ice.MAKEFILE, ice.LICENSE, ice.GO_MOD, ice.GO_SUM,
		ice.SRC_MAIN_SH, ice.SRC_MAIN_SHY, ice.SRC_MAIN_GO, ice.SRC_MAIN_JS,
		ice.ETC_MISS_SH, ice.ETC_INIT_SHY, ice.ETC_EXIT_SHY,
	}, func(file string) {
		if nfs.Exists(m, kit.Path(m.Option(cli.CMD_DIR), file)) {
			return
		}
		switch m.Cmdy(nfs.COPY, kit.Path(m.Option(cli.CMD_DIR), file), kit.Path(ice.USR_LOCAL_WORK, p, file)); file {
		case ice.GO_MOD:
			nfs.Rewrite(m, path.Join(p, file), func(line string) string {
				return kit.Select(line, nfs.MODULE+lex.SP+m.Option(mdb.NAME), strings.HasPrefix(line, nfs.MODULE))
			})
		}
	})
}

const (
	ALWAYS   = "always"
	STARTALL = "startall"
	STOPALL  = "stopall"
	PUBLISH  = "publish"
	FOR_EACH = "forEach"
	FOR_FLOW = "forFlow"

	DREAM_CREATE = "dream.create"
	DREAM_REMOVE = "dream.remove"
	DREAM_START  = "dream.start"
	DREAM_STOP   = "dream.stop"
	DREAM_OPEN   = "dream.open"
	DREAM_CLOSE  = "dream.close"
	DREAM_TRASH  = "dream.trash"

	DREAM_INPUTS = "dream.inputs"
	DREAM_TABLES = "dream.tables"
	DREAM_ACTION = "dream.action"
)
const DREAM = "dream"

func init() {
	Index.MergeCommands(ice.Commands{
		DREAM: {Name: "dream refresh", Help: "梦想家", Icon: "Launchpad.png", Role: aaa.VOID, Meta: kit.Dict(
			ice.CTX_TRANS, kit.Dict(html.INPUT, kit.Dict(WORKER, "空间", SERVER, "机器", MASTER, "服务")),
		), Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m = m.Spawn()
				m.GoSleep("10s", func() {
					mdb.HashSelects(m).Table(func(value ice.Maps) {
						if value[cli.RESTART] == ALWAYS && nfs.Exists(m, path.Join(ice.USR_LOCAL_WORK+value[mdb.NAME])) {
							m.Cmd(DREAM, cli.START, kit.Dict(mdb.NAME, value[mdb.NAME]))
						}
					})
				})
				m.GoSleep("1s", func() {
					for _, cmd := range kit.Reverse(kit.Split(mdb.Config(m, html.BUTTON))) {
						m.Cmd(gdb.EVENT, gdb.LISTEN, gdb.EVENT, DREAM_TABLES, ice.CMD, cmd)
						m.Cmd(gdb.EVENT, gdb.LISTEN, gdb.EVENT, DREAM_ACTION, ice.CMD, cmd)
					}
					aaa.White(m, kit.Keys(m.PrefixKey(), ctx.ACTION, DREAM_ACTION, ctx.RUN))
				})
			}},
			html.BUTTON: {Hand: func(m *ice.Message, arg ...string) { mdb.Config(m, html.BUTTON, kit.Join(arg)) }},
			mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
				if mdb.IsSearchPreview(m, arg) {
					mdb.HashSelects(m.Spawn()).Table(func(value ice.Maps) { m.PushSearch(mdb.TYPE, WORKER, mdb.TEXT, m.MergePod(value[mdb.NAME]), value) })
				}
			}},
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch m.Option(ctx.ACTION) {
				case mdb.CREATE:
					switch arg[0] {
					case mdb.NAME, nfs.TEMPLATE:
						_dream_list(m, true).Cut("name,status,time")
						return
					case nfs.BINARY:
						m.Cmdy(nfs.DIR, ice.BIN, "path,size,time", kit.Dict(nfs.DIR_TYPE, nfs.TYPE_BIN))
						m.Cmd(nfs.DIR, ice.USR_LOCAL_WORK, kit.Dict(nfs.DIR_TYPE, nfs.TYPE_BOTH), func(value ice.Maps) {
							m.Cmdy(nfs.DIR, path.Join(value[nfs.PATH], ice.BIN), "path,size,time", kit.Dict(nfs.DIR_TYPE, nfs.TYPE_BIN))
						})
						m.RenameAppend(nfs.PATH, arg[0])
						mdb.HashInputs(m, arg)
						DreamListSpide(m, []string{ice.DEV}, MASTER, func(dev, origin string) {
							m.Spawn().SplitIndex(m.Cmdx(SPIDE, dev, SPIDE_RAW, http.MethodGet, S(), cli.GOOS, runtime.GOOS, cli.GOARCH, runtime.GOARCH)).Table(func(value ice.Maps) {
								m.Push(arg[0], origin+S(value[mdb.NAME])).Push(nfs.SIZE, value[nfs.SIZE]).Push(mdb.TIME, value[mdb.TIME])
							})
						})
					case mdb.ICONS:
						mdb.HashInputs(m, arg)
						return
					}
				case STARTALL:
					DreamEach(m, "", cli.STOP, func(name string) { m.Push(arg[0], name) })
					return
				case MAIN:
					m.Cmdy(SPACE, m.Option(mdb.NAME), SPACE, mdb.INPUTS, arg)
					return
				case "send":
					m.Cmd(SPACE, func(value ice.Maps) {
						kit.If(kit.IsIn(value[mdb.TYPE], SERVER), func() { m.Push(arg[0], value[mdb.NAME]) })
					})
					return
				}
				switch arg[0] {
				case mdb.NAME:
					DreamEach(m, "", cli.START, func(name string) { m.Push(arg[0], name) })
				case nfs.FILE:
					m.Options(nfs.DIR_TYPE, nfs.TYPE_CAT, ice.MSG_FIELDS, nfs.PATH)
					m.Cmdy(nfs.DIR, nfs.SRC).Cmdy(nfs.DIR, nfs.ETC).Cmdy(nfs.DIR, "")
				case ctx.CMDS:
					m.Cmdy(ctx.COMMAND)
				default:
					gdb.Event(m, DREAM_INPUTS, arg)
				}
			}},
			mdb.CREATE: {Name: "create name*=hi icons repos binary template", Hand: func(m *ice.Message, arg ...string) {
				kit.If(!strings.Contains(m.Option(mdb.NAME), "-") || !strings.HasPrefix(m.Option(mdb.NAME), "20"), func() { m.Option(mdb.NAME, m.Time("20060102-")+m.Option(mdb.NAME)) })
				kit.If(mdb.Config(m, nfs.BINARY), func(p string) { m.OptionDefault(nfs.BINARY, p+m.Option(mdb.NAME)) })
				kit.If(mdb.Config(m, nfs.REPOS), func(p string) { m.OptionDefault(nfs.REPOS, p+m.Option(mdb.NAME)) })
				m.Option(nfs.REPOS, kit.Select("", kit.Slice(kit.Split(m.Option(nfs.REPOS)), -1), 0))
				m.OptionDefault(mdb.ICONS, nfs.USR_ICONS_CONTEXTS)
				if mdb.HashCreate(m); !m.IsCliUA() {
					_dream_start(m, m.Option(mdb.NAME))
				}
			}},
			STARTALL: {Name: "startall name", Help: "启动", Icon: "bi bi-play-circle", Hand: func(m *ice.Message, arg ...string) {
				DreamEach(m, m.Option(mdb.NAME), cli.STOP, func(name string) {
					m.Cmd("", cli.START, ice.Maps{mdb.NAME: name, ice.MSG_DAEMON: ""})
				})
			}},
			STOPALL: {Name: "stopall name", Help: "停止", Icon: "bi bi-stop-circle", Hand: func(m *ice.Message, arg ...string) {
				DreamEach(m, m.Option(mdb.NAME), cli.START, func(name string) {
					m.Cmd("", cli.STOP, ice.Maps{mdb.NAME: name, ice.MSG_DAEMON: ""})
				})
			}},
			cli.BUILD: {Name: "build name", Hand: func(m *ice.Message, arg ...string) {
				compile := cli.SystemFind(m, "go") != ""
				m.Option(ice.MSG_TITLE, kit.Keys(m.Option(ice.MSG_USERPOD), m.CommandKey(), m.ActionKey()))
				m.Cmd("", FOR_FLOW, m.Option(mdb.NAME), kit.JoinWord(cli.SH, ice.ETC_MISS_SH), func(p string) bool {
					if compile && nfs.Exists(m, path.Join(p, ice.SRC_MAIN_GO)) {
						return false
					} else {
						m.Cmd(SPACE, path.Base(p), cli.RUNTIME, UPGRADE)
						return true
					}
				})
				kit.If(m.Option(mdb.NAME) == "", func() { m.Sleep3s().Cmdy(ROUTE, cli.BUILD).ProcessInner() })
			}},
			PUBLISH: {Name: "publish name", Help: "发布", Icon: "bi bi-send-check", Hand: func(m *ice.Message, arg ...string) {
				m.Option(ice.MSG_TITLE, kit.Keys(m.CommandKey(), m.ActionKey()))
				defer ToastProcess(m)()
				list := []string{cli.LINUX, cli.DARWIN, cli.WINDOWS}
				msg := m.Spawn(ice.Maps{ice.MSG_DAEMON: ""})
				m.Cmd(AUTOGEN, BINPACK)
				kit.For(list, func(goos string) {
					PushNoticeRich(m, mdb.NAME, ice.Info.NodeName, msg.Cmd(COMPILE, goos, cli.AMD64).AppendSimple())
				})
				DreamEach(m, m.Option(mdb.NAME), "", func(name string) {
					m.Cmd(SPACE, name, AUTOGEN, BINPACK)
					kit.For(list, func(goos string) {
						PushNoticeRich(m, mdb.NAME, name, msg.Cmd(SPACE, name, COMPILE, goos, cli.AMD64, kit.Dict(ice.MSG_USERPOD, name)).AppendSimple())
					})
				})
				m.ProcessHold()
			}},
			FOR_FLOW: {Name: "forFlow name cmd*='sh etc/miss.sh'", Help: "流程", Icon: "bi bi-terminal", Hand: func(m *ice.Message, arg ...string) {
				list := []string{}
				DreamEach(m.Spawn(ice.Maps{ice.MSG_DAEMON: ""}), m.Option(mdb.NAME), "", func(name string) { list = append(list, name) })
				m.Options(ctx.DISPLAY, PLUGIN_XTERM, cli.CMD_OUTPUT, nfs.NewWriteCloser(func(buf []byte) (int, error) {
					PushNoticeGrow(m.Options(ice.MSG_COUNT, "0"), strings.ReplaceAll(string(buf), lex.NL, "\r\n"))
					return len(buf), nil
				}, func() error { return nil }))
				GoToast(m, "", func(toast func(string, int, int)) []string {
					kit.For(list, func(index int, name string) {
						toast(name, index, len(list))
						p := path.Join(ice.USR_LOCAL_WORK, name)
						switch cb := m.OptionCB("").(type) {
						case func(string) bool:
							if cb(p) {
								return
							}
						}
						PushNoticeGrow(m, strings.ReplaceAll(kit.Format("[%s]%s$ %s\n", time.Now().Format(ice.MOD_TIME_ONLY), name, m.Option(ice.CMD)), lex.NL, "\r\n"))
						m.Cmd(cli.SYSTEM, kit.Split(m.Option(ice.CMD)), kit.Dict(cli.CMD_DIR, p)).Sleep300ms()
						PushNoticeGrow(m, "\r\n\r\n\r\n")
					})
					return nil
				})
			}},
			ctx.CMDS: {Name: "cmds name cmds*", Help: "命令", Icon: "bi bi-terminal", Hand: func(m *ice.Message, arg ...string) {
				DreamEach(m, m.Option(mdb.NAME), "", func(name string) {
					m.Push(mdb.NAME, name).Push(mdb.TEXT, m.Cmdx(SPACE, name, kit.Split(m.Option(ice.CMD))))
				}).StatusTimeCount(m.OptionSimple(ctx.CMDS))
			}},
			nfs.CAT: {Name: "cat name file*", Help: "文件", Icon: "bi bi-file-earmark-code", Hand: func(m *ice.Message, arg ...string) {
				DreamEach(m, m.Option(mdb.NAME), "", func(name string) {
					m.Push(mdb.NAME, name).Push(mdb.TEXT, m.Cmdx(SPACE, name, nfs.CAT, m.Option(nfs.FILE)))
				}).StatusTimeCount(m.OptionSimple(nfs.FILE))
			}},
			cli.START: {Hand: func(m *ice.Message, arg ...string) {
				defer gdb.Event(m, DREAM_START, arg)
				_dream_start(m, m.Option(mdb.NAME))
			}},
			cli.STOP: {Hand: func(m *ice.Message, arg ...string) {
				defer ToastProcess(m)()
				gdb.Event(m, DREAM_STOP, arg)
				m.Cmd(SPACE, mdb.MODIFY, m.OptionSimple(mdb.NAME), mdb.STATUS, cli.STOP)
				m.Cmd(SPACE, m.Option(mdb.NAME), ice.EXIT)
				m.Sleep3s()
			}},
			nfs.TRASH: {Hand: func(m *ice.Message, arg ...string) {
				gdb.Event(m, DREAM_TRASH, arg)
				nfs.Trash(m, path.Join(ice.USR_LOCAL_WORK, m.Option(mdb.NAME)))
			}},
			tcp.SEND: {Name: "send space*", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(SPACE, m.Option(SPACE), DREAM, mdb.CREATE, m.OptionSimple(mdb.NAME, mdb.ICONS, nfs.REPOS, nfs.BINARY))
				m.Cmd(SPACE, m.Option(SPACE), DREAM, cli.START, m.OptionSimple(mdb.NAME))
				ProcessIframe(m, "", m.MergePod(kit.Keys(m.Option(SPACE), m.Option(mdb.NAME))))
			}},
			GRANT: {Help: "授权", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(CHAT_GRANT, aaa.CONFIRM, kit.Dict(SPACE, m.Option(mdb.NAME)))
			}},
			OPEN: {Role: aaa.VOID, Hand: func(m *ice.Message, arg ...string) {
				ProcessIframe(m, m.Option(mdb.NAME), kit.Select(m.MergePod(m.Option(mdb.NAME), SpideOrigin(m, m.Option(mdb.NAME))+C(ADMIN)), m.Option(mdb.TYPE) == MASTER), arg...)
			}},
			MAIN: {Name: "main index", Help: "首页", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(SPACE, m.Option(mdb.NAME), SPACE, ice.MAIN, m.Option(ctx.INDEX))
			}},
			DREAM_OPEN: {Hand: func(m *ice.Message, arg ...string) {}},
			DREAM_CLOSE: {Hand: func(m *ice.Message, arg ...string) {
				if m.Option(cli.DAEMON) == ice.OPS && m.Cmdv(SPACE, m.Option(mdb.NAME), mdb.STATUS) != cli.STOP {
					m.GoSleep300ms(func() { m.Cmd(DREAM, cli.START, m.OptionSimple(mdb.NAME)) })
				}
			}},
			DREAM_TABLES: {Hand: func(m *ice.Message, arg ...string) {
				switch m.Option(mdb.TYPE) {
				case WORKER:
					m.PushButton(tcp.SEND, OPEN)
				default:
					m.PushButton(OPEN)
				}
			}},
			STATS_TABLES: {Hand: func(m *ice.Message, arg ...string) {
				if msg := mdb.HashSelects(m.Spawn()); msg.Length() > 0 {
					stat := map[string]int{}
					list := m.CmdMap(SPACE, mdb.NAME)
					msg.Table(func(value ice.Maps) {
						if _, ok := list[value[mdb.NAME]]; ok {
							stat[cli.START]++
						}
					})
					PushStats(m, kit.Keys(m.CommandKey(), cli.START), stat[cli.START], "", "空间总数")
					PushStats(m, kit.Keys(m.CommandKey(), mdb.TOTAL), msg.Length(), "", "已启动空间")
				}
			}},
		}, StatsAction(), DreamAction(), mdb.ImportantHashAction(
			mdb.SHORT, mdb.NAME, mdb.FIELD, "time,name,icons,repos,binary,template,restart", ctx.TOOLS, kit.Simple(STORE, SPIDE),
			html.BUTTON, kit.JoinWord(PORTAL, ADMIN, DESKTOP, WIKI_WORD, STATUS, VIMER, XTERM, COMPILE),
		)), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				simple := m.Option("dream.simple") == ice.TRUE
				_dream_list(m, simple)
				_dream_list_icon(m)
				_dream_list_more(m, simple)
				stat := map[string]int{}
				m.Table(func(value ice.Maps) { stat[value[mdb.TYPE]]++; stat[value[mdb.STATUS]]++ })
				kit.If(stat[cli.START] == stat[WORKER], func() { delete(stat, cli.START) })
				m.Sort("type,status,name", []string{aaa.LOGIN, WORKER, SERVER, MASTER}, []string{cli.START, cli.STOP, cli.BEGIN}, ice.STR_R).StatusTimeCount(stat)
				if ctx.DisplayTableCard(m); ice.Info.NodeType == WORKER || !aaa.IsTechOrRoot(m) || m.IsCliUA() {
					m.Action()
				} else if cli.SystemFind(m, "go") == "" {
					m.Action(html.FILTER, mdb.CREATE, STARTALL, STOPALL)
				} else {
					m.Action(html.FILTER, mdb.CREATE, STARTALL, STOPALL, cli.BUILD, PUBLISH)
				}
			} else if arg[0] == ctx.ACTION {
				gdb.Event(m, DREAM_ACTION, arg)
			} else {
				mdb.HashSelects(m, arg[0])
			}
		}},
	})
}

func DreamAction() ice.Actions {
	return ice.MergeActions(ice.Actions{
		DREAM_ACTION: {Hand: func(m *ice.Message, arg ...string) { DreamProcess(m, nil, arg...) }},
	}, gdb.EventsAction(DREAM_OPEN, DREAM_CLOSE, DREAM_INPUTS, DREAM_CREATE, DREAM_TRASH, DREAM_TABLES, DREAM_ACTION, SERVE_START))
}
func DreamProcess(m *ice.Message, args ice.Any, arg ...string) {
	if kit.HasPrefixList(arg, ctx.RUN) {
		ctx.ProcessFloat(m, m.PrefixKey(), args, kit.Slice(arg, 1)...)
	} else if kit.HasPrefixList(arg, ctx.ACTION, m.PrefixKey()) || kit.HasPrefixList(arg, ctx.ACTION, m.CommandKey()) {
		if m.Option(mdb.TYPE) == MASTER && (kit.IsIn(ctx.ShortCmd(m.PrefixKey()), PORTAL, DESKTOP)) {
			if ProcessIframe(m, "", SpideOrigin(m, m.Option(mdb.NAME))+C(m.PrefixKey()), arg...); !m.IsMetaKey() {
				m.ProcessField(ctx.ACTION, ctx.RUN, CHAT_IFRAME)
			}
		} else if arg = kit.Slice(arg, 2); kit.HasPrefixList(arg, DREAM) {
			m.Cmdy(SPACE, m.Option(ice.MSG_USERPOD, arg[1]), m.PrefixKey(), ctx.ACTION, DREAM_ACTION, ctx.RUN, arg[2:])
		} else if dream := m.Option(mdb.NAME); dream != "" {
			m.Cmdy(SPACE, dream, m.PrefixKey(), ctx.ACTION, DREAM_ACTION, ctx.RUN, arg).Optionv(ice.FIELD_PREFIX, kit.Simple(ctx.ACTION, m.PrefixKey(), DREAM, dream, ctx.RUN))
			m.Push("_space", dream)
		}
	}
}
func DreamEach(m *ice.Message, name string, status string, cb func(string)) *ice.Message {
	reg, err := regexp.Compile(name)
	if m.WarnNotValid(err) {
		return m
	}
	list := []string{}
	m.Spawn().Cmds(DREAM).Table(func(value ice.Maps) {
		if value[mdb.STATUS] == kit.Select(cli.START, status) && reg.MatchString(value[mdb.NAME]) && value[mdb.TYPE] == WORKER {
			list = append(list, value[mdb.NAME])
		}
	})
	if len(list) == 0 {
		return m
	}
	GoToast(m, "", func(toast func(string, int, int)) []string {
		kit.For(list, func(index int, name string) { toast(name, index, len(list)); cb(name) })
		return nil
	})
	return m
}
func DreamList(m *ice.Message) *ice.Message {
	return m.SplitIndex(AdminCmd(m, DREAM).Result())
}
func DreamWhiteHandle(m *ice.Message, arg ...string) {
	aaa.White(m, kit.Keys(DREAM, ctx.ACTION, m.CommandKey()))
	aaa.White(m, kit.Keys(DREAM, ctx.ACTION, m.PrefixKey()))
	aaa.White(m, kit.Keys(ctx.ShortCmd(m.PrefixKey()), ctx.ACTION, DREAM_ACTION))
}
func DreamListSpide(m *ice.Message, list []string, types string, cb func(dev, origin string)) {
	m.Cmds(DREAM).Table(func(value ice.Maps) {
		kit.If(value[mdb.TYPE] == types, func() { list = append(list, value[mdb.NAME]) })
	})
	has := map[string]bool{}
	GoToast(m, "", func(toast func(name string, count, total int)) []string {
		kit.For(list, func(index int, dev string) {
			toast(dev, index, len(list))
			if origin := m.Cmdv(SPIDE, dev, CLIENT_ORIGIN); !has[origin] {
				has[origin] = true
				cb(dev, origin)
			}
		})
		return nil
	})
}
