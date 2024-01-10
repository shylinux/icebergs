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
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _volcanos(m *ice.Message, p ...string) string { return ice.USR_VOLCANOS + path.Join(p...) }
func _publish(m *ice.Message, p ...string) string  { return ice.USR_PUBLISH + path.Join(p...) }
func _require(m *ice.Message, p string) string {
	if kit.HasPrefix(p, nfs.USR_MODULES) {
		return path.Join("/require/modules/", strings.TrimPrefix(p, nfs.USR_MODULES))
	} else if kit.HasPrefix(p, ice.USR_VOLCANOS) {
		return path.Join("/volcanos/", strings.TrimPrefix(p, ice.USR_VOLCANOS))
	} else if kit.HasPrefix(p, nfs.SRC, nfs.USR) {
		return path.Join("/require/", p)
	} else {
		return path.Join("/volcanos/", p)
	}
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
	if m.Option(ice.MSG_USERPOD) != "" {
		return
	} else if _, e := nfs.DiskFile.StatFile(ice.USR_VOLCANOS); os.IsNotExist(e) {
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
	kit.For([]string{LIB, PANEL, PLUGIN}, func(k string) {
		nfs.DirDeepAll(m, dir, k, func(value ice.Maps) {
			kit.If(kit.Ext(value[nfs.PATH]) == CSS, func() {
				_webpack_css(m, css, js, value[nfs.PATH])
			})
		})
	})
	kit.For([]string{LIB, PANEL, PLUGIN}, func(k string) {
		nfs.DirDeepAll(m, dir, k, func(value ice.Maps) {
			kit.If(kit.Ext(value[nfs.PATH]) == JS, func() {
				_webpack_js(m, js, value[nfs.PATH])
			})
		})
	})
	kit.For([]string{ice.FRAME_JS}, func(k string) {
		_webpack_js(m, js, path.Join(dir, k))
	})
	m.Cmd(nfs.DIR, "src/template/web.chat.header/theme/", func(value ice.Maps) {
		_webpack_css(m, css, js, value[nfs.PATH])
	})
	mdb.HashSelects(m).Sort(nfs.PATH).Table(func(value ice.Maps) {
		defer fmt.Fprintln(js, "")
		if p := value[nfs.PATH]; kit.Ext(p) == nfs.CSS {
			_webpack_css(m, css, js, path.Join(nfs.USR_MODULES, p))
		} else {
			p = kit.Select(path.Join(p, LIB, kit.Keys(p, JS)), p, kit.Ext(p) == nfs.JS)
			_webpack_node(m, js, path.Join(nfs.USR_MODULES, p))
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
		fmt.Fprintf(f, nfs.Template(m, "index.html"), m.Cmdx(nfs.CAT, USR_PUBLISH_CAN_CSS), m.Cmdx(nfs.CAT, USR_PUBLISH_CAN_JS), kit.JoinKV(mdb.EQ, lex.NL,
			`Volcanos.meta.args`, kit.Formats(kit.Dict(m.OptionSimple(kit.Split(m.Option(ctx.ARGS))...))),
			`Volcanos.meta.pack`, kit.Formats(kit.UnMarshal(kit.Select("{}", m.Option(nfs.CONTENT)))),
			`Volcanos.meta.webpack`, ice.TRUE,
		)+lex.NL, m.Cmdx(nfs.CAT, ice.SRC_MAIN_JS))
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
			mdb.CREATE: {Name: "create path", Hand: func(m *ice.Message, arg ...string) {
				_webpack_cache(m.Spawn(), _volcanos(m), true)
				// _webpack_can(m)
				m.Cmdy("")
			}},
			mdb.REMOVE: {Hand: func(m *ice.Message, arg ...string) {
				_webpack_cache(m.Spawn(), _volcanos(m), false)
				m.Cmdy("")
			}},
			mdb.INSERT: {Name: "insert path*", Hand: func(m *ice.Message, arg ...string) { mdb.HashCreate(m) }},
			cli.BUILD: {Name: "build name*=hi", Hand: func(m *ice.Message, arg ...string) {
				// kit.If(!nfs.Exists(m, USR_PUBLISH_CAN_JS), func() { m.Cmd("", mdb.CREATE) })
				_webpack_build(m, _publish(m, m.Option(mdb.NAME)))
			}},
			nfs.TRASH: {Hand: func(m *ice.Message, arg ...string) {
				kit.If(!strings.Contains(m.Option(nfs.PATH), "/page/"), func() { nfs.Trash(m, m.Option(nfs.PATH)) })
			}},
		}, mdb.HashAction(mdb.SHORT, nfs.PATH, mdb.FIELD, "time,path")), Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(nfs.DIR, _volcanos(m, PAGE), kit.Dict(nfs.DIR_REG, kit.ExtReg(HTML, CSS, JS))).Cmdy(nfs.DIR, ice.USR_PUBLISH)
		}},
	})
}
