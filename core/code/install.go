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
	link = kit.Select(m.Option(mdb.LINK), link)
	if p := path.Join(ice.USR_INSTALL, kit.TrimExt(link)); nfs.ExistsFile(m, p) {
		return p
	} else if pp := path.Join(ice.USR_INSTALL, path.Base(link)); nfs.ExistsFile(m, pp) {
		return path.Join(ice.USR_INSTALL, strings.Split(m.Cmd(nfs.TAR, pp, "", "1").Append(nfs.FILE), ice.PS)[0])
	} else {
		return p
	}
}
func _install_download(m *ice.Message) {
	link := m.Option(mdb.LINK)
	name := path.Base(strings.Split(link, "?")[0])
	file := path.Join(kit.Select(ice.USR_INSTALL, m.Option(nfs.PATH)), name)
	defer web.ToastSuccess(m)
	defer m.Cmdy(nfs.DIR, file)
	if nfs.ExistsFile(m, file) {
		return
	}
	defer m.SetResult()
	m.Cmd(nfs.SAVE, file, "")
	mdb.HashCreate(m, mdb.NAME, name, nfs.PATH, file, mdb.LINK, link)
	web.GoToast(m, name, func(toast func(string, int, int)) {
		defer nfs.TarExport(m, file)
		begin := time.Now()
		web.SpideSave(m, file, link, func(count, total, step int) {
			cost := time.Now().Sub(begin)
			mdb.HashSelectUpdate(m, name, func(value ice.Map) { value[mdb.COUNT], value[mdb.TOTAL], value[mdb.VALUE] = count, total, step })
			toast(kit.FormatShow("from", begin.Format("15:04:05"), "cost", kit.FmtDuration(cost), "rest", kit.FmtDuration(cost*time.Duration(101)/time.Duration(step+1)-cost)), count, total)
		})
	})
}
func _install_build(m *ice.Message, arg ...string) string {
	p := m.Option(cli.CMD_DIR, _install_path(m, ""))
	pp := kit.Path(path.Join(p, _INSTALL))
	switch cb := m.Optionv(PREPARE).(type) {
	case func(string):
		cb(p)
	case nil:
		if msg := m.Cmd(cli.SYSTEM, "./configure", "--prefix="+pp, arg[1:]); !cli.IsSuccess(msg) {
			return msg.Append(cli.CMD_ERR) + msg.Append(cli.CMD_OUT)
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
	if m.Option(nfs.PATH) == "" {
		for _, v := range []string{"_install/bin", "bin", "sbin", ""} {
			if nfs.ExistsFile(m, path.Join(p, v)) {
				m.Option(nfs.PATH, v)
				break
			}
		}
	}
	m.Cmdy(cli.SYSTEM, nfs.PUSH, path.Join(p, m.Option(nfs.PATH)))
}
func _install_spawn(m *ice.Message, arg ...string) {
	if kit.Int(m.Option(tcp.PORT)) >= 10000 {
		if p := path.Join(ice.USR_LOCAL_DAEMON, m.Option(tcp.PORT)); nfs.ExistsFile(m, p) {
			m.Echo(p)
			return
		}
	} else {
		m.Option(tcp.PORT, m.Cmdx(tcp.PORT, aaa.RIGHT))
	}
	target, source := path.Join(ice.USR_LOCAL_DAEMON, m.Option(tcp.PORT)), _install_path(m, "")
	nfs.MkdirAll(m, target)
	defer m.Echo(target)
	if m.Option(INSTALL) == "" && nfs.ExistsFile(m, kit.Path(source, _INSTALL)) {
		m.Option(INSTALL, _INSTALL)
	}
	nfs.DirDeepAll(m.Spawn(), path.Join(source, m.Option(INSTALL)), "", func(value ice.Maps) {
		m.Cmd(nfs.LINK, path.Join(target, value[nfs.PATH]), path.Join(source, m.Option(INSTALL), value[nfs.PATH]))
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
	m.Cmdy(cli.DAEMON, kit.Select(path.Join(ice.BIN, kit.Split(path.Base(arg[0]), "-.")[0]), arg, 1), kit.Slice(arg, 2), args)
}
func _install_stop(m *ice.Message, arg ...string) {
	m.Cmd(cli.DAEMON, func(value ice.Maps) {
		if value[cli.PID] == m.Option(cli.PID) {
			m.Cmd(cli.DAEMON, cli.STOP, kit.Dict(mdb.HASH, value[mdb.HASH]))
		}
	})
	m.Cmd(gdb.SIGNAL, gdb.KILL, m.Option(cli.PID))
}
func _install_end(m *ice.Message, arg ...string) {
	m.Cmd(cli.DAEMON, func(value ice.Maps) {
		if value[cli.PID] == m.Option(cli.PID) {
			m.Cmd(cli.DAEMON, mdb.REMOVE, kit.Dict(mdb.HASH, value[mdb.HASH]))
		}
	})
}
func _install_trash(m *ice.Message, arg ...string) {
	m.Cmd(cli.DAEMON, func(value ice.Maps) {
		if value[cli.PID] == m.Option(cli.PID) {
			m.Cmd(cli.DAEMON, mdb.REMOVE, kit.Dict(mdb.HASH, value[mdb.HASH]))
		}
	})
	nfs.Trash(m, kit.Path(ice.USR_LOCAL_DAEMON, m.Option(tcp.PORT), m.Option(nfs.PATH)))
}
func _install_service(m *ice.Message, arg ...string) {
	arg = kit.Split(path.Base(arg[0]), "_-.")[:1]
	m.Fields(len(arg[1:]), "time,port,status,pid,cmd,dir")
	m.Cmd(mdb.SELECT, cli.DAEMON, "", mdb.HASH, func(value ice.Maps) {
		if strings.Contains(value[ice.CMD], path.Join(ice.BIN, arg[0])) {
			m.Push("", value, kit.Split(m.OptionFields()))
			switch value[mdb.STATUS] {
			case cli.START:
				m.PushButton(gdb.DEBUG, cli.STOP)
			case cli.STOP:
				m.PushButton(cli.START, cli.END)
			default:
				m.PushButton("")
			}
		}
	})
	m.Set(tcp.PORT).Tables(func(value ice.Maps) { m.Push(tcp.PORT, path.Base(value[nfs.DIR])) }).StatusTimeCount()
}

const (
	PREPARE  = "prepare"
	_INSTALL = "_install"
)
const INSTALL = "install"

func init() {
	Index.MergeCommands(ice.Commands{
		INSTALL: {Name: "install name port path:text auto download", Help: "安装", Actions: ice.MergeActions(ice.Actions{
			nfs.PATH: {Hand: func(m *ice.Message, arg ...string) { m.Echo(_install_path(m, kit.Select("", arg, 0))) }},
			web.DOWNLOAD: {Name: "download link* path", Help: "下载", Hand: func(m *ice.Message, arg ...string) {
				_install_download(m)
			}},
			cli.BUILD: {Name: "build link*", Help: "构建", Hand: func(m *ice.Message, arg ...string) {
				web.PushStream(m)
				defer m.ProcessHold()
				if err := _install_build(m, arg...); m.Warn(err != "", err) {
					web.ToastFailure(m, cli.BUILD, err)
				} else {
					web.ToastSuccess(m, cli.BUILD)
				}
			}},
			cli.ORDER: {Name: "order link* path", Help: "加载", Hand: func(m *ice.Message, arg ...string) {
				_install_order(m, arg...)
			}},
			cli.SPAWN: {Name: "spawn link*", Help: "新建", Hand: func(m *ice.Message, arg ...string) {
				_install_spawn(m, arg...)
			}},
			cli.START: {Name: "start link* cmd", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
				_install_start(m, arg...)
			}},
			cli.STOP: {Help: "停止", Hand: func(m *ice.Message, arg ...string) {
				_install_stop(m, arg...)
			}},
			cli.END: {Help: "清理", Hand: func(m *ice.Message, arg ...string) {
				_install_end(m, arg...)
			}},
			gdb.DEBUG: {Help: "调试", Hand: func(m *ice.Message, arg ...string) {
				ctx.Process(m, XTERM, []string{mdb.TYPE, "gdb"}, arg...)
			}},
			nfs.TRASH: {Hand: func(m *ice.Message, arg ...string) {
				_install_trash(m, arg...)
			}},
			nfs.SOURCE: {Name: "source link* path", Help: "源码", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(nfs.DIR_ROOT, path.Join(_install_path(m, ""), _INSTALL)); !nfs.ExistsFile(m, m.Option(nfs.DIR_ROOT)) {
					m.Option(nfs.DIR_ROOT, path.Join(_install_path(m, "")))
				}
				m.Cmdy(nfs.DIR, m.Option(nfs.PATH)).StatusTimeCount(nfs.PATH, m.Option(nfs.DIR_ROOT))
			}},
			mdb.REMOVE: {Hand: func(m *ice.Message, arg ...string) {
				nfs.Trash(m, m.Option(nfs.PATH))
				mdb.HashRemove(m)
			}},
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
		web.DOWNLOAD: {Name: "download", Help: "下载", Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(INSTALL, web.DOWNLOAD, m.Config(nfs.SOURCE))
		}},
		cli.BUILD: {Name: "build", Help: "构建", Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(INSTALL, cli.BUILD, m.Config(nfs.SOURCE))
		}},
		cli.ORDER: {Name: "order", Help: "加载", Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(INSTALL, cli.ORDER, m.Config(nfs.SOURCE), path.Join(_INSTALL, ice.BIN))
		}},
		nfs.TRASH: {Hand: func(m *ice.Message, arg ...string) {
			nfs.Trash(m, m.Option(nfs.PATH))
		}},
	}
}
