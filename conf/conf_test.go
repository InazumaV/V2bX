package conf

import (
	"log"
	"strings"
	"testing"
)

func TestConf_LoadFromPath(t *testing.T) {
	c := New()
	t.Log(c.LoadFromPath("../example/config.yml.example"))
}

func TestConf_Watch(t *testing.T) {
	//c := New()
	log.Println(strings.Split("aaaa", " "))
}
