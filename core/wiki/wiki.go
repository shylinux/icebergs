package wiki

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
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
func _option(m *ice.Message, kind, name, text string, arg ...string) {
	m.Option(mdb.TYPE, kind)
	m.Option(mdb.NAME, name)
	m.Option(mdb.TEXT, text)

	extra := kit.Dict()
	m.Optionv(mdb.EXTRA, extra)
	for i := 0; i < len(arg); i += 2 {
		extra[arg[i]] = kit.Format(kit.Parse(nil, "", kit.Split(arg[i+1])...))
	}
}

func _wiki_path(m *ice.Message, cmd string, arg ...string) string {
	return path.Join(m.Conf(cmd, kit.Keym(nfs.PATH)), path.Join(arg...))
}
func _wiki_link(m *ice.Message, cmd string, text string) string {
	if !strings.HasPrefix(text, ice.HTTP) && !strings.HasPrefix(text, ice.PS) {
		text = path.Join(web.SHARE_LOCAL, _wiki_path(m, cmd, text))
	}
	return text
}
func _wiki_list(m *ice.Message, cmd string, arg ...string) bool {
	m.Option(nfs.DIR_ROOT, _wiki_path(m, cmd))
	if len(arg) == 0 || strings.HasSuffix(arg[0], ice.PS) {
		if m.Option(nfs.DIR_DEEP) != ice.TRUE { // 目录列表
			m.Option(nfs.DIR_TYPE, nfs.DIR)
			m.Cmdy(nfs.DIR, kit.Select(nfs.PWD, arg, 0))
		}

		// 文件列表
		m.Option(nfs.DIR_TYPE, nfs.CAT)
		m.Cmdy(nfs.DIR, kit.Select(nfs.PWD, arg, 0))
		return true
	}
	return false
}
func _wiki_show(m *ice.Message, cmd, name string, arg ...string) {
	m.Option(nfs.DIR_ROOT, _wiki_path(m, cmd))
	m.Cmdy(nfs.CAT, name)
}
func _wiki_save(m *ice.Message, cmd, name, text string, arg ...string) {
	m.Option(nfs.DIR_ROOT, _wiki_path(m, cmd))
	m.Cmd(nfs.SAVE, name, text)
}
func _wiki_upload(m *ice.Message, cmd string, dir string) {
	m.Cmdy(web.CACHE, web.UPLOAD_WATCH, _wiki_path(m, cmd, dir))
}
func _wiki_template(m *ice.Message, cmd string, name, text string, arg ...string) {
	_option(m, cmd, name, strings.TrimSpace(text), arg...)
	m.RenderTemplate(m.Conf(cmd, kit.Keym(nfs.TEMPLATE)), &Message{m})
}

const WIKI = "wiki"

var Index = &ice.Context{Name: WIKI, Help: "文档中心"}

func init() {
	web.Index.Register(Index, &web.Frame{},
		TITLE, BRIEF, REFER, SPARK, CHART,
		ORDER, TABLE, IMAGE, VIDEO,
		FIELD, SHELL, LOCAL, PARSE,
		FEEL, DRAW, WORD, DATA,
	)
}

type Message struct {
	*ice.Message
}

func (m *Message) OptionTemplate() string {
	res := []string{`class="story"`}
	for _, key := range kit.Split(ctx.STYLE) {
		if m.Option(key) != "" {
			res = append(res, kit.Format(`s="%s"`, key, m.Option(key)))
		}
	}
	for _, key := range kit.Split("type,name,text") {
		if key == mdb.TEXT && m.Option(mdb.TYPE) == "spark" {
			continue
		}
		if m.Option(key) != "" {
			res = append(res, kit.Format(`data-%s="%s"`, key, m.Option(key)))
		}
	}
	kit.Fetch(m.Optionv("extra"), func(key string, value string) {
		if value != "" {
			res = append(res, kit.Format(`data-%s="%s"`, key, value))
		}
	})
	return kit.Join(res, ice.SP)
}
