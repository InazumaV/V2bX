package core

import (
	"errors"
	"github.com/Yuzuki616/V2bX/conf"
)

var (
	cores = map[string]func(c *conf.CoreConfig) (Core, error){}
)

func NewCore(c *conf.CoreConfig) (Core, error) {
	if f, ok := cores[c.Type]; ok {
		return f(c)
	} else {
		return nil, errors.New("unknown core type")
	}
}

func RegisterCore(t string, f func(c *conf.CoreConfig) (Core, error)) {
	cores[t] = f
}
