package node

import (
	"fmt"

	"github.com/InazumaV/V2bX/api/panel"
	"github.com/InazumaV/V2bX/conf"
	vCore "github.com/InazumaV/V2bX/core"
)

type Node struct {
	controllers []*Controller
}

func New() *Node {
	return &Node{}
}

func (n *Node) Start(nodes []*conf.NodeConfig, core vCore.Core) error {
	n.controllers = make([]*Controller, len(nodes))
	for i, c := range nodes {
		p, err := panel.New(c.ApiConfig)
		if err != nil {
			return err
		}
		// Register controller service
		n.controllers[i] = NewController(core, p, c.Options)
		err = n.controllers[i].Start()
		if err != nil {
			return fmt.Errorf("start node controller [%s-%s-%d] error: %s",
				c.ApiConfig.APIHost,
				c.ApiConfig.NodeType,
				c.ApiConfig.NodeID,
				err)
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
