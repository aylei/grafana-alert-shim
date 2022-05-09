package rule

import (
	"github.com/pkg/errors"
	"github.com/prometheus/prometheus/model/rulefmt"
)

// ValidateRuleGroup validate the given rule group, modified from cortex
func ValidateRuleGroup(rg *rulefmt.RuleGroup) []error {
	if rg.Name == "" {
		return []error{errors.New("Name must not be empty")}
	}
	if len(rg.Rules) < 1 {
		return []error{errors.Errorf("Rule group %s has no rule", rg.Name)}
	}

	var errs []error
	for i, r := range rg.Rules {
		if errList := r.Validate(); errList != nil {
			for _, err := range errList {
				var ruleName string
				if r.Alert.Value != "" {
					ruleName = r.Alert.Value
				} else {
					ruleName = r.Record.Value
				}
				errs = append(errs, &rulefmt.Error{
					Group:    rg.Name,
					Rule:     i,
					RuleName: ruleName,
					Err:      err,
				})
			}
		}
	}

	return errs
}
