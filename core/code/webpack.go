package code

import (
	"fmt"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _volcanos(m *ice.Message, file ...string) string {
	return path.Join(ice.USR_VOLCANOS, path.Join(file...))
}
func _publish(m *ice.Message, file ...string) string {
	return path.Join(ice.USR_PUBLISH, path.Join(file...))
}
func _webpack_can(m *ice.Message) {
	m.Option(nfs.DIR_ROOT, "")
	m.Cmd(nfs.COPY, _volcanos(m, PAGE_CAN_CSS), _volcanos(m, PAGE_INDEX_CSS), _volcanos(m, PAGE_CACHE_CSS))
	m.Cmd(nfs.COPY, _volcanos(m, PAGE_CAN_JS), _volcanos(m, ice.PROTO_JS), _volcanos(m, PAGE_CACHE_JS))
	m.Cmdy(nfs.DIR, _volcanos(m, PAGE))
}
func _webpack_cache(m *ice.Message, dir string, write bool) {
	if len(ice.Info.Pack) > 0 {
		return
	}

	css, _, e := kit.Create(path.Join(dir, PAGE_CACHE_CSS))
	m.Assert(e)
	defer css.Close()

	js, _, e := kit.Create(path.Join(dir, PAGE_CACHE_JS))
	m.Assert(e)
	defer js.Close()
	defer fmt.Fprintln(js, `_can_name = ""`)

	defer _webpack_can(m)
	if !write {
		return
	}

	m.Option(nfs.DIR_ROOT, "")
	m.Option(nfs.DIR_DEEP, true)
	m.Option(nfs.DIR_TYPE, nfs.CAT)

	// m.Cmd(nfs.DIR, ice.SRC).Tables(func(value ice.Maps) {
	// 	if kit.Ext(value[nfs.PATH]) == JS {
	// 		fmt.Fprintln(js, `_can_name = "`+path.Join("/require", ice.Info.Make.Module, value[nfs.PATH])+`"`)
	// 		fmt.Fprintln(js, m.Cmdx(nfs.CAT, value[nfs.PATH]))
	// 	}
	// })

	m.Option(nfs.DIR_ROOT, dir)
	for _, k := range []string{LIB, PANEL, PLUGIN} {
		m.Cmd(nfs.DIR, k).Sort(nfs.PATH).Tables(func(value ice.Maps) {
			if kit.Ext(value[nfs.PATH]) == CSS {
				fmt.Fprintln(css, kit.Format("/* %s */", path.Join(ice.PS, value[nfs.PATH])))
				fmt.Fprintln(css, m.Cmdx(nfs.CAT, value[nfs.PATH]))
				fmt.Fprintln(js, `Volcanos.meta.cache["`+path.Join(ice.PS, value[nfs.PATH])+`"] = []`)
			}
		})
	}
	fmt.Fprintln(js)
	for _, k := range []string{LIB, PANEL, PLUGIN} {
		m.Cmd(nfs.DIR, k).Sort(nfs.PATH).Tables(func(value ice.Maps) {
			if kit.Ext(value[nfs.PATH]) == JS {
				fmt.Fprintln(js, `_can_name = "`+path.Join(ice.PS, value[nfs.PATH])+`"`)
				fmt.Fprintln(js, m.Cmdx(nfs.CAT, value[nfs.PATH]))
			}
		})
	}
	for _, k := range []string{ice.FRAME_JS} {
		fmt.Fprintln(js, `_can_name = "`+path.Join(ice.PS, k)+`"`)
		fmt.Fprintln(js, m.Cmdx(nfs.CAT, k))
	}

	m.Cmd(mdb.SELECT, m.PrefixKey(), "", mdb.HASH, ice.OptionFields(nfs.PATH)).Sort(nfs.PATH).Tables(func(value ice.Maps) {
		defer fmt.Fprintln(js)

		p := value[nfs.PATH]
		switch kit.Ext(p) {
		case nfs.CSS:
			fmt.Fprintln(css, kit.Format("/* %s */", path.Join("/require/node_modules/", p)))
			fmt.Fprintln(css, m.Cmdx(nfs.CAT, path.Join("node_modules", p)))
			fmt.Fprintln(js, `Volcanos.meta.cache["`+path.Join("/require/node_modules/", p)+`"] = []`)
			return
		case nfs.JS:
		default:
			p = p + "/lib/" + p + ".js"
		}

		fmt.Fprintln(js, `_can_name = "`+path.Join("/require/node_modules/", p)+`"`)
		fmt.Fprintln(js, m.Cmdx(nfs.CAT, path.Join("node_modules", p)))
		fmt.Fprintln(js, `Volcanos.meta.cache["`+path.Join("/require/node_modules/", p)+`"] = []`)
	})
}
func _webpack_build(m *ice.Message, file string) {
	if f, _, e := kit.Create(kit.Keys(file, JS)); m.Assert(e) {
		defer f.Close()
		fmt.Fprintln(f, `Volcanos.meta.webpack = true`)
		fmt.Fprintln(f, `Volcanos.meta.pack = `+kit.Formats(kit.UnMarshal(kit.Select("{}", m.Option(nfs.CONTENT)))))
		fmt.Fprintln(f, `Volcanos.meta.args = `+kit.Formats(kit.Dict(m.OptionSimple(kit.Split(m.Option(ctx.ARGS))...))))
	}

	if f, p, e := kit.Create(kit.Keys(file, HTML)); m.Assert(e) {
		defer f.Close()
		defer m.Echo(p)

		fmt.Fprintf(f, `
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
</body>
`,
			m.Cmdx(nfs.CAT, _volcanos(m, PAGE_INDEX_CSS)),
			m.Cmdx(nfs.CAT, _volcanos(m, PAGE_CACHE_CSS)),
			m.Cmdx(nfs.CAT, _volcanos(m, ice.PROTO_JS)), m.Cmdx(nfs.CAT, kit.Keys(file, JS)),
			m.Cmdx(nfs.CAT, _volcanos(m, PAGE_CACHE_JS)),
			m.Cmdx(nfs.CAT, _volcanos(m, PAGE_INDEX_JS)),
		)
	}
}

const (
	LIB    = "lib"
	PAGE   = "page"
	PANEL  = "panel"
	PLUGIN = "plugin"
)
const (
	PUBLISH_ORDER_JS = "publish/order.js"
	PAGE_INDEX_CSS   = "page/index.css"
	PAGE_CACHE_CSS   = "page/cache.css"
	PAGE_CACHE_JS    = "page/cache.js"
	PAGE_INDEX_JS    = "page/index.js"
	PAGE_CAN_CSS     = "page/can.css"
	PAGE_CAN_JS      = "page/can.js"
)

const DEVPACK = "devpack"
const WEBPACK = "webpack"

func init() {
	Index.MergeCommands(ice.Commands{
		WEBPACK: {Name: "webpack path auto create remove", Help: "打包", Actions: ice.MergeAction(ice.Actions{
			mdb.CREATE: {Name: "create", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				_webpack_cache(m.Spawn(), _volcanos(m), true)
			}},
			mdb.INSERT: {Name: "insert", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(mdb.INSERT, m.PrefixKey(), "", mdb.HASH, nfs.PATH, arg[0])
			}},
			mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				_webpack_cache(m.Spawn(), _volcanos(m), false)
			}},
			nfs.TRASH: {Name: "trash", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				if !strings.Contains(m.Option(nfs.PATH), "page/index") {
					m.Cmd(nfs.TRASH, m.Option(nfs.PATH))
				}
			}},
			cli.BUILD: {Name: "build name=hi", Help: "构建", Hand: func(m *ice.Message, arg ...string) {
				_webpack_cache(m.Spawn(), _volcanos(m), true)
				_webpack_build(m, _publish(m, WEBPACK, m.Option(mdb.NAME)))
			}},
		}, mdb.HashAction(mdb.SHORT, nfs.PATH)), Hand: func(m *ice.Message, arg ...string) {
			m.Option(nfs.DIR_DEEP, true)
			m.Option(nfs.DIR_TYPE, nfs.CAT)
			m.OptionFields(nfs.DIR_WEB_FIELDS)
			m.Cmdy(nfs.DIR, _volcanos(m, PAGE))
			m.Cmdy(nfs.DIR, _publish(m, WEBPACK))
		}},
	})
}
