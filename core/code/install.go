package code

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/file"
)

func _install_path(m *ice.Message, link string) string {
	link = kit.Select(m.Option(mdb.LINK), link)
	if p := path.Join(ice.USR_INSTALL, kit.TrimExt(link)); kit.FileExists(p) {
		return p
	}
	if p := path.Join(ice.USR_INSTALL, path.Base(link)); kit.FileExists(p) {
		return path.Join(ice.USR_INSTALL, strings.Split(m.Cmdx(cli.SYSTEM, "sh", "-c", kit.Format("tar tf %s| head -n1", p), ice.Option{cli.CMD_OUTPUT, ""}), ice.PS)[0])
	}
	m.ErrorNotImplement(link)
	return ""
}
func _install_download(m *ice.Message) {
	link := m.Option(mdb.LINK)
	name := path.Base(strings.Split(link, "?")[0])
	file := path.Join(kit.Select(ice.USR_INSTALL, m.Option(nfs.PATH)), name)

	defer m.Cmdy(nfs.DIR, file)
	if kit.FileExists(file) {
		return // 文件存在
	}

	// 创建文件
	m.Cmd(nfs.SAVE, file, "")
	web.GoToast(m, web.DOWNLOAD, func(toast func(string, int, int)) {
		m.Cmd(mdb.INSERT, INSTALL, "", mdb.HASH, mdb.NAME, name, nfs.PATH, file, mdb.LINK, link)
		defer web.ToastSuccess(m)

		// 下载进度
		mdb.Richs(m, INSTALL, "", name, func(key string, value ice.Map) {
			value = kit.GetMeta(value)
			m.OptionCB(web.SPIDE, func(count int, total int, step int) {
				value[mdb.COUNT], value[mdb.TOTAL], value[mdb.VALUE] = count, total, step
				toast(name, count, total)
			})
		})

		// 下载解压
		m.Cmd("web.spide", ice.DEV, web.SPIDE_SAVE, file, web.SPIDE_GET, link)
		m.Cmd(nfs.TAR, mdb.EXPORT, name, kit.Dict(cli.CMD_DIR, path.Dir(file)))
	})
}
func _install_build(m *ice.Message, arg ...string) string {
	p := m.Option(cli.CMD_DIR, _install_path(m, ""))
	pp := kit.Path(path.Join(p, "_install"))

	// 推流
	web.PushStream(m)
	defer m.ProcessHold()

	// 配置
	switch cb := m.Optionv(PREPARE).(type) {
	case func(string):
		cb(p)
	default:
		if msg := m.Cmd(cli.SYSTEM, "./configure", "--prefix="+pp, arg[1:]); !cli.IsSuccess(msg) {
			return msg.Append(cli.CMD_ERR)
		}
	}

	// 编译
	if msg := m.Cmd(cli.SYSTEM, cli.MAKE, "-j8"); !cli.IsSuccess(msg) {
		return msg.Append(cli.CMD_ERR)
	}

	// 安装
	if msg := m.Cmd(cli.SYSTEM, cli.MAKE, "PREFIX="+pp, INSTALL); !cli.IsSuccess(msg) {
		return msg.Append(cli.CMD_ERR)
	}
	return ""
}
func _install_order(m *ice.Message, arg ...string) {
	m.Cmdy(cli.SYSTEM, nfs.PUSH, path.Join(_install_path(m, ""), m.Option(nfs.PATH)+ice.NL))
}
func _install_spawn(m *ice.Message, arg ...string) {
	if kit.Int(m.Option(tcp.PORT)) >= 10000 {
		if p := path.Join(ice.USR_LOCAL_DAEMON, m.Option(tcp.PORT)); kit.FileExists(p) {
			m.Echo(p)
			return
		}
	} else {
		m.Option(tcp.PORT, m.Cmdx(tcp.PORT, aaa.RIGHT))
	}

	target := path.Join(ice.USR_LOCAL_DAEMON, m.Option(tcp.PORT))
	source := _install_path(m, "")
	file.MkdirAll(target, ice.MOD_DIR)
	defer m.Echo(target)

	if m.Option(INSTALL) == "" && kit.FileExists(kit.Path(source, "_install")) {
		m.Option(INSTALL, "_install")
	}
	m.Cmd(nfs.DIR, path.Join(source, m.Option(INSTALL))).Tables(func(value ice.Maps) {
		m.Cmd(cli.SYSTEM, "cp", "-r", strings.TrimSuffix(value[nfs.PATH], ice.PS), target+ice.PS)
	})
}
func _install_start(m *ice.Message, arg ...string) {
	p := m.Option(cli.CMD_DIR, m.Cmdx(INSTALL, cli.SPAWN))

	args := []string{}
	switch cb := m.Optionv(PREPARE).(type) {
	case func(string) []string:
		args = append(args, cb(p)...)
	case func(string):
		cb(p)
	default:
		m.ErrorNotImplement(cb)
	}

	if m.Cmdy(cli.DAEMON, arg[1:], args); cli.IsSuccess(m) {
		m.SetAppend()
	}
}
func _install_stop(m *ice.Message, arg ...string) {
	m.Cmd(cli.DAEMON).Tables(func(value ice.Maps) {
		if value[cli.PID] == m.Option(cli.PID) {
			m.Cmd(cli.DAEMON, cli.STOP, kit.Dict(mdb.HASH, value[mdb.HASH]))
		}
	})
	m.Cmd(gdb.SIGNAL, gdb.KILL, m.Option(cli.PID))
}
func _install_service(m *ice.Message, arg ...string) {
	arg = kit.Split(path.Base(arg[0]), "-.")[:1]
	m.Fields(len(arg[1:]), "time,port,status,pid,cmd,dir")
	m.Cmd(mdb.SELECT, cli.DAEMON, "", mdb.HASH).Tables(func(value ice.Maps) {
		if strings.Contains(value[ice.CMD], path.Join(ice.BIN, arg[0])) {
			m.Push("", value, kit.Split(m.OptionFields()))
		}
	})
	m.Set(tcp.PORT).Tables(func(value ice.Maps) { m.Push(tcp.PORT, path.Base(value[nfs.DIR])) })
	m.StatusTimeCount()
}

const (
	PREPARE = "prepare"
)
const INSTALL = "install"

func init() {
	Index.MergeCommands(ice.Commands{
		INSTALL: {Name: "install name port path auto download", Help: "安装", Meta: kit.Dict(), Actions: ice.MergeAction(ice.Actions{
			web.DOWNLOAD: {Name: "download link path", Help: "下载", Hand: func(m *ice.Message, arg ...string) {
				_install_download(m)
			}},
			nfs.SOURCE: {Name: "source link path", Help: "源码", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(nfs.DIR_ROOT, path.Join(_install_path(m, ""), "_install")); !kit.FileExists(m.Option(nfs.DIR_ROOT)) {
					m.Option(nfs.DIR_ROOT, path.Join(_install_path(m, "")))
				}
				defer m.StatusTime(nfs.PATH, m.Option(nfs.DIR_ROOT))
				m.Cmdy(nfs.DIR, m.Option(nfs.PATH))
			}},
			cli.BUILD: {Name: "build link", Help: "构建", Hand: func(m *ice.Message, arg ...string) {
				if err := _install_build(m, arg...); err != "" {
					web.ToastFailure(m, cli.BUILD)
					m.Echo(err)
				} else {
					web.ToastSuccess(m, cli.BUILD)
				}
			}},
			cli.ORDER: {Name: "order link path", Help: "加载", Hand: func(m *ice.Message, arg ...string) {
				_install_order(m, arg...)
			}},
			cli.SPAWN: {Name: "spawn link", Help: "新建", Hand: func(m *ice.Message, arg ...string) {
				_install_spawn(m, arg...)
			}},
			cli.START: {Name: "start link cmd", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
				_install_start(m, arg...)
			}},
			cli.STOP: {Name: "stop", Help: "停止", Hand: func(m *ice.Message, arg ...string) {
				_install_stop(m, arg...)
			}},
		}, mdb.HashAction(mdb.SHORT, mdb.NAME, mdb.FIELD, "time,name,path,link", nfs.PATH, ice.USR_INSTALL)), Hand: func(m *ice.Message, arg ...string) {
			switch len(arg) {
			case 0: // 源码列表
				mdb.HashSelect(m, arg...)

			case 1: // 服务列表
				_install_service(m, arg...)

			default: // 目录列表
				m.Option(nfs.DIR_ROOT, path.Join(ice.USR_LOCAL_DAEMON, arg[1]))
				m.Cmdy(nfs.CAT, kit.Select(nfs.PWD, arg, 2))
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
			m.Cmdy(INSTALL, cli.ORDER, m.Config(nfs.SOURCE), "_install/bin")
		}},
		nfs.TRASH: {Name: "trash", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
			m.Cmd(nfs.TRASH, m.Option(nfs.PATH))
		}},
	}
}
func InstallSoftware(m *ice.Message, bin string, list ice.Any) (ok bool) {
	if cli.SystemFind(m, bin) != "" {
		return true
	}
	kit.Fetch(list, func(index int, value ice.Map) {
		if strings.Contains(m.Cmdx(cli.RUNTIME, kit.Keys(tcp.HOST, cli.OSID)), kit.Format(value[cli.OSID])) {
			web.PushStream(m)
			m.Cmd(cli.SYSTEM, value[ice.CMD])
			ok = true
		}
	})
	return ok
}
