package chart

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/core/wiki"
	kit "shylinux.com/x/toolkits"
)

type Label struct {
	max  map[int]int
	data [][]string
	Block
}

func (s *Label) Init(m *ice.Message, arg ...string) wiki.Chart {
	(&s.Block).Init(m)
	s.max = map[int]int{}
	m.Cmd(lex.SPLIT, "", kit.Dict(lex.SPLIT_BLOCK, lex.SP, nfs.CAT_CONTENT, arg[0]), func(ls []string) {
		s.data = append(s.data, ls)
		for i, v := range ls {
			if w := s.GetWidth(kit.SplitWord(v)[0]); w > s.max[i] {
				s.max[i] = w
			}
		}
	})
	s.Height = len(s.data) * s.GetHeights()
	for _, v := range s.max {
		s.Width += v + s.MarginX
	}
	return s
}
func (s *Label) Draw(m *ice.Message, x, y int) wiki.Chart {
	gs := wiki.NewGroup(m, RECT, TEXT)
	wiki.AddGroupOption(m, RECT, wiki.STROKE, m.Option(wiki.STROKE), wiki.FILL, m.Option(wiki.FILL))
	wiki.AddGroupOption(m, TEXT, wiki.STROKE, m.Option(wiki.STROKE), wiki.FILL, m.Option(wiki.STROKE))
	defer gs.DumpAll(m, RECT, TEXT)
	top := y
	for _, line := range s.data {
		left, height := x, 0
		for i, text := range line {
			item := s.Fork(m)
			ls := kit.SplitWord(text)
			if item.Init(m, ls[0]); len(ls) > 1 {
				data := kit.Dict()
				for i := 1; i < len(ls)-1; i += 2 {
					kit.Value(data, ls[i], ls[i+1])
				}
				item.Data(m, data)
			}
			switch m.Option(COMPACT) {
			case ice.TRUE:
			case "max":
				item.Width = s.Width/len(line) - s.MarginX
			default:
				item.Width = s.max[i]
			}
			if m.Option(HIDE_BLOCK) != ice.TRUE {
				gs.EchoRect(RECT, item.GetHeight(), item.GetWidth(), left+item.MarginX/2, top+item.MarginY/2)
			}
			gs.EchoTexts(TEXT, left+item.GetWidths()/2, top+item.GetHeights()/2, item.Text)
			if left += item.GetWidths(); item.GetHeights() > height {
				height = item.GetHeights()
			}
		}
		top += height
	}
	return s
}

const (
	HIDE_BLOCK = "hide-block"
	COMPACT    = "compact"
)
const LABEL = "label"

func init() { wiki.AddChart(LABEL, func(m *ice.Message) wiki.Chart { return &Label{} }) }
