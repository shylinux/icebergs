package web

import (
	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"

	"os"
	"path"
	"strings"
	"time"
)

func _story_share(m *ice.Message, story string, list string, arg ...string) {
	if m.Echo("share: "); list == "" {
		msg := m.Cmd(ice.WEB_STORY, ice.STORY_INDEX, story)
		m.Cmdy(ice.WEB_SHARE, "add", "story", story, msg.Append("list"))
	} else {
		msg := m.Cmd(ice.WEB_STORY, ice.STORY_INDEX, list)
		m.Cmdy(ice.WEB_SHARE, "add", msg.Append("scene"), msg.Append("story"), msg.Append("text"))
	}
}
func _story_list(m *ice.Message, arg ...string) {
	// 故事列表
	m.Richs(ice.WEB_STORY, "head", "*", func(key string, value map[string]interface{}) {
		m.Push(key, value, []string{"time", "story", "count"})
	})
	m.Sort("time", "time_r")
}
func _story_show(m *ice.Message, arg ...string) {
	if len(arg) == 1 {
		// 故事记录
		m.Cmdy(ice.WEB_STORY, "history", arg)
		return
	}
	// 故事详情
	m.Cmd(ice.WEB_STORY, ice.STORY_INDEX, arg[1]).Table(func(index int, value map[string]string, head []string) {
		for k, v := range value {
			m.Push("key", k)
			m.Push("value", v)
		}
		m.Sort("key")
	})
}

func _story_pull(m *ice.Message, arg ...string) {
	// 起止节点
	prev, begin, end := "", arg[3], ""
	repos := kit.Keys("remote", arg[2], arg[3])
	m.Richs(ice.WEB_STORY, "head", arg[1], func(key string, val map[string]interface{}) {
		end = kit.Format(kit.Value(val, kit.Keys(repos, "pull", "list")))
		prev = kit.Format(val["list"])
	})

	pull := end
	var first map[string]interface{}
	for begin != "" && begin != end {
		if m.Cmd(ice.WEB_SPIDE, arg[2], "msg", "/story/pull", "begin", begin, "end", end).Table(func(index int, value map[string]string, head []string) {
			if m.Richs(ice.WEB_CACHE, nil, value["data"], nil) == nil {
				m.Log(ice.LOG_IMPORT, "%v: %v", value["data"], value["save"])
				if node := kit.UnMarshal(value["save"]); kit.Format(kit.Value(node, "file")) != "" {
					// 下载文件
					m.Cmd(ice.WEB_SPIDE, arg[2], "cache", "GET", "/story/download/"+value["data"])
				} else {
					// 导入缓存
					m.Conf(ice.WEB_CACHE, kit.Keys("hash", value["data"]), node)
				}
			}

			node := kit.UnMarshal(value["node"]).(map[string]interface{})
			if m.Richs(ice.WEB_STORY, nil, value["list"], nil) == nil {
				// 导入节点
				m.Log(ice.LOG_IMPORT, "%v: %v", value["list"], value["node"])
				m.Conf(ice.WEB_STORY, kit.Keys("hash", value["list"]), node)
			}

			if first == nil {
				if m.Richs(ice.WEB_STORY, "head", arg[1], nil) == nil {
					// 自动创建
					h := m.Rich(ice.WEB_STORY, "head", kit.Dict(
						"scene", node["scene"], "story", arg[1],
						"count", node["count"], "list", value["list"],
					))
					m.Log(ice.LOG_CREATE, "%v: %v", h, node["story"])
				}

				pull, first = kit.Format(value["list"]), node
				m.Richs(ice.WEB_STORY, "head", arg[1], func(key string, val map[string]interface{}) {
					prev = kit.Format(val["list"])
					if kit.Int(node["count"]) > kit.Int(kit.Value(val, kit.Keys(repos, "pull", "count"))) {
						// 更新分支
						m.Log(ice.LOG_IMPORT, "%v: %v", arg[2], pull)
						kit.Value(val, kit.Keys(repos, "pull"), kit.Dict(
							"head", arg[3], "time", node["time"], "list", pull, "count", node["count"],
						))
					}
				})
			}

			if prev == kit.Format(node["prev"]) || prev == kit.Format(node["push"]) {
				// 快速合并
				m.Log(ice.LOG_IMPORT, "%v: %v", pull, arg[2])
				m.Richs(ice.WEB_STORY, "head", arg[1], func(key string, val map[string]interface{}) {
					val["count"] = first["count"]
					val["time"] = first["time"]
					val["list"] = pull
				})
				prev = pull
			}

			begin = kit.Format(node["prev"])
		}).Appendv("list") == nil {
			break
		}
	}

}
func _story_push(m *ice.Message, arg ...string) {
	// 更新分支
	m.Cmdx(ice.WEB_STORY, "pull", arg[1:])

	repos := kit.Keys("remote", arg[2], arg[3])
	// 查询索引
	prev, pull, some, list := "", "", "", ""
	m.Richs(ice.WEB_STORY, "head", arg[1], func(key string, val map[string]interface{}) {
		prev = kit.Format(val["list"])
		pull = kit.Format(kit.Value(val, kit.Keys(repos, "pull", "list")))
		for some = pull; prev != some && some != ""; {
			local := m.Richs(ice.WEB_STORY, nil, prev, nil)
			remote := m.Richs(ice.WEB_STORY, nil, some, nil)
			if diff := kit.Time(kit.Format(remote["time"])) - kit.Time(kit.Format(local["time"])); diff > 0 {
				some = kit.Format(remote["prev"])
			} else if diff < 0 {
				prev = kit.Format(local["prev"])
			}
		}

		if prev = kit.Format(val["list"]); prev == pull {
			// 相同节点
			return
		}

		if some != pull {
			// 合并节点
			local := m.Richs(ice.WEB_STORY, nil, prev, nil)
			remote := m.Richs(ice.WEB_STORY, nil, pull, nil)
			list = m.Rich(ice.WEB_STORY, nil, kit.Dict(
				"scene", val["scene"], "story", val["story"], "count", kit.Int(remote["count"])+1,
				"data", local["data"], "prev", pull, "push", prev,
			))
			m.Log(ice.LOG_CREATE, "merge: %s %s->%s", list, prev, pull)
			val["list"] = list
			prev = list
			val["count"] = kit.Int(remote["count"]) + 1
		}

		// 查询节点
		nodes := []string{}
		for list = prev; list != some; {
			m.Richs(ice.WEB_STORY, nil, list, func(key string, value map[string]interface{}) {
				nodes, list = append(nodes, list), kit.Format(value["prev"])
			})
		}

		for _, v := range kit.Revert(nodes) {
			m.Richs(ice.WEB_STORY, nil, v, func(list string, node map[string]interface{}) {
				m.Richs(ice.WEB_CACHE, nil, node["data"], func(data string, save map[string]interface{}) {
					if kit.Format(save["file"]) != "" {
						// 推送缓存
						m.Cmd(ice.WEB_SPIDE, arg[2], "/story/upload",
							"part", "upload", "@"+kit.Format(save["file"]),
						)
					}

					// 推送节点
					m.Log(ice.LOG_EXPORT, "%s: %s", v, kit.Format(node))
					m.Cmd(ice.WEB_SPIDE, arg[2], "/story/push",
						"story", arg[3], "list", v, "node", kit.Format(node),
						"data", node["data"], "save", kit.Format(save),
					)
				})
			})
		}
	})

	// 更新分支
	m.Cmd(ice.WEB_STORY, "pull", arg[1:])

}
func _story_commit(m *ice.Message, arg ...string) {
	// 查询索引
	head, prev, value, count := "", "", map[string]interface{}{}, 0
	m.Richs(ice.WEB_STORY, "head", arg[1], func(key string, val map[string]interface{}) {
		head, prev, value, count = key, kit.Format(val["list"]), val, kit.Int(val["count"])
		m.Log("info", "head: %v prev: %v count: %v", head, prev, count)
	})

	// 提交信息
	arg[2] = m.Cmdx(ice.WEB_STORY, "add", "submit", arg[2], "hostname,username")

	// 节点信息
	menu := map[string]string{}
	for i := 3; i < len(arg); i++ {
		menu[arg[i]] = m.Cmdx(ice.WEB_STORY, ice.STORY_INDEX, arg[i])
	}

	// 添加节点
	list := m.Rich(ice.WEB_STORY, nil, kit.Dict(
		"scene", "commit", "story", arg[1], "count", count+1, "data", arg[2], "list", menu, "prev", prev,
	))
	m.Log(ice.LOG_CREATE, "commit: %s %s: %s", list, arg[1], arg[2])
	m.Push("list", list)

	if head == "" {
		// 添加索引
		m.Rich(ice.WEB_STORY, "head", kit.Dict("scene", "commit", "story", arg[1], "count", count+1, "list", list))
	} else {
		// 更新索引
		value["count"] = count + 1
		value["time"] = m.Time()
		value["list"] = list
	}
	m.Echo(list)

}

func _story_trash(m *ice.Message, arg ...string) {
	bak := kit.Select(kit.Keys(arg[1], "bak"), arg, 2)
	os.Remove(bak)
	os.Rename(arg[1], bak)
}
func _story_watch(m *ice.Message, index string, arg ...string) {
	// 备份文件
	name := kit.Select(index, arg, 0)
	m.Cmd(ice.WEB_STORY, ice.STORY_TRASH, name)

	if msg := m.Cmd(ice.WEB_STORY, ice.STORY_INDEX, index); msg.Append("file") != "" {
		p := path.Dir(name)
		os.MkdirAll(p, 0777)

		// 导出文件
		os.Link(msg.Append("file"), name)
		m.Log(ice.LOG_EXPORT, "%s: %s", msg.Append("file"), name)
	} else {
		if f, p, e := kit.Create(name); m.Assert(e) {
			defer f.Close()
			// 导出数据
			f.WriteString(msg.Append("text"))
			m.Log(ice.LOG_EXPORT, "%s: %s", msg.Append("text"), p)
		}
	}
	m.Echo(name)
}
func _story_catch(m *ice.Message, arg ...string) {
	if last := m.Richs(ice.WEB_STORY, "head", arg[2], nil); last != nil {
		if t, e := time.ParseInLocation(ice.ICE_TIME, kit.Format(last["time"]), time.Local); e == nil {
			// 文件对比
			if s, e := os.Stat(arg[2]); e == nil && s.ModTime().Before(t) {
				m.Info("%s last: %s", arg[2], kit.Format(t))
				m.Echo("%s", last["list"])
				return
			}
		}
	}
	_story_add(m, arg...)
}
func _story_add(m *ice.Message, arg ...string) {
	if m.Richs(ice.WEB_CACHE, nil, kit.Select("", arg, 3), func(key string, value map[string]interface{}) {
		// 复用缓存
		arg[3] = key
	}) == nil {
		// 添加缓存
		m.Cmdy(ice.WEB_CACHE, arg)
		arg = []string{arg[0], m.Append("type"), m.Append("name"), m.Append("data")}
	}

	// 查询索引
	head, prev, value, count := "", "", map[string]interface{}{}, 0
	m.Richs(ice.WEB_STORY, "head", arg[2], func(key string, val map[string]interface{}) {
		head, prev, value, count = key, kit.Format(val["list"]), val, kit.Int(val["count"])
		m.Log("info", "head: %v prev: %v count: %v", head, prev, count)
	})

	if last := m.Richs(ice.WEB_STORY, nil, prev, nil); prev != "" && last != nil && last["data"] == arg[3] {
		// 重复提交
		m.Echo(prev)
	} else {
		// 添加节点
		list := m.Rich(ice.WEB_STORY, nil, kit.Dict(
			"scene", arg[1], "story", arg[2], "count", count+1, "data", arg[3], "prev", prev,
		))
		m.Log(ice.LOG_CREATE, "story: %s %s: %s", list, arg[1], arg[2])
		m.Push("list", list)

		if head == "" {
			// 添加索引
			m.Rich(ice.WEB_STORY, "head", kit.Dict("scene", arg[1], "story", arg[2], "count", count+1, "list", list))
		} else {
			// 更新索引
			value["count"] = count + 1
			value["time"] = m.Time()
			value["list"] = list
		}
		m.Echo(list)
	}

	// 分发数据
	if p := kit.Select(m.Conf(ice.WEB_FAVOR, "meta.proxy"), m.Option("you")); p != "" {
		m.Option("you", "")
		m.Cmd(ice.WEB_PROXY, p, ice.WEB_STORY, ice.STORY_PULL, arg[2], "dev", arg[2])
	}

}

func _story_index(m *ice.Message, name string) {
	m.Richs(ice.WEB_STORY, "head", name, func(key string, value map[string]interface{}) {
		// 查询索引
		name = kit.Format(value["list"])
	})

	m.Richs(ice.WEB_STORY, nil, name, func(key string, value map[string]interface{}) {
		// 查询节点
		m.Push("list", key)
		m.Push(key, value, []string{"scene", "story"})
		name = kit.Format(value["data"])
	})

	m.Richs(ice.WEB_CACHE, nil, name, func(key string, value map[string]interface{}) {
		// 查询数据
		m.Push("data", key)
		m.Push(key, value, []string{"text", "time", "size", "type", "name", "file"})
		if value["file"] != "" {
			m.Echo("%s", m.Cmdx("nfs.cat", value["file"]))
		}
	})
}
func _story_history(m *ice.Message, name string) {
	// 历史记录
	list := m.Cmd(ice.WEB_STORY, ice.STORY_INDEX, name).Append("list")
	for i := 0; i < kit.Int(kit.Select("30", m.Option("cache.limit"))) && list != ""; i++ {

		m.Richs(ice.WEB_STORY, nil, list, func(key string, value map[string]interface{}) {
			// 直连节点
			m.Push(key, value, []string{"time", "key", "count", "scene", "story"})
			m.Richs(ice.WEB_CACHE, nil, value["data"], func(key string, value map[string]interface{}) {
				m.Push("drama", value["text"])
				m.Push("data", key)
			})

			kit.Fetch(value["list"], func(key string, val string) {
				m.Richs(ice.WEB_STORY, nil, val, func(key string, value map[string]interface{}) {
					// 复合节点
					m.Push(key, value, []string{"time", "key", "count", "scene", "story"})
					m.Richs(ice.WEB_CACHE, nil, value["data"], func(key string, value map[string]interface{}) {
						m.Push("drama", value["text"])
						m.Push("data", key)
					})
				})
			})

			// 切换节点
			list = kit.Format(value["prev"])
		})
	}
}

func StoryIndex(m *ice.Message, name string)               { _story_index(m, name) }
func StoryHistory(m *ice.Message, name string)             { _story_history(m, name) }
func StoryWatch(m *ice.Message, index string, file string) { _story_watch(m, index, file) }
func StoryCatch(m *ice.Message, mime string, file string) {
	_story_catch(m, "catch", kit.Select(mime, strings.TrimPrefix(path.Ext(file), ".")), file)
}

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			ice.WEB_STORY: {Name: "story", Help: "故事会", Value: kit.Dict(
				kit.MDB_META, kit.Dict(kit.MDB_SHORT, "data"),
				"head", kit.Data(kit.MDB_SHORT, "story"),
				"mime", kit.Dict("md", "txt"),
			)},
		},
		Commands: map[string]*ice.Command{
			ice.WEB_STORY: {Name: "story story=auto key=auto auto", Help: "故事会", Meta: kit.Dict(
				"exports", []string{"top", "story"}, "detail", []string{"共享", "更新", "推送"},
			), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) > 1 && arg[0] == "action" {
					story, list := m.Option("story"), m.Option("list")
					switch arg[2] {
					case "story":
						story = arg[3]
					case "list":
						list = arg[3]
					}

					switch arg[1] {
					case "share", "共享":
						_story_share(m, story, list, arg...)
					}
					return
				}

				if len(arg) == 0 {
					_story_list(m, arg...)
					return
				}

				switch arg[0] {
				case ice.STORY_PULL: // story [spide [story]]
					_story_pull(m, arg...)
				case ice.STORY_PUSH:
					_story_push(m, arg...)

				case "commit":
					_story_commit(m, arg...)

				case ice.STORY_TRASH:
					_story_trash(m, arg...)
				case ice.STORY_WATCH:
					_story_watch(m, arg[1], arg[2:]...)
				case ice.STORY_CATCH:
					_story_catch(m, arg...)
				case "add", ice.STORY_UPLOAD, ice.STORY_DOWNLOAD:
					_story_add(m, arg...)

				case ice.STORY_INDEX:
					_story_index(m, arg[1])
				case ice.STORY_HISTORY:
					_story_history(m, arg[1])
				default:
					_story_show(m, arg...)
				}
			}},
			"/story/": {Name: "/story/", Help: "故事会", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {

				switch arg[0] {
				case ice.STORY_PULL:
					list := m.Cmd(ice.WEB_STORY, ice.STORY_INDEX, m.Option("begin")).Append("list")
					for i := 0; i < 10 && list != "" && list != m.Option("end"); i++ {
						if m.Richs(ice.WEB_STORY, nil, list, func(key string, value map[string]interface{}) {
							// 节点信息
							m.Push("list", key)
							m.Push("node", kit.Format(value))
							m.Push("data", value["data"])
							m.Push("save", kit.Format(m.Richs(ice.WEB_CACHE, nil, value["data"], nil)))
							list = kit.Format(value["prev"])
						}) == nil {
							break
						}
					}
					m.Log(ice.LOG_EXPORT, "%s %s", m.Option("begin"), m.Format("append"))

				case ice.STORY_PUSH:
					if m.Richs(ice.WEB_CACHE, nil, m.Option("data"), nil) == nil {
						// 导入缓存
						m.Log(ice.LOG_IMPORT, "%v: %v", m.Option("data"), m.Option("save"))
						m.Conf(ice.WEB_CACHE, kit.Keys("hash", m.Option("data")), kit.UnMarshal(m.Option("save")))
					}

					node := kit.UnMarshal(m.Option("node")).(map[string]interface{})
					if m.Richs(ice.WEB_STORY, nil, m.Option("list"), nil) == nil {
						// 导入节点
						m.Log(ice.LOG_IMPORT, "%v: %v", m.Option("list"), m.Option("node"))
						m.Conf(ice.WEB_STORY, kit.Keys("hash", m.Option("list")), node)
					}

					if head := m.Richs(ice.WEB_STORY, "head", m.Option("story"), nil); head == nil {
						// 自动创建
						h := m.Rich(ice.WEB_STORY, "head", kit.Dict(
							"scene", node["scene"], "story", m.Option("story"),
							"count", node["count"], "list", m.Option("list"),
						))
						m.Log(ice.LOG_CREATE, "%v: %v", h, m.Option("story"))
					} else if head["list"] == kit.Format(node["prev"]) || head["list"] == kit.Format(node["pull"]) {
						// 快速合并
						head["list"] = m.Option("list")
						head["count"] = node["count"]
						head["time"] = node["time"]
					} else {
						// 推送失败
					}

				case ice.STORY_UPLOAD:
					// 上传数据
					m.Cmdy(ice.WEB_CACHE, "upload")

				case ice.STORY_DOWNLOAD:
					// 下载数据
					m.Cmdy(ice.WEB_STORY, ice.STORY_INDEX, arg[1])
					m.Render(kit.Select(ice.RENDER_DOWNLOAD, ice.RENDER_RESULT, m.Append("file") == ""), m.Append("text"))
				}
			}},
		}}, nil)
}
