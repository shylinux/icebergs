package wiki

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

type item struct {
	list []string
	args []interface{}
}

func newItem(list []string, args ...interface{}) *item {
	return &item{list, args}
}
func (item *item) echo(str string, arg ...interface{}) *item {
	item.list = append(item.list, kit.Format(str, arg...))
	return item
}
func (item *item) push(str string, arg interface{}) *item {
	switch arg := arg.(type) {
	case string:
		if arg == "" {
			return item
		}
	case int:
		if arg == 0 {
			return item
		}
	}
	item.list, item.args = append(item.list, str), append(item.args, arg)
	return item
}
func (item *item) dump(m *ice.Message) *item {
	m.Echo(kit.Join(item.list, ice.SP), item.args...)
	m.Echo(ice.NL)
	return item
}

// 图形接口
type Chart interface {
	Init(*ice.Message, ...string) Chart
	Data(*ice.Message, interface{}) Chart
	Draw(*ice.Message, int, int) Chart

	GetHeight(...string) int
	GetWidth(...string) int
}

// 图形基类
type Block struct {
	Text       string
	FontSize   int
	FontColor  string
	BackGround string

	TextData string
	RectData string

	Padding int
	MarginX int
	MarginY int

	Height int
	Width  int
	x, y   int
}

func (b *Block) Init(m *ice.Message, arg ...string) Chart {
	if len(arg) > 0 {
		b.Text = arg[0]
	}
	if len(arg) > 1 {
		b.FontSize = kit.Int(arg[1])
	}
	if len(arg) > 2 {
		b.FontColor = arg[2]
	}
	if len(arg) > 3 {
		b.BackGround = arg[3]
	}
	if len(arg) > 4 {
		b.Padding = kit.Int(arg[4])
	}
	if len(arg) > 5 {
		b.MarginX = kit.Int(arg[5])
		b.MarginY = kit.Int(arg[5])
	}
	if len(arg) > 6 {
		b.MarginY = kit.Int(arg[6])
	}
	return b
}
func (b *Block) Data(m *ice.Message, meta interface{}) Chart {
	b.Text = kit.Select(b.Text, kit.Value(meta, kit.MDB_TEXT))
	kit.Fetch(meta, func(key string, value string) {
		switch key {
		case FG:
			b.TextData += kit.Format("%s='%s' ", FILL, value)
		case BG:
			b.RectData += kit.Format("%s='%s' ", FILL, value)
		}
	})
	kit.Fetch(kit.Value(meta, "data"), func(key string, value string) {
		b.TextData += kit.Format("%s='%s' ", key, value)
	})
	kit.Fetch(kit.Value(meta, "rect"), func(key string, value string) {
		b.RectData += kit.Format("%s='%s' ", key, value)
	})
	return b
}
func (b *Block) Draw(m *ice.Message, x, y int) Chart {
	float := 0
	if strings.Contains(m.Option(ice.MSG_USERUA), "iPhone") {
		float += 5
	}
	if m.Option(HIDE_BLOCK) != ice.TRUE {
		item := newItem([]string{`<rect height="%d" width="%d" rx="4" ry="4" x="%d" y="%d"`}, b.GetHeight(), b.GetWidth(), x+b.MarginX/2, y+b.MarginY/2)
		item.push(`fill="%s"`, b.BackGround).push(`%v`, b.RectData).echo("/>").dump(m)
	}
	item := newItem([]string{`<text x="%d" y="%d"`}, x+b.GetWidths()/2, y+b.GetHeights()/2+float)
	item.push(`stroke="%s"`, b.FontColor).push(`fill="%s"`, b.FontColor).push("%v", b.TextData).push(`>%v</text>`, b.Text).dump(m)
	return b
}

func (b *Block) GetHeight(str ...string) int {
	if b.Height != 0 {
		return b.Height
	}
	return b.FontSize + b.Padding
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
	b.FontSize = kit.Int(m.Option(FONT_SIZE))
	b.Padding = kit.Int(m.Option(PADDING))
	b.MarginX = kit.Int(m.Option(MARGINX))
	b.MarginY = kit.Int(m.Option(MARGINY))

	// 解析数据
	b.max = map[int]int{}
	for _, v := range strings.Split(arg[0], ice.NL) {
		l := kit.Split(v, ice.SP, ice.SP)
		b.data = append(b.data, l)

		for i, v := range l {
			switch data := kit.Parse(nil, "", kit.Split(v)...).(type) {
			case map[string]interface{}:
				v = kit.Select("", data[kit.MDB_TEXT])
			}
			if w := b.GetWidth(v); w > b.max[i] {
				b.max[i] = w
			}
		}
	}

	// 计算尺寸
	b.Height = len(b.data) * b.GetHeights()
	for _, v := range b.max {
		b.Width += v + b.MarginX
	}
	return b
}
func (b *Label) Draw(m *ice.Message, x, y int) Chart {
	var item *Block
	top := y
	for _, line := range b.data {
		left := x
		for i, text := range line {

			// 数据
			item = &Block{FontSize: b.FontSize, Padding: b.Padding, MarginX: b.MarginX, MarginY: b.MarginY}
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

			left += item.GetWidths()
		}
		top += item.GetHeights()
	}
	return b
}

// 链
type Chain struct {
	data map[string]interface{}
	ship []string
	Block
}

func (b *Chain) Init(m *ice.Message, arg ...string) Chart {
	b.FontSize = kit.Int(m.Option(FONT_SIZE))
	b.Padding = kit.Int(m.Option(PADDING))
	b.MarginX = kit.Int(m.Option(MARGINX))
	b.MarginY = kit.Int(m.Option(MARGINY))

	// 解析数据
	m.Option(nfs.CAT_CONTENT, arg[0])
	m.Option(lex.SPLIT_SPACE, "\t \n")
	m.Option(lex.SPLIT_BLOCK, "\t \n")
	b.data = lex.Split(m, "", kit.MDB_TEXT)

	// 计算尺寸
	b.Height = b.size(m, b.data) * b.GetHeights()
	b.Draw(m, 0, 0)
	b.Width += 200
	m.Set(ice.MSG_RESULT)
	return b
}
func (b *Chain) Draw(m *ice.Message, x, y int) Chart {
	b.draw(m, b.data, x, y, &b.Block)
	m.Echo(`<g stroke=%s stroke-width=%s>`, m.Option(SHIP_STROKE), m.Option(SHIP_STROKE_WIDTH))
	defer m.Echo(`</g>`)
	for _, ship := range b.ship {
		m.Echo(ship)
	}
	return b
}
func (b *Chain) size(m *ice.Message, root map[string]interface{}) (height int) {
	meta := kit.GetMeta(root)
	if list, ok := root[kit.MDB_LIST].([]interface{}); ok && len(list) > 0 {
		kit.Fetch(root[kit.MDB_LIST], func(index int, value map[string]interface{}) {
			height += b.size(m, value)
		})
	} else {
		height = 1
	}
	meta[HEIGHT] = height
	return height
}
func (b *Chain) draw(m *ice.Message, root map[string]interface{}, x, y int, p *Block) int {
	meta := kit.GetMeta(root)
	b.Height, b.Width = 0, 0

	// 当前节点
	item := &Block{
		BackGround: kit.Select(p.BackGround, meta[BG]),
		FontColor:  kit.Select(p.FontColor, meta[FG]),
		FontSize:   p.FontSize,
		Padding:    p.Padding,
		MarginX:    p.MarginX,
		MarginY:    p.MarginY,
	}
	item.x, item.y = x, y+(kit.Int(meta[HEIGHT])-1)*b.GetHeights()/2
	item.Init(m, kit.Format(meta[kit.MDB_TEXT])).Data(m, meta)
	item.Draw(m, item.x, item.y)

	// 画面尺寸
	if item.y+item.GetHeight()+b.MarginY > b.Height {
		b.Height = item.y + item.GetHeight() + b.MarginY
	}
	if item.x+item.GetWidth()+b.MarginX > b.Width {
		b.Width = item.x + item.GetWidth() + b.MarginX
	}

	// 模块连线
	if p != nil && p.y != 0 {
		x1, y1 := p.x+p.GetWidths()-(p.MarginX+item.MarginX)/4, p.y+p.GetHeights()/2
		x4, y4 := item.x+(p.MarginX+item.MarginX)/4, item.y+item.GetHeights()/2
		b.ship = append(b.ship, kit.Format(`<path d="M %d,%d Q %d,%d %d,%d T %d %d" fill=none></path>`,
			x1, y1, x1+(x4-x1)/4, y1, x1+(x4-x1)/2, y1+(y4-y1)/2, x4, y4))
	}

	// 递归节点
	h, x := 0, x+item.GetWidths()
	if kit.Fetch(root[kit.MDB_LIST], func(index int, value map[string]interface{}) {
		h += b.draw(m, value, x, y+h, item)
	}); h == 0 {
		return item.GetHeights()
	}
	return h
}

var chart_list = map[string]func(m *ice.Message) Chart{}

func _chart_show(m *ice.Message, kind, text string, arg ...string) {
	// 画笔参数
	m.Option(STROKE_WIDTH, "2")
	m.Option(STROKE, cli.BLUE)
	m.Option(FILL, cli.YELLOW)
	m.Option(FONT_SIZE, "24")
	m.Option(FONT_FAMILY, "monospace")

	// 几何参数
	m.Option(PADDING, "10")
	m.Option(MARGINX, "10")
	m.Option(MARGINY, "10")

	chart := chart_list[kind](m)

	// 解析参数
	for i := 0; i < len(arg)-1; i++ {
		m.Option(arg[i], arg[i+1])
	}
	m.Option(FILL, kit.Select(m.Option(FILL), m.Option(BG)))
	m.Option(STROKE, kit.Select(m.Option(STROKE), m.Option(FG)))

	// 计算尺寸
	chart.Init(m, text)
	m.Option(WIDTH, chart.GetWidth())
	m.Option(HEIGHT, chart.GetHeight())

	// 渲染引擎
	_wiki_template(m, CHART, "", text)
	defer m.Echo("</svg>")
	chart.Draw(m, 0, 0)
	m.RenderResult()
}
func AddChart(name string, hand func(m *ice.Message) Chart) {
	chart_list[name] = hand
}

const (
	FG = "fg"
	BG = "bg"

	STROKE_WIDTH = "stroke-width"
	STROKE       = "stroke"
	FILL         = "fill"
	FONT_SIZE    = "font-size"
	FONT_FAMILY  = "font-family"

	PADDING = "padding"
	MARGINX = "marginx"
	MARGINY = "marginy"
	HEIGHT  = "height"
	WIDTH   = "width"

	COMPACT           = "compact"
	HIDE_BLOCK        = "hide-block"
	SHIP_STROKE       = "ship-stroke"
	SHIP_STROKE_WIDTH = "ship-stroke-width"
)
const (
	LABEL = "label"
	CHAIN = "chain"
)
const CHART = "chart"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		CHART: {Name: "chart type=label,chain auto text", Help: "图表", Action: map[string]*ice.Action{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				AddChart(LABEL, func(m *ice.Message) Chart {
					return &Label{}
				})
				AddChart(CHAIN, func(m *ice.Message) Chart {
					m.Option(STROKE_WIDTH, "1")
					m.Option(FILL, cli.BLUE)
					m.Option(MARGINX, "40")
					m.Option(MARGINY, "0")

					m.Option(COMPACT, ice.TRUE)
					m.Option(HIDE_BLOCK, ice.TRUE)
					m.Option(SHIP_STROKE, cli.CYAN)
					m.Option(SHIP_STROKE_WIDTH, "1")
					return &Chain{}
				})
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 1 {
				_chart_show(m, arg[0], strings.TrimSpace(arg[1]), arg[2:]...)
			}
		}},
	}, Configs: map[string]*ice.Config{
		CHART: {Name: CHART, Help: "图表", Value: kit.Data(
			kit.MDB_TEMPLATE, `<svg {{.OptionTemplate}}
vertion="1.1" xmlns="http://www.w3.org/2000/svg" height="{{.Option "height"}}" width="{{.Option "width"}}"
stroke-width="{{.Option "stroke-width"}}" stroke="{{.Option "stroke"}}" fill="{{.Option "fill"}}"
font-size="{{.Option "font-size"}}" font-family="{{.Option "font-family"}}" text-anchor="middle" dominant-baseline="middle">`,
		)},
	}})
}
