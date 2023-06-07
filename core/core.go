package core

import (
	"errors"
	"strings"

	"github.com/Yuzuki616/V2bX/conf"
)

var (
	cores = map[string]func(c *conf.CoreConfig) (Core, error){}
)

func NewCore(c *conf.CoreConfig) (Core, error) {
	if f, ok := cores[strings.ToLower(c.Type)]; ok {
		return f(c)
	} else {
		return nil, errors.New("unknown core type")
	}
}

func RegisterCore(t string, f func(c *conf.CoreConfig) (Core, error)) {
	cores[t] = f
}

func RegisteredCore() []string {
	cs := make([]string, 0, len(cores))
	for k := range cores {
		cs = append(cs, k)
	}
	return cs
}
