package wiki

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/ctx"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/base/ssh"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"
)

func _name(m *ice.Message, arg []string) []string {
	if len(arg) == 1 {
		return []string{"", arg[0]}
	}
	return arg
}
func _option(m *ice.Message, kind, name, text string, arg ...string) {
	m.Option(kit.MDB_TYPE, kind)
	m.Option(kit.MDB_NAME, name)
	m.Option(kit.MDB_TEXT, text)

	extra := kit.Dict()
	m.Optionv(kit.MDB_EXTRA, extra)
	for i := 0; i < len(arg); i += 2 {
		extra[arg[i]] = kit.Format(kit.Parse(nil, "", kit.Split(arg[i+1])...))
	}
}

func _word_show(m *ice.Message, name string, arg ...string) {
	m.Set(ice.MSG_RESULT)
	m.Option(TITLE, map[string]int{})
	m.Option(kit.MDB_MENU, kit.Dict(kit.MDB_LIST, []interface{}{}))

	m.Option(ice.MSG_ALIAS, m.Confv(WORD, kit.Keym(kit.MDB_ALIAS)))
	m.Option(nfs.DIR_ROOT, _wiki_path(m, WORD))
	m.Option(ice.MSG_RENDER, ice.RENDER_RAW)
	m.Cmdy(ssh.SOURCE, name)
}

func _word_template(m *ice.Message, cmd string, arg ...string) {
	arg = _name(m, arg)
	_wiki_template(m, cmd, arg[0], kit.Select(arg[0], arg[1]), arg[2:]...)
}

const WORD = "word"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			WORD: {Name: WORD, Help: "语言文字", Value: kit.Data(
				kit.MDB_PATH, "", kit.MDB_REGEXP, ".*\\.shy", kit.MDB_ALIAS, kit.Dict(
					PREMENU, []interface{}{TITLE, PREMENU},
					CHAPTER, []interface{}{TITLE, CHAPTER},
					SECTION, []interface{}{TITLE, SECTION},
					ENDMENU, []interface{}{TITLE, ENDMENU},
					LABEL, []interface{}{CHART, LABEL},
					CHAIN, []interface{}{CHART, CHAIN},
				),
			)},
		},
		Commands: map[string]*ice.Command{
			WORD: {Name: "word path=src/main.shy auto 演示", Help: "语言文字", Meta: kit.Dict(
				ice.Display("/plugin/local/wiki/word.js", WORD),
			), Action: map[string]*ice.Action{
				web.STORY: {Name: "story", Help: "运行", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(arg[0], ctx.ACTION, cli.RUN, arg[1:])
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(nfs.DIR_REG, m.Conf(WORD, kit.Keym(kit.MDB_REGEXP)))
				if m.Option(nfs.DIR_DEEP, ice.TRUE); !_wiki_list(m, cmd, arg...) {
					_word_show(m, arg[0])
				}
			}},
		},
	})
}
