package theme

import (
	"strconv"
	"strings"

	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
)

type color struct {
	ice.Hash
	checkbox string `data:"true"`
	export   string `data:"true"`
	limit    string `data:"300"`
	short    string `data:"type,name,text"`
	field    string `data:"type,name,text,help,group,order,hash"`
	trans    string `data:"https://likexia.gitee.io/tools/colors/yansezhongwenming.html"`
	vendor   string `data:"https://www.w3.org/TR/css-color-3/"`
	group    string `name:"group group"`
}

func (s color) Load(m *ice.Message, arg ...string) {
	defer m.ToastProcess()()
	trans := map[string]string{}
	s.goquerys(m, m.Config(nfs.TRANS), "table#color").Table(func(value ice.Maps) {
		trans[strings.ToLower(value["英文代码"])] = value["形像颜色"]
	})
	s.goquerys(m, m.Config(mdb.VENDOR), "table.colortable").Table(func(value ice.Maps) {
		s.Hash.Create(m, mdb.TYPE, "css-color-3", mdb.NAME, value["Color name"], mdb.TEXT, strings.ToUpper(value["Hex rgb"]), mdb.HELP, trans[value["Color name"]])
	})
}
func (s color) Matrix(m *ice.Message, arg ...string) {
	m.Cmd("").Table(func(value ice.Maps) { key, info := Group(value[mdb.NAME], value[mdb.TEXT]); m.Push(key, info) })
	m.Action(s.Matrix).Display("")
}
func (s color) List(m *ice.Message, arg ...string) {
	s.Hash.List(m, arg...).PushAction(s.Group, s.Remove).Action(s.Create, s.Load, s.Vendor, s.Matrix, html.FILTER)
	m.Table(func(value ice.Maps) {
		key, info := Group(value[mdb.NAME], value[mdb.TEXT])
		m.Push(cli.COLOR, key).Push(mdb.WEIGHT, info)
	})
	m.Cut("type,name,text,help,group,order,color,weight,hash,action")
	m.Sort("type,group,order,color,text", ice.STR, group, ice.INT, group, ice.STR)
	m.Display("")
}
func (s color) Group(m *ice.Message, arg ...string) { s.Hash.Modify(m, mdb.GROUP, m.Option(mdb.GROUP)) }

func init() { ice.ChatCmd(color{}) }

var group = []string{cli.RED, cli.YELLOW, cli.GREEN, cli.CYAN, cli.BLUE, cli.PURPLE, "brown", "pink", "meet", cli.GRAY}

func (s color) goquerys(m *ice.Message, path, tags string, arg ...string) *ice.Message {
	return s.goquery(m, s.goquery(m, mdb.CREATE, nfs.PATH, path, nfs.TAGS, tags).Result())
}
func (s color) goquery(m *ice.Message, arg ...string) *ice.Message {
	return m.Cmd(web.SPACE, "20230511-golang-story", "web.code.goquery.goquery", arg)
}

func Group(name, text string) (string, string) {
	if text == "#000000" {
		return cli.GRAY, ""
	}
	r, _ := strconv.ParseInt(text[1:3], 16, 32)
	g, _ := strconv.ParseInt(text[3:5], 16, 32)
	b, _ := strconv.ParseInt(text[5:7], 16, 32)
	n := r + g + b
	red := float32(r) / float32(n)
	green := float32(g) / float32(n)
	blue := float32(b) / float32(n)
	name = kit.Format("%s %0.2f %0.2f %0.2f %s", text,
		float32(r)/float32(n),
		float32(g)/float32(n),
		float32(b)/float32(n),
		name,
	)
	if red > 0.33 && green > 0.33 && blue > 0.33 {
		return cli.GRAY, name
	}
	if red > 0.57 {
		return cli.RED, name
	} else if green > 0.57 {
		return cli.GREEN, name
	} else if blue > 0.57 {
		return cli.BLUE, name
	}
	if red < 0.15 && green > 0.3 && blue > 0.3 {
		return cli.CYAN, name
	} else if red > 0.3 && green < 0.15 && blue > 0.3 {
		return cli.PURPLE, name
	} else if red > 0.3 && green > 0.3 && blue < 0.15 {
		return cli.YELLOW, name
	}
	if red-blue > 0.3 && green-blue > 0.3 {
		return cli.YELLOW, name
	} else if red-green > 0.3 && blue-green > 0.3 {
		return cli.PURPLE, name
	} else if green-red > 0.3 && blue-red > 0.3 {
		return cli.CYAN, name
	}
	if red-green > 0.3 && red-blue > 0.3 {
		return cli.RED, name
	} else if green-red > 0.3 && green-blue > 0.3 {
		return cli.GREEN, name
	} else if blue-red > 0.3 && blue-green > 0.3 {
		return cli.BLUE, name
	}
	if red > green && red > blue {
		return cli.RED, name
	} else if green > red && green > blue {
		return cli.GREEN, name
	} else if blue > red && blue > green {
		return cli.BLUE, name
	}
	if red < green && red < blue {
		return cli.CYAN, name
	} else if green < red && green < blue {
		return cli.PURPLE, name
	} else if blue < red && blue < green {
		return cli.YELLOW, name
	}
	return "", name
}
