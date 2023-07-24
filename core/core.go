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
	// multi core
	if types := strings.Split(c.Type, " "); len(types) > 1 {
		var cs []Core
		for _, t := range types {
			f, ok := cores[strings.ToLower(t)]
			if !ok {
				return nil, errors.New("unknown core type: " + t)
			}
			core1, err := f(c)
			if err != nil {
				return nil, err
			}
			cs = append(cs, core1)
		}
		return &Selector{
			cores: cs,
		}, nil
	}
	// one core
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
