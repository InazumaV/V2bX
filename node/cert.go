package node

import (
	"fmt"
	"github.com/Yuzuki616/V2bX/common/file"
	"github.com/Yuzuki616/V2bX/node/lego"
	"log"
)

func (c *Controller) renewCertTask() {
	l, err := lego.New(c.CertConfig)
	if err != nil {
		log.Print("new lego error: ", err)
		return
	}
	err = l.RenewCert()
	if err != nil {
		log.Print("renew cert error: ", err)
		return
	}
}

func (c *Controller) requestCert() error {
	switch c.CertConfig.CertMode {
	case "reality", "none", "":
		return nil
	case "file":
		if c.CertConfig.CertFile == "" || c.CertConfig.KeyFile == "" {
			return fmt.Errorf("cert file path or key file path not exist")
		}
		return nil
	case "dns", "http":
		if c.CertConfig.CertFile == "" || c.CertConfig.KeyFile == "" {
			return fmt.Errorf("cert file path or key file path not exist")
		}
		if file.IsExist(c.CertConfig.CertFile) && file.IsExist(c.CertConfig.KeyFile) {
			return nil
		}
		l, err := lego.New(c.CertConfig)
		if err != nil {
			return fmt.Errorf("create lego object error: %s", err)
		}
		err = l.CreateCert()
		if err != nil {
			return fmt.Errorf("create cert error: %s", err)
		}
		return nil
	}
	return fmt.Errorf("unsupported certmode: %s", c.CertConfig.CertMode)
}
