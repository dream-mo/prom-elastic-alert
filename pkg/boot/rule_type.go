package boot

import (
	"github.com/openinsight-proj/elastic-alert/pkg/utils/xtime"
	"strings"
	"time"

	"github.com/openinsight-proj/elastic-alert/pkg/conf"
)

type Match struct {
	r          *conf.Rule
	Ids        []string
	StartsAt   time.Time
	EndsAt     time.Time
	HitsNumber int
}

func (mc *Match) Fingerprint() string {
	return mc.r.UniqueId
}

type RuleType interface {
	GetMatches(r *conf.Rule, hits []any) []Match
	FilterMatchCondition(r *conf.Rule, matches []Match) *Match
}

type FrequencyRule struct {
}

func (fr *FrequencyRule) FilterMatchCondition(r *conf.Rule, matches []Match) *Match {
	var match *Match
	for _, m := range matches {
		if uint(len(m.Ids)) >= r.Query.Config.NumEvents {
			match = &m
			break
		}
	}
	return match
}
func (fr *FrequencyRule) GetMatches(r *conf.Rule, resultHits []any) []Match {
	matches := make([]Match, 10)
	hasAgg := false
	var match Match
	for i := 0; i < len(resultHits); i++ {
		item := resultHits[i]
		m := item.(map[string]any)
		_id := m["_id"].(string)
		_source := m["_source"].(map[string]any)
		timestamp := _source["@timestamp"].(string)
		ts := xtime.Parse(timestamp)
		match.HitsNumber = len(resultHits)
		if !hasAgg {
			match.StartsAt = ts
			match.EndsAt = ts.Add(r.Query.Config.Timeframe.GetTimeDuration())
			match.Ids = []string{_id}
			match.r = r
			hasAgg = true
		} else {
			if ts.Before(match.EndsAt) {
				match.Ids = append(match.Ids, _id)
			} else {
				matches = append(matches, match)
				resultHits = resultHits[i:]
				i = 0
				match = Match{}
				hasAgg = false
			}
		}
	}
	if match.r != nil {
		matches = append(matches, match)
	}
	return matches
}

func NewRuleType(t string) RuleType {
	t = strings.ToLower(t)
	m := map[string]RuleType{
		"frequency": &FrequencyRule{},
	}
	rt, _ := m[t]
	return rt
}
