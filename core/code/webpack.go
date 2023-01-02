package code

import (
	"fmt"
	"io"
	"os"
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
func _webpack_css(m *ice.Message, css, js io.Writer, p string) {
	fmt.Fprintln(css, kit.Format("/* %s */", path.Join(ice.PS, p)))
	fmt.Fprintln(css, m.Cmdx(nfs.CAT, strings.Replace(p, "require/node_modules/", "src/node_modules/", 1)))
	fmt.Fprintln(js, `Volcanos.meta.cache["`+path.Join(ice.PS, p)+`"] = []`)
}
func _webpack_js(m *ice.Message, js io.Writer, p string) {
	fmt.Fprintln(js, `_can_name = "`+path.Join(ice.PS, p)+`";`)
	fmt.Fprintln(js, m.Cmdx(nfs.CAT, strings.TrimPrefix(p, ice.REQUIRE+ice.PS)))
}
func _webpack_node(m *ice.Message, js io.Writer, p string) {
	fmt.Fprintln(js, `_can_name = "`+path.Join(ice.PS, p)+`";`)
	fmt.Fprintln(js, m.Cmdx(nfs.CAT, strings.Replace(p, "require/node_modules/", "src/node_modules/", 1)))
	fmt.Fprintln(js, `Volcanos.meta.cache["`+path.Join(ice.PS, p)+`"] = []`)
}
func _webpack_cache(m *ice.Message, dir string, write bool) {
	if _, e := nfs.DiskFile.StatFile(ice.USR_VOLCANOS); os.IsNotExist(e) {
		return
	}
	css, _, e := nfs.CreateFile(m, path.Join(dir, PAGE_CACHE_CSS))
	m.Assert(e)
	defer css.Close()
	js, _, e := nfs.CreateFile(m, path.Join(dir, PAGE_CACHE_JS))
	m.Assert(e)
	defer js.Close()
	defer fmt.Fprintln(js, `_can_name = ""`)
	defer _webpack_can(m)
	if !write {
		return
	}
	for _, k := range []string{LIB, PANEL, PLUGIN} {
		nfs.DirDeepAll(m, dir, k, func(value ice.Maps) {
			if kit.Ext(value[nfs.PATH]) == CSS {
				_webpack_css(m, css, js, value[nfs.PATH])
			}
		})
	}
	fmt.Fprintln(js)
	for _, k := range []string{LIB, PANEL, PLUGIN} {
		nfs.DirDeepAll(m, dir, k, func(value ice.Maps) {
			if kit.Ext(value[nfs.PATH]) == JS {
				_webpack_js(m, js, value[nfs.PATH])
			}
		})
	}
	for _, k := range []string{ice.FRAME_JS} {
		_webpack_js(m, js, k)
	}
	m.Option(nfs.DIR_ROOT, "")
	mdb.HashSelects(m).Sort(nfs.PATH).Tables(func(value ice.Maps) {
		defer fmt.Fprintln(js)
		p := value[nfs.PATH]
		switch kit.Ext(p) {
		case nfs.CSS:
			_webpack_css(m, css, js, path.Join(ice.REQUIRE, ice.NODE_MODULES, p))
			return
		case nfs.JS:
		default:
			p = path.Join(p, LIB, p+".js")
		}
		_webpack_node(m, js, path.Join(ice.REQUIRE, ice.NODE_MODULES, p))
	})
}
func _webpack_build(m *ice.Message, file string) {
	if f, _, e := nfs.CreateFile(m, kit.Keys(file, JS)); m.Assert(e) {
		defer f.Close()
		fmt.Fprintln(f, `Volcanos.meta.webpack = true`)
		fmt.Fprintln(f, `Volcanos.meta.pack = `+kit.Formats(kit.UnMarshal(kit.Select("{}", m.Option(nfs.CONTENT)))))
		fmt.Fprintln(f, `Volcanos.meta.args = `+kit.Formats(kit.Dict(m.OptionSimple(kit.Split(m.Option(ctx.ARGS))...))))
	}
	if f, p, e := nfs.CreateFile(m, kit.Keys(file, HTML)); m.Assert(e) {
		defer f.Close()
		defer m.Echo(p)
		main_js := _volcanos(m, PAGE_INDEX_JS)
		if nfs.ExistsFile(m, ice.SRC_MAIN_JS) {
			main_js = ice.SRC_MAIN_JS
		}
		fmt.Fprintf(f, _webpack_template,
			m.Cmdx(nfs.CAT, _volcanos(m, PAGE_INDEX_CSS)), m.Cmdx(nfs.CAT, _volcanos(m, PAGE_CACHE_CSS)),
			m.Cmdx(nfs.CAT, _volcanos(m, ice.PROTO_JS)), m.Cmdx(nfs.CAT, kit.Keys(file, JS)),
			m.Cmdx(nfs.CAT, _volcanos(m, PAGE_CACHE_JS)), m.Cmdx(nfs.CAT, main_js),
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
	PAGE_INDEX_CSS = "page/index.css"
	PAGE_CACHE_CSS = "page/cache.css"
	PAGE_INDEX_JS  = "page/index.js"
	PAGE_CACHE_JS  = "page/cache.js"
	PAGE_CAN_CSS   = "page/can.css"
	PAGE_CAN_JS    = "page/can.js"
)

const DEVPACK = "devpack"
const WEBPACK = "webpack"

func init() {
	Index.MergeCommands(ice.Commands{
		WEBPACK: {Name: "webpack path auto create remove", Help: "打包", Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Name: "create", Help: "发布", Hand: func(m *ice.Message, arg ...string) {
				_webpack_cache(m.Spawn(), _volcanos(m), true)
			}},
			mdb.REMOVE: {Name: "remove", Help: "调试", Hand: func(m *ice.Message, arg ...string) {
				_webpack_cache(m.Spawn(), _volcanos(m), false)
			}},
			mdb.INSERT: {Name: "insert path*", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashCreate(m, m.OptionSimple(nfs.PATH))
			}},
			cli.BUILD: {Name: "build name*=hi", Hand: func(m *ice.Message, arg ...string) {
				_webpack_cache(m.Spawn(), _volcanos(m), true)
				_webpack_build(m, _publish(m, WEBPACK, m.Option(mdb.NAME)))
			}},
			nfs.TRASH: {Hand: func(m *ice.Message, arg ...string) {
				if !strings.Contains(m.Option(nfs.PATH), "page/index") {
					nfs.Trash(m, m.Option(nfs.PATH))
				}
			}},
		}, mdb.HashAction(mdb.SHORT, nfs.PATH, mdb.FIELD, "time,path"), mdb.ClearHashOnExitAction()), Hand: func(m *ice.Message, arg ...string) {
			m.Options(nfs.DIR_TYPE, nfs.CAT, nfs.DIR_DEEP, ice.TRUE)
			m.Cmdy(nfs.DIR, _volcanos(m, PAGE)).Cmdy(nfs.DIR, _publish(m, WEBPACK))
		}},
	})
}

var _webpack_template = `
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
`
