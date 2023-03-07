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
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _volcanos(m *ice.Message, p ...string) string { return ice.USR_VOLCANOS + path.Join(p...) }
func _publish(m *ice.Message, p ...string) string  { return ice.USR_PUBLISH + path.Join(p...) }
func _require(m *ice.Message, p string) string {
	return path.Join(ice.PS, strings.TrimPrefix(strings.Replace(p, ice.USR_NODE_MODULES, web.REQUIRE_MODULES, 1), ice.USR_VOLCANOS))
}
func _webpack_css(m *ice.Message, css, js io.Writer, p string) {
	fmt.Fprintln(css, kit.Format("/* %s */", _require(m, p)))
	fmt.Fprintln(css, m.Cmdx(nfs.CAT, p))
	_webpack_end(m, js, p)
}
func _webpack_js(m *ice.Message, js io.Writer, p string) {
	fmt.Fprintln(js, `_can_name = "`+_require(m, p)+`";`)
	fmt.Fprintln(js, m.Cmdx(nfs.CAT, p))
}
func _webpack_node(m *ice.Message, js io.Writer, p string) {
	_webpack_js(m, js, p)
	_webpack_end(m, js, p)
}
func _webpack_end(m *ice.Message, js io.Writer, p string) {
	fmt.Fprintln(js, `Volcanos.meta.cache["`+_require(m, p)+`"] = []`)
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
	for _, k := range []string{LIB, PANEL, PLUGIN} {
		nfs.DirDeepAll(m, dir, k, func(value ice.Maps) {
			if kit.Ext(value[nfs.PATH]) == JS {
				_webpack_js(m, js, value[nfs.PATH])
			}
		})
	}
	for _, k := range []string{ice.FRAME_JS} {
		_webpack_js(m, js, _volcanos(m, k))
	}
	mdb.HashSelects(m).Sort(nfs.PATH).Tables(func(value ice.Maps) {
		defer fmt.Fprintln(js, "")
		if p := value[nfs.PATH]; kit.Ext(p) == nfs.CSS {
			_webpack_css(m, css, js, path.Join(ice.USR_NODE_MODULES, p))
		} else {
			p = kit.Select(path.Join(p, LIB, kit.Keys(p, JS)), p, kit.Ext(p) == nfs.JS)
			_webpack_node(m, js, path.Join(ice.USR_NODE_MODULES, p))
		}
	})
}
func _webpack_can(m *ice.Message) {
	m.Cmd(nfs.COPY, USR_PUBLISH_CAN_JS, _volcanos(m, ice.PROTO_JS), _volcanos(m, PAGE_CACHE_JS))
	m.Cmd(nfs.COPY, USR_PUBLISH_CAN_CSS, _volcanos(m, ice.INDEX_CSS), _volcanos(m, PAGE_CACHE_CSS))
}
func _webpack_build(m *ice.Message, name string) {
	if f, p, e := nfs.CreateFile(m, kit.Keys(name, HTML)); m.Assert(e) {
		defer f.Close()
		defer m.Echo(p)
		fmt.Fprintf(f, nfs.Template(m, ice.INDEX_HTML), m.Cmdx(nfs.CAT, USR_PUBLISH_CAN_CSS), m.Cmdx(nfs.CAT, USR_PUBLISH_CAN_JS), kit.JoinKV(ice.EQ, ice.NL,
			`Volcanos.meta.args`, kit.Formats(kit.Dict(m.OptionSimple(kit.Split(m.Option(ctx.ARGS))...))),
			`Volcanos.meta.pack`, kit.Formats(kit.UnMarshal(kit.Select("{}", m.Option(nfs.CONTENT)))),
			`Volcanos.meta.webpack`, ice.TRUE,
		)+ice.NL, m.Cmdx(nfs.CAT, ice.SRC_MAIN_JS))
	}
}

const (
	LIB    = "lib"
	PAGE   = "page"
	PANEL  = "panel"
	PLUGIN = "plugin"
)
const (
	PAGE_CACHE_JS       = "page/cache.js"
	PAGE_CACHE_CSS      = "page/cache.css"
	USR_PUBLISH_CAN_CSS = "usr/publish/can.css"
	USR_PUBLISH_CAN_JS  = "usr/publish/can.js"
)

const DEVPACK = "devpack"
const WEBPACK = "webpack"

func init() {
	Index.MergeCommands(ice.Commands{
		WEBPACK: {Name: "webpack path auto create remove", Help: "打包", Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Help: "发布", Hand: func(m *ice.Message, arg ...string) {
				_webpack_cache(m.Spawn(), _volcanos(m), true)
				_webpack_can(m)
				m.Cmdy("")
			}},
			mdb.REMOVE: {Help: "调试", Hand: func(m *ice.Message, arg ...string) {
				_webpack_cache(m.Spawn(), _volcanos(m), false)
				m.Cmdy(nfs.DIR, _volcanos(m, PAGE))
			}},
			mdb.INSERT: {Name: "insert path*", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashCreate(m, m.OptionSimple(nfs.PATH))
			}},
			cli.BUILD: {Name: "build name*=hi", Hand: func(m *ice.Message, arg ...string) {
				kit.If(!nfs.ExistsFile(m, USR_PUBLISH_CAN_JS), func() { m.Cmd("", mdb.CREATE) })
				_webpack_build(m, _publish(m, m.Option(mdb.NAME)))
			}},
			nfs.TRASH: {Hand: func(m *ice.Message, arg ...string) {
				kit.If(!strings.Contains(m.Option(nfs.PATH), "/page/"), func() { nfs.Trash(m, m.Option(nfs.PATH)) })
			}},
		}, mdb.HashAction(mdb.SHORT, nfs.PATH, mdb.FIELD, "time,path")), Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(nfs.DIR, ice.USR_PUBLISH, kit.Dict(nfs.DIR_REG, kit.ExtReg(HTML, CSS, JS))).Cmdy(nfs.DIR, _volcanos(m, PAGE))
		}},
	})
}
