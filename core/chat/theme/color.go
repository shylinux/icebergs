package theme

import (
	"strconv"
	"strings"

	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
)

type color struct {
	ice.Hash
	limit  string `data:"300"`
	short  string `data:"type,name,text"`
	field  string `data:"type,name,text,group,order,hash"`
	vendor string `data:"https://www.w3.org/TR/css-color-3/"`
	group  string `name:"group group*"`
	load   string `name:"load file*"`
}

func (s color) Load(m *ice.Message, arg ...string) {
	nfs.ScanCSV(m.Message, m.Option(nfs.FILE), func(data []string) {
		s.Hash.Create(m, mdb.TYPE, m.Option(nfs.FILE), mdb.NAME, data[0], mdb.TEXT, strings.ToUpper(data[1]))
	}, mdb.NAME, mdb.TEXT)
}
func (s color) Matrix(m *ice.Message, arg ...string) {
	m.Cmd("").Table(func(value ice.Maps) {
		key, info := Group(value[mdb.NAME], value[mdb.TEXT])
		m.Push(key, info)
	})
	m.Display("")
}
func (s color) List(m *ice.Message, arg ...string) {
	s.Hash.List(m, arg...).PushAction(s.Group, s.UnGroup, s.Remove).Action(s.Create, s.Load, s.Matrix, s.Vendor, html.FILTER)
	m.Table(func(value ice.Maps) {
		key, info := Group(value[mdb.NAME], value[mdb.TEXT])
		m.Push("color", key).Push("weight", info)
	})
	m.Cut("type,name,text,group,order,color,weight,hash,action")
	m.Sort("type,group,order,text", ice.STR, []string{"red", "yellow", "green", "cyan", "blue", "purple", "brown", "meet", "gray"}, "int", ice.STR).Display("")
}
func (s color) UnGroup(m *ice.Message, arg ...string) {
	s.Hash.Modify(m, mdb.GROUP, "")
}
func (s color) Group(m *ice.Message, arg ...string) {
	s.Hash.Modify(m, mdb.GROUP, m.Option(mdb.GROUP))
}
func init() { ice.ChatCmd(color{}) }

func Group(name, text string) (string, string) {
	if text == "#000000" {
		return "gray", ""
	}
	r, _ := strconv.ParseInt(text[1:3], 16, 32)
	g, _ := strconv.ParseInt(text[3:5], 16, 32)
	b, _ := strconv.ParseInt(text[5:7], 16, 32)
	n := r + g + b
	red := float32(r) / float32(n)
	green := float32(g) / float32(n)
	blue := float32(b) / float32(n)
	name = kit.Format(
		"%s %0.2f %0.2f %0.2f %s",
		text,
		float32(r)/float32(n),
		float32(g)/float32(n),
		float32(b)/float32(n),
		name,
	)
	if red > 0.33 && green > 0.33 && blue > 0.33 {
		return "gray", name
	}
	if red < 0.01 && green > 0.3 && blue > 0.3 {
		return "cyan", name
	} else if red > 0.3 && green < 0.01 && blue > 0.3 {
		return "purple", name
	} else if red > 0.3 && green > 0.3 && blue < 0.01 {
		return "yellow", name
	}
	if red > 0.57 {
		return "red", name
	} else if green > 0.57 {
		return "green", name
	} else if blue > 0.57 {
		return "blue", name
	}
	if red < 0.1 && green > 0.3 && blue > 0.3 {
		return "cyan", name
	} else if red > 0.3 && green < 0.1 && blue > 0.3 {
		return "purple", name
	} else if red > 0.3 && green > 0.3 && blue < 0.1 {
		return "yellow", name
	}
	if red-blue > 0.3 && green-blue > 0.3 {
		return "yellow", name
	} else if red-green > 0.3 && blue-green > 0.3 {
		return "purple", name
	} else if green-red > 0.3 && blue-red > 0.3 {
		return "cyan", name
	}
	if red-green > 0.3 && red-blue > 0.3 {
		return "red", name
	} else if green-red > 0.3 && green-blue > 0.3 {
		return "green", name
	} else if blue-red > 0.3 && blue-green > 0.3 {
		return "blue", name
	}
	if red-blue > 0.2 && green-blue > 0.2 {
		return "yellow", name
	} else if red-green > 0.2 && blue-green > 0.2 {
		return "purple", name
	} else if green-red > 0.2 && blue-red > 0.2 {
		return "cyan", name
	}
	if red > green && red > blue {
		return "red", name
	} else if green > red && green > blue {
		return "green", name
	} else if blue > red && blue > green {
		return "blue", name
	} else {
		return "", name
	}
}
