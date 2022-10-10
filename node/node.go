package node

import (
	"github.com/Yuzuki616/V2bX/api/panel"
	"github.com/Yuzuki616/V2bX/conf"
	"github.com/Yuzuki616/V2bX/core"
	"github.com/Yuzuki616/V2bX/node/controller"
)

type Node struct {
	controllers []*controller.Node
}

func New() *Node {
	return &Node{}
}

func (n *Node) Start(nodes []*conf.NodeConfig, core *core.Core) error {
	n.controllers = make([]*controller.Node, len(nodes))
	for i, c := range nodes {
		// Register controller service
		n.controllers[i] = controller.New(core, panel.New(c.ApiConfig), c.ControllerConfig)
		err := n.controllers[i].Start()
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
