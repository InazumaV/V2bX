package xray

import (
	"bytes"
	"github.com/InazumaV/V2bX/api/panel"
	"github.com/goccy/go-json"
	log "github.com/sirupsen/logrus"
	coreConf "github.com/xtls/xray-core/infra/conf"
	"os"
)

func updateDNSConfig(node *panel.NodeInfo) (err error) {
	dnsPath := os.Getenv("XRAY_DNS_PATH")
	if len(node.RawDNS.DNSJson) != 0 {
		err = saveDnsConfig(node.RawDNS.DNSJson, dnsPath)
	} else if len(node.RawDNS.DNSMap) != 0 {
		dnsConfig := DNSConfig{
			Servers: []interface{}{
				"1.1.1.1",
				"localhost"},
			Tag: "dns_inbound",
		}
		for _, value := range node.RawDNS.DNSMap {
			dnsConfig.Servers = append(dnsConfig.Servers, value)
		}
		dnsConfigJSON, err := json.MarshalIndent(dnsConfig, "", "  ")
		if err != nil {
			log.WithField("err", err).Error("Error marshaling dnsConfig to JSON")
			return err
		}
		err = saveDnsConfig(dnsConfigJSON, dnsPath)
	}
	return err
}

func saveDnsConfig(dns []byte, dnsPath string) (err error) {
	currentData, err := os.ReadFile(dnsPath)
	if err != nil {
		log.WithField("err", err).Error("Failed to read XRAY_DNS_PATH")
		return err
	}
	if !bytes.Equal(currentData, dns) {
		coreDnsConfig := &coreConf.DNSConfig{}
		if err = json.NewDecoder(bytes.NewReader(dns)).Decode(coreDnsConfig); err != nil {
			log.WithField("err", err).Error("Failed to unmarshal DNS config")
		}
		_, err := coreDnsConfig.Build()
		if err != nil {
			log.WithField("err", err).Error("Failed to understand DNS config, Please check: https://xtls.github.io/config/dns.html for help")
			return err
		}
		if err = os.Truncate(dnsPath, 0); err != nil {
			log.WithField("err", err).Error("Failed to clear XRAY DNS PATH file")
		}
		if err = os.WriteFile(dnsPath, dns, 0644); err != nil {
			log.WithField("err", err).Error("Failed to write DNS to XRAY DNS PATH file")
		}
	}
	return err
}
