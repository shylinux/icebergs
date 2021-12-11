package wiki

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	kit "shylinux.com/x/toolkits"
)

type Item struct {
	list []string
	args []interface{}
}

func NewItem(list []string, args ...interface{}) *Item {
	return &Item{list, args}
}
func (item *Item) Echo(str string, arg ...interface{}) *Item {
	item.list = append(item.list, kit.Format(str, arg...))
	return item
}
func (item *Item) Push(str string, arg interface{}) *Item {
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
func (item *Item) Dump(m *ice.Message) *ice.Message {
	m.Echo(kit.Join(item.list, ice.SP), item.args...)
	m.Echo(ice.NL)
	return m
}

type Group struct {
	list map[string]*ice.Message
}

func NewGroup(m *ice.Message, arg ...string) *Group {
	g := &Group{list: map[string]*ice.Message{}}
	for _, k := range arg {
		g.list[k] = m.Spawn()
	}
	return g
}
func AddGroupOption(m *ice.Message, group string, arg ...string) {
	for i := 0; i < len(arg)-1; i += 2 {
		m.Option(group+"-"+arg[i], arg[i+1])
	}
}
func (g *Group) Option(group string, key string, arg ...interface{}) string {
	return g.Get(group).Option(group+"-"+key, arg...)
}
func (g *Group) Get(group string) *ice.Message { return g.list[group] }

func (g *Group) DefsArrow(group string, height, width int, arg ...string) *ice.Message { // name
	return g.Echo(group, `<defs>
<marker id="%s" markerHeight="%d" markerWidth="%d" refX="0" refY="%d" orient="auto"><polygon points="0 0, %d %d, 0 %d"/></marker>
</defs>`, kit.Select("arrowhead", arg, 0), height, width, height/2, width, height/2, height)
}
func (g *Group) Echo(group string, str string, arg ...interface{}) *ice.Message {
	return g.Get(group).Echo(str, arg...)
}
func (g *Group) EchoText(group string, x, y int, text string) *ice.Message {
	return g.Echo(group, "<text x=%d y=%d>%s</text>", x, y, text)
}
func (g *Group) EchoRect(group string, height, width, x, y int, arg ...string) *ice.Message { // rx ry
	return g.Echo(group, `<rect height=%d width=%d rx=%s ry=%s x=%d y=%d />`, height, width, kit.Select("4", arg, 0), kit.Select("4", arg, 1), x, y)
}
func (g *Group) EchoLine(group string, x1, y1, x2, y2 int) *ice.Message {
	return g.Echo(group, "<line x1=%d y1=%d x2=%d y2=%d></line>", x1, y1, x2, y2)
}
func (g *Group) EchoArrowLine(group string, x1, y1, x2, y2 int, arg ...string) *ice.Message { // marker-end
	return g.Echo(group, "<line x1=%d y1=%d x2=%d y2=%d marker-end='url(#%s)'></line>", x1, y1, x2, y2, kit.Select("arrowhead", arg, 0))
}
func (g *Group) Dump(m *ice.Message, group string, arg ...string) *Group {
	item := NewItem([]string{"<g name=%s"}, group)
	for _, k := range kit.Simple(STROKE_DASHARRAY, STROKE_WIDTH, STROKE, FILL, FONT_SIZE, arg) {
		item.Push(kit.Format(`%s="%%v"`, k), m.Option(group+"-"+k))
	}
	item.Echo(">").Dump(m).Copy(g.Get(group)).Echo("</g>")
	return g
}

type Chart interface {
	Init(*ice.Message, ...string) Chart
	Data(*ice.Message, interface{}) Chart
	Draw(*ice.Message, int, int) Chart

	GetHeight(...string) int
	GetWidth(...string) int
}

var chart_list = map[string]func(m *ice.Message) Chart{}

func AddChart(name string, hand func(m *ice.Message) Chart) {
	chart_list[name] = hand
}

func _chart_show(m *ice.Message, kind, text string, arg ...string) {
	// 默认参数
	m.Option(STROKE_WIDTH, "2")
	m.Option(STROKE, cli.BLUE)
	m.Option(FILL, cli.YELLOW)
	m.Option(FONT_SIZE, "24")
	m.Option(FONT_FAMILY, "monospace")
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

const (
	FG = "fg"
	BG = "bg"

	STROKE_DASHARRAY = "stroke-dasharray"
	STROKE_WIDTH     = "stroke-width"
	STROKE           = "stroke"
	FILL             = "fill"
	FONT_SIZE        = "font-size"
	FONT_FAMILY      = "font-family"

	PADDING = "padding"
	MARGINX = "marginx"
	MARGINY = "marginy"
	HEIGHT  = "height"
	WIDTH   = "width"
)
const (
	LABEL = "label"
	CHAIN = "chain"
)
const CHART = "chart"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		CHART: {Name: "chart type=label,chain,sequence auto text", Help: "图表", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
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
