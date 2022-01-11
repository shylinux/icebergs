package code

import (
	"fmt"
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _volcanos(m *ice.Message, file ...string) string {
	return path.Join(m.Conf(web.SERVE, kit.Keym(ice.VOLCANOS, nfs.PATH)), path.Join(file...))
}
func _publish(m *ice.Message, file ...string) string {
	return path.Join(m.Conf(PUBLISH, kit.Keym(nfs.PATH)), path.Join(file...))
}

const (
	PUBLISH_ORDER_JS = "publish/order.js"
	PAGE_CACHE_CSS   = "page/cache.css"
	PAGE_INDEX_CSS   = "page/index.css"
	PAGE_CACHE_JS    = "page/cache.js"
	PAGE_INDEX_JS    = "page/index.js"
	PAGE_CAN_CSS     = "page/can.css"
	PAGE_CAN_JS      = "page/can.js"
)

const DEVPACK = "devpack"
const WEBPACK = "webpack"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		WEBPACK: {Name: "webpack path auto create prunes", Help: "打包", Action: map[string]*ice.Action{
			mdb.CREATE: {Name: "create name=hi", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				dir := _volcanos(m)
				css, _, e := kit.Create(path.Join(dir, PAGE_CACHE_CSS))
				m.Assert(e)
				defer css.Close()

				js, _, e := kit.Create(path.Join(dir, PAGE_CACHE_JS))
				m.Assert(e)
				defer js.Close()

				m.Option(nfs.DIR_ROOT, dir)
				m.Option(nfs.DIR_DEEP, true)
				m.Option(nfs.DIR_TYPE, nfs.CAT)

				for _, k := range []string{"lib", "panel", "plugin"} {
					m.Cmd(nfs.DIR, k).Table(func(index int, value map[string]string, head []string) {
						if kit.Ext(value[nfs.PATH]) == CSS {
							js.WriteString(`Volcanos.meta.cache["` + path.Join(ice.PS, value[nfs.PATH]) + "\"] = []\n")
							css.WriteString(m.Cmdx(nfs.CAT, value[nfs.PATH]))
						}
					})
				}
				js.WriteString(ice.NL)
				for _, k := range []string{"lib", "panel", "plugin"} {
					m.Cmd(nfs.DIR, k).Table(func(index int, value map[string]string, head []string) {
						if kit.Ext(value[nfs.PATH]) == JS {
							js.WriteString(`_can_name = "` + path.Join(ice.PS, value[nfs.PATH]) + "\";\n")
							js.WriteString(m.Cmdx(nfs.CAT, value[nfs.PATH]))
						}
					})
				}

				for _, k := range []string{PUBLISH_ORDER_JS, ice.FRAME_JS} {
					js.WriteString(`_can_name = "` + path.Join(ice.PS, k) + "\"\n")
					js.WriteString(m.Cmdx(nfs.CAT, k))
				}

				if f, _, e := kit.Create(_publish(m, WEBPACK, kit.Keys(m.Option(mdb.NAME), JS))); m.Assert(e) {
					defer f.Close()

					f.WriteString(ice.NL)
					f.WriteString(kit.Format(`Volcanos.meta.args = {river: "%s", storm: "%s"}`, m.Option(web.RIVER), m.Option(web.STORM)))
					f.WriteString(ice.NL)
					f.WriteString(`Volcanos.meta.pack = ` + kit.Formats(kit.UnMarshal(kit.Select("{}", m.Option("content")))))
				}

				m.Option(nfs.DIR_ROOT, "")
				if f, p, e := kit.Create(_publish(m, WEBPACK, kit.Keys(m.Option(mdb.NAME), HTML))); m.Assert(e) {
					defer f.Close()

					f.WriteString(fmt.Sprintf(_pack,
						m.Cmdx(nfs.CAT, _volcanos(m, PAGE_CACHE_CSS)),
						m.Cmdx(nfs.CAT, _volcanos(m, PAGE_INDEX_CSS)),

						m.Cmdx(nfs.CAT, _volcanos(m, ice.PROTO_JS)),
						m.Cmdx(nfs.CAT, _publish(m, path.Join(WEBPACK, kit.Keys(m.Option(mdb.NAME), JS)))),

						m.Cmdx(nfs.CAT, _volcanos(m, PAGE_CACHE_JS)),
						m.Cmdx(nfs.CAT, _volcanos(m, PAGE_INDEX_JS)),
					))
					m.Echo(p)
				}

				m.Cmd(nfs.COPY, _volcanos(m, PAGE_CAN_CSS), _volcanos(m, PAGE_INDEX_CSS), _volcanos(m, PAGE_CACHE_CSS))
				m.Cmd(nfs.COPY, _volcanos(m, PAGE_CAN_JS), _volcanos(m, ice.PROTO_JS), _volcanos(m, PAGE_CACHE_JS))
			}},
			mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(nfs.SAVE, _volcanos(m, PAGE_CACHE_JS))
				m.Cmd(nfs.SAVE, _volcanos(m, PAGE_CACHE_CSS))

				m.Cmd(nfs.COPY, _volcanos(m, PAGE_CAN_CSS), _volcanos(m, PAGE_INDEX_CSS), _volcanos(m, PAGE_CACHE_CSS))
				m.Cmd(nfs.COPY, _volcanos(m, PAGE_CAN_JS), _volcanos(m, ice.PROTO_JS), _volcanos(m, PAGE_CACHE_JS))
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Option(nfs.DIR_DEEP, true)
			m.Option(nfs.DIR_TYPE, nfs.CAT)
			m.Option(nfs.DIR_ROOT, m.Conf(PUBLISH, kit.Keym(nfs.PATH)))

			m.Cmdy(nfs.DIR, WEBPACK, "time,size,path,action,link")
		}},
	}})
}

const _pack = `
<!DOCTYPE html>
<head>
    <meta charset="utf-8">
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
