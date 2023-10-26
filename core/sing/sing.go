package sing

import (
	"context"
	"fmt"
	"os"

	"github.com/sagernet/sing-box/log"

	"github.com/InazumaV/V2bX/conf"
	vCore "github.com/InazumaV/V2bX/core"
	"github.com/goccy/go-json"
	box "github.com/sagernet/sing-box"
	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/option"
)

var _ vCore.Core = (*Sing)(nil)

type DNSConfig struct {
	Servers []map[string]interface{} `json:"servers"`
	Rules   []map[string]interface{} `json:"rules"`
}

type Sing struct {
	box        *box.BoxEx
	ctx        context.Context
	hookServer *HookServer
	router     adapter.RouterEx
	logFactory log.Factory
	inbounds   map[string]adapter.Inbound
}

func init() {
	vCore.RegisterCore("sing", New)
}

func New(c *conf.CoreConfig) (vCore.Core, error) {
	options := option.Options{}
	if len(c.SingConfig.OriginalPath) != 0 {
		f, err := os.Open(c.SingConfig.OriginalPath)
		if err != nil {
			return nil, fmt.Errorf("open original config error: %s", err)
		}
		defer f.Close()
		err = json.NewDecoder(f).Decode(&options)
		if err != nil {
			return nil, fmt.Errorf("decode original config error: %s", err)
		}
	}
	options.Log = &option.LogOptions{
		Disabled:  c.SingConfig.LogConfig.Disabled,
		Level:     c.SingConfig.LogConfig.Level,
		Timestamp: c.SingConfig.LogConfig.Timestamp,
		Output:    c.SingConfig.LogConfig.Output,
	}
	options.NTP = &option.NTPOptions{
		Enabled:       c.SingConfig.NtpConfig.Enable,
		WriteToSystem: true,
		ServerOptions: option.ServerOptions{
			Server:     c.SingConfig.NtpConfig.Server,
			ServerPort: c.SingConfig.NtpConfig.ServerPort,
		},
	}
	os.Setenv("SING_DNS_PATH", "")
	options.DNS = &option.DNSOptions{}
	if c.SingConfig.DnsConfigPath != "" {
		f, err := os.OpenFile(c.SingConfig.DnsConfigPath, os.O_RDWR|os.O_CREATE, 0755)
		if err != nil {
			return nil, fmt.Errorf("failed to open or create sing dns config file: %s", err)
		}
		defer f.Close()
		if err := json.NewDecoder(f).Decode(options.DNS); err != nil {
			log.Warn(fmt.Sprintf(
				"Failed to unmarshal sing dns config from file '%v': %v. Using default DNS options",
				f.Name(), err))
			options.DNS = &option.DNSOptions{}
		}
		os.Setenv("SING_DNS_PATH", c.SingConfig.DnsConfigPath)
	}
	ctx := context.Background()
	b, err := box.NewEx(box.Options{
		Context: ctx,
		Options: options,
	})
	if err != nil {
		return nil, err
	}
	hs := NewHookServer(c.SingConfig.EnableConnClear)
	b.RouterEx().SetClashServer(hs)
	b.LogFactory()
	return &Sing{
		ctx:        ctx,
		box:        b,
		hookServer: hs,
		router:     b.RouterEx(),
		logFactory: b.LogFactory(),
		inbounds:   make(map[string]adapter.Inbound),
	}, nil
}

func (b *Sing) Start() error {
	return b.box.Start()
}

func (b *Sing) Close() error {
	return b.box.Close()
}

func (b *Sing) Protocols() []string {
	return []string{
		"vmess",
		"vless",
		"shadowsocks",
		"trojan",
		"hysteria",
	}
}

func (b *Sing) Type() string {
	return "sing"
}
