package es

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/code"
	kit "github.com/shylinux/toolkits"

	"path"
	"runtime"
	"strings"
)

const ES = "es"

var Index = &ice.Context{Name: ES, Help: "搜索",
	Configs: map[string]*ice.Config{
		ES: {Name: ES, Help: "搜索", Value: kit.Data(
			"address", "http://localhost:9200",
			"windows", "https://elasticsearch.thans.cn/downloads/elasticsearch/elasticsearch-7.3.2-windows-x86_64.zip",
			"darwin", "https://elasticsearch.thans.cn/downloads/elasticsearch/elasticsearch-7.3.2-darwin-x86_64.tar.gz",
			"linux", "https://elasticsearch.thans.cn/downloads/elasticsearch/elasticsearch-7.3.2-linux-x86_64.tar.gz",
		)},
	},
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},

		ES: {Name: "es hash=auto auto 启动:button 安装:button", Help: "搜索", Action: map[string]*ice.Action{
			"download": {Name: "download", Help: "安装", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy("web.code.install", "download", m.Conf(ES, kit.Keys(kit.MDB_META, runtime.GOOS)))
			}},

			"start": {Name: "start", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
				name := path.Base(m.Conf(ES, kit.Keys(kit.MDB_META, runtime.GOOS)))
				name = strings.Join(strings.Split(name, "-")[:2], "-")
				m.Cmdy("web.code.install", "start", name, "bin/elasticsearch")
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				m.Cmdy(cli.DAEMON)
				return
			}

			m.Richs(cli.DAEMON, "", arg[0], func(key string, value map[string]interface{}) {
				m.Cmdy(web.SPIDE, web.SPIDE_DEV, web.SPIDE_RAW, web.SPIDE_GET, m.Conf(ES, "meta.address"))
			})
		}},

		"GET": {Name: "GET 查看:button cmd=/", Help: "命令", Action: map[string]*ice.Action{}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if pod := m.Option("_pod"); pod != "" {
				m.Option("_pod", "")
				m.Cmdy(web.SPACE, pod, m.Prefix(cmd), arg)

				if m.Result(0) != ice.ErrWarn || m.Result(1) != ice.ErrNotFound {
					return
				}
				m.Set(ice.MSG_RESULT)
			}

			m.Option(web.SPIDE_HEADER, web.ContentType, web.ContentJSON)
			m.Echo(kit.Formats(kit.UnMarshal(m.Cmdx(web.SPIDE, web.SPIDE_DEV, web.SPIDE_RAW,
				web.SPIDE_GET, kit.MergeURL2(m.Conf(ES, "meta.address"), kit.Select("/", arg, 0))))))
		}},
		"CMD": {Name: "CMD 执行:button method:select=GET|PUT|POST|DELETE cmd=/ data:textarea", Help: "命令", Action: map[string]*ice.Action{}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if pod := m.Option("_pod"); pod != "" {
				m.Option("_pod", "")
				m.Cmdy(web.SPACE, pod, m.Prefix(cmd), arg)

				if m.Result(0) != ice.ErrWarn || m.Result(1) != ice.ErrNotFound {
					return
				}
				m.Set(ice.MSG_RESULT)
			}

			m.Option(web.SPIDE_HEADER, web.ContentType, web.ContentJSON)
			prefix := []string{web.SPIDE, web.SPIDE_DEV, web.SPIDE_RAW, arg[0], kit.MergeURL2(m.Conf(ES, "meta.address"), arg[1])}

			if len(arg) > 2 {
				prefix = append(prefix, web.SPIDE_DATA, arg[2])
			}
			m.Echo(kit.Formats(kit.UnMarshal(m.Cmdx(prefix))))
		}},
	},
}

func init() { code.Index.Register(Index, nil) }
