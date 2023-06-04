package node

import (
	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

type vue struct {
	ice.Code
	ice.Lang
	list string `name:"list path auto" help:"框架"`
}

func (s vue) Render(m *ice.Message, arg ...string) {
	list := kit.Dict(kit.UnMarshal(m.Cmdx(nfs.CAT, kit.Path(m.Option(nfs.PATH), "display.json"))))
	m.EchoIFrame(kit.Format("%s%s", list[tcp.HOST], kit.Format(list[m.Option(nfs.FILE)])))
}
func (s vue) Init(m *ice.Message) {
	s.Lang.Init(m, code.PREPARE, ice.Map{
		code.KEYWORD: kit.List(
			"template", "script", "style",
			"router-link", "router-view",
			"el-container", "el-aside", "el-header", "el-main",

			"el-tabs",
			"el-tab-pane",
			"el-dialog",
			"el-form",
			"el-form-item",

			"el-input",
			"el-select",
			"el-option",
			"el-button",
			"el-checkbox-group",
			"el-checkbox",
			"el-pagination",

			"el-menu",
			"el-submenu",
			"el-menu-item",

			"el-table",
			"el-table-column",
		),
		code.FUNCTION: kit.List(
			"data",
			"props",
			"inject",
			"provide",
			"components",
			"computed",
			"created",
			"mounted",
			"methods",
			"watch",
		),
	}, "include", kit.List(nfs.HTML, nfs.CSS, nfs.JS), "split.operator", "{[(.,:</>#)]}")
}
func (s vue) List(m *ice.Message) { m.Cmdy(nfs.DIR, nfs.USR) }

func init() { ice.CodeCtxCmd(vue{}) }
