package server

import (
	"encoding/json"
	common "scheduler/appointment-service/internal"
	"scheduler/appointment-service/internal/storage"
)

type rruleStorage struct {
	*storage.Storage
}

func (r *rruleStorage) AddRRule(user common.ID, rule RRuleWithType) error {
	return r.Storage.AddBusinessRule(user, rule.JSON)
}

func (r *rruleStorage) RemoveRRule(user common.ID, ruleId common.ID) error {
	return r.Storage.DeleteBusinessRule(user, ruleId)
}

func (r *rruleStorage) GetRRules(user common.ID) ([]RRuleResult, error) {
	dbRules, err := r.Storage.GetBusinessRules(user)

	if err != nil {
		return nil, err
	}

	result := make([]RRuleResult, len(dbRules))
	for i, rule := range dbRules {
		result[i].Id = rule.Id
		err = json.Unmarshal([]byte(rule.Rule), &result[i].Rrule)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}
