package input

import (
	"shylinux.com/x/ice"
)

type wubi struct {
	input

	short string `data:"zone"`
	store string `data:"usr/local/export/input/wubi"`
	fsize string `data:"300000"`
	limit string `data:"50000"`
	least string `data:"1000"`

	insert string `name:"insert zone=person text code weight" help:"添加"`
	load   string `name:"load file=usr/wubi-dict/wubi86 zone=wubi86" help:"加载"`
	save   string `name:"save file=usr/wubi-dict/person zone=person" help:"保存"`
	list   string `name:"list method=word,line code auto" help:"五笔"`
}

func init() { ice.Cmd("web.code.input.wubi", wubi{}) }
