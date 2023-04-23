package chart

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/core/wiki"
	kit "shylinux.com/x/toolkits"
)

type Chain struct {
	data ice.Map
	Block
}

func (c *Chain) Init(m *ice.Message, arg ...string) wiki.Chart {
	(&c.Block).Init(m)
	const _DEEP = "_deep"
	stack, max := kit.List(kit.Dict(_DEEP, -1, wiki.WIDTH, "0")), 0
	last := func(key string) int { return kit.Int(kit.Value(stack[len(stack)-1], key)) }
	m.Cmd(lex.SPLIT, "", mdb.TEXT, kit.Dict(lex.SPLIT_BLOCK, lex.SP, nfs.CAT_CONTENT, arg[0]), func(deep int, ls []string, data, root ice.Map) {
		for deep <= last(_DEEP) {
			stack = stack[:len(stack)-1]
		}
		width := last(wiki.WIDTH) + c.GetWidths(ls[0])
		if stack = append(stack, kit.Dict(_DEEP, deep, wiki.WIDTH, width)); width > max {
			max = width
		}
		c.data = root
	})
	c.Height, c.Width = c.height(m, c.data)*c.GetHeights(), max
	return c
}
func (c *Chain) Draw(m *ice.Message, x, y int) wiki.Chart {
	gs := wiki.NewGroup(m, SHIP, LINE, RECT, TEXT)
	wiki.AddGroupOption(m, LINE, wiki.STROKE, gs.Option(SHIP, wiki.STROKE))
	wiki.AddGroupOption(m, TEXT, wiki.STROKE, m.Option(wiki.STROKE), wiki.FILL, m.Option(wiki.STROKE))
	defer gs.DumpAll(m, SHIP, LINE, RECT, TEXT)
	c.Height, c.Width = 0, 0
	c.draw(m, c.data, x, y, &c.Block, gs)
	return c
}
func (c *Chain) height(m *ice.Message, root ice.Map) (height int) {
	meta := kit.GetMeta(root)
	if list, ok := root[mdb.LIST].([]ice.Any); ok && len(list) > 0 {
		kit.For(root[mdb.LIST], func(index int, value ice.Map) { height += c.height(m, value) })
	} else {
		height = 1
	}
	meta[wiki.HEIGHT] = height
	return height
}
func (c *Chain) draw(m *ice.Message, root ice.Map, x, y int, p *Block, gs *wiki.Group) int {
	meta := kit.GetMeta(root)
	item := p.Fork(m, kit.Format(meta[mdb.TEXT]))
	item.x, item.y = x, y+(kit.Int(meta[wiki.HEIGHT])-1)*c.GetHeights()/2
	item.Data(m, meta)
	if p != nil && p.y != 0 {
		padding := item.GetHeight() / 2
		kit.If(m.Option(SHOW_BLOCK) == ice.TRUE, func() { padding = 0 })
		x4, y4 := item.x+(p.MarginX+item.MarginX)/4, item.y+item.GetHeights()/2+padding
		x1, y1 := p.x+p.GetWidths()-(p.MarginX+item.MarginX)/4, p.y+p.GetHeights()/2+padding
		gs.EchoPath(SHIP, "M %d,%d Q %d,%d %d,%d T %d %d", x1, y1, x1+(x4-x1)/4, y1, x1+(x4-x1)/2, y1+(y4-y1)/2, x4, y4)
	}
	if m.Option(SHOW_BLOCK) == ice.TRUE {
		gs.EchoRect(RECT, item.GetHeight(), item.GetWidth(), item.x+item.MarginX/2, item.y+item.MarginY/2)
	} else {
		gs.EchoLine(LINE, item.x+item.MarginX/2, item.y+item.GetHeights()-item.MarginY/2, item.x+item.GetWidths()-item.MarginX/2, item.y+item.GetHeights()-item.MarginY/2)
	}
	gs.EchoTexts(TEXT, item.x+item.GetWidths()/2, item.y+item.GetHeights()/2, item.Text, item.TextData)
	h, x := 0, x+item.GetWidths()
	if kit.For(root[mdb.LIST], func(value ice.Map) { h += c.draw(m, value, x, y+h, item, gs) }); h == 0 {
		return item.GetHeights()
	}
	return h
}

const (
	SHOW_BLOCK = "show-block"

	SHIP = "ship"
	LINE = "line"
	RECT = "rect"
	TEXT = "text"
)
const CHAIN = "chain"

func init() {
	wiki.AddChart(CHAIN, func(m *ice.Message) wiki.Chart {
		m.Option(wiki.FONT_SIZE, "18")
		m.Option(wiki.MARGINX, "60")
		m.Option(wiki.MARGINY, "20")
		m.Option(wiki.PADDING, "10")
		wiki.AddGroupOption(m, SHIP, wiki.FILL, cli.GLASS)
		return &Chain{}
	})
}
