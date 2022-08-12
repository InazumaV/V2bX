package core

import (
	"github.com/Yuzuki616/V2bX/api"
)

func (p *Core) UpdateRule(tag string, newRuleList []api.DetectRule) error {
	return p.dispatcher.RuleManager.UpdateRule(tag, newRuleList)
}

func (p *Core) UpdateProtocolRule(tag string, newRuleList []string) error {

	return p.dispatcher.RuleManager.UpdateProtocolRule(tag, newRuleList)
}

func (p *Core) GetDetectResult(tag string) ([]api.DetectResult, error) {
	return p.dispatcher.RuleManager.GetDetectResult(tag)
}
