package wiki

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	kit "shylinux.com/x/toolkits"

	"strings"
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
	MarginX int
	MarginY int

	Width, Height int

	TextData string
	RectData string

	x, y int
}

func (b *Block) Init(m *ice.Message, arg ...string) Chart {
	b.Text = kit.Select(b.Text, arg, 0)
	b.FontSize = kit.Int(kit.Select(m.Option(FONT_SIZE), kit.Select(kit.Format(b.FontSize), arg, 1)))
	b.FontColor = kit.Select(m.Option(STROKE), kit.Select(b.FontColor, arg, 2))
	b.BackGround = kit.Select(m.Option(FILL), kit.Select(b.BackGround, arg, 3))
	b.Padding = kit.Int(kit.Select(m.Option(PADDING), kit.Select(kit.Format(b.Padding), arg, 4)))
	b.MarginX = kit.Int(kit.Select(m.Option(MARGINX), kit.Select(kit.Format(b.MarginX), arg, 5)))
	b.MarginY = kit.Int(kit.Select(m.Option(MARGINY), kit.Select(kit.Format(b.MarginY), arg, 5)))
	return b
}
func (b *Block) Data(m *ice.Message, root interface{}) Chart {
	b.Text = kit.Select(b.Text, kit.Value(root, kit.MDB_TEXT))
	kit.Fetch(root, func(key string, value string) {
		switch key {
		case FG:
			b.TextData += kit.Format("%s='%s' ", FILL, value)
		case BG:
			b.RectData += kit.Format("%s='%s' ", FILL, value)
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
	float := 0
	if strings.Contains(m.Option(ice.MSG_USERUA), "iPhone") {
		float += 5
	}
	m.Echo(`<rect x="%d" y="%d" width="%d" height="%d" rx="4" ry="4" fill="%s" %v/>`,
		x+b.MarginX/2, y+b.MarginY/2, b.GetWidth(), b.GetHeight(), b.BackGround, b.RectData)
	m.Echo(ice.NL)
	m.Echo(`<text x="%d" y="%d" stroke-width="1" fill="%s" stroke=%s %v>%v</text>`,
		x+b.GetWidths()/2, y+b.GetHeights()/2+float, b.FontColor, b.FontColor, b.TextData, b.Text)
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
	return b.GetWidth(str...) + b.MarginX
}
func (b *Block) GetHeights(str ...string) int {
	return b.GetHeight() + b.MarginY
}

// 框
type Label struct {
	data [][]string
	max  map[int]int
	Block
}

func (b *Label) Init(m *ice.Message, arg ...string) Chart {
	b.Text = kit.Select(b.Text, arg, 0)
	b.FontSize = kit.Int(m.Option(FONT_SIZE))
	b.Padding = kit.Int(m.Option(PADDING))
	b.MarginX = kit.Int(m.Option(MARGINX))
	b.MarginY = kit.Int(m.Option(MARGINY))

	// 解析数据
	b.max = map[int]int{}
	for _, v := range strings.Split(arg[0], ice.NL) {
		l := kit.Split(v, ice.SP, ice.SP)
		for i, v := range l {
			switch data := kit.Parse(nil, "", kit.Split(v)...).(type) {
			case map[string]interface{}:
				v = kit.Select("", data[kit.MDB_TEXT])
			}

			if w := b.GetWidth(v); w > b.max[i] {
				b.max[i] = w
			}
		}
		b.data = append(b.data, l)
	}

	// 计算尺寸
	for _, v := range b.max {
		b.Width += v + b.MarginX
	}
	b.Height = len(b.data) * b.GetHeights()
	return b
}
func (b *Label) Draw(m *ice.Message, x, y int) Chart {
	order, _ := kit.Parse(nil, "", kit.Split(m.Option("order"))...).(map[string]interface{})

	top := y
	for _, line := range b.data {
		left := x
		for i, text := range line {

			// 数据
			item := &Block{
				FontSize: b.FontSize,
				Padding:  b.Padding,
				MarginX:  b.MarginX,
				MarginY:  b.MarginY,
			}
			if order != nil {
				if w := kit.Int(kit.Value(order, "index")); w != 0 && i%w == 0 {
					for k, v := range order {
						switch k {
						case FG:
							item.FontColor = kit.Format(v)
						case BG:
							item.BackGround = kit.Format(v)
						}
					}
				}
			}

			switch data := kit.Parse(nil, "", kit.Split(text)...).(type) {
			case map[string]interface{}:
				item.Init(m, kit.Select(text, data[kit.MDB_TEXT])).Data(m, data)
			default:
				item.Init(m, text)
			}

			// 输出
			switch m.Option(COMPACT) {
			case "max":
				item.Width = b.Width/len(line) - b.MarginX
			case ice.TRUE:

			default:
				item.Width = b.max[i]
			}
			item.Draw(m, left, top)

			left += item.GetWidth() + item.MarginX
			b.Height = item.GetHeight()
		}
		top += b.Height + b.MarginY
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
	b.FontSize = kit.Int(m.Option(FONT_SIZE))
	b.Padding = kit.Int(m.Option(PADDING))
	b.MarginX = kit.Int(m.Option(MARGINX))
	b.MarginY = kit.Int(m.Option(MARGINY))

	// 解析数据
	b.data = kit.Parse(nil, "", b.show(m, arg[0])...).(map[string]interface{})

	// 计算尺寸
	b.max = map[int]int{}
	b.Height = b.size(m, b.data, 0, b.max) * b.GetHeights()
	for _, v := range b.max {
		b.Width += v + b.MarginX
	}
	return b
}
func (b *Chain) Draw(m *ice.Message, x, y int) Chart {
	b.draw(m, b.data, 0, b.max, x, y, &Block{})
	return b
}
func (b *Chain) show(m *ice.Message, str string) (res []string) {
	miss := []int{}
	for _, line := range kit.Split(str, ice.NL, ice.NL) {
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
		word := kit.Split(line, "\t ", "\t ")
		res = append(res, "{", kit.MDB_META, "{", "text")
		res = append(res, word...)
		res = append(res, "}", kit.MDB_LIST, "[")
	}
	return
}
func (b *Chain) size(m *ice.Message, root map[string]interface{}, depth int, width map[int]int) int {
	meta := root[kit.MDB_META].(map[string]interface{})

	// 最大宽度
	if w := b.GetWidths(kit.Format(meta[kit.MDB_TEXT])); w > width[depth] {
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

	meta[HEIGHT] = height
	return height
}
func (b *Chain) draw(m *ice.Message, root map[string]interface{}, depth int, width map[int]int, x, y int, p *Block) int {
	meta := root[kit.MDB_META].(map[string]interface{})
	b.Width, b.Height = 0, 0

	// 当前节点
	item := &Block{
		BackGround: kit.Select(b.BackGround, kit.Select(p.BackGround, meta[BG])),
		FontColor:  kit.Select(b.FontColor, kit.Select(p.FontColor, meta[FG])),
		FontSize:   b.FontSize,
		Padding:    b.Padding,
		MarginX:    b.MarginX,
		MarginY:    b.MarginY,
	}
	if m.Option(COMPACT) != ice.TRUE {
		item.Width = b.max[depth]
	}
	item.x = x
	item.y = y + (kit.Int(meta[HEIGHT])-1)*b.GetHeights()/2
	item.Init(m, kit.Format(meta[kit.MDB_TEXT])).Data(m, meta)
	item.Draw(m, item.x, item.y)

	if p != nil && p.y != 0 {
		x1 := p.x + p.GetWidths() - (p.MarginX+item.MarginX)/4
		y1 := p.y + p.GetHeights()/2
		x4 := item.x + (p.MarginX+item.MarginX)/4
		y4 := item.y + item.GetHeights()/2
		m.Echo(`<path d="M %d,%d Q %d,%d %d,%d T %d %d" stroke=cyan fill=none></path>`,
			x1, y1, x1+(x4-x1)/4, y1, x1+(x4-x1)/2, y1+(y4-y1)/2, x4, y4)
	}

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

func _chart_show(m *ice.Message, kind, text string, arg ...string) {
	var chart Chart
	switch kind {
	case LABEL: // 标签
		chart = &Label{}
	case CHAIN: // 链接
		chart = &Chain{}
	}

	// 扩展参数
	m.Option(FONT_SIZE, "24")
	m.Option(STROKE, cli.BLUE)
	m.Option(FILL, cli.YELLOW)
	// 扩展参数
	m.Option(STYLE, "")
	m.Option(STROKE_WIDTH, "2")
	m.Option(FONT_FAMILY, "monospace")
	// 扩展参数
	m.Option(COMPACT, ice.FALSE)
	m.Option(PADDING, "10")
	m.Option(MARGINX, "10")
	m.Option(MARGINY, "10")
	// m.Option("font-family", kit.Select("", "monospace", len(text) == len([]rune(text))))

	for i := 0; i < len(arg)-1; i++ {
		m.Option(arg[i], arg[i+1])
	}
	if m.Option(BG) != "" {
		m.Option(FILL, m.Option(BG))
	}
	if m.Option(FG) != "" {
		m.Option(STROKE, m.Option(FG))
	}

	// 计算尺寸
	chart.Init(m, text)
	m.Option(WIDTH, chart.GetWidth())
	m.Option(HEIGHT, chart.GetHeight())

	// 渲染引擎
	_wiki_template(m, CHART, "", text)
	defer m.Echo(`</svg>`)
	chart.Draw(m, 0, 0)
}

const (
	FG     = "fg"
	BG     = "bg"
	FILL   = "fill"
	STROKE = "stroke"

	STYLE   = "style"
	WIDTH   = "width"
	HEIGHT  = "height"
	PADDING = "padding"
	MARGINX = "marginx"
	MARGINY = "marginy"

	FONT_SIZE   = "font-size"
	FONT_FAMILY = "font-family"

	STROKE_WIDTH = "stroke-width"

	COMPACT = "compact"

	LABEL = "label"
	CHAIN = "chain"
)
const CHART = "chart"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		CHART: {Name: "chart label|chain text", Help: "图表", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			_chart_show(m, arg[0], strings.TrimSpace(arg[1]), arg[2:]...)
		}},
	}, Configs: map[string]*ice.Config{
		CHART: {Name: CHART, Help: "图表", Value: kit.Data(
			kit.MDB_TEMPLATE, `<svg {{.OptionTemplate}}
vertion="1.1" xmlns="http://www.w3.org/2000/svg" dominant-baseline="middle" text-anchor="middle"
font-size="{{.Option "font-size"}}" stroke="{{.Option "stroke"}}" fill="{{.Option "fill"}}"
stroke-width="{{.Option "stroke-width"}}" font-family="{{.Option "font-family"}}"
width="{{.Option "width"}}" height="{{.Option "height"}}"
>`,
		)},
	}})
}
