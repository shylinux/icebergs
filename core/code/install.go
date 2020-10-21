package code

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/gdb"
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
				web.DOWNLOAD: {Name: "download link", Help: "下载", Hand: func(m *ice.Message, arg ...string) {
					link := m.Option(kit.MDB_LINK)
					name := path.Base(link)
					p := path.Join(m.Conf(INSTALL, kit.META_PATH), name)

					m.Option(ice.MSG_PROCESS, "_progress")
					m.Option(mdb.FIELDS, m.Conf(INSTALL, kit.META_FIELD))
					if m.Cmd(mdb.SELECT, INSTALL, "", mdb.HASH, kit.MDB_NAME, name).Table(func(index int, value map[string]string, head []string) {
						if _, e := os.Stat(p); e == nil {
							m.Push("", value, kit.Split(m.Option(mdb.FIELDS)))
						}
					}); len(m.Appendv(kit.MDB_NAME)) > 0 {
						return // 已经下载
					}

					// 占位
					m.Cmd(nfs.SAVE, p, "")

					// 进度
					m.Cmd(mdb.INSERT, INSTALL, "", mdb.HASH, kit.MDB_NAME, name, kit.MDB_LINK, link)
					m.Richs(INSTALL, "", name, func(key string, value map[string]interface{}) {
						value = kit.GetMeta(value)
						m.Optionv(web.DOWNLOAD_CB, func(size int, total int) {
							s := size * 100 / total
							if s != kit.Int(value[kit.SSH_STEP]) && s%10 == 0 {
								m.Log_IMPORT(kit.MDB_FILE, name, kit.SSH_STEP, s, kit.MDB_SIZE, kit.FmtSize(int64(size)), kit.MDB_TOTAL, kit.FmtSize(int64(total)))
							}
							value[kit.SSH_STEP], value[kit.MDB_SIZE], value[kit.MDB_TOTAL] = s, size, total
						})
					})

					// 下载
					m.Go(func() {
						m.Option(cli.CMD_DIR, m.Conf(INSTALL, kit.META_PATH))
						msg := m.Cmd(web.SPIDE, web.SPIDE_DEV, web.SPIDE_CACHE, web.SPIDE_GET, link)

						m.Cmdy(nfs.LINK, p, msg.Append(kit.MDB_FILE))
						m.Cmd(cli.SYSTEM, "tar", "xvf", name)
					})
				}},
				gdb.BUILD: {Name: "build link", Help: "构建", Hand: func(m *ice.Message, arg ...string) {
					if cli.Follow(m) {
						return
					}

					m.Go(func() {
						defer m.Cmdy(cli.OUTPUT, mdb.MODIFY, kit.MDB_STATUS, cli.Status.Stop)
						defer m.Option(kit.MDB_HASH, m.Option("cache.hash"))

						p := m.Option(cli.CMD_DIR, path.Join(m.Conf(INSTALL, kit.META_PATH), kit.TrimExt(m.Option(kit.MDB_LINK))))
						pp := kit.Path(path.Join(p, "_install"))
						switch cb := m.Optionv("prepare").(type) {
						case func(string):
							cb(p)
						default:
							if m.Cmdy(cli.SYSTEM, "./configure", "--prefix="+pp, arg[1:]); m.Append(cli.CMD_CODE) != "0" {
								return
							}
						}

						if m.Cmdy(cli.SYSTEM, "make", "-j8"); m.Append(cli.CMD_CODE) != "0" {
							return
						}

						m.Cmdy(cli.SYSTEM, "make", "PREFIX="+pp, "install")
					})
				}},
				gdb.SPAWN: {Name: "spawn link", Help: "新建", Hand: func(m *ice.Message, arg ...string) {
					port := m.Cmdx(tcp.PORT, aaa.RIGHT)
					target := path.Join(m.Conf(cli.DAEMON, kit.META_PATH), port)
					source := path.Join(m.Conf(INSTALL, kit.META_PATH), kit.TrimExt(m.Option(kit.MDB_LINK)))

					m.Cmd(nfs.DIR, path.Join(source, kit.Select("_install", m.Option("install")))).Table(func(index int, value map[string]string, head []string) {
						m.Cmd(cli.SYSTEM, "cp", "-r", strings.TrimSuffix(value[kit.MDB_PATH], "/"), target)
					})
					m.Echo(target)
				}},
				gdb.START: {Name: "start link cmd", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
					p := m.Option(cli.CMD_DIR, m.Cmdx(INSTALL, gdb.SPAWN))

					args := []string{}
					switch cb := m.Optionv("prepare").(type) {
					case func(string) []string:
						args = append(args, cb(p)...)
					}

					m.Cmdy(cli.DAEMON, arg[1:], args)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					// 源码列表
					m.Option(mdb.FIELDS, m.Conf(INSTALL, kit.META_FIELD))
					m.Cmdy(mdb.SELECT, INSTALL, "", mdb.HASH)
					return
				}

				if len(arg) == 1 {
					// 服务列表
					arg = kit.Split(path.Base(arg[0]), "-.")
					m.Option(mdb.FIELDS, "time,port,status,pid,cmd,dir")
					m.Cmd(mdb.SELECT, cli.DAEMON, "", mdb.HASH).Table(func(index int, value map[string]string, head []string) {
						if strings.Contains(value["cmd"], "bin/"+arg[0]) {
							m.Push("", value, kit.Split(m.Option(mdb.FIELDS)))
						}
					})
					m.Appendv(kit.SSH_PORT, []string{})
					m.Table(func(index int, value map[string]string, head []string) {
						m.Push(kit.SSH_PORT, path.Base(value[kit.SSH_DIR]))
					})
					return
				}

				// 目录列表
				m.Option(nfs.DIR_ROOT, path.Join(m.Conf(cli.DAEMON, kit.META_PATH), arg[1]))
				if strings.HasSuffix(kit.Select("./", arg, 2), "/") {
					m.Cmdy(nfs.DIR, kit.Select("./", arg, 2))
				} else {
					m.Cmdy(nfs.CAT, kit.Select("./", arg, 2))
				}
			}},
		},
	}, nil)
}
