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
	list := m.CmdMap(SPACE, mdb.NAME)
	mdb.HashSelects(m.Spawn()).Table(func(value ice.Maps, index int, head []string) {
		if value[aaa.ACCESS] == aaa.PRIVATE && (m.Option(ice.FROM_SPACE) != "" || !aaa.IsTechOrRoot(m)) {
			return
		}
		if space, ok := list[value[mdb.NAME]]; ok {
			m.Push("", value, kit.Slice(head, 0, -1))
			if m.IsCliUA() || simple {
				m.Push(mdb.TYPE, space[mdb.TYPE]).Push(cli.STATUS, cli.START)
				m.Push(nfs.MODULE, space[nfs.MODULE]).Push(nfs.VERSION, space[nfs.VERSION]).Push(mdb.TEXT, DreamStat(m, value[mdb.NAME]))
				if aaa.IsTechOrRoot(m) {
					m.PushButton(cli.STOP)
				} else {
					m.PushButton()
				}
			} else {
				msg := gdb.Event(m.Spawn(value, space), DREAM_TABLES)
				if aaa.IsTechOrRoot(m) {
					msg.Copy(m.Spawn().PushButton(cli.STOP))
				}
				m.Push(mdb.TYPE, space[mdb.TYPE]).Push(cli.STATUS, cli.START)
				m.Push(nfs.MODULE, space[nfs.MODULE]).Push(nfs.VERSION, space[nfs.VERSION]).Push(mdb.TEXT, msg.Append(mdb.TEXT))
				m.PushButton(strings.Join(msg.Appendv(ctx.ACTION), ""))
			}
		} else if aaa.IsTechOrRoot(m) {
			m.Push("", value, kit.Slice(head, 0, -1))
			if m.Push(mdb.TYPE, WORKER); nfs.Exists(m, path.Join(ice.USR_LOCAL_WORK, value[mdb.NAME])) {
				m.Push(cli.STATUS, cli.STOP)
				m.Push(nfs.MODULE, "").Push(nfs.VERSION, "").Push(mdb.TEXT, "")
				if aaa.IsTechOrRoot(m) {
					m.PushButton(cli.START, nfs.TRASH)
				} else {
					m.PushButton()
				}
			} else {
				m.Push(cli.STATUS, cli.BEGIN)
				m.Push(nfs.MODULE, "").Push(nfs.VERSION, "").Push(mdb.TEXT, "")
				if aaa.IsTechOrRoot(m) {
					m.PushButton(cli.START, mdb.REMOVE)
				} else {
					m.PushButton()
				}
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
	list := m.CmdMap(SPIDE, CLIENT_NAME)
	m.Cmds(SPACE).Table(func(value ice.Maps) {
		value[mdb.ICONS] = nfs.USR_ICONS_VOLCANOS
		value[nfs.REPOS] = "https://" + value[nfs.MODULE]
		value[aaa.ACCESS] = kit.Select("", value[aaa.USERROLE], value[aaa.USERROLE] != aaa.VOID)
		value[mdb.STATUS] = cli.START
		switch value[mdb.TYPE] {
		case SERVER:
			value[mdb.ICONS] = m.FileURI(nfs.USR_ICONS_ICEBERGS)
			value[mdb.TEXT] = kit.JoinLine(value[nfs.MODULE], value[mdb.TEXT])
			if simple {
				defer m.PushButton("")
			} else {
				msg := gdb.Event(m.Spawn(value), DREAM_TABLES)
				defer m.PushButton(strings.Join(msg.Appendv(ctx.ACTION), ""))
			}
		case ORIGIN:
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
	if !m.IsCliUA() {
		defer m.ProcessOpenAndRefresh(m.MergePod(name))
		defer ToastProcess(m, mdb.CREATE, name)()
	}
	defer mdb.Lock(m, m.PrefixKey(), cli.START, name)()
	p := _dream_check(m, name)
	if p == "" {
		return
	}
	if !nfs.Exists(m, p) {
		gdb.Event(m, DREAM_CREATE, m.OptionSimple(mdb.NAME))
	}
	defer m.Options(cli.CMD_DIR, "", cli.CMD_ENV, "", cli.CMD_OUTPUT, "")
	m.Options(cli.CMD_DIR, kit.Path(p), cli.CMD_ENV, kit.EnvList(kit.Simple(m.OptionSimple(ice.TCP_DOMAIN),
		cli.CTX_OPS, HostPort(m, tcp.LOCALHOST, m.Cmdv(SERVE, tcp.PORT)), cli.CTX_LOG, ice.VAR_LOG_BOOT_LOG, cli.CTX_PID, ice.VAR_LOG_ICE_PID,
		cli.CTX_ROOT, kit.Path(""), cli.PATH, cli.BinPath(p, ""), cli.USER, ice.Info.Username,
	)...), cli.CMD_OUTPUT, path.Join(p, ice.VAR_LOG_BOOT_LOG), mdb.CACHE_CLEAR_ONEXIT, ice.TRUE)
	if !nfs.Exists(m, p) {
		kit.If(m.Option(nfs.BINARY) == "" && !cli.SystemFindGo(m), func(p string) { m.Option(nfs.BINARY, S(name)) })
		kit.If(m.Option(nfs.BINARY), func(p string) { _dream_binary(m, p) })
		kit.If(m.Option(nfs.TEMPLATE), func(p string) { _dream_template(m, p) })
	}
	bin := kit.Select(kit.Path(os.Args[0]), cli.SystemFind(m, ice.ICE_BIN, nfs.PWD+path.Join(p, ice.BIN), nfs.PWD+ice.BIN))
	if cli.IsSuccess(m.Cmd(cli.DAEMON, bin, SPACE, tcp.DIAL, ice.DEV, ice.OPS, mdb.TYPE, WORKER, m.OptionSimple(mdb.NAME), cli.DAEMON, ice.OPS)) {
		gdb.WaitEvent(m, DREAM_OPEN, func(m *ice.Message, arg ...string) bool { return m.Option(mdb.NAME) == name })
		m.Sleep300ms()
	}
}
func _dream_check(m *ice.Message, name string) string {
	p := path.Join(ice.USR_LOCAL_WORK, name)
	if p := path.Join(p, ice.VAR_LOG_ICE_PID); nfs.Exists(m, p) {
		if pid := m.Cmdx(nfs.CAT, p, kit.Dict(ice.MSG_USERROLE, aaa.TECH)); pid != "" && nfs.Exists(m, "/proc/"+pid) {
			m.Info("already exists %v", pid)
			return ""
		} else {
			for i := 0; i < 3; i++ {
				if m.Cmd(SPACE, name).Length() > 0 {
					m.Info("already exists %v", name)
					return ""
				}
				m.Sleep300ms()
			}
		}
	}
	return p
}
func _dream_binary(m *ice.Message, p string) {
	if bin := path.Join(m.Option(cli.CMD_DIR), ice.BIN_ICE_BIN); nfs.Exists(m, bin) {
		return
	} else if kit.IsUrl(p) || strings.HasPrefix(p, S()) {
		m.Cmd(DREAM, DOWNLOAD, bin, p)
	} else {
		m.Cmd(nfs.LINK, bin, kit.Path(p))
	}
}
func _dream_template(m *ice.Message, p string) {
	kit.For([]string{
		ice.LICENSE, ice.README_MD, ice.MAKEFILE, ice.GO_MOD, ice.GO_SUM,
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
	FOR_EACH = "forEach"
	FOR_FLOW = "forFlow"

	DREAM_INPUTS = "dream.inputs"
	DREAM_CREATE = "dream.create"
	DREAM_REMOVE = "dream.remove"
	DREAM_TRASH  = "dream.trash"
	DREAM_START  = "dream.start"
	DREAM_STOP   = "dream.stop"
	DREAM_OPEN   = "dream.open"
	DREAM_CLOSE  = "dream.close"

	DREAM_TABLES = "dream.tables"
	DREAM_ACTION = "dream.action"
)
const DREAM = "dream"

func init() {
	Index.MergeCommands(ice.Commands{
		DREAM: {Name: "dream refresh", Help: "梦想家", Icon: "Launchpad.png", Role: aaa.VOID, Meta: kit.Dict(
			ice.CTX_TRANS, kit.Dict(html.INPUT, kit.Dict(WORKER, "空间", SERVER, "机器", ORIGIN, "主机")),
		), Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m = m.Spawn()
				m.GoSleep(cli.TIME_1s, func() {
					for _, cmd := range kit.Reverse(kit.Split(mdb.Config(m, html.BUTTON))) {
						m.Cmd(gdb.EVENT, gdb.LISTEN, gdb.EVENT, DREAM_TABLES, ice.CMD, cmd)
						m.Cmd(gdb.EVENT, gdb.LISTEN, gdb.EVENT, DREAM_ACTION, ice.CMD, cmd)
						aaa.White(m, kit.Keys(m.ShortKey(), ctx.ACTION, cmd))
					}
				})
				m.GoSleep(cli.TIME_3s, func() {
					mdb.HashSelects(m).Table(func(value ice.Maps) {
						if value[cli.RESTART] == ALWAYS && nfs.Exists(m, path.Join(ice.USR_LOCAL_WORK+value[mdb.NAME])) {
							m.Cmd(DREAM, cli.START, kit.Dict(mdb.NAME, value[mdb.NAME]))
						}
					})
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
					case mdb.ICONS:
						mdb.HashInputs(m, arg)
						return
					case nfs.BINARY:
						m.Cmdy(nfs.DIR, ice.BIN, "path,size,time", kit.Dict(nfs.DIR_TYPE, nfs.TYPE_BIN))
						m.Cmd(nfs.DIR, ice.USR_LOCAL_WORK, kit.Dict(nfs.DIR_TYPE, nfs.TYPE_BOTH), func(value ice.Maps) {
							m.Cmdy(nfs.DIR, path.Join(value[nfs.PATH], ice.BIN), "path,size,time", kit.Dict(nfs.DIR_TYPE, nfs.TYPE_BIN))
						})
						m.RenameAppend(nfs.PATH, arg[0])
						// mdb.HashInputs(m, arg)
						DreamListSpide(m, []string{ice.DEV}, ORIGIN, func(dev, origin string) {
							m.Spawn().SplitIndex(m.Cmdx(SPIDE, dev, SPIDE_RAW, http.MethodGet, S(), cli.GOOS, runtime.GOOS, cli.GOARCH, runtime.GOARCH)).Table(func(value ice.Maps) {
								m.Push(arg[0], origin+S(value[mdb.NAME])).Push(nfs.SIZE, value[nfs.SIZE]).Push(mdb.TIME, value[mdb.TIME])
							})
						})
					}
				case STARTALL:
					DreamEach(m, "", cli.STOP, func(name string) { m.Push(arg[0], name) })
					return
				case tcp.SEND:
					m.Cmd(SPACE, func(value ice.Maps) {
						kit.If(kit.IsIn(value[mdb.TYPE], SERVER), func() { m.Push(arg[0], value[mdb.NAME]) })
					})
					return
				}
				switch arg[0] {
				case mdb.NAME:
					DreamEach(m, "", cli.START, func(name string) { m.Push(arg[0], name) })
				case ctx.CMDS:
					m.Cmdy(ctx.COMMAND)
				case nfs.FILE:
					m.Options(nfs.DIR_TYPE, nfs.TYPE_CAT, ice.MSG_FIELDS, nfs.PATH)
					m.Cmdy(nfs.DIR, nfs.SRC).Cmdy(nfs.DIR, nfs.ETC).Cmdy(nfs.DIR, "")
				default:
					gdb.Event(m, DREAM_INPUTS, arg)
				}
			}},
			mdb.CREATE: {Name: "create name*=hi icons repos binary template", Hand: func(m *ice.Message, arg ...string) {
				kit.If(!strings.Contains(m.Option(mdb.NAME), "-") || !strings.HasPrefix(m.Option(mdb.NAME), "20"), func() { m.Option(mdb.NAME, m.Time("20060102-")+m.Option(mdb.NAME)) })
				kit.If(mdb.Config(m, nfs.BINARY), func(p string) { m.OptionDefault(nfs.BINARY, p+m.Option(mdb.NAME)) })
				kit.If(mdb.Config(m, nfs.REPOS), func(p string) { m.OptionDefault(nfs.REPOS, p+m.Option(mdb.NAME)) })
				m.Option(nfs.REPOS, kit.Select("", kit.Split(m.Option(nfs.REPOS)), -1))
				m.OptionDefault(mdb.ICONS, nfs.USR_ICONS_CONTEXTS)
				if mdb.HashCreate(m); ice.Info.Important == true {
					_dream_start(m, m.Option(mdb.NAME))
					StreamPushRefreshConfirm(m, m.Trans("refresh for new space ", "刷新列表查看新空间 ")+m.Option(mdb.NAME))
				}
			}},
			mdb.REMOVE: {Hand: func(m *ice.Message, arg ...string) {
				gdb.Event(m, DREAM_REMOVE, m.OptionSimple(mdb.NAME))
				mdb.HashRemove(m)
			}},
			DOWNLOAD: {Name: "download path link", Hand: func(m *ice.Message, arg ...string) {
				GoToast(m, func(toast func(string, int, int)) []string {
					SpideSave(m, m.Option(nfs.PATH), kit.MergeURL(m.Option(mdb.LINK), cli.GOOS, runtime.GOOS, cli.GOARCH, runtime.GOARCH), func(count, total, value int) {
						toast(m.Option(mdb.NAME), count, total)
					})
					return nil
				})
				os.Chmod(m.Option(nfs.PATH), ice.MOD_DIR)
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
				compile := cli.SystemFindGo(m)
				m.Option(ice.MSG_TITLE, kit.Keys(m.Option(ice.MSG_USERPOD0), m.Option(ice.MSG_USERPOD), m.CommandKey(), m.ActionKey()))
				m.Cmd("", FOR_FLOW, m.Option(mdb.NAME), kit.JoinWord(cli.SH, ice.ETC_MISS_SH), func(p string) bool {
					if compile && nfs.Exists(m, path.Join(p, ice.SRC_MAIN_GO)) {
						return false
					} else {
						m.Cmd(SPACE, path.Base(p), cli.RUNTIME, UPGRADE)
						return true
					}
				})
				kit.If(m.Option(mdb.NAME) == "", func() { m.Sleep("5s").Cmdy(ROUTE, cli.BUILD).ProcessInner() })
			}},
			PUBLISH: {Name: "publish name", Hand: func(m *ice.Message, arg ...string) {
				m.Option(ice.MSG_TITLE, kit.Keys(m.Option(ice.MSG_USERPOD0), m.Option(ice.MSG_USERPOD), m.CommandKey(), m.ActionKey()))
				list := []string{cli.LINUX, cli.DARWIN, cli.WINDOWS}
				msg := m.Spawn(ice.Maps{ice.MSG_DAEMON: ""})
				func() {
					defer ToastProcess(m, PUBLISH, ice.Info.Pathname)()
					m.Cmd(AUTOGEN, BINPACK)
					kit.For(list, func(goos string) {
						PushNoticeRich(m, mdb.NAME, ice.Info.NodeName, msg.Cmd(COMPILE, goos, cli.AMD64).AppendSimple())
					})
				}()
				DreamEach(m, m.Option(mdb.NAME), "", func(name string) {
					m.Cmd(SPACE, name, AUTOGEN, BINPACK)
					kit.For(list, func(goos string) {
						PushNoticeRich(m.Options(ice.MSG_COUNT, "0", ice.LOG_DISABLE, ice.TRUE), mdb.NAME, name, msg.Cmd(SPACE, name, COMPILE, goos, cli.AMD64, kit.Dict(ice.MSG_USERPOD, name)).AppendSimple())
					})
				})
				m.ProcessHold()
			}},
			VERSION: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy("web.code.version") }},
			FOR_FLOW: {Name: "forFlow name cmd*='sh etc/miss.sh'", Help: "流程", Icon: "bi bi-terminal", Hand: func(m *ice.Message, arg ...string) {
				m.Options(ctx.DISPLAY, html.PLUGIN_XTERM, cli.CMD_OUTPUT, nfs.NewWriteCloser(func(buf []byte) (int, error) {
					PushNoticeGrow(m.Options(ice.MSG_COUNT, "0", ice.LOG_DEBUG, ice.FALSE, ice.LOG_DISABLE, ice.TRUE), strings.ReplaceAll(string(buf), lex.NL, "\r\n"))
					return len(buf), nil
				}, nil))
				msg := m.Spawn(ice.Maps{ice.MSG_DEBUG: ice.FALSE})
				DreamEach(m, m.Option(mdb.NAME), "", func(name string) {
					p := path.Join(ice.USR_LOCAL_WORK, name)
					if cb, ok := m.OptionCB("").(func(string) bool); ok && cb(p) {
						return
					}
					defer PushNoticeGrow(msg, "\r\n\r\n\r\n")
					PushNoticeGrow(msg, kit.Format("[%s]%s$ %s\r\n", time.Now().Format(ice.MOD_TIME_ONLY), name, m.Option(ice.CMD)))
					m.Cmd(cli.SYSTEM, kit.Split(m.Option(ice.CMD)), kit.Dict(cli.CMD_DIR, p)).Sleep300ms()
				})
			}},
			ctx.CMDS: {Name: "cmds name cmds*", Help: "命令", Icon: "bi bi-terminal", Hand: func(m *ice.Message, arg ...string) {
				DreamEach(m, m.Option(mdb.NAME), "", func(name string) {
					m.Push(mdb.NAME, name).Push(mdb.TEXT, m.Cmdx(SPACE, name, kit.Split(m.Option(ctx.CMDS))))
				}).StatusTimeCount(m.OptionSimple(ctx.CMDS))
			}},
			nfs.FILE: {Name: "file name file*", Help: "文件", Icon: "bi bi-file-earmark-code", Hand: func(m *ice.Message, arg ...string) {
				DreamEach(m, m.Option(mdb.NAME), "", func(name string) {
					m.Push(mdb.NAME, name).Push(mdb.TEXT, m.Cmdx(SPACE, name, nfs.CAT, m.Option(nfs.FILE)))
				}).StatusTimeCount(m.OptionSimple(nfs.FILE))
			}},
			cli.START: {Hand: func(m *ice.Message, arg ...string) {
				_dream_start(m, m.Option(mdb.NAME))
				gdb.Event(m, DREAM_START, arg)
			}},
			cli.STOP: {Hand: func(m *ice.Message, arg ...string) {
				defer ToastProcess(m)()
				gdb.Event(m, DREAM_STOP, arg)
				m.Cmd(SPACE, mdb.MODIFY, m.OptionSimple(mdb.NAME), mdb.STATUS, cli.STOP)
				m.Cmd(SPACE, m.Option(mdb.NAME), ice.EXIT).Sleep3s()
			}},
			cli.RUNTIME: {Hand: func(m *ice.Message, arg ...string) {
				ProcessPodCmd(m, m.Option(mdb.NAME), "", nil, arg...)
			}},
			tcp.SEND: {Name: "send space*", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(SPACE, m.Option(SPACE), DREAM, mdb.CREATE, m.OptionSimple(mdb.NAME, mdb.ICONS, nfs.REPOS, nfs.BINARY))
				m.Cmd(SPACE, m.Option(SPACE), DREAM, cli.START, m.OptionSimple(mdb.NAME))
				ProcessIframe(m, "", m.MergePod(kit.Keys(m.Option(SPACE), m.Option(mdb.NAME))))
			}},
			nfs.TRASH: {Hand: func(m *ice.Message, arg ...string) {
				gdb.Event(m, DREAM_TRASH, arg)
				nfs.Trash(m, path.Join(ice.USR_LOCAL_WORK, m.Option(mdb.NAME)))
			}},
			OPEN: {Role: aaa.VOID, Hand: func(m *ice.Message, arg ...string) {
				if m.Option(mdb.TYPE) == ORIGIN && m.IsLocalhost() {
					m.ProcessOpen(SpideOrigin(m, m.Option(mdb.NAME)))
				} else {
					m.ProcessOpen(S(m.Option(mdb.NAME)))
				}
			}},
			GRANT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(CHAT_GRANT, aaa.CONFIRM, kit.Dict(SPACE, m.Option(mdb.NAME)))
			}},
			DREAM_OPEN: {Hand: func(m *ice.Message, arg ...string) {}},
			DREAM_CLOSE: {Hand: func(m *ice.Message, arg ...string) {
				kit.For(arg, func(k, v string) {
					if k == cli.DAEMON && v == ice.OPS && m.Cmdv(SPACE, m.Option(mdb.NAME), mdb.STATUS) != cli.STOP {
						m.GoSleep300ms(func() { m.Cmd(DREAM, cli.START, m.OptionSimple(mdb.NAME)) })
					}
				})
			}},
			TOKEN: {Hand: func(m *ice.Message, arg ...string) {
				m.Options(m.Cmd(SPIDE, m.Option(mdb.NAME)).AppendSimple()).Cmdy(SPIDE, mdb.DEV_REQUEST)
			}},
			DREAM_TABLES: {Hand: func(m *ice.Message, arg ...string) {
				if !aaa.IsTechOrRoot(m) {
					m.PushButton(OPEN)
					return
				}
				list := []ice.Any{}
				kit.If(m.IsDebug(), func() { list = append(list, cli.RUNTIME) })
				switch m.Option(mdb.TYPE) {
				case WORKER:
					list = append(list, "settings", tcp.SEND, OPEN)
				default:
					list = append(list, TOKEN, DREAM, OPEN)
				}
				m.PushButton(list...)
			}},
			"settings": {Name: "settings restart=manual,always access=public,private", Help: "设置", Hand: func(m *ice.Message, arg ...string) {
				kit.If(m.Option(cli.RESTART) == "manual", func() { m.Option(cli.RESTART, "") })
				kit.If(m.Option(aaa.ACCESS) == aaa.PUBLIC, func() { m.Option(aaa.ACCESS, "") })
				mdb.HashModify(m, m.OptionSimple(mdb.NAME, cli.RESTART, aaa.ACCESS))
			}},
			STATS_TABLES: {Hand: func(m *ice.Message, arg ...string) {
				if msg := _dream_list(m.Spawn(), true); msg.Length() > 0 {
					stat := map[string]int{}
					msg.Table(func(value ice.Maps) { stat[value[mdb.TYPE]]++; stat[value[mdb.STATUS]]++ })
					PushStats(m, kit.Keys(m.CommandKey(), cli.START), stat[cli.START], "", "已启动空间")
					PushStats(m, kit.Keys(m.CommandKey(), SERVER), stat[SERVER], "", "已连接机器")
					PushStats(m, kit.Keys(m.CommandKey(), ORIGIN), stat[ORIGIN], "", "已连接主机")
				}
			}},
		}, StatsAction(), DreamAction(), DreamTablesAction(), mdb.ImportantHashAction(
			mdb.SHORT, mdb.NAME, mdb.FIELD, "time,name,icons,repos,binary,template,restart,access",
			html.BUTTON, kit.JoinWord(PORTAL, DESKTOP, MESSAGE, ADMIN, WORD, STATUS, VIMER, COMPILE, XTERM, DREAM),
			ctx.TOOLS, kit.Simple(SPIDE, ROUTE), ONLINE, ice.TRUE,
		)), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				simple := m.Option(ice.DREAM_SIMPLE) == ice.TRUE
				if ice.Info.NodeType != WORKER {
					_dream_list(m, simple)
					_dream_list_icon(m)
					if m.Length() == 0 {
						m.EchoInfoButton(m.Trans("please create new dream", "请创建新空间"), mdb.CREATE).Action(mdb.CREATE)
						return
					}
				}
				if !m.IsCliUA() && aaa.IsTechOrRoot(m) {
					_dream_list_more(m, simple)
				}
				if ice.Info.NodeType == WORKER || !aaa.IsTechOrRoot(m) || m.IsCliUA() {
					m.Action()
				} else if m.IsDebug() && cli.SystemFindGo(m) {
					m.Action(html.FILTER, mdb.CREATE, STARTALL, STOPALL, cli.BUILD, PUBLISH)
				} else {
					m.Action(html.FILTER, mdb.CREATE, STARTALL, STOPALL)
				}
				m.Sort("type,status,name", []string{aaa.LOGIN, WORKER, SERVER, ORIGIN}, []string{cli.START, cli.STOP, cli.BEGIN}, ice.STR_R)
				m.StatusTimeCountStats(mdb.TYPE, mdb.STATUS)
				ctx.DisplayTableCard(m)
				if !m.IsDebug() {
					m.Options(ice.MSG_TOOLKIT, "")
				}
				if !aaa.IsTechOrRoot(m) {
					m.Options(ice.MSG_TOOLKIT, "")
					m.Option(ice.MSG_ONLINE, ice.FALSE)
				}
			} else if arg[0] == ctx.ACTION {
				gdb.Event(m, DREAM_ACTION, arg)
			} else {
				mdb.HashSelects(m, arg[0]).PushAction(PORTAL, DESKTOP, ADMIN, OPEN, mdb.REMOVE)
			}
		}},
	})
}

func DreamTablesAction(arg ...string) ice.Actions {
	return ice.Actions{ice.CTX_INIT: {Hand: DreamWhiteHandle},
		DREAM_TABLES: {Hand: func(m *ice.Message, _ ...string) {
			m.PushButton(kit.Dict(m.CommandKey(), kit.Select(m.Commands("").Help, arg, 0)))
		}},
		DREAM_ACTION: {Hand: func(m *ice.Message, arg ...string) { DreamProcess(m, "", nil, arg...) }},
	}
}
func DreamAction() ice.Actions {
	return gdb.EventsAction(DREAM_INPUTS, DREAM_CREATE, DREAM_REMOVE, DREAM_TRASH, DREAM_OPEN, DREAM_CLOSE, SPACE_LOGIN, SERVE_START)
}
func DreamWhiteHandle(m *ice.Message, arg ...string) {
	aaa.White(m, kit.Keys(DREAM, ctx.ACTION, m.ShortKey()))
	aaa.White(m, kit.Keys(m.ShortKey(), ctx.ACTION, DREAM_ACTION))
}
func DreamProcessIframe(m *ice.Message, arg ...string) {
	if !kit.HasPrefixList(arg, ctx.ACTION, m.ShortKey()) && !kit.HasPrefixList(arg, ctx.ACTION, m.CommandKey()) {
		return
	}
	if len(arg) == 2 {
		defer m.Push(TITLE, kit.Keys(m.Option(mdb.NAME), m.ShortKey()))
	}
	DreamProcess(m, CHAT_IFRAME, func() string {
		return kit.MergeURL(S(kit.Keys(m.Option(ice.MSG_USERPOD), m.Option(mdb.NAME)))+C(m.ShortKey()), ice.MSG_DEBUG, m.Option(ice.MSG_DEBUG))
	}, arg...)
}
func DreamProcess(m *ice.Message, cmd string, args ice.Any, arg ...string) {
	if !kit.HasPrefixList(arg, ctx.ACTION, m.ShortKey()) && !kit.HasPrefixList(arg, ctx.ACTION, m.CommandKey()) {
		return
	} else if arg = arg[2:]; len(arg) == 0 {
		arg = append(arg, m.Option(mdb.NAME))
		defer m.ProcessField(ctx.ACTION, m.ShortKey(), arg[0], ctx.RUN)
		defer processSpace(m, arg[0], arg[0], m.ShortKey())
	}
	ctx.ProcessFloat(m.Options(ice.POD, arg[0]), kit.Select(m.ShortKey(), cmd), args, arg[1:]...)
}
func DreamEach(m *ice.Message, name string, status string, cb func(string)) *ice.Message {
	reg, err := regexp.Compile(name)
	if m.WarnNotValid(err) {
		return m
	}
	msg := m.Spawn()
	m.Cmds(DREAM, kit.Dict(ice.DREAM_SIMPLE, ice.TRUE)).Table(func(value ice.Maps) {
		if value[mdb.STATUS] == kit.Select(cli.START, status) && value[mdb.TYPE] == WORKER && (value[mdb.NAME] == name || reg.MatchString(kit.Format("%s:%s=%s@%d", value[mdb.NAME], value[mdb.TYPE], value[nfs.MODULE], value[nfs.VERSION]))) {
			msg.Push(mdb.NAME, value[mdb.NAME])
		}
	})
	return GoToastTable(msg, mdb.NAME, func(value ice.Maps) { cb(value[mdb.NAME]) })
}
func DreamListSpide(m *ice.Message, list []string, types string, cb func(dev, origin string)) {
	msg := m.Spawn()
	kit.For(list, func(name string) { msg.Push(mdb.NAME, name) })
	m.Cmds(SPACE).Table(func(value ice.Maps) { kit.If(value[mdb.TYPE] == types, func() { msg.Push(mdb.NAME, value[mdb.NAME]) }) })
	has := map[string]bool{}
	GoToastTable(msg, mdb.NAME, func(value ice.Maps) {
		origin := SpideOrigin(m, value[mdb.NAME])
		kit.If(!has[origin], func() { has[origin] = true; cb(value[mdb.NAME], origin) })
	})
}
func DreamList(m *ice.Message) *ice.Message {
	return AdminCmd(m.Options(ice.DREAM_SIMPLE, ice.TRUE), DREAM)
}
func DreamStat(m *ice.Message, name string) (res string) {
	if cli.SystemFindGit(m) {
		text := []string{}
		for _, line := range kit.Split(m.Cmdx(cli.SYSTEM, cli.GIT, "diff", "--shortstat", kit.Dict(cli.CMD_DIR, path.Join(ice.USR_LOCAL_WORK, name))), mdb.FS, mdb.FS) {
			if list := kit.Split(line); strings.Contains(line, nfs.FILE) {
				text = append(text, kit.Format("<span class='files'>%s file</span>", list[0]))
			} else if strings.Contains(line, "ins") {
				text = append(text, kit.Format("<span class='add'>%s+++</span>", list[0]))
			} else if strings.Contains(line, "del") {
				text = append(text, kit.Format("<span class='del'>%s---</span>", list[0]))
			}
		}
		res = strings.Join(text, "")
	}
	return
}
