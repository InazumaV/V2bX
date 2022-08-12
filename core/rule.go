package core

import (
	"github.com/Yuzuki616/V2bX/api"
)

func (p *Core) UpdateRule(tag string, newRuleList []api.DetectRule) error {
	err := p.dispatcher.RuleManager.UpdateRule(tag, newRuleList)
	return err
}

func (p *Core) UpdateProtocolRule(tag string, newRuleList []string) error {
	err := p.dispatcher.RuleManager.UpdateProtocolRule(tag, newRuleList)
	return err
}

func (p *Core) GetDetectResult(tag string) ([]api.DetectResult, error) {
	return p.dispatcher.RuleManager.GetDetectResult(tag)
}
