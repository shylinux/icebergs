package code

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	kit "github.com/shylinux/toolkits"

	"fmt"
	"path"
	"strings"
)

const WEBPACK = "webpack"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			WEBPACK: {Name: WEBPACK, Help: "webpack", Value: kit.Data(kit.MDB_PATH, "usr/volcanos")},
		},
		Commands: map[string]*ice.Command{
			WEBPACK: {Name: "webpack path auto create", Help: "打包", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create name=demo", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					css, _, e := kit.Create(path.Join(m.Conf(WEBPACK, kit.META_PATH), "page/cache.css"))
					m.Assert(e)
					defer css.Close()

					js, _, e := kit.Create(path.Join(m.Conf(WEBPACK, kit.META_PATH), "page/cache.js"))
					m.Assert(e)
					defer js.Close()

					m.Option(nfs.DIR_ROOT, m.Conf(WEBPACK, kit.META_PATH))
					m.Option(nfs.DIR_TYPE, nfs.CAT)
					m.Option(nfs.DIR_DEEP, true)

					for _, k := range []string{"lib", "panel", "plugin"} {
						m.Cmd(nfs.DIR, k).Table(func(index int, value map[string]string, head []string) {
							if strings.HasSuffix(value[kit.MDB_PATH], ".css") {
								js.WriteString(`Volcanos.meta.cache["` + path.Join("/", value[kit.MDB_PATH]) + "\"] = []\n")
								css.WriteString(m.Cmdx(nfs.CAT, value[kit.MDB_PATH]))
							}
						})

						m.Cmd(nfs.DIR, k).Table(func(index int, value map[string]string, head []string) {
							if strings.HasSuffix(value[kit.MDB_PATH], ".js") {
								js.WriteString(`_can_name = "` + path.Join("/", value[kit.MDB_PATH]) + "\";\n")
								js.WriteString(m.Cmdx(nfs.CAT, value[kit.MDB_PATH]))
							}
						})
					}

					for _, k := range []string{"frame.js", "publish/order.js"} {
						js.WriteString(`_can_name = "` + path.Join("/", k) + "\"\n")
						js.WriteString(m.Cmdx(nfs.CAT, k))
					}

					if f, _, e := kit.Create("usr/publish/webpack/" + m.Option("name") + ".js"); m.Assert(e) {
						defer f.Close()

						f.WriteString("\n")
						f.WriteString(kit.Format(`Volcanos.meta.args = {river: "%s", storm: "%s"}`, m.Option("river"), m.Option("storm")))
						f.WriteString("\n")
						f.WriteString(`Volcanos.meta.pack = ` + kit.Formats(kit.UnMarshal(kit.Select("{}", m.Option("content")))))
					}

					m.Option(nfs.DIR_ROOT, "")
					if f, p, e := kit.Create("usr/publish/webpack/" + m.Option("name") + ".html"); m.Assert(e) {
						f.WriteString(fmt.Sprintf(_pack,
							m.Cmdx(nfs.CAT, "usr/volcanos/page/cache.css"),
							m.Cmdx(nfs.CAT, "usr/volcanos/page/index.css"),

							m.Cmdx(nfs.CAT, "usr/volcanos/proto.js"),
							m.Cmdx(nfs.CAT, "usr/publish/webpack/"+m.Option("name")+".js"),
							m.Cmdx(nfs.CAT, "usr/volcanos/page/cache.js"),
							m.Cmdx(nfs.CAT, "usr/volcanos/page/index.js"),
						))
						m.Echo(p)
					}
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(nfs.DIR_ROOT, m.Conf(PUBLISH, kit.META_PATH))
				m.Option(nfs.DIR_TYPE, nfs.CAT)
				m.Option(nfs.DIR_DEEP, true)

				m.Cmdy(nfs.DIR, WEBPACK).Table(func(index int, value map[string]string, head []string) {
					m.PushDownload(kit.MDB_LINK, path.Join(m.Option(nfs.DIR_ROOT), value[kit.MDB_PATH]))
				})
			}},
		},
	})
}

const _pack = `
<!DOCTYPE html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width,initial-scale=0.7,user-scalable=no">
    <title>volcanos</title>
    <style type="text/css">%s</style>
    <style type="text/css">%s</style>
</head>
<body>
<script>%s</script>
<script>%s</script>
<script>%s</script>
<script>%s</script>
<script>Volcanos.meta.webpack = true</script>
</body>
`
