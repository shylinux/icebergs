package wiki

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/toolkits"
	"strings"
)

// 图形接口
type Chart interface {
	Init(*ice.Message, ...string) Chart
	Draw(*ice.Message, int, int) Chart

	GetWidth(...string) int
	GetHeight(...string) int
}

// 图形基类
type Block struct {
	Text       string
	FontColor  string
	FontFamily string
	BackGround string

	FontSize int
	LineSize int
	Padding  int
	Margin   int

	Width, Height int

	TextData string
	RectData string
}

func (b *Block) Init(m *ice.Message, arg ...string) Chart {
	b.Text = kit.Select(b.Text, arg, 0)
	b.FontColor = kit.Select("white", kit.Select(b.FontColor, arg, 1))
	b.BackGround = kit.Select("red", kit.Select(b.BackGround, arg, 2))
	b.FontSize = kit.Int(kit.Select("24", kit.Select(kit.Format(b.FontSize), arg, 3)))
	b.LineSize = kit.Int(kit.Select("12", kit.Select(kit.Format(b.LineSize), arg, 4)))
	return b
}
func (b *Block) Draw(m *ice.Message, x, y int) Chart {
	m.Echo(`<rect x="%d" y="%d" width="%d" height="%d" fill="%s" %v/>`,
		x+b.Margin/2, y+b.Margin/2, b.GetWidth(), b.GetHeight(), b.BackGround, b.RectData)
	m.Echo("\n")
	m.Echo(`<text x="%d" y="%d" font-size="%d" fill="%s" %v>%v</text>`,
		x+b.GetWidths()/2, y+b.GetHeights()/2, b.FontSize, b.FontColor, b.TextData, b.Text)
	m.Echo("\n")
	return b
}
func (b *Block) Data(root interface{}) {
	kit.Fetch(kit.Value(root, "data"), func(key string, value string) {
		b.TextData += key + "='" + value + "' "
	})
	kit.Fetch(kit.Value(root, "rect"), func(key string, value string) {
		b.RectData += key + "='" + value + "' "
	})
	b.FontColor = kit.Select(b.FontColor, kit.Value(root, "fg"))
	b.BackGround = kit.Select(b.BackGround, kit.Value(root, "bg"))
}
func (b *Block) GetWidth(str ...string) int {
	if b.Width != 0 {
		return b.Width
	}
	s := kit.Select(b.Text, str, 0)
	cn := (len(s) - len([]rune(s))) / 2
	en := len([]rune(s)) - cn
	return cn*b.FontSize + en*b.FontSize*10/16 + b.Padding
}
func (b *Block) GetHeight(str ...string) int {
	if b.Height != 0 {
		return b.Height
	}
	return b.FontSize + b.Padding
}
func (b *Block) GetWidths(str ...string) int {
	return b.GetWidth(str...) + b.Margin
}
func (b *Block) GetHeights(str ...string) int {
	return b.GetHeight() + b.Margin
}

// 框
type Label struct {
	data [][]string
	max  map[int]int
	Block
}

func (b *Label) Init(m *ice.Message, arg ...string) Chart {
	b.FontSize = kit.Int(kit.Select("24", arg, 1))
	b.Padding = kit.Int(kit.Select("16", arg, 6))
	b.Margin = kit.Int(kit.Select("8", arg, 7))

	// 解析数据
	b.max = map[int]int{}
	for _, v := range kit.Split(arg[0], "\n") {
		l := kit.Split(v)
		for i, v := range l {
			switch data := kit.Parse(nil, "", kit.Split(v)...).(type) {
			case map[string]interface{}:
				v = kit.Select("", data["text"])
			}

			if w := b.GetWidth(v); w > b.max[i] {
				b.max[i] = w
			}
		}
		b.data = append(b.data, l)
	}

	// 计算尺寸
	width := 0
	for _, v := range b.max {
		width += v + b.Margin
	}
	b.Width = width
	b.Height = len(b.data) * b.GetHeights()
	return b
}
func (b *Label) Draw(m *ice.Message, x, y int) Chart {
	b.Width, b.Height = 0, 0
	top := y
	for _, line := range b.data {
		left := x
		for i, text := range line {
			switch data := kit.Parse(nil, "", kit.Split(text)...).(type) {
			case map[string]interface{}:
				text = kit.Select(text, data["text"])
			}
			b.Text = text

			width := b.max[i]
			if m.Option("compact") == "true" {
				width = b.GetWidth()
			}

			m.Echo(`<rect x="%d" y="%d" width="%d" height="%d" rx="4" ry="4"/>`,
				left, top, width, b.GetHeight())
			m.Echo("\n")
			m.Echo(`<text x="%d" y="%d" fill="%s" stroke-width="1">%v</text>`,
				left+width/2, top+b.GetHeight()/2, m.Option("stroke"), text)
			m.Echo("\n")
			left += width + b.Margin
		}
		top += b.GetHeights()
	}
	return b
}

// 树
type Chain struct {
	data map[string]interface{}
	max  map[int]int
	Block
}

func (b *Chain) Init(m *ice.Message, arg ...string) Chart {
	// 解数据
	b.data = kit.Parse(nil, "", b.show(m, arg[0])...).(map[string]interface{})
	b.FontColor = kit.Select("white", arg, 1)
	b.BackGround = kit.Select("red", arg, 2)
	b.FontSize = kit.Int(kit.Select("24", arg, 3))
	b.LineSize = kit.Int(kit.Select("12", arg, 4))
	b.Padding = kit.Int(kit.Select("8", arg, 5))
	b.Margin = kit.Int(kit.Select("8", arg, 6))

	// 计算尺寸
	b.max = map[int]int{}
	b.Height = b.size(m, b.data, 0, b.max) * b.GetHeights()
	width := 0
	for _, v := range b.max {
		width += b.GetWidths(strings.Repeat(" ", v))
	}
	b.Width = width
	// m.Log("info", "data %v", kit.Formats(b.data))
	return b
}
func (b *Chain) Draw(m *ice.Message, x, y int) Chart {
	return b.draw(m, b.data, 0, b.max, x, y, &Block{})
}
func (b *Chain) show(m *ice.Message, str string) (res []string) {
	miss := []int{}
	list := kit.Split(str, "\n")
	for _, line := range list {
		// 计算缩进
		dep := 0
	loop:
		for _, v := range []rune(line) {
			switch v {
			case ' ':
				dep++
			case '\t':
				dep += 4
			default:
				break loop
			}
		}

		// 计算层次
		if len(miss) > 0 {
			if miss[len(miss)-1] > dep {
				for i := len(miss) - 1; i >= 0; i-- {
					if miss[i] < dep {
						break
					}
					res = append(res, "]", "}")
					miss = miss[:i]
				}
				miss = append(miss, dep)
			} else if miss[len(miss)-1] < dep {
				miss = append(miss, dep)
			} else {
				res = append(res, "]", "}")
			}
		} else {
			miss = append(miss, dep)
		}

		// 输出节点
		word := kit.Split(line)
		res = append(res, "{", kit.MDB_META, "{", "text")
		res = append(res, word...)
		res = append(res, "}", kit.MDB_LIST, "[")
	}
	return
}
func (b *Chain) size(m *ice.Message, root map[string]interface{}, depth int, width map[int]int) int {
	meta := root[kit.MDB_META].(map[string]interface{})

	// 最大宽度
	text := kit.Format(meta["text"])
	if len(text) > width[depth] {
		width[depth] = len(text)
	}

	// 计算高度
	height := 0
	if list, ok := root[kit.MDB_LIST].([]interface{}); ok && len(list) > 0 {
		kit.Fetch(root[kit.MDB_LIST], func(index int, value map[string]interface{}) {
			height += b.size(m, value, depth+1, width)
		})
	} else {
		height = 1
	}

	meta["height"] = height
	return height
}
func (b *Chain) draw(m *ice.Message, root map[string]interface{}, depth int, width map[int]int, x, y int, p *Block) Chart {
	meta := root[kit.MDB_META].(map[string]interface{})
	b.Width, b.Height = 0, 0

	// 当前节点
	block := &Block{
		BackGround: kit.Select(b.BackGround, kit.Select(p.BackGround, meta["bg"])),
		FontColor:  kit.Select(b.FontColor, kit.Select(p.FontColor, meta["fg"])),
		FontSize:   b.FontSize,
		LineSize:   b.LineSize,
		Padding:    b.Padding,
		Margin:     b.Margin,
		Width:      b.GetWidth(strings.Repeat(" ", width[depth])),
	}

	block.Data(root[kit.MDB_META])
	block.Init(m, kit.Format(meta["text"])).Draw(m, x, y+(kit.Int(meta["height"])-1)*b.GetHeights()/2)

	// 递归节点
	kit.Fetch(root[kit.MDB_LIST], func(index int, value map[string]interface{}) {
		b.draw(m, value, depth+1, width, x+b.GetWidths(strings.Repeat(" ", width[depth])), y, block)
		y += kit.Int(kit.Value(value, "meta.height")) * b.GetHeights()
	})
	return b
}

func Stack(m *ice.Message, name string, level int, data interface{}) {
	style := []string{}
	kit.Fetch(kit.Value(data, "meta"), func(key string, value string) {
		switch key {
		case "bg":
			style = append(style, "background:"+value)
		case "fg":
			style = append(style, "color:"+value)
		}
	})

	l, ok := kit.Value(data, "list").([]interface{})
	if !ok || len(l) == 0 {
		m.Echo(`<div class="%s" style="%s"><span class="state">o</span> %s</div>`, name, strings.Join(style, ";"), kit.Value(data, "meta.text"))
		return
	}
	m.Echo(`<div class="%s %s" style="%s"><span class="state">%s</span> %s</div>`,
		kit.Select("span", "fold", level > 2), name, strings.Join(style, ";"), kit.Select("v", ">", level > 2), kit.Value(data, "meta.text"))

	m.Echo("<ul class='%s' %s>", name, kit.Select("", `style="display:none"`, level > 2))
	kit.Fetch(kit.Value(data, "list"), func(index int, value map[string]interface{}) {
		m.Echo("<li>")
		Stack(m, name, level+1, value)
		m.Echo("</li>")
	})
	m.Echo("</ul>")
}
