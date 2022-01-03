package code

import (
	"os"
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
	name := path.Base(link)
	file := path.Join(kit.Select(m.Config(nfs.PATH), m.Option(nfs.PATH)), name)

	defer m.Cmdy(nfs.DIR, file)
	if _, e := os.Stat(file); e == nil {
		return // 文件存在
	}

	// 创建文件
	m.Cmd(nfs.SAVE, file, "")

	m.GoToast(web.DOWNLOAD, func(toast func(string, int, int)) {
		// 进度
		m.Cmd(mdb.INSERT, INSTALL, "", mdb.HASH, mdb.NAME, name, mdb.LINK, link)
		m.Richs(INSTALL, "", name, func(key string, value map[string]interface{}) {
			value = kit.GetMeta(value)

			p := 0
			m.OptionCB(web.SPIDE, func(size int, total int) {
				if n := size * 100 / total; p != n {
					value[mdb.VALUE], value[mdb.COUNT], value[mdb.TOTAL] = n, size, total
					toast(name, size, total)
					p = n
				}
			})
		})

		// 下载
		msg := m.Cmd("web.spide", ice.DEV, web.SPIDE_CACHE, web.SPIDE_GET, link)
		m.Cmd(nfs.LINK, file, msg.Append(nfs.FILE))

		// 解压
		m.Option(cli.CMD_DIR, path.Dir(file))
		m.Cmd(cli.SYSTEM, "tar", "xvf", name)
	})
}
func _install_build(m *ice.Message, arg ...string) {
	p := m.Option(cli.CMD_DIR, path.Join(m.Config(nfs.PATH), kit.TrimExt(m.Option(mdb.LINK))))
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
			m.Echo(msg.Append(cli.CMD_ERR))
			m.Toast(ice.FAILURE, cli.BUILD)
			return
		}
	}

	// 编译
	if msg := m.Cmd(cli.SYSTEM, "make", "-j8"); !cli.IsSuccess(msg) {
		m.Echo(msg.Append(cli.CMD_ERR))
		m.Toast(ice.FAILURE, cli.BUILD)
		return
	}

	// 安装
	if msg := m.Cmd(cli.SYSTEM, "make", "PREFIX="+pp, "install"); !cli.IsSuccess(msg) {
		m.Echo(msg.Append(cli.CMD_ERR))
		m.Toast(ice.FAILURE, cli.BUILD)
		return
	}

	m.Toast(ice.SUCCESS, cli.BUILD)
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
		p := path.Join(m.Conf(cli.DAEMON, kit.Keym(nfs.PATH)), m.Option(tcp.PORT))
		if _, e := os.Stat(p); e == nil {
			m.Echo(p)
			return
		}
	} else {
		m.Option(tcp.PORT, m.Cmdx(tcp.PORT, aaa.RIGHT))
	}

	target := path.Join(m.Conf(cli.DAEMON, kit.Keym(nfs.PATH)), m.Option(tcp.PORT))
	source := path.Join(m.Config(nfs.PATH), kit.TrimExt(m.Option(mdb.LINK)))

	m.Cmd(nfs.DIR, path.Join(source, kit.Select("_install", m.Option("install")))).Table(func(index int, value map[string]string, head []string) {
		m.Cmd(cli.SYSTEM, "cp", "-r", strings.TrimSuffix(value[nfs.PATH], ice.PS), target)
	})
	m.Echo(target)
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
func _install_package(m *ice.Message, arg ...string) {
	m.Fields(len(arg), "time,name,path")
	m.Cmdy(mdb.SELECT, INSTALL, "", mdb.HASH)
}
func _install_service(m *ice.Message, arg ...string) {
	arg = kit.Split(path.Base(arg[0]), "-.")[:1]

	m.Fields(len(arg[1:]), "time,port,status,pid,cmd,dir")
	m.Cmd(mdb.SELECT, cli.DAEMON, "", mdb.HASH).Table(func(index int, value map[string]string, head []string) {
		if strings.Contains(value[ice.CMD], "bin/"+arg[0]) {
			m.Push("", value, kit.Split(m.Option(mdb.FIELDS)))
		}
	})

	m.Appendv(tcp.PORT, []string{})
	m.Table(func(index int, value map[string]string, head []string) {
		m.Push(tcp.PORT, path.Base(value[nfs.DIR]))
	})
}

const (
	PREPARE = "prepare"
)
const INSTALL = "install"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		INSTALL: {Name: INSTALL, Help: "安装", Value: kit.Data(
			mdb.SHORT, mdb.NAME, nfs.PATH, ice.USR_INSTALL,
		)},
	}, Commands: map[string]*ice.Command{
		INSTALL: {Name: "install name port path auto download", Help: "安装", Meta: kit.Dict(), Action: map[string]*ice.Action{
			web.DOWNLOAD: {Name: "download link path", Help: "下载", Hand: func(m *ice.Message, arg ...string) {
				_install_download(m)
			}},
			cli.BUILD: {Name: "build link", Help: "构建", Hand: func(m *ice.Message, arg ...string) {
				_install_build(m, arg...)
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
			nfs.SOURCE: {Name: "source link path", Help: "源码", Hand: func(m *ice.Message, arg ...string) {
				m.Option(nfs.DIR_ROOT, path.Join(m.Config(nfs.PATH), kit.TrimExt(m.Option(mdb.LINK)), "_install"))
				defer m.StatusTime(nfs.PATH, m.Option(nfs.DIR_ROOT))
				m.Cmdy(nfs.DIR, m.Option(nfs.PATH))
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch len(arg) {
			case 0: // 源码列表
				_install_package(m, arg...)

			case 1: // 服务列表
				_install_service(m, arg...)

			default: // 目录列表
				m.Option(nfs.DIR_ROOT, path.Join(m.Conf(cli.DAEMON, kit.Keym(nfs.PATH)), arg[1]))
				m.Cmdy(nfs.CAT, kit.Select(ice.PWD, arg, 2))
			}
		}},
	}})
}

func InstallAction(fields ...string) map[string]*ice.Action {
	return ice.SelectAction(map[string]*ice.Action{
		web.DOWNLOAD: {Name: "download", Help: "下载", Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(INSTALL, web.DOWNLOAD, m.Config(nfs.SOURCE))
		}},
		cli.BUILD: {Name: "build", Help: "构建", Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(INSTALL, cli.BUILD, m.Config(nfs.SOURCE))
		}},
		cli.ORDER: {Name: "order", Help: "加载", Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(INSTALL, cli.ORDER, m.Config(nfs.SOURCE), "_install/bin")
		}},
	}, fields...)
}
