package code

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _install_download(m *ice.Message) {
	link := m.Option(mdb.LINK)
	name := path.Base(strings.Split(link, "?")[0])
	file := path.Join(kit.Select(m.Config(nfs.PATH), m.Option(nfs.PATH)), name)

	defer m.Cmdy(nfs.DIR, file)
	if kit.FileExists(file) {
		return // 文件存在
	}

	// 创建文件
	m.Cmd(nfs.SAVE, file, "")
	m.GoToast(web.DOWNLOAD, func(toast func(string, int, int)) {
		m.Cmd(mdb.INSERT, INSTALL, "", mdb.HASH, mdb.NAME, name, nfs.PATH, file, mdb.LINK, link)
		defer m.ToastSuccess()

		// 下载进度
		m.Richs(INSTALL, "", name, func(key string, value map[string]interface{}) {
			prev, value := 0, kit.GetMeta(value)
			m.OptionCB(web.SPIDE, func(count int, total int, step int) {
				if step >= prev {
					value[mdb.COUNT], value[mdb.TOTAL], value[mdb.VALUE] = count, total, step
					toast(name, count, total)
					prev = step
				}
			})
		})

		// 下载解压
		m.Cmd("web.spide", ice.DEV, web.SPIDE_SAVE, file, web.SPIDE_GET, link)
		m.Cmd(nfs.TAR, mdb.EXPORT, name, kit.Dict(cli.CMD_DIR, path.Dir(file)))
	})
}
func _install_build(m *ice.Message, arg ...string) string {
	p := m.Option(cli.CMD_DIR, path.Join(m.Config(nfs.PATH), kit.TrimExt(m.Option(mdb.LINK))))
	pp := kit.Path(path.Join(p, "_install"))

	// 推流
	cli.PushStream(m)
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
	p := path.Join(m.Config(nfs.PATH), kit.TrimExt(m.Option(mdb.LINK)), m.Option(nfs.PATH)+ice.NL)
	if !strings.Contains(m.Cmdx(nfs.CAT, ice.ETC_PATH), p) {
		m.Cmd(nfs.PUSH, ice.ETC_PATH, p)
	}
	m.Cmdy(nfs.CAT, ice.ETC_PATH)
}
func _install_spawn(m *ice.Message, arg ...string) {
	if kit.Int(m.Option(tcp.PORT)) >= 10000 {
		if p := path.Join(m.Conf(cli.DAEMON, kit.Keym(nfs.PATH)), m.Option(tcp.PORT)); kit.FileExists(p) {
			m.Echo(p)
			return
		}
	} else {
		m.Option(tcp.PORT, m.Cmdx(tcp.PORT, aaa.RIGHT))
	}

	target := path.Join(m.Conf(cli.DAEMON, kit.Keym(nfs.PATH)), m.Option(tcp.PORT))
	source := path.Join(m.Config(nfs.PATH), kit.TrimExt(m.Option(mdb.LINK)))
	nfs.MkdirAll(m, target)
	defer m.Echo(target)

	if m.Option(INSTALL) == "" && kit.FileExists(kit.Path(source, "_install")) {
		m.Option(INSTALL, "_install")
	}
	m.Cmd(nfs.DIR, path.Join(source, m.Option(INSTALL))).Tables(func(value map[string]string) {
		m.Cmd(cli.SYSTEM, "cp", "-r", strings.TrimSuffix(value[nfs.PATH], ice.PS), target+ice.PS)
	})
}
func _install_start(m *ice.Message, arg ...string) {
	p := m.Option(cli.CMD_DIR, m.Cmdx(INSTALL, cli.SPAWN))

	args := []string{}
	switch cb := m.Optionv(PREPARE).(type) {
	case func(string) []string:
		args = append(args, cb(p)...)
	}

	m.Cmdy(cli.DAEMON, arg[1:], args)
}
func _install_service(m *ice.Message, arg ...string) {
	arg = kit.Split(path.Base(arg[0]), "-.")[:1]
	m.Fields(len(arg[1:]), "time,port,status,pid,cmd,dir")
	m.Cmd(mdb.SELECT, cli.DAEMON, "", mdb.HASH).Tables(func(value map[string]string) {
		if strings.Contains(value[ice.CMD], path.Join(ice.BIN, arg[0])) {
			m.Push("", value, kit.Split(m.OptionFields()))
		}
	})
	m.Set(tcp.PORT).Tables(func(value map[string]string) { m.Push(tcp.PORT, path.Base(value[nfs.DIR])) })
}
func _install_stop(m *ice.Message, arg ...string) {
	m.Cmd(cli.DAEMON).Tables(func(value map[string]string) {
		if value[cli.PID] == m.Option(cli.PID) {
			m.Cmd(cli.DAEMON, cli.STOP, kit.Dict(mdb.HASH, value[mdb.HASH]))
		}
	})
	m.Cmd(cli.SYSTEM, cli.KILL, m.Option(cli.PID))
}

const (
	PREPARE = "prepare"
)
const INSTALL = "install"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		INSTALL: {Name: INSTALL, Help: "安装", Value: kit.Data(
			mdb.SHORT, mdb.NAME, mdb.FIELD, "time,name,path,link", nfs.PATH, ice.USR_INSTALL,
		)},
	}, Commands: map[string]*ice.Command{
		INSTALL: {Name: "install name port path auto download", Help: "安装", Meta: kit.Dict(), Action: ice.MergeAction(map[string]*ice.Action{
			web.DOWNLOAD: {Name: "download link path", Help: "下载", Hand: func(m *ice.Message, arg ...string) {
				_install_download(m)
			}},
			nfs.SOURCE: {Name: "source link path", Help: "源码", Hand: func(m *ice.Message, arg ...string) {
				m.Option(nfs.DIR_ROOT, path.Join(m.Config(nfs.PATH), kit.TrimExt(m.Option(mdb.LINK)), "_install"))
				defer m.StatusTime(nfs.PATH, m.Option(nfs.DIR_ROOT))
				m.Cmdy(nfs.DIR, m.Option(nfs.PATH))
			}},
			cli.BUILD: {Name: "build link", Help: "构建", Hand: func(m *ice.Message, arg ...string) {
				if err := _install_build(m, arg...); err != "" {
					m.ToastFailure(cli.BUILD)
					m.Echo(err)
				} else {
					m.ToastSuccess(cli.BUILD)
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
		}, mdb.HashAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch len(arg) {
			case 0: // 源码列表
				mdb.HashSelect(m, arg...)

			case 1: // 服务列表
				_install_service(m, arg...)

			default: // 目录列表
				m.Option(nfs.DIR_ROOT, path.Join(m.Conf(cli.DAEMON, kit.Keym(nfs.PATH)), arg[1]))
				m.Cmdy(nfs.CAT, kit.Select(nfs.PWD, arg, 2))
			}
		}},
	}})
}

func InstallSoftware(m *ice.Message, bin string, list interface{}) (ok bool) {
	if cli.SystemFind(m, bin) != "" {
		return true
	}
	kit.Fetch(list, func(index int, value map[string]interface{}) {
		if strings.Contains(m.Cmdx(cli.RUNTIME, kit.Keys(tcp.HOST, cli.OSID)), kit.Format(value[cli.OSID])) {
			cli.PushStream(m)
			m.Cmd(cli.SYSTEM, value[ice.CMD])
			ok = true
		}
	})
	return ok
}
func InstallAction(args ...interface{}) map[string]*ice.Action {
	return ice.SelectAction(map[string]*ice.Action{ice.CTX_INIT: mdb.AutoConfig(args...),
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
	})
}
