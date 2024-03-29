package alpha

import (
	"shylinux.com/x/ice"
)

type cache struct {
	ice.Hash
	short string `data:"word"`
	field string `data:"time,word,phonetic,translation,definition"`
	list  string `name:"list word auto create prunes" help:"缓存"`
}

func init() { ice.Cmd("web.wiki.alpha.cache", cache{}) }
