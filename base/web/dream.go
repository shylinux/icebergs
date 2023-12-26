package web

import (
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
	kit "shylinux.com/x/toolkits"
)

func _dream_list(m *ice.Message) *ice.Message {
	stats := map[string]int{}
	list := m.CmdMap(SPACE, mdb.NAME)
	mdb.HashSelect(m).Table(func(value ice.Maps) {
		if space, ok := list[value[mdb.NAME]]; ok {
			msg := gdb.Event(m.Spawn(value, space), DREAM_TABLES).Copy(m.Spawn().PushButton(cli.STOP))
			m.Push(nfs.VERSION, space[nfs.VERSION])
			m.Push(mdb.TYPE, space[mdb.TYPE])
			m.Push(cli.STATUS, cli.START)
			m.Push(mdb.TEXT, msg.Append(mdb.TEXT))
			m.PushButton(strings.Join(msg.Appendv(ctx.ACTION), ""))
			stats[cli.START]++
		} else {
			m.Push(nfs.VERSION, "")
			m.Push(mdb.TYPE, WORKER)
			if nfs.Exists(m, path.Join(ice.USR_LOCAL_WORK, value[mdb.NAME])) {
				m.Push(cli.STATUS, cli.STOP)
			} else {
				m.Push(cli.STATUS, cli.BEGIN)
			}
			m.Push(mdb.TEXT, "")
			if nfs.Exists(m, path.Join(ice.USR_LOCAL_WORK, value[mdb.NAME])) {
				m.PushButton(cli.START, nfs.TRASH)
				stats[cli.STOP]++
			} else {
				m.PushButton(cli.START, mdb.REMOVE)
				stats[ice.INIT]++
			}
		}
	})
	return m.Sort("status,type,name", []string{cli.START, cli.STOP, cli.BEGIN}, ice.STR, ice.STR_R).StatusTimeCount()
}
func _dream_start(m *ice.Message, name string) {
	if m.Warn(name == "", ice.ErrNotValid, mdb.NAME) {
		return
	}
	defer m.ProcessOpen(m.MergePod(m.Option(mdb.NAME, name)))
	p := path.Join(ice.USR_LOCAL_WORK, name)
	if p := path.Join(p, ice.Info.PidPath); nfs.Exists(m, p) {
		if pid := m.Cmdx(nfs.CAT, p, kit.Dict(ice.MSG_USERROLE, aaa.TECH)); pid != "" && nfs.Exists(m, "/proc/"+pid) {
			m.Info("already exists %v", pid)
			return
		}
	}
	for i := 0; i < 3; i++ {
		if msg := m.Cmd(SPACE, name); msg.Length() > 0 {
			m.Info("already exists %v", name)
			return
		}
		m.Sleep300ms()
	}
	defer ToastProcess(m)()
	defer m.Options(cli.CMD_DIR, "", cli.CMD_ENV, "", cli.CMD_OUTPUT, "")
	m.Options(cli.CMD_DIR, kit.Path(p), cli.CMD_ENV, kit.EnvList(kit.Simple(m.OptionSimple("tcp_domain"),
		cli.CTX_OPS, Domain(tcp.LOCALHOST, m.Cmdv(SERVE, tcp.PORT)), cli.CTX_LOG, ice.VAR_LOG_BOOT_LOG, cli.CTX_PID, ice.VAR_LOG_ICE_PID,
		cli.CTX_ROOT, kit.Path(""), cli.PATH, cli.BinPath(p, ""), cli.USER, ice.Info.Username,
	)...), cli.CMD_OUTPUT, path.Join(p, ice.VAR_LOG_BOOT_LOG), mdb.CACHE_CLEAR_ONEXIT, ice.TRUE)
	gdb.Event(m, DREAM_CREATE, m.OptionSimple(mdb.NAME, mdb.TYPE))
	kit.If(m.Option(nfs.BINARY), func(p string) { _dream_binary(m, p) })
	kit.If(m.Option(nfs.TEMPLATE), func(p string) { _dream_template(m, p) })
	bin := kit.Select(kit.Path(os.Args[0]), cli.SystemFind(m, ice.ICE_BIN, nfs.PWD+path.Join(p, ice.BIN), nfs.PWD+ice.BIN))
	if bin != kit.Path(ice.BIN_ICE_BIN) && strings.Count(m.Cmdx(cli.SYSTEM, "sh", "-c", "ps aux | grep "+bin+" | grep -v grep"), bin) > 0 {
		return
	}
	m.Cmd(cli.DAEMON, bin, SPACE, tcp.DIAL, ice.DEV, ice.OPS, mdb.TYPE, WORKER, m.OptionSimple(mdb.NAME), cli.DAEMON, ice.OPS)
	gdb.WaitEvent(m, DREAM_OPEN, func(m *ice.Message, arg ...string) bool { return m.Option(mdb.NAME) == name })
}
func _dream_binary(m *ice.Message, p string) {
	if bin := path.Join(m.Option(cli.CMD_DIR), ice.BIN_ICE_BIN); nfs.Exists(m, bin) {
		return
	} else if kit.IsUrl(p) {
		GoToast(m, DOWNLOAD, func(toast func(string, int, int)) (list []string) {
			begin := time.Now()
			SpideSave(m, bin, kit.MergeURL(p, cli.GOOS, runtime.GOOS, cli.GOARCH, runtime.GOARCH), func(count, total, value int) {
				cost := time.Now().Sub(begin)
				toast(kit.FormatShow(nfs.FROM, begin.Format("15:04:05"), cli.COST, kit.FmtDuration(cost), cli.REST, kit.FmtDuration(cost*time.Duration(101)/time.Duration(value+1)-cost)), count, total)
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
		ice.SRC_MAIN_SHY, ice.SRC_MAIN_SH, ice.SRC_MAIN_GO, ice.SRC_MAIN_JS,
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
		DREAM: {Name: "dream name@key auto create repos startall stopall publish cmd cat", Help: "梦想家", Icon: "Launchpad.png", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m = m.Spawn()
				m.GoSleep("10s", func() {
					mdb.HashSelects(m).Table(func(value ice.Maps) {
						if value[cli.RESTART] == "always" && nfs.Exists(m, path.Join(ice.USR_LOCAL_WORK+value[mdb.NAME])) {
							m.Cmd(DREAM, cli.START, kit.Dict(mdb.NAME, value[mdb.NAME]))
						}
					})
				})
			}},
			ctx.CONFIG: {Hand: func(m *ice.Message, arg ...string) {
				for _, cmd := range kit.Reverse(arg) {
					m.Cmd(gdb.EVENT, gdb.LISTEN, gdb.EVENT, DREAM_TABLES, ice.CMD, cmd)
					m.Cmd(gdb.EVENT, gdb.LISTEN, gdb.EVENT, DREAM_ACTION, ice.CMD, cmd)
				}
			}},
			mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
				if mdb.IsSearchPreview(m, arg) {
					m.Cmds("", func(value ice.Maps) { m.PushSearch(mdb.TEXT, m.MergePod(value[mdb.NAME]), value) })
				}
			}},
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch m.Option(ctx.ACTION) {
				case mdb.CREATE:
					switch arg[0] {
					case mdb.NAME, nfs.TEMPLATE:
						_dream_list(m).Cut("name,status,time")
						return
					case mdb.ICONS:
						mdb.HashInputs(m, arg)
						return
					}
				case "startall":
					DreamEach(m, "", cli.STOP, func(name string) { m.Push(arg[0], name) })
					return
				case ice.MAIN:
					m.Cmdy(SPACE, m.Option(mdb.NAME), SPACE, mdb.INPUTS, arg)
					return
				}
				switch arg[0] {
				case mdb.NAME:
					DreamEach(m, "", cli.START, func(name string) { m.Push(arg[0], name) })
				case nfs.BINARY:
					m.Cmdy(nfs.DIR, ice.BIN, "path,size,time", kit.Dict(nfs.DIR_TYPE, nfs.TYPE_BIN))
					m.Cmd(nfs.DIR, ice.USR_LOCAL_WORK, kit.Dict(nfs.DIR_TYPE, nfs.TYPE_BOTH), func(value ice.Maps) {
						m.Cmdy(nfs.DIR, path.Join(value[nfs.PATH], ice.BIN), "path,size,time", kit.Dict(nfs.DIR_TYPE, nfs.TYPE_BIN))
					})
					m.RenameAppend(nfs.PATH, arg[0])
					mdb.HashInputs(m, arg)
				case ice.CMD:
					m.Cmdy(ctx.COMMAND)
				case nfs.FILE:
					m.Options(nfs.DIR_TYPE, nfs.TYPE_CAT, ice.MSG_FIELDS, nfs.PATH)
					m.Cmdy(nfs.DIR, nfs.SRC).Cmdy(nfs.DIR, nfs.ETC).Cmdy(nfs.DIR, "")
				default:
					gdb.Event(m, DREAM_INPUTS, arg)
				}
			}},
			mdb.CREATE: {Name: "create name*=hi icon@icons repos binary template", Hand: func(m *ice.Message, arg ...string) {
				kit.If(!strings.Contains(m.Option(mdb.NAME), "-") || !strings.HasPrefix(m.Option(mdb.NAME), "20"), func() { m.Option(mdb.NAME, m.Time("20060102-")+m.Option(mdb.NAME)) })
				kit.If(mdb.Config(m, nfs.BINARY), func(p string) { m.OptionDefault(nfs.BINARY, p+m.Option(mdb.NAME)) })
				kit.If(mdb.Config(m, nfs.REPOS), func(p string) { m.OptionDefault(nfs.REPOS, p+m.Option(mdb.NAME)) })
				m.Option(nfs.REPOS, kit.Select("", kit.Slice(kit.Split(m.Option(nfs.REPOS)), -1), 0))
				m.OptionDefault(mdb.ICON, nfs.USR_ICONS_ICEBERGS)
				if mdb.HashCreate(m); !m.IsCliUA() {
					_dream_start(m, m.Option(mdb.NAME))
				}
			}},
			nfs.REPOS: {Help: "仓库", Icon: "bi bi-git", Hand: func(m *ice.Message, arg ...string) {
				m.ProcessOpen(m.MergePodCmd("", CODE_GIT_SEARCH, nfs.REPOS, nfs.REPOS))
			}},
			"startall": {Name: "startall name", Help: "启动", Icon: "bi bi-play-circle", Hand: func(m *ice.Message, arg ...string) {
				DreamEach(m, m.Option(mdb.NAME), cli.STOP, func(name string) {
					m.Cmd("", cli.START, ice.Maps{mdb.NAME: name, ice.MSG_DAEMON: ""})
				})
			}},
			"stopall": {Name: "stopall name", Help: "停止", Icon: "bi bi-stop-circle", Hand: func(m *ice.Message, arg ...string) {
				DreamEach(m, m.Option(mdb.NAME), "", func(name string) {
					m.Cmd("", cli.STOP, ice.Maps{mdb.NAME: name, ice.MSG_DAEMON: ""})
				})
			}},
			"publish": {Name: "publish name", Help: "发布", Icon: "bi bi-send-check", Hand: func(m *ice.Message, arg ...string) {
				DreamEach(m, m.Option(mdb.NAME), "", func(name string) {
					m.Push(mdb.NAME, name).Push(mdb.TEXT, m.Cmdx(SPACE, name, "compile", cli.LINUX))
					m.Push(mdb.NAME, name).Push(mdb.TEXT, m.Cmdx(SPACE, name, "compile", cli.DARWIN))
					m.Push(mdb.NAME, name).Push(mdb.TEXT, m.Cmdx(SPACE, name, "compile", cli.WINDOWS))
				})
			}},
			ice.CMD: {Name: "cmd name cmd*", Help: "命令", Icon: "bi bi-terminal", Hand: func(m *ice.Message, arg ...string) {
				DreamEach(m, m.Option(mdb.NAME), "", func(name string) {
					m.Push(mdb.NAME, name).Push(mdb.TEXT, m.Cmdx(SPACE, name, kit.Split(m.Option(ice.CMD))))
				}).StatusTimeCount(ice.CMD, m.Option(ice.CMD))
			}},
			nfs.CAT: {Name: "cat name file*", Help: "文件", Icon: "bi bi-file-earmark-code", Hand: func(m *ice.Message, arg ...string) {
				DreamEach(m, m.Option(mdb.NAME), "", func(name string) {
					m.Push(mdb.NAME, name).Push(mdb.TEXT, m.Cmdx(SPACE, name, nfs.CAT, m.Option(nfs.FILE)))
				}).StatusTimeCount(nfs.FILE, m.Option(nfs.FILE))
			}},
			cli.START: {Hand: func(m *ice.Message, arg ...string) {
				gdb.Event(m, DREAM_START, arg)
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
			OPEN: {Hand: func(m *ice.Message, arg ...string) { m.ProcessOpen(m.MergePod(m.Option(mdb.NAME))) }},
			MAIN: {Name: "main index", Help: "首页", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(SPACE, m.Option(mdb.NAME), SPACE, ice.MAIN, m.Option(ctx.INDEX))
			}},
			DREAM_CLOSE: {Hand: func(m *ice.Message, arg ...string) {
				if m.Option(cli.DAEMON) == ice.OPS && m.Cmdv(SPACE, m.Option(mdb.NAME), mdb.STATUS) != cli.STOP {
					m.GoSleep300ms(func() { m.Cmd(DREAM, cli.START, m.OptionSimple(mdb.NAME)) })
				}
			}},
			DREAM_TABLES: {Hand: func(m *ice.Message, arg ...string) {
				kit.Switch(m.Option(mdb.TYPE), []string{WORKER, SERVER}, func() { m.PushButton(OPEN, ice.MAIN) })
			}},
			DREAM_OPEN: {Hand: func(m *ice.Message, arg ...string) {
			}},
			STATS_TABLES: {Hand: func(m *ice.Message, arg ...string) {
				if msg := mdb.HashSelects(m.Spawn()); msg.Length() > 0 {
					stats := map[string]int{}
					list := m.CmdMap(SPACE, mdb.NAME)
					msg.Table(func(value ice.Maps) {
						if _, ok := list[value[mdb.NAME]]; ok {
							stats[cli.START]++
						}
					})
					PushStats(m, kit.Keys(m.CommandKey(), cli.START), stats[cli.START], "", "空间总数")
					PushStats(m, kit.Keys(m.CommandKey(), mdb.TOTAL), msg.Length(), "", "已启动空间")
				}
			}},
		}, aaa.RoleAction(), StatsAction(), DreamAction(), mdb.ImportantHashAction(ctx.TOOLS, "web.route",
			mdb.SHORT, mdb.NAME, mdb.FIELD, "time,name,icon,repos,binary,template,restart")), Hand: func(m *ice.Message, arg ...string) {
			if ice.Info.NodeType == WORKER {
				return
			}
			if len(arg) == 0 {
				_dream_list(m).RewriteAppend(func(value, key string, index int) string {
					if key == mdb.ICON {
						if kit.HasPrefix(value, HTTP, nfs.PS) {
							return value
						}
						if nfs.ExistsFile(m, path.Join(ice.USR_LOCAL_WORK, m.Appendv(mdb.NAME)[index], value)) {
							return kit.MergeURL(ctx.FileURI(value), ice.POD, m.Appendv(mdb.NAME)[index])
						}
						if nfs.ExistsFile(m, value) {
							return kit.MergeURL(ctx.FileURI(value))
						}
						return kit.MergeURL(ctx.FileURI(nfs.USR_ICONS_ICEBERGS))
					}
					return value
				}).Option(ice.MSG_ACTION, "")
				ctx.DisplayTableCard(m)
			} else if arg[0] == ctx.ACTION {
				gdb.Event(m, DREAM_ACTION, arg)
			} else {
				mdb.HashSelects(m, arg[0])
				// m.EchoIFrame(m.MergePod(arg[0]))
			}
		}},
	})
}

func DreamAction() ice.Actions {
	return ice.MergeActions(ice.Actions{
		DREAM_ACTION: {Hand: func(m *ice.Message, arg ...string) { DreamProcess(m, []string{}, arg...) }},
	}, gdb.EventsAction(DREAM_OPEN, DREAM_CLOSE, DREAM_INPUTS, DREAM_CREATE, DREAM_TRASH, DREAM_TABLES, DREAM_ACTION, SERVE_START))
}
func DreamProcess(m *ice.Message, args ice.Any, arg ...string) {
	if kit.HasPrefixList(arg, ctx.RUN) {
		ctx.ProcessField(m, m.PrefixKey(), args, kit.Slice(arg, 1)...)
	} else if kit.HasPrefixList(arg, ctx.ACTION, m.PrefixKey()) || kit.HasPrefixList(arg, ctx.ACTION, m.CommandKey()) {
		if arg = kit.Slice(arg, 2); kit.HasPrefixList(arg, DREAM) {
			m.Cmdy(SPACE, m.Option(ice.MSG_USERPOD, arg[1]), m.PrefixKey(), ctx.ACTION, DREAM_ACTION, ctx.RUN, arg[2:])
		} else if dream := m.Option(mdb.NAME); dream != "" {
			m.Cmdy(SPACE, dream, m.PrefixKey(), ctx.ACTION, DREAM_ACTION, ctx.RUN, arg).Optionv(ice.FIELD_PREFIX, kit.Simple(ctx.ACTION, m.PrefixKey(), DREAM, dream, ctx.RUN))
			m.Push("_space", dream)
		}
	}
}
func DreamEach(m *ice.Message, name string, status string, cb func(string)) *ice.Message {
	reg, err := regexp.Compile(name)
	if m.Warn(err) {
		return m
	}
	list := []string{}
	m.Spawn().Cmds(DREAM).Table(func(value ice.Maps) {
		if value[mdb.STATUS] == kit.Select(cli.START, status) && reg.MatchString(value[mdb.NAME]) {
			list = append(list, value[mdb.NAME])
		}
	})
	if len(list) == 0 {
		return m
	}
	GoToast(m, "", func(toast func(string, int, int)) []string {
		kit.For(list, func(index int, name string) {
			toast(name, index, len(list))
			cb(name)
		})
		return nil
	})
	return m
}
