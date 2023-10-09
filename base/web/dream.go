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
			m.Push(cli.STATUS, cli.STOP)
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
	return m.Sort("status,type,name", ice.STR, ice.STR, ice.STR_R).StatusTimeCount(stats)

}
func _dream_start(m *ice.Message, name string) {
	if m.Warn(name == "", ice.ErrNotValid, mdb.NAME) {
		return
	}
	defer m.ProcessOpen(m.MergePod(m.Option(mdb.NAME, name)))
	p := path.Join(ice.USR_LOCAL_WORK, name)
	if pid := m.Cmdx(nfs.CAT, path.Join(p, ice.Info.PidPath), kit.Dict(ice.MSG_USERROLE, aaa.TECH)); pid != "" {
		if nfs.Exists(m, "/proc/"+pid) {
			m.Info("already exists %v", pid)
			return
		}
		for i := 0; i < 3; i++ {
			if msg := m.Cmd(SPACE, name); msg.Length() > 0 {
				m.Info("already exists %v", name)
				return
			}
			m.Sleep300ms()
		}
	}
	defer ToastProcess(m)()
	defer m.Sleep300ms()
	m.Options(cli.CMD_DIR, kit.Path(p), cli.CMD_ENV, kit.EnvList(kit.Simple(
		cli.CTX_OPS, Domain(tcp.LOCALHOST, m.Cmdv(SERVE, tcp.PORT)), cli.CTX_LOG, ice.VAR_LOG_BOOT_LOG, cli.CTX_PID, ice.VAR_LOG_ICE_PID,
		cli.PATH, cli.BinPath(p, ""), cli.USER, ice.Info.Username,
	)...), cli.CMD_OUTPUT, path.Join(p, ice.VAR_LOG_BOOT_LOG), mdb.CACHE_CLEAR_ONEXIT, ice.TRUE)
	defer m.Options(cli.CMD_DIR, "", cli.CMD_ENV, "", cli.CMD_OUTPUT, "")
	gdb.Event(m, DREAM_CREATE, m.OptionSimple(mdb.NAME, mdb.TYPE))
	kit.If(m.Option(nfs.BINARY), func(p string) { _dream_binary(m, p) })
	kit.If(m.Option(nfs.TEMPLATE), func(p string) { _dream_template(m, p) })
	m.Cmd(cli.DAEMON, kit.Select(kit.Path(os.Args[0]), cli.SystemFind(m, ice.ICE_BIN, nfs.PWD+path.Join(p, ice.BIN), nfs.PWD+ice.BIN)),
		SPACE, tcp.DIAL, ice.DEV, ice.OPS, mdb.TYPE, WORKER, m.OptionSimple(mdb.NAME), cli.DAEMON, ice.OPS)
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
		DREAM: {Name: "dream name@key auto create repos startall stopall build cmd cat", Icon: "Launchpad.png", Help: "梦想家", Actions: ice.MergeActions(ice.Actions{
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
				if m.Option(ctx.ACTION) == ice.MAIN {
					m.Cmdy(SPACE, m.Option(mdb.NAME), SPACE, mdb.INPUTS, arg)
					return
				}
				switch arg[0] {
				case mdb.NAME, nfs.TEMPLATE:
					_dream_list(m).Cut("name,status,time")
				case nfs.BINARY:
					m.Cmdy(nfs.DIR, ice.BIN, "path,size,time", kit.Dict(nfs.DIR_TYPE, nfs.TYPE_BIN))
					m.Cmd(nfs.DIR, ice.USR_LOCAL_WORK, kit.Dict(nfs.DIR_TYPE, nfs.TYPE_BOTH), func(value ice.Maps) {
						m.Cmdy(nfs.DIR, path.Join(value[nfs.PATH], ice.BIN), "path,size,time", kit.Dict(nfs.DIR_TYPE, nfs.TYPE_BIN))
					})
					m.RenameAppend(nfs.PATH, arg[0])
					mdb.HashInputs(m, arg)
				default:
					gdb.Event(m, DREAM_INPUTS, arg)
				}
			}},
			mdb.CREATE: {Name: "create name*=hi icon@icon repos binary template", Hand: func(m *ice.Message, arg ...string) {
				m.OptionDefault(mdb.ICON, nfs.USR_ICONS_ICEBERGS)
				m.Option(nfs.REPOS, kit.Select("", kit.Slice(kit.Split(m.Option(nfs.REPOS)), -1), 0))
				kit.If(!strings.Contains(m.Option(mdb.NAME), "-") || !strings.HasPrefix(m.Option(mdb.NAME), "20"), func() { m.Option(mdb.NAME, m.Time("20060102-")+m.Option(mdb.NAME)) })
				if mdb.HashCreate(m); !m.IsCliUA() {
					_dream_start(m, m.OptionDefault(mdb.NAME, path.Base(m.Option(nfs.REPOS))))
				}
			}},
			nfs.REPOS: {Help: "仓库", Hand: func(m *ice.Message, arg ...string) {
				m.ProcessOpen(m.MergePodCmd("", CODE_GIT_SEARCH, nfs.REPOS, nfs.REPOS))
			}},
			"startall": {Name: "startall name", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
				reg, err := regexp.Compile(m.Option(mdb.NAME))
				if m.Warn(err) {
					return
				}
				list := []string{}
				m.Spawn().Cmds("").Table(func(value ice.Maps) {
					if value[mdb.STATUS] == cli.STOP && reg.MatchString(value[mdb.NAME]) {
						list = append(list, value[mdb.NAME])
					}
				})
				if len(list) == 0 {
					return
				}
				GoToast(m, "", func(toast func(string, int, int)) []string {
					kit.For(list, func(index int, name string) {
						toast(name, index, len(list))
						m.Cmd("", cli.START, ice.Maps{mdb.NAME: name, ice.MSG_DAEMON: ""})
					})
					return nil
				})
			}},
			"stopall": {Name: "stopall name", Help: "停止", Hand: func(m *ice.Message, arg ...string) {
				reg, err := regexp.Compile(m.Option(mdb.NAME))
				if m.Warn(err) {
					return
				}
				list := []string{}
				m.Spawn().Cmds("").Table(func(value ice.Maps) {
					if value[mdb.STATUS] == cli.START && reg.MatchString(value[mdb.NAME]) {
						list = append(list, value[mdb.NAME])
					}
				})
				if len(list) == 0 {
					return
				}
				GoToast(m, "", func(toast func(string, int, int)) []string {
					kit.For(list, func(index int, name string) {
						toast(name, index, len(list))
						m.Cmd("", cli.STOP, ice.Maps{mdb.NAME: name, ice.MSG_DAEMON: ""})
					})
					return nil
				})
			}},
			cli.BUILD: {Hand: func(m *ice.Message, arg ...string) {
				GoToast(m, "", func(toast func(string, int, int)) []string {
					msg := mdb.HashSelect(m.Spawn())
					msg.Table(func(index int, value ice.Maps) {
						toast(value[mdb.NAME], index, msg.Length())
						m.Push(mdb.NAME, value[mdb.NAME])
						m.Push(mdb.TEXT, m.Cmdx(SPACE, value[mdb.NAME], "compile", cli.LINUX))
						m.Push(mdb.NAME, value[mdb.NAME])
						m.Push(mdb.TEXT, m.Cmdx(SPACE, value[mdb.NAME], "compile", cli.DARWIN))
						m.Push(mdb.NAME, value[mdb.NAME])
						m.Push(mdb.TEXT, m.Cmdx(SPACE, value[mdb.NAME], "compile", cli.WINDOWS))
					})
					return nil
				})
				m.StatusTimeCount()
			}},
			nfs.CAT: {Name: "cat file*", Help: "文件", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashSelect(m.Spawn()).Table(func(value ice.Maps) {
					m.Push(mdb.NAME, value[mdb.NAME])
					m.Push(mdb.TEXT, m.Cmdx(SPACE, value[mdb.NAME], nfs.CAT, m.Option(nfs.FILE)))
				})
				m.StatusTimeCount()
			}},
			ice.CMD: {Name: "cmd cmd*", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
				GoToast(m, "", func(toast func(string, int, int)) []string {
					msg := mdb.HashSelect(m.Spawn())
					msg.Table(func(index int, value ice.Maps) {
						toast(value[mdb.NAME], index, msg.Length())
						m.Push(mdb.NAME, value[mdb.NAME])
						m.Push(mdb.TEXT, m.Cmdx(SPACE, value[mdb.NAME], kit.Split(m.Option("cmd"))))
					})
					return nil
				})
				m.StatusTimeCount()
			}},
			cli.START: {Hand: func(m *ice.Message, arg ...string) {
				gdb.Event(m, DREAM_START, arg)
				_dream_start(m, m.Option(mdb.NAME))
			}},
			cli.STOP: {Hand: func(m *ice.Message, arg ...string) {
				gdb.Event(m, DREAM_STOP, arg)
				m.Cmd(SPACE, mdb.MODIFY, m.OptionSimple(mdb.NAME), mdb.STATUS, cli.STOP)
				m.Go(func() { m.Cmd(SPACE, m.Option(mdb.NAME), ice.EXIT) })
				m.Sleep300ms()
			}},
			nfs.TRASH: {Hand: func(m *ice.Message, arg ...string) {
				gdb.Event(m, DREAM_TRASH, arg)
				nfs.Trash(m, path.Join(ice.USR_LOCAL_WORK, m.Option(mdb.NAME)))
			}},
			ice.MAIN: {Name: "main index", Help: "首页", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(SPACE, m.Option(mdb.NAME), SPACE, ice.MAIN, m.Option(ctx.INDEX))
			}},
			OPEN: {Hand: func(m *ice.Message, arg ...string) { m.ProcessOpen(m.MergePod(m.Option(mdb.NAME))) }},
			DREAM_CLOSE: {Hand: func(m *ice.Message, arg ...string) {
				if m.Option(cli.DAEMON) == ice.OPS && m.Cmdv(SPACE, m.Option(mdb.NAME), mdb.STATUS) != cli.STOP {
					m.Go(func() { m.Sleep300ms(DREAM, cli.START, m.OptionSimple(mdb.NAME)) })
				}
			}},
			DREAM_TABLES: {Hand: func(m *ice.Message, arg ...string) {
				kit.Switch(m.Option(mdb.TYPE), []string{WORKER, SERVER}, func() { m.PushButton(OPEN, ice.MAIN) })
			}},
		}, DreamAction(), mdb.ImportantHashAction(mdb.SHORT, mdb.NAME, mdb.FIELD, "time,name,icon,repos,binary,template")), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				_dream_list(m).RewriteAppend(func(value, key string, index int) string {
					if key == mdb.ICON {
						return kit.MergeURL(ctx.FileURI(value), ice.POD, m.Appendv(mdb.NAME)[index])
					}
					return value
				}).Option(ice.MSG_ACTION, "")
				ctx.DisplayTableCard(m)
			} else if arg[0] == ctx.ACTION {
				gdb.Event(m, DREAM_ACTION, arg)
			} else {
				m.EchoIFrame(m.MergePod(arg[0]))
			}
		}},
	})
}

func DreamAction() ice.Actions {
	return ice.MergeActions(ice.Actions{
		DREAM_ACTION: {Hand: func(m *ice.Message, arg ...string) { DreamProcess(m, []string{}, arg...) }},
	}, gdb.EventsAction(DREAM_OPEN, DREAM_CLOSE, DREAM_INPUTS, DREAM_CREATE, DREAM_TRASH, DREAM_TABLES, DREAM_ACTION))
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
