package ice

import (
	kit "github.com/shylinux/toolkits"

	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"sort"
	"strings"
)

const (
	ZONE = "zone"
)

func (m *Message) Prefile(favor string, id string) map[string]string {
	res := map[string]string{}
	m.Option("render", "")
	m.Option("_action", "")
	m.Cmd(WEB_FAVOR, kit.Select(m.Option("favor"), favor), id).Table(func(index int, value map[string]string, head []string) {
		res[value["key"]] = value["value"]
	})

	res["content"] = m.Cmdx(CLI_SYSTEM, "sed", "-n", kit.Format("%d,%dp", kit.Int(res["extra.row"]), kit.Int(res["extra.row"])+3), res["extra.buf"])
	return res
}
func (m *Message) Prefix(arg ...string) string {
	return kit.Keys(m.Cap(CTX_FOLLOW), arg)
}
func (m *Message) Save(arg ...string) *Message {
	list := []string{}
	for _, k := range arg {
		list = append(list, kit.Keys(m.Cap(CTX_FOLLOW), k))
	}
	m.Cmd(CTX_CONFIG, "save", kit.Keys(m.Cap(CTX_FOLLOW), "json"), list)
	return m
}
func (m *Message) Load(arg ...string) *Message {
	list := []string{}
	for _, k := range arg {
		list = append(list, kit.Keys(m.Cap(CTX_FOLLOW), k))
	}
	m.Cmd(CTX_CONFIG, "load", kit.Keys(m.Cap(CTX_FOLLOW), "json"), list)
	return m
}

func (m *Message) Richs(key string, chain interface{}, raw interface{}, cb interface{}) (res map[string]interface{}) {
	// 数据结构
	cache := m.Confm(key, chain)
	if cache == nil {
		return nil
	}
	meta, ok := cache[kit.MDB_META].(map[string]interface{})
	hash, ok := cache[kit.MDB_HASH].(map[string]interface{})
	if !ok {
		return nil
	}

	h := kit.Format(raw)
	switch h {
	case "*":
		// 全部遍历
		switch cb := cb.(type) {
		case func(string, string):
			for k, v := range hash {
				cb(k, kit.Format(v))
			}
		case func(string, map[string]interface{}):
			for k, v := range hash {
				res = v.(map[string]interface{})
				cb(k, res)
			}
		}
		return res
	case "%":
		// 随机选取
		if len(hash) > 0 {
			list := []string{}
			for k := range hash {
				list = append(list, k)
			}
			h = list[rand.Intn(len(list))]
			res, _ = hash[h].(map[string]interface{})
		}
	default:
		// 单个查询
		if res, ok = hash[h].(map[string]interface{}); !ok {
			switch kit.Format(kit.Value(meta, kit.MDB_SHORT)) {
			case "", "uniq":
			default:
				hh := kit.Hashs(h)
				if res, ok = hash[hh].(map[string]interface{}); ok {
					h = hh
					break
				}

				prefix := path.Join(kit.Select(m.Conf(WEB_CACHE, "meta.store"), kit.Format(meta["store"])), key)
				for _, k := range []string{h, hh} {
					if f, e := os.Open(path.Join(prefix, kit.Keys(k, "json"))); e == nil {
						defer f.Close()
						if b, e := ioutil.ReadAll(f); e == nil {
							if json.Unmarshal(b, &res) == e {
								h = k
								m.Log(LOG_IMPORT, "%s/%s.json", prefix, k)
								break
							}
						}
					}
				}
			}
		}
	}

	// 返回数据
	if res != nil {
		switch cb := cb.(type) {
		case func(map[string]interface{}):
			cb(res)
		case func(string, map[string]interface{}):
			cb(h, res)
		}
	}
	return res
}
func (m *Message) Rich(key string, chain interface{}, data interface{}) string {
	// 数据结构
	cache := m.Confm(key, chain)
	if cache == nil {
		cache = map[string]interface{}{}
		m.Confv(key, chain, cache)
	}
	meta, ok := cache[kit.MDB_META].(map[string]interface{})
	if !ok {
		meta = map[string]interface{}{}
		cache[kit.MDB_META] = meta
	}
	hash, ok := cache[kit.MDB_HASH].(map[string]interface{})
	if !ok {
		hash = map[string]interface{}{}
		cache[kit.MDB_HASH] = hash
	}

	// 通用数据
	prefix := kit.Select("", "meta.", kit.Value(data, "meta") != nil)
	if kit.Value(data, prefix+kit.MDB_TIME) == nil {
		kit.Value(data, prefix+kit.MDB_TIME, m.Time())
	}

	// 生成键值
	h := ""
	switch short := kit.Format(kit.Value(meta, kit.MDB_SHORT)); short {
	case "":
		h = kit.ShortKey(hash, 6)
	case "uniq":
		h = kit.Hashs("uniq")
	case "data":
		h = kit.Hashs(kit.Format(data))
	default:
		if kit.Value(data, "meta") != nil {
			h = kit.Hashs(kit.Format(kit.Value(data, "meta."+short)))
		} else {
			h = kit.Hashs(kit.Format(kit.Value(data, short)))
		}
	}

	// 添加数据
	if hash[h] = data; len(hash) >= kit.Int(kit.Select(m.Conf(WEB_CACHE, "meta.limit"), kit.Format(meta["limit"]))) {
		least := kit.Int(kit.Select(m.Conf(WEB_CACHE, "meta.least"), kit.Format(meta["least"])))

		// 时间淘汰
		list := []int{}
		for _, v := range hash {
			list = append(list, kit.Time(kit.Format(kit.Value(v, "time"))))
		}
		sort.Ints(list)
		dead := list[len(list)-1-least]

		prefix := path.Join(kit.Select(m.Conf(WEB_CACHE, "meta.store"), kit.Format(meta["store"])), key)
		for k, v := range hash {
			if kit.Time(kit.Format(kit.Value(v, "time"))) > dead {
				break
			}

			name := path.Join(prefix, kit.Keys(k, "json"))
			if f, p, e := kit.Create(name); m.Assert(e) {
				defer f.Close()
				// 保存数据
				if n, e := f.WriteString(kit.Format(v)); m.Assert(e) {
					m.Log(LOG_EXPORT, "%s: %d", p, n)
					delete(hash, k)
				}
			}
		}
	}

	return h
}
func (m *Message) Grow(key string, chain interface{}, data interface{}) int {
	// 数据结构
	cache := m.Confm(key, chain)
	if cache == nil {
		cache = map[string]interface{}{}
		m.Confv(key, chain, cache)
	}
	meta, ok := cache[kit.MDB_META].(map[string]interface{})
	if !ok {
		meta = map[string]interface{}{}
		cache[kit.MDB_META] = meta
	}
	list, _ := cache[kit.MDB_LIST].([]interface{})

	// 通用数据
	id := kit.Int(meta["count"]) + 1
	prefix := kit.Select("", "meta.", kit.Value(data, "meta") != nil)
	if kit.Value(data, prefix+kit.MDB_ID, id); kit.Value(data, prefix+kit.MDB_TIME) == nil {
		kit.Value(data, prefix+kit.MDB_TIME, kit.Select(m.Time(), m.Option("time")))
	}

	// 添加数据
	list = append(list, data)
	cache[kit.MDB_LIST] = list
	meta["count"] = id

	// 保存数据
	if len(list) >= kit.Int(kit.Select(m.Conf(WEB_CACHE, "meta.limit"), kit.Format(meta["limit"]))) {
		least := kit.Int(kit.Select(m.Conf(WEB_CACHE, "meta.least"), kit.Format(meta["least"])))

		record, _ := meta["record"].([]interface{})

		// 文件命名
		prefix := path.Join(kit.Select(m.Conf(WEB_CACHE, "meta.store"), kit.Format(meta["store"])), key)
		name := path.Join(prefix, kit.Keys(kit.Select("list", chain), "csv"))
		if len(record) > 0 {
			name = kit.Format(kit.Value(record, kit.Keys(len(record)-1, "file")))
			if s, e := os.Stat(name); e == nil {
				if s.Size() > kit.Int64(kit.Select(m.Conf(WEB_CACHE, "meta.fsize"), kit.Format(meta["fsize"]))) {
					name = fmt.Sprintf("%s/%s_%d.csv", prefix, kit.Select("list", chain), kit.Int(meta["offset"]))
				}
			}
		}

		// 打开文件
		f, e := os.OpenFile(name, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
		if e != nil {
			f, _, e = kit.Create(name)
			m.Info("%s.%v create: %s", key, chain, name)
		} else {
			m.Info("%s.%v append: %s", key, chain, name)
		}
		defer f.Close()
		s, e := f.Stat()
		m.Assert(e)

		// 保存表头
		keys := []string{}
		w := csv.NewWriter(f)
		if s.Size() == 0 {
			for k := range list[0].(map[string]interface{}) {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			w.Write(keys)
			m.Info("write head: %v", keys)
			w.Flush()
			s, e = f.Stat()
		} else {
			r := csv.NewReader(f)
			keys, e = r.Read()
			m.Info("read head: %v", keys)
		}

		// 创建索引
		count := len(list) - least
		offset := kit.Int(meta["offset"])
		meta["record"] = append(record, map[string]interface{}{
			"time": m.Time(), "offset": offset, "count": count,
			"file": name, "position": s.Size(),
		})

		// 保存数据
		for i, v := range list {
			if i >= count {
				break
			}

			val := v.(map[string]interface{})

			values := []string{}
			for _, k := range keys {
				values = append(values, kit.Format(val[k]))
			}
			w.Write(values)

			if i < least {
				list[i] = list[count+i]
			}
		}

		m.Log(LOG_INFO, "%s.%v save %s offset %v+%v", key, chain, name, offset, count)
		meta["offset"] = offset + count
		list = list[count:]
		cache[kit.MDB_LIST] = list
		w.Flush()
	}
	return id
}
func (m *Message) Grows(key string, chain interface{}, match string, value string, cb interface{}) map[string]interface{} {
	// 数据结构
	cache := m.Confm(key, chain)
	if cache == nil {
		return nil
	}
	meta, ok := cache[kit.MDB_META].(map[string]interface{})
	list, ok := cache[kit.MDB_LIST].([]interface{})
	if !ok {
		return nil
	}

	// 数据范围
	offend := kit.Int(kit.Select("0", m.Option("cache.offend")))
	limit := kit.Int(kit.Select("10", m.Option("cache.limit")))
	current := kit.Int(meta["offset"])
	end := current + len(list) - offend
	begin := end - limit
	switch limit {
	case -1:
		begin = current
	case -2:
		begin = 0
	}

	if match == kit.MDB_ID {
		begin, end = kit.Int(value)-1, kit.Int(value)
		match, value = "", ""
	}

	order := 0
	if begin < current {
		// 读取文件
		m.Log(LOG_INFO, "%s.%v read %v-%v from %v-%v", key, chain, begin, end, current, current+len(list))
		store, _ := meta["record"].([]interface{})
		for s := len(store) - 1; s > -1; s-- {
			item, _ := store[s].(map[string]interface{})
			line := kit.Int(item["offset"])
			m.Logs(LOG_INFO, "action", "check", "record", s, "offset", line, "count", item["count"])
			if begin < line && s > 0 {
				if kit.Int(item["count"]) != 0 {
					s -= (line - begin) / kit.Int(item["count"])
				}
				// 向后查找
				continue
			}

			for ; begin < end && s < len(store); s++ {
				item, _ := store[s].(map[string]interface{})
				name := kit.Format(item["file"])
				pos := kit.Int(item["position"])
				offset := kit.Int(item["offset"])
				if offset+kit.Int(item["count"]) <= begin {
					m.Logs(LOG_INFO, "action", "check", "record", s, "offset", line, "count", item["count"])
					// 向前查找
					continue
				}

				if f, e := os.Open(name); m.Assert(e) {
					defer f.Close()
					// 打开文件
					r := csv.NewReader(f)
					heads, _ := r.Read()
					m.Logs(LOG_IMPORT, "head", heads)

					f.Seek(int64(pos), os.SEEK_SET)
					r = csv.NewReader(f)
					for i := offset; i < end; i++ {
						lines, e := r.Read()
						if e != nil {
							m.Log(LOG_IMPORT, "load head %v", e)
							break
						}
						if i < begin {
							m.Logs(LOG_INFO, "action", "skip", "offset", i)
							continue
						}

						// 读取数据
						item := map[string]interface{}{}
						for i := range heads {
							if heads[i] == "extra" {
								item[heads[i]] = kit.UnMarshal(lines[i])
							} else {
								item[heads[i]] = lines[i]
							}
						}
						m.Logs(LOG_IMPORT, "offset", i, "type", item["type"], "name", item["name"], "text", item["text"])

						if match == "" || strings.Contains(kit.Format(item[match]), value) {
							// 匹配成功
							switch cb := cb.(type) {
							case func(int, map[string]interface{}):
								cb(order, item)
							case func(int, map[string]interface{}) bool:
								if cb(order, item) {
									return meta
								}
							}
							order++
						}
						begin = i + 1
					}
				}
			}
			break
		}
	}

	if begin < current {
		begin = current
	}
	for i := begin - current; i < end-current; i++ {
		// 读取缓存
		if match == "" || strings.Contains(kit.Format(kit.Value(list[i], match)), value) {
			switch cb := cb.(type) {
			case func(int, map[string]interface{}):
				cb(order, list[i].(map[string]interface{}))
			case func(int, map[string]interface{}) bool:
				if cb(order, list[i].(map[string]interface{})) {
					return meta
				}
			}
			order++
		}
	}
	return meta
}
func (m *Message) Show(cmd string, arg ...string) bool {
	if len(arg) == 0 {
		// 日志分类
		m.Richs(cmd, nil, "*", func(key string, value map[string]interface{}) {
			m.Push(key, value["meta"])
		})
		return true
	}
	if len(arg) < 3 {
		if m.Richs(cmd, nil, arg[0], func(key string, val map[string]interface{}) {
			if len(arg) == 1 {
				// 日志列表
				m.Grows(cmd, kit.Keys(kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
					m.Push(key, value)
				})
				return
			}
			// 日志详情
			m.Grows(cmd, kit.Keys(kit.MDB_HASH, key), "id", arg[1], func(index int, value map[string]interface{}) {
				m.Push("detail", value)
			})
		}) != nil {
			return true
		}
	}
	return false
}

func (m *Message) RichCreate(prefix string, zone string, arg ...string) {
}
func (m *Message) RichInsert(prefix string, zone string, kind, name, text string, data []string, arg ...string) {
}
func ListLook(name ...string) []interface{} {
	list := []interface{}{}
	for _, k := range name {
		list = append(list, kit.MDB_INPUT, "text", "name", k, "action", "auto")
	}
	return kit.List(append(list,
		kit.MDB_INPUT, "button", "name", "查看", "action", "auto",
		kit.MDB_INPUT, "button", "name", "返回", "cb", "Last",
	)...)
}
