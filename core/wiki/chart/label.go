package chart

import (
	"strings"

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
	m.Cmd(lex.SPLIT, "", kit.Dict(lex.SPLIT_BLOCK, ice.SP, nfs.CAT_CONTENT, arg[0]), func(ls []string) {
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
	wiki.AddGroupOption(m, TEXT, wiki.FILL, m.Option(wiki.STROKE))
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
				args := []string{"4", "4"}
				if mod := kit.Int(m.Option("order.mod")); mod != 0 && i%mod == 0 {
					args = append(args, wiki.FILL, m.Option("order.bg"))
				}
				gs.EchoRect(RECT, item.GetHeight(), item.GetWidth(), left+item.MarginX/2, top+item.MarginY/2, args...)
			}

			args := []string{}
			if mod := kit.Int(m.Option("order.mod")); mod != 0 && i%mod == 0 {
				args = append(args, wiki.STROKE, m.Option("order.fg"))
				args = append(args, wiki.FILL, m.Option("order.fg"))
			}
			if strings.Contains(m.Option(ice.MSG_USERUA), "Chrome") || strings.Contains(m.Option(ice.MSG_USERUA), "Mobile") {
				gs.EchoTexts(TEXT, left+item.GetWidths()/2, top+item.GetHeights()/2, item.Text, args...)
			} else {
				gs.EchoTexts(TEXT, left+item.GetWidths()/2, top+item.GetHeights()/2+4, item.Text, args...)
			}

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

func init() {
	wiki.AddChart(LABEL, func(m *ice.Message) wiki.Chart {
		wiki.AddGroupOption(m, TEXT, wiki.STROKE_WIDTH, "1")
		return &Label{}
	})
}
