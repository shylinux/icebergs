package lex

import ice "shylinux.com/x/icebergs"

const LEX = "lex"

var Index = &ice.Context{Name: LEX, Help: "词法模块"}

func init() { ice.Index.Register(Index, nil, SPLIT) }
