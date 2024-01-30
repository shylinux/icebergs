package disk

import (
	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/chat/oauth"
	kit "shylinux.com/x/toolkits"
)

const (
	BAIDU     = "baidu"
	AUTH_URL  = "http://openapi.baidu.com"
	API_URL   = "https://pan.baidu.com/rest/2.0/xpan/"
	USER_INFO = "nas?method=uinfo"
	FILE_LIST = "file?method=list"
	FILE_META = "multimedia?method=filemetas"
)

type Client struct {
	oauth.Client
	list string `name:"list hash path orgs:text repo:text auto" help:"仓库" icon:"netdisk.png"`
}

func init() {
	oauth.Inputs[BAIDU] = map[string]string{
		oauth.OAUTH_URL: "/oauth/2.0/authorize",
		oauth.GRANT_URL: "/oauth/2.0/token",
		oauth.TOKEN_URL: "/oauth/2.0/token",
		oauth.USERS_URL: API_URL + USER_INFO,
		oauth.NICK_KEY:  "baidu_name",
		oauth.USER_KEY:  "uk",
		oauth.SCOPE:     "basic,netdisk",
	}
}
func (s Client) Init(m *ice.Message, arg ...string) {
	m.Cmd(web.SPIDE, mdb.CREATE, BAIDU, AUTH_URL, "", "usr/icons/netdisk.png")
	s.Hash.Init(m, arg...)
}
func (s Client) Login(m *ice.Message, arg ...string) {
	s.Client.Login2(m, arg...)
}
func (s Client) List(m *ice.Message, arg ...string) {
	if len(arg) == 0 {
		s.Client.List(m, arg...)
		return
	}
	res := s.Client.Get(m, arg[0], API_URL+FILE_LIST, nfs.DIR, kit.Select("", arg, 1))
	kit.For(kit.Value(res, mdb.LIST), func(value ice.Map) {
		m.Push(mdb.TIME, kit.TimeUnix(value["server_mtime"]))
		m.Push(nfs.PATH, kit.Format(value[nfs.PATH])+kit.Select("", nfs.PS, kit.Format(value["isdir"]) == "1"))
		m.Push(nfs.SIZE, kit.FmtSize(kit.Int(value[nfs.SIZE])))
		m.Push(mdb.ID, value["fs_id"])
	})
	m.PushAction("show")
}
func (s Client) Show(m *ice.Message, arg ...string) {
	res := s.Client.Get(m, m.Option(mdb.HASH), API_URL+FILE_META, "dlink", "1", "fsids", kit.Format("%v", []string{m.Option(mdb.ID)}))
	p := "usr/local/disk/" + m.Option(mdb.ID)
	s.Save(m, m.Option(mdb.HASH), p, kit.Format(kit.Value(res, "list.0.dlink")))
	defer web.ToastSuccess(m.Message)
	switch kit.Format(kit.Value(res, "list.0.category")) { // 1 视频、2 音频、3 图片、4 文档、5 应用、6 其他、7 种子
	case "4":
		m.Cmdy(nfs.CAT, p)
	}
}
func init() { ice.WikiCtxCmd(Client{}) }
