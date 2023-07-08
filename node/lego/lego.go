package lego

import (
	"fmt"
	"github.com/Yuzuki616/V2bX/common/file"
	"github.com/Yuzuki616/V2bX/conf"
	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/lego"
	"os"
	"path"
)

type Lego struct {
	client *lego.Client
	config *conf.CertConfig
}

func New(config *conf.CertConfig) (*Lego, error) {
	user, err := NewUser(path.Join(path.Dir(config.CertFile),
		"user",
		fmt.Sprintf("user-%s.json", config.Email)),
		config.Email)
	if err != nil {
		return nil, fmt.Errorf("create user error: %s", err)
	}
	c := lego.NewConfig(user)
	//c.CADirURL = "http://192.168.99.100:4000/directory"
	c.Certificate.KeyType = certcrypto.RSA2048
	client, err := lego.NewClient(c)
	if err != nil {
		return nil, err
	}
	l := Lego{
		client: client,
		config: config,
	}
	err = l.SetProvider()
	if err != nil {
		return nil, fmt.Errorf("set provider error: %s", err)
	}
	return &l, nil
}

func checkPath(p string) error {
	if !file.IsExist(path.Dir(p)) {
		err := os.MkdirAll(path.Dir(p), 0755)
		if err != nil {
			return fmt.Errorf("create dir error: %s", err)
		}
	}
	return nil
}
