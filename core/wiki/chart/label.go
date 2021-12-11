package chart

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/core/wiki"
	kit "shylinux.com/x/toolkits"
)

type Label struct {
	data [][]string
	max  map[int]int
	Block
}

func (l *Label) Init(m *ice.Message, arg ...string) wiki.Chart {
	(&l.Block).Init(m)

	// 解析数据
	l.max = map[int]int{}
	for _, v := range strings.Split(arg[0], ice.NL) {
		ls := kit.Split(v, ice.SP, ice.SP)
		l.data = append(l.data, ls)

		for i, v := range ls {
			switch data := kit.Parse(nil, "", kit.Split(v)...).(type) {
			case map[string]interface{}:
				v = kit.Select("", data[kit.MDB_TEXT])
			}
			if w := l.GetWidth(v); w > l.max[i] {
				l.max[i] = w
			}
		}
	}

	// 计算尺寸
	l.Height = len(l.data) * l.GetHeights()
	for _, v := range l.max {
		l.Width += v + l.MarginX
	}
	return l
}
func (l *Label) Draw(m *ice.Message, x, y int) wiki.Chart {
	var item *Block
	top := y
	for _, line := range l.data {
		left := x
		for i, text := range line {

			// 数据
			item = &Block{FontSize: l.FontSize, Padding: l.Padding, MarginX: l.MarginX, MarginY: l.MarginY}
			switch data := kit.Parse(nil, "", kit.Split(text)...).(type) {
			case map[string]interface{}:
				item.Init(m, kit.Select(text, data[kit.MDB_TEXT])).Data(m, data)
			default:
				item.Init(m, text)
			}

			// 输出
			switch m.Option(COMPACT) {
			case "max":
				item.Width = l.Width/len(line) - l.MarginX
			case ice.TRUE:

			default:
				item.Width = l.max[i]
			}
			item.Draw(m, left, top)

			left += item.GetWidths()
		}
		top += item.GetHeights()
	}
	return l
}

const (
	COMPACT = "compact"
)
const LABEL = "label"

func init() {
	wiki.AddChart(LABEL, func(m *ice.Message) wiki.Chart {
		return &Label{}
	})
}
