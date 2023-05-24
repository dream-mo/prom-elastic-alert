package domain

import (
	"github.com/openinsight-proj/elastic-alert/pkg/model"
)

type Rule struct {
	Policy string     `json:"policy"`
	Rule   model.Rule `json:"rule"`
}

type Rules struct {
	Rules []Rule `json:"rules"`
}
