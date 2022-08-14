package chat

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const TOPIC = "topic"

func init() {
	form := ice.Map{
		"body.background": ice.Map{
			mdb.TYPE: "text", mdb.NAME: "background", mdb.VALUE: "black",
		},
		"header.height": ice.Map{
			"tags":   "panel.Header,panel.Header>div.output",
			mdb.TYPE: "text", mdb.NAME: "height", mdb.VALUE: "31",
		},
	}
	Index.MergeCommands(ice.Commands{
		TOPIC: {Name: "topic zone id auto create insert", Help: "主题", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(aaa.ROLE, aaa.WHITE, aaa.VOID, TOPIC)
			}},
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case "tags":
					for k := range form {
						m.Push(arg[0], k)
					}
				case "type":
					m.Push(arg[0], "text")
					m.Push(arg[0], "textarea")
					m.Push(arg[0], "select")
				case "name", "value":
					if tags, ok := form[m.Option("tags")]; ok {
						m.Push(arg[0], kit.Format(kit.Value(tags, arg[0])))
					} else if arg[0] == "name" {
						m.Push(arg[0], "background-color")
						m.Push(arg[0], "color")
					}
				default:
					m.Push(arg[0], "red")
					m.Push(arg[0], "blue")
					m.Push(arg[0], "yellow")
					m.Push(arg[0], "green")
					m.Push(arg[0], "blue")
					m.Push(arg[0], "cyan")
					m.Push(arg[0], "magenta")
					m.Push(arg[0], "white")
					m.Push(arg[0], "black")
				}
				m.Cmdy(mdb.INPUTS, m.PrefixKey(), "", mdb.ZONE, arg)
			}},
			"choose": {Name: "choose", Help: "切换", Hand: func(m *ice.Message, arg ...string) {
				m.ProcessLocation(web.MergeURL2(m, "", "topic", kit.TrimExt(m.Option(mdb.ZONE), nfs.CSS)))
			}},
			"create": {Name: "create topic=demo hover=gray float=lightgray color=black background=white", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				buf, err := kit.Render(`
body.{{.Option "topic"}} { background-color:{{.Option "background" }}; color:{{.Option "color" }}; }
body.{{.Option "topic"}} legend { background-color:{{.Option "hover" }}; color:{{.Option "color" }}; }
body.{{.Option "topic"}} select { background-color:{{.Option "background" }}; color:{{.Option "color" }}; }
body.{{.Option "topic"}} textarea { background-color:{{.Option "background" }}; }
body.{{.Option "topic"}} input[type=text] { background-color:{{.Option "background" }}; }
body.{{.Option "topic"}} input[type=button] { background-color:{{.Option "float" }}; color:{{.Option "color" }}; }
legend, select, textarea, input[type=text], div.code, div.story[data-type=spark] { box-shadow:4px 4px 20px 4px {{.Option "float" }}; }

body.{{.Option "topic"}} div.carte { background-color:{{.Option "float" }}; }
body.{{.Option "topic"}} div.input { background-color:{{.Option "float" }}; }
body.{{.Option "topic"}} div.story[data-type=spark] { background:{{.Option "float" }}; }
body.{{.Option "topic"}} fieldset.input { background-color:{{.Option "float" }}; }

body.{{.Option "topic"}} table.content tr { background-color:{{.Option "background" }}; }
body.{{.Option "topic"}} table.content th { background-color:{{.Option "float" }}; }
body.{{.Option "topic"}} table.content.action th:last-child { background-color:{{.Option "float" }}; }
body.{{.Option "topic"}} table.content.action td:last-child { background-color:{{.Option "float" }}; }

body.{{.Option "topic"}} fieldset.panel.Action { background-color:{{.Option "background" }}; }
body.{{.Option "topic"}} fieldset.panel.Footer>div.output div.toast { background-color:{{.Option "float" }}; }
body.{{.Option "topic"}} fieldset.plugin { background-color:{{.Option "background" }}; }
body.{{.Option "topic"}} fieldset.plugin { box-shadow:2px 2px 10px 4px {{.Option "float" }}; }
body.{{.Option "topic"}} fieldset.story { box-shadow:4px 4px 10px 1px {{.Option "float" }}; }
body.{{.Option "topic"}} fieldset.draw div.output { background:{{.Option "background" }}; }
body.{{.Option "topic"}} fieldset.draw div.output div.content svg { background:{{.Option "float" }}; }

body.{{.Option "topic"}} input[type=text]:hover { background-color:{{.Option "float" }}; }
body.{{.Option "topic"}} input[type=button]:hover { background-color:{{.Option "hover" }}; }
body.{{.Option "topic"}} div.item:hover { background-color:{{.Option "float" }}; }
body.{{.Option "topic"}} div.item.select { background-color:{{.Option "float" }}; }
body.{{.Option "topic"}} div.list div.item:hover { background-color:{{.Option "hover" }}; }
body.{{.Option "topic"}} div.list div.item.select { background-color:{{.Option "hover" }}; }
body.{{.Option "topic"}} div.carte div.item:hover { background-color:{{.Option "hover" }}; }
body.{{.Option "topic"}} table.content tr:hover { background-color:{{.Option "float" }}; }
body.{{.Option "topic"}} table.content th:hover { background-color:{{.Option "hover" }}; }
body.{{.Option "topic"}} table.content td:hover { background-color:{{.Option "hover" }}; }
body.{{.Option "topic"}} table.content td.select { background-color:{{.Option "hover" }}; }
body.{{.Option "topic"}} fieldset.plugin:hover { box-shadow:12px 12px 12px 6px {{.Option "float" }}; }
body.{{.Option "topic"}} fieldset.story:hover { box-shadow:12px 12px 12px 6px {{.Option "float" }}; }
body.{{.Option "topic"}} fieldset.panel.Header>div.output div:hover { background-color:{{.Option "float" }}; }
`, m)
				m.Assert(err)
				m.Cmd(nfs.SAVE, path.Join("src/website/topic/"+m.Option("topic")+".css"), string(buf))
			}},
			"form": {Name: "form", Help: "表单", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(m.CommandKey(), m.Option(mdb.ZONE)).Tables(func(value ice.Maps) {
					tags, _ := form[value["tags"]]
					m.Push("tags", value["tags"])
					m.Push("type", kit.Select(kit.Format(kit.Value(tags, "type")), value["type"]))
					m.Push("name", kit.Select(kit.Format(kit.Value(tags, "name")), value["name"]))
					m.Push("value", kit.Select(kit.Format(kit.Value(tags, "value")), value["value"]))
				})
				kit.Fetch(form, func(k string, v ice.Map) {
					m.Push("tags", k)
					m.Push("", v, kit.Split("type,name,value"))
				})
				m.EchoButton("submit")
			}},
			"submit": {Name: "form zone", Help: "提交", Hand: func(m *ice.Message, arg ...string) {
				m.Echo("dome")
			}},
		}, mdb.ZoneAction(mdb.SHORT, "zone", mdb.FIELD, "time,id,tags,type,name,value")), Hand: func(m *ice.Message, arg ...string) {
			if mdb.ZoneSelect(m, arg...); len(arg) == 0 {
				m.Cmd(nfs.DIR, nfs.PWD, ice.OptionFields(), kit.Dict(nfs.DIR_ROOT, "src/website/topic/")).RenameAppend(nfs.PATH, mdb.ZONE, nfs.SIZE, mdb.COUNT).Tables(func(values ice.Maps) {
					m.Push("", values)
				})
				m.PushAction("choose", "form", mdb.REMOVE)
			} else {
				if p := "src/website/topic/" + arg[0]; nfs.ExistsFile(m, p) {
					m.Cmdy(nfs.CAT, p)
				} else {
					m.Tables(func(value ice.Maps) {
						m.Echo("body.%s %s { %s:%s }\n", arg[0], value["tag"], value["name"], value["value"])
					})
				}
			}
		}},
		"/topic/": {Name: "/topic/", Help: "主题", Hand: func(m *ice.Message, arg ...string) {
			if p := "src/website/topic/" + arg[0]; nfs.ExistsFile(m, p) {
				m.RenderDownload(p)
				return
			}
			m.Cmdy(m.CommandKey(), kit.TrimExt(kit.Select("black", arg, 0), nfs.CSS))
			m.W.Header()[web.ContentType] = kit.Simple(web.ContentCSS)
			m.RenderResult()
		}},
	})
}
