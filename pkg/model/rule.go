package model

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aquasecurity/esquery"

	"github.com/openinsight-proj/elastic-alert/pkg/utils"
	"github.com/openinsight-proj/elastic-alert/pkg/utils/logger"
	"github.com/openinsight-proj/elastic-alert/pkg/utils/xtime"
)

type EsConfig struct {
	Addresses   []string `json:"addresses" yaml:"addresses"`
	Username    string   `json:"username" yaml:"username"`
	Password    string   `json:"password" yaml:"password"`
	ConnTimeout uint     `json:"conn_timeout" yaml:"conn_timeout" default:"10"`
	Version     string   `json:"version" yaml:"version" default:"v7"`
}

type Query struct {
	Type         string            `json:"type" yaml:"type"`
	Config       QueryConfig       `json:"config" yaml:"config"`
	QueryString  string            `json:"query_string" yaml:"query_string"`
	BooleanQuery BooleanQuery      `json:"boolean_query" yaml:"boolean_query"`
	Labels       map[string]string `json:"labels" yaml:"labels"`
	Annotations  map[string]string `json:"annotations" yaml:"annotations"`
}

type QueryConfig struct {
	Timeframe xtime.TimeLimit `json:"timeframe" yaml:"timeframe"`
	NumEvents uint            `json:"num_events" yaml:"num_events"`
}

type Rule struct {
	UniqueId   string          `json:"unique_id" yaml:"unique_id"`
	Enabled    bool            `json:"enabled" yaml:"enabled" default:"true"`
	ES         EsConfig        `json:"es" yaml:"es"`
	Index      string          `json:"index" yaml:"index"`
	RunEvery   xtime.TimeLimit `json:"run_every" yaml:"run_every"`
	Query      Query           `json:"query" yaml:"query"`
	RawContent string          `json:"-"`
	FilePath   string          `json:"-"`
}

type BooleanQuery string

// addTimeRangeFilter will add or override a time range clause into bool.filter clause
func addTimeRangeFilter(booleanQuery map[string]any, start, end time.Time) map[string]any {
	timeRange := map[string]any{
		"range": map[string]any{
			"time": map[string]any{
				"format": "strict_date_optional_time",
				"gte":    start.UTC().Format(time.RFC3339Nano),
				"lte":    end.UTC().Format(time.RFC3339Nano),
			},
		},
	}

	// check bool exist
	if booleanQuery["bool"] == nil {
		return booleanQuery
	}

	// check filter exist in bool
	if _, ok := booleanQuery["bool"].(map[string]any)["filter"]; !ok {
		b := booleanQuery["bool"].(map[string]any)
		b["filter"] = []any{timeRange}
		return booleanQuery
	}

	// loop each filterList and delete exist range time
	if fl, ok := booleanQuery["bool"].(map[string]any)["filter"].([]any); ok {
		for index, v := range fl {
			if value, ok := v.(map[string]any); ok && value["range"] != nil {
				if r, ok := value["range"].(map[string]any); ok && r["time"] != nil {
					booleanQuery["bool"].(map[string]any)["filter"] = append(booleanQuery["bool"].(map[string]any)["filter"].([]any)[:index], booleanQuery["bool"].(map[string]any)["filter"].([]any)[index+1:]...)
					break
				}
			}
		}
	}

	booleanQuery["bool"].(map[string]any)["filter"] = append(booleanQuery["bool"].(map[string]any)["filter"].([]any), timeRange)

	return booleanQuery
}

func (b BooleanQuery) GetDSL(from int, size int, start time.Time, end time.Time) string {
	booleanQuery := make(map[string]any)
	err := json.Unmarshal([]byte(b), &booleanQuery)
	if err != nil {
		logger.Logger.Errorln(fmt.Sprintf("error Unmarshal booleanQuery: %q, error message: %q, use a default booleanQuery.", b, err.Error()))
	}

	booleanQuery = addTimeRangeFilter(booleanQuery, start, end)

	searchRequest := esquery.Search().
		Query(esquery.CustomQuery(booleanQuery)).
		Sort("time", esquery.OrderAsc).
		From(uint64(from)).Size(uint64(size))

	dsl, err := searchRequest.MarshalJSON()
	if err != nil {
		logger.Logger.Errorln(fmt.Sprintf("error build DSL, raw booleanQuery: %q, error message: %q", booleanQuery, err.Error()))
		return ""
	}
	return string(dsl)
}

func (b BooleanQuery) GetCountDSL(start time.Time, end time.Time) string {
	booleanQuery := make(map[string]any)
	err := json.Unmarshal([]byte(b), &booleanQuery)
	if err != nil {
		logger.Logger.Errorln(fmt.Sprintf("error Unmarshal booleanQuery: %q, error message: %q, use a default booleanQuery.", b, err.Error()))
	}

	booleanQuery = addTimeRangeFilter(booleanQuery, start, end)

	searchRequest := esquery.Search().
		Query(esquery.CustomQuery(booleanQuery))

	dsl, err := searchRequest.MarshalJSON()
	if err != nil {
		logger.Logger.Errorln(fmt.Sprintf("error build DSL, raw booleanQuery: %q, error message: %q", booleanQuery, err.Error()))
		return ""
	}
	return string(dsl)
}

func (rl *Rule) GetQueryStringDSL(from int, size int, start time.Time, end time.Time) string {
	if rl.Query.BooleanQuery != "" {
		return rl.Query.BooleanQuery.GetDSL(from, size, start, end)
	}
	q := `
{
    "query":{
        "bool":{
            "must":[
                {
                    "query_string":{
                        "query": %s
                    }
                },
                {
                    "range":{
                        "@timestamp":{
                            "format":"strict_date_optional_time",
                            "gte":"%s",
                            "lte":"%s"
                        }
                    }
                }
            ]
        }
    },
    "sort":[
        {
            "time":{
                "order":"asc"
            }
        }
    ],
	"from": %d,
    "size": %d
}
    `
	dsl := fmt.Sprintf(q, strconv.Quote(rl.Query.QueryString), xtime.TimeFormatISO8601(start), xtime.TimeFormatISO8601(end), from, size)
	return dsl
}

func (rl *Rule) GetQueryStringCountDSL(start time.Time, end time.Time) string {
	if rl.Query.BooleanQuery != "" {
		return rl.Query.BooleanQuery.GetCountDSL(start, end)
	}
	q := `
{
    "query":{
        "bool":{
            "must":[
                {
                    "query_string":{
                        "query": %s
                    }
                },
                {
                    "range":{
                        "time":{
                            "format":"strict_date_optional_time",
                            "gte":"%s",
                            "lte":"%s"
                        }
                    }
                }
            ]
        }
    }
}
    `
	dsl := fmt.Sprintf(q, strconv.Quote(rl.Query.QueryString), xtime.TimeFormatISO8601(start), xtime.TimeFormatISO8601(end))
	return dsl
}

func (rl *Rule) GetMetricsQueryFingerprint(statusCode int) string {
	f := []string{rl.UniqueId, rl.FilePath, rl.GetEsAddress(), rl.Index, strconv.Itoa(statusCode)}
	return utils.MD5(strings.Join(f, ""))
}

func (rl *Rule) GetMetricsOpRedisFingerprint(cmd string, key string, statusCode int) string {
	return GetMetricsOpRedisFingerprint(rl.UniqueId, rl.FilePath, cmd, key, statusCode)
}

func (rl *Rule) GetEsAddress() string {
	return strings.Join(rl.ES.Addresses, ",")
}

func GetMetricsOpRedisFingerprint(uniqueId string, path string, cmd string, key string, statusCode int) string {
	f := []string{uniqueId, path, cmd, key, strconv.Itoa(statusCode)}
	return utils.MD5(strings.Join(f, ""))
}

func GetMetricsAlertFingerprint(uniqueId string, path string, key string) string {
	f := []string{uniqueId, path, key}
	return utils.MD5(strings.Join(f, ""))
}

func GetMetricsWebhookNotifyFingerprint(uniqueId string, path string, statusCode int) string {
	f := []string{uniqueId, path, strconv.Itoa(statusCode)}
	return utils.MD5(strings.Join(f, ""))
}

func BuildFindByIdsDSLBody(ids []string) string {
	m := map[string]any{
		"query": map[string]any{
			"ids": map[string]any{
				"values": ids,
			},
		},
		"sort": []map[string]any{
			{
				"@timestamp": map[string]string{
					"order": "asc",
				},
			},
		},
	}
	bs, _ := json.Marshal(m)
	return string(bs)
}

var RuleYamlSchema = `
type: object
required: ["unique_id", "es", "index", "run_every", "query"]
properties:
  unique_id:
    type: string
  enabled:
    type: boolean
  es:
    type: object
    required: []
    properties:
      addresses: {type: array, items: {type: string}}
      username: {type: string}
      password: {type: string}
      version: {type: string, enum: ["v7", "v8"]}
  index:
    type: string
  run_every:
    type: object
    required: []
    properties:
      seconds: {type: number}
      minutes: {type: number}
      days: {type: number}
  query:
    type: object
    required: ["type", "config", "labels", "annotations"]
    properties:
      type: {type: string, enum: ["frequency"]}
      query_string: {type: string}
      boolean_query: {type: string}
      config: {type: object, required: ["timeframe", "num_events"], properties: {timeframe: {type: object, required: [], properties: {minutes: {type: number}}}, num_events: {type: number}}}
      labels: {type: object, required: ["alertname"], properties: {alertname: {type: string}, instance: {type: string}, severity: {type: string}, for_time: {type: string}, threshold: {type: string}}}
      annotations: {type: object, required: [], properties: {description: {type: string}, summary: {type: string}}}
`
