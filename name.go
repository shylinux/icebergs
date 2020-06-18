package ice

import (
	"errors"
)

var names = map[string]interface{}{}

var ErrNameExists = errors.New("name already exists")

func Name(name string, value interface{}) string {
	if _, ok := names[name]; ok {
		println(name)
		panic(ErrNameExists)
	}
	names[name] = value
	return name
}
