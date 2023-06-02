package conf

import (
	"log"
	"testing"
)

func TestConf_LoadFromPath(t *testing.T) {
	c := New()
	t.Log(c.LoadFromPath("../example/config.yml.example"), c.NodesConfig[0].ControllerConfig.EnableXtls)
}

func TestConf_Watch(t *testing.T) {
	c := New()
	c.Watch("../example/config.yml.example", func() {
		log.Println(1)
	})
	select {}
}
