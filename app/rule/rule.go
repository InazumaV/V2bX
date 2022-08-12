// Package rule is to control the audit rule behaviors
package rule

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/Yuzuki616/V2bX/api"
	mapset "github.com/deckarep/golang-set"
)

type Rule struct {
	InboundRule         *sync.Map // Key: Tag, Value: []api.DetectRule
	InboundProtocolRule *sync.Map // Key: Tag, Value: []string
	InboundDetectResult *sync.Map // key: Tag, Value: mapset.NewSet []api.DetectResult
}

func New() *Rule {
	return &Rule{
		InboundRule:         new(sync.Map),
		InboundProtocolRule: new(sync.Map),
		InboundDetectResult: new(sync.Map),
	}
}

func (r *Rule) UpdateRule(tag string, newRuleList []api.DetectRule) error {
	if value, ok := r.InboundRule.LoadOrStore(tag, newRuleList); ok {
		oldRuleList := value.([]api.DetectRule)
		if !reflect.DeepEqual(oldRuleList, newRuleList) {
			r.InboundRule.Store(tag, newRuleList)
		}
	}
	return nil
}

func (r *Rule) UpdateProtocolRule(tag string, ruleList []string) error {
	if value, ok := r.InboundProtocolRule.LoadOrStore(tag, ruleList); ok {
		old := value.([]string)
		if !reflect.DeepEqual(old, ruleList) {
			r.InboundProtocolRule.Store(tag, ruleList)
		}
	}
	return nil
}

func (r *Rule) GetDetectResult(tag string) ([]api.DetectResult, error) {
	detectResult := make([]api.DetectResult, 0)
	if value, ok := r.InboundDetectResult.LoadAndDelete(tag); ok {
		resultSet := value.(mapset.Set)
		it := resultSet.Iterator()
		for result := range it.C {
			detectResult = append(detectResult, result.(api.DetectResult))
		}
	}
	return detectResult, nil
}

func (r *Rule) Detect(tag string, destination string, email string) (reject bool) {
	reject = false
	var hitRuleID = -1
	// If we have some rule for this inbound
	if value, ok := r.InboundRule.Load(tag); ok {
		ruleList := value.([]api.DetectRule)
		for _, r := range ruleList {
			if r.Pattern.Match([]byte(destination)) {
				hitRuleID = r.ID
				reject = true
				break
			}
		}
		// If we hit some rule
		if reject && hitRuleID != -1 {
			l := strings.Split(email, "|")
			uid, err := strconv.Atoi(l[len(l)-1])
			if err != nil {
				newError(fmt.Sprintf("Record illegal behavior failed! Cannot find user's uid: %s", email)).AtDebug().WriteToLog()
				return reject
			}
			newSet := mapset.NewSetWith(api.DetectResult{UID: uid, RuleID: hitRuleID})
			// If there are any hit history
			if v, ok := r.InboundDetectResult.LoadOrStore(tag, newSet); ok {
				resultSet := v.(mapset.Set)
				// If this is a new record
				if resultSet.Add(api.DetectResult{UID: uid, RuleID: hitRuleID}) {
					r.InboundDetectResult.Store(tag, resultSet)
				}
			}
		}
	}
	return reject
}
func (r *Rule) ProtocolDetect(tag string, protocol string) bool {
	if value, ok := r.InboundProtocolRule.Load(tag); ok {
		ruleList := value.([]string)
		for _, r := range ruleList {
			if r == protocol {
				return true
			}
		}
	}
	return false
}
