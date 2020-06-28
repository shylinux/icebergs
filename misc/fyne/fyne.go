package fyne

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/core/chat"
	"github.com/shylinux/toolkits"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/widget"
	"os"
	"strings"
)

var Index = &ice.Context{Name: "fyne", Help: "fyne",
	Configs: map[string]*ice.Config{},
	Commands: map[string]*ice.Command{
		"field": {Name: "field", Help: "field", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			if len(arg) == 0 {
				arg = append(arg, "space")
			}
			newField(m, kit.Select("contexts", m.Option("title"))).update(m.Cmd(arg))
		}},
		"hide": {Name: "hide", Help: "hide", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			field := m.Optionv("field").(*Field)
			field.w.Hide()
		}},
		"close": {Name: "close", Help: "close", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			field := m.Optionv("field").(*Field)
			field.w.Close()
		}},
	},
}

type Label struct {
	*widget.Label
	width  int
	height int
}

func newLabel(str string, width int, height int) *Label {
	return &Label{width: width, height: height, Label: widget.NewLabel(str)}
}
func (label *Label) MinSize() fyne.Size {
	return fyne.NewSize(label.width*10, label.height*20)
}

type Board struct {
	width  int
	height int
	*widget.ScrollContainer
}

func newBoard(list fyne.CanvasObject, width int, height int) *Board {
	return &Board{width: width, height: height, ScrollContainer: widget.NewScrollContainer(list)}
}
func (board *Board) MinSize() fyne.Size {
	return fyne.NewSize(board.width*10, board.height*20)
}

type Field struct {
	widget.Entry
	w fyne.Window
	m *ice.Message
}

func newField(m *ice.Message, title string) *Field {
	w := win.NewWindow(title)
	w.CenterOnScreen()
	w.SetMainMenu(fyne.NewMainMenu(
		fyne.NewMenu("action",
			fyne.NewMenuItem("contexts", func() { win.OpenURL(kit.ParseURL("http://localhost:9020")) }),
			fyne.NewMenuItem("fnye", func() { win.OpenURL(kit.ParseURL("https://developer.fyne.io")) }),
			fyne.NewMenuItem("quit", func() { os.Exit(0) }),
		),
	))
	return &Field{w: w}
}
func (field *Field) update(m *ice.Message) {
	field.ExtendBaseWidget(field)
	field.m = m

	cols := 0
	rows := 0
	list := []fyne.CanvasObject{}
	width := map[int]int{}
	m.Table(func(index int, value map[string]string, head []string) {
		if index == 0 {
			for i, k := range head {
				if len(k) > width[i] {
					width[i] = len(k)
				}
			}
		}
		for i, k := range head {
			if len(value[k]) > width[i] {
				width[i] = len(value[k])
			}
		}
	})
	m.Table(func(index int, value map[string]string, head []string) {
		rows = index + 1
		if cols = len(head); index == 0 {
			line := []fyne.CanvasObject{}
			for i, k := range head {
				item := newLabel(k, width[i], 1)
				line = append(line, item)
			}
			list = append(list, widget.NewHBox(line...))
		}

		line := []fyne.CanvasObject{}
		for i, k := range head {
			v := value[k]
			if len(v) > 40 {
				v = v[:40] + "..."
				width[i] = 40
			}
			item := newLabel(v, width[i], 1)
			line = append(line, item)
		}
		list = append(list, widget.NewHBox(line...))
	})
	table := widget.NewVBox(list...)
	count := strings.Count(m.Result(), "\n")
	// board := newBoard(newLabel(m.Result(), 20, count), 20, lines)
	board := widget.NewScrollContainer(newLabel(m.Result(), 20, count))

	w := field.w
	w.Resize(fyne.NewSize(kit.Int(kit.Select("600", m.Option("width"))), kit.Int(kit.Select("200", m.Option("height")))))
	w.SetContent(widget.NewVBox(field, table, board, widget.NewHBox(
		newLabel(m.Time(), 20, 1),
	)))
	w.Show()
}
func (field *Field) KeyDown(key *fyne.KeyEvent) {
	switch m := field.m; key.Name {
	case fyne.KeyReturn:
		m.Optionv("field", field)
		field.update(m.Cmd(kit.Split(field.Text)))
		field.Entry.SetText("")
	default:
	}
}

var win = app.New()

func init() { ice.Loop = win.Run }

func init() { chat.Index.Register(Index, nil) }
