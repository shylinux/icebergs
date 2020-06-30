package web

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/tcp"
	kit "github.com/shylinux/toolkits"

	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
)

func _spide_list(m *ice.Message, name string) {
	if name == "" {
		m.Richs(SPIDE, nil, kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
			m.Push(key, value["client"], []string{"name", "share", "login", "method", "url"})
		})
		m.Sort("name")
		return
	}

	m.Richs(SPIDE, nil, name, func(key string, value map[string]interface{}) {
		m.Push("detail", value)
		if kit.Value(value, "client.share") != nil {
			m.Push("key", "share")
			m.Push("value", fmt.Sprintf(m.Conf(SHARE, "meta.template.text"), m.Conf(SHARE, "meta.domain"), kit.Value(value, "client.share")))
		}
	})
}
func _spide_show(m *ice.Message, name string) {
}
func _spide_login(m *ice.Message, name string) {
	m.Richs(SPIDE, nil, name, func(key string, value map[string]interface{}) {
		msg := m.Cmd(SPIDE, name, "msg", "/route/login", "login")
		if msg.Append(ice.MSG_USERNAME) != "" {
			m.Echo(msg.Append(ice.MSG_USERNAME))
			return
		}
		if msg.Result() != "" {
			kit.Value(value, "client.login", msg.Result())
			kit.Value(value, "client.share", m.Cmdx(SHARE, SPIDE, name,
				kit.Format("%s?sessid=%s", kit.Value(value, "client.url"), kit.Value(value, "cookie.sessid"))))
		}
		m.Render(ice.RENDER_QRCODE, kit.Dict(
			kit.MDB_TYPE, "login", kit.MDB_NAME, name,
			kit.MDB_TEXT, kit.Value(value, "cookie.sessid"),
		))
	})
}
func _spide_create(m *ice.Message, name, address string, arg ...string) {
	if uri, e := url.Parse(address); e == nil && address != "" {
		if uri.Host == "random" {
			uri.Host = ":" + m.Cmdx(tcp.PORT, "get")
			address = strings.Replace(address, "random", uri.Host, -1)
		}

		if m.Richs(SPIDE, nil, name, func(key string, value map[string]interface{}) {
			kit.Value(value, "client.hostname", uri.Host)
			kit.Value(value, "client.url", address)
		}) == nil {
			dir, file := path.Split(uri.EscapedPath())
			m.Rich(SPIDE, nil, kit.Dict(
				"cookie", kit.Dict(), "header", kit.Dict(), "client", kit.Dict(
					"share", ShareCreate(m.Spawn(), SPIDE, name, address),
					"name", name, "url", address, "method", "POST",
					"protocol", uri.Scheme, "hostname", uri.Host,
					"path", dir, "file", file, "query", uri.RawQuery,
					"timeout", "100s", "logheaders", false,
				),
			))
		}
		m.Log_CREATE(SPIDE, name, "address", address)
	}
}
func _spide_search(m *ice.Message, kind, name, text string, arg ...string) {
	m.Richs(SPIDE, nil, kit.Select(kit.MDB_FOREACH, ""), func(key string, value map[string]interface{}) {
		if kit.Format(kit.Value(value, "client.name")) != name && (text == "" || !strings.Contains(kit.Format(kit.Value(value, "client.url")), text)) {
			return
		}

		m.Push("pod", m.Option("pod"))
		m.Push("ctx", "web")
		m.Push("cmd", SPIDE)
		m.Push(key, value, []string{kit.MDB_TIME})
		m.Push(kit.MDB_SIZE, 0)
		m.Push("type", SPIDE)
		// m.Push("type", kit.Format(kit.Value(value, "client.protocol")))
		m.Push("name", kit.Format(kit.Value(value, "client.name")))
		m.Push("text", kit.Format(kit.Value(value, "client.url")))
	})
}
func _spide_render(m *ice.Message, kind, name, text string, arg ...string) {
	m.Echo(`<iframe src="%s" width=800 height=400></iframe>`, text)
}

const SPIDE = "spide"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			SPIDE: {Name: "spide", Help: "蜘蛛侠", Value: kit.Data(kit.MDB_SHORT, "client.name")},
		},
		Commands: map[string]*ice.Command{
			SPIDE: {Name: "spide name=auto [action:select=msg|raw|cache] [method:select=POST|GET] url [format:select=json|form|part|data|file] arg... auto", Help: "蜘蛛侠", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create name address", Help: "", Hand: func(m *ice.Message, arg ...string) {
					_spide_create(m, arg[0], arg[1])
				}},
				mdb.SEARCH: {Name: "search type name text arg...", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
					_spide_search(m, arg[0], arg[1], arg[2], arg[3:]...)
				}},
				mdb.RENDER: {Name: "render type name text arg...", Help: "渲染", Hand: func(m *ice.Message, arg ...string) {
					_spide_render(m, arg[0], arg[1], arg[2], arg[3:]...)
				}},
				"login": {Name: "login name", Help: "", Hand: func(m *ice.Message, arg ...string) {
					_spide_login(m, arg[0])
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) < 2 || len(arg) > 3 && arg[3] == "" {
					_spide_list(m, kit.Select("", arg, 1))
					return
				}

				m.Richs(SPIDE, nil, arg[0], func(key string, value map[string]interface{}) {
					client := value["client"].(map[string]interface{})
					// 缓存数据
					cache := ""
					switch arg[1] {
					case "raw":
						cache, arg = arg[1], arg[1:]
					case "msg":
						cache, arg = arg[1], arg[1:]
					case "cache":
						cache, arg = arg[1], arg[1:]
					}

					// 请求方法
					method := kit.Select("POST", client["method"])
					switch arg = arg[1:]; arg[0] {
					case "POST":
						method, arg = "POST", arg[1:]
					case "GET":
						method, arg = "GET", arg[1:]
					}

					// 请求地址
					uri, arg := arg[0], arg[1:]

					// 渲染引擎
					head := map[string]string{}
					body, ok := m.Optionv("body").(io.Reader)
					if !ok && len(arg) > 0 && method != "GET" {
						switch arg[0] {
						case "file":
							if f, e := os.Open(arg[1]); m.Warn(e != nil, "%s", e) {
								return
							} else {
								defer f.Close()
								body, arg = f, arg[2:]
							}
						case "data":
							body, arg = bytes.NewBufferString(arg[1]), arg[2:]
						case "part":
							buf := &bytes.Buffer{}
							mp := multipart.NewWriter(buf)
							for i := 1; i < len(arg)-1; i += 2 {
								if strings.HasPrefix(arg[i+1], "@") {
									if f, e := os.Open(arg[i+1][1:]); m.Assert(e) {
										defer f.Close()
										if p, e := mp.CreateFormFile(arg[i], path.Base(arg[i+1][1:])); m.Assert(e) {
											io.Copy(p, f)
										}
									}
								} else {
									mp.WriteField(arg[i], arg[i+1])
								}
							}
							mp.Close()
							head["Content-Type"] = mp.FormDataContentType()
							body = buf
						case "form":
							data := []string{}
							for i := 1; i < len(arg)-1; i += 2 {
								data = append(data, url.QueryEscape(arg[i])+"="+url.QueryEscape(arg[i+1]))
							}
							body = bytes.NewBufferString(strings.Join(data, "&"))
							head["Content-Type"] = "application/x-www-form-urlencoded"
						case "json":
							arg = arg[1:]
							fallthrough
						default:
							data := map[string]interface{}{}
							for i := 0; i < len(arg)-1; i += 2 {
								kit.Value(data, arg[i], arg[i+1])
							}
							if b, e := json.Marshal(data); m.Assert(e) {
								head["Content-Type"] = "application/json"
								body = bytes.NewBuffer(b)
							}
							m.Log(ice.LOG_EXPORT, "json: %s", kit.Format(data))
						}
						arg = arg[:0]
					} else {
						body = bytes.NewBuffer([]byte{})
					}

					// 请求地址
					uri = kit.MergeURL2(kit.Format(client["url"]), uri, arg)
					req, e := http.NewRequest(method, uri, body)
					m.Info("%s %s", req.Method, req.URL)
					m.Assert(e)

					// 请求变量
					kit.Fetch(value["cookie"], func(key string, value string) {
						req.AddCookie(&http.Cookie{Name: key, Value: value})
						m.Info("%s: %s", key, value)
					})
					kit.Fetch(value["header"], func(key string, value string) {
						req.Header.Set(key, value)
					})
					list := kit.Simple(m.Optionv("header"))
					for i := 0; i < len(list)-1; i += 2 {
						req.Header.Set(list[i], list[i+1])
					}
					for k, v := range head {
						req.Header.Set(k, v)
					}

					// 请求代理
					web := m.Target().Server().(*Frame)
					if web.Client == nil {
						web.Client = &http.Client{Timeout: kit.Duration(kit.Format(client["timeout"]))}
					}
					if method == "POST" {
						m.Info("%s: %s", req.Header.Get("Content-Length"), req.Header.Get("Content-Type"))
					}

					// 发送请求
					res, e := web.Client.Do(req)
					if m.Warn(e != nil, "%s", e) {
						return
					}

					// 检查结果
					if m.Cost("%s %s: %s", res.Status, res.Header.Get("Content-Length"), res.Header.Get("Content-Type")); m.Warn(res.StatusCode != http.StatusOK, "%s", res.Status) {
						return
					}

					// 缓存变量
					for _, v := range res.Cookies() {
						kit.Value(value, "cookie."+v.Name, v.Value)
						m.Log(ice.LOG_IMPORT, "%s: %s", v.Name, v.Value)
					}

					// 解析引擎
					switch cache {
					case "cache":
						m.Optionv("response", res)
						m.Cmdy(CACHE, DOWNLOAD, res.Header.Get("Content-Type"), uri)
						m.Echo(m.Append(DATA))
					case "raw":
						if b, e := ioutil.ReadAll(res.Body); m.Assert(e) {
							m.Echo(string(b))
						}
					case "msg":
						var data map[string][]string
						m.Assert(json.NewDecoder(res.Body).Decode(&data))
						m.Info("res: %s", kit.Formats(data))
						for _, k := range data[ice.MSG_APPEND] {
							for i := range data[k] {
								m.Push(k, data[k][i])
							}
						}
						m.Resultv(data[ice.MSG_RESULT])
					default:
						if strings.HasPrefix(res.Header.Get("Content-Type"), "text/html") {
							b, _ := ioutil.ReadAll(res.Body)
							m.Echo(string(b))
							break
						}

						var data interface{}
						m.Assert(json.NewDecoder(res.Body).Decode(&data))
						data = kit.KeyValue(map[string]interface{}{}, "", data)
						m.Info("res: %s", kit.Format(data))
						m.Push("", data)
					}
				})
			}},
		}}, nil)
}
