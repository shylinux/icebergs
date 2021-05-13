package code

import (
	"os"
	"path"
	"strings"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/gdb"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/base/tcp"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"
)

type Buffer struct {
	data []byte
	m    *ice.Message
	n    string
}

func (b *Buffer) Write(buf []byte) (int, error) {
	b.data = append(b.data, buf...)
	b.m.Cmd(web.SPACE, b.n, "grow", string(buf))
	return len(buf), nil
}
func (b *Buffer) Close() error {
	return nil
}

const PREPARE = "prepare"
const INSTALL = "install"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			INSTALL: {Name: INSTALL, Help: "安装", Value: kit.Data(
				kit.MDB_SHORT, kit.MDB_NAME, kit.MDB_PATH, ice.USR_INSTALL,
			)},
		},
		Commands: map[string]*ice.Command{
			INSTALL: {Name: "install name port path auto download", Help: "安装", Meta: kit.Dict(), Action: map[string]*ice.Action{
				web.DOWNLOAD: {Name: "download link", Help: "下载", Hand: func(m *ice.Message, arg ...string) {
					link := m.Option(kit.MDB_LINK)
					name := path.Base(link)
					file := path.Join(m.Conf(INSTALL, kit.META_PATH), name)

					defer m.Cmdy(nfs.DIR, file)
					if _, e := os.Stat(file); e == nil {
						return // 文件存在
					}

					// 文件占位
					m.Cmd(nfs.SAVE, file, "")

					m.GoToast("download", func(toast func(string, int, int)) {
						// 进度
						m.Cmd(mdb.INSERT, INSTALL, "", mdb.HASH, kit.MDB_NAME, name, kit.MDB_LINK, link)
						m.Richs(INSTALL, "", name, func(key string, value map[string]interface{}) {
							value = kit.GetMeta(value)

							p := 0
							m.Optionv(web.DOWNLOAD_CB, func(size int, total int) {
								if n := size * 100 / total; p != n {
									value[kit.SSH_STEP], value[kit.MDB_SIZE], value[kit.MDB_TOTAL] = n, size, total
									toast(name, size, total)
									p = n
								}
							})
						})

						// 下载
						os.MkdirAll(m.Option(cli.CMD_DIR, m.Conf(INSTALL, kit.META_PATH)), ice.MOD_DIR)
						msg := m.Cmd(web.SPIDE, web.SPIDE_DEV, web.SPIDE_CACHE, web.SPIDE_GET, link)

						// 解压
						m.Cmdy(nfs.LINK, file, msg.Append(kit.MDB_FILE))
						m.Cmd(cli.SYSTEM, "tar", "xvf", name)
					})
				}},
				gdb.BUILD: {Name: "build link", Help: "构建", Hand: func(m *ice.Message, arg ...string) {

					m.Option(cli.CMD_OUTPUT, &Buffer{m: m, n: m.Option(ice.MSG_DAEMON)})
					defer func() {
						m.Toast("success", "build")
						m.ProcessHold()
					}()

					p := m.Option(cli.CMD_DIR, path.Join(m.Conf(INSTALL, kit.META_PATH), kit.TrimExt(m.Option(kit.MDB_LINK))))
					pp := kit.Path(path.Join(p, "_install"))
					switch cb := m.Optionv("prepare").(type) {
					case func(string):
						cb(p)
					default:
						if m.Cmd(cli.SYSTEM, "./configure", "--prefix="+pp, arg[1:]).Append(cli.CMD_CODE) != "0" {
							return
						}
					}

					if m.Cmd(cli.SYSTEM, "make", "-j8"); m.Append(cli.CMD_CODE) != "0" {
						return
					}

					m.Cmd(cli.SYSTEM, "make", "PREFIX="+pp, "install")
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
	})
}
