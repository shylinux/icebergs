package web

import (
	"os"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
)

func _dream_list(m *ice.Message) *ice.Message {
	list := m.CmdMap(SPACE, mdb.NAME)
	m.Cmdy(nfs.DIR, ice.USR_LOCAL_WORK, "time,size,name").Table(func(value ice.Maps) {
		if space, ok := list[value[mdb.NAME]]; ok {
			msg := gdb.Event(m.Spawn(value, space), DREAM_TABLES).Copy(m.Spawn().PushButton(cli.STOP))
			m.Push(mdb.TYPE, space[mdb.TYPE])
			m.Push(cli.STATUS, cli.START)
			m.Push(mdb.TEXT, msg.Append(mdb.TEXT))
			m.PushButton(strings.Join(msg.Appendv(ctx.ACTION), ""))
		} else {
			m.Push(mdb.TYPE, WORKER)
			m.Push(cli.STATUS, cli.STOP)
			m.Push(mdb.TEXT, "")
			m.PushButton(cli.START, nfs.TRASH)
		}
	})
	return m.Sort("status,type,name", ice.STR, ice.STR, ice.STR_R).StatusTimeCount(cli.START, len(list))
}
func _dream_show(m *ice.Message, name string) {
	if m.Warn(name == "", ice.ErrNotValid, mdb.NAME) {
		return
	}
	kit.If(!strings.Contains(name, "-") || !strings.HasPrefix(name, "20"), func() { name = m.Time("20060102-") + name })
	defer m.ProcessOpen(m.MergePod(m.Option(mdb.NAME, name)))
	p := path.Join(ice.USR_LOCAL_WORK, name)
	if pid := m.Cmdx(nfs.CAT, path.Join(p, ice.Info.PidPath), kit.Dict(ice.MSG_USERROLE, aaa.TECH)); pid != "" && nfs.Exists(m, "/proc/"+pid) {
		m.Info("already exists %v", pid)
		return
	} else if m.Cmd(SPACE, name).Length() > 0 {
		m.Info("already exists %v", name)
		return
	}
	defer ToastProcess(m)()
	defer m.Sleep300ms()
	m.Options(cli.CMD_DIR, kit.Path(p), cli.CMD_ENV, kit.Simple(
		cli.CTX_OPS, Domain(tcp.LOCALHOST, m.Cmdv(SERVE, tcp.PORT)), cli.CTX_LOG, ice.VAR_LOG_BOOT_LOG, cli.CTX_PID, ice.VAR_LOG_ICE_PID,
		cli.PATH, cli.BinPath(p, ""), cli.USER, ice.Info.Username, kit.EnvSimple(cli.HOME, cli.TERM, cli.SHELL), mdb.Configv(m, cli.ENV),
	), cli.CMD_OUTPUT, path.Join(p, ice.VAR_LOG_BOOT_LOG))
	defer m.Options(cli.CMD_DIR, "", cli.CMD_ENV, "", cli.CMD_OUTPUT, "")
	gdb.Event(m, DREAM_CREATE, m.OptionSimple(mdb.NAME, mdb.TYPE))
	kit.If(m.Option(nfs.TEMPLATE), func() { _dream_template(m, p) })
	m.Cmd(cli.DAEMON, kit.Select(kit.Path(os.Args[0]), cli.SystemFind(m, ice.ICE_BIN, nfs.PWD+path.Join(p, ice.BIN), nfs.PWD+ice.BIN)),
		SPACE, tcp.DIAL, ice.DEV, ice.OPS, mdb.TYPE, WORKER, m.OptionSimple(mdb.NAME), cli.DAEMON, ice.OPS)
}
func _dream_template(m *ice.Message, p string) {
	kit.For([]string{ice.ETC_MISS_SH,
		ice.LICENSE, ice.MAKEFILE, ice.README_MD, ice.GO_MOD, ice.GO_SUM,
		ice.SRC_MAIN_GO, ice.SRC_MAIN_SH, ice.SRC_MAIN_SHY, ice.SRC_MAIN_JS,
	}, func(file string) {
		if nfs.Exists(m, path.Join(p, file)) {
			return
		}
		switch m.Cmdy(nfs.COPY, path.Join(p, file), path.Join(ice.USR_LOCAL_WORK, m.Option(nfs.TEMPLATE), file)); file {
		case ice.GO_MOD:
			nfs.Rewrite(m, path.Join(p, file), func(line string) string {
				return kit.Select(line, nfs.MODULE+ice.SP+m.Option(mdb.NAME), strings.HasPrefix(line, nfs.MODULE))
			})
		}
	})
}

const (
	DREAM_CREATE = "dream.create"
	DREAM_OPEN   = "dream.open"
	DREAM_CLOSE  = "dream.close"

	DREAM_INPUTS = "dream.inputs"
	DREAM_TABLES = "dream.tables"
	DREAM_ACTION = "dream.action"
)
const DREAM = "dream"

func init() {
	Index.MergeCommands(ice.Commands{
		DREAM: {Name: "dream name path auto create", Help: "梦想家", Actions: ice.MergeActions(ice.Actions{
			mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == mdb.FOREACH && arg[1] == "" {
					m.Cmds("", func(value ice.Maps) { m.PushSearch(mdb.TEXT, m.MergePod(value[mdb.NAME]), value) })
				}
			}},
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case mdb.NAME, nfs.TEMPLATE:
					_dream_list(m).Cut("name,status,time")
				case nfs.REPOS:
					if msg := m.Cmd(SPIDE, ice.OPS, SPIDE_MSG, UserHost(m)+"/x/list"); !msg.IsErr() {
						m.Copy(msg)
					}
					kit.For([]string{ice.OPS, ice.DEV, ice.SHY}, func(dev string) {
						if msg := m.Cmd(SPIDE, dev, SPIDE_MSG, "/x/list"); !msg.IsErr() {
							m.Copy(msg)
						}
					})
				default:
					gdb.Event(m, "", arg)
				}
			}},
			mdb.CREATE: {Name: "create name*=hi repos template", Hand: func(m *ice.Message, arg ...string) {
				m.Option(nfs.REPOS, kit.Select("", kit.Slice(kit.Split(m.Option(nfs.REPOS)), -1), 0))
				_dream_show(m, m.OptionDefault(mdb.NAME, path.Base(m.Option(nfs.REPOS))))
			}},
			cli.START: {Hand: func(m *ice.Message, arg ...string) {
				_dream_show(m, m.Option(mdb.NAME))
			}},
			cli.STOP: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(SPACE, mdb.MODIFY, m.OptionSimple(mdb.NAME), mdb.STATUS, cli.STOP)
				m.Go(func() { m.Cmd(SPACE, m.Option(mdb.NAME), ice.EXIT) })
				m.Sleep30ms()
			}},
			nfs.TRASH: {Hand: func(m *ice.Message, arg ...string) {
				nfs.Trash(m, path.Join(ice.USR_LOCAL_WORK, m.Option(mdb.NAME)))
			}},
			DREAM_CLOSE: {Hand: func(m *ice.Message, arg ...string) {
				if m.Option(cli.DAEMON) == ice.OPS && m.Cmdv(SPACE, m.Option(mdb.NAME), mdb.STATUS) != cli.STOP {
					m.Go(func() { m.Sleep30ms(DREAM, cli.START, m.OptionSimple(mdb.NAME)) })
				}
			}},
			DREAM_TABLES: {Hand: func(m *ice.Message, arg ...string) {
				kit.Switch(m.Option(mdb.TYPE), []string{SERVER, WORKER}, func() { m.PushButton(OPEN) })
			}},
			OPEN: {Hand: func(m *ice.Message, arg ...string) { ctx.ProcessOpen(m, m.MergePod(m.Option(mdb.NAME), arg)) }},
		}, ctx.CmdAction(), DreamAction()), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				_dream_list(m)
				ctx.DisplayTableCard(m)
			} else if arg[0] == ctx.ACTION {
				gdb.Event(m, DREAM_ACTION, arg)
			} else {
				m.Cmdy(nfs.CAT, arg[1:], kit.Dict(nfs.DIR_ROOT, path.Join(ice.USR_LOCAL_WORK, arg[0])))
			}
		}},
	})
}

func DreamAction() ice.Actions {
	return ice.MergeActions(ice.Actions{
		DREAM_ACTION: {Hand: func(m *ice.Message, arg ...string) { DreamProcess(m, []string{}, arg...) }},
	}, gdb.EventsAction(DREAM_OPEN, DREAM_CLOSE, DREAM_INPUTS, DREAM_TABLES, DREAM_ACTION))
}
func DreamProcess(m *ice.Message, args ice.Any, arg ...string) {
	if kit.HasPrefixList(arg, ice.RUN) {
		ctx.ProcessField(m, m.PrefixKey(), args, kit.Slice(arg, 1)...)
	} else if kit.HasPrefixList(arg, ctx.ACTION, m.CommandKey()) {
		if arg = kit.Slice(arg, 2); kit.HasPrefixList(arg, DREAM) {
			m.Cmdy(SPACE, m.Option(ice.MSG_USERPOD, arg[1]), m.PrefixKey(), ctx.ACTION, DREAM_ACTION, ice.RUN, arg[2:])
		} else if dream := m.Option(mdb.NAME); dream != "" {
			m.Cmdy(SPACE, dream, m.PrefixKey(), ctx.ACTION, DREAM_ACTION, ice.RUN, arg).Optionv(ice.FIELD_PREFIX, kit.Simple(ctx.ACTION, m.CommandKey(), DREAM, dream, ice.RUN))
		}
	}
}
