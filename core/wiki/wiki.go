package wiki

import (
	"html"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _name(m *ice.Message, arg []string) []string {
	kit.If(len(arg) == 1, func() { arg = []string{"", arg[0]} })
	return arg
}
func _option(m *ice.Message, kind, name, text string, arg ...string) *ice.Message {
	extra := kit.Dict()
	kit.For(arg, func(k, v string) {
		kit.If(k == cli.FG, func() { k = "style.color" })
		kit.If(k == cli.BG, func() { k = "style.background" })
		kit.Value(extra, k, kit.Format(kit.Parse(nil, "", kit.Split(v)...)))
	})
	m.OptionDefault(mdb.META, kit.Format(extra))
	return m.Options(mdb.TYPE, kind, mdb.NAME, name, mdb.TEXT, text)
}
func _wiki_path(m *ice.Message, arg ...string) string {
	return path.Join(mdb.Config(m, nfs.PATH), path.Join(arg...))
}
func _wiki_link(m *ice.Message, text string) string {
	return web.SharePath(m, text)
}
func _wiki_list(m *ice.Message, arg ...string) bool {
	if m.OptionDefault(nfs.DIR_ROOT, _wiki_path(m)); len(arg) == 0 || kit.HasSuffix(arg[0], nfs.PS) {
		kit.If(m.Option(nfs.DIR_DEEP) != ice.TRUE, func() { m.Cmdy(nfs.DIR, kit.Slice(arg, 0, 1), kit.Dict(nfs.DIR_TYPE, nfs.DIR)) })
		m.Copy(m.Cmd(nfs.DIR, kit.Slice(arg, 0, 1), kit.Dict(nfs.DIR_TYPE, nfs.CAT, nfs.DIR_REG, mdb.Config(m, lex.REGEXP))).SortStr(nfs.PATH))
		return true
	} else {
		m.Display(m.FileURI(nfs.Relative(m, ctx.GetCmdFile(m, m.PrefixKey()))))
		// ctx.DisplayLocal(m, path.Join(kit.PathName(2), kit.Keys(kit.FileName(2), nfs.JS)))
		return false
	}
}
func _wiki_show(m *ice.Message, name string, arg ...string) {
	m.Cmdy(nfs.CAT, name, kit.Dict(nfs.DIR_ROOT, _wiki_path(m)))
}
func _wiki_save(m *ice.Message, name, text string, arg ...string) {
	m.Cmd(nfs.SAVE, name, text, kit.Dict(nfs.DIR_ROOT, _wiki_path(m)))
}
func _wiki_upload(m *ice.Message, dir string) {
	m.Cmdy(web.CACHE, web.WATCH, m.Option(ice.MSG_UPLOAD), _wiki_path(m, dir, m.Option(mdb.NAME)))
}
func _wiki_template(m *ice.Message, file, name, text string, arg ...string) *ice.Message {
	msg := _option(m, m.CommandKey(), name, strings.TrimSpace(text), arg...)
	return m.Echo(nfs.Template(msg, kit.Keys(kit.Select(m.CommandKey(), file), nfs.HTML), &Message{msg}))
}

const WIKI = "wiki"

var Index = &ice.Context{Name: WIKI, Help: "文档中心"}

func init() {
	web.Index.Register(Index, &web.Frame{},
		FEEL, DRAW, DATA, WORD, PORTAL, STYLE,
		TITLE, BRIEF, REFER, SPARK, FIELD,
		ORDER, TABLE, CHART, IMAGE, VIDEO, AUDIO,
	)
}
func Prefix(arg ...string) string { return web.Prefix(WIKI, kit.Keys(arg)) }

func WikiAction(dir string, ext ...string) ice.Actions {
	return ice.Actions{ice.CTX_INIT: mdb.AutoConfig(nfs.PATH, dir, lex.REGEXP, kit.ExtReg(ext...)),
		mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
			switch arg[0] {
			case nfs.PATH:
				m.Option(nfs.DIR_REG, mdb.Config(m, lex.REGEXP))
				m.Cmdy(nfs.DIR, path.Join(mdb.Config(m, nfs.PATH), kit.Select("", arg, 1)))
			default:
				mdb.HashInputs(m, arg)
			}
		}},
		web.UPLOAD: {Hand: func(m *ice.Message, arg ...string) { _wiki_upload(m, m.Option(nfs.PATH)) }},
		nfs.SAVE: {Hand: func(m *ice.Message, arg ...string) {
			_wiki_save(m, m.Option(nfs.PATH), kit.Select(m.Option(mdb.TEXT), arg, 1))
		}},
		nfs.TRASH: {Hand: func(m *ice.Message, arg ...string) {
			nfs.Trash(m, _wiki_path(m, kit.Select("some", kit.Select(m.Option(nfs.PATH), arg, 0))))
		}},
		mdb.SELECT: {Hand: func(m *ice.Message, arg ...string) {
			kit.If(!_wiki_list(m, arg...), func() { _wiki_show(m, arg[0]) })
		}},
	}
}

type Message struct{ *ice.Message }

func (m *Message) OptionTemplate() string {
	res := []string{`class="story"`}
	add := func(pre, key string) {
		kit.If(m.Option(key), func() { res = append(res, kit.Format(`%s%s=%q`, pre, key, html.EscapeString(m.Option(key)))) })
	}
	kit.For(kit.Split("type,name,text,meta"), func(k string) { add("data-", k) })
	kit.For(kit.Split(ctx.STYLE), func(k string) { add("", k) })
	return kit.Join(res, lex.SP)
}
func (m *Message) OptionKV(key ...string) string {
	res := []string{}
	kit.For(kit.Split(kit.Join(key)), func(k string) {
		kit.If(m.Option(k), func() { res = append(res, kit.Format("%s='%s'", k, m.Option(k))) })
	})
	return kit.Join(res, lex.SP)
}
