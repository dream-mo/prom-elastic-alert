package services

import (
	"errors"
	"github.com/openinsight-proj/elastic-alert/pkg/model"
	"github.com/openinsight-proj/elastic-alert/pkg/utils/xtime"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func Test_ValidateRule(t *testing.T) {
	testCases := []struct {
		name string
		rule *model.Rule
		err  error
	}{{
		name: "alertname is required",
		rule: &model.Rule{
			UniqueId: "dfdfd",
			Enabled:  false,
			ES: model.EsConfig{
				Addresses: []string{"http://localhost:9200"},
				Version:   "v7",
			},
			Index: "fdfdf",
			RunEvery: xtime.TimeLimit{
				Seconds: 5,
			},
			Query: model.Query{
				Type: "frequency",
				Config: struct {
					Timeframe xtime.TimeLimit `json:"timeframe" yaml:"timeframe"`
					NumEvents uint            `json:"num_events" yaml:"num_events"`
				}{},
				QueryString: "",
				Labels: map[string]string{
					"foo": "bar",
				},
				Annotations: map[string]string{
					"foo": "bar",
				},
			},
		},
		err: errors.New("query.labels: alertname is required"),
	},
	}

	for _, tc := range testCases {
		err := validateRuleConfig(tc.rule)
		if err != nil {
			assert.Equal(t, true, strings.Contains(err.Error(), tc.err.Error()))
		} else {
			assert.NoError(t, err)
		}
	}

}
