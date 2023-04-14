package wiki

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

type Item struct {
	list []string
	args []ice.Any
}

func NewItem(str string, args ...ice.Any) *Item { return &Item{[]string{str}, args} }
func (item *Item) Push(str string, arg ice.Any) *Item {
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
func (item *Item) Echo(str string, arg ...ice.Any) *Item {
	item.list = append(item.list, kit.Format(str, arg...))
	return item
}
func (item *Item) Dump(m *ice.Message) *ice.Message {
	return m.Echo(kit.Join(item.list, ice.SP), item.args...).Echo(ice.NL)
}

type Group struct{ list ice.Messages }

func NewGroup(m *ice.Message, arg ...string) *Group {
	g := &Group{list: ice.Messages{}}
	kit.For(arg, func(k string) { g.list[k] = m.Spawn() })
	return g
}
func AddGroupOption(m *ice.Message, group string, arg ...string) {
	kit.For(arg, func(k, v string) { m.Option(kit.Keys(group, k), v) })
}
func (g *Group) Option(group string, key string, arg ...ice.Any) string {
	return g.Get(group).Option(kit.Keys(group, key), arg...)
}
func (g *Group) Get(group string) *ice.Message { return g.list[group] }

func (g *Group) Echo(group string, str string, arg ...ice.Any) *ice.Message {
	return g.Get(group).Echo(str, arg...)
}
func (g *Group) EchoPath(group string, str string, arg ...ice.Any) *ice.Message {
	return g.Echo(group, `<path d="%s"></path>`, kit.Format(str, arg...))
}
func (g *Group) EchoLine(group string, x1, y1, x2, y2 int) *ice.Message {
	return g.Echo(group, "<line x1=%d y1=%d x2=%d y2=%d></line>", x1, y1, x2, y2)
}
func (g *Group) EchoRect(group string, height, width, x, y int, arg ...string) *ice.Message { // rx ry
	return g.Echo(group, `<rect height=%d width=%d rx=%s ry=%s x=%d y=%d %s/>`,
		height, width, kit.Select("4", arg, 0), kit.Select("4", arg, 1), x, y, kit.JoinKV(ice.EQ, ice.SP, kit.Slice(arg, 2)...))
}
func (g *Group) EchoText(group string, x, y int, text string, arg ...string) *ice.Message {
	float := kit.Int(kit.Select("2", "6", strings.Contains(g.Get(group).Option(ice.MSG_USERUA), "Chrome")))
	return g.Echo(group, "<text x=%d y=%d %s>%s</text>", x, y+float, kit.JoinKV(ice.EQ, ice.SP, arg...), text)
}
func (g *Group) EchoTexts(group string, x, y int, text string, arg ...string) *ice.Message {
	m := g.Get(group)
	float := kit.Int(kit.Select("6", "2", strings.Contains(m.Option(ice.MSG_USERUA), "Chrome") || (strings.Contains(m.Option(ice.MSG_USERUA), "Mobile") && !kit.Contains(m.Option(ice.MSG_USERUA), "iPhone"))))
	return g.EchoText(group, x, y+float, text, arg...)
}
func (g *Group) EchoArrowLine(group string, x1, y1, x2, y2 int, arg ...string) *ice.Message { // marker-end
	return g.Echo(group, "<line x1=%d y1=%d x2=%d y2=%d marker-end='url(#%s)'></line>", x1, y1, x2, y2, kit.Select("arrowhead", arg, 0))
}
func (g *Group) DefsArrow(group string, height, width int, arg ...string) *ice.Message { // name
	return g.Echo(group, `<defs>
<marker id="%s" markerHeight="%d" markerWidth="%d" refX="0" refY="%d" stroke-dasharray="none" orient="auto"><polygon points="0 0, %d %d, 0 %d"/></marker>
</defs>`, kit.Select("arrowhead", arg, 0), height, width, height/2, width, height/2, height)
}
func (g *Group) Dump(m *ice.Message, group string, arg ...string) *Group {
	item := NewItem("<g class=%s", group)
	for _, k := range kit.Simple(STROKE_DASHARRAY, STROKE_WIDTH, STROKE, FILL, FONT_SIZE, FONT_FAMILY, arg) {
		item.Push(kit.Format(`%s="%%v"`, k), m.Option(kit.Keys(group, k)))
	}
	item.Echo(">").Dump(m).Copy(g.Get(group)).Echo("</g>")
	return g
}
func (g *Group) DumpAll(m *ice.Message, group ...string) {
	kit.For(group, func(grp string) { g.Dump(m, grp) })
}

type Chart interface {
	Init(*ice.Message, ...string) Chart
	Draw(*ice.Message, int, int) Chart
	GetHeight(...string) int
	GetWidth(...string) int
}

var chart_list = map[string]func(m *ice.Message) Chart{}

func AddChart(name string, hand func(m *ice.Message) Chart) { chart_list[name] = hand }

func _chart_show(m *ice.Message, name, text string, arg ...string) {
	m.Options(FONT_SIZE, "24", STROKE_WIDTH, "2")
	kit.For(arg, func(k, v string) { m.Option(k, v) })
	m.Option(FILL, kit.Select(m.Option(FILL), m.Option(BG)))
	m.Option(STROKE, kit.Select(m.Option(STROKE), m.Option(FG)))
	chart := chart_list[name](m)
	chart.Init(m, text)
	m.Options(HEIGHT, chart.GetHeight(), WIDTH, chart.GetWidth())
	_wiki_template(m, "", name, text, arg...)
	defer m.Echo("</svg>")
	defer m.RenderResult()
	chart.Draw(m, 0, 0)
}

const (
	FG = "fg"
	BG = "bg"

	FONT_SIZE        = "font-size"
	FONT_FAMILY      = "font-family"
	STROKE_DASHARRAY = "stroke-dasharray"
	STROKE_WIDTH     = "stroke-width"
	STROKE           = "stroke"
	FILL             = "fill"

	PADDING = "padding"
	MARGINX = "marginx"
	MARGINY = "marginy"
	HEIGHT  = "height"
	WIDTH   = "width"
)
const (
	LABEL    = "label"
	CHAIN    = "chain"
	SEQUENCE = "sequence"
)
const CHART = "chart"

func init() {
	Index.MergeCommands(ice.Commands{
		CHART: {Name: "chart type=label,chain,sequence auto text", Help: "图表", Hand: func(m *ice.Message, arg ...string) {
			_chart_show(m, arg[0], strings.TrimSpace(arg[1]), arg[2:]...)
		}},
	})
}
