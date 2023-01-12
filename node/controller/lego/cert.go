package lego

import (
	"fmt"
	"github.com/Yuzuki616/V2bX/common/file"
	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge/http01"
	"github.com/go-acme/lego/v4/providers/dns"
	"os"
	"path"
	"strings"
	"time"
)

func (l *Lego) SetProvider() error {
	switch l.config.CertMode {
	case "http":
		err := l.client.Challenge.SetHTTP01Provider(http01.NewProviderServer("", "5002"))
		if err != nil {
			return err
		}
	case "dns":
		for k, v := range l.config.DNSEnv {
			os.Setenv(k, v)
		}
		p, err := dns.NewDNSChallengeProviderByName(l.config.Provider)
		if err != nil {
			return fmt.Errorf("create dns challenge provider error: %s", err)
		}
		err = l.client.Challenge.SetDNS01Provider(p)
		if err != nil {
			return fmt.Errorf("set dns provider error: %s", err)
		}
	}
	return nil
}

func (l *Lego) CreateCert() (err error) {
	request := certificate.ObtainRequest{
		Domains: []string{l.config.CertDomain},
	}
	certificates, err := l.client.Certificate.Obtain(request)
	if err != nil {
		return fmt.Errorf("obtain certificate error: %s", err)
	}
	err = l.writeCert(certificates)
	return nil
}

func (l *Lego) RenewCert() error {
	file, err := os.ReadFile(l.config.CertFile)
	if err != nil {
		return fmt.Errorf("read cert file error: %s", err)
	}
	if e, err := l.CheckCert(file); !e {
		return nil
	} else if err != nil {
		return fmt.Errorf("check cert error: %s", err)
	}
	res, err := l.client.Certificate.Renew(certificate.Resource{
		Domain:      l.config.CertDomain,
		Certificate: file,
	}, false, false, "")
	if err != nil {
		return err
	}
	err = l.writeCert(res)
	return nil
}

func (l *Lego) CheckCert(file []byte) (bool, error) {
	cert, err := certcrypto.ParsePEMCertificate(file)
	if err != nil {
		return false, err
	}
	notAfter := int(time.Until(cert.NotAfter).Hours() / 24.0)
	if notAfter > 30 {
		return false, nil
	}
	return true, nil
}
func (l *Lego) parseParmas(path string) string {
	r := strings.NewReplacer("{domain}", l.config.CertDomain,
		"{email}", l.config.Email)
	return r.Replace(path)
}
func (l *Lego) writeCert(certificates *certificate.Resource) error {
	if !file.IsExist(path.Dir(l.config.CertFile)) {
		err := os.MkdirAll(path.Dir(l.config.CertFile), 0755)
		if err != nil {
			return fmt.Errorf("create dir error: %s", err)
		}
	}
	err := os.WriteFile(l.parseParmas(l.config.CertFile), certificates.Certificate, 0644)
	if err != nil {
		return err
	}
	if !file.IsExist(path.Dir(l.config.KeyFile)) {
		err := os.MkdirAll(path.Dir(l.config.CertFile), 0755)
		if err != nil {
			return fmt.Errorf("create dir error: %s", err)
		}
	}
	err = os.WriteFile(l.parseParmas(l.config.KeyFile), certificates.PrivateKey, 0644)
	if err != nil {
		return err
	}
	return nil
}
