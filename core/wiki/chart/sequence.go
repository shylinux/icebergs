package chart

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/core/wiki"
	kit "shylinux.com/x/toolkits"
)

type Sequence struct {
	head []string
	list [][]ice.Map
	pos  []int
	max  int
	Block
}

func (s *Sequence) push(m *ice.Message, list string, arg ...ice.Any) ice.Map {
	node, node_list := kit.Dict(arg...), kit.Int(list)
	s.list[node_list] = append(s.list[node_list], node)
	_max := kit.Max(len(s.list[node_list])-1, s.max)
	node[ORDER], s.pos[node_list] = _max, _max+1
	return node
}
func (s *Sequence) Init(m *ice.Message, arg ...string) wiki.Chart {
	(&s.Block).Init(m)
	m.Cmd(lex.SPLIT, "", kit.Dict(lex.SPLIT_BLOCK, lex.SP, nfs.CAT_CONTENT, arg[0]), func(ls []string) {
		if len(s.head) == 0 {
			s.head, s.pos = ls, make([]int, len(ls))
			for i := 0; i < len(ls); i++ {
				s.list = append(s.list, []ice.Map{})
			}
			return
		}

		from_node := s.push(m, ls[0])
		list := map[string]ice.Map{ls[0]: from_node}
		for step, i := 0, 1; i < len(ls)-1; i += 2 {
			to_node := list[ls[i+1]]
			if to_node == nil {
				to_node = s.push(m, ls[i+1])
				list[ls[i+1]] = to_node
				step++

				_max := kit.Max(kit.Int(from_node[ORDER]), kit.Int(to_node[ORDER]), s.max)
				s.pos[kit.Int(ls[i-1])], s.pos[kit.Int(ls[i+1])] = _max+1, _max+1
				from_node[ORDER], to_node[ORDER] = _max, _max

				from_node[TEXT], from_node[NEXT] = kit.Format("%d.%d %s", s.max+1, step, ls[i]), ls[i+1]
				to_node[ECHO], to_node[PREV] = "", ls[i-1]
			} else {
				from_node[ECHO], from_node[PREV] = ls[i], ls[i+1]
			}
			from_node = to_node
		}
		s.max++
		return
	})

	width := 0
	for _, v := range s.head {
		width += s.Block.GetWidths(v)
	}
	rect_height := kit.Int(m.Option(kit.Keys(RECT, wiki.HEIGHT)))
	s.Width, s.Height = width, kit.Max(s.pos...)*(rect_height+s.MarginY)+s.MarginY+s.GetHeights()
	return s
}
func (s *Sequence) Draw(m *ice.Message, x, y int) wiki.Chart {
	gs := wiki.NewGroup(m, HEAD, TITLE, LINE, RECT, NEXT, PREV, TEXT, ECHO, ARROW)
	wiki.AddGroupOption(m, TITLE, wiki.FILL, m.Option(wiki.STROKE))
	wiki.AddGroupOption(m, TEXT, wiki.FILL, m.Option(wiki.STROKE))
	wiki.AddGroupOption(m, ECHO, wiki.FILL, m.Option(wiki.STROKE))
	defer gs.DumpAll(m, HEAD, TITLE, LINE, RECT, NEXT, PREV, TEXT, ECHO)

	rect_width := kit.Int(gs.Option(RECT, wiki.WIDTH))
	rect_height := kit.Int(gs.Option(RECT, wiki.HEIGHT))
	text_size := kit.Int(gs.Option(TEXT, wiki.FONT_SIZE))
	echo_size := kit.Int(gs.Option(ECHO, wiki.FONT_SIZE))
	arrow_height := kit.Int(gs.Option(ARROW, wiki.HEIGHT))
	arrow_width := kit.Int(gs.Option(ARROW, wiki.WIDTH))
	gs.DefsArrow(NEXT, arrow_height, arrow_width, NEXT)
	gs.DefsArrow(PREV, arrow_height, arrow_width, PREV)

	line_pos, height := make([]int, len(s.list)), 0
	for i, head := range s.head {
		item := s.Fork(m, head)
		pos := x + item.GetWidths()/2
		gs.EchoLine(LINE, pos, item.GetHeight()+item.MarginY/2, pos, s.Height-s.MarginY/2)
		gs.EchoRect(HEAD, item.GetHeight(), item.GetWidth(), x+item.MarginX/2, y+item.MarginY/2)
		gs.EchoText(TITLE, pos, y+item.GetHeights()/2, head)
		height = item.GetHeight() + item.MarginY/2
		line_pos[i], x = pos, x+item.GetWidths()
	}
	y += height

	for i, x := range line_pos {
		for _, v := range s.list[i] {
			pos := kit.Int(v[ORDER])
			gs.EchoRect(RECT, rect_height, rect_width, x-rect_width/2, y+pos*(rect_height+s.MarginY)+s.MarginY, "0", "0")

			yy := y + pos*(rect_height+s.MarginY) + s.MarginY + rect_height/4
			if kit.Format(v[NEXT]) != "" {
				xx := line_pos[kit.Int(v[NEXT])]
				if x < xx {
					gs.EchoArrowLine(NEXT, x+rect_width/2, yy, xx-rect_width/2-arrow_width, yy, NEXT)
				} else {
					gs.EchoArrowLine(NEXT, x-rect_width/2, yy, xx+rect_width/2+arrow_width, yy, NEXT)
				}
				gs.EchoText(TEXT, (x+xx)/2, yy-text_size/2, kit.Format(v[TEXT]))
			}

			yy += rect_height / 2
			if kit.Format(v[PREV]) != "" {
				xx := line_pos[kit.Int(v[PREV])]
				if x < xx {
					gs.EchoArrowLine(PREV, x+rect_width/2, yy, xx-rect_width/2-arrow_width, yy, PREV)
				} else {
					gs.EchoArrowLine(PREV, x-rect_width/2, yy, xx+rect_width/2+arrow_width, yy, PREV)
				}
				gs.EchoText(ECHO, (x+xx)/2, yy-echo_size/2, kit.Format(v[ECHO]))
			}
		}
	}
	return s
}

const (
	ORDER = "order"
)
const (
	HEAD  = "head"
	TITLE = "title"
	NEXT  = "next"
	PREV  = "prev"
	ECHO  = "echo"
	ARROW = "arrow"
)

const SEQUENCE = "sequence"

func init() {
	wiki.AddChart(SEQUENCE, func(m *ice.Message) wiki.Chart {
		m.Option(wiki.STROKE_WIDTH, "1")
		m.Option(wiki.MARGINX, "40")
		m.Option(wiki.MARGINY, "40")
		wiki.AddGroupOption(m, LINE, wiki.STROKE_DASHARRAY, "20 4 4 4")
		wiki.AddGroupOption(m, RECT, wiki.HEIGHT, "40", wiki.WIDTH, "14")
		wiki.AddGroupOption(m, PREV, wiki.STROKE_DASHARRAY, "10 2")
		wiki.AddGroupOption(m, TEXT, wiki.FONT_SIZE, "16")
		wiki.AddGroupOption(m, ECHO, wiki.FONT_SIZE, "12")
		wiki.AddGroupOption(m, ARROW, wiki.HEIGHT, "6", wiki.WIDTH, "12", wiki.FILL, cli.TRANS)
		return &Sequence{}
	})
}
