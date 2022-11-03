package wiki

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _name(m *ice.Message, arg []string) []string {
	if len(arg) == 1 {
		return []string{"", arg[0]}
	}
	return arg
}
func _option(m *ice.Message, kind, name, text string, arg ...string) *ice.Message {
	m.Option(mdb.TYPE, kind)
	m.Option(mdb.NAME, name)
	m.Option(mdb.TEXT, text)

	extra := kit.Dict()
	m.Optionv(mdb.EXTRA, extra)
	for i := 0; i < len(arg); i += 2 {
		extra[arg[i]] = kit.Format(kit.Parse(nil, "", kit.Split(arg[i+1])...))
	}
	return m
}

func _wiki_path(m *ice.Message, arg ...string) string {
	return path.Join(m.Config(nfs.PATH), path.Join(arg...))
}
func _wiki_link(m *ice.Message, text string) string {
	if !strings.HasPrefix(text, ice.PS) && !strings.HasPrefix(text, ice.HTTP) {
		text = path.Join(web.SHARE_LOCAL, _wiki_path(m, text))
	}
	return text
}
func _wiki_list(m *ice.Message, arg ...string) bool {
	if m.Option(nfs.DIR_ROOT, _wiki_path(m)); len(arg) == 0 || strings.HasSuffix(arg[0], ice.PS) {
		if m.Option(nfs.DIR_DEEP) != ice.TRUE {
			m.Cmdy(nfs.DIR, kit.Slice(arg, 0, 1), kit.Dict(nfs.DIR_TYPE, nfs.DIR))
		}
		m.Cmdy(nfs.DIR, kit.Slice(arg, 0, 1), kit.Dict(nfs.DIR_TYPE, nfs.CAT, nfs.DIR_REG, m.Config(lex.REGEXP)))
		return true
	}
	ctx.DisplayLocal(m, path.Join(kit.PathName(2), kit.Keys(kit.FileName(2), ice.JS)))
	return false
}
func _wiki_show(m *ice.Message, name string, arg ...string) {
	m.Cmdy(nfs.CAT, name, kit.Dict(nfs.DIR_ROOT, _wiki_path(m)))
}
func _wiki_save(m *ice.Message, name, text string, arg ...string) {
	m.Cmd(nfs.SAVE, name, text, kit.Dict(nfs.DIR_ROOT, _wiki_path(m)))
}
func _wiki_upload(m *ice.Message, dir string) {
	m.Cmdy(web.CACHE, web.UPLOAD_WATCH, _wiki_path(m, dir))
}
func _wiki_template(m *ice.Message, name, text string, arg ...string) *ice.Message {
	return _option(m, m.CommandKey(), name, strings.TrimSpace(text), arg...).RenderTemplate(m.Config(nfs.TEMPLATE), &Message{m})
}

const WIKI = "wiki"

var Index = &ice.Context{Name: WIKI, Help: "文档中心"}

func init() {
	web.Index.Register(Index, &web.Frame{},
		TITLE, BRIEF, REFER, SPARK, FIELD, PARSE,
		ORDER, TABLE, CHART, IMAGE, VIDEO, AUDIO,
		FEEL, DRAW, WORD, DATA,
	)
}

func WikiAction(dir string, ext ...string) ice.Actions {
	return ice.Actions{
		ice.CTX_INIT: &ice.Action{Hand: func(m *ice.Message, arg ...string) {
			if cs := m.Target().Configs; cs[m.CommandKey()] == nil {
				cs[m.CommandKey()] = &ice.Config{Value: kit.Data()}
				ice.Info.Load(m, m.CommandKey())
			}
			m.Config(nfs.PATH, dir)
			m.Config(lex.REGEXP, kit.FileReg(ext...))
		}},
		nfs.TRASH: {Hand: func(m *ice.Message, arg ...string) {
			m.Cmd(nfs.TRASH, path.Join(m.Config(nfs.PATH), m.Option(nfs.PATH)))
		}},
		nfs.SAVE: {Name: "save path text", Hand: func(m *ice.Message, arg ...string) {
			m.Cmd(nfs.SAVE, arg[0], arg[1], kit.Dict(nfs.DIR_ROOT, m.Config(nfs.PATH)))
		}},
		web.UPLOAD: {Hand: func(m *ice.Message, arg ...string) {
			_wiki_upload(m, m.Option(nfs.PATH))
		}},
	}
}

type Message struct{ *ice.Message }

func (m *Message) OptionTemplate() string {
	res := []string{`class="story"`}
	add := func(pre, key string) {
		if m.Option(key) != "" {
			res = append(res, kit.Format(`%s%s="%s"`, pre, key, m.Option(key)))
		}
	}
	for _, key := range kit.Split("type,name,text") {
		if key == mdb.TEXT && m.Option(mdb.TYPE) == SPARK {
			continue
		}
		add("data-", key)
	}
	kit.Fetch(m.Optionv(mdb.EXTRA), func(key string, value string) {
		add("data-", key)
	})
	for _, key := range kit.Split(ctx.STYLE) {
		add("", key)
	}
	return kit.Join(res, ice.SP)
}
