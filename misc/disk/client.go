package disk

import (
	"time"

	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/chat/oauth"
	kit "shylinux.com/x/toolkits"
)

type Client struct {
	oauth.Client
	list string `name:"list hash path orgs:text repo:text auto" help:"仓库" icon:"gitea.png"`
}

func init() {
	oauth.Inputs["baidu"] = map[string]string{
		oauth.OAUTH_URL:    "/oauth/2.0/authorize",
		oauth.GRANT_URL:    "/oauth/2.0/token",
		oauth.TOKEN_URL:    "/oauth/2.0/token",
		oauth.USERS_URL:    "https://pan.baidu.com/rest/2.0/xpan/nas?method=uinfo",
		oauth.NICK_KEY:     "baidu_name",
		oauth.USER_KEY:     "uk",
		oauth.SCOPE:        "basic,netdisk",
		oauth.API_PREFIX:   "/api/v1/",
		oauth.TOKEN_PREFIX: "",
	}
}
func (s Client) Login(m *ice.Message, arg ...string) {
	s.Client.Login2(m, arg...)
}
func (s Client) Show(m *ice.Message, arg ...string) {
	res := s.Client.Get(m, m.Option(mdb.HASH), "https://pan.baidu.com/rest/2.0/xpan/multimedia?method=filemetas", "dlink", "1", "fsids", kit.Format("%v", []string{m.Option(mdb.ID)}))
	// 1 视频、2 音频、3 图片、4 文档、5 应用、6 其他、7 种子
	p := "usr/local/disk/" + m.Option(mdb.ID)
	s.Save(m, m.Option(mdb.HASH), p, kit.Format(kit.Value(res, "list.0.dlink")))
	web.ToastSuccess(m.Message)
	switch kit.Format(kit.Value(res, "list.0.category")) {
	case "4":
		m.Cmdy(nfs.CAT, p)
	}
}
func (s Client) List(m *ice.Message, arg ...string) {
	if len(arg) == 0 {
		s.Client.List(m, arg...)
		return
	}
	res := s.Client.Get(m, arg[0], "https://pan.baidu.com/rest/2.0/xpan/file?method=list", "dir", kit.Select("", arg, 1))
	kit.For(kit.Value(res, mdb.LIST), func(value ice.Map) {
		m.Push(mdb.TIME, time.Unix(kit.Int64(value["server_mtime"]), 0))
		m.Push(nfs.PATH, kit.Format(value[nfs.PATH])+kit.Select("", nfs.PS, kit.Format(value["isdir"]) == "1"))
		m.Push(nfs.SIZE, kit.FmtSize(kit.Int(value[nfs.SIZE])))
		m.Push(mdb.ID, value["fs_id"])
	})
	m.PushAction("show")
	m.Echo(kit.Formats(res))
}

func init() { ice.WikiCtxCmd(Client{}) }
