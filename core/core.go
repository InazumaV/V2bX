package core

import (
	"errors"
	"strings"

	"github.com/InazumaV/V2bX/conf"
)

var (
	cores = map[string]func(c *conf.CoreConfig) (Core, error){}
)

func NewCore(c []conf.CoreConfig) (Core, error) {
	if len(c) < 0 {
		return nil, errors.New("no have vail core")
	}
	// multi core
	if len(c) > 1 {
		var cs []Core
		for _, t := range c {
			f, ok := cores[strings.ToLower(t.Type)]
			if !ok {
				return nil, errors.New("unknown core type: " + t.Type)
			}
			core1, err := f(&t)
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
	if f, ok := cores[c[0].Type]; ok {
		return f(&c[0])
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
