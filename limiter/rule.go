package limiter

import (
	"github.com/Yuzuki616/V2bX/api/panel"
	"reflect"
)

func (l *Limiter) CheckDomainRule(destination string) (reject bool) {
	// have rule
	for i := range l.Rules {
		if l.Rules[i].Pattern.MatchString(destination) {
			reject = true
			break
		}
	}
	return
}

func (l *Limiter) CheckProtocolRule(protocol string) (reject bool) {
	for i := range l.ProtocolRules {
		if l.ProtocolRules[i] == protocol {
			reject = true
			break
		}
	}
	return
}

func (l *Limiter) UpdateRule(newRuleList []panel.DestinationRule) error {
	if !reflect.DeepEqual(l.Rules, newRuleList) {
		l.Rules = newRuleList
	}
	return nil
}
