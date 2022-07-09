package code

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const (
	SRC_WEBSITE  = "src/website/"
	CHAT_WEBSITE = "/chat/website/"
)

func _website_url(m *ice.Message, file string) string {
	const WEBSITE = "website"
	p := path.Join(WEBSITE, file)
	if m.Option(ice.MSG_USERPOD) != "" {
		p = path.Join(ice.POD, m.Option(ice.MSG_USERPOD), WEBSITE, file)
	}
	return strings.Split(kit.MergeURL2(m.Option(ice.MSG_USERWEB), path.Join("/chat", p)), "?")[0]
}

const ZML = nfs.ZML

func init() {
	Index.Register(&ice.Context{Name: ZML, Help: "网页", Commands: ice.Commands{
		ZML: {Name: "zml", Help: "网页", Actions: ice.MergeAction(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(mdb.PLUGIN, mdb.CREATE, nfs.ZML, m.PrefixKey())
				m.Cmd(mdb.RENDER, mdb.CREATE, nfs.ZML, m.PrefixKey())
				LoadPlug(m, ZML)
			}},
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
				m.EchoIFrame(_website_url(m, strings.TrimPrefix(path.Join(arg[2], arg[1]), SRC_WEBSITE)))
			}},
		}, PlugAction())},
	}, Configs: ice.Configs{
		ZML: {Name: ZML, Help: "网页", Value: kit.Data(PLUG, kit.Dict(PREFIX, kit.Dict("# ", COMMENT), PREPARE, kit.Dict(
			KEYWORD, kit.Simple(
				"head", "left", "main", "foot",
				"tabs",
			),
			CONSTANT, kit.Simple(
				"auto", "username",
			),
			FUNCTION, kit.Simple(
				"index", "action", "args", "type",
				"style", "width",
			),
		), KEYWORD, kit.Dict(),
		))},
	}}, nil)
}
