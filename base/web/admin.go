package web

import (
	"net/http"
	"os"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
)

const ADMIN = "admin"

func init() {
	Index.MergeCommands(ice.Commands{
		ADMIN: {Name: "admin index list", Help: "后台", Role: aaa.VOID, Actions: ice.MergeActions(ice.Actions{
			DREAM_ACTION: {Hand: func(m *ice.Message, arg ...string) { DreamProcessIframe(m, arg...) }},
		}, DreamTablesAction()), Hand: func(m *ice.Message, arg ...string) {
			if m.Option(ice.MSG_SOURCE) != "" {
				RenderMain(m)
			} else {
				kit.If(len(arg) == 0, func() { arg = append(arg, SPACE, DOMAIN) })
				m.Cmd(SPIDE, mdb.CREATE, HostPort(m, tcp.LOCALHOST, kit.GetValid(
					func() string { return m.Cmdx(nfs.CAT, ice.VAR_LOG_ICE_PORT) },
					func() string { return m.Cmdx(nfs.CAT, kit.Path(os.Args[0], "../", ice.VAR_LOG_ICE_PORT)) },
					func() string { return m.Cmdx(nfs.CAT, kit.Path(os.Args[0], "../../", ice.VAR_LOG_ICE_PORT)) },
					func() string { return tcp.PORT_9020 },
				)), ice.OPS)
				args := []string{}
				for i := range arg {
					if arg[i] == "--" {
						arg, args = arg[:i], arg[i+1:]
						break
					}
				}
				kit.If(os.Getenv(cli.CTX_OPS), func(p string) { m.Optionv(SPIDE_HEADER, html.XHost, p) })
				m.Cmdy(SPIDE, ice.OPS, SPIDE_RAW, http.MethodPost, C(arg...), cli.PWD, kit.Path(""), args)
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
func OpsCmd(m *ice.Message, cmd string, arg ...ice.Any) *ice.Message {
	if ice.Info.NodeType == WORKER {
		return m.Cmd(append([]ice.Any{SPACE, ice.OPS, cmd}, arg...)...)
	} else {
		return m.Cmd(append([]ice.Any{cmd}, arg...)...)
	}
}
func DevCmd(m *ice.Message, cmd string, arg ...ice.Any) *ice.Message {
	if ice.Info.NodeType == WORKER {
		return m.Cmd(append([]ice.Any{SPACE, ice.OPS, SPACE, ice.DEV, cmd}, arg...)...)
	} else {
		return m.Cmd(append([]ice.Any{SPACE, ice.DEV, cmd}, arg...)...)
	}
}
