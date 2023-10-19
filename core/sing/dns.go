package sing

import (
	"bytes"
	"github.com/InazumaV/V2bX/api/panel"
	"github.com/goccy/go-json"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
)

func updateDNSConfig(node *panel.NodeInfo) (err error) {
	dnsPath := os.Getenv("SING_DNS_PATH")
	if len(node.RawDNS.DNSJson) != 0 {
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, node.RawDNS.DNSJson, "", " "); err != nil {
			return err
		}
		err = saveDnsConfig(prettyJSON.Bytes(), dnsPath)
	} else if len(node.RawDNS.DNSMap) != 0 {
		dnsConfig := DNSConfig{
			Servers: []map[string]interface{}{
				{
					"tag":     "default",
					"address": "https://8.8.8.8/dns-query",
					"detour":  "direct",
				},
			},
		}
		for id, value := range node.RawDNS.DNSMap {
			dnsConfig.Servers = append(dnsConfig.Servers,
				map[string]interface{}{
					"tag":              id,
					"address":          value["address"],
					"address_resolver": "default",
					"detour":           "direct",
				},
			)
			rule := map[string]interface{}{
				"server":        id,
				"disable_cache": true,
			}
			for _, ruleType := range []string{"domain_suffix", "domain_keyword", "domain_regex", "geosite"} {
				var domains []string
				for _, v := range value["domains"].([]string) {
					split := strings.SplitN(v, ":", 2)
					prefix := strings.ToLower(split[0])
					if prefix == ruleType || (prefix == "domain" && ruleType == "domain_suffix") {
						if len(split) > 1 {
							domains = append(domains, split[1])
						}
						if len(domains) > 0 {
							rule[ruleType] = domains
						}
					}
				}
			}
			dnsConfig.Rules = append(dnsConfig.Rules, rule)
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
		log.WithField("err", err).Error("Failed to read SING_DNS_PATH")
		return err
	}
	if !bytes.Equal(currentData, dns) {
		if err = os.Truncate(dnsPath, 0); err != nil {
			log.WithField("err", err).Error("Failed to clear SING DNS PATH file")
		}
		if err = os.WriteFile(dnsPath, dns, 0644); err != nil {
			log.WithField("err", err).Error("Failed to write DNS to SING DNS PATH file")
		}
	}
	return err
}
