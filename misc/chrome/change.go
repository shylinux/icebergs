package chrome

import (
	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/core/wiki"
	kit "shylinux.com/x/toolkits"
)

type change struct {
	ice.Hash
	operate

	short string `data:"property"`
	list  string `name:"list wid tid selector:text@key property:textarea@key auto export import" help:"编辑"`
}

func (c change) Inputs(m *ice.Message, arg ...string) {
	switch arg[0] {
	case SELECTOR:
		m.Push(arg[0], wiki.VIDEO)
		fallthrough
	default:
		c.Hash.Inputs(m, arg...)
	}
}
func (c change) List(m *ice.Message, arg ...string) {
	if len(arg) < 2 || arg[2] == "" {
		c.send(m, kit.Slice(arg, 0, 2))
		return
	}
	c.send(m.Spawn(), kit.Slice(arg, 0, 2), m.CommandKey(), kit.Slice(arg, 2)).Table(func(index int, value ice.Maps, head []string) {
		m.Push(mdb.TEXT, kit.ReplaceAll(value[mdb.TEXT], "<", "&lt;", ">", "&gt;"))
	})
	if len(arg) > 3 {
		c.Hash.Create(m, SELECTOR, arg[2], PROPERTY, arg[3])
	}
}

const (
	SELECTOR = "selector"
	PROPERTY = "property"
)

func init() { ice.CodeCtxCmd(change{}) }
