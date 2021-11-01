package wiki

import (
	"encoding/json"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _json_show(m *ice.Message, data interface{}) {
	switch data := data.(type) {
	case map[string]interface{}:
		i := 0
		if m.Echo(`{`); len(data) > 0 {
			m.Echo(`<span class="toggle">...</span>`)
		}
		m.Echo(`<div class="list">`)
		for k, v := range data {
			m.Echo(`<div class="item">`)
			m.Echo(`"<span class="key">%s</span>": `, k)
			_json_show(m, v)
			if i++; i < len(data) {
				m.Echo(",")
			}
			m.Echo("</div>")
		}
		m.Echo(`</div>`)
		m.Echo("}")
	case []interface{}:
		if m.Echo(`[`); len(data) > 0 {
			m.Echo(`<span class="toggle">...</span>`)
		}
		m.Echo(`<div class="list">`)
		for i, v := range data {
			_json_show(m, v)
			if i < len(data)-1 {
				m.Echo(",")
			}
		}
		m.Echo(`</div>`)
		m.Echo("]")
	case string:
		m.Echo(`"<span class="value str">%v</span>"`, data)
	default:
		m.Echo(`<span class="value">%v</span>`, data)
	}
}

const JSON = "json"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		JSON: {Name: JSON, Help: "数据结构", Value: kit.Data(
			kit.MDB_PATH, ice.USR_LOCAL_EXPORT, kit.MDB_REGEXP, ".*\\.json",
		)},
	}, Commands: map[string]*ice.Command{
		JSON: {Name: "json path auto", Help: "数据结构", Meta: kit.Dict(
			ice.Display("/plugin/local/wiki/json.js"),
		), Action: map[string]*ice.Action{
			nfs.SAVE: {Name: "save path text", Help: "保存", Hand: func(m *ice.Message, arg ...string) {
				_wiki_save(m, JSON, arg[0], arg[1])
			}},
			ice.RUN: {Name: "run", Help: "执行", Hand: func(m *ice.Message, arg ...string) {
				var data interface{}
				json.Unmarshal([]byte(m.Cmdx(arg)), &data)
				m.Option("type", "json")
				m.RenderTemplate(`<div {{.OptionTemplate}}>`)
				_json_show(m, data)
				m.Echo(`</div>`)
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if !_wiki_list(m, JSON, kit.Select("./", arg, 0)) {
				m.Cmdy(nfs.CAT, arg[0])
			}
		}},
	}})
}
