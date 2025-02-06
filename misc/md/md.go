package md

import (
	"path"

	"github.com/gomarkdown/markdown"
	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/nfs"
)

type md struct {
	ice.Code
	ice.Lang
	list string `name:"list name auto" help:"示例"`
}

func (s md) Init(m *ice.Message, arg ...string) {
	s.Lang.Init(m, nfs.SCRIPT, m.Resource(""))
}
func (s md) List(m *ice.Message, arg ...string) {
	s.Code.Source(m, "", arg...)
}
func (s md) Render(m *ice.Message, arg ...string) {
	md := []byte(m.Cmdx(nfs.CAT, path.Join(arg[2], arg[1])))
	html := markdown.ToHTML(md, nil, nil)
	m.Echo(string(html))
}
func (s md) Engine(m *ice.Message, arg ...string) {
	md := []byte(m.Cmdx(nfs.CAT, path.Join(arg[2], arg[1])))
	html := markdown.ToHTML(md, nil, nil)
	m.Echo(string(html))
}

func init() { ice.Cmd("web.wiki.md", md{}) }
