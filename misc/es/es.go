package es

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/tcp"
	"github.com/shylinux/icebergs/core/code"
	kit "github.com/shylinux/toolkits"

	"os"
	"path"
	"runtime"
	"strings"
)

const ES = "es"

var Index = &ice.Context{Name: ES, Help: "搜索",
	Configs: map[string]*ice.Config{
		ES: {Name: ES, Help: "搜索", Value: kit.Data(
			"windows", "https://elasticsearch.thans.cn/downloads/elasticsearch/elasticsearch-7.3.2-windows-x86_64.zip",
			"darwin", "https://elasticsearch.thans.cn/downloads/elasticsearch/elasticsearch-7.3.2-darwin-x86_64.tar.gz",
			"linux", "https://elasticsearch.thans.cn/downloads/elasticsearch/elasticsearch-7.3.2-linux-x86_64.tar.gz",
		)},
	},
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},

		ES: {Name: "es hash=auto auto 启动:button 安装:button", Help: "搜索", Action: map[string]*ice.Action{
			"install": {Name: "install", Help: "安装", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy("web.code.install", "download", m.Conf(ES, kit.Keys(kit.MDB_META, runtime.GOOS)))
			}},

			"start": {Name: "start", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
				name := path.Base(m.Conf(ES, kit.Keys(kit.MDB_META, runtime.GOOS)))
				name = strings.Join(strings.Split(name, "-")[:2], "-")

				port := m.Cmdx(tcp.PORT, "get")
				p := "var/daemon/" + port
				os.MkdirAll(p, ice.MOD_DIR)
				for _, dir := range []string{"bin", "jdk", "lib", "logs", "config", "modules", "plugins"} {
					m.Cmd(cli.SYSTEM, "cp", "-r", "usr/install/"+name+"/"+dir, p)
				}

				m.Option(cli.CMD_DIR, p)
				m.Cmdy(cli.DAEMON, "bin/elasticsearch")
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				m.Cmdy(cli.DAEMON)
				return
			}

			if len(arg) == 1 {
				m.Richs(cli.DAEMON, "", arg[0], func(key string, value map[string]interface{}) {
					m.Cmdy("web.spide", "dev", "raw", "GET", "http://localhost:9200")
				})
			}

			if len(arg) == 2 {
				m.Richs(cli.DAEMON, "", arg[0], func(key string, value map[string]interface{}) {
					m.Cmdy("web.spide", "dev", "raw", "GET", "http://localhost:9200")
				})
			}
		}},

		"index": {Name: "table index 创建:button", Help: "索引", Action: map[string]*ice.Action{}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				m.Cmdy(cli.DAEMON)
				return
			}

			m.Option("header", "Content-Type", "application/json")
			m.Echo(kit.Formats(kit.UnMarshal(m.Cmdx("web.spide", "dev", "raw", "PUT", "http://localhost:9200/"+arg[0]))))
		}},

		"mapping": {Name: "mapping index mapping 创建:button text:textarea", Help: "映射", Action: map[string]*ice.Action{}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				m.Cmdy(cli.DAEMON)
				return
			}

			m.Option("header", "Content-Type", "application/json")
			m.Echo(kit.Formats(kit.UnMarshal(m.Cmdx("web.spide", "dev", "raw", "PUT", "http://localhost:9200/"+arg[0]+"/_mapping/"+arg[1], "data", arg[2]))))
		}},

		"document": {Name: "table index=index_test mapping=mapping_test id=1 查看:button 添加:button data:textarea", Help: "文档", Action: map[string]*ice.Action{
			mdb.INSERT: {Name: "insert", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				if len(arg) > 3 {
					m.Option("header", "Content-Type", "application/json")
					m.Echo(kit.Formats(kit.UnMarshal(m.Cmdx("web.spide", "dev", "raw", "PUT", "http://localhost:9200/"+arg[0]+"/"+arg[1]+"/"+arg[2], "data", arg[3]))))
				}
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Option("header", "Content-Type", "application/json")
			m.Echo(kit.Formats(kit.UnMarshal(m.Cmdx("web.spide", "dev", "raw", "GET", "http://localhost:9200/"+arg[0]+"/"+arg[1]+"/"+arg[2]))))
		}},
	},
}

func init() { code.Index.Register(Index, nil) }
