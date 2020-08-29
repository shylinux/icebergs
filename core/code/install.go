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
			)},
		},
		Commands: map[string]*ice.Command{
			INSTALL: {Name: "install name=auto port=auto path=auto auto", Help: "安装", Meta: kit.Dict(), Action: map[string]*ice.Action{
				"download": {Name: "download link", Help: "下载", Hand: func(m *ice.Message, arg ...string) {
					name := path.Base(arg[0])
					if m.Richs(INSTALL, "", name, func(key string, value map[string]interface{}) {
						if _, e := os.Stat(path.Join(m.Conf(INSTALL, kit.META_PATH), kit.Format(value[kit.MDB_NAME]))); e == nil {
							m.Push(key, value, []string{kit.MDB_TIME, kit.MDB_STEP, kit.MDB_SIZE, kit.MDB_NAME, kit.MDB_LINK})
						}
					}) != nil && len(m.Appendv(kit.MDB_NAME)) > 0 {
						// 已经下载
						return
					}

					// 进度
					m.Cmd(mdb.INSERT, m.Prefix(INSTALL), "", mdb.HASH, kit.MDB_NAME, name, kit.MDB_LINK, arg[0])
					m.Richs(INSTALL, "", name, func(key string, value map[string]interface{}) {
						m.Optionv("progress", func(size int, total int) {
							s := size * 100 / total
							if s != kit.Int(value[kit.MDB_STEP]) && s%10 == 0 {
								m.Log_IMPORT(kit.MDB_FILE, name, kit.MDB_STEP, s, kit.MDB_SIZE, kit.FmtSize(int64(size)), kit.MDB_TOTAL, kit.FmtSize(int64(total)))
							}
							value[kit.MDB_STEP], value[kit.MDB_SIZE], value[kit.MDB_TOTAL] = s, size, total
						})
					})

					// 占位
					p := path.Join(m.Conf(INSTALL, kit.META_PATH), name)
					m.Cmd(cli.SYSTEM, "touch", p)

					// 代理
					to := m.Cmd("web.spide_rewrite", arg[0]).Append("to")
					m.Debug("to: %s", to)
					arg[0] = kit.Select(arg[0], to)

					// 下载
					m.Option(cli.CMD_DIR, m.Conf(INSTALL, kit.META_PATH))
					if strings.HasPrefix(arg[0], "ftp") {
						m.Cmdy(cli.SYSTEM, "wget", arg[0])
					} else {
						msg := m.Cmd(web.SPIDE, web.SPIDE_DEV, web.SPIDE_CACHE, web.SPIDE_GET, arg[0])
						m.Cmdy(nfs.LINK, p, msg.Append(kit.MDB_FILE))
					}

					// 解压
					m.Cmd(cli.SYSTEM, "tar", "xvf", name)
				}},
				"build": {Name: "build link", Help: "构建", Hand: func(m *ice.Message, arg ...string) {
					p := m.Option(cli.CMD_DIR, path.Join(m.Conf(INSTALL, kit.META_PATH), kit.TrimExt(arg[0])))
					switch cb := m.Optionv("prepare").(type) {
					case func(string):
						cb(p)
					default:
						m.Cmdy(cli.SYSTEM, "./configure", "--prefix="+kit.Path(path.Join(p, INSTALL)), arg[1:])
					}

					m.Cmdy(cli.SYSTEM, "make", "-j8")
					m.Cmdy(cli.SYSTEM, "make", "PREFIX="+kit.Path(path.Join(p, INSTALL)), "install")
				}},
				"spawn": {Name: "spawn link", Help: "新建", Hand: func(m *ice.Message, arg ...string) {
					port := m.Cmdx(tcp.PORT, "select")
					target := path.Join(m.Conf(cli.DAEMON, kit.META_PATH), port)
					source := path.Join(m.Conf(INSTALL, kit.META_PATH), kit.TrimExt(arg[0]))

					m.Cmd(nfs.DIR, path.Join(source, kit.Select("install", m.Option("install")))).Table(func(index int, value map[string]string, head []string) {
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
					m.Option(mdb.FIELDS, "time,step,size,total,name,link")
					m.Cmdy(mdb.SELECT, m.Prefix(INSTALL), "", mdb.HASH)
					m.Sort(kit.MDB_TIME, "time_r")
					return
				}

				arg[0] = path.Base(arg[0])
				if key := strings.Split(strings.Split(arg[0], "-")[0], ".")[0]; len(arg) == 1 {
					// 服务列表
					u := kit.ParseURL(m.Option(ice.MSG_USERWEB))
					m.Cmd(cli.DAEMON).Table(func(index int, value map[string]string, head []string) {
						if strings.Contains(value[kit.MDB_NAME], key) {
							m.Push(kit.MDB_TIME, value[kit.MDB_TIME])
							m.Push(kit.MDB_PORT, path.Base(value[kit.MDB_DIR]))
							m.Push(kit.MDB_STATUS, value[kit.MDB_STATUS])
							m.Push(kit.MDB_NAME, value[kit.MDB_NAME])
							m.Push(kit.MDB_LINK, m.Cmdx(mdb.RENDER, web.RENDER.A,
								kit.Format("http://%s:%s", u.Hostname(), path.Base(value[kit.MDB_DIR]))))
						}
					})
					m.Sort(kit.MDB_TIME, "time_r")
					return
				}

				// 目录列表
				m.Option(nfs.DIR_ROOT, path.Join(m.Conf(cli.DAEMON, kit.META_PATH), arg[1]))
				m.Cmdy(nfs.DIR, kit.Select("./", arg, 2))
				m.Sort(kit.MDB_TIME, "time_r")
			}},
		},
	}, nil)
}
