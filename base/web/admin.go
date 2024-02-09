package web

import (
	"net/http"
	"os"
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
)

const ADMIN = "admin"

func init() {
	Index.MergeCommands(ice.Commands{
		ADMIN: {Name: "admin index list", Help: "管理", Role: aaa.VOID, Actions: ice.Actions{
			ice.CTX_INIT: {Hand: DreamWhiteHandle},
			DREAM_TABLES: {Hand: func(m *ice.Message, arg ...string) { m.PushButton(kit.Dict(ADMIN, "后台")) }},
			DREAM_ACTION: {Hand: func(m *ice.Message, arg ...string) {
				if kit.HasPrefixList(arg, ctx.ACTION, ADMIN) && len(arg) == 2 {
					link := m.MergePodCmd(m.Option(mdb.NAME), "")
					if m.Option(mdb.TYPE) == MASTER {
						link = SpideOrigin(m, m.Option(mdb.NAME)) + C(m.PrefixKey())
					}
					if m.IsMetaKey() {
						m.ProcessOpen(link)
					} else {
						ctx.ProcessFloat(m, CHAT_IFRAME, link, arg...)
						m.ProcessField(ctx.ACTION, ctx.RUN, CHAT_IFRAME)
					}
				}
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			if m.Option(ice.MSG_SOURCE) != "" {
				RenderMain(m)
			} else {
				kit.If(len(arg) == 0, func() { arg = append(arg, SPACE, DOMAIN) })
				m.Cmd(SPIDE, mdb.CREATE, ice.OPS, HostPort(m, tcp.LOCALHOST, kit.GetValid(
					func() string { return m.Cmdx(nfs.CAT, ice.VAR_LOG_ICE_PORT) },
					func() string { return m.Cmdx(nfs.CAT, kit.Path(os.Args[0], "../", ice.VAR_LOG_ICE_PORT)) },
					func() string { return m.Cmdx(nfs.CAT, kit.Path(os.Args[0], "../../", ice.VAR_LOG_ICE_PORT)) },
					func() string { return tcp.PORT_9020 },
				)))
				args := []string{}
				for i := range arg {
					if arg[i] == "--" {
						arg, args = arg[:i], arg[i+1:]
						break
					}
				}
				m.Cmdy(SPIDE, ice.OPS, SPIDE_RAW, http.MethodPost, C(path.Join(arg...)), cli.PWD, kit.Path(""), args)
			}
		}},
	})
}
func AdminCmd(m *ice.Message, cmd string, arg ...ice.Any) *ice.Message {
	if ice.Info.NodeType == WORKER {
		return m.Cmd(append([]ice.Any{SPACE, ice.OPS, cmd}, arg...)...)
	} else {
		return m.Cmd(append([]ice.Any{cmd}, arg...)...)
	}
}
