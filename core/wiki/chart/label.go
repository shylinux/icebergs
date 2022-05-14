package chart

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
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
	m.Option(lex.SPLIT_BLOCK, ice.SP)
	m.Cmd(lex.SPLIT, "", kit.Dict(nfs.CAT_CONTENT, arg[0]), func(ls []string, data map[string]interface{}) []string {
		l.data = append(l.data, ls)

		for i, v := range ls {
			switch data := kit.Parse(nil, "", kit.Split(v)...).(type) {
			case map[string]interface{}:
				v = kit.Select("", data[mdb.TEXT])
			}
			if w := l.GetWidth(v); w > l.max[i] {
				l.max[i] = w
			}
		}
		return ls
	})

	// 计算尺寸
	l.Height = len(l.data) * l.GetHeights()
	for _, v := range l.max {
		l.Width += v + l.MarginX
	}
	return l
}
func (l *Label) Draw(m *ice.Message, x, y int) wiki.Chart {
	gs := wiki.NewGroup(m, RECT, TEXT)
	wiki.AddGroupOption(m, TEXT, wiki.FILL, m.Option(wiki.STROKE))
	defer func() { gs.Dump(m, RECT).Dump(m, TEXT) }()

	var item *Block
	top := y
	for _, line := range l.data {
		left := x
		for i, text := range line {

			// 数据
			item = &Block{FontSize: l.FontSize, Padding: l.Padding, MarginX: l.MarginX, MarginY: l.MarginY}
			switch data := kit.Parse(nil, "", kit.Split(text)...).(type) {
			case map[string]interface{}:
				item.Init(m, kit.Select(text, data[mdb.TEXT])).Data(m, data)
			default:
				item.Init(m, text)
			}

			// 尺寸
			switch m.Option(COMPACT) {
			case "max":
				item.Width = l.Width/len(line) - l.MarginX
			case ice.TRUE:
			default:
				item.Width = l.max[i]
			}

			// 输出
			if m.Option(SHOW_BLOCK) == ice.TRUE {
				gs.EchoRect(RECT, item.GetHeight(), item.GetWidth(), left+item.MarginX/2, top+item.MarginY/2)
			}
			gs.EchoTexts(TEXT, left+item.GetWidths()/2, top+item.GetHeights()/2, item.Text)

			left += item.GetWidths()
		}
		top += item.GetHeights()
	}
	return l
}

const (
	SHOW_BLOCK = "show-block"
	COMPACT    = "compact"
)
const LABEL = "label"

func init() {
	wiki.AddChart(LABEL, func(m *ice.Message) wiki.Chart {
		m.Option(SHOW_BLOCK, ice.TRUE)
		wiki.AddGroupOption(m, TEXT, wiki.FILL, m.Option(wiki.STROKE))
		wiki.AddGroupOption(m, TEXT, wiki.STROKE_WIDTH, "1")
		wiki.AddGroupOption(m, TEXT, wiki.FONT_FAMILY, m.Option(wiki.FONT_FAMILY))
		return &Label{}
	})
}
