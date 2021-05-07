package web

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

func _spide_create(m *ice.Message, name, address string) {
	if uri, e := url.Parse(address); e == nil && address != "" {
		if m.Richs(SPIDE, nil, name, func(key string, value map[string]interface{}) {
			kit.Value(value, "client.hostname", uri.Host)
			kit.Value(value, "client.url", address)
		}) == nil {
			dir, file := path.Split(uri.EscapedPath())
			m.Echo(m.Rich(SPIDE, nil, kit.Dict(
				"cookie", kit.Dict(), "header", kit.Dict(), "client", kit.Dict(
					"name", name, "url", address, "method", SPIDE_POST,
					"protocol", uri.Scheme, "hostname", uri.Host,
					"path", dir, "file", file, "query", uri.RawQuery,
					"timeout", "600s", "logheaders", false,
				),
			)))
		}
		m.Log_CREATE(SPIDE, name, ADDRESS, address)
	}
}

const (
	SPIDE_SHY  = "shy"
	SPIDE_DEV  = "dev"
	SPIDE_SELF = "self"

	SPIDE_RAW   = "raw"
	SPIDE_MSG   = "msg"
	SPIDE_SAVE  = "save"
	SPIDE_CACHE = "cache"
	SPIDE_PROXY = "proxy"
	SPIDE_CB    = "spide.cb"

	SPIDE_GET    = "GET"
	SPIDE_PUT    = "PUT"
	SPIDE_POST   = "POST"
	SPIDE_DELETE = "DELETE"

	SPIDE_FORM = "form"
	SPIDE_PART = "part"
	SPIDE_JSON = "json"
	SPIDE_DATA = "data"
	SPIDE_FILE = "file"
	SPIDE_BODY = "body"

	SPIDE_CLIENT = "client"
	SPIDE_METHOD = "method"
	SPIDE_HEADER = "header"
	SPIDE_COOKIE = "cookie"

	ContentType   = "Content-Type"
	ContentLength = "Content-Length"

	ContentFORM = "application/x-www-form-urlencoded"
	ContentJSON = "application/json"
	ContentHTML = "text/html"
	ContentPNG  = "image/png"
)
const (
	ADDRESS  = "address"
	REQUEST  = "request"
	RESPONSE = "response"
	PROTOCOL = "protocol"
)
const SPIDE = "spide"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			SPIDE: {Name: SPIDE, Help: "蜘蛛侠", Value: kit.Data(kit.MDB_SHORT, "client.name")},
		},
		Commands: map[string]*ice.Command{
			SPIDE_GET: {Name: "GET url key value 执行:button", Help: "蜘蛛侠", Action: map[string]*ice.Action{
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, SPIDE, "", mdb.HASH, "client.name", m.Option("client.name"))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Echo(kit.Formats(kit.UnMarshal(m.Cmdx(SPIDE, SPIDE_DEV, SPIDE_RAW, SPIDE_GET, arg[0], arg[1:]))))
			}},
			SPIDE_POST: {Name: "POST url key value 执行:button", Help: "蜘蛛侠", Action: map[string]*ice.Action{
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, SPIDE, "", mdb.HASH, "client.name", m.Option("client.name"))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Echo(kit.Formats(kit.UnMarshal(m.Cmdx(SPIDE, SPIDE_DEV, SPIDE_RAW, SPIDE_POST, arg[0], SPIDE_JSON, arg[1:]))))
			}},

			SPIDE: {Name: "spide client.name action=raw,msg,save,cache method=GET,PUT,POST,DELETE url format=form,part,json,data,file arg 执行:button create", Help: "蜘蛛侠", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create name address", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					if arg[0] != "name" {
						m.Option("name", arg[0])
						m.Option(ADDRESS, arg[1])
					}
					_spide_create(m, m.Option(kit.MDB_NAME), m.Option(ADDRESS))
				}},
				mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.MODIFY, SPIDE, "", mdb.HASH, "client.name", m.Option("client.name"), arg)
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, SPIDE, "", mdb.HASH, "client.name", m.Option("client.name"))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) < 2 || arg[0] == "" || (len(arg) > 3 && arg[3] == "") {
					m.Fields(len(arg) == 0 || arg[0] == "", "time,client.name,client.url")
					m.Cmdy(mdb.SELECT, SPIDE, "", mdb.HASH, "client.name", arg)
					m.PushAction(mdb.REMOVE)
					return
				}

				m.Richs(SPIDE, nil, arg[0], func(key string, value map[string]interface{}) {
					client := value[SPIDE_CLIENT].(map[string]interface{})
					// 缓存数据
					cache, save := "", ""
					switch arg[1] {
					case SPIDE_RAW, SPIDE_PROXY:
						cache, arg = arg[1], arg[1:]
					case SPIDE_MSG:
						cache, arg = arg[1], arg[1:]
					case SPIDE_SAVE:
						cache, save, arg = arg[1], arg[2], arg[2:]
					case SPIDE_CACHE:
						cache, arg = arg[1], arg[1:]
					}

					// 请求方法
					method := kit.Select(SPIDE_POST, client[SPIDE_METHOD])
					switch arg = arg[1:]; arg[0] {
					case SPIDE_GET:
						method, arg = SPIDE_GET, arg[1:]
					case SPIDE_PUT:
						method, arg = SPIDE_PUT, arg[1:]
					case SPIDE_POST:
						method, arg = SPIDE_POST, arg[1:]
					case SPIDE_DELETE:
						method, arg = SPIDE_DELETE, arg[1:]
					}

					// 请求地址
					uri, arg := arg[0], arg[1:]
					if strings.HasPrefix(uri, "ftp") {
						m.Cmdy(cli.SYSTEM, "wget", uri, arg)
						return
					}

					// 渲染引擎
					head := map[string]string{}
					body, ok := m.Optionv(SPIDE_BODY).(io.Reader)
					if !ok && len(arg) > 0 && method != SPIDE_GET {
						switch arg[0] {
						case SPIDE_FILE:
							if f, e := os.Open(arg[1]); m.Assert(e) {
								defer f.Close()
								body, arg = f, arg[2:]
							}
						case SPIDE_DATA:
							body, arg = bytes.NewBufferString(arg[1]), arg[2:]
						case SPIDE_PART:
							buf := &bytes.Buffer{}
							mp := multipart.NewWriter(buf)
							cache := time.Now().Add(-time.Hour * 240000)
							for i := 1; i < len(arg)-1; i += 2 {
								if arg[i] == "cache" {
									if t, e := time.ParseInLocation(ice.MOD_TIME, arg[i+1], time.Local); e == nil {
										cache = t
									}
								}
								if strings.HasPrefix(arg[i+1], "@") {
									if s, e := os.Stat(arg[i+1][1:]); e == nil {
										if s.ModTime().Before(cache) {
											return
										}
									}
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
							head[ContentType] = mp.FormDataContentType()
							body = buf
						case SPIDE_FORM:
							data := []string{}
							for i := 1; i < len(arg)-1; i += 2 {
								data = append(data, url.QueryEscape(arg[i])+"="+url.QueryEscape(arg[i+1]))
							}
							body = bytes.NewBufferString(strings.Join(data, "&"))
							head[ContentType] = ContentFORM
						case SPIDE_JSON:
							arg = arg[1:]
							fallthrough
						default:
							data := map[string]interface{}{}
							for i := 0; i < len(arg)-1; i += 2 {
								kit.Value(data, arg[i], arg[i+1])
							}
							if b, e := json.Marshal(data); m.Assert(e) {
								head[ContentType] = ContentJSON
								body = bytes.NewBuffer(b)
							}
							m.Log_EXPORT(SPIDE_JSON, kit.Format(data))
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
					kit.Fetch(value[SPIDE_HEADER], func(key string, value string) {
						req.Header.Set(key, value)
					})
					kit.Fetch(value[SPIDE_COOKIE], func(key string, value string) {
						req.AddCookie(&http.Cookie{Name: key, Value: value})
						m.Info("%s: %s", key, value)
					})
					list := kit.Simple(m.Optionv(SPIDE_HEADER))
					for i := 0; i < len(list)-1; i += 2 {
						req.Header.Set(list[i], list[i+1])
						m.Info("%s: %s", list[i], list[i+1])
					}
					for k, v := range head {
						req.Header.Set(k, v)
					}

					// 请求代理
					web := m.Target().Server().(*Frame)
					if web.Client == nil {
						web.Client = &http.Client{Timeout: kit.Duration(kit.Format(client["timeout"]))}
					}
					if req.Method == SPIDE_POST {
						m.Info("%s: %s", req.Header.Get(ContentLength), req.Header.Get(ContentType))
					}

					// 发送请求
					res, e := web.Client.Do(req)
					if m.Warn(e != nil, ice.ErrNotFound, e) {
						return
					}

					// 检查结果
					defer res.Body.Close()
					m.Cost("status", res.Status, "length", res.Header.Get(ContentLength), "type", res.Header.Get(ContentType))

					switch cb := m.Optionv(SPIDE_CB).(type) {
					case func(*ice.Message, *http.Request, *http.Response):
						cb(m, req, res)
						return
					}

					if m.Warn(res.StatusCode != http.StatusOK, res.Status) {
						m.Set(ice.MSG_RESULT)
						// return
						switch res.StatusCode {
						case http.StatusNotFound:
							return
						default:
						}

					}

					// 缓存变量
					for _, v := range res.Cookies() {
						kit.Value(value, "cookie."+v.Name, v.Value)
						m.Log(ice.LOG_IMPORT, "%s: %s", v.Name, v.Value)
					}

					// 解析引擎
					switch cache {
					case SPIDE_PROXY:
						m.Optionv(RESPONSE, res)
						m.Cmdy(CACHE, DOWNLOAD, res.Header.Get(ContentType), uri)
						m.Echo(m.Append(DATA))

					case SPIDE_CACHE:
						m.Optionv(RESPONSE, res)
						m.Cmdy(CACHE, DOWNLOAD, res.Header.Get(ContentType), uri)
						m.Echo(m.Append(DATA))

					case SPIDE_SAVE:
						if f, p, e := kit.Create(save); m.Assert(e) {
							if n, e := io.Copy(f, res.Body); m.Assert(e) {
								m.Log_EXPORT(kit.MDB_SIZE, n, kit.MDB_FILE, p)
								m.Echo(p)
							}
						}

					case SPIDE_RAW:
						if b, e := ioutil.ReadAll(res.Body); m.Assert(e) {
							m.Echo(kit.Formats(kit.UnMarshal(string(b))))
							// m.Echo(string(b))
						}

					case SPIDE_MSG:
						var data map[string][]string
						m.Assert(json.NewDecoder(res.Body).Decode(&data))
						m.Info("res: %s", kit.Format(data))
						for _, k := range data[ice.MSG_APPEND] {
							for i := range data[k] {
								m.Push(k, data[k][i])
							}
						}
						m.Resultv(data[ice.MSG_RESULT])

					default:
						b, _ := ioutil.ReadAll(res.Body)
						m.Echo(string(b))
						break
						if strings.HasPrefix(res.Header.Get(ContentType), ContentHTML) {
							b, _ := ioutil.ReadAll(res.Body)
							m.Echo(string(b))
							break
						}

						var data interface{}
						m.Assert(json.NewDecoder(res.Body).Decode(&data))
						m.Optionv("content_data", data)

						data = kit.KeyValue(map[string]interface{}{}, "", data)
						m.Push("", data)
					}
				})
			}},
		}})
}
