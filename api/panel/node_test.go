package panel

import (
	"github.com/Yuzuki616/V2bX/conf"
	"github.com/Yuzuki616/V2bX/node/controller/legoCmd/log"
	"testing"
)

func TestClient_GetNodeInfo(t *testing.T) {
	c, err := New(&conf.ApiConfig{
		APIHost:  "http://127.0.0.1",
		Key:      "token",
		NodeType: "V2ray",
		NodeID:   1,
	})
	if err != nil {
		log.Print(err)
	}
	log.Println(c.GetNodeInfo())
}
