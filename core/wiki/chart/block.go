package chart

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/core/wiki"
	kit "shylinux.com/x/toolkits"
)

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

func (b *Block) Init(m *ice.Message, arg ...string) wiki.Chart {
	b.FontSize = kit.Int(kit.Select("24", m.Option(wiki.FONT_SIZE)))
	b.Padding = kit.Int(kit.Select("10", m.Option(wiki.PADDING)))
	b.MarginX = kit.Int(kit.Select("10", m.Option(wiki.MARGINX)))
	b.MarginY = kit.Int(kit.Select("10", m.Option(wiki.MARGINY)))

	if len(arg) > 0 {
		b.Text = arg[0]
	}
	return b
}
func (b *Block) Data(m *ice.Message, meta interface{}) wiki.Chart {
	b.Text = kit.Select(b.Text, kit.Value(meta, mdb.TEXT))
	kit.Fetch(meta, func(key string, value string) {
		switch key {
		case wiki.FG:
			b.TextData += kit.Format("%s='%s' ", wiki.FILL, value)
		case wiki.BG:
			b.RectData += kit.Format("%s='%s' ", wiki.FILL, value)
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
func (b *Block) Draw(m *ice.Message, x, y int) wiki.Chart {
	float := kit.Int(kit.Select("2", "7", strings.Contains(m.Option(ice.MSG_USERUA), "iPhone")))
	if m.Option(SHOW_BLOCK) == ice.TRUE {
		item := wiki.NewItem([]string{`<rect height="%d" width="%d" rx="4" ry="4" x="%d" y="%d"`}, b.GetHeight(), b.GetWidth(), x+b.MarginX/2, y+b.MarginY/2)
		item.Push(`fill="%s"`, b.BackGround).Push(`%v`, b.RectData).Echo("/>").Dump(m)
	}
	item := wiki.NewItem([]string{`<text x="%d" y="%d"`}, x+b.GetWidths()/2, y+b.GetHeights()/2+float)
	item.Push(`fill="%s"`, kit.Select(m.Option(wiki.STROKE), b.FontColor))
	item.Push(`stroke-width="%d"`, 1)
	item.Push(`stroke="%s"`, b.FontColor).Push(`fill="%s"`, b.FontColor).Push("%v", b.TextData).Push(`>%v</text>`, b.Text).Dump(m)
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
