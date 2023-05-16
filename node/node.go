package node

import (
	"github.com/Yuzuki616/V2bX/api/panel"
	"github.com/Yuzuki616/V2bX/conf"
	"github.com/Yuzuki616/V2bX/core"
)

type Node struct {
	controllers []*Controller
}

func New() *Node {
	return &Node{}
}

func (n *Node) Start(nodes []*conf.NodeConfig, core *core.Core) error {
	n.controllers = make([]*Controller, len(nodes))
	for i, c := range nodes {
		p, err := panel.New(c.ApiConfig)
		if err != nil {
			return err
		}
		// Register controller service
		n.controllers[i] = NewController(core, p, c.ControllerConfig)
		err = n.controllers[i].Start()
		if err != nil {
			return err
		}
	}
	return nil
}

func (n *Node) Close() {
	for _, c := range n.controllers {
		err := c.Close()
		if err != nil {
			panic(err)
		}
	}
	n.controllers = nil
}
