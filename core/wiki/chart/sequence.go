package chart

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/core/wiki"
	kit "shylinux.com/x/toolkits"
)

type Sequence struct {
	Head []string
	List [][]map[string]interface{}
	pos  []int
	Block
}

func (s *Sequence) push(m *ice.Message, list string, arg ...interface{}) map[string]interface{} {
	node, node_list := kit.Dict(arg...), kit.Int(list)
	s.List[node_list] = append(s.List[node_list], node)
	_max := kit.Max(len(s.List[node_list])-1, s.pos[node_list])
	node[ORDER], s.pos[node_list] = _max, _max+1
	return node
}
func (s *Sequence) Init(m *ice.Message, arg ...string) wiki.Chart {
	(&s.Block).Init(m)

	// 解析数据
	m.Cmd(lex.SPLIT, "", kit.Dict(nfs.CAT_CONTENT, arg[0]), func(ls []string, data map[string]interface{}) []string {
		if len(s.Head) == 0 {
			s.Head, s.pos = ls, make([]int, len(ls))
			for i := 0; i < len(ls); i++ {
				s.List = append(s.List, []map[string]interface{}{})
			}
			return ls
		}

		from_node := s.push(m, ls[0])
		list := map[string]map[string]interface{}{ls[0]: from_node}
		for i := 1; i < len(ls)-1; i += 2 {
			to_node := list[ls[i+1]]
			if to_node == nil {
				to_node = s.push(m, ls[i+1])
				list[ls[i+1]] = to_node

				_max := kit.Max(kit.Int(from_node[ORDER]), kit.Int(to_node[ORDER]))
				s.pos[kit.Int(ls[i-1])], s.pos[kit.Int(ls[i+1])] = _max+1, _max+1
				from_node[ORDER], to_node[ORDER] = _max, _max
				from_node[mdb.TEXT], from_node[mdb.NEXT] = ls[i], ls[i+1]
			} else {
				from_node[ECHO], from_node[PREV] = ls[i], ls[i+1]
			}
			from_node = to_node
		}
		return ls
	})

	// 计算尺寸
	width := 0
	for _, v := range s.Head {
		width += s.Block.GetWidths(v)
	}
	rect_height := kit.Int(m.Option(RECT + "-" + wiki.HEIGHT))
	s.Width, s.Height = width, kit.Max(s.pos...)*(rect_height+s.MarginY)+s.MarginY+s.GetHeights()
	return s
}
func (s *Sequence) Draw(m *ice.Message, x, y int) wiki.Chart {
	g := wiki.NewGroup(m, ARROW, HEAD, LINE, RECT, NEXT, PREV, TEXT, ECHO)
	arrow_height := kit.Int(g.Option(ARROW, wiki.HEIGHT))
	arrow_width := kit.Int(g.Option(ARROW, wiki.WIDTH))
	rect_height := kit.Int(g.Option(RECT, wiki.HEIGHT))
	rect_width := kit.Int(g.Option(RECT, wiki.WIDTH))
	g.DefsArrow(NEXT, arrow_height, arrow_width)

	height := s.Height
	s.Block.Height, s.Block.Width = 0, 0
	line_pos := make([]int, len(s.List))
	for i := range s.List {
		s.Block.Text = s.Head[i]
		s.Block.Draw(g.Get(HEAD), x, y)
		line_pos[i], x = x+s.Block.GetWidths()/2, x+s.Block.GetWidths()
	}

	y += s.Block.GetHeight() + s.MarginY/2
	for _, x := range line_pos {
		g.EchoLine(LINE, x, y, x, height-s.MarginY/2)
	}

	for i, x := range line_pos {
		for _, v := range s.List[i] {
			pos := kit.Int(v[ORDER])
			g.EchoRect(RECT, rect_height, rect_width, x-rect_width/2, y+pos*(rect_height+s.MarginY)+s.MarginY, "2", "2")

			yy := y + pos*(rect_height+s.MarginY) + s.MarginY + rect_height/4
			if kit.Format(v[mdb.NEXT]) != "" {
				xx := line_pos[kit.Int(v[mdb.NEXT])]
				if x < xx {
					g.EchoArrowLine(NEXT, x+rect_width/2, yy, xx-rect_width/2-arrow_width, yy)
				} else {
					g.EchoArrowLine(NEXT, x-rect_width/2, yy, xx+rect_width/2+arrow_width, yy)
				}
				g.EchoText(TEXT, (x+xx)/2, yy, kit.Format(v[mdb.TEXT]))
			}

			yy += rect_height / 2
			if kit.Format(v[PREV]) != "" {
				xx := line_pos[kit.Int(v[PREV])]
				if x < xx {
					g.EchoArrowLine(PREV, x+rect_width/2, yy, xx-rect_width/2-arrow_width, yy)
				} else {
					g.EchoArrowLine(PREV, x-rect_width/2, yy, xx+rect_width/2+arrow_width, yy)
				}
				g.EchoText(ECHO, (x+xx)/2, yy, kit.Format(v[ECHO]))
			}
		}
	}

	g.Dump(m, HEAD).Dump(m, LINE)
	g.Dump(m, RECT).Dump(m, NEXT).Dump(m, PREV)
	g.Dump(m, TEXT).Dump(m, ECHO)
	return s
}

const (
	ORDER = "order"
)
const (
	ARROW = "arrow"

	HEAD = "head"
	LINE = "line"
	RECT = "rect"
	NEXT = "next"
	PREV = "prev"
	TEXT = "text"
	ECHO = "echo"
)

const SEQUENCE = "sequence"

func init() {
	wiki.AddChart(SEQUENCE, func(m *ice.Message) wiki.Chart {
		m.Option(wiki.MARGINX, "60")
		m.Option(wiki.MARGINY, "20")
		m.Option(wiki.STROKE_WIDTH, "1")
		m.Option(wiki.STROKE, cli.WHITE)
		m.Option(wiki.FILL, cli.WHITE)
		wiki.AddGroupOption(m, ARROW, wiki.HEIGHT, "8", wiki.WIDTH, "18", wiki.FILL, cli.GLASS)
		wiki.AddGroupOption(m, HEAD, wiki.FILL, cli.GLASS)
		wiki.AddGroupOption(m, LINE, wiki.STROKE_DASHARRAY, "20 4 4 4")
		wiki.AddGroupOption(m, RECT, wiki.HEIGHT, "40", wiki.WIDTH, "14")
		wiki.AddGroupOption(m, NEXT, wiki.FILL, cli.GLASS)
		wiki.AddGroupOption(m, PREV, wiki.STROKE_DASHARRAY, "10 2")
		wiki.AddGroupOption(m, TEXT, wiki.FONT_SIZE, "16")
		wiki.AddGroupOption(m, ECHO, wiki.FONT_SIZE, "12")
		return &Sequence{}
	})
}
