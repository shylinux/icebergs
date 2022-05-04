package chat

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const CMD = "cmd"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		CMD: {Name: CMD, Help: "命令", Value: kit.Data(mdb.SHORT, "type", nfs.PATH, nfs.PWD)},
	}, Commands: map[string]*ice.Command{
		"/cmd/": {Name: "/cmd/", Help: "命令", Action: ice.MergeAction(map[string]*ice.Action{
			ice.CTX_INIT: {Name: "_init", Help: "初始化", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(CMD, mdb.CREATE, mdb.TYPE, "shy", mdb.NAME, "web.wiki.word")
				m.Cmdy(CMD, mdb.CREATE, mdb.TYPE, "svg", mdb.NAME, "web.wiki.draw")
				m.Cmdy(CMD, mdb.CREATE, mdb.TYPE, "csv", mdb.NAME, "web.wiki.data")
				m.Cmdy(CMD, mdb.CREATE, mdb.TYPE, "json", mdb.NAME, "web.wiki.json")

				for _, k := range []string{"mod", "sum"} {
					m.Cmdy(CMD, mdb.CREATE, mdb.TYPE, k, mdb.NAME, "web.code.inner")
				}
			}},
		}, ctx.CmdAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if strings.HasSuffix(m.R.URL.Path, ice.PS) {
				m.RenderCmd(CMD)
				return // 目录
			}

			if mdb.HashSelect(m.Spawn(), path.Join(arg...)).Table(func(index int, value map[string]string, head []string) {
				m.RenderCmd(value[mdb.NAME])
			}).Length() > 0 {
				return // 命令
			}

			p := path.Join(m.Config(nfs.PATH), path.Join(arg...))
			if mdb.HashSelect(m.Spawn(), kit.Ext(m.R.URL.Path)).Table(func(index int, value map[string]string, head []string) {
				m.RenderCmd(value[mdb.NAME], p)
			}).Length() > 0 {
				return // 插件
			}

			switch p := path.Join(arg...); kit.Ext(p) {
			case nfs.JS:
				if cmd := ice.GetFileCmd(p); cmd != "" {
					m.Display(ice.FileURI(p))
					m.RenderCmd(cmd)
				}
				return
			case nfs.GO:
				if cmd := ice.GetFileCmd(p); cmd != "" {
					m.RenderCmd(cmd)
				}
				return
			case nfs.SH:
				if cmd := ice.GetFileCmd(p); cmd != "" {
					msg := m.Cmd(cmd, ice.OptionFields(""))
					if msg.Length() > 0 {
						msg.Table()
					}
					m.Cmdy(cli.SYSTEM, "sh", p, msg.Result())
					m.RenderResult()
				}
				return
			}

			if m.PodCmd(ctx.COMMAND, arg[0]) {
				if !m.IsErr() {
					m.RenderCmd(arg[0], arg[1:]) // 远程命令
				}
			} else if m.Cmdy(ctx.COMMAND, arg[0]); m.Length() > 0 {
				m.RenderCmd(arg[0], arg[1:]) // 本地命令
			} else {
				m.RenderDownload(p) // 文件
			}
		}},
		CMD: {Name: "cmd path auto upload up home", Help: "命令", Action: ice.MergeAction(map[string]*ice.Action{
			web.UPLOAD: {Name: "upload", Help: "上传", Hand: func(m *ice.Message, arg ...string) {
				m.Upload(path.Join(m.Config(nfs.PATH), strings.TrimPrefix(path.Dir(m.R.URL.Path), "/cmd")))
			}},

			"home": {Name: "home", Help: "根目录", Hand: func(m *ice.Message, arg ...string) {
				m.ProcessLocation("/chat/cmd/")
			}},
			"up": {Name: "up", Help: "上一级", Hand: func(m *ice.Message, arg ...string) {
				if strings.TrimPrefix(m.R.URL.Path, "/cmd") == ice.PS {
					m.Cmdy(CMD)
				} else if strings.HasSuffix(m.R.URL.Path, ice.PS) {
					m.ProcessLocation("../")
				} else {
					m.ProcessLocation(nfs.PWD)
				}
			}},
		}, mdb.HashAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 0 {
				m.ProcessLocation(arg[0])
				return
			}
			switch p := path.Join(arg...); kit.Ext(p) {
			case "html":
				m.RenderResult(m.Cmdx(nfs.CAT, path.Join(ice.SRC, path.Join(arg...))))
				return
			}
			m.Option(nfs.DIR_ROOT, path.Join(m.Config(nfs.PATH), strings.TrimPrefix(path.Dir(m.R.URL.Path), "/cmd")))
			m.Cmdy(nfs.DIR, arg)
		}},
	}})
}
