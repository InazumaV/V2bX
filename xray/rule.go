package xray

import (
	"github.com/Yuzuki616/V2bX/api"
	"github.com/Yuzuki616/V2bX/app/dispatcher"
	"github.com/xtls/xray-core/features/routing"
)

func (p *Xray) UpdateRule(tag string, newRuleList []api.DetectRule) error {
	dispather := p.Server.GetFeature(routing.DispatcherType()).(*dispatcher.DefaultDispatcher)
	err := dispather.RuleManager.UpdateRule(tag, newRuleList)
	return err
}

func (p *Xray) UpdateProtocolRule(tag string, newRuleList []string) error {
	dispather := p.Server.GetFeature(routing.DispatcherType()).(*dispatcher.DefaultDispatcher)
	err := dispather.RuleManager.UpdateProtocolRule(tag, newRuleList)
	return err
}

func (p *Xray) GetDetectResult(tag string) (*[]api.DetectResult, error) {
	dispather := p.Server.GetFeature(routing.DispatcherType()).(*dispatcher.DefaultDispatcher)
	return dispather.RuleManager.GetDetectResult(tag)
}
