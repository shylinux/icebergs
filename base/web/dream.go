package web

import (
	"os"
	"path"
	"runtime"
	"strings"

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
	list := m.CmdMap(SPACE, mdb.NAME)
	stats := map[string]int{}
	mdb.HashSelect(m).Table(func(value ice.Maps) {
		if space, ok := list[value[mdb.NAME]]; ok {
			msg := gdb.Event(m.Spawn(value, space), DREAM_TABLES).Copy(m.Spawn().PushButton(cli.STOP))
			m.Push(mdb.TYPE, space[mdb.TYPE])
			m.Push(cli.STATUS, cli.START)
			m.Push(mdb.TEXT, msg.Append(mdb.TEXT))
			m.PushButton(strings.Join(msg.Appendv(ctx.ACTION), ""))
			stats[cli.START]++
		} else if nfs.Exists(m, path.Join(ice.USR_LOCAL_WORK, value[mdb.NAME])) {
			m.Push(mdb.TYPE, WORKER)
			m.Push(cli.STATUS, cli.STOP)
			m.Push(mdb.TEXT, "")
			m.PushButton(cli.START, nfs.TRASH)
			stats[cli.STOP]++
		} else {
			m.Push(mdb.TYPE, WORKER)
			m.Push(cli.STATUS, cli.STOP)
			m.Push(mdb.TEXT, "")
			m.PushButton(cli.START, mdb.REMOVE)
			stats[ice.INIT]++
		}
	})
	return m.Sort("status,type,name", ice.STR, ice.STR, ice.STR).StatusTimeCount(stats)
}
func _dream_start(m *ice.Message, name string) {
	if m.Warn(name == "", ice.ErrNotValid, mdb.NAME) {
		return
	}
	defer m.ProcessOpen(m.MergePod(m.Option(mdb.NAME, name)))
	p := path.Join(ice.USR_LOCAL_WORK, name)
	if pid := m.Cmdx(nfs.CAT, path.Join(p, ice.Info.PidPath), kit.Dict(ice.MSG_USERROLE, aaa.TECH)); pid != "" && nfs.Exists(m, "/proc/"+pid) {
		m.Info("already exists %v", pid)
		return
	} else if msg := m.Cmd(SPACE, name); msg.Length() > 0 {
		m.Info("already exists %v", name)
		return
	}
	defer ToastProcess(m)()
	defer m.Sleep("1s")
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

	} else if kit.IsUrl(p) {
		SpideSave(m, bin, kit.MergeURL(p, cli.GOOS, runtime.GOOS, cli.GOARCH, runtime.GOARCH), nil)
		os.Chmod(bin, ice.MOD_DIR)
	} else {
		m.Cmd(nfs.LINK, bin, kit.Path(p))
	}
}
func _dream_template(m *ice.Message, p string) {
	kit.For([]string{
		ice.LICENSE, ice.MAKEFILE, ice.README_MD, ice.GO_MOD, ice.GO_SUM,
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
	DREAM_CREATE = "dream.create"
	DREAM_START  = "dream.start"
	DREAM_STOP   = "dream.stop"
	DREAM_OPEN   = "dream.open"
	DREAM_CLOSE  = "dream.close"
	DREAM_TRASH  = "dream.trash"
	DREAM_REMOVE = "dream.remove"

	DREAM_INPUTS = "dream.inputs"
	DREAM_TABLES = "dream.tables"
	DREAM_ACTION = "dream.action"
)
const DREAM = "dream"

func init() {
	Index.MergeCommands(ice.Commands{
		DREAM: {Name: "dream name@key auto create", Help: "梦想家", Actions: ice.MergeActions(ice.Actions{
			mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
				if mdb.IsSearchPreview(m, arg) {
					m.Cmds("", func(value ice.Maps) { m.PushSearch(mdb.TEXT, m.MergePod(value[mdb.NAME]), value) })
				}
			}},
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case mdb.NAME, nfs.TEMPLATE:
					_dream_list(m).Cut("name,status,time")
				case nfs.BINARY:
					m.Cmdy(nfs.DIR, ice.BIN, "path,size,time", kit.Dict(nfs.DIR_TYPE, nfs.TYPE_BIN))
					m.Cmd(nfs.DIR, ice.USR_LOCAL_WORK, kit.Dict(nfs.DIR_TYPE, nfs.TYPE_BOTH), func(value ice.Maps) {
						m.Cmdy(nfs.DIR, path.Join(value[nfs.PATH], ice.BIN), "path,size,time", kit.Dict(nfs.DIR_TYPE, nfs.TYPE_BIN))
					})
				default:
					gdb.Event(m, DREAM_INPUTS, arg)
				}
			}},
			mdb.CREATE: {Name: "create name*=hi repos binary template", Hand: func(m *ice.Message, arg ...string) {
				m.Option(nfs.REPOS, kit.Select("", kit.Slice(kit.Split(m.Option(nfs.REPOS)), -1), 0))
				kit.If(!strings.Contains(m.Option(mdb.NAME), "-") || !strings.HasPrefix(m.Option(mdb.NAME), "20"), func() { m.Option(mdb.NAME, m.Time("20060102-")+m.Option(mdb.NAME)) })
				if mdb.HashCreate(m); !m.IsCliUA() {
					_dream_start(m, m.OptionDefault(mdb.NAME, path.Base(m.Option(nfs.REPOS))))
				}
			}},
			cli.START: {Hand: func(m *ice.Message, arg ...string) {
				gdb.Event(m, DREAM_START, arg)
				_dream_start(m, m.Option(mdb.NAME))
			}},
			cli.STOP: {Hand: func(m *ice.Message, arg ...string) {
				gdb.Event(m, DREAM_STOP, arg)
				m.Cmd(SPACE, mdb.MODIFY, m.OptionSimple(mdb.NAME), mdb.STATUS, cli.STOP)
				m.Go(func() { m.Cmd(SPACE, m.Option(mdb.NAME), ice.EXIT) })
				m.Sleep30ms()
			}},
			nfs.TRASH: {Hand: func(m *ice.Message, arg ...string) {
				gdb.Event(m, DREAM_TRASH, arg)
				nfs.Trash(m, path.Join(ice.USR_LOCAL_WORK, m.Option(mdb.NAME)))
			}},
			DREAM_CLOSE: {Hand: func(m *ice.Message, arg ...string) {
				if m.Option(cli.DAEMON) == ice.OPS && m.Cmdv(SPACE, m.Option(mdb.NAME), mdb.STATUS) != cli.STOP {
					m.Go(func() { m.Sleep300ms(DREAM, cli.START, m.OptionSimple(mdb.NAME)) })
				}
			}},
			DREAM_TABLES: {Hand: func(m *ice.Message, arg ...string) {
				kit.Switch(m.Option(mdb.TYPE), []string{SERVER, WORKER}, func() { m.PushButton(OPEN) })
			}},
			OPEN: {Hand: func(m *ice.Message, arg ...string) { ctx.ProcessOpen(m, m.MergePod(m.Option(mdb.NAME))) }},
			"button": {Hand: func(m *ice.Message, arg ...string) {
				for _, cmd := range kit.Reverse(arg) {
					m.Cmd(gdb.EVENT, gdb.LISTEN, gdb.EVENT, DREAM_TABLES, ice.CMD, cmd)
					m.Cmd(gdb.EVENT, gdb.LISTEN, gdb.EVENT, DREAM_ACTION, ice.CMD, cmd)
				}
			}},
		}, ctx.CmdAction(), DreamAction(), mdb.ImportantHashAction(mdb.SHORT, mdb.NAME, mdb.FIELD, "time,name,repos,binary,template")), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				_dream_list(m)
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
	}, gdb.EventsAction(DREAM_OPEN, DREAM_CLOSE, DREAM_INPUTS, DREAM_CREATE, DREAM_TABLES, DREAM_ACTION))
}
func DreamProcess(m *ice.Message, args ice.Any, arg ...string) {
	if kit.HasPrefixList(arg, ice.RUN) {
		ctx.ProcessField(m, m.PrefixKey(), args, kit.Slice(arg, 1)...)
	} else if kit.HasPrefixList(arg, ctx.ACTION, m.PrefixKey()) || kit.HasPrefixList(arg, ctx.ACTION, m.CommandKey()) {
		if arg = kit.Slice(arg, 2); kit.HasPrefixList(arg, DREAM) {
			m.Cmdy(SPACE, m.Option(ice.MSG_USERPOD, arg[1]), m.PrefixKey(), ctx.ACTION, DREAM_ACTION, ice.RUN, arg[2:])
		} else if dream := m.Option(mdb.NAME); dream != "" {
			m.Cmdy(SPACE, dream, m.PrefixKey(), ctx.ACTION, DREAM_ACTION, ice.RUN, arg).Optionv(ice.FIELD_PREFIX, kit.Simple(ctx.ACTION, m.PrefixKey(), DREAM, dream, ice.RUN))
			m.Push("_space", dream)
		}
	}
}
