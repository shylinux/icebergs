package wiki

import (
	"strings"

	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"
)

// 图形接口
type Chart interface {
	Init(*ice.Message, ...string) Chart
	Data(*ice.Message, interface{}) Chart
	Draw(*ice.Message, int, int) Chart

	GetWidth(...string) int
	GetHeight(...string) int
}

// 图形基类
type Block struct {
	Text       string
	FontSize   int
	FontColor  string
	BackGround string

	Padding int
	Margin  int

	Width, Height int

	TextData string
	RectData string
}

func (b *Block) Init(m *ice.Message, arg ...string) Chart {
	b.Text = kit.Select(b.Text, arg, 0)
	b.FontSize = kit.Int(kit.Select(m.Option("font-size"), kit.Select(kit.Format(b.FontSize), arg, 1)))
	b.FontColor = kit.Select(m.Option("stroke"), kit.Select(b.FontColor, arg, 2))
	b.BackGround = kit.Select(m.Option("fill"), kit.Select(b.BackGround, arg, 3))
	b.Padding = kit.Int(kit.Select(m.Option("padding"), kit.Select(kit.Format(b.Padding), arg, 4)))
	b.Margin = kit.Int(kit.Select(m.Option("margin"), kit.Select(kit.Format(b.Margin), arg, 5)))
	return b
}
func (b *Block) Data(m *ice.Message, root interface{}) Chart {
	b.Text = kit.Select(b.Text, kit.Value(root, "text"))
	kit.Fetch(root, func(key string, value string) {
		switch key {
		case "fg":
			b.TextData += kit.Format("%s='%s' ", "fill", value)
		case "bg":
			b.RectData += kit.Format("%s='%s' ", "fill", value)
		}
	})
	kit.Fetch(kit.Value(root, "data"), func(key string, value string) {
		b.TextData += kit.Format("%s='%s' ", key, value)
	})
	kit.Fetch(kit.Value(root, "rect"), func(key string, value string) {
		b.RectData += kit.Format("%s='%s' ", key, value)
	})
	return b
}
func (b *Block) Draw(m *ice.Message, x, y int) Chart {
	m.Echo(`<rect x="%d" y="%d" width="%d" height="%d" rx="4" ry="4" %v/>`,
		x+b.Margin/2, y+b.Margin/2, b.GetWidth(), b.GetHeight(), b.RectData)
	m.Echo("\n")
	m.Echo(`<text x="%d" y="%d" stroke-width="1" fill="%s" %v>%v</text>`,
		x+b.GetWidths()/2, y+b.GetHeights()/2, b.FontColor, b.TextData, b.Text)
	m.Echo("\n")
	return b
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
	b.Text = kit.Select(b.Text, arg, 0)
	b.FontSize = kit.Int(m.Option("font-size"))
	b.Padding = kit.Int(m.Option("padding"))
	b.Margin = kit.Int(m.Option("margin"))

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
	for _, v := range b.max {
		b.Width += v + b.Margin
	}
	b.Height = len(b.data) * b.GetHeights()
	return b
}
func (b *Label) Draw(m *ice.Message, x, y int) Chart {
	top := y
	for _, line := range b.data {
		left := x
		for i, text := range line {

			// 数据
			item := &Block{
				FontSize: b.FontSize,
				Padding:  b.Padding,
				Margin:   b.Margin,
			}
			switch data := kit.Parse(nil, "", kit.Split(text)...).(type) {
			case map[string]interface{}:
				item.Init(m, kit.Select(text, data["text"])).Data(m, data)
			default:
				item.Init(m, text)
			}

			// 输出
			if m.Option("compact") != "true" {
				item.Width = b.max[i]
			}
			item.Draw(m, left, top)

			left += item.GetWidth() + item.Margin
			b.Height = item.GetHeight()
		}
		top += b.Height + b.Margin
	}
	return b
}

// 链
type Chain struct {
	data map[string]interface{}
	max  map[int]int
	Block
}

func (b *Chain) Init(m *ice.Message, arg ...string) Chart {
	b.FontSize = kit.Int(m.Option("font-size"))
	b.Padding = kit.Int(m.Option("padding"))
	b.Margin = kit.Int(m.Option("margin"))

	// 解析数据
	b.data = kit.Parse(nil, "", b.show(m, arg[0])...).(map[string]interface{})

	// 计算尺寸
	b.max = map[int]int{}
	b.Height = b.size(m, b.data, 0, b.max) * b.GetHeights()
	for _, v := range b.max {
		b.Width += v
	}
	return b
}
func (b *Chain) Draw(m *ice.Message, x, y int) Chart {
	b.draw(m, b.data, 0, b.max, x, y, &Block{})
	return b
}
func (b *Chain) show(m *ice.Message, str string) (res []string) {
	miss := []int{}
	for _, line := range kit.Split(str, "\n") {
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
	if w := b.GetWidths(kit.Format(meta["text"])); w > width[depth] {
		width[depth] = w
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
func (b *Chain) draw(m *ice.Message, root map[string]interface{}, depth int, width map[int]int, x, y int, p *Block) int {
	meta := root[kit.MDB_META].(map[string]interface{})
	b.Width, b.Height = 0, 0

	// 当前节点
	item := &Block{
		BackGround: kit.Select(b.BackGround, kit.Select(p.BackGround)),
		FontColor:  kit.Select(b.FontColor, kit.Select(p.FontColor)),
		FontSize:   b.FontSize,
		Padding:    b.Padding,
		Margin:     b.Margin,
	}
	if m.Option("compact") != "true" {
		item.Width = b.max[depth]
	}
	item.Init(m, kit.Format(meta["text"])).Data(m, meta)
	item.Draw(m, x, y+(kit.Int(meta["height"])-1)*b.GetHeights()/2)

	// 递归节点
	h := 0
	x += item.GetWidths()
	kit.Fetch(root[kit.MDB_LIST], func(index int, value map[string]interface{}) {
		h += b.draw(m, value, depth+1, width, x, y+h, item)
	})
	if h == 0 {
		return item.GetHeights()
	}
	return h
}

// 栈
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
