package es

import (
	"net/http"
	"path"
	"runtime"

	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"

	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/code"
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

		ES: {Name: "es 安装:button", Help: "搜索", Action: map[string]*ice.Action{
			"install": {Name: "install", Help: "安装", Hand: func(m *ice.Message, arg ...string) {
				name := path.Base(m.Conf(ES, kit.Keys("meta", runtime.GOOS)))
				msg := m.Cmd(web.SPIDE, "dev", "cache", http.MethodGet, m.Conf(ES, kit.Keys("meta", runtime.GOOS)))
				m.Cmdy(nfs.LINK, path.Join("usr/install/", name), msg.Append("file"))
				m.Option(cli.CMD_DIR, "usr/install")
				m.Cmd(cli.SYSTEM, "tar", "xvf", name)
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		}},
	},
}

func init() { code.Index.Register(Index, nil) }
