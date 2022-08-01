package chat

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _cmd_file(m *ice.Message, arg ...string) bool {
	switch p := path.Join(arg...); kit.Ext(p) {
	case nfs.JS:
		m.Display(ice.FileURI(p))
		m.RenderCmd(kit.Select(ctx.CAN_PLUGIN, ice.GetFileCmd(p)))

	case nfs.GO:
		m.RenderCmd(ice.GetFileCmd(p))

	case nfs.SHY:
		m.RenderCmd("web.wiki.word", p)

	case nfs.IML:
		if m.Option(ice.MSG_USERPOD) == "" {
			m.RenderRedirect(path.Join(CHAT_WEBSITE, strings.TrimPrefix(p, SRC_WEBSITE)))
			m.Option(ice.MSG_ARGS, m.Option(ice.MSG_ARGS))
		} else {
			m.RenderRedirect(path.Join("/chat/pod", m.Option(ice.MSG_USERPOD), "website", strings.TrimPrefix(p, SRC_WEBSITE)))
			m.Option(ice.MSG_ARGS, m.Option(ice.MSG_ARGS))
		}

	case nfs.ZML:
		m.RenderCmd("can.parse", m.Cmdx(nfs.CAT, p))

	default:
		if p = strings.TrimPrefix(p, ice.SRC+ice.PS); kit.FileExists(path.Join(ice.SRC, p)) {
			if msg := m.Cmd(mdb.RENDER, kit.Ext(p)); msg.Length() > 0 {
				m.Cmdy(mdb.RENDER, kit.Ext(p), p, ice.SRC+ice.PS).RenderResult()
				break
			}
		}
		return false
	}
	return true
}

const CMD = "cmd"

func init() {
	Index.MergeCommands(ice.Commands{
		CMD: {Name: "cmd path auto upload up home", Help: "命令", Actions: ice.MergeAction(ice.Actions{
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
		}, mdb.HashAction(mdb.SHORT, "type", nfs.PATH, nfs.PWD)), Hand: func(m *ice.Message, arg ...string) {
			if _cmd_file(m, arg...) {
				return
			}
			if msg := m.Cmd(ctx.COMMAND, arg[0]); msg.Length() > 0 {
				m.RenderCmd(arg[0])
				return
			}

			if len(arg) > 0 {
				m.ProcessLocation(arg[0])
				return
			}
			m.Option(nfs.DIR_ROOT, path.Join(m.Config(nfs.PATH), strings.TrimPrefix(path.Dir(m.R.URL.Path), "/cmd")))
			m.Cmdy(nfs.DIR, arg)
		}},
		"/cmd/": {Name: "/cmd/", Help: "命令", Actions: ice.MergeAction(ice.Actions{
			ice.CTX_INIT: {Name: "_init", Help: "初始化", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(CMD, mdb.CREATE, mdb.TYPE, "shy", mdb.NAME, "web.wiki.word")
				m.Cmdy(CMD, mdb.CREATE, mdb.TYPE, "svg", mdb.NAME, "web.wiki.draw")
				m.Cmdy(CMD, mdb.CREATE, mdb.TYPE, "csv", mdb.NAME, "web.wiki.data")
				m.Cmdy(CMD, mdb.CREATE, mdb.TYPE, "json", mdb.NAME, "web.wiki.json")

				for _, k := range []string{"mod", "sum"} {
					m.Cmdy(CMD, mdb.CREATE, mdb.TYPE, k, mdb.NAME, "web.code.inner")
				}
			}},
		}, ctx.CmdAction()), Hand: func(m *ice.Message, arg ...string) {
			if strings.HasSuffix(m.R.URL.Path, ice.PS) {
				m.RenderCmd(CMD)
				return // 目录
			}
			if _cmd_file(m, arg...) {
				return
			}

			if m.PodCmd(ctx.COMMAND, arg[0]) {
				if !m.IsErr() {
					m.RenderCmd(arg[0], arg[1:]) // 远程命令
				}
			} else if m.Cmdy(ctx.COMMAND, arg[0]); m.Length() > 0 {
				m.RenderCmd(arg[0], arg[1:]) // 本地命令
			} else {
				m.RenderDownload(path.Join(m.Config(nfs.PATH), path.Join(arg...))) // 文件
			}
		}},
	})
}
