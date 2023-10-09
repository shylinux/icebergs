package code

import (
	"path"
	"strings"
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _install_path(m *ice.Message, link string) string {
	u := kit.ParseURL(kit.Select(m.Option(web.LINK), link))
	if p := path.Join(ice.USR_INSTALL, kit.TrimExt(u.Path)); nfs.Exists(m, p) {
		return p
	} else if pp := path.Join(ice.USR_INSTALL, path.Base(u.Path)); nfs.Exists(m, pp) {
		return path.Join(ice.USR_INSTALL, strings.Split(m.Cmd(nfs.TAR, pp, "", "1").Append(nfs.FILE), nfs.PS)[0])
	} else {
		return p
	}
}
func _install_download(m *ice.Message) {
	link := m.Option(web.LINK)
	name := path.Base(kit.ParseURL(link).Path)
	// file := path.Join(kit.Select(ice.USR_INSTALL, m.Option(nfs.PATH)), name)
	file := path.Join(ice.USR_INSTALL, name)
	defer m.Cmdy(nfs.DIR, file)
	if nfs.Exists(m, file) {
		return
	}
	mdb.HashCreate(m.Cmd(nfs.SAVE, file, ""), mdb.NAME, name, nfs.PATH, file, web.LINK, link)
	web.GoToast(m, name, func(toast func(string, int, int)) (list []string) {
		begin := time.Now()
		_toast := func(count, total, value int) {
			cost := time.Now().Sub(begin)
			toast(kit.FormatShow(nfs.FROM, begin.Format("15:04:05"), cli.COST, kit.FmtDuration(cost), cli.REST, kit.FmtDuration(cost*time.Duration(101)/time.Duration(value+1)-cost)), count, total)
			mdb.HashSelectUpdate(m, name, func(value ice.Map) { value[mdb.COUNT], value[mdb.TOTAL], value[mdb.VALUE] = count, total, value })
		}
		defer nfs.TarExport(m, file)
		if mdb.Config(m, nfs.REPOS) != "" {
			web.SpideSave(m, file, mdb.Config(m, nfs.REPOS)+path.Base(link), _toast)
			if s, e := nfs.StatFile(m, file); e == nil && s.Size() > 0 {
				return
			}
		}
		web.SpideSave(m, file, link, _toast)
		return
	})
	if s, e := nfs.StatFile(m, file); e == nil && s.Size() > 0 {
		web.ToastSuccess(m)
	} else {
		web.ToastFailure(m)
		nfs.Trash(m, file)
	}
}
func _install_build(m *ice.Message, arg ...string) string {
	p := m.Option(cli.CMD_DIR, _install_path(m, ""))
	pp := kit.Path(path.Join(p, _INSTALL))
	switch cb := m.Optionv(PREPARE).(type) {
	case func(string):
		cb(p)
	case nil:
		if nfs.Exists(m, path.Join(p, "./configure")) {
			if msg := m.Cmd(cli.SYSTEM, "./configure", "--prefix="+pp, arg[1:]); !cli.IsSuccess(msg) {
				return msg.Append(cli.CMD_ERR) + msg.Append(cli.CMD_OUT)
			}
		}
	default:
		return m.ErrorNotImplement(cb).Result()
	}
	if msg := m.Cmd(cli.SYSTEM, cli.MAKE, "-j"+m.Cmdx(cli.RUNTIME, cli.MAXPROCS)); !cli.IsSuccess(msg) {
		return msg.Append(cli.CMD_ERR) + msg.Append(cli.CMD_OUT)
	} else if msg := m.Cmd(cli.SYSTEM, cli.MAKE, "PREFIX="+pp, INSTALL); !cli.IsSuccess(msg) {
		return msg.Append(cli.CMD_ERR) + msg.Append(cli.CMD_OUT)
	} else {
		return ""
	}
}
func _install_order(m *ice.Message, arg ...string) {
	p := _install_path(m, "")
	kit.For([]string{"", "sbin", "bin", "_install/bin"}, func(v string) {
		kit.If(nfs.Exists(m, path.Join(p, v)), func() { m.Option(nfs.PATH, v) })
	})
	m.Cmdy(cli.SYSTEM, nfs.PUSH, path.Join(p, m.Option(nfs.PATH)))
}
func _install_spawn(m *ice.Message, arg ...string) {
	if kit.Int(m.Option(tcp.PORT)) >= 10000 {
		if p := path.Join(ice.USR_LOCAL_DAEMON, m.Option(tcp.PORT)); nfs.Exists(m, p) {
			m.Echo(p)
			return
		}
	} else {
		m.Option(tcp.PORT, m.Cmdx(tcp.PORT, aaa.RIGHT))
	}
	target, source := path.Join(ice.USR_LOCAL_DAEMON, m.Option(tcp.PORT)), _install_path(m, "")
	nfs.MkdirAll(m, target)
	defer m.Echo(target)
	kit.If(m.Option(INSTALL) == "" && nfs.Exists(m, kit.Path(source, _INSTALL)), func() { m.Option(INSTALL, _INSTALL) })
	nfs.DirDeepAll(m.Spawn(), path.Join(source, m.Option(INSTALL)), "", func(value ice.Maps) {
		m.Cmd(nfs.LINK, path.Join(target, value[nfs.PATH]), path.Join(source, m.Option(INSTALL), value[nfs.PATH]), kit.Dict(ice.LOG_DISABLE, ice.TRUE))
	})
}
func _install_start(m *ice.Message, arg ...string) {
	args, p := []string{}, m.Option(cli.CMD_DIR, m.Cmdx(INSTALL, cli.SPAWN))
	switch cb := m.Optionv(PREPARE).(type) {
	case func(string) []string:
		args = append(args, cb(p)...)
	case func(string, int):
		cb(p, kit.Int(path.Base(p)))
	case func(string):
		cb(p)
	case nil:
	default:
		m.ErrorNotImplement(cb)
		return
	}
	m.Cmdy(cli.DAEMON, kit.Select(path.Join(ice.BIN, path.Base(_install_path(m, ""))), arg, 1), kit.Slice(arg, 2), args)
}
func _install_stop(m *ice.Message, arg ...string) {
	m.Cmd(cli.DAEMON, func(value ice.Maps) {
		kit.If(value[cli.PID] == m.Option(cli.PID), func() { m.Cmd(cli.DAEMON, cli.STOP, kit.Dict(mdb.HASH, value[mdb.HASH])) })
	})
	kit.If(m.Option(cli.PID), func() { m.Cmd(gdb.SIGNAL, gdb.KILL, m.Option(cli.PID)) })
}
func _install_clear(m *ice.Message, arg ...string) {
	m.Cmd(cli.DAEMON, func(value ice.Maps) {
		kit.If(value[cli.PID] == m.Option(cli.PID), func() { m.Cmd(cli.DAEMON, mdb.REMOVE, kit.Dict(mdb.HASH, value[mdb.HASH])) })
	})
}
func _install_trash(m *ice.Message, arg ...string) {
	if m.Option(tcp.PORT) == "" {
		nfs.Trash(m, m.Option(nfs.PATH))
	} else {
		m.Cmd(cli.DAEMON, mdb.REMOVE)
		nfs.Trash(m, kit.Path(ice.USR_LOCAL_DAEMON, m.Option(tcp.PORT), m.Option(nfs.PATH)))
	}
}
func _install_service(m *ice.Message, arg ...string) {
	name := kit.Split(path.Base(arg[0]), "_-.")[0]
	m.Fields(len(kit.Slice(arg, 1)), "time,port,status,pid,cmd,dir")
	m.Cmd(mdb.SELECT, cli.DAEMON, "", mdb.HASH, func(value ice.Maps) {
		if strings.Contains(value[ice.CMD], path.Join(ice.BIN, name)) {
			switch m.Push("", value, kit.Split(m.OptionFields())); value[mdb.STATUS] {
			case cli.START:
				m.PushButton(gdb.DEBUG, cli.STOP)
			case cli.STOP:
				m.PushButton(cli.START, cli.END)
			default:
				m.PushButton("")
			}
		}
	})
	m.Set(tcp.PORT).Table(func(value ice.Maps) { m.Push(tcp.PORT, path.Base(value[nfs.DIR])) }).StatusTimeCount()
}

const (
	PREPARE  = "prepare"
	_INSTALL = "_install"
)
const INSTALL = "install"

func init() {
	Index.MergeCommands(ice.Commands{
		INSTALL: {Name: "install name port path:text auto download", Help: "安装", Actions: ice.MergeActions(ice.Actions{
			web.DOWNLOAD: {Name: "download link* path", Hand: func(m *ice.Message, arg ...string) { _install_download(m) }},
			cli.BUILD: {Name: "build link*", Hand: func(m *ice.Message, arg ...string) {
				web.PushStream(m)
				if err := _install_build(m, arg...); m.Warn(err != "", err) {
					web.ToastFailure(m, cli.BUILD, err)
				} else {
					web.ToastSuccess(m, cli.BUILD)
				}
			}},
			cli.ORDER: {Name: "order link* path", Hand: func(m *ice.Message, arg ...string) { _install_order(m, arg...) }},
			cli.SPAWN: {Name: "spawn link*", Hand: func(m *ice.Message, arg ...string) { _install_spawn(m, arg...) }},
			cli.START: {Name: "start link* cmd", Hand: func(m *ice.Message, arg ...string) { _install_start(m, arg...) }},
			cli.STOP:  {Hand: func(m *ice.Message, arg ...string) { _install_stop(m, arg...) }},
			cli.CLEAR: {Hand: func(m *ice.Message, arg ...string) { _install_clear(m, arg...) }},
			gdb.DEBUG: {Hand: func(m *ice.Message, arg ...string) { ctx.Process(m, XTERM, []string{mdb.TYPE, "gdb"}, arg...) }},
			nfs.TRASH: {Hand: func(m *ice.Message, arg ...string) { _install_trash(m, arg...) }},
			nfs.PATH:  {Hand: func(m *ice.Message, arg ...string) { m.Echo(_install_path(m, kit.Select("", arg, 0))) }},
			nfs.SOURCE: {Name: "source link* path", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(nfs.DIR_ROOT, path.Join(_install_path(m, ""), _INSTALL)); !nfs.Exists(m, m.Option(nfs.DIR_ROOT)) {
					m.Option(nfs.DIR_ROOT, path.Join(_install_path(m, "")))
				}
				m.Cmdy(nfs.DIR, m.Option(nfs.PATH))
			}},
			mdb.REMOVE: {Hand: func(m *ice.Message, arg ...string) { nfs.Trash(mdb.HashRemove(m), m.Option(nfs.PATH)) }},
		}, mdb.HashAction(mdb.SHORT, mdb.NAME, mdb.FIELD, "time,count,total,name,path,link")), Hand: func(m *ice.Message, arg ...string) {
			switch len(arg) {
			case 0:
				mdb.HashSelect(m, arg...).PushAction(cli.BUILD, cli.ORDER, mdb.REMOVE)
			case 1:
				_install_service(m, arg...)
			default:
				m.Cmdy(nfs.CAT, kit.Select(nfs.PWD, arg, 2), kit.Dict(nfs.DIR_ROOT, path.Join(ice.USR_LOCAL_DAEMON, arg[1])))
			}
		}},
	})
}

func InstallAction(args ...ice.Any) ice.Actions {
	return ice.Actions{ice.CTX_INIT: mdb.AutoConfig(args...),
		web.DOWNLOAD: {Help: "下载", Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(INSTALL, web.DOWNLOAD, mdb.Config(m, nfs.SOURCE))
		}},
		cli.BUILD: {Help: "构建", Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(INSTALL, cli.BUILD, mdb.Config(m, nfs.SOURCE))
		}},
		cli.ORDER: {Help: "加载", Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(INSTALL, cli.ORDER, mdb.Config(m, nfs.SOURCE), path.Join(_INSTALL, ice.BIN))
		}},
		nfs.TRASH: {Hand: func(m *ice.Message, arg ...string) {
			nfs.Trash(m, m.Option(nfs.PATH))
		}},
		mdb.SELECT: {Name: "select path auto order build download", Hand: func(m *ice.Message, arg ...string) {
			m.Options(nfs.PATH, "").Cmdy(INSTALL, ctx.ConfigSimple(m, nfs.SOURCE), arg)
		}},
	}
}
