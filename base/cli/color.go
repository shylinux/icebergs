package cli

import (
	"fmt"
	"image/color"
	"math/rand"
	"strconv"
	"strings"

	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

const (
	_DARK  = 255
	_LIGHT = 127
)

var _color_map = map[string]color.Color{
	BLACK:  color.RGBA{0, 0, 0, _DARK},
	RED:    color.RGBA{_DARK, 0, 0, _DARK},
	GREEN:  color.RGBA{0, _DARK, 0, _DARK},
	YELLOW: color.RGBA{_DARK, _DARK, 0, _DARK},
	BLUE:   color.RGBA{0, 0, _DARK, _DARK},
	PURPLE: color.RGBA{_DARK, 0, _DARK, _DARK},
	CYAN:   color.RGBA{0, _DARK, _DARK, _DARK},
	WHITE:  color.RGBA{_DARK, _DARK, _DARK, _DARK},
	SILVER: color.RGBA{0xC0, 0xC0, 0xC0, _DARK},
}

func _parse_color(str string) color.Color {
	if str == RANDOM {
		list := kit.SortedKey(_color_map)
		str = list[rand.Intn(len(list))]
	}
	if strings.HasPrefix(str, "#") {
		kit.If(len(str) == 7, func() { str += "ff" })
		if u, e := strconv.ParseUint(str[1:], 16, 64); e == nil {
			return color.RGBA{uint8((u & 0xFF000000) >> 24), uint8((u & 0x00FF0000) >> 16), uint8((u & 0x0000FF00) >> 8), uint8((u & 0x000000FF) >> 0)}
		}
	}
	if color, ok := _color_map[str]; ok {
		return color
	}
	return _color_map[WHITE]
}
func _parse_cli_color(str string) string {
	res := 0
	r, g, b, _ := _parse_color(str).RGBA()
	kit.If(r > _LIGHT, func() { res += 1 })
	kit.If(g > _LIGHT, func() { res += 2 })
	kit.If(b > _LIGHT, func() { res += 4 })
	return kit.Format(res)
}

const (
	BG     = "bg"
	FG     = "fg"
	COLOR  = "color"
	BLACK  = "black"
	WHITE  = "white"
	BLUE   = "blue"
	RED    = "red"
	GRAY   = "gray"
	CYAN   = "cyan"
	GREEN  = "green"
	SILVER = "silver"
	PURPLE = "purple"
	YELLOW = "yellow"
	RANDOM = "random"
	TRANS  = "#0000"
	LIGHT  = "light"
	DARK   = "dark"
)

func Color(m *ice.Message, c string, str ice.Any) string {
	wrap, color := `<span style="color:%s">%v</span>`, c
	kit.If(m.IsCliUA(), func() { wrap, color = "\033[3%sm%v\033[0m", _parse_cli_color(c) })
	return fmt.Sprintf(wrap, color, str)
}
func ColorRed(m *ice.Message, str ice.Any) string    { return Color(m, RED, str) }
func ColorGreen(m *ice.Message, str ice.Any) string  { return Color(m, GREEN, str) }
func ColorYellow(m *ice.Message, str ice.Any) string { return Color(m, YELLOW, str) }
func ParseCliColor(color string) string              { return _parse_cli_color(color) }
func ParseColor(color string) color.Color            { return _parse_color(color) }
