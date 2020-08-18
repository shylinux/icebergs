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
			INSTALL: {Name: "install", Help: "安装", Value: kit.Data(
				kit.MDB_SHORT, kit.MDB_NAME, kit.MDB_PATH, "usr/install",
			)},
		},
		Commands: map[string]*ice.Command{
			INSTALL: {Name: "install name=auto auto", Help: "安装", Action: map[string]*ice.Action{
				"download": {Name: "download link", Help: "下载", Hand: func(m *ice.Message, arg ...string) {
					name := path.Base(arg[0])
					if m.Richs(INSTALL, "", name, func(key string, value map[string]interface{}) {
						if _, e := os.Stat(path.Join(m.Conf(INSTALL, kit.META_PATH), kit.Format(value["name"]))); e == nil {
							m.Push(key, value, []string{"time", "progress", "size", "name", "link"})
						}
					}) != nil && len(m.Appendv("name")) > 0 {
						// 查询
						return
					}

					// 进度
					m.Cmd(mdb.INSERT, m.Prefix(INSTALL), "", mdb.HASH, kit.MDB_NAME, name, kit.MDB_LINK, arg[0])
					m.Richs(INSTALL, "", name, func(key string, value map[string]interface{}) {
						m.Optionv("progress", func(size int, total int) {
							p := size * 100 / total
							if p != kit.Int(value["progress"]) && p%10 == 0 {
								m.Log_IMPORT(kit.MDB_FILE, name, "per", size*100/total, kit.MDB_SIZE, kit.FmtSize(int64(size)), "total", kit.FmtSize(int64(total)))
							}
							value["progress"], value["size"], value["total"] = p, size, total
						})
					})

					// 下载
					m.Option(cli.CMD_DIR, m.Conf(INSTALL, kit.META_PATH))
					if strings.HasPrefix(arg[0], "ftp") {
						m.Cmdy(cli.SYSTEM, "wget", arg[0])
					} else {
						msg := m.Cmd(web.SPIDE, web.SPIDE_DEV, web.SPIDE_CACHE, web.SPIDE_GET, arg[0])
						p := path.Join(m.Conf(INSTALL, kit.META_PATH), name)
						m.Cmdy(nfs.LINK, p, msg.Append("file"))
					}

					// 解压
					m.Cmd(cli.SYSTEM, "tar", "xvf", name)
				}},
				"start": {Name: "start source binary", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
					port := m.Cmdx(tcp.PORT, "get")
					p := path.Join(m.Conf(cli.DAEMON, kit.META_PATH), port)
					os.MkdirAll(p, ice.MOD_DIR)

					// 复制
					m.Cmd(nfs.DIR, path.Join(m.Conf(INSTALL, kit.META_PATH), arg[0])).Table(func(index int, value map[string]string, head []string) {
						m.Cmd(cli.SYSTEM, "cp", "-r", strings.TrimSuffix(value[kit.MDB_PATH], "/"), p)
					})

					// 启动
					m.Option(cli.CMD_DIR, p)
					m.Cmdy(cli.DAEMON, arg[1])
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) > 0 {
					// 详情
					m.Cmdy(mdb.SELECT, m.Prefix(INSTALL), "", mdb.HASH, kit.MDB_NAME, arg[0])
					return
				}

				// 列表
				m.Option("fields", "time,progress,size,total,name,link")
				m.Cmdy(mdb.SELECT, m.Prefix(INSTALL), "", mdb.HASH)
				m.Sort(kit.MDB_TIME, "time_r")
			}},
		},
	}, nil)
}
