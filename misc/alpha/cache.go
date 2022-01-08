package alpha

import (
	"shylinux.com/x/ice"
)

type cache struct {
	ice.Hash
	short  string `data:"word"`
	field  string `data:"time,word,translation,definition"`
	create string `name:"create word translation definition" help:"创建"`
}

func (c cache) Create(m *ice.Message, arg ...string) {
	c.Hash.Create(m, arg...)
}

func init() { ice.Cmd("web.wiki.alpha.cache", cache{}) }
