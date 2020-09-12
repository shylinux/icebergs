package code

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/base/tcp"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"

	"os"
	"path"
	"strings"
)

const INSTALL = "install"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			INSTALL: {Name: INSTALL, Help: "安装", Value: kit.Data(
				kit.MDB_SHORT, kit.MDB_NAME, kit.MDB_PATH, "usr/install",
				kit.MDB_FIELD, "time,step,size,total,name,link",
			)},
		},
		Commands: map[string]*ice.Command{
			INSTALL: {Name: "install name port path auto", Help: "安装", Meta: kit.Dict(), Action: map[string]*ice.Action{
				"download": {Name: "download link", Help: "下载", Hand: func(m *ice.Message, arg ...string) {
					name := path.Base(arg[0])
					p := path.Join(m.Conf(INSTALL, kit.META_PATH), name)

					m.Option("_process", "_progress")
					m.Option(mdb.FIELDS, m.Conf(INSTALL, kit.META_FIELD))
					if m.Cmd(mdb.SELECT, m.Prefix(INSTALL), "", mdb.HASH, kit.MDB_NAME, name).Table(func(index int, value map[string]string, head []string) {
						if _, e := os.Stat(p); e == nil {
							m.Push("", value, kit.Split(m.Option(mdb.FIELDS)))
						}
					}); len(m.Appendv(kit.MDB_NAME)) > 0 {
						return // 已经下载
					}

					// 占位
					m.Cmd(nfs.SAVE, p, "")

					// 进度
					m.Cmd(mdb.INSERT, m.Prefix(INSTALL), "", mdb.HASH, kit.MDB_NAME, name, kit.MDB_LINK, arg[0])
					m.Richs(INSTALL, "", name, func(key string, value map[string]interface{}) {
						value = value[kit.MDB_META].(map[string]interface{})
						m.Optionv("progress", func(size int, total int) {
							s := size * 100 / total
							if s != kit.Int(value[kit.MDB_STEP]) && s%10 == 0 {
								m.Log_IMPORT(kit.MDB_FILE, name, kit.MDB_STEP, s, kit.MDB_SIZE, kit.FmtSize(int64(size)), kit.MDB_TOTAL, kit.FmtSize(int64(total)))
							}
							value[kit.MDB_STEP], value[kit.MDB_SIZE], value[kit.MDB_TOTAL] = s, size, total
						})
					})

					m.Gos(m, func(m *ice.Message) {
						// 下载
						m.Option(cli.CMD_DIR, m.Conf(INSTALL, kit.META_PATH))
						msg := m.Cmd(web.SPIDE, web.SPIDE_DEV, web.SPIDE_CACHE, web.SPIDE_GET, arg[0])
						m.Cmdy(nfs.LINK, p, msg.Append(kit.MDB_FILE))

						// 解压
						m.Cmd(cli.SYSTEM, "tar", "xvf", name)
					})
				}},
				"build": {Name: "build link", Help: "构建", Hand: func(m *ice.Message, arg ...string) {
					p := m.Option(cli.CMD_DIR, path.Join(m.Conf(INSTALL, kit.META_PATH), kit.TrimExt(arg[0])))
					switch cb := m.Optionv("prepare").(type) {
					case func(string):
						cb(p)
					default:
						if m.Cmdy(cli.SYSTEM, "./configure", "--prefix="+kit.Path(path.Join(p, kit.Select("_install", m.Option("install")))), arg[1:]); m.Append(cli.CMD_CODE) != "0" {
							return
						}
					}

					if m.Cmdy(cli.SYSTEM, "make", "-j8"); m.Append(cli.CMD_CODE) != "0" {
						return
					}

					m.Cmdy(cli.SYSTEM, "mv", "INSTALL", "INSTALLS")
					m.Cmdy(cli.SYSTEM, "make", "PREFIX="+kit.Path(path.Join(p, kit.Select("_install", m.Option("install")))), "install")
				}},
				"spawn": {Name: "spawn link", Help: "新建", Hand: func(m *ice.Message, arg ...string) {
					port := m.Cmdx(tcp.PORT, "select")
					target := path.Join(m.Conf(cli.DAEMON, kit.META_PATH), port)
					source := path.Join(m.Conf(INSTALL, kit.META_PATH), kit.TrimExt(arg[0]))

					m.Cmd(nfs.DIR, path.Join(source, kit.Select("_install", m.Option("install")))).Table(func(index int, value map[string]string, head []string) {
						m.Cmd(cli.SYSTEM, "cp", "-r", strings.TrimSuffix(value[kit.MDB_PATH], "/"), target)
					})
					m.Echo(target)
				}},
				"start": {Name: "start link cmd...", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
					p := m.Option(cli.CMD_DIR, m.Cmdx(INSTALL, "spawn", arg[0]))

					args := []string{}
					switch cb := m.Optionv("prepare").(type) {
					case func(string) []string:
						args = append(args, cb(p)...)
					}

					m.Cmdy(cli.DAEMON, arg[1:], args)
				}},
				"bench": {Name: "bench port cmd...", Help: "压测", Hand: func(m *ice.Message, arg ...string) {
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					// 源码列表
					m.Option(mdb.FIELDS, m.Conf(INSTALL, kit.META_FIELD))
					m.Cmdy(mdb.SELECT, m.Prefix(INSTALL), "", mdb.HASH)
					return
				}

				arg[0] = path.Base(arg[0])
				if key := strings.Split(strings.Split(arg[0], "-")[0], ".")[0]; len(arg) == 1 {
					u := kit.ParseURL(m.Option(ice.MSG_USERWEB))
					m.Cmd(cli.DAEMON).Table(func(index int, value map[string]string, head []string) {
						// 服务列表
						if strings.Contains(value[kit.MDB_NAME], key) {
							m.Push(kit.MDB_TIME, value[kit.MDB_TIME])
							m.Push(kit.MDB_PORT, path.Base(value[kit.MDB_DIR]))
							m.Push(kit.MDB_STATUS, value[kit.MDB_STATUS])
							m.Push(kit.MDB_NAME, value[kit.MDB_NAME])
							m.PushRender(kit.MDB_LINK, "a", kit.Format("http://%s:%s", u.Hostname(), path.Base(value[kit.MDB_DIR])))
						}
					})
					return
				}

				// 目录列表
				m.Option(nfs.DIR_ROOT, path.Join(m.Conf(cli.DAEMON, kit.META_PATH), arg[1]))
				m.Cmdy(nfs.DIR, kit.Select("./", arg, 2))
			}},
		},
	}, nil)
}
