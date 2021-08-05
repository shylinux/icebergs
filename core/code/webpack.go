package code

import (
	"fmt"
	"path"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"
)

func _volcanos(file ...string) string {
	return path.Join(ice.USR_VOLCANOS, path.Join(file...))
}
func _publish(file ...string) string {
	return path.Join(ice.USR_PUBLISH, path.Join(file...))
}

const WEBPACK = "webpack"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		WEBPACK: {Name: "webpack path auto create prunes", Help: "打包", Action: map[string]*ice.Action{
			mdb.CREATE: {Name: "create name=demo", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				dir := m.Conf(web.SERVE, kit.Keym(ice.VOLCANOS, kit.SSH_PATH))
				css, _, e := kit.Create(path.Join(dir, "page/cache.css"))
				m.Assert(e)
				defer css.Close()

				js, _, e := kit.Create(path.Join(dir, "page/cache.js"))
				m.Assert(e)
				defer js.Close()

				m.Option(nfs.DIR_ROOT, dir)
				m.Option(nfs.DIR_DEEP, true)
				m.Option(nfs.DIR_TYPE, nfs.CAT)

				for _, k := range []string{"lib", "panel", "plugin"} {
					m.Cmd(nfs.DIR, k).Table(func(index int, value map[string]string, head []string) {
						if kit.Ext(value[kit.MDB_PATH]) == CSS {
							js.WriteString(`Volcanos.meta.cache["` + path.Join("/", value[kit.MDB_PATH]) + "\"] = []\n")
							css.WriteString(m.Cmdx(nfs.CAT, value[kit.MDB_PATH]))
						}
					})
				}
				for _, k := range []string{"lib", "panel", "plugin"} {
					m.Cmd(nfs.DIR, k).Table(func(index int, value map[string]string, head []string) {
						if kit.Ext(value[kit.MDB_PATH]) == JS {
							js.WriteString(`_can_name = "` + path.Join("/", value[kit.MDB_PATH]) + "\";\n")
							js.WriteString(m.Cmdx(nfs.CAT, value[kit.MDB_PATH]))
						}
					})
				}

				for _, k := range []string{"publish/order.js", "frame.js"} {
					js.WriteString(`_can_name = "` + path.Join("/", k) + "\"\n")
					js.WriteString(m.Cmdx(nfs.CAT, k))
				}

				if f, _, e := kit.Create(path.Join(ice.USR_PUBLISH, WEBPACK, kit.Keys(m.Option(kit.MDB_NAME), JS))); m.Assert(e) {
					defer f.Close()

					f.WriteString("\n")
					f.WriteString(kit.Format(`Volcanos.meta.args = {river: "%s", storm: "%s"}`, m.Option("river"), m.Option("storm")))
					f.WriteString("\n")
					f.WriteString(`Volcanos.meta.pack = ` + kit.Formats(kit.UnMarshal(kit.Select("{}", m.Option("content")))))
				}

				m.Option(nfs.DIR_ROOT, "")
				if f, p, e := kit.Create(_publish(WEBPACK, kit.Keys(m.Option(kit.MDB_NAME), HTML))); m.Assert(e) {
					f.WriteString(fmt.Sprintf(_pack,
						m.Cmdx(nfs.CAT, _volcanos("page/cache.css")),
						m.Cmdx(nfs.CAT, _volcanos("page/index.css")),

						m.Cmdx(nfs.CAT, _volcanos(ice.PROTO_JS)),
						m.Cmdx(nfs.CAT, _publish(path.Join(WEBPACK, kit.Keys(m.Option(kit.MDB_NAME), JS)))),

						m.Cmdx(nfs.CAT, _volcanos("page/cache.js")),
						m.Cmdx(nfs.CAT, _volcanos("page/index.js")),
					))
					m.Echo(p)
				}

				m.Cmd(nfs.COPY, _volcanos("page/cmd.css"), _volcanos("page/index.css"), _volcanos("page/cache.css"))
				m.Cmd(nfs.COPY, _volcanos("page/cmd.js"), _volcanos("proto.js"), _volcanos("page/cache.js"))
			}},
			mdb.PRUNES: {Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(nfs.SAVE, path.Join(ice.USR_VOLCANOS, "page/cache.css"), "")
				m.Cmd(nfs.SAVE, path.Join(ice.USR_VOLCANOS, "page/cache.js"), "")
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Option(nfs.DIR_DEEP, true)
			m.Option(nfs.DIR_TYPE, nfs.CAT)
			m.Option(nfs.DIR_ROOT, m.Conf(PUBLISH, kit.META_PATH))

			m.Cmdy(nfs.DIR, WEBPACK, "time,size,path,action,link")
		}},
	}})
}

const _pack = `
<!DOCTYPE html>
<head>
    <link rel="apple-touch-icon-precomposed" href="/publish/app.png"/>
    <link rel="apple-touch-startup-image" href="/publish/splash.png"/>
    <meta name="apple-mobile-web-app-capable" content="yes"/>
    <meta name="apple-mobile-web-app-status-bar-style" content="black-translucent"/>

    <meta name="viewport" content="width=device-width,initial-scale=0.8,user-scalable=no">
    <meta charset="utf-8">
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
