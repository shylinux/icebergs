package wiki

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
)

func _spark_md(m *ice.Message, arg ...string) *ice.Message {
	block, code := "", []string{}
	show := func(arg ...string) {
		kit.If(len(code) > 0, func() { m.Cmdy(SPARK, arg, kit.Join(code, lex.NL)); code = []string{} })
	}
	defer show()
	m.Cmd(nfs.CAT, arg[0], func(line string) {
		for _, ls := range [][]string{[]string{"# ", TITLE}, []string{"## ", TITLE, CHAPTER}, []string{"### ", TITLE, SECTION}} {
			if strings.HasPrefix(line, ls[0]) {
				show()
				m.Cmdy(ls[1:], strings.TrimPrefix(line, ls[0]))
				return
			}
		}
		if strings.HasPrefix(line, "```") {
			if block == "" {
				show()
				block = "```"
			} else {
				show(SHELL)
			}
		} else {
			code = append(code, line)
		}
	})
	return m
}
func _spark_tabs(m *ice.Message, arg ...string) {
	defer m.Echo(`<div class="story" data-type="spark_tabs">`).Echo(`</div>`)
	func() {
		defer m.Echo(`<div class="tabs">`).Echo(`</div>`)
		kit.For(arg[1:], func(k, v string) { m.Echo(`<div class="item">%s</div>`, k) })
	}()
	kit.For(arg[1:], func(k, v string) { m.Cmdy("", arg[0], v) })
}
func _spark_show(m *ice.Message, name, text string, arg ...string) *ice.Message {
	if m.IsCliUA() {
		kit.For(kit.SplitLine(text), func(text string) { m.Echo(text).Echo("\n") })
		return m
	}
	return _wiki_template(m.Options(mdb.LIST, kit.SplitLine(text)), name, name, text, arg...)
}

const (
	INNER   = "inner"
	SHELL   = "shell"
	ROUTE   = "route"
	PROJECT = "project"
	PRODUCT = "product"
)

const SPARK = "spark"

func init() {
	Index.MergeCommands(ice.Commands{
		SPARK: {Name: "spark type=inner,shell,redis,mysql text", Help: "段落", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				ice.AddRender(ice.RENDER_SCRIPT, func(msg *ice.Message, args ...ice.Any) string {
					return msg.Cmdx(SPARK, SHELL, args)
				})
			}},
			ice.DEMO: {Hand: func(m *ice.Message, arg ...string) {
				if aaa.Right(m.Spawn(), arg[0]) {
					m.Cmdy(FIELD, "", arg[0], arg[1:])
				} else {
					p := kit.Format("https://demo.shylinux.com/c/%s", arg[0])
					m.Cmdy(SPARK, p, arg[1:]).Cmdy("web.wiki.iframe", p, arg[1:])
				}
			}},
			INNER: {Hand: func(m *ice.Message, arg ...string) {
				if strings.Contains(arg[0], lex.NL) {

				} else if nfs.Exists(m, arg[0]) {
					arg = append(arg, kit.Simple(ctx.ARGS, kit.Join(nfs.SplitPath(m, arg[0]), lex.SP))...)
					arg[0] = m.Cmdx(nfs.CAT, arg[0])
				} else if p := path.Join(path.Dir(m.Option(ice.MSG_SCRIPT)), arg[0]); nfs.Exists(m, p) {
					arg = append(arg, kit.Simple(ctx.ARGS, kit.Join(nfs.SplitPath(m, p), lex.SP))...)
					arg[0] = m.Cmdx(nfs.CAT, p)
				}
				m.Cmdy(FIELD, "", web.CODE_INNER, ice.MSG_RESULT, arg[0], ctx.DISPLAY, html.PLUGIN_INNER, ctx.STYLE, html.OUTPUT, arg[1:])
			}},
			PROJECT: {Hand: func(m *ice.Message, arg ...string) { _spark_project(m, arg...) }},
			PRODUCT: {Hand: func(m *ice.Message, arg ...string) { _spark_product(m, arg...) }},
		}), Hand: func(m *ice.Message, arg ...string) {
			m.Option(mdb.META, "")
			if kit.Ext(arg[0]) == "md" {
				_spark_md(m, arg...)
			} else if arg[0] == SHELL && kit.IsIn(kit.Select("", arg, 1), cli.ALPINE, "centos", cli.LINUX, cli.DARWIN, cli.MACOS, cli.WINDOWS) {
				_spark_tabs(m, arg...)
			} else {
				if arg = _name(m, arg); arg[0] == SHELL && len(arg) > 3 && arg[2] == "with" && arg[3] == cli.ECHO {
					m.Option(cli.ECHO, m.Cmdx(cli.SYSTEM, kit.Split(arg[1])))
				}
				_spark_show(m, arg[0], strings.TrimSpace(kit.Renders(arg[1], ice.Info)), arg[2:]...)
			}
		}},
	})
}
func _spark_project(m *ice.Message, arg ...string) {
	func() {
		defer m.Cmdy(STYLE, FLEX).Cmdy(STYLE, END)
		func() {
			defer m.Cmdy(STYLE, COLUMN).Cmdy(STYLE, END)
			m.Cmdy(TITLE, ice.Info.Title())
			m.Cmdy(SPARK, TITLE, arg[0]).Cmdy(ORDER, arg[1])
			func() {
				defer m.Cmdy(STYLE, FLEX).Cmdy(STYLE, END)
				m.Cmdy(SPARK, html.BUTTON, "体 验", ROUTE, web.SpideOrigin(m, ice.DEMO))
				m.Cmdy(SPARK, html.BUTTON, "下 载", ROUTE, "download/")
				m.Cmdy(SPARK, html.BUTTON, "文 档", ROUTE, "started/")
			}()
		}()
		func() {
			defer m.Cmdy(STYLE, COLUMN, FLEX, "0 0 500px", "padding", "10px").Cmdy(STYLE, END)
			m.Cmdy(SPARK, INNER, ice.SRC_MAIN_GO, html.WIDTH, "480px")
			m.Cmdy(SPARK, SHELL, kit.Renders(`
git clone {{ .Make.Remote }}
cd {{ .Make.Remote | Base }} && source etc/miss.sh

./bin/ice.bin

open http://localhost:9020
`, ice.Info), "style.width", "480px")
		}()
	}()
	if nfs.Exists(m, "src/qrcode.jpg") {
		defer m.Echo(`<div class="story" data-type="qrcode">`).Echo("</div>")
		m.Cmdy(IMAGE, "qrcode", "src/qrcode.jpg")
		m.Cmdy(SPARK, "请使用微信扫码，打开公众号体验服务")
	}
}
func _spark_product(m *ice.Message, arg ...string) {
	if len(arg) == 0 {
		m.Cmd("web.product").Table(func(value ice.Maps) {
			if value[mdb.DISABLE] == ice.TRUE {
				_spark_product(m, value[ctx.INDEX], value[mdb.NAME], value[mdb.TEXT])
			}
		})
		return
	}
	m.Cmdy(SPARK, TITLE, arg[1]).Cmdy(SPARK, arg[2])
	m.Cmdy(FIELD, arg[1], arg[0], arg[3:])
}
