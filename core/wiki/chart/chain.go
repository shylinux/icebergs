package chart

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/core/wiki"
	kit "shylinux.com/x/toolkits"
)

type Chain struct {
	data  map[string]interface{}
	Group *wiki.Group
	Block
}

func (c *Chain) Init(m *ice.Message, arg ...string) wiki.Chart {
	(&c.Block).Init(m)

	// 解析数据
	m.Option(nfs.CAT_CONTENT, arg[0])
	m.Option(lex.SPLIT_SPACE, "\t \n")
	m.Option(lex.SPLIT_BLOCK, "\t \n")
	c.data = lex.Split(m, "", kit.MDB_TEXT)

	// 计算尺寸
	c.Height = c.size(m, c.data) * c.GetHeights()
	c.Draw(m, 0, 0)
	c.Width += 200
	m.Set(ice.MSG_RESULT)
	return c
}
func (c *Chain) Draw(m *ice.Message, x, y int) wiki.Chart {
	c.Group = wiki.NewGroup(m, SHIP)
	defer c.Group.Dump(m, SHIP)
	c.draw(m, c.data, x, y, &c.Block)
	return c
}
func (c *Chain) size(m *ice.Message, root map[string]interface{}) (height int) {
	meta := kit.GetMeta(root)
	if list, ok := root[kit.MDB_LIST].([]interface{}); ok && len(list) > 0 {
		kit.Fetch(root[kit.MDB_LIST], func(index int, value map[string]interface{}) {
			height += c.size(m, value)
		})
	} else {
		height = 1
	}
	meta[wiki.HEIGHT] = height
	return height
}
func (c *Chain) draw(m *ice.Message, root map[string]interface{}, x, y int, p *Block) int {
	meta := kit.GetMeta(root)
	c.Height, c.Width = 0, 0

	if kit.Format(meta[wiki.FG]) != "" {
		items := wiki.NewItem([]string{"<g"})
		items.Push("stroke=%s", meta[wiki.FG])
		items.Push("fill=%s", meta[wiki.FG])
		items.Echo(">").Dump(m)
		defer m.Echo("</g>")
	}

	// 当前节点
	item := &Block{
		FontSize: p.FontSize,
		Padding:  p.Padding,
		MarginX:  p.MarginX,
		MarginY:  p.MarginY,
	}
	item.x, item.y = x, y+(kit.Int(meta[wiki.HEIGHT])-1)*c.GetHeights()/2
	item.Init(m, kit.Format(meta[kit.MDB_TEXT])).Data(m, meta)
	item.Draw(m, item.x, item.y)

	// 画面尺寸
	if item.y+item.GetHeight()+c.MarginY > c.Height {
		c.Height = item.y + item.GetHeight() + c.MarginY
	}
	if item.x+item.GetWidth()+c.MarginX > c.Width {
		c.Width = item.x + item.GetWidth() + c.MarginX
	}

	// 模块连线
	if p != nil && p.y != 0 {
		x1, y1 := p.x+p.GetWidths()-(p.MarginX+item.MarginX)/4, p.y+p.GetHeights()/2
		x4, y4 := item.x+(p.MarginX+item.MarginX)/4, item.y+item.GetHeights()/2
		c.Group.Echo(SHIP, `<path d="M %d,%d Q %d,%d %d,%d T %d %d"></path>`,
			x1, y1, x1+(x4-x1)/4, y1, x1+(x4-x1)/2, y1+(y4-y1)/2, x4, y4)
	}

	// 递归节点
	h, x := 0, x+item.GetWidths()
	if kit.Fetch(root[kit.MDB_LIST], func(index int, value map[string]interface{}) {
		h += c.draw(m, value, x, y+h, item)
	}); h == 0 {
		return item.GetHeights()
	}
	return h
}

const (
	SHIP = "ship"

	HIDE_BLOCK = "hide-block"
)
const CHAIN = "chain"

func init() {
	wiki.AddChart(CHAIN, func(m *ice.Message) wiki.Chart {
		m.Option(wiki.STROKE_WIDTH, "1")
		m.Option(wiki.FILL, cli.BLUE)
		m.Option(wiki.MARGINX, "40")
		m.Option(wiki.MARGINY, "0")

		m.Option(COMPACT, ice.TRUE)
		m.Option(HIDE_BLOCK, ice.TRUE)
		wiki.AddGroupOption(m, SHIP, wiki.STROKE_WIDTH, "1", wiki.STROKE, cli.CYAN, wiki.FILL, "none")
		return &Chain{}
	})
}
