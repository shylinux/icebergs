package code

import (
	"net/http"
	"path"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"
)

const INSTALL = "install"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			INSTALL: {Name: "install", Help: "安装", Value: kit.Data(
				kit.MDB_SHORT, kit.MDB_NAME, "path", "usr/install",
			)},
		},
		Commands: map[string]*ice.Command{
			INSTALL: {Name: "install name=auto auto", Help: "安装", Action: map[string]*ice.Action{
				"download": {Name: "download link", Help: "下载", Hand: func(m *ice.Message, arg ...string) {
					name := path.Base(arg[0])
					if m.Richs(INSTALL, "", name, func(key string, value map[string]interface{}) {
						m.Push(key, value, []string{"time", "progress", "size", "name", "link"})
					}) != nil {
						return
					}

					m.Cmd(mdb.INSERT, m.Prefix(INSTALL), "", mdb.HASH, kit.MDB_NAME, name, kit.MDB_LINK, arg[0])
					m.Richs(INSTALL, "", name, func(key string, value map[string]interface{}) {
						m.Optionv("progress", func(size int, total int) {
							value["progress"], value["size"], value["total"] = size*100/total, size, total
						})
					})

					msg := m.Cmd(web.SPIDE, "dev", "cache", http.MethodGet, arg[0])
					p := path.Join(m.Conf(INSTALL, "meta.path"), name)
					m.Cmdy(nfs.LINK, p, msg.Append("file"))

					m.Option(cli.CMD_DIR, m.Conf(INSTALL, "meta.path"))
					m.Cmd(cli.SYSTEM, "tar", "xvf", name)
					m.Echo(p)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option("fields", "time,progress,size,total,name,link")
				if len(arg) > 0 {
					m.Cmdy(mdb.SELECT, m.Prefix(INSTALL), "", mdb.HASH, kit.MDB_NAME, arg[0])
					return
				}
				m.Cmdy(mdb.SELECT, m.Prefix(INSTALL), "", mdb.HASH)
				m.Sort(kit.MDB_TIME, "time_r")
			}},
		},
	}, nil)
}
