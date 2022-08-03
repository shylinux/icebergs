package code

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
)

func _website_url(m *ice.Message, file string) string {
	return strings.Split(web.MergeWebsite(m, file), "?")[0]
}

const ZML = nfs.ZML

func init() {
	const (
		SRC_WEBSITE = "src/website/"
	)
	Index.Register(&ice.Context{Name: ZML, Help: "网页", Commands: ice.Commands{
		ZML: {Name: "zml", Help: "网页", Actions: ice.MergeAction(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(mdb.RENDER, mdb.CREATE, nfs.ZML, m.PrefixKey())
				m.Cmd(mdb.ENGINE, mdb.CREATE, nfs.ZML, m.PrefixKey())
			}},
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
				m.EchoIFrame(_website_url(m, strings.TrimPrefix(path.Join(arg[2], arg[1]), SRC_WEBSITE)))
			}},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) {
				m.Echo(_website_url(m, strings.TrimPrefix(path.Join(arg[2], arg[1]), SRC_WEBSITE)))
			}},
		})},
	}}, nil)
}
