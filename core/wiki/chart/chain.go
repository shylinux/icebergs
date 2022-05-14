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
	data map[string]interface{}
	gs   *wiki.Group
	Block
}

func (c *Chain) Init(m *ice.Message, arg ...string) wiki.Chart {
	(&c.Block).Init(m)

	m.Option(nfs.CAT_CONTENT, arg[0])
	m.Option(lex.SPLIT_BLOCK, ice.SP)
	max, stack := 0, kit.List(kit.Dict("_deep", -1, "width", "0"))
	m.OptionCB(lex.SPLIT, func(deep int, ls []string, data map[string]interface{}) []string {
		for deep <= kit.Int(kit.Value(stack[len(stack)-1], "_deep")) {
			stack = stack[:len(stack)-1]
		}
		width := kit.Int(kit.Value(stack[len(stack)-1], "width")) + c.GetWidths(ls[0])
		stack = append(stack, kit.Dict("_deep", deep, "width", width))
		if width > max {
			max = width
		}
		return ls
	})
	c.data = lex.Split(m, "", mdb.TEXT)

	c.Height, c.Width = c.size(m, c.data)*c.GetHeights(), max
	return c
}
func (c *Chain) Draw(m *ice.Message, x, y int) wiki.Chart {
	c.gs = wiki.NewGroup(m, SHIP, RECT, LINE, TEXT)
	wiki.AddGroupOption(m, LINE, wiki.STROKE, c.gs.Option(SHIP, wiki.STROKE))
	wiki.AddGroupOption(m, TEXT, wiki.FILL, m.Option(wiki.STROKE), wiki.STROKE_WIDTH, "1")
	defer func() { c.gs.Dump(m, SHIP).Dump(m, RECT).Dump(m, LINE).Dump(m, TEXT) }()

	c.Height, c.Width = 0, 0
	c.draw(m, c.data, x, y, &c.Block)
	return c
}

func (c *Chain) size(m *ice.Message, root map[string]interface{}) (height int) {
	meta := kit.GetMeta(root)
	if list, ok := root[mdb.LIST].([]interface{}); ok && len(list) > 0 {
		kit.Fetch(root[mdb.LIST], func(index int, value map[string]interface{}) { height += c.size(m, value) })
	} else {
		height = 1
	}
	meta[wiki.HEIGHT] = height
	return height
}
func (c *Chain) draw(m *ice.Message, root map[string]interface{}, x, y int, p *Block) int {
	meta := kit.GetMeta(root)

	item := &Block{FontSize: p.FontSize, Padding: p.Padding, MarginX: p.MarginX, MarginY: p.MarginY}
	item.x, item.y = x, y+(kit.Int(meta[wiki.HEIGHT])-1)*c.GetHeights()/2
	item.Text = kit.Format(meta[mdb.TEXT])
	if m.Option(SHOW_BLOCK) == ice.TRUE { // 方框
		c.gs.EchoRect(RECT, item.GetHeight(), item.GetWidth(), item.x+item.MarginX/2, item.y+item.MarginY/2)
	} else { // 横线
		c.gs.EchoLine(LINE, item.x+item.MarginX/2, item.y+item.GetHeight()+item.Padding/2, item.x+item.GetWidths()-item.MarginX/2, item.y+item.GetHeight()+item.Padding/2)
	}

	// 文本
	c.gs.EchoTexts(TEXT, item.x+item.GetWidths()/2, item.y+item.GetHeights()/2, kit.Format(meta[mdb.TEXT]))

	if p != nil && p.y != 0 { // 连线
		if m.Option(SHOW_BLOCK) == ice.TRUE {
			x1, y1 := p.x+p.GetWidths()-(p.MarginX+item.MarginX)/4, p.y+p.GetHeights()/2
			x4, y4 := item.x+(p.MarginX+item.MarginX)/4, item.y+item.GetHeights()/2
			c.gs.Echo(SHIP, `<path d="M %d,%d Q %d,%d %d,%d T %d %d"></path>`, x1, y1, x1+(x4-x1)/4, y1, x1+(x4-x1)/2, y1+(y4-y1)/2, x4, y4)
		} else {
			x1, y1 := p.x+p.GetWidths()-(p.MarginX+item.MarginX)/4, p.y+p.GetHeight()+item.Padding/2
			x4, y4 := item.x+(p.MarginX+item.MarginX)/4, item.y+item.GetHeight()+item.Padding/2
			c.gs.Echo(SHIP, `<path d="M %d,%d Q %d,%d %d,%d T %d %d"></path>`, x1, y1, x1+(x4-x1)/4, y1, x1+(x4-x1)/2, y1+(y4-y1)/2, x4, y4)
		}
	}

	// 递归
	h, x := 0, x+item.GetWidths()
	if kit.Fetch(root[mdb.LIST], func(index int, value map[string]interface{}) {
		h += c.draw(m, value, x, y+h, item)
	}); h == 0 {
		return item.GetHeights()
	}
	return h
}

const (
	SHIP = "ship"
)
const CHAIN = "chain"

func init() {
	wiki.AddChart(CHAIN, func(m *ice.Message) wiki.Chart {
		m.Option(wiki.STROKE_WIDTH, "2")
		m.Option(wiki.STROKE, cli.BLUE)
		m.Option(wiki.MARGINX, "40")
		m.Option(wiki.MARGINY, "10")
		wiki.AddGroupOption(m, SHIP, wiki.STROKE, cli.RED, wiki.FILL, cli.GLASS)
		return &Chain{}
	})
}
