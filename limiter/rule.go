package limiter

import (
	"reflect"
	"regexp"
)

func (l *Limiter) CheckDomainRule(destination string) (reject bool) {
	// have rule
	for i := range l.Rules {
		if l.Rules[i].MatchString(destination) {
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

func (l *Limiter) UpdateRule(newRuleList []*regexp.Regexp) error {
	if !reflect.DeepEqual(l.Rules, newRuleList) {
		l.Rules = newRuleList
	}
	return nil
}
